package worker

import (
	"context"
	"github.com/divyanshu-parihar/oxidized-scheduler/models"
)

type TaskHandler interface {
	Handle(ctx context.Context, task models.Task) error
	Type() string
}

type Worker struct {
	handlers map[string]TaskHandler
}

func NewWorker() *Worker {
	return &Worker{
		handlers: make(map[string]TaskHandler),
	}
}

func (w *Worker) Register(handler TaskHandler) {
	w.handlers[handler.Type()] = handler
}

func (w *Worker) Process(ctx context.Context, task models.Task) error {
	handler, ok := w.handlers[task.TaskType]
	if !ok {
		return nil // Or return an error if unknown type
	}
	return handler.Handle(ctx, task)
}
