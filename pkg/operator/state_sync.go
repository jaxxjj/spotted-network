package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	utils "github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
)

const (
	GenesisBlock = 0
	EpochPeriod  = 12
	stateHashCheckInterval = 30 * time.Second // Interval to check state hash

	StateSyncTopic = "/spotted/state-sync"
	
	// Protocol for state verification
	StateVerifyProtocol = "/spotted/state-verify/1.0.0"
	EpochUpdateTopic = "/spotted/epoch-update"

)

type ChainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	GetMinimumWeight(ctx context.Context) (*big.Int, error)
	GetThresholdWeight(ctx context.Context) (*big.Int, error)
	GetTotalWeight(ctx context.Context) (*big.Int, error)
	GetStateAtBlock(ctx context.Context, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error)
	GetCurrentEpoch(ctx context.Context) (uint32, error)
	Close() error
}	

// StateSyncProcessor handles state sync via pubsub
type StateSyncProcessor struct {
	node             *Node
	pubsub           PubSubService
	stateSyncTopic   ResponseTopic
	epochUpdateTopic ResponseTopic
}

// NewStateSyncProcessor creates a new state sync processor
func NewStateSyncProcessor(node *Node, pubsub PubSubService) (*StateSyncProcessor, error) {
	// Join state sync topic
	stateSyncTopic, err := pubsub.Join(StateSyncTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join state sync topic: %w", err)
	}
	log.Printf("[StateSync] Joined state sync topic: %s", StateSyncTopic)

	// Join epoch update topic
	epochUpdateTopic, err := pubsub.Join(EpochUpdateTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join epoch update topic: %w", err)
	}
	log.Printf("[StateSync] Joined epoch update topic: %s", EpochUpdateTopic)

	sp := &StateSyncProcessor{
		node:             node,
		pubsub:           pubsub,
		stateSyncTopic:   stateSyncTopic,
		epochUpdateTopic: epochUpdateTopic,
	}

	// Subscribe to state sync topic
	stateSub, err := stateSyncTopic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to state sync topic: %w", err)
	}
	log.Printf("[StateSync] Subscribed to state sync topic")

	// Subscribe to epoch update topic
	epochSub, err := epochUpdateTopic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to epoch update topic: %w", err)
	}
	log.Printf("[StateSync] Subscribed to epoch update topic")

	// Start handling updates
	go sp.handlePubSubStateUpdates(stateSub)
	go sp.handleEpochUpdates(epochSub)

	// Start state hash verification
	go sp.startStateHashVerification()

	return sp, nil
}

// handlePubSubStateUpdates handles incoming state updates from the topic
func (sp *StateSyncProcessor) handleStateSyncTopic(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(context.Background())
		if err != nil {
			log.Printf("[StateSync] Failed to get next state update message: %v", err)
			continue
		}

		var update pb.FullStateSync
		if err := proto.Unmarshal(msg.Data, &update); err != nil {
			log.Printf("[StateSync] Failed to unmarshal state update: %v", err)
			continue
		}

		// Update local state
		sp.updateOperatorStates(update.Operators)
		log.Printf("[StateSync] Updated operator states with %d operators", len(update.Operators))
	}
}

// handleEpochUpdates handles incoming epoch updates from the topic
func (sp *StateSyncProcessor) handleEpochUpdateTopic(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(context.Background())
		if err != nil {
			log.Printf("[StateSync] Failed to get next epoch update message: %v", err)
			continue
		}

		var update pb.FullStateSync
		if err := proto.Unmarshal(msg.Data, &update); err != nil {
			log.Printf("[StateSync] Failed to unmarshal epoch update: %v", err)
			continue
		}

		// Update local state with epoch update
		sp.updateOperatorStates(update.Operators)
		log.Printf("[StateSync] Updated operator states for epoch with %d operators", len(update.Operators))

		// Handle epoch specific logic here if needed
		// For example: reset epoch-related counters, update epoch number, etc.
	}
}

// updateOperatorStates updates the local operator states
func (sp *StateSyncProcessor) updateOperatorStates(operators []*pb.OperatorPeerState) {
	sp.node.activePeers.mu.Lock()
	defer sp.node.activePeers.mu.Unlock()

	// Clear existing state
	sp.node.activePeers.active = make(map[peer.ID]*OperatorState)

	// Update with new state
	for _, op := range operators {
		peerID, err := peer.Decode(op.PeerId)
		if err != nil {
			log.Printf("[StateSync] Failed to decode peer ID %s: %v", op.PeerId, err)
			continue
		}

		multiaddrs := make([]multiaddr.Multiaddr, 0, len(op.Multiaddrs))
		for _, addr := range op.Multiaddrs {
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				log.Printf("[StateSync] Failed to parse multiaddr %s for peer %s: %v", addr, op.PeerId, err)
				continue
			}
			multiaddrs = append(multiaddrs, maddr)
		}

		sp.node.activePeers.active[peerID] = &OperatorState{
			PeerID:     peerID,
			Multiaddrs: multiaddrs,
			Address:    op.Address,
			SigningKey: op.SigningKey,
			Weight:     utils.StringToBigInt(op.Weight),
		}
	}
}

// handleStreamStateUpdate handles state update from stream
func (sp *StateSyncProcessor) handleStreamStateUpdate(stream network.Stream) error {
	// Read update message
	data, err := utils.ReadLengthPrefixedDataWithDeadline(stream, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to read state update: %w", err)
	}

	var update pb.FullStateSync
	if err := proto.Unmarshal(data, &update); err != nil {
		return fmt.Errorf("failed to unmarshal state update: %w", err)
	}

	// Update local state
	sp.updateOperatorStates(update.Operators)
	return nil
}

// startStateHashVerification starts periodic state hash verification
func (sp *StateSyncProcessor) startStateHashVerification() {
	ticker := time.NewTicker(stateHashCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sp.verifyStateHash(); err != nil {
				log.Printf("[StateSync] Failed to verify state hash: %v", err)
			}
		}
	}
}

// verifyStateHash verifies local state hash with registry
func (sp *StateSyncProcessor) verifyStateHash() error {
	// Open stream to registry
	stream, err := sp.node.host.NewStream(context.Background(), sp.node.registryID, StateSyncProtocol)
	if err != nil {
		return fmt.Errorf("failed to open state hash stream: %w", err)
	}
	defer stream.Close()

	// Calculate current state hash
	stateHash := sp.node.GetActiveOperatorsRoot()

	// Send state hash request
	if _, err := stream.Write([]byte{MsgTypeStateHash}); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Send hash
	if err := utils.WriteLengthPrefixedDataWithDeadline(stream, []byte(stateHash), 5*time.Second); err != nil {
		return fmt.Errorf("failed to write state hash: %w", err)
	}

	// Read response
	respData, err := utils.ReadLengthPrefixedDataWithDeadline(stream, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var resp pb.StateHashResponse
	if err := proto.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// If hash matches, we're done
	if resp.HashMatched {
		log.Printf("[StateSync] State hash verified successfully")
		return nil
	}

	// If hash doesn't match and we got a full state, update our state
	if resp.FullState != nil {
		sp.updateOperatorStates(resp.FullState.Operators)
		log.Printf("[StateSync] Updated state from hash mismatch with %d operators", len(resp.FullState.Operators))
	}

	return nil
}







