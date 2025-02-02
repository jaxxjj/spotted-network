package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// getStateWithRetries attempts to get state with retries
func (tp *TaskProcessor) getStateWithRetries(ctx context.Context, chainClient ChainClient, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		// Get latest block for validation
		latestBlock, err := chainClient.BlockNumber(ctx)
		if err != nil {
			log.Printf("[StateCheck] Failed to get latest block: %v", err)
			time.Sleep(retryDelay)
			continue
		}

		// Validate block number
		if blockNumber > latestBlock {
			return nil, fmt.Errorf("block number %d is in the future (latest: %d)", blockNumber, latestBlock)
		}

		// Attempt to get state
		state, err := chainClient.GetStateAtBlock(ctx, target, key, blockNumber)
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

			log.Printf("[StateCheck] Attempt %d failed: %v", i+1, err)
			time.Sleep(retryDelay)
			continue
		}

		return state, nil
	}

	return nil, fmt.Errorf("failed to get state after %d retries", maxRetries)
}


// getOperatorWeight returns the weight of an operator in the operator's local storage
func (tp *TaskProcessor) getOperatorWeight(operatorAddr string) (*big.Int, error) {
	tp.node.operators.mu.RLock()
	defer tp.node.operators.mu.RUnlock()
	
	if state, exists := tp.node.operators.byAddress[operatorAddr]; exists {
		weight := new(big.Int)
		// Parse the weight string directly
		if _, ok := weight.SetString(state.Weight, 10); !ok {
			return big.NewInt(0), fmt.Errorf("[StateCheck] failed to parse weight for operator %s", operatorAddr)
		}
		// Log the actual weight value directly
		log.Printf("[StateCheck] Got operator %s weight: %v", operatorAddr, weight)
		return weight, nil
	}
	
	return big.NewInt(0), fmt.Errorf("[StateCheck] operator %s not found", operatorAddr)
}

// PrintOperatorStates prints all operator states stored in memory
func (n *Node) PrintOperatorStates() {
	n.operators.mu.RLock()
	defer n.operators.mu.RUnlock()

	log.Printf("\nOperator States (%d total):", len(n.operators.byAddress))
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
	log.Printf("| %-42s | %-12s | %-12s | %-10s |", "Address", "Status", "ActiveEpoch", "Weight")
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
	
	for _, op := range n.operators.byAddress {
		log.Printf("| %-42s | %-12s | %-12d | %-10s |", 
			op.Address,
			op.Status,
			op.ActiveEpoch,
			op.Weight,
		)
	}
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
}