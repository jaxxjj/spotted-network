package registry

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	utils "github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"google.golang.org/protobuf/proto"
)

const (
	maxMessageSize    = 1024 * 1024 // 1MB
	stateSyncInterval = 2 * time.Minute

	// Topic for broadcasting node updates
	StateSyncTopic = "/spotted/state-sync"
	
	// Protocol for state verification
	StateVerifyProtocol = protocol.ID("/spotted/state-verify/1.0.0")
)

// PubSubService defines the interface for pubsub functionality
type PubSubService interface {
	Join(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error)
}

// ResponseTopic defines the interface for topic functionality
type ResponseTopic interface {
	// Subscribe returns a new subscription for the topic
	Subscribe(opts ...pubsub.SubOpt) (*pubsub.Subscription, error)
	// Publish publishes data to the topic
	Publish(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error
	// ListPeers returns the peer IDs of peers in the topic
	ListPeers() []peer.ID
	// String returns the string representation of the topic
	String() string
}

// StateSyncProcessor handles state verification via stream and broadcasts updates via pubsub
type StateSyncProcessor struct {
	node           *Node
	pubsub         PubSubService
	stateSyncTopic ResponseTopic
}

// NewStateSyncProcessor creates a new state sync processor
func NewStateSyncProcessor(node *Node, pubsub PubSubService) (*StateSyncProcessor, error) {
	// Join state sync topic for broadcasting updates
	stateSyncTopic, err := pubsub.Join(StateSyncTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join state sync topic: %w", err)
	}
	log.Printf("[StateSync] Joined state sync topic: %s", StateSyncTopic)

	return &StateSyncProcessor{
		node:           node,
		pubsub:         pubsub,
		stateSyncTopic: stateSyncTopic,
	}, nil
}

// handleStateVerifyStream handles incoming state verification requests
func (sp *StateSyncProcessor) handleStateVerifyStream(stream network.Stream) {
	defer stream.Close()
	
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("[StateVerify] New state verification request from %s", remotePeer)

	// Read and parse message
	var msg pb.StateVerifyMessage
	if err := utils.ReadStreamMessage(stream, &msg, maxMessageSize); err != nil {
		log.Printf("[StateVerify] Failed to read message: %v", err)
		stream.Reset()
		return
	}

	// Verify state root
	matches := bytes.Equal(sp.node.getActiveOperatorsRoot(), msg.StateRoot)
	
	resp := &pb.StateVerifyResponse{
		Success: matches,
		Message: "State verification completed",
	}
	
	// If root doesn't match, include full state
	if !matches {
		resp.Operators = sp.node.BuildOperatorPeerStates()
		log.Printf("[StateVerify] State root mismatch for %s, sending full state", remotePeer)
	} else {
		log.Printf("[StateVerify] State root verified for %s", remotePeer)
	}
	
	// Send response
	if err := utils.WriteStreamMessage(stream, resp); err != nil {
		log.Printf("[StateVerify] Failed to send verification response: %v", err)
		stream.Reset()
		return
	}
}

// BroadcastStateUpdate broadcasts state update to all operators via pubsub
func (sp *StateSyncProcessor) BroadcastStateUpdate() error {
	// Get current active operators
	operators := sp.node.GetActiveOperators()
	
	// Create state update message
	update := &pb.FullStateSync{
		Operators: operators,
	}
	
	// Marshal update
	data, err := proto.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal state update: %w", err)
	}
	
	// Publish update
	if err := sp.stateSyncTopic.Publish(context.Background(), data); err != nil {
		return fmt.Errorf("failed to publish state update: %w", err)
	}
	
	log.Printf("[StateSync] Broadcasted state update with %d operators", len(operators))
	return nil
}

// GetActiveOperators returns list of active operators
func (n *Node) GetActiveOperators() []*pb.OperatorPeerState {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()

	operators := make([]*pb.OperatorPeerState, 0, len(n.activeOperators.active))
	for peerID, state := range n.activeOperators.active {
		// Get operator info from DB
		operator, err := n.opQuerier.GetOperatorByAddress(context.Background(), state.Address)
		if err != nil {
			log.Printf("[StateVerify] Failed to get operator for address %s: %v", state.Address, err)
			continue
		}

		op := &pb.OperatorPeerState{
			PeerId:     peerID.String(),
			Multiaddrs: make([]string, len(state.Multiaddrs)),
			Address:    state.Address,
			SigningKey: operator.SigningKey,
			Weight:     utils.NumericToString(operator.Weight),
		}
		for i, addr := range state.Multiaddrs {
			op.Multiaddrs[i] = addr.String()
		}
		operators = append(operators, op)
	}

	return operators
}

// BuildOperatorStates builds a list of all active operator states
func (n *Node) BuildOperatorPeerStates() []*pb.OperatorPeerState {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()

	operators := make([]*pb.OperatorPeerState, 0, len(n.activeOperators.active))
	for peerID, peerInfo := range n.activeOperators.active {
		operator, err := n.opQuerier.GetOperatorByAddress(context.Background(), peerInfo.Address)
		if err != nil {
			log.Printf("[StateVerify] Failed to get operator for address %s: %v", peerInfo.Address, err)
			continue
		}

		op := &pb.OperatorPeerState{
			PeerId:     peerID.String(),
			Multiaddrs: utils.MultiaddrsToStrings(peerInfo.Multiaddrs),
			Address:    peerInfo.Address,
			SigningKey: operator.SigningKey,
			Weight:     utils.NumericToString(operator.Weight),
		}
		operators = append(operators, op)
	}

	return operators
}