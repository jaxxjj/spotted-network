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
	"github.com/libp2p/go-libp2p/core/protocol"
	"google.golang.org/protobuf/proto"
)

const (
	GenesisBlock = 0
	EpochPeriod  = 12
	stateHashCheckInterval = 30 * time.Second // Interval to check state hash
	
	StateSyncTopic = "/spotted/state-sync"
	StateVerifyProtocol = protocol.ID("/spotted/state-verify/1.0.0")
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
}

// NewStateSyncProcessor creates a new state sync processor
func NewStateSyncProcessor(ctx context.Context, node *Node, pubsub PubSubService) (*StateSyncProcessor, error) {
	sp := &StateSyncProcessor{
		node:             node,
		pubsub:           pubsub,
	}

	// Subscribe to state sync topic
	stateSub, err := sp.SubscribeToStateSyncTopic()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to state sync topic: %w", err)
	}

	// Set up handler for state verify responses from registry
	node.host.SetStreamHandler(StateVerifyProtocol, sp.handleStateVerifyStream)
	
	// Start handling updates
	go sp.handleStateSyncTopic(ctx, stateSub)

	// Start state hash verification
	go sp.startStateHashVerification(ctx)

	return sp, nil
}

// Subscribe subscribes to the state sync topic
func (sp *StateSyncProcessor) SubscribeToStateSyncTopic() (*pubsub.Subscription, error) {
	// Join state sync topic if not already joined
	if sp.stateSyncTopic == nil {
		topic, err := sp.pubsub.Join(StateSyncTopic)
		if err != nil {
			return nil, fmt.Errorf("failed to join state sync topic: %w", err)
		}
		sp.stateSyncTopic = topic
		log.Printf("[StateSync] Joined state sync topic: %s", StateSyncTopic)
	}

	// Subscribe to the topic
	sub, err := sp.stateSyncTopic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to state sync topic: %w", err)
	}
	log.Printf("[StateSync] Subscribed to state sync topic")

	return sub, nil
}

// handleStateSyncTopic handles incoming state updates from pubsub
func (sp *StateSyncProcessor) handleStateSyncTopic(ctx context.Context, sub *pubsub.Subscription) {
	log.Printf("[StateSync] Started handling state sync topic")
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("[StateSync] Stopping state sync topic handler: %v", ctx.Err())
			return
		default:
			msg, err := sub.Next(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("[StateSync] Context cancelled while getting next message: %v", ctx.Err())
					return
				}
				log.Printf("[StateSync] Failed to get next message: %v", err)
				continue
			}
			
			log.Printf("[StateSync] Received message from peer: %s", msg.ReceivedFrom.String())
			
			// 在goroutine中处理消息
			go func(msg *pubsub.Message) {
				if err := sp.handleStateUpdateMessage(ctx, msg); err != nil {
					log.Printf("[StateSync] Failed to handle state update message: %v", err)
				}
			}(msg)
		}
	}
}

// handleStateUpdateMessage handles a single state update message
func (sp *StateSyncProcessor) handleStateUpdateMessage(ctx context.Context, msg *pubsub.Message) error {
	// Parse state update
	var update pb.FullStateSync
	if err := proto.Unmarshal(msg.Data, &update); err != nil {
		return fmt.Errorf("failed to unmarshal state update: %w", err)
	}
	
	log.Printf("[StateSync] Processing state update with %d operators", len(update.Operators))
	
	// Process operator states
	states := make([]*OperatorState, 0, len(update.Operators))
	for _, opState := range update.Operators {
		state, err := sp.node.convertToOperatorState(opState)
		if err != nil {
			log.Printf("[StateSync] Failed to convert operator state: %v", err)
			continue
		}
		states = append(states, state)
	}

	// Update states
	sp.node.updateOperatorStates(ctx, states)

	// Check if epoch update is included
	if update.Epoch != nil {
		epochNumber := uint32(*update.Epoch)
		log.Printf("[StateSync] Processing epoch update: %d", epochNumber)
		
		// Create context with timeout for epoch update
		updateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := sp.node.updateEpochState(updateCtx, epochNumber); err != nil {
			return fmt.Errorf("failed to update epoch state: %w", err)
		}
		log.Printf("[StateSync] Successfully updated epoch state to %d", epochNumber)
	}

	return nil
}

// startStateHashVerification periodically verifies state hash with registry
func (sp *StateSyncProcessor) startStateHashVerification(ctx context.Context) {
	log.Printf("[StateSync] Starting state hash verification")
	ticker := time.NewTicker(stateHashCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			log.Printf("[StateSync] Stopping state hash verification")
			return
		default:
			if err := sp.verifyStateWithRegistry(ctx); err != nil {
				log.Printf("[StateSync] Failed to verify state: %v", err)
			}
		}
	}
}

// verifyStateWithRegistry verifies local state hash with registry
func (sp *StateSyncProcessor) verifyStateWithRegistry(ctx context.Context) error {
	log.Printf("[StateSync] Verifying state with registry...")
	
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

	log.Printf("[StateSync] Sent state verify request to registry, waiting for response...")

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
			state, err := sp.node.convertToOperatorState(opState)
			if err != nil {
				log.Printf("[StateSync] Failed to convert operator state: %v", err)
				continue
			}
			states = append(states, state)
		}

		sp.node.updateOperatorStates(ctx, states)
	} else {
		log.Printf("[StateSync] State verified successfully with registry")
	}

	return nil
}

func (sp *StateSyncProcessor) handleStateVerifyStream(stream network.Stream) {
	defer stream.Close()
	
	log.Printf("[StateVerify] Received state verify response from registry")
	
	// Read response
	var resp pb.StateVerifyResponse
	if err := utils.ReadStreamMessage(stream, &resp, maxMessageSize); err != nil {
		log.Printf("[StateVerify] Failed to read verify response: %v", err)
		stream.Reset()
		return
	}

	// If state mismatch, update local state
	if !resp.Success {
		log.Printf("[StateVerify] State mismatch detected, updating local state...")
		states := make([]*OperatorState, 0, len(resp.Operators))
		
		for _, opState := range resp.Operators {
			state, err := sp.node.convertToOperatorState(opState)
			if err != nil {
				log.Printf("[StateVerify] Failed to convert operator state: %v", err)
				continue
			}
			states = append(states, state)
		}

		sp.node.updateOperatorStates(context.Background(), states)
	} else {
		log.Printf("[StateVerify] State verified successfully with registry")
	}
}










