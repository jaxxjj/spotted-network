package operator

import "github.com/galxe/spotted-network/pkg/common/types"

// GetActiveOperatorsRoot returns the merkle root hash of all active operators
func (n *Node) GetActiveOperatorsRoot() []byte {
	n.activePeers.mu.RLock()
	operators := make([]*types.OperatorState, 0, len(n.activePeers.active))
	
	// Convert active operators to OperatorState
	for peerID, op := range n.activePeers.active {
		operators = append(operators, &types.OperatorState{
			PeerID:     peerID,
			Multiaddrs: op.Multiaddrs,
			Address:    op.Address,
			SigningKey: op.SigningKey,
			Weight:     op.Weight,
		})
	}
	n.activePeers.mu.RUnlock()

	// Compute merkle root using common merkle tree
	return types.ComputeStateRoot(operators)
}

func (n *Node) getCurrentStateRoot() []byte {
	n.activePeers.mu.RLock()
	defer n.activePeers.mu.RUnlock()
	return n.activePeers.stateRoot
}
