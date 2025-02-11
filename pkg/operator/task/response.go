package task

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"runtime/debug"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	pb "github.com/galxe/spotted-network/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

type BlacklistRepo interface {
	IncrementViolationCount(ctx context.Context, arg blacklist.IncrementViolationCountParams, isBlocked *string) (*blacklist.Blacklist, error)
}

// handleResponses handles incoming task responses (backpressure pattern)
func (tp *taskProcessor) handleResponses(ctx context.Context, sub *pubsub.Subscription) {
	const (
		maxConcurrent  = 20
		processTimeout = 60 * time.Second
	)

	// Create semaphore for backpressure
	semaphore := make(chan struct{}, maxConcurrent)

	log.Printf("[Response] Starting response handler with max concurrent: %d", maxConcurrent)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Response] Stopping response handler: %v", ctx.Err())
			return
		default:
			// Get next message with context
			msg, err := sub.Next(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[Response] Failed to get next message: %v", err)
				continue
			}

			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				return
			case <-time.After(processTimeout):
				log.Printf("[Response] System overloaded, dropping message")
				continue
			}

			// Start processing goroutine
			go func(msg *pubsub.Message) {
				defer func() {
					<-semaphore // Release semaphore
				}()

				// Add recover for panics
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[Response] Recovered from panic: %v\n%s", r, debug.Stack())
					}
				}()

				// Create timeout context for processing
				processCtx, cancel := context.WithTimeout(ctx, processTimeout)
				defer cancel()

				if err := tp.processResponse(processCtx, msg); err != nil {
					log.Printf("[Response] Failed to process message: %v", err)
				}
			}(msg)
		}
	}
}

// processResponse handles a single response message
func (tp *taskProcessor) processResponse(ctx context.Context, msg *pubsub.Message) error {
	// Parse protobuf message
	var pbMsg pb.TaskResponseMessage
	if err := proto.Unmarshal(msg.Data, &pbMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	response := convertProtoToResponse(&pbMsg)
	// Get and verify operator
	peerID := msg.ReceivedFrom
	currentEpochNumber := tp.epochStateQuerier.GetCurrentEpochNumber()
	if response.epoch != currentEpochNumber {
		return nil // Not an error, just skip
	}

	p2pKey, err := utils.PeerIDToP2PKey(peerID)
	if err != nil {
		return tp.incrementPeerViolation(ctx, peerID, fmt.Errorf("failed to convert peer ID to P2P key: %w", err))
	}
	operator, err := tp.operatorRepo.GetOperatorByP2PKey(ctx, p2pKey)
	log.Printf("[Response] Operator: %+v", operator)
	if err != nil {
		return tp.incrementPeerViolation(ctx, peerID, fmt.Errorf("failed to get operator state: %w", err))
	}
	if operator == nil || !operator.IsActive {
		return tp.incrementPeerViolation(ctx, peerID, fmt.Errorf("operator not exist or not active: %w", err))
	}

	signingKey := operator.SigningKey
	// Check if already processed
	if tp.alreadyProcessed(pbMsg.TaskId, signingKey) {
		log.Printf("[Response] Task %s already processed by signing key %s", pbMsg.TaskId, signingKey)
		return nil
	}

	// Check existing consensus
	consensus, err := tp.consensusRepo.GetConsensusResponseByTaskId(ctx, pbMsg.TaskId)
	if err == nil && consensus != nil {
		log.Printf("[Response] Task %s already has consensus", pbMsg.TaskId)
		return nil
	}

	// Verify response signature
	if err := tp.verifyResponse(response, signingKey); err != nil {
		return tp.incrementPeerViolation(ctx, peerID, fmt.Errorf("failed to verify response: %w", err))
	}

	weight, err := utils.NumericToBigInt(operator.Weight)
	if err != nil {
		return fmt.Errorf("failed to convert operator weight to big int: %w", err)
	}
	// Store response and weight
	tp.storeResponseAndWeight(response, signingKey, weight)

	// Check consensus
	if err := tp.checkConsensus(ctx, response); err != nil {
		log.Printf("[Response] Failed to check consensus for task %s: %v", response.taskID, err)
	}

	// Check if we need to process this task
	tp.taskResponseTrack.mu.RLock()
	_, processed := tp.taskResponseTrack.responses[response.taskID][tp.signer.GetSigningAddress().Hex()]
	tp.taskResponseTrack.mu.RUnlock()

	if !processed {
		if err := tp.ProcessTask(ctx, &tasks.Tasks{
			TaskID:        response.taskID,
			Value:         utils.StringToNumeric(response.value),
			BlockNumber:   response.blockNumber,
			ChainID:       response.chainID,
			TargetAddress: response.targetAddress,
			Key:           utils.StringToNumeric(response.key),
			Epoch:         response.epoch,
		}); err != nil {
			return fmt.Errorf("failed to process task: %w", err)
		}
	}

	return nil
}

// helper function to store response and weight
func (tp *taskProcessor) storeResponseAndWeight(response taskResponse, signingKey string, weight *big.Int) {
	// Store in local map
	tp.taskResponseTrack.mu.Lock()
	// Initialize maps if they don't exist
	if _, exists := tp.taskResponseTrack.responses[response.taskID]; !exists {
		tp.taskResponseTrack.responses[response.taskID] = make(map[string]taskResponse)
		tp.taskResponseTrack.weights[response.taskID] = make(map[string]*big.Int)
	}
	// Add new response and weight
	tp.taskResponseTrack.responses[response.taskID][signingKey] = response
	tp.taskResponseTrack.weights[response.taskID][signingKey] = weight
	tp.taskResponseTrack.mu.Unlock()
}

// verifyResponse verifies a task response signature
func (tp *taskProcessor) verifyResponse(response taskResponse, signingKey string) error {
	key := utils.StringToBigInt(response.key)
	value := utils.StringToBigInt(response.value)

	params := signer.TaskSignParams{
		User:        ethcommon.HexToAddress(response.targetAddress),
		ChainID:     response.chainID,
		BlockNumber: response.blockNumber,
		Key:         key,
		Value:       value,
	}

	// signature verification
	return tp.signer.VerifyTaskResponse(params, response.signature, signingKey)
}

func (tp *taskProcessor) alreadyProcessed(taskId string, signingKey string) bool {
	// check if processed in memory
	tp.taskResponseTrack.mu.RLock()
	_, processed := tp.taskResponseTrack.responses[taskId][signingKey]
	tp.taskResponseTrack.mu.RUnlock()

	return processed
}

// incrementPeerViolation adds a violation record for the peer and returns the original error
func (tp *taskProcessor) incrementPeerViolation(ctx context.Context, peerID peer.ID, originalErr error) error {
	peerIDStr := peerID.String()

	_, err := tp.blacklistRepo.IncrementViolationCount(ctx, blacklist.IncrementViolationCountParams{
		PeerID:         peerIDStr,
		ViolationCount: 1,
	}, &peerIDStr)
	if err != nil {
		log.Printf("[Response] failed to increment violation count: %v", err)
	}

	return fmt.Errorf("%v: %w", originalErr, err)
}

// newTaskResponseFromProto creates a taskResponse from a protobuf message
func convertProtoToResponse(pbMsg *pb.TaskResponseMessage) taskResponse {
	return taskResponse{
		taskID:        pbMsg.TaskId,
		signature:     pbMsg.Signature,
		epoch:         pbMsg.Epoch,
		chainID:       pbMsg.ChainId,
		targetAddress: pbMsg.TargetAddress,
		key:           pbMsg.Key,
		value:         pbMsg.Value,
		blockNumber:   pbMsg.BlockNumber,
	}
}
