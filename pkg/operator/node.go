package operator

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	pb "github.com/galxe/spotted-network/proto"
)

type Node struct {
	host           host.Host
	registryID     peer.ID
	registryAddr   string
	signer         signer.Signer
	knownOperators map[peer.ID]*peer.AddrInfo
	operatorsMu    sync.RWMutex
	pingService    *ping.PingService
	
	// State sync related fields
	operatorStates map[string]*pb.OperatorState  // address -> state
	statesMu       sync.RWMutex
}

func NewNode(registryAddr string, s signer.Signer) (*Node, error) {
	// Parse the registry multiaddr
	maddr, err := multiaddr.NewMultiaddr(registryAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry address: %w", err)
	}

	// Extract peer ID from multiaddr
	addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry peer info: %w", err)
	}

	// Create a new host with default configuration
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Create ping service
	pingService := ping.NewPingService(host)

	return &Node{
		host:           host,
		registryID:     addrInfo.ID,
		registryAddr:   registryAddr,
		signer:         s,
		knownOperators: make(map[peer.ID]*peer.AddrInfo),
		operatorsMu:    sync.RWMutex{},
		pingService:    pingService,
		operatorStates: make(map[string]*pb.OperatorState),
		statesMu:      sync.RWMutex{},
	}, nil
}

func (n *Node) Start() error {
	log.Printf("Operator node started with ID: %s", n.host.ID())

	log.Printf("Connecting to registry...")
	if err := n.connectToRegistry(); err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	log.Printf("Successfully connected to registry")

	// Start message handler and health check
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleMessages)
	go n.healthCheck()
	log.Printf("Message handler and health check started")

	// Subscribe to state updates
	if err := n.subscribeToStateUpdates(); err != nil {
		return fmt.Errorf("failed to subscribe to state updates: %w", err)
	}
	log.Printf("Subscribed to state updates")

	// Get initial state
	if err := n.getFullState(); err != nil {
		return fmt.Errorf("failed to get initial state: %w", err)
	}
	log.Printf("Got initial state")

	// Announce to registry
	log.Printf("Announcing to registry...")
	if err := n.announceToRegistry(); err != nil {
		return fmt.Errorf("failed to announce to registry: %w", err)
	}
	log.Printf("Successfully announced to registry")

	return nil
}

func (n *Node) handleMessages(stream network.Stream) {
	defer stream.Close()
	
	log.Printf("New stream opened from: %s", stream.Conn().RemotePeer())

	// Read the entire message
	data, err := io.ReadAll(stream)
	if err != nil {
		log.Printf("Error reading from stream: %v", err)
		return
	}

	// Parse protobuf message
	var req pb.JoinRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Printf("Error parsing protobuf message: %v", err)
		return
	}

	log.Printf("Received join request from operator: %s", req.Address)

	// Send success response
	resp := &pb.JoinResponse{
		Success: true,
	}
	respData, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	if _, err := stream.Write(respData); err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}

	log.Printf("Successfully processed join request from: %s", req.Address)
}

func (n *Node) healthCheck() {
	// Ping registry node
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := <-n.pingService.Ping(ctx, n.registryID)
	if result.Error != nil {
		log.Printf("Failed to ping registry: %v\n", result.Error)
		return
	}
	log.Printf("Successfully pinged registry (RTT: %v)\n", result.RTT)

	// Ping all known operators
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()

	for operatorID, peerInfo := range n.knownOperators {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try to connect if not connected
		if err := n.host.Connect(ctx, *peerInfo); err != nil {
			log.Printf("Failed to connect to operator %s: %v\n", operatorID, err)
			continue
		}

		// Ping the operator
		result := <-n.pingService.Ping(ctx, operatorID)
		if result.Error != nil {
			log.Printf("Failed to ping operator %s: %v\n", operatorID, result.Error)
			continue
		}
		log.Printf("Successfully pinged operator %s (RTT: %v)\n", operatorID, result.RTT)
	}
}

func (n *Node) connectToRegistry() error {
	log.Printf("Attempting to connect to Registry Node at address: %s\n", n.registryAddr)

	// Create multiaddr from the provided address
	addr, err := multiaddr.NewMultiaddr(n.registryAddr)
	if err != nil {
		return fmt.Errorf("invalid registry address: %s", n.registryAddr)
	}

	// Parse peer info from multiaddr
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("failed to parse peer info from address: %v", err)
	}

	// Connect to registry
	if err := n.host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to registry: %v", err)
	}

	// Save the registry ID
	n.registryID = peerInfo.ID

	log.Printf("Successfully connected to Registry Node with ID: %s\n", n.registryID)
	return nil
}

func (n *Node) Stop() error {
	return n.host.Close()
}

func (n *Node) announceToRegistry() error {
	stream, err := n.host.NewStream(context.Background(), peer.ID(n.registryID), "/spotted/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open stream to registry: %w", err)
	}
	defer stream.Close()

	// Create join request message
	message := []byte("join_request")
	signature, err := n.signer.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	// Create protobuf request
	req := &pb.JoinRequest{
		Address:    n.signer.GetAddress().Hex(),
		Message:    string(message),
		Signature:  hex.EncodeToString(signature),
		SigningKey: n.signer.GetSigningKey(),
	}

	// Marshal and send request
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// State sync related functions

func (node *Node) subscribeToStateUpdates() error {
	log.Printf("Opening state sync stream to registry...")
	stream, err := node.host.NewStream(context.Background(), node.registryID, "/state-sync/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open state sync stream: %w", err)
	}

	// Send message type
	msgType := []byte{0x02} // 0x02 for SubscribeRequest
	if _, err := stream.Write(msgType); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Send subscribe request
	req := &pb.SubscribeRequest{}
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal subscribe request: %w", err)
	}

	// Send length prefix
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)
	
	if _, err := stream.Write(lengthBytes); err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}

	log.Printf("Sending state sync subscribe request (type: 0x02, length: %d)...", length)
	bytesWritten, err := stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send subscribe request: %w", err)
	}
	log.Printf("Sent %d bytes to registry", bytesWritten)
	log.Printf("Successfully subscribed to state updates")

	// Handle updates in background
	go node.handleStateUpdates(stream)
	return nil
}

func (node *Node) getFullState() error {
	log.Printf("Requesting full operator state from registry...")
	stream, err := node.host.NewStream(context.Background(), node.registryID, "/state-sync/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open state sync stream: %w", err)
	}
	defer stream.Close()

	// Send get full state request with message type
	msgType := []byte{0x01} // 0x01 for GetFullStateRequest
	if _, err := stream.Write(msgType); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Send get full state request
	req := &pb.GetFullStateRequest{}
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal get state request: %w", err)
	}

	// Send length prefix
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)
	
	if _, err := stream.Write(lengthBytes); err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}

	log.Printf("Sending GetFullState request to registry (type: 0x01, length: %d)...", length)
	bytesWritten, err := stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send get state request: %w", err)
	}
	log.Printf("Sent %d bytes to registry", bytesWritten)

	// Read response
	log.Printf("Waiting for response from registry...")
	
	// Read length prefix first
	respLengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(stream, respLengthBytes); err != nil {
		return fmt.Errorf("failed to read response length: %w", err)
	}
	
	respLength := uint32(respLengthBytes[0])<<24 | 
		uint32(respLengthBytes[1])<<16 | 
		uint32(respLengthBytes[2])<<8 | 
		uint32(respLengthBytes[3])
		
	log.Printf("Response length: %d bytes", respLength)
	
	// Read the full response
	respData := make([]byte, respLength)
	if _, err := io.ReadFull(stream, respData); err != nil {
		return fmt.Errorf("failed to read response data: %w", err)
	}

	var resp pb.GetFullStateResponse
	if err := proto.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal state response: %w", err)
	}

	log.Printf("Successfully unmarshaled response with %d operators", len(resp.Operators))
	
	// Update local state
	log.Printf("Updating local operator states...")
	node.updateOperatorStates(resp.Operators)
	
	// Print initial state table with header
	log.Printf("\n=== Initial Operator State Table ===")
	if len(resp.Operators) == 0 {
		log.Printf("No operators found in response")
	} else {
		log.Printf("Received operators:")
		for _, op := range resp.Operators {
			log.Printf("- Address: %s, Status: %s, ActiveEpoch: %d, Weight: %s", 
				op.Address, op.Status, op.ActiveEpoch, op.Weight)
		}
	}
	node.PrintOperatorStates()
	log.Printf("===================================\n")
	
	return nil
}

func (node *Node) handleStateUpdates(stream network.Stream) {
	defer stream.Close()
	log.Printf("Started handling state updates from registry")

	for {
		// Read length prefix first
		lengthBytes := make([]byte, 4)
		if _, err := io.ReadFull(stream, lengthBytes); err != nil {
			if err != io.EOF {
				log.Printf("Error reading update length: %v", err)
			} else {
				log.Printf("State update stream closed by registry")
			}
			return
		}
		
		length := uint32(lengthBytes[0])<<24 | 
			uint32(lengthBytes[1])<<16 | 
			uint32(lengthBytes[2])<<8 | 
			uint32(lengthBytes[3])
			
		// Sanity check on length
		if length > 1024*1024 { // Max 1MB
			log.Printf("Warning: Received very large update length: %d bytes, ignoring", length)
			stream.Reset()
			return
		}
			
		log.Printf("Received state update with length: %d bytes", length)
		
		// Read the full update
		data := make([]byte, length)
		if _, err := io.ReadFull(stream, data); err != nil {
			if err != io.EOF {
				log.Printf("Error reading state update data: %v", err)
			}
			return
		}

		var update pb.OperatorStateUpdate
		if err := proto.Unmarshal(data, &update); err != nil {
			log.Printf("Error unmarshaling state update: %v", err)
			continue
		}

		log.Printf("Received state update - Type: %s, Operators count: %d", update.Type, len(update.Operators))

		// Handle the update based on type
		if update.Type == "FULL" {
			log.Printf("Processing full state update...")
			node.updateOperatorStates(update.Operators)
			log.Printf("Full state update processed, total operators: %d", len(update.Operators))
		} else {
			log.Printf("Processing delta state update...")
			// For delta updates, merge with existing state
			node.mergeOperatorStates(update.Operators)
			log.Printf("Delta state update processed, updated operators: %d", len(update.Operators))
		}
	}
}

func (node *Node) updateOperatorStates(operators []*pb.OperatorState) {
	node.statesMu.Lock()
	defer node.statesMu.Unlock()

	log.Printf("Updating operator states in memory...")
	log.Printf("Current operator count before update: %d", len(node.operatorStates))
	
	// Update memory state
	node.operatorStates = make(map[string]*pb.OperatorState)
	for _, op := range operators {
		node.operatorStates[op.Address] = op
		log.Printf("Added/Updated operator in memory - Address: %s, Status: %s, ActiveEpoch: %d, Weight: %s", 
			op.Address, op.Status, op.ActiveEpoch, op.Weight)
	}
	
	log.Printf("Memory state updated - New operator count: %d", len(node.operatorStates))
	
	// Print current state of operatorStates map
	log.Printf("\nCurrent Operator States in Memory:")
	for addr, state := range node.operatorStates {
		log.Printf("Address: %s, Status: %s, ActiveEpoch: %d, Weight: %s",
			addr, state.Status, state.ActiveEpoch, state.Weight)
	}
}

func (node *Node) mergeOperatorStates(operators []*pb.OperatorState) {
	node.statesMu.Lock()
	defer node.statesMu.Unlock()

	// Update or add operators
	for _, op := range operators {
		node.operatorStates[op.Address] = op
		log.Printf("Merged operator state - Address: %s, Status: %s, ActiveEpoch: %d", 
			op.Address, op.Status, op.ActiveEpoch)
	}
	log.Printf("Memory state merged with %d operators, total operators: %d", 
		len(operators), len(node.operatorStates))
		
	// Print current state table
	log.Printf("\nOperator State Table:")
	node.PrintOperatorStates()
}

// PrintOperatorStates prints all operator states stored in memory
func (node *Node) PrintOperatorStates() {
	node.statesMu.RLock()
	defer node.statesMu.RUnlock()

	log.Printf("\nOperator States (%d total):", len(node.operatorStates))
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
	log.Printf("| %-42s | %-12s | %-12s | %-10s |", "Address", "Status", "ActiveEpoch", "Weight")
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
	
	for _, op := range node.operatorStates {
		log.Printf("| %-42s | %-12s | %-12d | %-10s |", 
			op.Address,
			op.Status,
			op.ActiveEpoch,
			op.Weight,
		)
	}
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
}

// GetOperatorState returns the state of a specific operator
func (node *Node) GetOperatorState(address string) *pb.OperatorState {
	node.statesMu.RLock()
	defer node.statesMu.RUnlock()
	return node.operatorStates[address]
}

// GetOperatorCount returns the number of operators in memory
func (node *Node) GetOperatorCount() int {
	node.statesMu.RLock()
	defer node.statesMu.RUnlock()
	return len(node.operatorStates)
} 