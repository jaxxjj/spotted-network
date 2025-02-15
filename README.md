# Spotted Network

A decentralized network for state verification and consensus formation.

## Overview

The Spotted Network consists of two main components:
- API Server: RESTful service for task creation and consensus querying
- Operator P2P Node: Decentralized network node for consensus formation

## Requirements

### Software Requirements
- Docker: [Installation Guide](https://docs.docker.com/get-docker/)
- Docker Compose: [Installation Guide](https://docs.docker.com/compose/install/)
- Go: version 1.23.3 or higher (for building from source)

### Verify Installation
```bash
# Check Docker
docker --version

# Check Docker Compose
docker compose version

# Check Go
go version
```

## Installation

### Binary Installation
```bash
# Download and install binary
curl -sSL https://raw.githubusercontent.com/jaxxjj/spotted-network/main/scripts/install.sh | bash

# Add to PATH
export PATH=$PATH:~/bin

# Initialize
spotted init
```

### Build from Source
```bash
# Clone repository
git clone https://github.com/jaxxjj/spotted-network
cd spotted-network

# Option 1: Using go build
go build -o spotted cmd/operator/main.go

# Option 2: Using make
make build

# Initialize
./spotted init
```

## Configuration

### Chain Configuration
During initialization, you'll be prompted for:
- RPC URLs for each supported chain (defaults available)
- Bootstrap node configuration
- Deployment mode selection

### Deployment Modes
1. Docker Mode
   - Customizable setup
   - Manual database and Redis configuration
   - Integration with cloud services

2. Docker Compose Mode
   - Simplified setup
   - Predefined configurations
   - Automatic service orchestration

Example configuration will be written to `config/operator.yaml`:

```yaml
chains:
  84532:
    rpc: https://base-sepolia-rpc.publicnode.com
    contracts:
      registry: ""
      epochManager: ""
      stateManager: 0xe8Cbc41961125A1B0F86465Ff9a6666e39104E9e
    required_confirmations: 2
    average_block_time: 2
  # ... other chain configs ...

database:
  host: postgres
  port: 5432
  username: spotted
  # ... other database configs ...

p2p:
  bootstrap_peers: []
  port: 10000
  rendezvous: spotted-network
```

## Operation

### Environment Setup
Configure required environment variables:

```bash
# Required: P2P Key (base64 encoded)
export P2P_KEY_64="your-p2p-key-base64-encoded"

# Choose ONE of the following signing methods:

# Option 1: Private Key
export SIGNING_KEY_PRIV="your-private-key"

# Option 2: Keystore File
export SIGNING_KEY_PATH="/path/to/keystore.json"
export KEYSTORE_PASSWORD="your-keystore-password"
```

⚠️ Important: Use either Option 1 OR Option 2 for signing key configuration.

### Running the Node

For Docker mode:
```bash
./scripts/run-docker.sh
```

For Docker Compose mode:
```bash
./scripts/run-docker-compose.sh
```

## Architecture

### API Server

Built using the Chi router framework, the API server provides HTTP endpoints for network interaction.

#### Endpoints

All endpoints are under `/api/v1`:
- `POST /tasks`: Creates new tasks for network processing
- `GET /consensus`: Queries consensus status by parameters
- `GET /consensus/tasks/{taskID}`: Retrieves consensus information for specific tasks

Detailed API documentation can be found in the [USER](./docs/USER.md) section.

### Operator P2P Node

A decentralized network node built on libp2p, enabling operators to participate in consensus formation and task processing.

#### Components

##### Node
Core networking infrastructure that enables decentralized operator interaction:
- Decentralized peer discovery using Kademlia DHT
- Automatic peer connection management
- NAT traversal and address discovery
- Bootstrap node functionality
- Network address broadcasting

Network topology management:
- Bootstrap peer connections
- New peer discovery
- DHT routing table maintenance
- Address announcements
- Multi-transport protocol support

##### Connection Gater
Network security controller that:
- Validates peer IDs against registered P2P keys
- Manages connection permissions
- Maintains blacklist of malicious peers
- Performs multi-stage connection validation:
  - Pre-connection
  - Post-encryption
  - Post-upgrade

##### Event Listener
Blockchain event monitor for operator lifecycle management:
- Tracks operator registration events
- Manages operator activation periods
- Handles operator deregistration
- Maintains operator state

##### Task Processor
Consensus formation manager using pub/sub system:
- Collects and validates task responses
- Tracks operator weights and thresholds
- Verifies signatures and state claims
- Manages consensus state
- Broadcasts consensus achievements

##### Epoch Updator
Mainnet synchronization and epoch state manager:
- Monitors blockchain epoch transitions
- Updates operator status
- Synchronizes operator metadata
- Maintains network epoch state:
  - Minimum weight
  - Total weight
  - Threshold

##### Health Checker
Network reliability monitor:
- Regular connection health checks
- Peer disconnection management
- Network stability monitoring
- Connection status reporting

#### Task Response Topic

The network uses `/spotted/task-response` pub/sub topic for consensus formation:
- Primary channel for task validation
- All nodes subscribe for consensus participation
- Propagates validated task responses

## Workflows

### Task Creation
1. API handler receives/validates request
2. Task processor retrieves blockchain state
3. Operator signs task response
4. Response stored in database
5. Signed response broadcast to task-response topic

### Response Processing
1. Check for duplicate processing
2. Verify existing consensus
3. Validate response signature
4. Store response and update consensus weight
5. Check consensus threshold
6. Process task locally if needed
7. Broadcast own response

## License


