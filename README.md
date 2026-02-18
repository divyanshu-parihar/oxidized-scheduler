# Oxidized-Scheduler

A highly scalable, production-grade distributed task scheduler built in Go, designed to handle millions of future-dated tasks with high precision and reliability.

## 🚀 Vision
The goal of **Oxidized-Scheduler** is to provide a "Redis-like" ease of use for scheduling tasks but with the persistence guarantees of PostgreSQL and the precision of Hierarchical Timing Wheels. It is designed to solve the "Hot-Path" problem where massive amounts of tasks scheduled for the same second don't overwhelm the system.

## 🏗️ Architecture (In Progress)
- **Engine:** Built on **Hierarchical Timing Wheels (HTW)** for O(1) timer management.
- **Persistence:** **PostgreSQL** backend using `SKIP LOCKED` and advisory locks for distributed coordination.
- **API:** **Gin-powered** REST interface for task submission and monitoring.
- **Reliability:** At-least-once delivery semantics with idempotent execution support.

## 🛠️ Current Progress
### Phase 1: Persistence & API (Completed ✅)
- [x] **Database Schema:** Optimized PostgreSQL schema with composite indexing on `(scheduled_at, status)`.
- [x] **Configuration:** Environment-based configuration (support for `.env`, `.env.development`, etc.).
- [x] **Migrations:** Automated migration system with both programmatic and CLI support.
- [x] **API Endpoints:**
  - `POST /events`: Schedule a new task/event.
  - `GET /events`: List recent tasks for observability.
  - `GET /health`: System health check.
- [x] **Tooling:** Development seed scripts for rapid testing.

### Phase 2: Timing Wheels & Core Engine (Current 🚧)
- [ ] In-memory Timing Wheel implementation.
- [ ] Sharded Database "Bucketing" for parallel task loading.
- [ ] Dispatcher-Executor decoupling.

## 🚦 Getting Started

### Prerequisites
- Go 1.24+
- PostgreSQL

### Setup
1. **Clone the repository:**
   ```bash
   git clone https://github.com/divyanshu-parihar/oxidized-scheduler.git
   cd oxidized-scheduler
   ```

2. **Configure environment:**
   ```bash
   cp .env.development .env
   # Edit .env with your DATABASE_URL
   ```

3. **Run Migrations:**
   ```bash
   go run cmd/migrate/main.go -cmd up
   ```

4. **Start the Server:**
   ```bash
   go run main.go
   ```

5. **Seed Tasks:**
   ```bash
   ./seed.sh
   ```

## 🧪 Testing & Performance

### Integration Tests
Run the integration tests to verify API and Database connectivity:
```bash
# Ensure DATABASE_URL is set in .env or environment
go test -v ./api/...
```

### Throughput Benchmarking
We provide a custom benchmarking tool to measure API throughput:
```bash
# Start the server first
go run main.go

# In another terminal, run the benchmarker
# -c: concurrency, -d: duration (seconds)
go run cmd/bench/main.go -c 50 -d 30
```

## 📊 API Examples

### Schedule a Task
```bash
curl -X POST http://localhost:8080/events 
  -H "Content-Type: application/json" 
  -d '{
    "task_type": "email_dispatch",
    "scheduled_time": "2026-02-15T10:00:00Z",
    "payload": {"email": "hello@example.com", "template": "welcome"},
    "max_attempts": 3
  }'
```

## 📜 Research & Principles
This project is built following the principles outlined in our **Architectural Curriculum** (`GEMINI.md`), focusing on:
- **Indexing Theory:** B-Tree fill factors for write-heavy workloads.
- **Thundering Herd Mitigation:** Using `SKIP LOCKED` for efficient task picking.
- **Distributed Consensus:** Managing bucket ownership across multiple nodes.

---
*Created as part of the Architectural Curriculum for Divyanshu Parihar.*
