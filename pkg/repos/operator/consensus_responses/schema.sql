-- Schema for consensus_responses table stores aggregated consensus results for tasks
CREATE TABLE IF NOT EXISTS consensus_responses (
    id BIGSERIAL PRIMARY KEY,                    -- Auto-incrementing unique identifier
    task_id VARCHAR(66) NOT NULL,                -- Task hash identifier
    epoch INT4 NOT NULL,                         -- Consensus epoch number
    value NUMERIC NOT NULL,                      -- Consensus value
    key NUMERIC NOT NULL,                        -- Consensus key
    total_weight NUMERIC NOT NULL,               -- Total weight of participating operators
    chain_id INT4 NOT NULL,                      -- Chain identifier
    block_number BIGINT NOT NULL,                -- Block number at consensus
    target_address VARCHAR(42) NOT NULL,         -- Target contract address
    aggregated_signatures BYTEA,                 -- Combined signatures of operators
    operator_signatures JSONB,                   -- Individual operator signatures
    consensus_reached_at TIMESTAMPTZ,            -- When consensus was achieved
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_task_consensus UNIQUE (task_id),
    CONSTRAINT unique_consensus_request UNIQUE (target_address, chain_id, block_number, key)
);

CREATE INDEX IF NOT EXISTS idx_consensus_responses_epoch ON consensus_responses(epoch);
CREATE INDEX IF NOT EXISTS idx_consensus_responses_chain_block ON consensus_responses(chain_id, block_number);
CREATE INDEX IF NOT EXISTS idx_consensus_responses_target ON consensus_responses(target_address);
CREATE INDEX IF NOT EXISTS idx_consensus_responses_created_at ON consensus_responses(created_at); 