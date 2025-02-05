CREATE TABLE IF NOT EXISTS task_responses (
    id BIGSERIAL PRIMARY KEY,
    task_id VARCHAR(66) NOT NULL,
    operator_address VARCHAR(42) NOT NULL,
    signature BYTEA NOT NULL,
    chain_id INT4 NOT NULL,
    target_address VARCHAR(42) NOT NULL,
    key NUMERIC NOT NULL,
    value NUMERIC NOT NULL,
    block_number BIGINT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT task_responses_operator_unique UNIQUE(task_id, operator_address)
);

CREATE INDEX IF NOT EXISTS idx_task_responses_task_id ON task_responses(task_id);
CREATE INDEX IF NOT EXISTS idx_task_responses_operator_addr ON task_responses(operator_address);

