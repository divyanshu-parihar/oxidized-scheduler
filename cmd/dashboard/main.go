package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/divyanshu-parihar/oxidized-scheduler/internal/config"
	"github.com/divyanshu-parihar/oxidized-scheduler/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

const dashTmpl = `
<!DOCTYPE html>
<html>
<head>
    <title>Oxidized Scheduler Dashboard</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; margin: 0; padding: 20px; background: #f4f7f6; }
        .container { max-width: 1200px; margin: 0 auto; }
        .card { background: white; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); padding: 20px; margin-bottom: 20px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { text-align: left; padding: 12px; border-bottom: 1px solid #eee; }
        th { background: #fafafa; }
        .status-pending { color: #f39c12; font-weight: bold; }
        .status-completed { color: #2ecc71; font-weight: bold; }
        .status-failed { color: #e74c3c; font-weight: bold; }
        .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
        .summary-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; margin-bottom: 20px; }
        .summary-card { text-align: center; }
        .summary-card h3 { margin-top: 0; color: #666; font-size: 0.9rem; text-transform: uppercase; }
        .summary-card .count { font-size: 2rem; font-weight: bold; }
        .wheel-container { display: flex; flex-wrap: wrap; gap: 2px; margin-top: 5px; background: #eee; padding: 5px; border-radius: 4px; }
        .bucket { width: 12px; height: 12px; background: white; border: 1px solid #ddd; border-radius: 2px; }
        .bucket.has-tasks { background: #3498db; border-color: #2980b9; }
        .bucket.active { background: #e74c3c; border-color: #c0392b; transform: scale(1.2); }
    </style>
</head>
<body>
<div class="container">
    <div class="header">
        <h1>Oxidized Scheduler Dashboard</h1>
        <div>Last Refresh: {{.LastRefresh}}</div>
    </div>

    <div class="summary-grid">
        <div class="card summary-card">
            <h3>Total Tasks</h3>
            <div class="count">{{.Stats.Total}}</div>
        </div>
        <div class="card summary-card">
            <h3>Pending</h3>
            <div class="count status-pending">{{.Stats.Pending}}</div>
        </div>
        <div class="card summary-card">
            <h3>Completed</h3>
            <div class="count status-completed">{{.Stats.Completed}}</div>
        </div>
        <div class="card summary-card">
            <h3>Failed</h3>
            <div class="count status-failed">{{.Stats.Failed}}</div>
        </div>
    </div>

    <div class="card">
        <h2>Timing Wheel Visualization</h2>
        <div style="margin-bottom: 10px;">
            <strong>Seconds Wheel:</strong> (Ptr: {{.Wheel.SecPtr}})
            <div class="wheel-container">
                {{range $i, $count := .Wheel.Seconds}}
                <div class="bucket {{if eq $i $.Wheel.SecPtr}}active{{else if gt $count 0}}has-tasks{{end}}" title="Bucket {{$i}}: {{$count}} tasks"></div>
                {{end}}
            </div>
        </div>
        <div style="margin-bottom: 10px;">
            <strong>Minutes Wheel:</strong> (Ptr: {{.Wheel.MinPtr}})
            <div class="wheel-container">
                {{range $i, $count := .Wheel.Minutes}}
                <div class="bucket {{if eq $i $.Wheel.MinPtr}}active{{else if gt $count 0}}has-tasks{{end}}" title="Bucket {{$i}}: {{$count}} tasks"></div>
                {{end}}
            </div>
        </div>
        <div>
            <strong>Hours Wheel:</strong> (Ptr: {{.Wheel.HourPtr}})
            <div class="wheel-container">
                {{range $i, $count := .Wheel.Hours}}
                <div class="bucket {{if eq $i $.Wheel.HourPtr}}active{{else if gt $count 0}}has-tasks{{end}}" title="Bucket {{$i}}: {{$count}} tasks"></div>
                {{end}}
            </div>
        </div>
    </div>

    <div class="card">
        <h2>Recent Tasks</h2>
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Type</th>
                    <th>Status</th>
                    <th>Scheduled For</th>
                    <th>Attempts</th>
                </tr>
            </thead>
            <tbody>
                {{range .Tasks}}
                <tr>
                    <td><code>{{.ID}}</code></td>
                    <td>{{.TaskType}}</td>
                    <td><span class="status-{{.Status}}">{{.Status}}</span></td>
                    <td>{{.ScheduledTime.Format "Jan 02 15:04:05"}}</td>
                    <td>{{.Attempts}}/{{.MaxAttempts}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</div>
<script>setTimeout(() => location.reload(), 2000);</script>
</body>
</html>
`

type WheelState struct {
	Seconds     [60]int `json:"seconds"`
	Minutes     [60]int `json:"minutes"`
	Hours       [24]int `json:"hours"`
	Unprocessed int     `json:"unprocessed"`
	SecPtr      int     `json:"sec_ptr"`
	MinPtr      int     `json:"min_ptr"`
	HourPtr     int     `json:"hour_ptr"`
}

type TaskStats struct {
	Total     int
	Pending   int
	Completed int
	Failed    int
}

type Dashboard struct {
	db *pgxpool.Pool
}

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}
	defer db.Close()

	d := &Dashboard{db: db}

	http.HandleFunc("/", d.handleIndex)
	slog.Info("Dashboard starting", "port", "8081")
	http.ListenAndServe(":8081", nil)
}

func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Fetch Wheel State from Scheduler
	var wheelState WheelState
	err := requests.
		URL("http://localhost:8080/wheel/state").
		ToJSON(&wheelState).
		Fetch(ctx)
	if err != nil {
		slog.Warn("failed to fetch wheel state", "error", err)
	}

	// 2. Fetch Stats from DB
	var stats TaskStats
	d.db.QueryRow(ctx, "SELECT COUNT(*) FROM tasks").Scan(&stats.Total)
	d.db.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE status = 'pending'").Scan(&stats.Pending)
	d.db.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE status = 'completed'").Scan(&stats.Completed)
	d.db.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE status = 'failed'").Scan(&stats.Failed)

	// 3. Fetch Recent Tasks from DB
	query := `SELECT id, task_type, status, scheduled_at, attempts, max_attempts FROM tasks ORDER BY created_at DESC LIMIT 20`
	rows, err := d.db.Query(ctx, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.TaskType, &t.Status, &t.ScheduledTime, &t.Attempts, &t.MaxAttempts); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	tmpl := template.Must(template.New("dashboard").Parse(dashTmpl))
	tmpl.Execute(w, struct {
		Tasks       []models.Task
		Stats       TaskStats
		Wheel       WheelState
		LastRefresh string
	}{
		Tasks:       tasks,
		Stats:       stats,
		Wheel:       wheelState,
		LastRefresh: time.Now().Format("15:04:05"),
	})
}
