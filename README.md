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

- Operator 1:
  - P2P: 10000
  - HTTP: 8000
- Operator 2:
  - P2P: 10001
  - HTTP: 8001
- Operator 3:
  - P2P: 10002
  - HTTP: 8002

## Development

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


Generated P2P key for operator1:
  Private key (base64): CAESQKW/y8x4MBT09AySrCDS1HXvsFEGoXLwqvWOQUifZ90TvdsBG0rSgcjJTH8qWwRYRysJaZ+7Z4egLxvShvBnQys=
  Public key (base64): CAESIL3bARtK0oHIyUx/KlsEWEcrCWmfu2eHoC8b0obwZ0Mr
  P2PKey: 0x310c8425b620980dcfcf756e46572bb6ac80eb07
  PeerId: 12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p

Generated P2P key for operator2:
  Private key (base64): CAESQHGMebvS8Wf6IZZh40yacCPzXhRlKqJCGfPySZyCFid6EdbnbwgelZkcZbllzWAZFfrdV/dcf2poB1OySA2mV0I=
  Public key (base64): CAESIBHW528IHpWZHGW5Zc1gGRX63Vf3XH9qaAdTskgNpldC
  P2PKey: 0x01078ffbf1de436d6f429f5ce6be8fd9d6e16165
  PeerId: 12D3KooWB21ALruLNKbH5vLjDjiG1mM9XVnCJMGdsdo2LKvoqneD

Generated P2P key for operator3:
  Private key (base64): CAESQM5ltPHuttHq7/HHHHymN5A/XSDKt5EPOwGWor2H3k0PXckF23DDwxzmdOhEtOy5f8szIAYWqSFH8cIlICumemo=
  Public key (base64): CAESIF3JBdtww8Mc5nToRLTsuX/LMyAGFqkhR/HCJSArpnpq
  P2PKey: 0x67aa23adde2459a1620be2ea28982310597521b0
  PeerId: 12D3KooWG8TsS8YsbnfArhnQznH6Rmumt6fwE9J66ZN8tF7n9GVf

```solidity
    address public constant OPERATOR_1 = address(0xCf593639B34CaE0ea3217dA27014ab5FbBAc8342);
    address public constant OPERATOR_2 = address(0xCCE3B4EC7681B4EcF5fD5b50e562A88a33E5137B);
    address public constant OPERATOR_3 = address(0xFE6B5379E861C79dB03eb3a01F3F1892FC4141D5);
    address public constant SIGNING_KEY_1 = address(0x9F0D8BAC11C5693a290527f09434b86651c66Bf2);
    address public constant SIGNING_KEY_2 = address(0xeBBAce05Db3D717A5BA82EAB8AdE712dFb151b13);
    address public constant SIGNING_KEY_3 = address(0x083739b681B85cc2c9e394471486321D6446b25b);
    address public constant P2P_KEY_1 = address(0x310C8425b620980DCFcf756e46572bb6ac80Eb07);
    address public constant P2P_KEY_2 = address(0x01078ffBf1De436d6f429f5Ce6Be8Fd9D6E16165);
    address public constant P2P_KEY_3 = address(0x67aa23adde2459a1620BE2Ea28982310597521b0);
```

# Spotted Network CLI

A command-line interface for interacting with the Spotted Network.

## Quick Installation

Install the CLI using the following command:

```bash
curl -sSfL https://raw.githubusercontent.com/galxe/spotted-network/master/scripts/install.sh | sh -s
```

The binary will be installed inside the `~/bin` directory.

To add the binary to your path, run:

```bash
export PATH=$PATH:~/bin
```

## Custom Installation Location

To install the CLI in a custom location, use:

```bash
curl -sSfL https://raw.githubusercontent.com/galxe/spotted-network/master/scripts/install.sh | sh -s -- -b <custom_location>
```

## Usage

### Basic Commands

```bash
# Show version information
spotted --version

# Run with config file
spotted --config /path/to/config.yaml

# Enable debug mode
spotted --config /path/to/config.yaml --debug
```

### Configuration

Create a configuration file (e.g., `config.yaml`) with the following structure:

```yaml
operator:
  signing_key: /path/to/signing.key
  p2p_key: /path/to/p2p.key
  
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: spotted

chain:
  mainnet_rpc: https://mainnet.infura.io/v3/your-project-id
  testnet_rpc: https://goerli.infura.io/v3/your-project-id
```

## Building from Source

1. Clone the repository:
```bash
git clone https://github.com/galxe/spotted-network.git
cd spotted-network
```

2. Build the binary:
```bash
make build
```

3. Install locally:
```bash
make install
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
