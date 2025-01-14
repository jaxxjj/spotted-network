package registry

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	host *p2p.Host
	
	// Connected operators
	operators map[peer.ID]*OperatorInfo
	operatorsMu sync.RWMutex
	
	// Health check interval
	healthCheckInterval time.Duration

	// Database connection
	db *registry.Queries

	// Event listener
	eventListener *EventListener

	// Chain clients manager
	chainClients *ethereum.ChainClients
}

type OperatorInfo struct {
	ID peer.ID
	Addrs []multiaddr.Multiaddr
	LastSeen time.Time
	Status string
}

func NewNode(ctx context.Context, cfg *p2p.Config, db *registry.Queries, chainClients *ethereum.ChainClients) (*Node, error) {
	host, err := p2p.NewHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	node := &Node{
		host:               host,
		operators:          make(map[peer.ID]*OperatorInfo),
		healthCheckInterval: 5 * time.Second,
		db:                db,
		chainClients:      chainClients,
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

// handleStream handles incoming p2p streams
func (n *Node) handleStream(stream network.Stream) {
	defer stream.Close()

	// Read join request
	req, err := p2p.ReadJoinRequest(stream)
	if err != nil {
		log.Printf("Failed to read join request: %v", err)
		stream.Reset()
		return
	}

	// Handle join request
	err = n.HandleJoinRequest(context.Background(), req)
	if err != nil {
		log.Printf("Failed to handle join request: %v", err)
		p2p.SendError(stream, err)
		return
	}

	// Send success response
	p2p.SendSuccess(stream)
}

func (n *Node) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(n.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.checkOperators(ctx)
		}
	}
}

func (n *Node) checkOperators(ctx context.Context) {
	n.operatorsMu.Lock()
	defer n.operatorsMu.Unlock()

	for id, info := range n.operators {
		// Ping the operator
		if err := n.host.PingPeer(ctx, id); err != nil {
			log.Printf("Operator %s is unreachable: %v\n", id, err)
			info.Status = string(OperatorStatusInactive)
		} else {
			info.LastSeen = time.Now()
			info.Status = string(OperatorStatusActive)
			log.Printf("Operator %s is healthy\n", id)
		}
	}
}

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

func (n *Node) Stop() error {
	return n.host.Close()
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

// HandleJoinRequest handles the join request from an operator
func (n *Node) HandleJoinRequest(ctx context.Context, req *p2p.JoinRequest) error {
	// Get mainnet client
	mainnetClient := n.chainClients.GetMainnetClient()

	// Verify operator is registered on chain
	operatorAddr := common.HexToAddress(req.OperatorAddress)
	isRegistered, err := mainnetClient.IsOperatorRegistered(ctx, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to check operator registration: %w", err)
	}

	if !isRegistered {
		return fmt.Errorf("operator not registered on chain")
	}

	// Get operator from database
	op, err := n.db.GetOperatorByAddress(ctx, req.OperatorAddress)
	if err != nil {
		return fmt.Errorf("failed to get operator: %w", err)
	}

	// Verify signing key matches
	if op.SigningKey != req.SigningKey {
		return fmt.Errorf("signing key mismatch")
	}

	// TODO: Implement signature verification
	// if !n.verifySignature(req.OperatorAddress, req.Message, req.Signature) {
	// 	return fmt.Errorf("invalid signature")
	// }

	// Update status to waitingActive
	_, err = n.db.UpdateOperatorStatus(ctx, registry.UpdateOperatorStatusParams{
		Address: req.OperatorAddress,
		Status:  string(OperatorStatusWaitingActive),
	})
	if err != nil {
		return fmt.Errorf("failed to update operator status: %w", err)
	}

	log.Printf("Operator %s joined successfully", req.OperatorAddress)
	return nil
}

// SetupProtocols sets up the protocol handlers for the registry node
func (n *Node) SetupProtocols() {
	// Handle join requests
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleStream)
}

// TODO: Implement these helper functions
func ReadJoinRequest(stream network.Stream) (*JoinRequest, error) {
	// TODO: Implement reading join request from stream
	return nil, fmt.Errorf("not implemented")
}

func SendError(stream network.Stream, err error) {
	// TODO: Implement sending error response
}

func SendSuccess(stream network.Stream) {
	// TODO: Implement sending success response
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