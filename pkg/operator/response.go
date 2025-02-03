package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
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

// handleResponses handles incoming task responses
func (tp *TaskProcessor) handleResponses(ctx context.Context, sub *pubsub.Subscription) {
	// use buffered channel to limit concurrent goroutine number
	semaphore := make(chan struct{}, 10) 
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("[Response] Stopping response handler: %v", ctx.Err())
			return
		default:
			msg, err := sub.Next(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("[Response] Context cancelled while getting next message")
					return
				}
				log.Printf("[Response] Failed to get next response message: %v", err)
				continue
			}

			semaphore <- struct{}{}
			
			go func(msg *pubsub.Message) {
				defer func() {
					<-semaphore
				}()

				// Parse protobuf message
				var pbMsg pb.TaskResponseMessage
				if err := proto.Unmarshal(msg.Data, &pbMsg); err != nil {
					log.Printf("[Response] Failed to unmarshal message: %v", err)
					return
				}

				// Get peer ID and verify it's an active operator
				peerID := msg.ReceivedFrom
				operatorState := tp.node.GetOperatorState(peerID)
				if operatorState == nil {
					log.Printf("[Response] Peer %s not found in active operators", peerID)
					return
				}

				// Verify operator address
				if !strings.EqualFold(operatorState.Address, pbMsg.OperatorAddress) {
					log.Printf("[Response] Operator address mismatch for peer %s", peerID)
					return
				}

				if tp.alreadyProcessed(pbMsg.TaskId, pbMsg.OperatorAddress) {
					return
				}

				// Check existing consensus
				consensus, err := tp.consensusResponse.GetConsensusResponseByTaskId(context.Background(), pbMsg.TaskId)
				if err == nil && consensus != nil {
					return
				}

				// Convert and verify response
				response, err := convertToTaskResponse(&pbMsg)
				if err != nil || response == nil {
					log.Printf("[Response] Invalid response: %v", err)
					return
				}

				// Get and verify operator weight
				weight, err := tp.node.getOperatorWeight(peerID)
				if err != nil {
					log.Printf("[Response] Invalid operator weight: %v", err)
					return
				}

				if err := tp.verifyResponse(response); err != nil {
					log.Printf("[Response] Invalid response signature: %v", err)
					return
				}

				// Store response and weight
				tp.storeResponseAndWeight(response, weight)

				// Process task if needed
				tp.responsesMutex.RLock()
				processed := tp.responses[response.TaskID][tp.signer.GetOperatorAddress().Hex()] != nil
				tp.responsesMutex.RUnlock()

				if !processed {
					if err := tp.ProcessTask(context.Background(), &tasks.Tasks{
						TaskID:        response.TaskID,
						Value:         response.Value,
						BlockNumber:   response.BlockNumber,
						ChainID:       response.ChainID,
						TargetAddress: response.TargetAddress,
						Key:          response.Key,
						Epoch:        response.Epoch,
						Timestamp:    response.Timestamp,
					}); err != nil {
						log.Printf("[Response] Failed to process task: %v", err)
					}
				}

				// Check consensus
				if err := tp.checkConsensus(response.TaskID); err != nil {
					log.Printf("[Response] Failed to check consensus: %v", err)
				}
				
			}(msg)
		}
	}
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
	// 1. check if processed in memory
	tp.responsesMutex.RLock()
	_, processed := tp.responses[taskId][operatorAddress]
	tp.responsesMutex.RUnlock()
	
	if processed {
		log.Printf("[Response] Task %s from operator %s already processed in memory", taskId, operatorAddress)
		return true
	}
	
	// 2. check if consensus exists in database
	consensus, err := tp.consensusResponse.GetConsensusResponseByTaskId(context.Background(), taskId)
	if err != nil {
		log.Printf("[Response] Error checking consensus for task %s: %v", taskId, err)
		return false
	}
	
	// if database has consensus record, task has reached consensus
	if consensus != nil {
		log.Printf("[Response] Task %s already has consensus in database", taskId)
		return true
	}
	
	return false
}
