package operator

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/contracts"
)

// getStateWithRetries attempts to get state with retries
func (tp *TaskProcessor) getStateWithRetries(ctx context.Context, stateClient contracts.StateClient, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		// Get latest block for validation
		latestBlock, err := stateClient.GetLatestBlockNumber(ctx)
		if err != nil {
			tp.logger.Printf("[StateCheck] Failed to get latest block: %v", err)
			time.Sleep(retryDelay)
			continue
		}

		// Validate block number
		if blockNumber > latestBlock {
			return nil, fmt.Errorf("block number %d is in the future (latest: %d)", blockNumber, latestBlock)
		}

		// Attempt to get state
		state, err := stateClient.GetStateAtBlock(ctx, target, key, blockNumber)
		if err != nil {
			// Check for specific contract errors
			if strings.Contains(err.Error(), "0x7c44ec9a") { // StateManager__BlockNotFound
				return nil, fmt.Errorf("block %d not found in state history", blockNumber)
			}
			if strings.Contains(err.Error(), "StateManager__KeyNotFound") {
				return nil, fmt.Errorf("key %s not found for address %s", key.String(), target.Hex())
			}
			if strings.Contains(err.Error(), "StateManager__NoHistoryFound") {
				return nil, fmt.Errorf("no state history found for block %d and key %s", blockNumber, key.String())
			}

			tp.logger.Printf("[StateCheck] Attempt %d failed: %v", i+1, err)
			time.Sleep(retryDelay)
			continue
		}

		return state, nil
	}

	return nil, fmt.Errorf("failed to get state after %d retries", maxRetries)
}

// isActiveOperator checks if a signing key belongs to an active operator
func (tp *TaskProcessor) isActiveOperator(signingKey string) bool {
	// Get all operator states
	tp.node.statesMu.RLock()
	defer tp.node.statesMu.RUnlock()

	// Check each operator's state
	for _, state := range tp.node.operatorStates {
		if state.Status == "active" {
			tp.logger.Printf("[Operator] Found active operator %s with status %s", state.Address, state.Status)
			return true
		}
	}

	tp.logger.Printf("[Operator] No active operator found for signing key %s", signingKey)
	return false
}

// getOperatorWeight returns the weight of an operator
func (tp *TaskProcessor) getOperatorWeight(operatorAddr string) (*big.Int, error) {
	tp.node.statesMu.RLock()
	defer tp.node.statesMu.RUnlock()
	
	if state, exists := tp.node.operatorStates[operatorAddr]; exists {
		weight := new(big.Int)
		// Parse the weight string directly
		if _, ok := weight.SetString(state.Weight, 10); !ok {
			return big.NewInt(0), fmt.Errorf("[TaskProcessor] failed to parse weight for operator %s", operatorAddr)
		}
		// Log the actual weight value directly
		tp.logger.Printf("[TaskProcessor] Got operator %s weight: %v", operatorAddr, weight)
		return weight, nil
	}
	return big.NewInt(0), fmt.Errorf("[TaskProcessor] operator %s not found", operatorAddr)
}
