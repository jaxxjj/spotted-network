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
func (tp *TaskProcessor) handleResponses(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(context.Background())
		if err != nil {
			log.Printf("[Response] Failed to get next response message: %v", err)
			continue
		}
		log.Printf("[Response] Received new message from peer: %s", msg.ReceivedFrom.String())

		// Parse protobuf message
		var pbMsg pb.TaskResponseMessage
		if err := proto.Unmarshal(msg.Data, &pbMsg); err != nil {
			log.Printf("[Response] Failed to unmarshal message: %v", err)
			continue
		}
		log.Printf("[Response] Received task response for task %s from operator %s", pbMsg.TaskId, pbMsg.OperatorAddress)

		// Check if consensus already exists for this task
		consensus, err := tp.consensusResponse.GetConsensusResponseByTaskId(context.Background(), pbMsg.TaskId)
		if err == nil {
			if consensus != nil {
				// Consensus already exists, skip this response
				log.Printf("[Response] Consensus already exists for task %s, skipping response from %s", 
					pbMsg.TaskId, pbMsg.OperatorAddress)
				continue
			}
		} else {
			log.Printf("[Response] Error checking consensus for task %s: %v", pbMsg.TaskId, err)
			continue
		}

		// Skip messages from self (operator address)
		if strings.EqualFold(pbMsg.OperatorAddress, tp.signer.GetOperatorAddress().Hex()) {
			log.Printf("[Response] Skipping message from self operator: %s", pbMsg.OperatorAddress)
			continue
		}

		// Convert to TaskResponse
		response, err := convertToTaskResponse(&pbMsg)
		if err != nil {
			log.Printf("[Response] Failed to convert message: %v", err)
			continue
		}
		if response == nil {
			log.Printf("[Response] Converted response is nil for task %s", pbMsg.TaskId)
			continue
		}
		log.Printf("[Response] Successfully converted task response for task %s", response.TaskID)

		// Verify operator is active by checking signing key
		if !tp.isActiveOperator(response.SigningKey) {
			log.Printf("[Response] Skipping response from inactive operator signing key: %s", response.SigningKey)
			continue
		}

		// Get operator weight
		weight, err := tp.getOperatorWeight(response.OperatorAddress)
		if err != nil {
			log.Printf("[Response] Invalid operator %s: %v", response.OperatorAddress, err)
			continue
		}
		log.Printf("[Response] Got operator weight for %s: %s", response.OperatorAddress, weight.String())

		// Verify response
		if err := tp.verifyResponse(response); err != nil {
			log.Printf("[Response] Invalid response from %s: %v", response.OperatorAddress, err)
			continue
		}
		log.Printf("[Response] Verified response signature from operator %s", response.OperatorAddress)

		// Store response in local map
		tp.responsesMutex.Lock()
		if _, exists := tp.responses[response.TaskID]; !exists {
			tp.responses[response.TaskID] = make(map[string]*task_responses.TaskResponses)
		}
		tp.responses[response.TaskID][response.OperatorAddress] = response
		tp.responsesMutex.Unlock()
		log.Printf("[Response] Stored response in memory for task %s from operator %s", response.TaskID, response.OperatorAddress)

		// Store weight in local map
		tp.weightsMutex.Lock()
		if _, exists := tp.taskWeights[response.TaskID]; !exists {
			tp.taskWeights[response.TaskID] = make(map[string]*big.Int)
		}
		tp.taskWeights[response.TaskID][response.OperatorAddress] = weight
		tp.weightsMutex.Unlock()
		log.Printf("[Response] Stored weight in memory for task %s from operator %s", response.TaskID, response.OperatorAddress)

		// Store in database TODO: need to store?
		if err := tp.storeResponse(context.Background(), response); err != nil {
			if !strings.Contains(err.Error(), "duplicate key value") {
				log.Printf("[Response] Failed to store response: %v", err)
			} else {
				log.Printf("[Response] Response already exists in database for task %s from operator %s", response.TaskID, response.OperatorAddress)
			}
			continue
		}
		log.Printf("[Response] Successfully stored response in database for task %s from operator %s", response.TaskID, response.OperatorAddress)

		// Check if we need to process this task
		tp.responsesMutex.RLock()
		_, processed := tp.responses[response.TaskID][tp.signer.GetOperatorAddress().Hex()]
		tp.responsesMutex.RUnlock()

		if !processed {
			// Process task and broadcast our response
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
			continue
		}
		log.Printf("[Response] Completed consensus check for task %s", response.TaskID)
	}
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

