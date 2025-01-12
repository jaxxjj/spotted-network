.PHONY: build clean run-registry run-operator stop

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

# Run operator node (requires registry address)
run-operator:
	@echo "Getting registry peer ID..."
	$(eval REGISTRY_ID := $(shell ./registry 2>&1 | grep "Host ID:" | cut -d' ' -f6))
	@echo "Starting operator node..."
	@./operator -registry "/ip4/127.0.0.1/tcp/9000/p2p/$(REGISTRY_ID)"

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