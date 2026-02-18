package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/divyanshu-parihar/oxidized-scheduler/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	const totalSets = 50
	const tasksPerSet = 1000
	const totalTasks = totalSets * tasksPerSet

	fmt.Printf("Seeding %d tasks in %d sets of %d (each set sharing a unique second)...\n", totalTasks, totalSets, tasksPerSet)

	start := time.Now()
	baseTime := time.Now().Add(10 * time.Second).Truncate(time.Second) // Start 10s from now

	for s := 0; s < totalSets; s++ {
		scheduledAt := baseTime.Add(time.Duration(s) * time.Second)
		rows := [][]interface{}{}

		for t := 0; t < tasksPerSet; t++ {
			id := uuid.New()
			taskType := "example_task"
			status := "pending"
			payload, _ := json.Marshal(map[string]interface{}{
				"set": s,
				"idx": t,
				"url": "http://localhost:8080/callback",
			})
			attempts := 0
			maxAttempts := 3
			now := time.Now()

			rows = append(rows, []interface{}{
				id, taskType, 1, scheduledAt, status, payload, attempts, maxAttempts, now, now,
			})
		}

		_, err := conn.CopyFrom(
			ctx,
			pgx.Identifier{"tasks"},
			[]string{"id", "task_type", "version", "scheduled_at", "status", "payload", "attempts", "max_attempts", "created_at", "updated_at"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			log.Fatalf("Error during batch insert for set %d: %v", s, err)
		}

		if (s+1)%10 == 0 {
			fmt.Printf("Seeded %d sets (%d tasks)...\n", s+1, (s+1)*tasksPerSet)
		}
	}

	duration := time.Since(start)
	fmt.Printf("Successfully seeded %d tasks in %v (avg %.2f tasks/sec)\n", totalTasks, duration, float64(totalTasks)/duration.Seconds())
}
