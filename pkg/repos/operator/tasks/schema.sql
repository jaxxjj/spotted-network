-- Tasks table stores task information
CREATE TABLE IF NOT EXISTS tasks (
    task_id VARCHAR(66) PRIMARY KEY,
    chain_id INT4 NOT NULL,
    target_address VARCHAR(42) NOT NULL,
    key NUMERIC NOT NULL,
    block_number BIGINT,
    timestamp BIGINT,
    value NUMERIC,
    epoch INT4 NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'confirming')),
    required_confirmations INT2,
    retry_count INT2 NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_task_id ON tasks(task_id);
CREATE INDEX IF NOT EXISTS idx_tasks_target_address ON tasks(target_address);
CREATE INDEX IF NOT EXISTS idx_tasks_chain_block ON tasks(chain_id, target_address, key, block_number);

