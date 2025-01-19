#!/bin/bash

set -e
set -u

function create_operator_schema() {
    local database=$1
    echo "Creating operator schema in database '$database'"
    
    # Create tasks table
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$database" <<-EOSQL
        CREATE TABLE IF NOT EXISTS tasks (
            task_id TEXT PRIMARY KEY,
            target_address TEXT NOT NULL,
            chain_id INT NOT NULL,
            block_number NUMERIC(78),
            timestamp NUMERIC(78),
            epoch INT NOT NULL,
            key NUMERIC(78) NOT NULL,
            value NUMERIC(78),
            retries INT DEFAULT 0,
            required_confirmations INT,          
            current_confirmations INT DEFAULT 0,  
            last_checked_block NUMERIC(78),      
            status TEXT NOT NULL CHECK (status IN ('pending', 'confirming', 'completed', 'expired', 'failed')),
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        );

        CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
        CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
        CREATE INDEX IF NOT EXISTS idx_tasks_confirming ON tasks(status, current_confirmations) WHERE status = 'confirming';

        CREATE TABLE IF NOT EXISTS task_responses (
            id BIGSERIAL PRIMARY KEY,
            task_id TEXT NOT NULL REFERENCES tasks(task_id) ON DELETE CASCADE,
            operator_address TEXT NOT NULL,
            signing_key TEXT NOT NULL,
            signature BYTEA NOT NULL,
            epoch INT NOT NULL,
            chain_id INT NOT NULL,
            target_address TEXT NOT NULL,
            key NUMERIC(78) NOT NULL,
            value NUMERIC(78) NOT NULL,
            block_number NUMERIC(78) NOT NULL,
            timestamp NUMERIC(78) NOT NULL,
            submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            UNIQUE(task_id, operator_address)
        );

        CREATE INDEX IF NOT EXISTS idx_task_responses_task_id ON task_responses(task_id);
        CREATE INDEX IF NOT EXISTS idx_task_responses_operator ON task_responses(operator_address);

        CREATE TABLE IF NOT EXISTS consensus_responses (
            task_id TEXT PRIMARY KEY,
            epoch INT NOT NULL,
            status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'failed')),
            aggregated_signatures BYTEA,
            operator_signatures JSONB,
            consensus_reached_at TIMESTAMP,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        );

        CREATE INDEX IF NOT EXISTS idx_consensus_responses_task_id ON consensus_responses(task_id);
        CREATE INDEX IF NOT EXISTS idx_consensus_responses_status ON consensus_responses(status);
EOSQL
}

function create_database() {
    local database=$1
    echo "Creating database '$database'"
    psql -v ON_ERROR_STOP=0 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
        CREATE DATABASE $database;
        GRANT ALL PRIVILEGES ON DATABASE $database TO $POSTGRES_USER;
EOSQL

    # Initialize schema for operator database
    if [[ $database == *"operator"* ]]; then
        create_operator_schema $database
    fi
}

if [ -n "$POSTGRES_MULTIPLE_DATABASES" ]; then
    echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"
    
    for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
        # Skip postgres database since it already exists
        if [ "$db" != "postgres" ]; then
            create_database $db
        fi
    done
    
    echo "Multiple databases created"
fi