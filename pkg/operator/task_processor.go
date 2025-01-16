package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5/pgtype"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/protobuf/proto"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	pb "github.com/galxe/spotted-network/proto"
)

const (
	TaskResponseProtocol = "/spotted/task-response/1.0.0"
)

// TaskProcessor handles task processing and consensus
type TaskProcessor struct {
	node           *Node
	signer         signer.Signer
	db             *task_responses.Queries
	consensusDB    *consensus_responses.Queries
	responseTopic  *pubsub.Topic
	responsesMutex sync.RWMutex
	responses      map[string]map[string]*types.TaskResponse // taskID -> operatorAddr -> response
}

// NewTaskProcessor creates a new task processor
func NewTaskProcessor(node *Node, db *task_responses.Queries, consensusDB *consensus_responses.Queries) (*TaskProcessor, error) {
	// Create response topic
	responseTopic, err := node.PubSub.Join(TaskResponseProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to join response topic: %w", err)
	}

	tp := &TaskProcessor{
		node:          node,
		signer:        node.signer,
		db:            db,
		consensusDB:   consensusDB,
		responseTopic: responseTopic,
		responses:     make(map[string]map[string]*types.TaskResponse),
	}

	// Subscribe to response topic
	sub, err := responseTopic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to response topic: %w", err)
	}

	go tp.handleResponses(sub)
	go tp.checkTimeouts(context.Background()) // Start timeout checker

	return tp, nil
}

// ProcessTask processes a new task and broadcasts response
func (tp *TaskProcessor) ProcessTask(ctx context.Context, task *types.Task) error {
	// Get state client for chain
	stateClient, err := tp.node.chainClient.GetStateClient(fmt.Sprintf("%d", task.ChainID))
	if err != nil {
		return fmt.Errorf("failed to get state client: %w", err)
	}

	// Get state from chain
	value, err := stateClient.GetStateAtBlock(
		ctx,
		common.HexToAddress(task.TargetAddress),
		task.Key,
		task.BlockNumber.Uint64(),
	)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	// Verify the state matches
	if value.Cmp(task.Value) != 0 {
		return fmt.Errorf("state value mismatch: expected %s, got %s", task.Value, value)
	}

	// Sign the response with all required fields
	signParams := signer.TaskSignParams{
		User:        common.HexToAddress(task.TargetAddress),
		ChainID:     uint32(task.ChainID),
		BlockNumber: task.BlockNumber.Uint64(),
		Timestamp:   task.Timestamp.Uint64(),
		Key:         task.Key,
		Value:       value,
	}
	
	signature, err := tp.signer.SignTaskResponse(signParams)
	if err != nil {
		return fmt.Errorf("failed to sign response: %w", err)
	}

	// Create response
	response := &types.TaskResponse{
		TaskID:        task.ID,
		OperatorAddr:  tp.signer.Address(),
		Signature:     signature,
		Value:         value,
		BlockNumber:   task.BlockNumber,
		ChainID:       task.ChainID,
		TargetAddress: task.TargetAddress,
		Key:          task.Key,
		Epoch:        task.Epoch,
		CreatedAt:    time.Now(),
	}

	// Store response in database
	if err := tp.storeResponse(ctx, response); err != nil {
		return fmt.Errorf("failed to store response: %w", err)
	}

	// Broadcast response
	if err := tp.broadcastResponse(response); err != nil {
		return fmt.Errorf("failed to broadcast response: %w", err)
	}

	return nil
}

// storeResponse stores a task response in the database
func (tp *TaskProcessor) storeResponse(ctx context.Context, resp *types.TaskResponse) error {
	params := task_responses.CreateTaskResponseParams{
		TaskID:         resp.TaskID,
		OperatorAddress: resp.OperatorAddr,
		Signature:      resp.Signature,
		Epoch:         int32(resp.Epoch),
		ChainID:       int32(resp.ChainID),
		TargetAddress: resp.TargetAddress,
		Key:           types.NumericFromBigInt(resp.Key),
		Value:         types.NumericFromBigInt(resp.Value),
		BlockNumber:   types.NumericFromBigInt(resp.BlockNumber),
		Status:        "verified",
	}

	_, err := tp.db.CreateTaskResponse(ctx, params)
	return err
}

// broadcastResponse broadcasts a task response to other operators
func (tp *TaskProcessor) broadcastResponse(resp *types.TaskResponse) error {
	// Create protobuf message
	msg := &pb.TaskResponseMessage{
		TaskId:        resp.TaskID,
		OperatorAddr:  resp.OperatorAddr,
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

	return tp.responseTopic.Publish(context.Background(), data)
}

// handleResponses handles incoming task responses
func (tp *TaskProcessor) handleResponses(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(context.Background())
		if err != nil {
			log.Printf("Failed to get next response message: %v", err)
			continue
		}

		// Skip messages from self
		if msg.ReceivedFrom == tp.node.host.ID() {
			continue
		}

		// Unmarshal protobuf message
		pbMsg := &pb.TaskResponseMessage{}
		if err := proto.Unmarshal(msg.Data, pbMsg); err != nil {
			log.Printf("Failed to unmarshal response: %v", err)
			continue
		}

		// Convert to TaskResponse
		value := new(big.Int)
		value.SetString(pbMsg.Value, 10)
		blockNumber := new(big.Int)
		blockNumber.SetString(pbMsg.BlockNumber, 10)
		key := new(big.Int)
		key.SetString(pbMsg.Key, 10)

		resp := &types.TaskResponse{
			TaskID:        pbMsg.TaskId,
			OperatorAddr:  pbMsg.OperatorAddr,
			Signature:     pbMsg.Signature,
			Value:         value,
			BlockNumber:   blockNumber,
			ChainID:       uint64(pbMsg.ChainId),
			TargetAddress: pbMsg.TargetAddress,
			Key:          key,
			Epoch:        pbMsg.Epoch,
		}

		// Verify signature
		if err := tp.verifyResponse(resp); err != nil {
			log.Printf("Failed to verify response: %v", err)
			continue
		}

		// Store response
		if err := tp.storeResponse(context.Background(), resp); err != nil {
			log.Printf("Failed to store response: %v", err)
			continue
		}

		// Update responses map
		tp.responsesMutex.Lock()
		if _, exists := tp.responses[resp.TaskID]; !exists {
			tp.responses[resp.TaskID] = make(map[string]*types.TaskResponse)
		}
		tp.responses[resp.TaskID][resp.OperatorAddr] = resp
		tp.responsesMutex.Unlock()

		// Check consensus
		if err := tp.checkConsensus(context.Background(), resp.TaskID); err != nil {
			log.Printf("Failed to check consensus: %v", err)
			continue
		}
	}
}

// verifyResponse verifies a task response signature
func (tp *TaskProcessor) verifyResponse(response *types.TaskResponse) error {
	return tp.signer.VerifyTaskResponse(
		response.TaskID,
		response.Value,
		response.BlockNumber,
		response.Signature,
		response.OperatorAddr,
	)
}

// aggregateSignatures combines multiple ECDSA signatures by concatenation
func (tp *TaskProcessor) aggregateSignatures(sigs map[string][]byte) []byte {
	// Convert map to sorted slice to ensure deterministic ordering
	sigSlice := make([][]byte, 0, len(sigs))
	for _, sig := range sigs {
		sigSlice = append(sigSlice, sig)
	}
	
	// Concatenate all signatures
	var aggregated []byte
	for _, sig := range sigSlice {
		aggregated = append(aggregated, sig...)
	}
	return aggregated
}

// checkConsensus checks if consensus has been reached for a task
func (tp *TaskProcessor) checkConsensus(ctx context.Context, taskID string) error {
	tp.responsesMutex.RLock()
	responses := tp.responses[taskID]
	tp.responsesMutex.RUnlock()

	if len(responses) == 0 {
		return nil
	}

	// Calculate total weight and collect signatures
	totalWeight := big.NewInt(0)
	operatorSigs := make(map[string][]byte)
	
	var firstResponse *types.TaskResponse
	for addr, resp := range responses {
		if firstResponse == nil {
			firstResponse = resp
		}
		// Get operator weight
		weight := tp.node.getOperatorWeight(addr)
		if weight.Cmp(big.NewInt(0)) <= 0 {
			log.Printf("Operator %s has no weight, skipping response", addr)
			continue
		}
		
		// Verify signature
		if err := tp.verifyResponse(resp); err != nil {
			log.Printf("Invalid signature from operator %s: %v", addr, err)
			continue
		}

		totalWeight.Add(totalWeight, weight)
		operatorSigs[addr] = resp.Signature
	}

	// Check if threshold is reached
	threshold := tp.node.getConsensusThreshold()
	if totalWeight.Cmp(threshold) >= 0 {
		// Create consensus response
		consensusResp := consensus_responses.CreateConsensusResponseParams{
			TaskID:              taskID,
			Epoch:              int32(firstResponse.Epoch),
			Status:             "completed",
			AggregatedSignatures: tp.aggregateSignatures(operatorSigs),
			OperatorSignatures:  tp.aggregateSignatures(operatorSigs),
			ConsensusReachedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		}
		
		// Store consensus result
		if _, err := tp.consensusDB.CreateConsensusResponse(ctx, consensusResp); err != nil {
			return fmt.Errorf("failed to store consensus: %w", err)
		}

		// Broadcast consensus reached message
		if err := tp.broadcastConsensusReached(taskID, &consensusResp); err != nil {
			log.Printf("Failed to broadcast consensus reached: %v", err)
		}

		// Clean up memory
		tp.responsesMutex.Lock()
		delete(tp.responses, taskID)
		tp.responsesMutex.Unlock()

		log.Printf("Consensus reached for task %s with weight %s/%s", taskID, totalWeight, threshold)
	}

	return nil
}

// broadcastConsensusReached broadcasts a consensus reached message
func (tp *TaskProcessor) broadcastConsensusReached(taskID string, consensus *consensus_responses.CreateConsensusResponseParams) error {
	msg := &pb.ConsensusReachedMessage{
		TaskId: taskID,
		Epoch: uint32(consensus.Epoch),
		AggregatedSignatures: consensus.AggregatedSignatures,
		ConsensusReachedAt: consensus.ConsensusReachedAt.Time.Unix(),
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal consensus message: %w", err)
	}

	return tp.responseTopic.Publish(context.Background(), data)
}

// checkTimeouts periodically checks for timed out tasks
func (tp *TaskProcessor) checkTimeouts(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			tp.responsesMutex.Lock()
			for taskID, responses := range tp.responses {
				// Check if any response is older than 5 minutes
				for _, resp := range responses {
					if time.Since(resp.CreatedAt) > 5*time.Minute {
						// Create failed consensus response
						consensusResp := consensus_responses.CreateConsensusResponseParams{
							TaskID: taskID,
							Epoch:  int32(resp.Epoch),
							Status: "failed",
						}
						
						if _, err := tp.consensusDB.CreateConsensusResponse(context.Background(), consensusResp); err != nil {
							log.Printf("Failed to store failed consensus: %v", err)
						}

						// Clean up memory
						delete(tp.responses, taskID)
						break
					}
				}
			}
			tp.responsesMutex.Unlock()
		}
	}
} 