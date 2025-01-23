package operator

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	pb "github.com/galxe/spotted-network/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/protobuf/proto"
)

// handleResponses handles incoming task responses
func (tp *TaskProcessor) handleResponses(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(context.Background())
		if err != nil {
			tp.logger.Printf("[Response] Failed to get next response message: %v", err)
			continue
		}
		tp.logger.Printf("[Response] Received new message from peer: %s", msg.ReceivedFrom.String())

		// Parse protobuf message
		var pbMsg pb.TaskResponseMessage
		if err := proto.Unmarshal(msg.Data, &pbMsg); err != nil {
			tp.logger.Printf("[Response] Failed to unmarshal message: %v", err)
			continue
		}
		tp.logger.Printf("[Response] Received task response for task %s from operator %s", pbMsg.TaskId, pbMsg.OperatorAddr)

		// Check if consensus already exists for this task
		_, err = tp.consensusDB.GetConsensusResponse(context.Background(), pbMsg.TaskId)
		if err == nil {
			// Consensus already exists, skip this response
			tp.logger.Printf("[Response] Consensus already exists for task %s, skipping response from %s", 
				pbMsg.TaskId, pbMsg.OperatorAddr)
			continue
		}

		// Skip messages from self (检查operator address)
		if strings.EqualFold(pbMsg.OperatorAddr, tp.signer.Address()) {
			tp.logger.Printf("[Response] Skipping message from self operator: %s", pbMsg.OperatorAddr)
			continue
		}

		// Skip messages with our signing key
		if strings.EqualFold(pbMsg.SigningKey, tp.signer.GetSigningKey()) {
			tp.logger.Printf("[Response] Skipping message with self signing key: %s", pbMsg.SigningKey)
			continue
		}

		// Convert to TaskResponse
		response, err := convertToTaskResponse(&pbMsg)
		if err != nil {
			tp.logger.Printf("[Response] Failed to convert message: %v", err)
			continue
		}
		tp.logger.Printf("[Response] Successfully converted task response for task %s", response.TaskID)

		// Verify operator is active by checking signing key
		if !tp.isActiveOperator(response.SigningKey) {
			tp.logger.Printf("[Response] Skipping response from inactive operator signing key: %s", response.SigningKey)
			continue
		}

		// Get operator weight
		weight, err := tp.getOperatorWeight(response.OperatorAddr)
		if err != nil {
			tp.logger.Printf("[Response] Invalid operator %s: %v", response.OperatorAddr, err)
			continue
		}
		tp.logger.Printf("[Response] Got operator weight for %s: %s", response.OperatorAddr, weight.String())

		// Verify response
		if err := tp.verifyResponse(response); err != nil {
			tp.logger.Printf("[Response] Invalid response from %s: %v", response.OperatorAddr, err)
			continue
		}
		tp.logger.Printf("[Response] Verified response signature from operator %s", response.OperatorAddr)

		// Store response in local map
		tp.responsesMutex.Lock()
		if _, exists := tp.responses[response.TaskID]; !exists {
			tp.responses[response.TaskID] = make(map[string]*types.TaskResponse)
		}
		tp.responses[response.TaskID][response.OperatorAddr] = response
		tp.responsesMutex.Unlock()
		tp.logger.Printf("[Response] Stored response in memory for task %s from operator %s", response.TaskID, response.OperatorAddr)

		// Store weight in local map
		tp.weightsMutex.Lock()
		if _, exists := tp.taskWeights[response.TaskID]; !exists {
			tp.taskWeights[response.TaskID] = make(map[string]*big.Int)
		}
		tp.taskWeights[response.TaskID][response.OperatorAddr] = weight
		tp.weightsMutex.Unlock()
		tp.logger.Printf("[Response] Stored weight in memory for task %s from operator %s", response.TaskID, response.OperatorAddr)

		// Store in database
		if err := tp.storeResponse(context.Background(), response); err != nil {
			if !strings.Contains(err.Error(), "duplicate key value") {
				tp.logger.Printf("[Response] Failed to store response: %v", err)
			} else {
				tp.logger.Printf("[Response] Response already exists in database for task %s from operator %s", response.TaskID, response.OperatorAddr)
			}
			continue
		}
		tp.logger.Printf("[Response] Successfully stored response in database for task %s from operator %s", response.TaskID, response.OperatorAddr)

		// Check if we need to process this task
		tp.responsesMutex.RLock()
		_, processed := tp.responses[response.TaskID][tp.signer.Address()]
		tp.responsesMutex.RUnlock()

		if !processed {
			// Process task and broadcast our response
			if err := tp.ProcessTask(context.Background(), &types.Task{
				ID:            response.TaskID,
				Value:         response.Value,
				BlockNumber:   response.BlockNumber,
				ChainID:       response.ChainID,
				TargetAddress: response.TargetAddress,
				Key:          response.Key,
				Epoch:        response.Epoch,
				Timestamp:    response.Timestamp,
			}); err != nil {
				tp.logger.Printf("[Response] Failed to process task: %v", err)
			}
		}

		// Check consensus
		if err := tp.checkConsensus(response.TaskID); err != nil {
			tp.logger.Printf("[Response] Failed to check consensus: %v", err)
			continue
		}
		tp.logger.Printf("[Response] Completed consensus check for task %s", response.TaskID)
	}
}


// broadcastResponse broadcasts a task response to other operators
func (tp *TaskProcessor) broadcastResponse(resp *types.TaskResponse) error {
	tp.logger.Printf("[Broadcast] Starting to broadcast response for task %s", resp.TaskID)
	
	// Create protobuf message
	msg := &pb.TaskResponseMessage{
		TaskId:        resp.TaskID,
		OperatorAddr:  resp.OperatorAddr,
		SigningKey:    tp.signer.GetSigningKey(),
		Signature:     resp.Signature,
		Value:         resp.Value.String(),
		BlockNumber:   resp.BlockNumber.String(),
		ChainId:       int32(resp.ChainID),
		TargetAddress: resp.TargetAddress,
		Key:          resp.Key.String(),
		Epoch:        resp.Epoch,
	}

	// Marshal message
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	tp.logger.Printf("[Broadcast] Marshaled response message for task %s", resp.TaskID)

	// Publish to topic
	if err := tp.responseTopic.Publish(context.Background(), data); err != nil {
		tp.logger.Printf("[Broadcast] Failed to publish response: %v", err)
		return err
	}
	tp.logger.Printf("[Broadcast] Successfully published response for task %s to topic %s", resp.TaskID, tp.responseTopic.String())

	return nil
}

// convertToTaskResponse converts a protobuf task response to a TaskResponse
func convertToTaskResponse(msg *pb.TaskResponseMessage) (*types.TaskResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	value := new(big.Int)
	if _, ok := value.SetString(msg.Value, 10); !ok {
		return nil, fmt.Errorf("invalid value: %s", msg.Value)
	}

	blockNumber := new(big.Int)
	if _, ok := blockNumber.SetString(msg.BlockNumber, 10); !ok {
		return nil, fmt.Errorf("invalid block number: %s", msg.BlockNumber)
	}

	key := new(big.Int)
	if _, ok := key.SetString(msg.Key, 10); !ok {
		return nil, fmt.Errorf("invalid key: %s", msg.Key)
	}

	// 设置timestamp
	timestamp := new(big.Int).SetInt64(time.Now().Unix())

	return &types.TaskResponse{
		TaskID:        msg.TaskId,
		OperatorAddr:  msg.OperatorAddr,
		SigningKey:    msg.SigningKey,
		Signature:     msg.Signature,
		Value:         value,
		BlockNumber:   blockNumber,
		ChainID:       uint64(msg.ChainId),
		TargetAddress: msg.TargetAddress,
		Key:          key,
		Epoch:        msg.Epoch,
		Timestamp:    timestamp,
		CreatedAt:    time.Now(),
	}, nil
}

// storeResponse stores a task response in the database
func (tp *TaskProcessor) storeResponse(ctx context.Context, resp *types.TaskResponse) error {
	params := task_responses.CreateTaskResponseParams{
		TaskID:         resp.TaskID,
		OperatorAddress: resp.OperatorAddr,
		SigningKey:     resp.SigningKey,
		Signature:      resp.Signature,
		Epoch:         int32(resp.Epoch),
		ChainID:       int32(resp.ChainID),
		TargetAddress: resp.TargetAddress,
		Key:           types.NumericFromBigInt(resp.Key),
		Value:         types.NumericFromBigInt(resp.Value),
		BlockNumber:   types.NumericFromBigInt(resp.BlockNumber),
		Timestamp:     types.NumericFromBigInt(resp.Timestamp),
	}

	_, err := tp.db.CreateTaskResponse(ctx, params)
	return err
}
// verifyResponse verifies a task response signature
func (tp *TaskProcessor) verifyResponse(response *types.TaskResponse) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	if response.BlockNumber == nil {
		return fmt.Errorf("block number is nil")
	}
	if response.Key == nil {
		return fmt.Errorf("key is nil")
	}
	if response.Value == nil {
		return fmt.Errorf("value is nil")
	}
	if response.Timestamp == nil {
		return fmt.Errorf("timestamp is nil")
	}

	params := signer.TaskSignParams{
		User:        ethcommon.HexToAddress(response.TargetAddress),
		ChainID:     uint32(response.ChainID),
		BlockNumber: response.BlockNumber.Uint64(),
		Key:         response.Key,
		Value:       response.Value,
	}

	return tp.signer.VerifyTaskResponse(params, response.Signature, response.OperatorAddr)
}