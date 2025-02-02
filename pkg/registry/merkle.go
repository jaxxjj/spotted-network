package registry

import (
	"context"
	"log"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/types"
)

// GetActiveOperatorsRoot returns the merkle root hash of all active operators
func (n *Node) computeActiveOperatorsRoot() []byte {
	n.activeOperators.mu.RLock()
	operators := make([]*types.OperatorState, 0, len(n.activeOperators.active))
	ctx := context.Background()
	// Convert active operators to OperatorState
	for peerID, peerInfo := range n.activeOperators.active {
		operator, err  := n.opQuerier.GetOperatorByAddress(ctx, peerInfo.Address)
		if err != nil {
			log.Printf("[StateSync] Failed to get operator for address %s: %v", peerInfo.Address, err)
			continue
		}
		weight, err := utils.NumericToBigInt(operator.Weight)
		if err != nil {
			log.Printf("[StateSync] Failed to get operator for address %s: %v", peerInfo.Address, err)
			continue
		}
		operators = append(operators, &types.OperatorState{
			PeerID:     peerID,
			Multiaddrs: peerInfo.Multiaddrs,
			Address:    peerInfo.Address,
			SigningKey: operator.SigningKey,
			Weight:     weight,
		})
	}
	n.activeOperators.mu.RUnlock()

	// Compute merkle root
	return types.ComputeStateRoot(operators)
}

func (n *Node) getActiveOperatorsRoot() []byte {
	n.activeOperators.mu.RLock()
	stateRoot := n.activeOperators.stateRoot
	n.activeOperators.mu.RUnlock()
	return stateRoot
}

func (n *Node) setActiveOperatorsRoot(stateRoot []byte) {
	n.activeOperators.mu.Lock()
	n.activeOperators.stateRoot = stateRoot
	n.activeOperators.mu.Unlock()
}



