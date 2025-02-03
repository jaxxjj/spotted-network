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
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
)

const (
	GenesisBlock = 0
	EpochPeriod  = 12
	stateHashCheckInterval = 30 * time.Second // Interval to check state hash
	
	// Protocol IDs
	StateSyncProtocol = protocol.ID("/spotted/state-sync/1.0.0")
	StateSyncTopic = "/spotted/state-sync"
	StateVerifyProtocol = protocol.ID("/spotted/state-verify/1.0.0")
)

type OperatorStateController interface {
	UpsertActivePeerState(state *OperatorState) error
	GetActiveOperatorsRoot() []byte
	GetOperatorState(id peer.ID) *OperatorState
	RemoveOperator(id peer.ID)
	createStreamToRegistry(ctx context.Context, pid protocol.ID) (network.Stream, error)
	getCurrentStateRoot() []byte
	UpsertActivePeerStates(states []*OperatorState) error
}

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
	node             OperatorStateController
	pubsub           PubSubService
	stateSyncTopic   ResponseTopic
}

// NewStateSyncProcessor creates a new state sync processor
func NewStateSyncProcessor(node OperatorStateController, pubsub PubSubService) (*StateSyncProcessor, error) {
	// Join state sync topic
	stateSyncTopic, err := pubsub.Join(StateSyncTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join state sync topic: %w", err)
	}
	log.Printf("[StateSync] Joined state sync topic: %s", StateSyncTopic)

	sp := &StateSyncProcessor{
		node:             node,
		pubsub:           pubsub,
		stateSyncTopic:   stateSyncTopic,
	}

	// Subscribe to state sync topic
	stateSub, err := stateSyncTopic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to state sync topic: %w", err)
	}
	log.Printf("[StateSync] Subscribed to state sync topic")

	// Start handling updates
	go sp.handleStateSyncTopic(stateSub)

	// Start state hash verification
	go sp.startStateHashVerification()

	return sp, nil
}

// handleStateSyncTopic handles incoming state updates from pubsub
func (sp *StateSyncProcessor) handleStateSyncTopic(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(context.Background())
		if err != nil {
			log.Printf("[StateSync] Failed to get next message: %v", err)
			continue
		}

		// Parse state update
		var update pb.FullStateSync
		if err := proto.Unmarshal(msg.Data, &update); err != nil {
			log.Printf("[StateSync] Failed to unmarshal state update: %v", err)
			continue
		}

		// Process operator states
		states := make([]*OperatorState, 0, len(update.Operators))
		for _, opState := range update.Operators {
			state, err := sp.convertToOperatorState(opState)
			if err != nil {
				continue
			}
			states = append(states, state)
		}

		// Batch update all states
		if len(states) > 0 {
			if err := sp.node.UpsertActivePeerStates(states); err != nil {
				log.Printf("[StateSync] Failed to batch update operator states: %v", err)
			}
		}
	}
}

// startStateHashVerification periodically verifies state hash with registry
func (sp *StateSyncProcessor) startStateHashVerification() {
	ticker := time.NewTicker(stateHashCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sp.verifyStateWithRegistry(); err != nil {
				log.Printf("[StateSync] Failed to verify state: %v", err)
			}
		}
	}
}

// verifyStateWithRegistry verifies local state hash with registry
func (sp *StateSyncProcessor) verifyStateWithRegistry() error {
	// Create stream to registry
	stream, err := sp.node.createStreamToRegistry(context.Background(), StateVerifyProtocol)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Send verify request
	req := &pb.StateVerifyMessage{
		StateRoot: sp.node.getCurrentStateRoot(),
	}
	if err := utils.WriteStreamMessage(stream, req); err != nil {
		return fmt.Errorf("failed to write verify request: %w", err)
	}

	// Read response
	var resp pb.StateVerifyResponse
	if err := utils.ReadStreamMessage(stream, &resp, maxMessageSize); err != nil {
		return fmt.Errorf("failed to read verify response: %w", err)
	}

	// If state mismatch, update local state
	if !resp.Success {
		log.Printf("[StateSync] State mismatch detected, updating local state...")
		states := make([]*OperatorState, 0, len(resp.Operators))
		
		for _, opState := range resp.Operators {
			state, err := sp.convertToOperatorState(opState)
			if err != nil {
				continue
			}
			states = append(states, state)
		}

		// Batch update all states
		if len(states) > 0 {
			if err := sp.node.UpsertActivePeerStates(states); err != nil {
				log.Printf("[StateSync] Failed to batch update operator states: %v", err)
			}
		}
	}

	return nil
}

// convertToOperatorState converts a protobuf operator state to internal operator state
func (sp *StateSyncProcessor) convertToOperatorState(opState *pb.OperatorPeerState) (*OperatorState, error) {
	// Convert multiaddrs
	addrs := make([]multiaddr.Multiaddr, 0, len(opState.Multiaddrs))
	for _, addrStr := range opState.Multiaddrs {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			log.Printf("[StateSync] Failed to parse multiaddr %s: %v", addrStr, err)
			continue
		}
		addrs = append(addrs, addr)
	}

	// Parse peer ID
	peerID, err := peer.Decode(opState.PeerId)
	if err != nil {
		log.Printf("[StateSync] Failed to decode peer ID %s: %v", opState.PeerId, err)
		return nil, err
	}

	// Create operator state
	state := &OperatorState{
		PeerID:     peerID,
		Multiaddrs: addrs,
		Address:    opState.Address,
		SigningKey: opState.SigningKey,
		Weight:     utils.StringToBigInt(opState.Weight),
	}

	return state, nil
}











