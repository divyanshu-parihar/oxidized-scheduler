package wheel

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/divyanshu-parihar/oxidized-scheduler/internal/metrics"
	"github.com/divyanshu-parihar/oxidized-scheduler/internal/worker"
	"github.com/divyanshu-parihar/oxidized-scheduler/models"
	"github.com/google/uuid"
)

type Wheel struct {
	context context.Context
	tick    time.Duration
	baseURL string

	// Hierarchical Buckets
	SecondsEvent  [60][]models.Task
	MinutesEvents [60][]models.Task
	HoursEvents   [24][]models.Task

	UnProcessedQueue []models.Task
	inMemoryTasks    map[uuid.UUID]bool

	secPtr  int
	minPtr  int
	hourPtr int

	worker *worker.Worker
	mu     sync.Mutex
}

func NewWheel(baseURL string) *Wheel {
	w := worker.NewWorker()
	w.Register(worker.NewWebhookHandler())
	w.Register(worker.NewTelegramHandler())

	return &Wheel{
		context:       context.Background(),
		tick:          1 * time.Second,
		inMemoryTasks: make(map[uuid.UUID]bool),
		worker:        w,
		baseURL:       baseURL,
	}
}

func (wheel *Wheel) FetchTasks() {
	var events []models.Task
	err := requests.
		URL(fmt.Sprintf("%s/events", wheel.baseURL)).
		ToJSON(&events).
		Fetch(wheel.context)

	if err != nil {
		slog.Error("failed to fetch tasks from API", "error", err)
		return
	}

	wheel.mu.Lock()
	defer wheel.mu.Unlock()

	now := time.Now()
	for _, t := range events {
		if t.Status != models.TASK_PENDING || wheel.inMemoryTasks[t.ID] {
			continue
		}

		wheel.placeTask(t, now)
		wheel.inMemoryTasks[t.ID] = true
	}

	slog.Info("Wheel hydrated", "new_tasks", len(events))
}

func (wheel *Wheel) placeTask(t models.Task, now time.Time) {
	delta := t.ScheduledTime.Sub(now)

	if delta < 0 {
		wheel.SecondsEvent[wheel.secPtr] = append(wheel.SecondsEvent[wheel.secPtr], t)
		return
	}

	if delta < 1*time.Minute {
		idx := (wheel.secPtr + int(delta.Seconds())) % 60
		wheel.SecondsEvent[idx] = append(wheel.SecondsEvent[idx], t)
	} else if delta < 1*time.Hour {
		idx := (wheel.minPtr + int(delta.Minutes())) % 60
		wheel.MinutesEvents[idx] = append(wheel.MinutesEvents[idx], t)
	} else if delta < 24*time.Hour {
		idx := (wheel.hourPtr + int(delta.Hours())) % 24
		wheel.HoursEvents[idx] = append(wheel.HoursEvents[idx], t)
	} else {
		wheel.UnProcessedQueue = append(wheel.UnProcessedQueue, t)
	}
}

func (wheel *Wheel) Start() {
	wheel.FetchTasks()

	ticker := time.NewTicker(wheel.tick)
	go func() {
		for range ticker.C {
			wheel.processTick()
		}
	}()

	go func() {
		fetchTicker := time.NewTicker(10 * time.Second)
		for range fetchTicker.C {
			wheel.FetchTasks()
		}
	}()
}

func (wheel *Wheel) processTick() {
	wheel.mu.Lock()
	defer wheel.mu.Unlock()

	tasks := wheel.SecondsEvent[wheel.secPtr]
	wheel.SecondsEvent[wheel.secPtr] = nil

	for _, t := range tasks {
		go wheel.executeTask(t)
		delete(wheel.inMemoryTasks, t.ID)
	}

	wheel.secPtr++
	if wheel.secPtr == 60 {
		wheel.secPtr = 0
		wheel.tickMinute()
	}
}

func (wheel *Wheel) tickMinute() {
	wheel.minPtr++
	if wheel.minPtr == 60 {
		wheel.minPtr = 0
		wheel.tickHour()
	}

	tasks := wheel.MinutesEvents[wheel.minPtr]
	wheel.MinutesEvents[wheel.minPtr] = nil

	now := time.Now()
	for _, t := range tasks {
		wheel.placeTask(t, now)
	}
}

func (wheel *Wheel) tickHour() {
	wheel.hourPtr++
	if wheel.hourPtr == 24 {
		wheel.hourPtr = 0
		wheel.rebalanceUnprocessed()
	}

	tasks := wheel.HoursEvents[wheel.hourPtr]
	wheel.HoursEvents[wheel.hourPtr] = nil

	now := time.Now()
	for _, t := range tasks {
		wheel.placeTask(t, now)
	}
}

func (wheel *Wheel) rebalanceUnprocessed() {
	now := time.Now()
	oldQueue := wheel.UnProcessedQueue
	wheel.UnProcessedQueue = nil
	for _, t := range oldQueue {
		wheel.placeTask(t, now)
	}
}

type WheelState struct {
	Seconds     [60]int `json:"seconds"`
	Minutes     [60]int `json:"minutes"`
	Hours       [24]int `json:"hours"`
	Unprocessed int     `json:"unprocessed"`
	SecPtr      int     `json:"sec_ptr"`
	MinPtr      int     `json:"min_ptr"`
	HourPtr     int     `json:"hour_ptr"`
}

func (wheel *Wheel) GetState() WheelState {
	wheel.mu.Lock()
	defer wheel.mu.Unlock()

	var state WheelState
	for i := 0; i < 60; i++ {
		state.Seconds[i] = len(wheel.SecondsEvent[i])
		state.Minutes[i] = len(wheel.MinutesEvents[i])
	}
	for i := 0; i < 24; i++ {
		state.Hours[i] = len(wheel.HoursEvents[i])
	}
	state.Unprocessed = len(wheel.UnProcessedQueue)
	state.SecPtr = wheel.secPtr
	state.MinPtr = wheel.minPtr
	state.HourPtr = wheel.hourPtr
	return state
}

func (wheel *Wheel) executeTask(t models.Task) {
	start := time.Now()
	t.Attempts++

	drift := time.Since(t.ScheduledTime)
	metrics.DriftHistogram.WithLabelValues(t.TaskType).Observe(drift.Seconds())

	slog.Info("Executing Task", "id", t.ID, "type", t.TaskType, "scheduled_at", t.ScheduledTime, "attempt", t.Attempts, "drift", drift)

	err := wheel.worker.Process(wheel.context, t)
	duration := time.Since(start)

	status := models.TASK_COMPLETED
	metricsStatus := "completed"
	if err != nil {
		slog.Error("task execution failed", "id", t.ID, "error", err, "drift", drift)
		if t.Attempts < t.MaxAttempts {
			status = models.TASK_PENDING // Will be picked up again
			metricsStatus = "retrying"
		} else {
			status = models.TASK_FAILED
			metricsStatus = "failed"
		}
	}

	metrics.LatencyHistogram.WithLabelValues(t.TaskType, metricsStatus).Observe(duration.Seconds())
	metrics.TaskCounter.WithLabelValues(t.TaskType, metricsStatus).Inc()

	// Update DB status via API
	updateErr := requests.
		URL(fmt.Sprintf("%s/tasks/%s/status", wheel.baseURL, t.ID)).
		Method("PUT").
		BodyJSON(map[string]interface{}{
			"status":   status,
			"attempts": t.Attempts,
		}).
		Fetch(wheel.context)

	if updateErr != nil {
		slog.Error("failed to update task status via API", "id", t.ID, "error", updateErr)
	}

	if status == models.TASK_COMPLETED {
		slog.Info("Task completed", "id", t.ID, "duration", duration, "drift", drift)
	}
}
