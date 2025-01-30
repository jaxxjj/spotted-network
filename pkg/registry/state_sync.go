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
	maxMessageSize    = 1024 * 1024 // 1MB
	peerSyncInterval  = 2 * time.Minute
)

// State sync message types
const (
	MsgTypeGetFullState    byte = 0x01
	MsgTypeSubscribe       byte = 0x02
	MsgTypeStateUpdate     byte = 0x03
	MsgTypePeerSync        byte = 0x04  // New message type for peer sync
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

	// Write length-prefixed data with deadline
	if err := commonHelpers.WriteLengthPrefixedDataWithDeadline(stream, data, 5*time.Second); err != nil {
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

	// Read request with deadline
	data, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, 5*time.Second)
	if err != nil {
		log.Printf("[StateSync] Failed to read request from %s: %v", peer, err)
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
	log.Printf("[StateSync] Successfully subscribed peer %s", peer)

	// Start peer sync ticker with longer interval
	peerSyncTicker := time.NewTicker(peerSyncInterval)
	defer peerSyncTicker.Stop()

	// Send initial peer sync
	if err := n.syncPeersToOperator(stream); err != nil {
		log.Printf("[StateSync] Failed to send initial peer sync: %v", err)
	}

	// Keep reading from the stream to detect when it's closed
	buf := make([]byte, 1)
	for {
		select {
		case <-peerSyncTicker.C:
			// Send periodic peer sync
			if err := n.syncPeersToOperator(stream); err != nil {
				log.Printf("[StateSync] Failed to sync peers: %v", err)
				continue
			}
		default:
			_, err := stream.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[StateSync] Error reading from stream: %v", err)
				} else {
					log.Printf("[StateSync] Stream closed by peer %s", peer)
				}
				n.Unsubscribe(peer)
				return
			}
		}
	}
}

// syncPeersToOperator sends current peer information to a single operator
func (n *Node) syncPeersToOperator(stream network.Stream) error {
	// Get current operator info
	n.operatorsInfoMu.RLock()
	peers := make([]*pb.PeerInfo, 0, len(n.operatorsInfo))
	for peerID, info := range n.operatorsInfo {
		peers = append(peers, &pb.PeerInfo{
			PeerId:     peerID.String(),
			Multiaddrs: make([]string, len(info.Addrs)),
			Status:     info.Status,
			LastSeen:   info.LastSeen.Unix(),
		})
		for i, addr := range info.Addrs {
			peers[len(peers)-1].Multiaddrs[i] = addr.String()
		}
	}
	n.operatorsInfoMu.RUnlock()

	// Create peer sync message
	msg := &pb.PeerSyncMessage{
		Peers: peers,
	}

	// Marshal message
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal peer sync message: %w", err)
	}

	// Write message type
	if _, err := stream.Write([]byte{MsgTypePeerSync}); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Write length-prefixed data with deadline
	if err := commonHelpers.WriteLengthPrefixedDataWithDeadline(stream, data, 5*time.Second); err != nil {
		return fmt.Errorf("failed to write peer sync data: %w", err)
	}

	log.Printf("[StateSync] Sent peer sync with %d operators", len(peers))
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

