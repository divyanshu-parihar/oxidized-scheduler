package wheel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/divyanshu-parihar/oxidized-scheduler/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWheelFlow(t *testing.T) {
	// 1. Mock API Server
	var mu sync.Mutex
	executed := false
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/events" {
			task := models.Task{
				ID:            uuid.New(),
				TaskType:      "webhook_dispatch",
				Status:        models.TASK_PENDING,
				ScheduledTime: time.Now().Add(2 * time.Second),
				Payload:       json.RawMessage(fmt.Sprintf(`{"url":"http://%s/callback","data":{}}`, r.Host)),
			}
			json.NewEncoder(w).Encode([]models.Task{task})
		} else if r.Method == "POST" && r.URL.Path == "/callback" {
			mu.Lock()
			executed = true
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
		} else if r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/status") {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// 2. Setup Wheel
	// Note: In a real test, you'd use a real DB or a mock pool. 
	// For this flow test, we focus on the Wheel logic.
	wheel := NewWheel(server.URL)
	wheel.tick = 100 * time.Millisecond // Speed up ticks for testing
	
	// 3. Manually trigger hydration
	wheel.FetchTasks()
	
	// 4. Run ticks manually to simulate time passage
	for i := 0; i < 30; i++ {
		wheel.processTick()
		time.Sleep(10 * time.Millisecond)
		
		mu.Lock()
		if executed {
			mu.Unlock()
			break
		}
		mu.Unlock()
	}

	assert.True(t, executed, "Task should have been executed by the wheel")
}
