-- Schema for consensus_responses table
CREATE TABLE IF NOT EXISTS consensus_responses (
    id BIGSERIAL PRIMARY KEY,
    task_id TEXT NOT NULL,
    epoch INT NOT NULL,
    value NUMERIC(78) NOT NULL, 
    block_number NUMERIC(78) NOT NULL,
    chain_id INT NOT NULL,
    target_address TEXT NOT NULL,
    key NUMERIC(78) NOT NULL,
    aggregated_signatures BYTEA,
    operator_signatures JSONB,
    total_weight NUMERIC(78) NOT NULL,
    consensus_reached_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_task_consensus UNIQUE (task_id)
);

CREATE INDEX IF NOT EXISTS idx_consensus_epoch ON consensus_responses(epoch); 