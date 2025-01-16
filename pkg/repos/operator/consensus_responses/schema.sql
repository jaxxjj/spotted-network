-- Schema for consensus_responses table
CREATE TABLE consensus_responses (
    id BIGSERIAL PRIMARY KEY,
    task_id TEXT NOT NULL,
    epoch INT NOT NULL,
    status TEXT NOT NULL, -- pending, completed, failed
    aggregated_signatures BYTEA,
    operator_signatures JSONB, -- stores operator signatures with their weights
    consensus_reached_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'completed', 'failed')),
    CONSTRAINT unique_task_consensus UNIQUE (task_id)
); 