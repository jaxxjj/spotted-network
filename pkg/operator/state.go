package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	utils "github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
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

	n.activeOperators.mu.Lock()
	defer n.activeOperators.mu.Unlock()

	log.Printf("[StateSync] Starting batch update of %d operator states", len(states))
	
	// Update or add each state
	for _, state := range states {
		if state == nil {
			log.Printf("[StateSync] Skipping nil state in batch update")
			continue
		}
		
		n.activeOperators.active[state.PeerID] = state
		log.Printf("[StateSync] Updated state for peer %s, weight: %s, address: %s, multiaddrs: %v, signing key: %s", state.PeerID, state.Weight.String(), state.Address, state.Multiaddrs, state.SigningKey)
	}

	log.Printf("[StateSync] Completed batch update of operator states")
	return nil
}

func (n *Node) getOperatorWeight(peerID peer.ID) (*big.Int, error) {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()

	operator, ok := n.activeOperators.active[peerID]
	if !ok {
		return nil, fmt.Errorf("operator %s not found", peerID)
	}

	return operator.Weight, nil
}

// GetConnectedOperators returns all connected operator IDs
func (n *Node) getActivePeerIDs() []peer.ID {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()

	operators := make([]peer.ID, 0, len(n.activeOperators.active))
	for id := range n.activeOperators.active {
		operators = append(operators, id)
	}
	return operators
}

// convertToOperatorState converts a protobuf operator state to internal operator state
func (n *Node) convertToOperatorState(opState *pb.OperatorPeerState) (*OperatorState, error) {
	// Convert multiaddrs
	addrs := make([]multiaddr.Multiaddr, 0, len(opState.Multiaddrs))
	for _, addrStr := range opState.Multiaddrs {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			log.Printf("[StateSync] Failed to parse multiaddr %s: %v", addrStr, err)
			continue
		}
		addrs = append(addrs, addr)
	}

	// Parse peer ID
	peerID, err := peer.Decode(opState.PeerId)
	if err != nil {
		log.Printf("[StateSync] Failed to decode peer ID %s: %v", opState.PeerId, err)
		return nil, err
	}

	// Create operator state
	state := &OperatorState{
		PeerID:     peerID,
		Multiaddrs: addrs,
		Address:    opState.Address,
		SigningKey: opState.SigningKey,
		Weight:     utils.StringToBigInt(opState.Weight),
	}

	return state, nil
}

// updateOperatorStates updates operator states and related state root
func (n *Node) updateOperatorStates(ctx context.Context, states []*OperatorState) {
	if len(states) > 0 {
		if err := n.UpsertActivePeerStates(states); err != nil {
			log.Printf("[StateSync] Failed to batch update operator states: %v", err)
			return
		}
		rootHash := n.computeActiveOperatorsRoot()
		n.setCurrentStateRoot(rootHash)
		n.PrintOperatorStates()
		n.UpdateActiveConnections(ctx, n.activeOperators.active)
		log.Printf("[StateSync] Successfully updated local state with %d operators", len(n.activeOperators.active))
	}
}

// PrintOperatorStates prints all operator states stored in memory
func (n *Node) PrintOperatorStates() {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()

	for peerID, state := range n.activeOperators.active {
		log.Printf("Operator ID: %s, Address: %s, Weight: %s", peerID, state.Address, state.Weight.String())
	}
	log.Printf("Current State Root: %s", n.activeOperators.stateRoot)
}

