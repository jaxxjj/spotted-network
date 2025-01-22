.PHONY: build clean run-registry run-operator stop generate-keys check-tasks create-task get-final-task

# Generate operator keys
generate-keys:
	@echo "Generating operator keys..."
	@go run scripts/generate_keys.go
	@echo "Keys generated successfully in ./keys directory"

check-operator-status-operator1:
	@echo "Querying operator status from registry database..."
	@PGPASSWORD=spotted psql -h localhost -p 5432 -U spotted -d spotted -c "SELECT * FROM operators;"

check-operator-status-operator2:
	@echo "Querying operator2 status..."
	@PGPASSWORD=spotted psql -h localhost -p 5434 -U spotted -d operator2 -c "SELECT * FROM operators;"

check-operator-status-operator3:
	@echo "Querying operator3 status..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM operators;"

# Check operator1 tasks
check-tasks-operator1:
	@echo "Querying tasks from operator1 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5433 -U spotted -d operator1 -c "SELECT * FROM tasks;"
check-tasks-operator2:
	@echo "Querying tasks from operator2 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5434 -U spotted -d operator2 -c "SELECT * FROM tasks;"
check-tasks-operator3:
	@echo "Querying tasks from operator3 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM tasks;"

check-task-responses-operator1:
	@echo "Querying task responses from operator1 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5433 -U spotted -d operator1 -c "SELECT * FROM task_responses;"

check-task-responses-operator2:
	@echo "Querying task responses from operator2 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5434 -U spotted -d operator2 -c "SELECT * FROM task_responses;"

check-task-responses-operator3:
	@echo "Querying task responses from operator3 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM task_responses;"
check-consensus-operator1:
	@echo "Querying consensus responses from operator1 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5433 -U spotted -d operator1 -c "SELECT * FROM consensus_responses;"

check-consensus-operator2:
	@echo "Querying consensus responses from operator2 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5434 -U spotted -d operator2 -c "SELECT * FROM consensus_responses;"

check-consensus-operator3:
	@echo "Querying consensus responses from operator3 database..."
	@PGPASSWORD=spotted psql -h localhost -p 5435 -U spotted -d operator3 -c "SELECT * FROM consensus_responses;"

# Create new task
create-task:
	@echo "Creating new task..."
	@curl -X POST -H "Content-Type: application/json" -d '{"chain_id":31337,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":8}' http://localhost:8001/api/v1/task

# Mine 15 blocks
mine-15:
	@echo "Mining 15 blocks..."
	@curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"anvil_mine","params":["0xF"],"id":1}' http://localhost:8545

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

# Run registry node
run-registry:
	@echo "Starting registry node..."
	@./registry

# Run registry nodes
run-registry1:
	@echo "Starting registry node 1..."
	@./registry -port 9000

run-registry2:
	@echo "Starting registry node 2..."
	@./registry -port 9001

# Run operator nodes (requires registry address)
run-operator1:
	@echo "Starting operator node 1..."
	@./operator -registry "/ip4/127.0.0.1/tcp/9000/p2p/12D3KooWAiHjQh86GPwa3yHakinreSYzFKop2kCW5853zk7kLFpx" -port 10000

run-operator2:
	@echo "Starting operator node 2..."
	@./operator -registry "/ip4/127.0.0.1/tcp/9000/p2p/12D3KooWAiHjQh86GPwa3yHakinreSYzFKop2kCW5853zk7kLFpx" -port 10001

run-operator3:
	@echo "Starting operator node 3..."
	@./operator -registry "/ip4/127.0.0.1/tcp/9000/p2p/12D3KooWAiHjQh86GPwa3yHakinreSYzFKop2kCW5853zk7kLFpx" -port 10002

# Build and run all (in separate terminals)
run-all: build
	@echo "Starting all nodes..."
	@osascript -e 'tell app "Terminal" to do script "cd $(PWD) && make run-registry"'
	@sleep 2
	@osascript -e 'tell app "Terminal" to do script "cd $(PWD) && make run-operator"'

# docker clean
docker-clean:
	@rm -rf ~/Library/Containers/com.docker.docker/Data/*

# Stop all running nodes
stop:
	@echo "Stopping all nodes..."
	@pkill -f "./registry" || true
	@pkill -f "./operator" || true
	@echo "All nodes stopped"

# Get final task
get-final-task:
	@curl -X GET "http://localhost:8001/api/v1/task/827c41edd51a9ad4da0ce4218eb42c7f62c09563d74123214493a069655934fb/final" 