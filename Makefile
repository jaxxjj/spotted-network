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

BINARY_NAME=spotted
VERSION=$(shell git describe --tags --always --dirty)
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/galxe/spotted-network/pkg/version.Version=${VERSION} \
                  -X github.com/galxe/spotted-network/pkg/version.GitCommit=${GIT_COMMIT} \
                  -X github.com/galxe/spotted-network/pkg/version.BuildTime=${BUILD_TIME}"

.PHONY: build clean run-registry run-operator stop generate-keys check-tasks create-task get-final-task start-registry get-registry-id start-operators start-monitoring test lint codecov install-lint test-infra-up test-infra-down test-infra-clean generate-bindings clean-bindings test test-signer test-signer-verbose test-signer-coverage clean-coverage build-docker

# Start monitoring infrastructure
start-prometheus:
	@docker compose up -d prometheus

start-otel:
	@docker compose up -d otel-collector

start-monitoring: start-prometheus start-otel

build:
	go build -o spotted cmd/operator/main.go
	
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

create-task-ec2:
	curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"chain_id":31337,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":8}' \
  http://18.116.97.149:8000/api/v1/tasks
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
	@go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/operator/main.go
	@echo "Build complete"

# Clean built binaries
clean:
	@echo "Cleaning up..."
	@rm -f registry bin/${BINARY_NAME}
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

.PHONY: install
install: build
	cp bin/${BINARY_NAME} ~/bin/${BINARY_NAME}

.PHONY: release
release:
	# Build for different platforms
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 cmd/operator/main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 cmd/operator/main.go
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 cmd/operator/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-arm64 cmd/operator/main.go
	
	# Create checksums
	cd bin && sha256sum ${BINARY_NAME}-* > checksums.txt

# Run all tests
test:
	@echo "Running all tests..."
	@go test -race ./...

# Run signer tests
test-signer:
	@echo "Running signer tests..."
	@go test -race ./pkg/common/crypto/signer

# Run signer tests with coverage
test-signer-cov:
	@echo "Running signer tests with coverage..."
	@go test -race -cover ./pkg/common/crypto/signer

# Run all tests with coverage
test-cov:
	@echo "Running all tests with coverage..."
	@( \
		export POSTGRES_USERNAME=$(POSTGRES_USERNAME) && \
		export POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) && \
		export POSTGRES_APPNAME=$(POSTGRES_APPNAME) && \
		export POSTGRES_HOST=$(POSTGRES_HOST) && \
		export POSTGRES_PORT=$(POSTGRES_PORT) && \
		export POSTGRES_DBNAME=$(POSTGRES_DBNAME) && \
		export POSTGRES_SSLMODE=$(POSTGRES_SSLMODE) && \
		go test -p 1 -coverprofile=coverage.out ./... && \
		go tool cover -func=coverage.out \
	)

# Build Docker image
build-operator:
	docker build -f Dockerfile.operator -t spotted-operator:latest .
	docker tag spotted-operator:latest jaxonch/spotted-operator:latest
	docker push jaxonch/spotted-operator:latest
