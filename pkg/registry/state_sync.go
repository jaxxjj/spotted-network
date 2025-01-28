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
	heartbeatInterval = 30 * time.Second
	heartbeatTimeout  = 10 * time.Second
)

// Start initializes the state sync service
func (n *Node) startStateSync(ctx context.Context) error {
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

	// Set read and write deadlines
	stream.SetReadDeadline(time.Now().Add(heartbeatTimeout))
	stream.SetWriteDeadline(time.Now().Add(heartbeatTimeout))

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

	// Read buffer
	buf := make([]byte, 1024)

	for {
		select {
		case <-heartbeat.C:
			// Send heartbeat
			if err := n.sendHeartbeat(peer, stream); err != nil {
				log.Printf("[StateSync] Heartbeat failed for %s: %v", peer, err)
				n.Unsubscribe(peer)
				return
			}
		default:
			// Set read deadline
			stream.SetReadDeadline(time.Now().Add(heartbeatTimeout))
			
			// Read data
			_, err := stream.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[StateSync] Error reading from subscriber %s: %v", peer, err)
				} else {
					log.Printf("[StateSync] Subscriber %s closed connection", peer)
				}
				n.Unsubscribe(peer)
				return
			}
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
		log.Printf("Error marshaling state update: %v", err)
		return
	}

	// Prepare message type and length prefix
	msgType := []byte{0x02} // 0x02 for state update
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)

	n.subscribersMu.RLock()
	defer n.subscribersMu.RUnlock()

	for peer, stream := range n.subscribers {
		// Set write deadline
		if err := stream.SetWriteDeadline(time.Now().Add(heartbeatTimeout)); err != nil {
			log.Printf("Error setting write deadline for %s: %v", peer, err)
			n.Unsubscribe(peer)
			continue
		}

		// Write message type
		if _, err := stream.Write(msgType); err != nil {
			log.Printf("Error sending message type to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
			continue
		}

		// Write length prefix
		if _, err := stream.Write(lengthBytes); err != nil {
			log.Printf("Error sending length prefix to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
			continue
		}

		// Write protobuf data
		if _, err := stream.Write(data); err != nil {
			log.Printf("Error sending update data to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
		} else {
			log.Printf("Sent state update to %s - Type: %s, Length: %d, Operators: %d", 
				peer, updateType, length, len(operators))
		}
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

