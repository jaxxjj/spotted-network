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


// UpsertActivePeerStates updates or adds multiple operator states to the active map
func (n *Node) UpsertActivePeerStates(states []*OperatorState) error {
	if len(states) == 0 {
		return fmt.Errorf("cannot update empty states")
	}

	n.activePeers.mu.Lock()
	defer n.activePeers.mu.Unlock()

	log.Printf("[StateSync] Starting batch update of %d operator states", len(states))
	
	// Update or add each state
	for _, state := range states {
		if state == nil {
			log.Printf("[StateSync] Skipping nil state in batch update")
			continue
		}
		
		n.activePeers.active[state.PeerID] = state
		log.Printf("[StateSync] Updated state for peer %s", state.PeerID)
	}

	log.Printf("[StateSync] Completed batch update of operator states")
	return nil
}



// PrintOperatorStates prints all operator states stored in memory
func (n *Node) PrintOperatorStates() {

}

