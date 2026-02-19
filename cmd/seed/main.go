package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

	taskTypes := []struct {
		Name    string
		Payload func() map[string]interface{}
	}{
		{
			Name: "webhook_dispatch",
			Payload: func() map[string]interface{} {
				return map[string]interface{}{
					"url": "https://oxhmpwkwbaawptdxtjpm.supabase.co/functions/v1/bright-api",
					"headers": map[string]string{
						"Authorization": "Bearer sb_publishable_k8K7gA7PZGGFvnR9NMiC4g_HXOsxLdT",
						"apikey":        "sb_publishable_k8K7gA7PZGGFvnR9NMiC4g_HXOsxLdT",
					},
					"data": map[string]string{"name": "Functions"},
				}
			},
		},
		{
			Name: "telegram_message",
			Payload: func() map[string]interface{} {
				return map[string]interface{}{
					"bot_token": "7581789785:AAEYTUviDG4Sqs-lWChmcAbbMgWhDTKwshM",
					"chat_id":   "6501231451", // Typical format, ensuring it's set for testing
					"message":   "System Alert: High CPU Usage detected",
				}
			},
		},
	}

	const tasksPerBucket = 10
	fmt.Printf("Seeding tasks to fill EVERY wheel bucket (Seconds, Minutes, Hours)...\n")

	start := time.Now()
	rows := [][]interface{}{}

	// 1. Fill Seconds Wheel (next 60 seconds)
	for s := 0; s < 60; s++ {
		scheduledAt := time.Now().Add(time.Duration(s) * time.Second)
		for t := 0; t < tasksPerBucket; t++ {
			tt := taskTypes[rand.Intn(len(taskTypes))]
			rows = append(rows, []interface{}{
				uuid.New(), tt.Name, 1, scheduledAt, "pending", tt.Payload(), 0, 3, time.Now(), time.Now(),
			})
		}
	}

	// 2. Fill Minutes Wheel (next 60 minutes)
	for m := 1; m <= 60; m++ {
		scheduledAt := time.Now().Add(time.Duration(m) * time.Minute)
		for t := 0; t < tasksPerBucket; t++ {
			tt := taskTypes[rand.Intn(len(taskTypes))]
			rows = append(rows, []interface{}{
				uuid.New(), tt.Name, 1, scheduledAt, "pending", tt.Payload(), 0, 3, time.Now(), time.Now(),
			})
		}
	}

	// 3. Fill Hours Wheel (next 24 hours)
	for h := 1; h <= 24; h++ {
		scheduledAt := time.Now().Add(time.Duration(h) * time.Hour)
		for t := 0; t < tasksPerBucket; t++ {
			tt := taskTypes[rand.Intn(len(taskTypes))]
			rows = append(rows, []interface{}{
				uuid.New(), tt.Name, 1, scheduledAt, "pending", tt.Payload(), 0, 3, time.Now(), time.Now(),
			})
		}
	}

	// Helper to handle Payload marshalling in the loop above was tricky, refactored here:
	finalRows := [][]interface{}{}
	for _, row := range rows {
		payloadJson, _ := json.Marshal(row[5])
		row[5] = payloadJson
		finalRows = append(finalRows, row)
	}

	_, err = conn.CopyFrom(
		ctx,
		pgx.Identifier{"tasks"},
		[]string{"id", "task_type", "version", "scheduled_at", "status", "payload", "attempts", "max_attempts", "created_at", "updated_at"},
		pgx.CopyFromRows(finalRows),
	)
	if err != nil {
		log.Fatalf("Error during bulk insert: %v", err)
	}

	fmt.Printf("Seeded %d tasks across all wheel tiers in %v\n", len(finalRows), time.Since(start))
}
