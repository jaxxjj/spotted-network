.PHONY: build clean run-registry run-operator stop generate-keys

# Generate operator keys
generate-keys:
	@echo "Generating operator keys..."
	@go run scripts/generate_keys.go
	@echo "Keys generated successfully in ./keys directory"

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

# Stop all running nodes
stop:
	@echo "Stopping all nodes..."
	@pkill -f "./registry" || true
	@pkill -f "./operator" || true
	@echo "All nodes stopped" 