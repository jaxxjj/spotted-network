package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"runtime/debug"

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
	sub   			 *pubsub.Subscription
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

// NewStateSyncProcessor creates a new state sync processor
func NewStateSyncProcessor(ctx context.Context, node *Node, pubsub PubSubService) (*StateSyncProcessor, error) {
	// Create a new context with cancel
	ctx, cancel := context.WithCancel(ctx)
	
	sp := &StateSyncProcessor{
		node:             node,
		pubsub:           pubsub,
		cancel:           cancel,
	}

	// Subscribe to state sync topic
	sub, err := sp.SubscribeToStateSyncTopic()
	if err != nil {
		cancel() // Clean up if subscription fails
		return nil, fmt.Errorf("failed to subscribe to state sync topic: %w", err)
	}

	// Set up handler for state verify responses from registry
	node.host.SetStreamHandler(StateVerifyProtocol, sp.handleStateVerifyStream)
	
	// Start handling updates
	sp.wg.Add(2) 
	go func() {
		defer sp.wg.Done()
		sp.handleStateSync(ctx, sub)
	}()

	// Start state hash verification
	go func() {
		defer sp.wg.Done()
		sp.startStateHashVerification(ctx)
	}()

	return sp, nil
}

// Stop gracefully stops the state sync processor
func (sp *StateSyncProcessor) Stop() {
    log.Printf("[StateSync] Stopping state sync processor...")
    
    sp.cancel()
    
    // Wait for all goroutines to finish
    sp.wg.Wait()
    
    // Remove stream handler
    sp.node.host.RemoveStreamHandler(StateVerifyProtocol)
    
    // close topic related resources
    if sp.stateSyncTopic != nil {
        // 1. cancel subscription
        if sp.sub != nil {
            sp.sub.Cancel()
            sp.sub = nil
        }
        
        // 2. clean up topic
        sp.stateSyncTopic = nil
    }
    
    log.Printf("[StateSync] State sync processor stopped")
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

// handleStateSync handles incoming state updates from pubsub (worker pool pattern)
func (sp *StateSyncProcessor) handleStateSync(ctx context.Context, sub *pubsub.Subscription) {
	const (
		workerCount = 2    
		queueSize   = 60  
	)

	taskQueue := make(chan *pubsub.Message, queueSize)
	var wg sync.WaitGroup
	
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			log.Printf("[StateSync] Worker %d started", workerID)
			defer log.Printf("[StateSync] Worker %d stopped", workerID)
			
			for msg := range taskQueue {
				processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				
				func() {
					defer cancel()
					defer func() {
						if r := recover(); r != nil {
							log.Printf("[StateSync] Worker %d recovered from panic: %v\n%s", 
								workerID, r, debug.Stack())
						}
					}()

					if err := sp.processStateSync(processCtx, msg); err != nil {
						log.Printf("[StateSync] Worker %d failed to process message: %v", 
							workerID, err)
					}
				}()
			}
		}(i)
	}

	log.Printf("[StateSync] Started %d workers", workerCount)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[StateSync] Context cancelled, stopping...")
			close(taskQueue)
			wg.Wait()
			log.Printf("[StateSync] All workers stopped")
			return
			
		default:
			msg, err := sub.Next(ctx)
			if err != nil {
				if ctx.Err() != nil {
					continue
				}
				log.Printf("[StateSync] Error getting next message: %v", err)
				continue
			}

			select {
			case taskQueue <- msg:
				// message added to queue
			case <-ctx.Done():
				close(taskQueue)
				wg.Wait()
				return
			}
		}
	}
}

// processStateSync handles a single state update message
func (sp *StateSyncProcessor) processStateSync(ctx context.Context, msg *pubsub.Message) error {
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

	// Update states
	sp.node.updateOperatorStates(ctx, states)

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










