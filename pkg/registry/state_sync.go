package registry

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

const (
	heartbeatInterval = 10 * time.Second
	heartbeatTimeout  = 30 * time.Second
	maxMessageSize   = 1024 * 1024
)

// Start initializes the state sync service
func (n *Node) startStateSync() error {
	// Set up protocol handler for state sync requests
	n.host.SetStreamHandler("/state-sync/1.0.0", n.handleStateSync)
	
	log.Printf("[StateSync] Service started successfully")
	return nil
}

// handleStateSync handles incoming state sync requests
func (n *Node) handleStateSync(stream network.Stream) {
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("[StateSync] New state sync request from %s", remotePeer)

	msgType := make([]byte, 1)
	if _, err := stream.Read(msgType); err != nil {
		log.Printf("[StateSync] Error reading message type from %s: %v", remotePeer, err)
		stream.Reset()
		return
	}

	switch msgType[0] {
	case 0x01:
		log.Printf("[StateSync] Processing GetFullState request from %s", remotePeer)
		n.handleGetFullState(stream)
	case 0x02:
		log.Printf("[StateSync] Processing Subscribe request from %s", remotePeer)
		n.handleSubscribe(stream)
	default:
		log.Printf("[StateSync] Unknown message type from %s: %x", remotePeer, msgType[0])
		stream.Reset()
	}
}

// handleGetFullState handles a request for full state
func (n *Node) handleGetFullState(stream network.Stream) {
	// Get all operators from database
	allOperators, err := n.operators.ListAllOperators(context.Background())
	if err != nil {
		log.Printf("[StateSync] Failed to get operators: %v", err)
		stream.Reset()
		return
	}

	log.Printf("[StateSync] Found %d total operators", len(allOperators))

	// Create response
	resp := &pb.GetFullStateResponse{
		Operators: make([]*pb.OperatorState, len(allOperators)),
	}
	for i, op := range allOperators {
		resp.Operators[i] = convertToProtoOperator(op)
	}

	// Marshal response
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("[StateSync] Failed to marshal response: %v", err)
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
		log.Printf("[StateSync] Failed to write length prefix: %v", err)
		stream.Reset()
		return
	}

	// Write response
	if _, err := stream.Write(data); err != nil {
		log.Printf("[StateSync] Failed to write response: %v", err)
		stream.Reset()
		return
	}

	log.Printf("[StateSync] Sent full state with %d operators", len(allOperators))
}


func (n *Node) Subscribe(peerID peer.ID, stream network.Stream) {
	n.subscribersMu.Lock()
	defer n.subscribersMu.Unlock()

	// Close existing stream
	if existingStream, ok := n.subscribers[peerID]; ok {
		existingStream.Reset()
		delete(n.subscribers, peerID)
	}

	// Add new subscriber without setting fixed deadlines
	n.subscribers[peerID] = stream
	log.Printf("[StateSync] New subscriber added: %s", peerID)
}

func (n *Node) handleSubscribe(stream network.Stream) {
	peer := stream.Conn().RemotePeer()
	log.Printf("[StateSync] Handling subscribe request from peer %s", peer)
	
	n.Subscribe(peer, stream)
	
	// Start heartbeat check
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	// Read buffer for message type
	msgType := make([]byte, 1)

	// Set initial read deadline
	stream.SetReadDeadline(time.Now().Add(heartbeatTimeout))

	for {
		select {
		case <-heartbeat.C:
			// Send heartbeat
			if err := n.sendHeartbeat(peer, stream); err != nil {
				log.Printf("[StateSync] Heartbeat failed for %s: %v", peer, err)
				n.Unsubscribe(peer)
				return
			}
			// Update read deadline after successful heartbeat
			stream.SetReadDeadline(time.Now().Add(heartbeatTimeout))
		default:
			// Read message type
			if _, err := io.ReadFull(stream, msgType); err != nil {
				if err != io.EOF {
					log.Printf("[StateSync] Error reading message type from %s: %v", peer, err)
				} else {
					log.Printf("[StateSync] Connection closed by %s", peer)
				}
				n.Unsubscribe(peer)
				return
			}

			// Handle different message types
			switch msgType[0] {
			case 0x04: // Heartbeat response
				log.Printf("[StateSync] Received heartbeat response from %s", peer)
				// Update read deadline after receiving heartbeat response
				stream.SetReadDeadline(time.Now().Add(heartbeatTimeout))
			default:
				log.Printf("[StateSync] Received unknown message type 0x%02x from %s", msgType[0], peer)
				n.Unsubscribe(peer)
				return
			}

			// Small sleep to prevent tight loop
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (n *Node) sendHeartbeat(peer peer.ID, stream network.Stream) error {
	// Set write deadline
	stream.SetWriteDeadline(time.Now().Add(heartbeatTimeout))
	
	// Send heartbeat message (0x03 represents heartbeat message)
	_, err := stream.Write([]byte{0x03})
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	
	return nil
}

// Unsubscribe removes a subscriber
func (n *Node) Unsubscribe(peerID peer.ID) {
	n.subscribersMu.Lock()
	defer n.subscribersMu.Unlock()

	if stream, ok := n.subscribers[peerID]; ok {
		stream.Reset()
		delete(n.subscribers, peerID)
		log.Printf("[StateSync] Subscriber removed: %s", peerID)
	}
}

func (n *Node) BroadcastStateUpdate(operators []*pb.OperatorState, updateType string) {
	update := &pb.OperatorStateUpdate{
		Type:      updateType,
		Operators: operators,
	}

	data, err := proto.Marshal(update)
	if err != nil {
		log.Printf("[StateSync] Error marshaling state update: %v", err)
		return
	}

	// Validate message size
	if len(data) > maxMessageSize {
		log.Printf("[StateSync] State update too large (%d bytes), max allowed size is %d bytes", len(data), maxMessageSize)
		return
	}

	// Prepare message
	length := uint32(len(data))
	message := make([]byte, 5+len(data)) // 1 byte type + 4 bytes length + data
	message[0] = 0x02 // State update type
	message[1] = byte(length >> 24)
	message[2] = byte(length >> 16)
	message[3] = byte(length >> 8)
	message[4] = byte(length)
	copy(message[5:], data)

	n.subscribersMu.RLock()
	defer n.subscribersMu.RUnlock()

	for peer, stream := range n.subscribers {
		// Set write deadline
		if err := stream.SetWriteDeadline(time.Now().Add(heartbeatTimeout)); err != nil {
			log.Printf("[StateSync] Error setting write deadline for %s: %v", peer, err)
			n.Unsubscribe(peer)
			continue
		}

		// Write entire message in one call
		if _, err := stream.Write(message); err != nil {
			log.Printf("[StateSync] Error sending update to %s: %v", peer, err)
			n.Unsubscribe(peer)
			continue
		}

		log.Printf("[StateSync] Sent state update to %s - Type: %s, Length: %d, Operators: %d", 
			peer, updateType, length, len(operators))
	}
}

// Helper functions

func convertToProtoOperator(operator operators.Operators) *pb.OperatorState {

	return &pb.OperatorState{
		Address:                operator.Address,
		SigningKey:            operator.SigningKey,
		RegisteredAtBlockNumber: operator.RegisteredAtBlockNumber,
		RegisteredAtTimestamp:   operator.RegisteredAtTimestamp,
		ActiveEpoch:        operator.ActiveEpoch,
		ExitEpoch:              &operator.ExitEpoch,
		Status:                 string(operator.Status),
		Weight:                 common.NumericToString(operator.Weight),
	}
}

