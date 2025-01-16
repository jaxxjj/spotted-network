-- Task responses table to store individual operator responses
CREATE TABLE IF NOT EXISTS task_responses (
    id BIGSERIAL PRIMARY KEY,
    task_id TEXT NOT NULL,
    operator_address TEXT NOT NULL,
    signature BYTEA NOT NULL,
    epoch INT NOT NULL,
    chain_id INT NOT NULL,
    target_address TEXT NOT NULL,
    key NUMERIC(78) NOT NULL,
    value NUMERIC(78) NOT NULL,
    block_number NUMERIC(78),
    timestamp NUMERIC(78),
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'invalid')),
    
    -- Add foreign key to tasks table
    CONSTRAINT fk_task FOREIGN KEY (task_id) REFERENCES tasks(task_id) ON DELETE CASCADE,
    
    -- Ensure one response per operator per task
    CONSTRAINT unique_operator_task UNIQUE (task_id, operator_address)
);

-- Index for querying responses by task
CREATE INDEX idx_task_responses_task_id ON task_responses(task_id);

-- Index for querying responses by operator
CREATE INDEX idx_task_responses_operator ON task_responses(operator_address);

-- Index for querying by status
CREATE INDEX idx_task_responses_status ON task_responses(status);