# Distributed Task Scheduler: Architectural Curriculum

Welcome, Divyanshu. This document serves as your research roadmap for building **Oxidized-Scheduler** (or your chosen name) in Go. We will focus on the "why" and the "how" from a first-principles perspective.

---

## Phase 1: Persistence & The "Hot-Path" Problem
To handle millions of tasks, your database isn't just a storage bin; it's a high-throughput priority queue.

### Core Concepts to Master:
*   **Indexing Theory:** Understand why a composite index on `(execution_time, status)` is superior to two separate indexes. Research **B-Tree fill factors** and how they impact write-heavy workloads.
*   **Database Partitioning:** Study **Declarative Partitioning** in PostgreSQL. Why would you partition by `execution_time` (range) vs. `bucket_id` (list)?
*   **The Thundering Herd:** Research what happens when 1,000 nodes all run `SELECT ... WHERE status = 'PENDING'` at the same millisecond. 
    *   *Keywords:* `SKIP LOCKED`, Advisory Locks.

### Research Exercise:
Compare the performance of a Postgres-based queue vs. a specialized LSM-tree based engine for "future-dated" events.

---

## Phase 2: Timing Wheels (The Heart of the Engine)
Operating system kernels use Timing Wheels to manage millions of TCP timeouts. You will use them to manage tasks.

### Core Concepts to Master:
*   **Hierarchical Timing Wheels (HTW):** Read the seminal paper: *"Hashed and Hierarchical Timing Wheels: Data Structures for the Efficient Implementation of a Timer Facility"* by George Varghese and Tony Lauck (1987).
*   **Precision vs. Overhead:** In Go, `time.Timer` is backed by a 4-ary heap. Why does this fail at $10^7$ tasks?
*   **The "Tick" Mechanism:** Explore the trade-offs between a "Busy-Wait" tick vs. a `sleep` based tick.

### Research Exercise:
Draft a diagram of how a task "cascades" from a "Day" wheel down to a "Minute" wheel and finally to the "Millisecond" wheel.

---

## Phase 3: Distributed Coordination (The Brain)
How do nodes decide who owns which task without a single "Leader" becoming a bottleneck?

### Core Concepts to Master:
*   **Consistent Hashing:** Study the **Karger et al. (1997)** paper. How does adding a node impact task reassignment?
*   **Distributed Consensus vs. Shared State:** Compare **Raft/Paxos** (High Consistency) vs. **Postgres-backed Heartbeats** (Eventual Consistency with Shared DB).
*   **The Split-Brain Problem:** What happens when a network partition occurs and two nodes think they own the same "Bucket"?

### Research Exercise:
Investigate **Fencing Tokens**. How can you use a database-incremented version number to prevent a "zombie" node from executing a task?

---

## Phase 4: Reliability Semantics
Building a system that "mostly works" is easy. Building one that guarantees at-least-once delivery is the PhD-level challenge.

### Core Concepts to Master:
*   **Idempotency Keys:** How do you design a system where a task can be safely executed twice?
*   **The "Ack" Pattern:** Research the two-phase status transition (`PENDING` -> `DISPATCHED` -> `COMPLETED`).
*   **The Reaper/Janitor Process:** How to handle "zombie" tasks stuck in the `DISPATCHED` state after a worker crash.

---

## Phase 5: Push-Based Scaling & Transport
Scaling workers up and down requires a decouple-first approach between the Dispatcher and the Executor.

### Core Concepts to Master:
*   **Transport Layers:** Research **NATS Queue Groups** for load balancing. Why is a message broker better for scaling than direct gRPC calls?
*   **The Distributed Dispatcher:** How to shard database "buckets" across multiple dispatcher nodes to avoid a Single Point of Failure (SPOF).
*   **Auto-scaling Metrics:** Defining **Scheduling Lag** (`ActualTime - TargetTime`) as the primary trigger for scaling.

---

## Phase 6: Production Task Modeling
Your scheduler must handle heterogeneous workloads with different resource requirements.

### Task Types to Implement:
1.  **Webhook Dispatcher:** High-concurrency, I/O bound (Tests your networking stack).
2.  **Delayed Notifications:** Long-horizon, high reliability (Tests your persistence).
3.  **Report Generator:** Resource intensive, CPU/RAM bound (Tests your **Bulkhead** isolation).
4.  **System Cleanup:** Periodic/Cron-like (Tests **Overlapping Execution** logic).

---

## Phase 7: Production-Grade Observability
In a distributed system, logs are useless without context.

### Core Concepts to Master:
*   **Structured Logging (Slog/Zap):** Why is JSON logging mandatory for high-volume systems?
*   **Semantic Monitoring:** Don't just track "CPU Usage." Track "Scheduling Drift"—the delta between $	ext{TargetTime}$ and $	ext{ActualExecutionTime}$.
*   **Distributed Tracing:** Look into **OpenTelemetry**. How do you trace a task from the API request that created it to the worker that executed it 3 days later?

---

## Recommended Reading List
1.  **Varghese, G., & Lauck, T. (1987).** *Hashed and Hierarchical Timing Wheels.*
2.  **Karger, D., et al. (1997).** *Consistent Hashing and Random Trees.*
3.  **Kleppmann, M.** *Designing Data-Intensive Applications* (Chapters 5, 7, and 9).

---

**Next Step:** Once you have internalized the HTW paper, let's discuss how you plan to map PostgreSQL rows into your in-memory Timing Wheel slots.
