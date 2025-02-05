# Project variables
NAME=spotted-network
TEST_DIRS := $(shell go list ./...)
POSTGRES_USERNAME=spotted
POSTGRES_PASSWORD=spotted
POSTGRES_APPNAME=registry_test
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DBNAME=spotted

.PHONY: build clean run-registry run-operator stop generate-keys check-tasks create-task get-final-task start-registry get-registry-id start-operators start-all start-monitoring test lint codecov install-lint test-infra-up test-infra-down test-infra-clean

# Install golangci-lint
install-lint:
	@echo "Installing golangci-lint..."
	@if [ "$(shell uname)" = "Darwin" ]; then \
		brew install golangci-lint; \
	else \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2; \
	fi

# Start monitoring infrastructure
start-prometheus:
	@docker compose up -d prometheus

start-otel:
	@docker compose up -d otel-collector

start-monitoring: start-prometheus start-otel


# Start registry node
start-registry: 
	@docker compose up -d registry 


# Start operator nodes
start-operators: 
	@docker compose --profile operators up -d operator1 operator2 operator3 --build

# Start everything in sequence
start-all: start-monitoring start-registry get-registry-id start-operators
	@echo "All services started"
	@echo "Prometheus UI: http://localhost:9090"
	@echo "Registry metrics: http://localhost:4014/metrics"
	@echo "Operator1 metrics: http://localhost:4015/metrics"
	@echo "Operator2 metrics: http://localhost:4016/metrics"
	@echo "Operator3 metrics: http://localhost:4017/metrics"

# Check monitoring status
check-monitoring:
	@echo "Checking Prometheus status..."
	@curl -s http://localhost:9090/-/healthy || echo "Prometheus is not healthy"
	@echo "\nChecking OpenTelemetry Collector status..."
	@curl -s http://localhost:8888/metrics > /dev/null && echo "OpenTelemetry Collector is healthy" || echo "OpenTelemetry Collector is not healthy"

# Generate operator keys
generate-keys:
	@echo "Generating operator keys..."
	@go run scripts/generate_keys.go
	@echo "Keys generated successfully in ./keys directory"

# Check operator status table
check-operator-status-operator1:
	@echo "Querying operator status from registry database..."
	@PGPASSWORD=spotted psql -h localhost -p 5432 -U spotted -d spotted -c "SELECT * FROM operators;"

check-operator-status-operator2:
	@echo "Querying operator2 status..."
	@PGPASSWORD=spotted psql -h localhost -p 5434 -U spotted -d operator2 -c "SELECT * FROM operators;"

check-operator-status-operator3:
	@echo "Querying operator3 status..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM operators;"

# Check tasks table
check-tasks-operator1:
	@echo "Querying tasks from operator1 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5433 -U spotted -d operator1 -c "SELECT * FROM tasks;"

check-tasks-operator2:
	@echo "Querying tasks from operator2 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5434 -U spotted -d operator2 -c "SELECT * FROM tasks;"

check-tasks-operator3:
	@echo "Querying tasks from operator3 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM tasks;"

# Check task responses table
check-task-responses-operator1:
	@echo "Querying task responses from operator1 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5433 -U spotted -d operator1 -c "SELECT * FROM task_responses;"

check-task-responses-operator2:
	@echo "Querying task responses from operator2 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5434 -U spotted -d operator2 -c "SELECT * FROM task_responses;"

check-task-responses-operator3:
	@echo "Querying task responses from operator3 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM task_responses;"

# Check consensus responses table
check-consensus-operator1:
	@echo "Querying consensus responses from operator1 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5433 -U spotted -d operator1 -c "SELECT * FROM consensus_responses;"

check-consensus-operator2:
	@echo "Querying consensus responses from operator2 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5434 -U spotted -d operator2 -c "SELECT * FROM consensus_responses;"

check-consensus-operator3:
	@echo "Querying consensus responses from operator3 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM consensus_responses;"

# Create a new sample task
create-task-operator1:
	@echo "Creating new task..."
	@curl -X POST -H "Content-Type: application/json" -d '{"chain_id":31337,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":8}' http://localhost:8001/api/v1/tasks

create-task-operator2:
	@echo "Creating new task..."
	@curl -X POST -H "Content-Type: application/json" -d '{"chain_id":31337,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":8}' http://localhost:8002/api/v1/tasks

create-task-operator3:
	@echo "Creating new task..."
	@curl -X POST -H "Content-Type: application/json" -d '{"chain_id":31337,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":8}' http://localhost:8003/api/v1/tasks

# Mine 15 blocks
mine-15:
	@echo "Mining 15 blocks..."
	@curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"anvil_mine","params":["0xF"],"id":1}' http://localhost:8545

# Get final task
get-final-response-taskId:
	@curl -X GET "http://localhost:8001/api/v1/consensus/tasks/fb562be126cced839de0da912247c4eaf591a594726e6e7672d4555df68d49ce"

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
	docker-compose up -d postgres redis

# 关闭测试基础设施
test-infra-down:
	docker-compose stop postgres redis
	docker-compose rm -f postgres redis

# 清理测试基础设施（包括数据）
test-infra-clean: test-infra-down
	docker volume rm -f spotted-network_postgres_data

