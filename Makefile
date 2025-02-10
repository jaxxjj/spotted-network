# Project variables
NAME=spotted-network
TEST_DIRS := $(shell go list ./...)
POSTGRES_USERNAME=spotted
POSTGRES_PASSWORD=spotted
POSTGRES_APPNAME=operator_test
POSTGRES_HOST=localhost
POSTGRES_PORT=5435
POSTGRES_DBNAME=spotted
POSTGRES_SSLMODE=disable

.PHONY: build clean run-registry run-operator stop generate-keys check-tasks create-task get-final-task start-registry get-registry-id start-operators start-monitoring test lint codecov install-lint test-infra-up test-infra-down test-infra-clean generate-bindings clean-bindings

# Start monitoring infrastructure
start-prometheus:
	@docker compose up -d prometheus

start-otel:
	@docker compose up -d otel-collector

start-monitoring: start-prometheus start-otel


# Generate operator keys
generate-keys:
	@echo "Generating operator keys..."
	@go run scripts/generate_keys.go
	@echo "Keys generated successfully in ./keys directory"

# Generate keys
generate-ecdsa-keys:
	@echo "Generating ECDSA operator keys..."
	@go run scripts/gen_keys/generate_keys.go -type ecdsa
	@echo "ECDSA keys generated successfully in ./keys directory"

generate-ed25519-keys:
	@echo "Generating Ed25519 operator keys..."
	@go run scripts/gen_keys/generate_keys.go -type ed25519
	@echo "Ed25519 keys generated successfully in ./keys directory"

generate-all-keys: generate-ecdsa-keys generate-ed25519-keys generate-p2p-keys
	@echo "All keys generated successfully"
# Check operator status table
check-operator-status-operator1:
	docker-compose exec postgres_operator1 psql -U spotted -d operator1 -c "SELECT * FROM operators;"

check-operator-status-operator2:
	docker-compose exec postgres_operator2 psql -U spotted -d operator2 -c "SELECT * FROM operators;"

check-operator-status-operator3:
	docker-compose exec postgres_operator3 psql -U spotted -d operator3 -c "SELECT * FROM operators;"

# Check tasks table
check-tasks-operator1:
	docker-compose exec postgres_operator1 psql -U spotted -d operator1 -c "SELECT * FROM tasks;"

check-tasks-operator2:
	docker-compose exec postgres_operator2 psql -U spotted -d operator2 -c "SELECT * FROM tasks;"

check-tasks-operator3:
	docker-compose exec postgres_operator3 psql -U spotted -d operator3 -c "SELECT * FROM tasks;"

	@echo "Querying consensus responses from operator3 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM consensus_responses;"

# Create a new sample task
create-task-operator1:
	@echo "Creating new task..."
	@curl -X POST -H "Content-Type: application/json" -d '{"chain_id":31337,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":8}' http://localhost:8000/api/v1/tasks

create-task-operator2:
	@echo "Creating new task..."
	@curl -X POST -H "Content-Type: application/json" -d '{"chain_id":31337,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":8}' http://localhost:8001/api/v1/tasks

create-task-operator3:
	@echo "Creating new task..."
	@curl -X POST -H "Content-Type: application/json" -d '{"chain_id":31337,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":8}' http://localhost:8002/api/v1/tasks

# Mine 15 blocks
mine-15:
	@echo "Mining 15 blocks..."
	@curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"anvil_mine","params":["0xF"],"id":1}' http://localhost:8545

# Get final task
get-consensus-response-operator1:
	@curl -X GET "http://localhost:8000/api/v1/consensus/tasks/ecd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6"

get-consensus-response-operator2:
	@curl -X GET "http://localhost:8001/api/v1/consensus/tasks/ecd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6"

get-consensus-response-operator3:
	@curl -X GET "http://localhost:8002/api/v1/consensus/tasks/ecd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6"
# Build both binaries
build:
	@echo "Building registry and operator..."
	@go build -o registry ./cmd/registry
	@go build -o operator ./cmd/operator
	@echo "Build complete"

# Clean built binaries
clean:
	@echo "Cleaning up..."
	@rm -f registry operator
	@echo "Clean complete"

# docker clean
docker-clean:
	@rm -rf ~/Library/Containers/com.docker.docker/Data/*

# Stop all running nodes
stop:
	@echo "Stopping all services..."
	@docker compose down
	@echo "All services stopped"

# Restart all services
restart: stop start-all
	@echo "All services restarted"

# Run tests
test-registry: 
	export POSTGRES_USERNAME=$(POSTGRES_USERNAME) && \
	export POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) && \
	export POSTGRES_APPNAME=$(POSTGRES_APPNAME) && \
	export POSTGRES_HOST=$(POSTGRES_HOST) && \
	export POSTGRES_PORT=$(POSTGRES_PORT) && \
	export POSTGRES_DBNAME=$(POSTGRES_DBNAME) && \
	go test ./pkg/registry -v  


test-operator: 
	export POSTGRES_USERNAME=$(POSTGRES_USERNAME) && \
	export POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) && \
	export POSTGRES_APPNAME=$(POSTGRES_APPNAME) && \
	export POSTGRES_HOST=$(POSTGRES_HOST) && \
	export POSTGRES_PORT=$(POSTGRES_PORT) && \
	export POSTGRES_DBNAME=$(POSTGRES_DBNAME) && \
	go test ./pkg/operator -v  

registry-cov:
	export POSTGRES_USERNAME=$(POSTGRES_USERNAME) && \
	export POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) && \
	export POSTGRES_APPNAME=$(POSTGRES_APPNAME) && \
	export POSTGRES_HOST=$(POSTGRES_HOST) && \
	export POSTGRES_PORT=$(POSTGRES_PORT) && \
	export POSTGRES_DBNAME=$(POSTGRES_DBNAME) && \
	go test -cover ./pkg/registry

# Run tests with coverage
codecov:
	@echo "Running tests with coverage..."
	@export POSTGRES_USERNAME=$(POSTGRES_USERNAME) && \
	export POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) && \
	export POSTGRES_APPNAME=$(POSTGRES_APPNAME) && \
	export POSTGRES_HOST=$(POSTGRES_HOST) && \
	export POSTGRES_PORT=$(POSTGRES_PORT) && \
	export POSTGRES_DBNAME=$(POSTGRES_DBNAME) && \
	go test $(TEST_DIRS) -coverprofile=coverage.txt -covermode=atomic -p 1

# Run linter (with automatic installation if needed)
lint: 
	golangci-lint run

# Run linter with auto-fix (with automatic installation if needed)
lint-fix:
	golangci-lint run --fix

# 启动测试所需的基础设施
test-infra-up:
	docker-compose up -d postgres_test redis

# 关闭测试基础设施
test-infra-down:
	docker-compose stop postgres redis
	docker-compose rm -f postgres redis

# Generate contract bindings
generate-bindings: clean-bindings
	@echo "Creating bindings directory..."
	@mkdir -p pkg/common/contracts/bindings
	
	@echo "Generating ECDSA Stake Registry bindings..."
	@abigen --abi abi/ecdsa_stake_registry.json --pkg bindings --type ECDSAStakeRegistry --out pkg/common/contracts/bindings/ecdsa_stake_registry.go
	
	@echo "Generating Epoch Manager bindings..."
	@abigen --abi abi/epoch_manager.json --pkg bindings --type EpochManager --out pkg/common/contracts/bindings/epoch_manager.go
	
	@echo "Generating State Manager bindings..."
	@abigen --abi abi/state_manager.json --pkg bindings --type StateManager --out pkg/common/contracts/bindings/state_manager.go
	
	@echo "All contract bindings generated successfully"

# Clean old bindings
clean-bindings:
	@echo "Cleaning old bindings..."
	@rm -f pkg/common/contracts/bindings/*.go

# Build operators
build-operator1:
	@echo "Building operator1..."
	@docker compose build operator1

build-operator2:
	@echo "Building operator2..."
	@docker compose build operator2

build-operator3:
	@echo "Building operator3..."
	@docker compose build operator3

# Build all operators
build-operators: build-operator1 build-operator2 build-operator3
	@echo "All operators built"

# Start individual operators with build
start-operator1: build-operator1
	@echo "Starting operator1..."
	@docker compose up -d postgres_operator1 redis
	@sleep 5  # Wait for dependencies
	@docker compose --profile operators up -d operator1
	@echo "Operator1 started"

start-operator2: build-operator2
	@echo "Starting operator2..."
	@docker compose up -d postgres_operator2 redis
	@sleep 5  # Wait for dependencies
	@docker compose --profile operators up -d operator2
	@echo "Operator2 started"

start-operator3: build-operator3
	@echo "Starting operator3..."
	@docker compose up -d postgres_operator3 redis
	@sleep 5  # Wait for dependencies
	@docker compose --profile operators up -d operator3
	@echo "Operator3 started"

# Start all operators in sequence with build
start-operators: build-operators start-operator1 start-operator2 start-operator3
	@echo "All operators started"

check-operator1:
	docker exec -it spotted-network-postgres_operator2-1 psql -U spotted -d operator2 -c "SELECT * FROM operators WHERE LOWER(p2p_key) = LOWER('0x310c8425b620980dcfcf756e46572bb6ac80eb07');"

# Define test command template
define test_cmd
	@( \
		export POSTGRES_USERNAME=$(POSTGRES_USERNAME) && \
		export POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) && \
		export POSTGRES_APPNAME=$(POSTGRES_APPNAME) && \
		export POSTGRES_HOST=$(POSTGRES_HOST) && \
		export POSTGRES_PORT=$(POSTGRES_PORT) && \
		export POSTGRES_DBNAME=$(POSTGRES_DBNAME) && \
		export POSTGRES_SSLMODE=$(POSTGRES_SSLMODE) && \
		go test $(2) ./pkg/operator/$(1)/... \
	)
endef

# Test targets with package name parameter
test-%:
	$(call test_cmd,$*,-v)

# Coverage test targets with package name parameter
testcov-%:
	$(call test_cmd,$(subst -cov,,$*),-cover)