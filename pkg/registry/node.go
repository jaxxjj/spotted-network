package registry

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry"
	pb "github.com/galxe/spotted-network/proto"
	"google.golang.org/protobuf/proto"
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

	// State sync subscribers
	subscribers map[peer.ID]network.Stream
	subscribersMu sync.RWMutex
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

// handleStream handles incoming p2p streams
func (n *Node) handleStream(stream network.Stream) {
	defer stream.Close()

	// Read join request
	req, err := p2p.ReadJoinRequest(stream)
	if err != nil {
		log.Printf("Failed to read join request: %v", err)
		p2p.SendError(stream, fmt.Errorf("failed to read request: %v", err))
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

// HandleJoinRequest handles a join request from an operator
func (n *Node) HandleJoinRequest(ctx context.Context, req *pb.JoinRequest) error {
	log.Printf("Handling join request from operator %s", req.Address)

	// Check if operator is registered on chain
	mainnetClient := n.chainClients.GetMainnetClient()
	isRegistered, err := mainnetClient.IsOperatorRegistered(ctx, common.HexToAddress(req.Address))
	if err != nil {
		return fmt.Errorf("failed to check operator registration: %w", err)
	}

	if !isRegistered {
		return fmt.Errorf("operator not registered on chain")
	}

	// Get operator from database
	op, err := n.db.GetOperatorByAddress(ctx, req.Address)
	if err != nil {
		return fmt.Errorf("failed to get operator: %w", err)
	}

	// Verify signing key matches
	if op.SigningKey != req.SigningKey {
		return fmt.Errorf("signing key mismatch")
	}

	// Update status to waitingActive
	updatedOp, err := n.db.UpdateOperatorStatus(ctx, registry.UpdateOperatorStatusParams{
		Address: req.Address,
		Status:  string(OperatorStatusWaitingActive),
	})
	if err != nil {
		return fmt.Errorf("failed to update operator status: %w", err)
	}

	// Broadcast state update
	var exitEpoch *int32
	if updatedOp.ExitEpoch.Valid {
		val := updatedOp.ExitEpoch.Int32
		exitEpoch = &val
	}

	stateUpdate := []*pb.OperatorState{
		{
			Address:                 updatedOp.Address,
			SigningKey:             updatedOp.SigningKey,
			RegisteredAtBlockNumber: updatedOp.RegisteredAtBlockNumber.Int.Int64(),
			RegisteredAtTimestamp:   updatedOp.RegisteredAtTimestamp.Int.Int64(),
			ActiveEpoch:            int32(updatedOp.ActiveEpoch),
			ExitEpoch:              exitEpoch,
			Status:                 updatedOp.Status,
			Weight:                 updatedOp.Weight.Int.String(),
			Missing:                int32(updatedOp.Missing.Int32),
			SuccessfulResponseCount: int32(updatedOp.SuccessfulResponseCount.Int32),
		},
	}
	n.BroadcastStateUpdate(stateUpdate, "DELTA")

	log.Printf("Operator %s joined successfully", req.Address)
	return nil
}

// SetupProtocols sets up the protocol handlers for the registry node
func (n *Node) SetupProtocols() {
	// Set up join request handler
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleStream)
	
	// Set up state sync handler
	n.host.SetStreamHandler("/state-sync/1.0.0", n.handleStateSync)
	
	log.Printf("Protocol handlers set up")
}

func (n *Node) handleStateSync(stream network.Stream) {
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("Received state sync request from %s", remotePeer)
	
	// Read message type
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		log.Printf("Error reading message type: %v", err)
		stream.Reset()
		return
	}
	log.Printf("Received message type: 0x%x", msgType[0])

	// Read length prefix
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(stream, lengthBytes); err != nil {
		log.Printf("Error reading length prefix: %v", err)
		stream.Reset()
		return
	}
	length := uint32(lengthBytes[0])<<24 | 
		uint32(lengthBytes[1])<<16 | 
		uint32(lengthBytes[2])<<8 | 
		uint32(lengthBytes[3])
	log.Printf("Message length: %d bytes", length)

	// Read request data
	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		log.Printf("Error reading request data: %v", err)
		stream.Reset()
		return
	}

	switch msgType[0] {
	case 0x01: // GetFullStateRequest
		var req pb.GetFullStateRequest
		if err := proto.Unmarshal(data, &req); err != nil {
			log.Printf("Error unmarshaling GetFullStateRequest: %v", err)
			stream.Reset()
			return
		}
		log.Printf("Successfully identified request as GetFullStateRequest")
		n.handleGetFullState(stream)
		return

	case 0x02: // SubscribeRequest
		var req pb.SubscribeRequest
		if err := proto.Unmarshal(data, &req); err != nil {
			log.Printf("Error unmarshaling SubscribeRequest: %v", err)
			stream.Reset()
			return
		}
		log.Printf("Successfully identified request as SubscribeRequest")
		n.handleSubscribe(stream)
		return

	default:
		log.Printf("Unknown message type: 0x%x", msgType[0])
		stream.Reset()
		return
	}
}

func (n *Node) handleGetFullState(stream network.Stream) {
	log.Printf("Handling GetFullState request from %s", stream.Conn().RemotePeer())
	
	// Get all operators from database regardless of status
	operators, err := n.db.ListAllOperators(context.Background())
	if err != nil {
		log.Printf("Error getting operators from database: %v", err)
		stream.Reset()
		return
	}

	log.Printf("Found %d total operators in database", len(operators))
	if len(operators) == 0 {
		log.Printf("Warning: No operators found in database")
	}

	// Convert to proto message
	protoOperators := make([]*pb.OperatorState, 0, len(operators))
	for _, op := range operators {
		var exitEpoch *int32
		if op.ExitEpoch.Valid {
			val := op.ExitEpoch.Int32
			exitEpoch = &val
		}

		// Log operator details before conversion
		log.Printf("Converting operator - Address: %s, Status: %s, BlockNumber: %v, Timestamp: %v, Weight: %v",
			op.Address, op.Status, op.RegisteredAtBlockNumber, op.RegisteredAtTimestamp, op.Weight)

		protoOperators = append(protoOperators, &pb.OperatorState{
			Address:                 op.Address,
			SigningKey:             op.SigningKey,
			RegisteredAtBlockNumber: op.RegisteredAtBlockNumber.Int.Int64(),
			RegisteredAtTimestamp:   op.RegisteredAtTimestamp.Int.Int64(),
			ActiveEpoch:            int32(op.ActiveEpoch),
			ExitEpoch:              exitEpoch,
			Status:                 op.Status,
			Weight:                 op.Weight.Int.String(),
			Missing:                int32(op.Missing.Int32),
			SuccessfulResponseCount: int32(op.SuccessfulResponseCount.Int32),
		})
		log.Printf("Added operator to response - Address: %s, Status: %s", op.Address, op.Status)
	}

	// Create response
	resp := &pb.GetFullStateResponse{
		Operators: protoOperators,
	}

	// Marshal response
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		stream.Reset()
		return
	}

	// Write length prefix
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)
	
	if _, err := stream.Write(lengthBytes); err != nil {
		log.Printf("Error writing response length: %v", err)
		stream.Reset()
		return
	}

	// Write response data
	bytesWritten, err := stream.Write(data)
	if err != nil {
		log.Printf("Error sending response: %v", err)
		stream.Reset()
		return
	}

	log.Printf("Successfully sent full state with %d operators (%d bytes) to %s", 
		len(protoOperators), bytesWritten, stream.Conn().RemotePeer())
}

func (n *Node) handleSubscribe(stream network.Stream) {
	peer := stream.Conn().RemotePeer()
	
	n.subscribersMu.Lock()
	n.subscribers[peer] = stream
	n.subscribersMu.Unlock()

	log.Printf("Added subscriber %s", peer)

	// Keep stream open until closed by peer or error
	buf := make([]byte, 1024)
	for {
		_, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from subscriber %s: %v", peer, err)
			}
			n.subscribersMu.Lock()
			delete(n.subscribers, peer)
			n.subscribersMu.Unlock()
			stream.Close()
			log.Printf("Removed subscriber %s", peer)
			return
		}
	}
}

func (n *Node) BroadcastStateUpdate(operators []*pb.OperatorState, updateType string) {
	update := &pb.OperatorStateUpdate{
		Type:      updateType,
		Operators: operators,
	}

	data, err := proto.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling state update: %v", err)
		return
	}

	n.subscribersMu.RLock()
	defer n.subscribersMu.RUnlock()

	for peer, stream := range n.subscribers {
		if _, err := stream.Write(data); err != nil {
			log.Printf("Error sending update to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
		} else {
			log.Printf("Sent state update to %s - Type: %s, Operators: %d", 
				peer, updateType, len(operators))
		}
	}
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