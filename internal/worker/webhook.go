package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/divyanshu-parihar/oxidized-scheduler/models"
)

type WebhookHandler struct {
	client *http.Client
}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *WebhookHandler) Type() string {
	return "webhook_dispatch"
}

func (h *WebhookHandler) Handle(ctx context.Context, task models.Task) error {
	var payload struct {
		URL     string                 `json:"url"`
		Headers map[string]string      `json:"headers"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.URL == "" {
		return fmt.Errorf("webhook URL is missing")
	}

	slog.Info("dispatching webhook", "task_id", task.ID, "url", payload.URL)

	req := requests.
		URL(payload.URL).
		Method("POST").
		BodyJSON(payload.Data).
		Client(h.client)

	for k, v := range payload.Headers {
		req.Header(k, v)
	}

	err := req.Fetch(ctx)

	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}

	return nil
}
