CREATE TABLE consensus_responses (
    task_id TEXT PRIMARY KEY,
    epoch INT NOT NULL,
    status TEXT NOT NULL, -- pending, completed, failed
    aggregated_signatures BYTEA,
    operator_signatures JSONB, -- stores operator signatures with their weights
    consensus_reached_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'completed', 'failed'))
); 