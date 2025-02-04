package operator

import (
	"log"

	"github.com/galxe/spotted-network/pkg/common/types"
)

// GetActiveOperatorsRoot returns the merkle root hash of all active operators
func (n *Node) computeActiveOperatorsRoot() []byte {
	n.activeOperators.mu.RLock()
	log.Printf("[Merkle] Starting to compute root for %d operators", len(n.activeOperators.active))
	operators := make([]*types.OperatorState, 0, len(n.activeOperators.active))
	
	// Convert active operators to OperatorState
	for peerID, op := range n.activeOperators.active {
		// Log operator details
		log.Printf("[Merkle] Processing operator - PeerID: %s", peerID)
		
		// Skip if required fields are nil
		if op == nil || op.Weight == nil {
			log.Printf("[Merkle] Skipping operator %s due to nil fields", peerID)
			continue
		}
		
		operators = append(operators, &types.OperatorState{
			PeerID:     peerID,
			Multiaddrs: op.Multiaddrs,
			Address:    op.Address,
			SigningKey: op.SigningKey,
			Weight:     op.Weight,
		})
	}
	n.activeOperators.mu.RUnlock()

	// If no valid operators, return empty root
	if len(operators) == 0 {
		log.Printf("[Merkle] No valid operators to compute root")
		return nil
	}

	// Compute merkle root using common merkle tree
	root := types.ComputeStateRoot(operators)
	
	// Log result
	if root == nil {
		log.Printf("[Merkle] WARNING: ComputeStateRoot returned nil")
	} else {
		log.Printf("[Merkle] Successfully computed root hash: %x", root)
	}
	
	return root
}

func (n *Node) getCurrentStateRoot() []byte {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()
	return n.activeOperators.stateRoot
}

func (n *Node) setCurrentStateRoot(root []byte) {
	n.activeOperators.mu.Lock()
	defer n.activeOperators.mu.Unlock()
	n.activeOperators.stateRoot = root
}
