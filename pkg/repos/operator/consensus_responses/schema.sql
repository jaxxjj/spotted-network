-- Schema for consensus_responses table
CREATE TABLE consensus_responses (
    id BIGSERIAL PRIMARY KEY,
    task_id TEXT NOT NULL,
    epoch INT NOT NULL,
    value NUMERIC(78) NOT NULL, -- 添加value字段，记录最终共识的值
    block_number NUMERIC(78) NOT NULL, -- 添加block_number
    chain_id INT NOT NULL, -- 添加chain_id 
    target_address TEXT NOT NULL, -- 添加target_address
    key NUMERIC(78) NOT NULL, -- 添加key
    aggregated_signatures BYTEA,
    operator_signatures JSONB, -- {operator_address: {signature: bytes, weight: string}}
    total_weight NUMERIC(78) NOT NULL, -- 添加total_weight字段
    consensus_reached_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_task_consensus UNIQUE (task_id),
    CONSTRAINT fk_task FOREIGN KEY (task_id) REFERENCES tasks(task_id) ON DELETE CASCADE
);

CREATE INDEX idx_consensus_epoch ON consensus_responses(epoch); 