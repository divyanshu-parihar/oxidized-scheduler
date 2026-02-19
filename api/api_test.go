package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/divyanshu-parihar/oxidized-scheduler/internal/config"
	"github.com/divyanshu-parihar/oxidized-scheduler/internal/database"
	"github.com/divyanshu-parihar/oxidized-scheduler/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	cfg := config.LoadConfig()
	// Use the DB URL from config, but ensure migrations are run
	err := database.RunMigrations(cfg.DatabaseURL, "../migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	return db
}

func TestAddEvent(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test; DATABASE_URL not set")
	}

	db := setupTestDB(t)
	defer db.Close()

	api := NewAPI(db, nil)
	router := api.CreateServer()

	taskPayload := map[string]interface{}{
		"task_type":      "test_task",
		"scheduled_time": time.Now().Add(time.Hour).Format(time.RFC3339),
		"payload":        map[string]string{"foo": "bar"},
		"max_attempts":   5,
	}
	body, _ := json.Marshal(taskPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Event scheduled successfully", response["message"])
	assert.NotNil(t, response["task_id"])
}

func TestListEvents(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test; DATABASE_URL not set")
	}

	db := setupTestDB(t)
	defer db.Close()

	api := NewAPI(db, nil)
	router := api.CreateServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/events", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var tasks []models.Task
	err := json.Unmarshal(w.Body.Bytes(), &tasks)
	assert.NoError(t, err)
	assert.True(t, len(tasks) >= 0)
}
