package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type MainnetClient interface {
	GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error)
	WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorRegisteredEvent) (event.Subscription, error)
	WatchOperatorDeregistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorDeregisteredEvent) (event.Subscription, error) 
	BlockNumber(ctx context.Context) (uint64, error)
	GetOperatorWeight(ctx context.Context, operator common.Address) (*big.Int, error)
	IsOperatorRegistered(ctx context.Context, operator common.Address) (bool, error) 
}

type Node struct {
	host *p2p.Host
	
	// Connected operators
	operatorsInfo map[peer.ID]*OperatorInfo
	operatorsInfoMu sync.RWMutex
	
	// Health check interval
	healthCheckInterval time.Duration

	// Database connection
	operators OperatorsQuerier

	// Event listener
	eventListener *EventListener
	
	// Epoch updator
	epochUpdator *EpochUpdator

	// Chain clients manager
	mainnetClient MainnetClient

	// State sync subscribers
	subscribers map[peer.ID]network.Stream
	subscribersMu sync.RWMutex


}

func NewNode(ctx context.Context, cfg *p2p.Config, operatorsQuerier OperatorsQuerier, chainClients MainnetClient) (*Node, error) {
	host, err := p2p.NewHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	node := &Node{
		host:               host,
		operatorsInfo:          make(map[peer.ID]*OperatorInfo),
		healthCheckInterval: 100 * time.Second,
		operators:                operatorsQuerier,
		mainnetClient:      chainClients,
		subscribers:       make(map[peer.ID]network.Stream),
	}

	// Create and initialize event listener
	node.eventListener = NewEventListener(node, chainClients, operatorsQuerier)

	// Create and initialize epoch updator
	node.epochUpdator = NewEpochUpdator(node)

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
	// Validate required components
	if n.host == nil {
		return fmt.Errorf("[Registry] host not initialized")
	}
	if n.operators == nil {
		return fmt.Errorf("[Registry] operators database not initialized") 
	}
	if n.epochUpdator == nil {
		return fmt.Errorf("[Registry] epoch updator not initialized")
	}
	if n.eventListener == nil {
		return fmt.Errorf("[Registry] event listener not initialized")
	}

	// Start state sync service
	if err := n.startStateSync(ctx); err != nil {
		return fmt.Errorf("failed to start state sync service: %w", err)
	}
	log.Printf("[Registry] State sync service started")

	// Start epoch monitoring
	go n.epochUpdator.Start(ctx)
	log.Printf("[Registry] Epoch monitoring started")

	return nil
}

func (n *Node) Stop() error {
	return n.host.Close()
}

// GetOperatorInfo returns information about a connected operator
func (n *Node) GetOperatorInfo(id peer.ID) *OperatorInfo {
	n.operatorsInfoMu.RLock()
	defer n.operatorsInfoMu.RUnlock()
	return n.operatorsInfo[id]
}

func (n *Node) GetConnectedOperators() []peer.ID {
	n.operatorsInfoMu.RLock()
	defer n.operatorsInfoMu.RUnlock()

	operators := make([]peer.ID, 0, len(n.operatorsInfo))
	for id := range n.operatorsInfo {
		operators = append(operators, id)
	}
	return operators
}

// GetHostID returns the node's libp2p host ID
func (n *Node) GetHostID() string {
	return n.host.ID().String()
}

