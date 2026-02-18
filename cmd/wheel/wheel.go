package wheel

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/divyanshu-parihar/oxidized-scheduler/api"
	"github.com/divyanshu-parihar/oxidized-scheduler/models"
)

type WheelTaskStatus bool

const (
	NotInWheel WheelTaskStatus = false
	InWheel    WheelTaskStatus = true
)

type WheelTask struct {
	task      models.Task
	startTime *time.Time
}

type Wheel struct {
	dbApi            *api.API
	context          *context.Context
	tick             time.Duration
	SecondsEvent     []models.Task
	HoursEvents      []models.Task
	MinutesEvents    []models.Task
	UnProcessesQueue []models.Task
	mu               sync.Mutex
}

func NewWheel(dbApi *api.API) Wheel {
	return Wheel{
		dbApi: dbApi,
		tick:  time.Duration(1000),
	}
}

func (wheel *Wheel) processTick(tickDuration time.Duration) {

	ticker := time.NewTicker(tickDuration)

	for currTime := range ticker.C {

		var events []models.Task
		err := requests.
			URL("http://localhost:8080/events").
			ToJSON(&events).
			Fetch(*wheel.context)
		if err != nil {
			slog.Error("could not connect to example.com:", err.Error)
		}

		slog.Info("New Tick at", currTime.String())
	}

}
