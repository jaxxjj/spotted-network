-- Tasks table stores task information
CREATE TABLE IF NOT EXISTS tasks (
    task_id TEXT PRIMARY KEY,
    chain_id INTEGER NOT NULL,
    target_address TEXT NOT NULL,
    key NUMERIC NOT NULL,
    block_number NUMERIC,
    timestamp NUMERIC,
    value NUMERIC,
    epoch INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'confirming')),
    required_confirmations INTEGER,
    current_confirmations INTEGER DEFAULT 0,
    last_checked_block NUMERIC,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add indexes for query performance
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_confirming ON tasks(status, current_confirmations) WHERE status = 'confirming'; 