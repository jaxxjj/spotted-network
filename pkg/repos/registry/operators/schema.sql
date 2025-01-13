CREATE TABLE IF NOT EXISTS operators (
    address TEXT PRIMARY KEY,
    signing_key TEXT NOT NULL,
    registered_at_block_number NUMERIC(78) NOT NULL,
    registered_at_timestamp NUMERIC(78) NOT NULL,
    active_epoch INT NOT NULL,
    exit_epoch INT,
    status TEXT NOT NULL CHECK (status IN ('waitingJoin', 'waitingActive', 'active', 'unactive', 'suspended')),
    weight NUMERIC(78),
    missing INT DEFAULT 0,
    successful_response_count INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for frequently accessed fields
CREATE INDEX idx_operators_status ON operators(status);
CREATE INDEX idx_operators_active_epoch ON operators(active_epoch); 