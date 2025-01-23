CREATE TABLE IF NOT EXISTS operators (
    address TEXT PRIMARY KEY,
    signing_key TEXT NOT NULL,
    registered_at_block_number NUMERIC(78) NOT NULL,
    registered_at_timestamp NUMERIC(78) NOT NULL,
    active_epoch NUMERIC(78) NOT NULL,
    exit_epoch NUMERIC(78) NOT NULL DEFAULT 4294967295,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive', 'suspended')),
    weight NUMERIC(78),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for frequently accessed fields
CREATE INDEX idx_operators_status ON operators(status);
CREATE INDEX idx_operators_active_epoch ON operators(active_epoch);