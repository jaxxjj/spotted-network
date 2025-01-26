# Spotted Network

Spotted oracle network implementation with a registry node and multiple operator nodes.

## Overview

The Spotted Network consists of a central registry node that manages operator registration and state synchronization, and multiple operator nodes that process state query tasks and reach consensus on results.

## Architecture

### Components

- **Registry Node**: Central coordinator that manages operator registration and maintains network state
- **Operator Nodes**: Process state query tasks and participate in consensus (3 test operators included), then allow user to query the final result
- **Mock Contracts**: 
  - MockRegistry: Handles operator registration
  - MockEpochManager: Manages epoch transitions
  - MockStateManager: Manages state queries and updates

### Database Structure

Each node has its own PostgreSQL database:
- Registry DB: Stores operator information and network state
- Operator DBs: Store tasks, responses, and consensus data

## Setup & Installation

### Prerequisites

- Go 1.23
- Docker & Docker Compose
- Anvil 

### Installation Steps

1. Clone the repository:
```bash
git clone https://github.com/your-org/spotted-network.git
cd spotted-network
```

2. Start Anvil local node:
```bash
anvil
```

3. Deploy mock contracts:
```bash
cd mock
make deploy-all
```
This will:
- Deploy MockRegistry, MockEpochManager, and MockStateManager contracts
- Emit registration events for 3 test operators

4. Start the network:
```bash
docker-compose up --build
```

## Testing

The repository includes 3 dummy test operator nodes with keys stored in the `keys` folder.

### Test Flow

1. Ensure Anvil is running
2. Deploy mock contracts and register operators
3. Start the network with docker-compose
4. Use make commands to interact with the network

### Available Make Commands

#### Check Database Status
```bash
# Check operator status
make check-operator-status-operator1
make check-operator-status-operator2
make check-operator-status-operator3

# Check tasks
make check-tasks-operator1
make check-tasks-operator2
make check-tasks-operator3

# Check task responses
make check-task-responses-operator1
make check-task-responses-operator2
make check-task-responses-operator3

# Check consensus
make check-consensus-operator1
make check-consensus-operator2
make check-consensus-operator3
```

#### Create and Process Tasks
```bash
# Create a new task on any operator
make create-task-operator1

# Mine 15 blocks to simulate block progression
make mine-15

# Get final task result (replace with actual task ID)
curl -X GET "http://localhost:8001/api/v1/task/YOUR_TASK_ID/final"
```

## Network Configuration

### Ports

- Registry Node: 
  - P2P: 9000
  - HTTP: 8000
- Operator 1:
  - P2P: 10000
  - HTTP: 8001
- Operator 2:
  - P2P: 10001
  - HTTP: 8002
- Operator 3:
  - P2P: 10002
  - HTTP: 8003

### Databases

- Registry: Port 5432
- Operator 1: Port 5433
- Operator 2: Port 5434
- Operator 3: Port 5435

## Development

### Project Structure

```
.
├── cmd/
│   ├── registry/        # Registry node entry point
│   └── operator/        # Operator node entry point
├── pkg/
│   ├── common/          # Shared utilities and types
│   ├── operator/        # Operator node implementation
│   └── registry/        # Registry node implementation
├── mock/               # Mock contracts for testing
├── keys/               # Test operator keys
└── docker-compose.yml  # Network deployment configuration
```

### Key Features

- P2P communication between nodes
- Task processing and consensus
- State synchronization
- Epoch-based operator management
- Database persistence per node

## Notes

- The test operator keys are located in the `keys` folder
- When getting final task results, replace the task ID in the curl command with the actual task ID from the create-task response
- The network uses PostgreSQL for persistence, with separate databases for each node
- Mock contracts are used for testing and simulate the actual network contracts
