package api

import (
	models "github.com/divyanshu-parihar/oxidized-scheduler/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type RouteHandler func(ctx *gin.Context)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

func (api *API) GetWheelState(c *gin.Context) {
	if api.wheel == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Wheel not initialized"})
		return
	}
	c.JSON(http.StatusOK, api.wheel.GetState())
}

func (api *API) AddEvent(c *gin.Context) {
	var task models.Task

	if err := c.ShouldBindJSON(&task); err != nil {
		slog.Error("failed to bind event", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Initialize metadata
	task.ID = uuid.New()
	task.Status = models.TASK_PENDING
	task.CreatedAt = time.Now()
	task.Version = 1

	if task.MaxAttempts == 0 {
		task.MaxAttempts = 3
	}

	// Persist to Postgres
	query := `
		INSERT INTO tasks (id, task_type, version, scheduled_at, status, payload, max_attempts, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := api.db.Exec(c.Request.Context(), query,
		task.ID, task.TaskType, task.Version, task.ScheduledTime, task.Status, task.Payload, task.MaxAttempts, task.CreatedAt,
	)

	if err != nil {
		slog.Error("failed to save task to db", "error", err, "task_id", task.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule event"})
		return
	}

	slog.Info("event scheduled", "task_id", task.ID, "type", task.TaskType)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Event scheduled successfully",
		"task_id": task.ID,
	})
}

func (api *API) UpdateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status   models.TaskStatus `json:"status"`
		Attempts int               `json:"attempts"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE tasks 
		SET status = $1, attempts = $2, updated_at = NOW(), version = version + 1
		WHERE id = $3`

	_, err := api.db.Exec(c.Request.Context(), query, req.Status, req.Attempts, id)
	if err != nil {
		slog.Error("failed to update task status", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated"})
}

func (api *API) ListEvents(c *gin.Context) {
	// Fetch pending tasks scheduled for the next 24 hours
	query := `SELECT id, task_type, version, scheduled_at, status, payload, attempts, max_attempts, created_at 
			  FROM tasks 
			  WHERE status = 'pending' AND scheduled_at <= NOW() + INTERVAL '24 hours'
			  ORDER BY scheduled_at ASC 
			  LIMIT 1000`

	rows, err := api.db.Query(c.Request.Context(), query)
	if err != nil {
		slog.Error("failed to fetch tasks", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(
			&t.ID, &t.TaskType, &t.Version, &t.ScheduledTime, &t.Status, &t.Payload, &t.Attempts, &t.MaxAttempts, &t.CreatedAt,
		)
		if err != nil {
			slog.Error("failed to scan task", "error", err)
			continue
		}
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	c.JSON(http.StatusOK, tasks)
}

func AddTask(c *gin.Context) {
	var task models.Task

	if err := c.ShouldBindJSON(&task); err != nil {
		slog.Error("failed to bind task", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	slog.Info("task added", "task_id", task.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Task added successfully", "task": task})
}
