-- Operators table stores information about registered network operators
CREATE TABLE IF NOT EXISTS operators (
    address VARCHAR(42) PRIMARY KEY,             -- Operator's Ethereum address
    signing_key VARCHAR(66) NOT NULL,            -- Operator's public key for signing
    registered_at_block_number BIGINT NOT NULL,  -- Block number when operator registered
    registered_at_timestamp BIGINT NOT NULL,     -- Timestamp when operator registered
    active_epoch BIGINT NOT NULL,                  -- Epoch when operator will (or already) become active
    exit_epoch BIGINT NOT NULL DEFAULT 4294967295, -- Epoch when operator will exit (max uint32 as default)
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive', 'suspended')),
    weight NUMERIC NOT NULL,                     -- Operator's stake weight
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_operators_status ON operators(status);
CREATE INDEX IF NOT EXISTS idx_operators_active_epoch ON operators(active_epoch);
CREATE INDEX IF NOT EXISTS idx_operators_exit_epoch ON operators(exit_epoch);
CREATE INDEX IF NOT EXISTS idx_operators_weight ON operators(weight);
CREATE INDEX IF NOT EXISTS idx_operators_updated_at ON operators(updated_at);