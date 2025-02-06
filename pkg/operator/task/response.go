package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"runtime/debug"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	pb "github.com/galxe/spotted-network/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/protobuf/proto"
)


type TaskResponseQuerier interface {
	CreateTaskResponse(ctx context.Context, arg task_responses.CreateTaskResponseParams) (*task_responses.TaskResponses, error)
	GetTaskResponse(ctx context.Context, arg task_responses.GetTaskResponseParams) (*task_responses.TaskResponses, error)
}

// handleResponses handles incoming task responses (backpressure pattern)
func (tp *TaskProcessor) handleResponses(ctx context.Context, sub *pubsub.Subscription) {
	const (
		maxConcurrent = 10
		processTimeout = 30 * time.Second
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
func (tp *TaskProcessor) processResponse(ctx context.Context, msg *pubsub.Message) error {
	// Parse protobuf message
	var pbMsg pb.TaskResponseMessage
	if err := proto.Unmarshal(msg.Data, &pbMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Get and verify operator
	peerID := msg.ReceivedFrom
	operatorState := tp.node.GetOperatorState(peerID)
	if operatorState == nil {
		return fmt.Errorf("peer %s not found in active operators", peerID)
	}

	// Verify operator address
	if !strings.EqualFold(operatorState.Address, pbMsg.OperatorAddress) {
		return fmt.Errorf("operator address mismatch for peer %s", peerID)
	}

	// Check if already processed
	if tp.alreadyProcessed(pbMsg.TaskId, pbMsg.OperatorAddress) {
		return nil // Not an error, just skip
	}

	// Check existing consensus
	consensus, err := tp.consensusResponse.GetConsensusResponseByTaskId(ctx, pbMsg.TaskId)
	if err == nil && consensus != nil {
		return nil // Not an error, just skip
	}

	// Convert and verify response
	response, err := convertToTaskResponse(&pbMsg)
	if err != nil || response == nil {
		return fmt.Errorf("invalid response: %w", err)
	}

	// Get and verify operator weight
	weight, err := tp.node.getOperatorWeight(peerID)
	if err != nil {
		return fmt.Errorf("invalid operator weight: %w", err)
	}

	// Verify response signature
	if err := tp.verifyResponse(response); err != nil {
		return fmt.Errorf("invalid response signature: %w", err)
	}

	// Store response and weight
	tp.storeResponseAndWeight(response, weight)

	// Check if we need to process this task
	if err := tp.maybeProcessTask(ctx, response); err != nil {
		return fmt.Errorf("failed to process task: %w", err)
	}

	// Check consensus
	if err := tp.checkConsensus(response.TaskID); err != nil {
		return fmt.Errorf("failed to check consensus: %w", err)
	}

	return nil
}

// maybeProcessTask processes the task if not already processed
func (tp *TaskProcessor) maybeProcessTask(ctx context.Context, response *task_responses.TaskResponses) error {
	tp.responsesMutex.RLock()
	processed := tp.responses[response.TaskID][tp.signer.GetOperatorAddress().Hex()] != nil
	tp.responsesMutex.RUnlock()

	if !processed {
		return tp.ProcessTask(ctx, &tasks.Tasks{
			TaskID:        response.TaskID,
			Value:         response.Value,
			BlockNumber:   response.BlockNumber,
			ChainID:       response.ChainID,
			TargetAddress: response.TargetAddress,
			Key:          response.Key,
			Epoch:        response.Epoch,
			Timestamp:    response.Timestamp,
		})
	}
	return nil
}

// helper function to store response and weight
func (tp *TaskProcessor) storeResponseAndWeight(response *task_responses.TaskResponses, weight *big.Int) {
	// Store response
	tp.responsesMutex.Lock()
	if _, exists := tp.responses[response.TaskID]; !exists {
		tp.responses[response.TaskID] = make(map[string]*task_responses.TaskResponses)
	}
	tp.responses[response.TaskID][response.OperatorAddress] = response
	tp.responsesMutex.Unlock()

	// Store weight
	tp.weightsMutex.Lock()
	if _, exists := tp.taskWeights[response.TaskID]; !exists {
		tp.taskWeights[response.TaskID] = make(map[string]*big.Int)
	}
	tp.taskWeights[response.TaskID][response.OperatorAddress] = weight
	tp.weightsMutex.Unlock()
}

// broadcastResponse broadcasts a task response to other operators
func (tp *TaskProcessor) broadcastResponse(response *task_responses.TaskResponses) error {
	log.Printf("[Response] Starting to broadcast response for task %s", response.TaskID)
	
	// Create protobuf message
	msg := &pb.TaskResponseMessage{
		TaskId:        response.TaskID,
		OperatorAddress:  response.OperatorAddress,
		SigningKey:    tp.signer.GetSigningAddress().Hex(),
		Signature:     response.Signature,
		Value:         utils.NumericToString(response.Value),
		BlockNumber:   response.BlockNumber,
		ChainId:       response.ChainID,
		TargetAddress: response.TargetAddress,
		Key:          utils.NumericToString(response.Key),
		Epoch:        response.Epoch,
		Timestamp:    response.Timestamp,
	}

	// Marshal message
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	log.Printf("[Broadcast] Marshaled response message for task %s", response.TaskID)

	// Publish to topic
	if err := tp.responseTopic.Publish(context.Background(), data); err != nil {
		log.Printf("[Response] Failed to publish response: %v", err)
		return err
	}
	log.Printf("[Response] Successfully published response for task %s to topic %s", response.TaskID, tp.responseTopic.String())

	return nil
}

// convertToTaskResponse converts a protobuf task response to a TaskResponse
func convertToTaskResponse(msg *pb.TaskResponseMessage) (*task_responses.TaskResponses, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	return &task_responses.TaskResponses{
		TaskID:        msg.TaskId,
		OperatorAddress:  msg.OperatorAddress,
		SigningKey:    msg.SigningKey,
		Signature:     msg.Signature,
		Value:         utils.StringToNumeric(msg.Value),
		BlockNumber:   msg.BlockNumber,
		ChainID:       msg.ChainId,
		TargetAddress: msg.TargetAddress,
		Key:          utils.StringToNumeric(msg.Key),
		Epoch:        msg.Epoch,
		Timestamp:    msg.Timestamp,
		SubmittedAt:    time.Now(),
	}, nil
}

// storeResponse stores a task response in the database
func (tp *TaskProcessor) storeResponse(ctx context.Context, response *task_responses.TaskResponses) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	if tp.taskResponse == nil {
		return fmt.Errorf("task response querier not initialized")
	}

	params := task_responses.CreateTaskResponseParams{
		TaskID:         response.TaskID,
		OperatorAddress: response.OperatorAddress,
		SigningKey:     response.SigningKey,
		Signature:      response.Signature,
		Epoch:         response.Epoch,
		ChainID:       response.ChainID,
		TargetAddress: response.TargetAddress,
		Key:           response.Key,
		Value:         response.Value,
		BlockNumber:   response.BlockNumber,
		Timestamp:     response.Timestamp,
	}

	_, err := tp.taskResponse.CreateTaskResponse(ctx, params)
	return err
}
// verifyResponse verifies a task response signature
func (tp *TaskProcessor) verifyResponse(response *task_responses.TaskResponses) error {
	// basic check
	if response == nil {
		return fmt.Errorf("[Response] response is nil")
	}
	if response.BlockNumber == 0 {
		return fmt.Errorf("[Response] block number is nil")
	}
	if response.Timestamp == 0 {
		return fmt.Errorf("[Response] timestamp is nil")
	}

	key, err := utils.NumericToBigInt(response.Key)
	if err != nil {
		return fmt.Errorf("[Response] failed to convert key to big int: %w", err)
	}
	value, err := utils.NumericToBigInt(response.Value)
	if err != nil {
		return fmt.Errorf("[Response] failed to convert value to big int: %w", err)
	}

	params := signer.TaskSignParams{
		User:        ethcommon.HexToAddress(response.TargetAddress),
		ChainID:     response.ChainID,
		BlockNumber: response.BlockNumber,
		Key:         key,
		Value:       value,
	}

	// signature verification
	return tp.signer.VerifyTaskResponse(params, response.Signature, response.SigningKey)
}

func (tp *TaskProcessor) alreadyProcessed(taskId string, operatorAddress string) bool {
	// check if processed in memory
	tp.responsesMutex.RLock()
	_, processed := tp.responses[taskId][operatorAddress]
	tp.responsesMutex.RUnlock()
	
	if processed {
		log.Printf("[Response] Task %s from operator %s already processed in memory", taskId, operatorAddress)
		return true
	}
	
	return false
}
