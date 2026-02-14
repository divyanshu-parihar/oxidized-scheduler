package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type TaskStatus string

const (
	TASK_PENDING    TaskStatus = "pending"
	TASK_DISPATCHED TaskStatus = "dispatched"
	TASK_COMPLETED  TaskStatus = "completed"
	TASK_FAILED     TaskStatus = "failed"
	TASK_RETRYING   TaskStatus = "retrying"
)

type Task struct {
	// Task metadata
	ID            uuid.UUID       `json:"id"`
	TaskType      string          `json:"task_type"`
	Version       int             `json:"version"`
	ScheduledTime time.Time       `json:"scheduled_time"`
	Status        TaskStatus      `json:"status"`
	Payload       json.RawMessage `json:"payload"`
	Attempts      int             `json:"attempts"`
	MaxAttempts   int             `json:"max_attempts"`
	CreatedAt     time.Time       `json:"created_at"`
}
