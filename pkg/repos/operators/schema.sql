CREATE TABLE IF NOT EXISTS operators (
    address VARCHAR(42) PRIMARY KEY,            
    signing_key VARCHAR(42) NOT NULL,           
    p2p_key VARCHAR(42) NOT NULL,             
    registered_at_block_number BIGINT NOT NULL,       
    active_epoch BIGINT NOT NULL,                 
    exit_epoch BIGINT NOT NULL DEFAULT 4294967295, 
    is_active BOOLEAN NOT NULL DEFAULT false,      
    weight NUMERIC NOT NULL,                    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_operators_address ON operators(signing_key);
CREATE INDEX IF NOT EXISTS idx_operators_p2p_key ON operators(p2p_key);
CREATE INDEX IF NOT EXISTS idx_operators_is_active ON operators(is_active);