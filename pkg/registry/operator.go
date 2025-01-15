package registry

import (
	"context"
	"log"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/registry"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) AddOperator(id peer.ID) {
	n.operatorsMu.Lock()
	defer n.operatorsMu.Unlock()

	// Check if operator is already registered
	if _, exists := n.operators[id]; exists {
		return
	}

	// Get operator's addresses from the host
	peerInfo := n.host.Peerstore().PeerInfo(id)

	// Add new operator with addresses
	n.operators[id] = &OperatorInfo{
		ID: id,
		Addrs: peerInfo.Addrs,
		LastSeen: time.Now(),
		Status: string(OperatorStatusActive),
	}
	log.Printf("New operator added: %s with addresses: %v\n", id, peerInfo.Addrs)

	// Broadcast new operator to all connected operators
	for opID := range n.operators {
		if opID != id {
			log.Printf("Broadcasting new operator %s to %s\n", id, opID)
			// Send operator info through p2p network
			if err := n.host.SendOperatorInfo(context.Background(), opID, id, peerInfo.Addrs); err != nil {
				log.Printf("Failed to broadcast operator info to %s: %v\n", opID, err)
			}
		}
	}

	// Send existing operators to the new operator
	for existingID, existingInfo := range n.operators {
		if existingID != id {
			log.Printf("Sending existing operator %s to new operator %s\n", existingID, id)
			if err := n.host.SendOperatorInfo(context.Background(), id, existingID, existingInfo.Addrs); err != nil {
				log.Printf("Failed to send existing operator info to %s: %v\n", id, err)
			}
		}
	}
}

func (n *Node) RemoveOperator(id peer.ID) {
	n.operatorsMu.Lock()
	defer n.operatorsMu.Unlock()

	delete(n.operators, id)
	log.Printf("Operator removed: %s\n", id)
}

func (n *Node) GetOperatorCount() int {
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()
	return len(n.operators)
}

// Get connected operators
func (n *Node) GetConnectedOperators() []peer.ID {
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()

	operators := make([]peer.ID, 0, len(n.operators))
	for id := range n.operators {
		operators = append(operators, id)
	}
	return operators
}

// GetHostID returns the node's libp2p host ID
func (n *Node) GetHostID() string {
	return n.host.ID().String()
}

// GetOperatorByAddress gets operator info from database
func (n *Node) GetOperatorByAddress(ctx context.Context, address string) (registry.Operators, error) {
	return n.db.GetOperatorByAddress(ctx, address)
}

// UpdateOperatorStatus updates operator status in database
func (n *Node) UpdateOperatorStatus(ctx context.Context, address string, status string) error {
	_, err := n.db.UpdateOperatorStatus(ctx, registry.UpdateOperatorStatusParams{
		Address: address,
		Status:  status,
	})
	return err
} 