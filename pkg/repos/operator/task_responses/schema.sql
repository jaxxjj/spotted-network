-- Task responses table to store individual operator responses
CREATE TABLE IF NOT EXISTS task_responses (
    id BIGSERIAL PRIMARY KEY,
    task_id TEXT NOT NULL,
    operator_address TEXT NOT NULL,
    signing_key TEXT NOT NULL,
    signature BYTEA NOT NULL,
    epoch INTEGER NOT NULL,
    chain_id INTEGER NOT NULL,
    target_address TEXT NOT NULL,
    key NUMERIC NOT NULL,
    value NUMERIC NOT NULL,
    block_number NUMERIC NOT NULL,
    timestamp NUMERIC NOT NULL,
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(task_id, operator_address)
);

-- Index for querying responses by task
CREATE INDEX IF NOT EXISTS idx_task_responses_task_id ON task_responses(task_id);

-- Index for querying responses by operator
CREATE INDEX IF NOT EXISTS idx_task_responses_operator ON task_responses(operator_address);

-- Index for querying by status
CREATE INDEX idx_task_responses_status ON task_responses(status);
