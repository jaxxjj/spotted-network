RPC_URL := http://localhost:8545

.PHONY: deploy-all deploy-registry deploy-epoch-manager emit-register-event emit-deregister-event deploy-mock-state-manager

deploy-all: deploy-registry deploy-epoch-manager emit-register-event deploy-mock-state-manager

deploy-registry:
	forge script script/DeployRegistry.s.sol:DeployRegistryScript --rpc-url $(RPC_URL) --broadcast

deploy-epoch-manager: 
	forge script script/DeployEpochManager.s.sol:DeployEpochManagerScript --rpc-url $(RPC_URL) --broadcast

emit-register-event: 
	forge script script/EmitRegisterEvent.s.sol:EmitRegisterEvent --rpc-url $(RPC_URL) --broadcast

emit-deregister-event: 
	forge script script/EmitDeregisterEvent.s.sol:EmitDeregisterEvent --rpc-url $(RPC_URL) --broadcast

deploy-mock-state-manager: 
	forge script script/DeployMockStateManager.s.sol:DeployMockStateManager --rpc-url $(RPC_URL) --broadcast