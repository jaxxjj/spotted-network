-- Epoch states table stores consensus state for each epoch
CREATE TABLE IF NOT EXISTS epoch_states (
    epoch_number INT4 PRIMARY KEY,
    block_number BIGINT NOT NULL,
    minimum_weight NUMERIC NOT NULL,
    total_weight NUMERIC NOT NULL,
    threshold_weight NUMERIC NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_epoch_states_block_number ON epoch_states(block_number);
CREATE INDEX IF NOT EXISTS idx_epoch_states_updated_at ON epoch_states(updated_at); 