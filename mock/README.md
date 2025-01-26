# Mock Registry Environment

This directory contains a mock environment for testing the Registry node's event listening functionality using Foundry.

## Prerequisites

1. Install Foundry:
```bash
curl -L https://foundry.paradigm.xyz | bash
foundryup
```

## Setup

1. Install dependencies:
```bash
forge install
```

2. Start the local environment:
```bash
docker-compose up -d postgres anvil
```

3. Deploy the mock contract:
```bash
# Get the default anvil private key
export PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

# Deploy the contract
forge script script/Deploy.s.sol --rpc-url http://localhost:8545 --broadcast
```
The deployment script will output the contract address. Save this address.

4. Set the environment variable:
```bash
export REGISTRY_ADDRESS=<deployed-contract-address>
```

5. Start the registry node:
```bash
docker-compose up registry
```

## Testing

To emit a test event:

```bash
forge script script/EmitEvent.s.sol --rpc-url http://localhost:8545 --broadcast
```

This will:
1. Generate a signing key from the deployer's private key
2. Emit an OperatorRegistered event
3. The registry node should detect this event and store it in the database

## Verifying

To verify the event was processed:

1. Connect to the database:
```bash
docker exec -it spotted-network-postgres-1 psql -U spotted -d spotted
```

2. Query the operators table:
```sql
SELECT * FROM operators;
```

You should see a record for the operator that was registered in the event.

## Cleanup

To stop all services:
```bash
docker-compose down -v
```
