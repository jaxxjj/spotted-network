-- Epoch states table
CREATE TABLE IF NOT EXISTS epoch_states (
    epoch_number INT PRIMARY KEY,
    block_number NUMERIC(78) NOT NULL,
    minimum_weight NUMERIC(78) NOT NULL,
    total_weight NUMERIC(78) NOT NULL,
    threshold_weight NUMERIC(78) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_epoch_states_block_number ON epoch_states(block_number); 