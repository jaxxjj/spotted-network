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
	BlacklistPeer(peer.ID)
}

// ResponseTopic defines the interface for topic functionality
type PubsubTopic interface {
	Publish(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error
}

type PubsubConfig struct {
	node *Node
	pubsub PubSubService
}

// StateSyncProcessor handles state verification via stream and broadcasts updates via pubsub
type StateSyncProcessor struct {
	node           *Node
	pubsub         PubSubService
	stateSyncTopic PubsubTopic
}

// NewStateSyncProcessor creates a new state sync processor
func NewStateSyncProcessor(cfg *PubsubConfig) (*StateSyncProcessor, error) {
	if cfg.node == nil {
		return nil, fmt.Errorf("node is nil")
	}
	if cfg.pubsub == nil {
		return nil, fmt.Errorf("pubsub is nil")
	}
	// Join state sync topic for broadcasting updates
	stateSyncTopic, err := cfg.pubsub.Join(StateSyncTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join state sync topic: %w", err)
	}
	log.Printf("[StateSync] Joined state sync topic: %s", StateSyncTopic)

	return &StateSyncProcessor{
		node:           cfg.node,
		pubsub:         cfg.pubsub,
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
		resp.Operators = sp.node.buildOperatorPeerStates()
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
func (sp *StateSyncProcessor) broadcastStateUpdate(epoch *uint32) error {
	// Get current active operators
	operators := sp.node.buildOperatorPeerStates()
	var update *pb.FullStateSync
	if epoch != nil {
		epochU64 := uint64(*epoch)
		update = &pb.FullStateSync{
			Operators: operators,
			Epoch:     &epochU64,
		}
	} else {
		update = &pb.FullStateSync{
			Operators: operators,
		}
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

// BuildOperatorStates builds a list of all active operator states
func (n *Node) buildOperatorPeerStates() []*pb.OperatorPeerState {
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

// Stop cleans up state sync processor resources
func (sp *StateSyncProcessor) Stop() {
	sp.stateSyncTopic = nil
	log.Printf("[StateSync] State sync processor stopped")
}