CREATE TABLE IF NOT EXISTS epoch_states (
    epoch_number INT4 PRIMARY KEY,
    minimum_weight NUMERIC NOT NULL,
    total_weight NUMERIC NOT NULL,
    threshold_weight NUMERIC NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_epoch_states_epoch_number ON epoch_states(epoch_number);

