package registry

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

const (
	heartbeatInterval = 10 * time.Second
	heartbeatTimeout  = 30 * time.Second
	maxMessageSize    = 1024 * 1024 // 1MB
)

// State sync message types
const (
	MsgTypeGetFullState    byte = 0x01
	MsgTypeSubscribe       byte = 0x02
	MsgTypeStateUpdate     byte = 0x03
	MsgTypeHeartbeat       byte = 0x04
	MsgTypeHeartbeatResp   byte = 0x05
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

	// Read message type
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		log.Printf("[StateSync] Error reading message type from %s: %v", remotePeer, err)
		stream.Reset()
		return
	}

	switch msgType[0] {
	case MsgTypeGetFullState:
		log.Printf("[StateSync] Processing GetFullState request from %s", remotePeer)
		n.handleGetFullState(stream)
	case MsgTypeSubscribe:
		log.Printf("[StateSync] Processing Subscribe request from %s", remotePeer)
		n.handleSubscribe(stream)
	default:
		log.Printf("[StateSync] Unknown message type from %s: 0x%02x", remotePeer, msgType[0])
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

	// Write length prefix and data
	if err := commonHelpers.WriteLengthPrefix(stream, uint32(len(data))); err != nil {
		log.Printf("[StateSync] Failed to write length prefix: %v", err)
		stream.Reset()
		return
	}

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

	// Add new subscriber
	n.subscribers[peerID] = stream
	log.Printf("[StateSync] New subscriber added: %s", peerID)
}

func (n *Node) handleSubscribe(stream network.Stream) {
	peer := stream.Conn().RemotePeer()
	log.Printf("[StateSync] Handling subscribe request from peer %s", peer)

	// Read request length and data
	length, err := commonHelpers.ReadLengthPrefix(stream)
	if err != nil {
		log.Printf("[StateSync] Failed to read length prefix from %s: %v", peer, err)
		stream.Reset()
		return
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		log.Printf("[StateSync] Failed to read request data from %s: %v", peer, err)
		stream.Reset()
		return
	}

	// Parse request
	var req pb.SubscribeRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Printf("[StateSync] Failed to unmarshal subscribe request from %s: %v", peer, err)
		stream.Reset()
		return
	}
	
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
			case MsgTypeHeartbeatResp: // Heartbeat response
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
	if err := stream.SetWriteDeadline(time.Now().Add(heartbeatTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline for %s: %w", peer, err)
	}
	defer stream.SetWriteDeadline(time.Time{})
	
	// Send heartbeat message
	if _, err := stream.Write([]byte{MsgTypeHeartbeat}); err != nil {
		return fmt.Errorf("failed to send heartbeat to %s: %w", peer, err)
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

	n.subscribersMu.RLock()
	defer n.subscribersMu.RUnlock()

	for peer, stream := range n.subscribers {
		// Set write deadline
		if err := stream.SetWriteDeadline(time.Now().Add(heartbeatTimeout)); err != nil {
			log.Printf("[StateSync] Error setting write deadline for %s: %v", peer, err)
			n.Unsubscribe(peer)
			continue
		}

		// Write message type
		if _, err := stream.Write([]byte{MsgTypeStateUpdate}); err != nil {
			log.Printf("[StateSync] Error writing message type to %s: %v", peer, err)
			n.Unsubscribe(peer)
			continue
		}

		// Write length prefix and data
		if err := commonHelpers.WriteLengthPrefix(stream, uint32(len(data))); err != nil {
			log.Printf("[StateSync] Error writing length prefix to %s: %v", peer, err)
			n.Unsubscribe(peer)
			continue
		}

		if _, err := stream.Write(data); err != nil {
			log.Printf("[StateSync] Error writing update data to %s: %v", peer, err)
			n.Unsubscribe(peer)
			continue
		}

		log.Printf("[StateSync] Sent state update to %s - Type: %s, Length: %d, Operators: %d", 
			peer, updateType, len(data), len(operators))
	}
}

// Helper functions

func convertToProtoOperator(operator operators.Operators) *pb.OperatorState {
	return &pb.OperatorState{
		Address:                 operator.Address,
		SigningKey:             operator.SigningKey,
		RegisteredAtBlockNumber: operator.RegisteredAtBlockNumber,
		RegisteredAtTimestamp:   operator.RegisteredAtTimestamp,
		ActiveEpoch:            operator.ActiveEpoch,
		ExitEpoch:              &operator.ExitEpoch,
		Status:                 string(operator.Status),
		Weight:                 commonHelpers.NumericToString(operator.Weight),
	}
}

