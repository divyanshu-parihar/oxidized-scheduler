package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080/events", "API URL to test")
	concurrency := flag.Int("c", 10, "Number of concurrent workers")
	duration := flag.Int("d", 10, "Duration of test in seconds")
	flag.Parse()

	fmt.Printf("Starting throughput test: %s
", *url)
	fmt.Printf("Concurrency: %d, Duration: %ds
", *concurrency, *duration)

	payload := map[string]interface{}{
		"task_type":      "bench_task",
		"scheduled_time": time.Now().Add(time.Hour).Format(time.RFC3339),
		"payload":        map[string]string{"data": "benchmarking"},
		"max_attempts":   3,
	}
	body, _ := json.Marshal(payload)

	var successCount int64
	var errorCount int64
	var totalLatency int64 // in microseconds

	start := time.Now()
	stop := time.After(time.Duration(*duration) * time.Second)

	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{
				Timeout: 2 * time.Second,
			}
			for {
				select {
				case <-stop:
					return
				default:
					reqStart := time.Now()
					resp, err := client.Post(*url, "application/json", bytes.NewBuffer(body))
					latency := time.Since(reqStart).Microseconds()

					if err != nil || resp.StatusCode >= 400 {
						atomic.AddInt64(&errorCount, 1)
						if resp != nil {
							resp.Body.Close()
						}
					} else {
						atomic.AddInt64(&successCount, 1)
						atomic.AddInt64(&totalLatency, latency)
						resp.Body.Close()
					}
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start).Seconds()

	totalReqs := successCount + errorCount
	rps := float64(totalReqs) / elapsed
	avgLatency := float64(0)
	if successCount > 0 {
		avgLatency = float64(totalLatency) / float64(successCount) / 1000 // to ms
	}

	fmt.Println("
--- Results ---")
	fmt.Printf("Total Requests: %d
", totalReqs)
	fmt.Printf("Successful:     %d
", successCount)
	fmt.Printf("Errors:         %d
", errorCount)
	fmt.Printf("Throughput:     %.2f req/s
", rps)
	fmt.Printf("Avg Latency:    %.2f ms
", avgLatency)
	fmt.Printf("Total Time:     %.2fs
", elapsed)
}
