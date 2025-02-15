# Spotted Network

A decentralized network for state verification and consensus formation.

## Overview

The Spotted Network consists of two main components:
- API Server: RESTful service for task creation and consensus querying
- Operator P2P Node: Decentralized network node for consensus formation

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


