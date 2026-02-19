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

type TelegramHandler struct {
	client *http.Client
}

func NewTelegramHandler() *TelegramHandler {
	return &TelegramHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *TelegramHandler) Type() string {
	return "telegram_message"
}

func (h *TelegramHandler) Handle(ctx context.Context, task models.Task) error {
	var payload struct {
		BotToken string `json:"bot_token"`
		ChatID   string `json:"chat_id"`
		Message  string `json:"message"`
	}

	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal telegram payload: %w", err)
	}

	if payload.BotToken == "" || payload.ChatID == "" {
		return fmt.Errorf("missing bot_token or chat_id in telegram payload")
	}

	slog.Info("sending telegram message", "task_id", task.ID, "chat_id", payload.ChatID)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", payload.BotToken)
	
	err := requests.
		URL(apiURL).
		Method("POST").
		BodyJSON(map[string]string{
			"chat_id": payload.ChatID,
			"text":    payload.Message,
		}).
		Client(h.client).
		Fetch(ctx)

	if err != nil {
		return fmt.Errorf("telegram api call failed: %w", err)
	}

	return nil
}
