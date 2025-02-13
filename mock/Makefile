# Include .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: deploy-all deploy-registry deploy-epoch-manager emit-register-event deploy-mock-state-manager generate-keys check-balance clear-pending check-nonce deploy-high-gas deploy-with-nonce replace-tx cancel-tx test-state query-genesis

check-nonce:
	@echo "Checking account nonce..."
	@echo "Current nonce: $$(cast nonce --rpc-url $(RPC_URL) $$(cast wallet address --private-key $(PRIVATE_KEY)))"
	@echo "Pending nonce: $$(cast nonce --rpc-url $(RPC_URL) --pending $$(cast wallet address --private-key $(PRIVATE_KEY)))"

check-balance:
	@echo "Checking account balance..."
	@cast balance --rpc-url $(RPC_URL) $(shell cast wallet address --private-key $(PRIVATE_KEY))

clear-pending:
	@echo "Clearing pending transactions..."
	@cast send --rpc-url $(RPC_URL) --private-key $(PRIVATE_KEY) $(shell cast wallet address --private-key $(PRIVATE_KEY)) --value 0 --gas-price 20000000000 --legacy --nonce 607

# Deploy with specific nonce
deploy-with-nonce:
	@echo "Deploying with nonce 607..."
	forge script script/DeployMockStateManager.s.sol:DeployMockStateManager --rpc-url $(RPC_URL) --broadcast --private-key $(PRIVATE_KEY) --legacy --gas-price 30000000000 --gas-limit 5000000 --nonce 607

# High gas deployment commands
deploy-high-gas:
	@echo "Deploying with high gas..."
	forge script script/DeployMockStateManager.s.sol:DeployMockStateManager --rpc-url $(RPC_URL) --broadcast --private-key $(PRIVATE_KEY) --legacy --gas-price 20000000000 --gas-limit 5000000

deploy-all: check-balance deploy-registry deploy-epoch-manager emit-register-event deploy-mock-state-manager

deploy-registry:
	forge script script/DeployRegistry.s.sol:DeployRegistryScript --rpc-url $(RPC_URL) --broadcast --private-key $(PRIVATE_KEY) --legacy --gas-price 20000000000

deploy-epoch-manager: 
	forge script script/DeployEpochManager.s.sol:DeployEpochManagerScript --rpc-url $(RPC_URL) --broadcast --private-key $(PRIVATE_KEY) --legacy --gas-price 20000000000

emit-register-event: 
	forge script script/EmitRegisterEvent.s.sol:EmitRegisterEvent --rpc-url $(RPC_URL) --broadcast --private-key $(PRIVATE_KEY) --legacy --gas-price 20000000000

deploy-mock-state-manager: 
	forge script script/DeployMockStateManager.s.sol:DeployMockStateManager --rpc-url $(RPC_URL) --broadcast --private-key $(PRIVATE_KEY) --legacy --gas-price 20000000000

emit-deregister-event: 
	forge script script/EmitDeregisterEvent.s.sol:EmitDeregisterEvent --rpc-url $(RPC_URL) --broadcast --private-key $(PRIVATE_KEY) --legacy --gas-price 20000000000

# Test state queries
test-state:
	@echo "Testing getHistoryAtBlock for ALICE at block 7698667..."
	@cast call --rpc-url $(RPC_URL) 0xe8Cbc41961125A1B0F86465Ff9a6666e39104E9e "getHistoryAtBlock(address,uint256,uint256)(uint256)" 0x0000000000000000000000000000000000001111 1 7698667

# Query genesis block
query-genesis:
	@echo "Querying GENESIS_BLOCK value..."
	@cast call --rpc-url $(RPC_URL) 0x5bFB7609a51F8577D90e8576DE6e85BC7fBf08F7 "GENESIS_BLOCK()(uint64)"

spotted-network % curl -v -X POST \
  -H "Content-Type: application/json" \
  -d '{"chain_id":11155111,"target_address":"0x0000000000000000000000000000000000001111","key":"1","block_number":7698667}' \
  http://64.111.64.88:8080/api/v1/tasks