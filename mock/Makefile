RPC_URL := http://localhost:8545

.PHONY: deploy-all deploy-registry deploy-epoch-manager emit-event deploy-mock-state-manager

deploy-all: deploy-registry deploy-epoch-manager emit-event deploy-mock-state-manager

deploy-registry:
	forge script script/DeployRegistry.s.sol:DeployRegistryScript --rpc-url $(RPC_URL) --broadcast

deploy-epoch-manager: deploy-registry
	forge script script/DeployEpochManager.s.sol:DeployEpochManagerScript --rpc-url $(RPC_URL) --broadcast

emit-event: deploy-epoch-manager
	forge script script/EmitEvent.s.sol:EmitEventScript --rpc-url $(RPC_URL) --broadcast

deploy-mock-state-manager: emit-event
	forge script script/DeployMockStateManager.s.sol:DeployMockStateManager --rpc-url $(RPC_URL) --broadcast
