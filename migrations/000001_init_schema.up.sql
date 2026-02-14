
CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    task_type VARCHAR(100) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payload JSONB NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_scheduled_status ON tasks (scheduled_at, status)
WHERE status = 'pending';
