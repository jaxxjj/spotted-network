package registry

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func NewNode(ctx context.Context, cfg *p2p.Config, db *operators.Queries, chainClients *ethereum.ChainClients) (*Node, error) {
	host, err := p2p.NewHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	node := &Node{
		host:               host,
		operators:          make(map[peer.ID]*OperatorInfo),
		healthCheckInterval: 100 * time.Second,
		db:                db,
		chainClients:      chainClients,
		subscribers:       make(map[peer.ID]network.Stream),
	}

	// Create and initialize event listener
	node.eventListener = NewEventListener(chainClients, db)

	// Start listening for chain events
	if err := node.eventListener.StartListening(ctx); err != nil {
		return nil, fmt.Errorf("failed to start event listener: %w", err)
	}

	// Set stream handler for operator connections
	node.SetupProtocols()

	// Start health check
	go node.startHealthCheck(ctx)

	log.Printf("Registry Node started. %s\n", host.GetHostInfo())
	return node, nil
}

func (n *Node) Start(ctx context.Context) error {
	// Start epoch monitoring
	go n.monitorEpochUpdates(ctx)

	return nil
}

func (n *Node) Stop() error {
	return n.host.Close()
}

// GetOperatorInfo returns information about a connected operator
func (n *Node) GetOperatorInfo(id peer.ID) *OperatorInfo {
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()
	return n.operators[id]
}

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
func (n *Node) GetOperatorByAddress(ctx context.Context, address string) (operators.Operators, error) {
	return n.db.GetOperatorByAddress(ctx, address)
}

// UpdateOperatorStatus updates operator status in database
func (n *Node) UpdateOperatorStatus(ctx context.Context, address string, status string) error {
	_, err := n.db.UpdateOperatorStatus(ctx, operators.UpdateOperatorStatusParams{
		Address: address,
		Status:  status,
	})
	return err
}
