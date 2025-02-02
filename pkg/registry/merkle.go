package registry

import (
	"bytes"
	"context"
	"log"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

// GetActiveOperatorsRoot returns the merkle root hash of all active operators
func (n *Node) GetActiveOperatorsRoot() []byte {
	n.activeOperators.mu.RLock()
	operators := make([]*types.OperatorState, 0, len(n.activeOperators.active))
	ctx := context.Background()
	// Convert active operators to OperatorState
	for peerID, peerInfo := range n.activeOperators.active {
		operator, err  := n.opQuerier.GetOperatorByAddress(ctx, peerInfo.Address)
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

// VerifyOperatorStateHash verifies if operator's state root matches registry's state
func (n *Node) VerifyOperatorStateHash(peerID peer.ID, stateRoot []byte) bool {
	// Get current root hash
	registryStateRoot := n.GetActiveOperatorsRoot()
	
	// Return if hashes match
	return bytes.Equal(registryStateRoot, stateRoot)
}


