CREATE TABLE IF NOT EXISTS consensus_responses (
    task_id VARCHAR(66) PRIMARY KEY,                
    epoch INT4 NOT NULL,
    chain_id INT4 NOT NULL,
    target_address VARCHAR(42) NOT NULL,
    block_number BIGINT NOT NULL,        
    key NUMERIC NOT NULL,                        
    value NUMERIC NOT NULL,                                                                                        
    aggregated_signatures BYTEA[] NOT NULL,
    operator_addresses VARCHAR(42)[] NOT NULL,                                        
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_consensus_request UNIQUE (target_address, chain_id, block_number, key)
);

CREATE INDEX IF NOT EXISTS idx_consensus_responses_epoch ON consensus_responses(epoch);
CREATE INDEX IF NOT EXISTS idx_consensus_responses_chain_block ON consensus_responses(chain_id, target_address, key, block_number);
