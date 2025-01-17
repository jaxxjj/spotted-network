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
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	pb "github.com/galxe/spotted-network/proto"
)

const (
	TaskResponseProtocol = "/spotted/task-response/1.0.0"
)

// TaskProcessor handles task processing and consensus
type TaskProcessor struct {
	node           *Node
	signer         signer.Signer
	taskQueries    *tasks.Queries
	db             *task_responses.Queries
	consensusDB    *consensus_responses.Queries
	responseTopic  *pubsub.Topic
	responsesMutex sync.RWMutex
	responses      map[string]map[string]*types.TaskResponse // taskID -> operatorAddr -> response
	logger         *log.Logger
}

// NewTaskProcessor creates a new task processor
func NewTaskProcessor(node *Node, taskQueries *tasks.Queries, responseQueries *task_responses.Queries, consensusDB *consensus_responses.Queries) (*TaskProcessor, error) {
	// Create response topic
	responseTopic, err := node.PubSub.Join(TaskResponseProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to join response topic: %w", err)
	}

	tp := &TaskProcessor{
		node:          node,
		signer:        node.signer,
		taskQueries:   taskQueries,
		db:            responseQueries,
		consensusDB:   consensusDB,
		responseTopic: responseTopic,
		responses:     make(map[string]map[string]*types.TaskResponse),
		logger:        log.Default(),
	}

	// Subscribe to response topic
	sub, err := responseTopic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to response topic: %w", err)
	}

	go tp.handleResponses(sub)
	go tp.checkTimeouts(context.Background()) // Start timeout checker
	go tp.checkConfirmations(context.Background()) // Start confirmation checker

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
			tp.logger.Printf("[Response] Failed to get next response message: %v", err)
			continue
		}

		// Skip messages from self
		if msg.ReceivedFrom == tp.node.host.ID() {
			tp.logger.Printf("[Response] Skipping message from self")
			continue
		}

		// Unmarshal protobuf message
		pbMsg := &pb.TaskResponseMessage{}
		if err := proto.Unmarshal(msg.Data, pbMsg); err != nil {
			tp.logger.Printf("[Response] Failed to unmarshal response: %v", err)
			continue
		}

		tp.logger.Printf("[Response] Received response for task %s from operator %s", pbMsg.TaskId, pbMsg.OperatorAddr)

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
			tp.logger.Printf("[Response] Failed to verify response signature from %s: %v", resp.OperatorAddr, err)
			continue
		}
		tp.logger.Printf("[Response] Verified signature from operator %s", resp.OperatorAddr)

		// Store response
		if err := tp.storeResponse(context.Background(), resp); err != nil {
			tp.logger.Printf("[Response] Failed to store response in database: %v", err)
			continue
		}
		tp.logger.Printf("[Response] Stored response from operator %s in database", resp.OperatorAddr)

		// Get task details
		task, err := tp.taskQueries.GetTaskByID(context.Background(), resp.TaskID)
		if err != nil {
			tp.logger.Printf("[Response] Failed to get task details: %v", err)
			continue
		}

		// Update responses map
		tp.responsesMutex.Lock()
		if _, exists := tp.responses[resp.TaskID]; !exists {
			tp.responses[resp.TaskID] = make(map[string]*types.TaskResponse)
		}
		tp.responses[resp.TaskID][resp.OperatorAddr] = resp
		responseCount := len(tp.responses[resp.TaskID])
		tp.responsesMutex.Unlock()

		tp.logger.Printf("[Response] Added response to memory (total responses for task %s: %d)", resp.TaskID, responseCount)

		// For confirming tasks, we need to wait for block confirmations
		if task.Status == "confirming" {
			tp.logger.Printf("[Response] Task %s is in confirming status, waiting for block confirmations", resp.TaskID)
			continue
		}

		// Check consensus for non-confirming tasks
		if err := tp.checkConsensus(context.Background(), resp.TaskID); err != nil {
			tp.logger.Printf("[Response] Failed to check consensus: %v", err)
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
	tp.logger.Printf("[Consensus] Starting consensus check for task %s", taskID)
	
	tp.responsesMutex.RLock()
	responses := tp.responses[taskID]
	tp.responsesMutex.RUnlock()

	if len(responses) == 0 {
		tp.logger.Printf("[Consensus] No responses yet for task %s", taskID)
		return nil
	}

	// Calculate total weight and collect signatures
	totalWeight := big.NewInt(0)
	operatorSigs := make(map[string][]byte)
	
	var firstResponse *types.TaskResponse
	tp.logger.Printf("[Consensus] Processing %d responses for task %s", len(responses), taskID)
	
	for addr, resp := range responses {
		if firstResponse == nil {
			firstResponse = resp
		}
		// Get operator weight
		weight := tp.node.getOperatorWeight(addr)
		if weight.Cmp(big.NewInt(0)) <= 0 {
			tp.logger.Printf("[Consensus] Operator %s has no weight, skipping response", addr)
			continue
		}
		
		// Verify signature
		if err := tp.verifyResponse(resp); err != nil {
			tp.logger.Printf("[Consensus] Invalid signature from operator %s: %v", addr, err)
			continue
		}

		totalWeight.Add(totalWeight, weight)
		operatorSigs[addr] = resp.Signature
		tp.logger.Printf("[Consensus] Added operator %s weight: %s (total: %s)", addr, weight.String(), totalWeight.String())
	}

	// Check if threshold is reached
	threshold := tp.node.getConsensusThreshold()
	tp.logger.Printf("[Consensus] Current total weight: %s, Required threshold: %s", totalWeight.String(), threshold.String())
	
	if totalWeight.Cmp(threshold) >= 0 {
		tp.logger.Printf("[Consensus] Threshold reached for task %s! Creating consensus response", taskID)
		
		// Get task details to check if it needs confirmation
		task, err := tp.taskQueries.GetTaskByID(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to get task details: %w", err)
		}

		// If task is in confirming status, we need to wait for block confirmations
		if task.Status == "confirming" {
			tp.logger.Printf("[Consensus] Task %s is in confirming status, waiting for block confirmations", taskID)
			return nil
		}

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
		tp.logger.Printf("[Consensus] Stored consensus response for task %s", taskID)

		// Update task status to completed if not in confirming status
		if task.Status != "confirming" {
			_, err = tp.taskQueries.UpdateTaskStatus(ctx, tasks.UpdateTaskStatusParams{
				TaskID: taskID,
				Status: "completed",
			})
			if err != nil {
				tp.logger.Printf("[Consensus] Failed to update task status to completed: %v", err)
				return fmt.Errorf("failed to update task status: %w", err)
			}
			tp.logger.Printf("[Consensus] Updated task %s status to completed", taskID)
		}

		// Broadcast consensus reached message
		if err := tp.broadcastConsensusReached(taskID, &consensusResp); err != nil {
			tp.logger.Printf("[Consensus] Failed to broadcast consensus reached: %v", err)
		} else {
			tp.logger.Printf("[Consensus] Broadcasted consensus reached message for task %s", taskID)
		}

		// Clean up memory
		tp.responsesMutex.Lock()
		delete(tp.responses, taskID)
		tp.responsesMutex.Unlock()
		tp.logger.Printf("[Consensus] Cleaned up memory for task %s", taskID)

		tp.logger.Printf("[Consensus] Consensus process completed for task %s with weight %s/%s", taskID, totalWeight, threshold)
	} else {
		tp.logger.Printf("[Consensus] Threshold not yet reached for task %s (%s/%s)", taskID, totalWeight, threshold)
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
			for taskID := range tp.responses {
				// Get task details to check status
				task, err := tp.taskQueries.GetTaskByID(ctx, taskID)
				if err != nil {
					log.Printf("Failed to get task details: %v", err)
					continue
				}

				// Different timeout durations based on status
				var timeoutDuration time.Duration
				if task.Status == "confirming" {
					timeoutDuration = 15 * time.Minute
				} else {
					timeoutDuration = 15 * time.Second
				}

				// Check if task has timed out based on last update
				if time.Since(task.UpdatedAt.Time) > timeoutDuration {
					// Create failed consensus response
					consensusResp := consensus_responses.CreateConsensusResponseParams{
						TaskID: taskID,
						Epoch:  task.Epoch,
						Status: "failed",
					}
					
					if _, err := tp.consensusDB.CreateConsensusResponse(context.Background(), consensusResp); err != nil {
						log.Printf("Failed to store failed consensus: %v", err)
					}

					// Update task status to failed
					_, err = tp.taskQueries.UpdateTaskStatus(ctx, tasks.UpdateTaskStatusParams{
						TaskID: taskID,
						Status: "failed",
					})
					if err != nil {
						log.Printf("Failed to update task status to failed: %v", err)
					}

					// Clean up memory
					delete(tp.responses, taskID)
				}
			}
			tp.responsesMutex.Unlock()
		}
	}
}

// checkConfirmations periodically checks block confirmations for tasks in confirming status
func (tp *TaskProcessor) checkConfirmations(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get all tasks in confirming status
			tasks, err := tp.taskQueries.ListConfirmingTasks(ctx)
			if err != nil {
				continue
			}

			for _, task := range tasks {
				// Get state client for the chain
				stateClient, err := tp.node.chainClient.GetStateClient(fmt.Sprintf("%d", task.ChainID))
				if err != nil {
					continue
				}

				// Get latest block number
				latestBlock, err := stateClient.GetLatestBlockNumber(ctx)
				if err != nil {
					continue
				}

				var lastCheckedBlock pgtype.Numeric
				if err := task.LastCheckedBlock.Scan(&lastCheckedBlock); err != nil {
					continue
				}

				// Calculate new confirmations
				var targetBlock uint64
				if err := task.BlockNumber.Scan(&targetBlock); err != nil {
					continue
				}

				newConfirmations := int32(latestBlock - targetBlock)
				if newConfirmations < 0 {
					newConfirmations = 0
				}

				// Convert current confirmations to int32
				var currentConfirmations pgtype.Int4
				if err := task.CurrentConfirmations.Scan(&currentConfirmations); err != nil {
					continue
				}

				var currentValue int32
				if err := currentConfirmations.Scan(&currentValue); err != nil {
					log.Printf("Failed to scan current confirmations value: %v", err)
					continue
				}

				// Update task confirmations
				if newConfirmations > currentValue {
					var lastCheckedBlockNumeric pgtype.Numeric
					if err := lastCheckedBlockNumeric.Scan(fmt.Sprintf("%d", latestBlock)); err != nil {
						log.Printf("Failed to scan last checked block: %v", err)
						return
					}

					var newConfirmationsNumeric pgtype.Int4
					if err := newConfirmationsNumeric.Scan(newConfirmations); err != nil {
						log.Printf("Failed to scan new confirmations: %v", err)
						continue
					}

					// Update task confirmations directly using struct literal
					err = tp.taskQueries.UpdateTaskConfirmations(ctx, struct {
						CurrentConfirmations pgtype.Int4    `json:"current_confirmations"`
						LastCheckedBlock     pgtype.Numeric `json:"last_checked_block"`
						TaskID               string         `json:"task_id"`
					}{
						TaskID:               task.TaskID,
						CurrentConfirmations: pgtype.Int4{Int32: newConfirmations, Valid: true},
						LastCheckedBlock:     lastCheckedBlockNumeric,
					})
					if err != nil {
						log.Printf("Failed to update task confirmations: %v", err)
						continue
					}

					var requiredConfirmations int32
					if err := task.RequiredConfirmations.Scan(&requiredConfirmations); err != nil {
						log.Printf("Failed to scan required confirmations: %v", err)
						continue
					}

					if newConfirmations >= requiredConfirmations {
						// Change task status to pending
						err = tp.taskQueries.UpdateTaskToPending(ctx, task.TaskID)
						if err != nil {
							log.Printf("Failed to update task status to pending: %v", err)
							continue
						}
						tp.logger.Printf("[Confirmation] Task %s has reached required confirmations, changed status to pending", task.TaskID)

						// Query state and process task
						stateClient, err := tp.node.chainClient.GetStateClient(fmt.Sprintf("%d", task.ChainID))
						if err != nil {
							log.Printf("Failed to get state client for chain %d: %v", task.ChainID, err)
							continue
						}

						var blockNum uint64
						if err := task.BlockNumber.Scan(&blockNum); err != nil {
							log.Printf("Failed to scan block number: %v", err)
							continue
						}

						var keyBig big.Int
						if err := task.Key.Scan(&keyBig); err != nil {
							log.Printf("Failed to scan key: %v", err)
							continue
						}

						var blockNumBig big.Int
						if err := task.BlockNumber.Scan(&blockNumBig); err != nil {
							log.Printf("Failed to scan block number: %v", err)
							continue
						}

						state, err := stateClient.GetStateAtBlock(ctx, 
							common.HexToAddress(task.TargetAddress),
							&keyBig,
							blockNum)
						if err != nil {
							log.Printf("Failed to get state at block: %v", err)
							continue
						}

						typesTask := &types.Task{
							ID:            task.TaskID,
							ChainID:       uint64(task.ChainID),
							TargetAddress: task.TargetAddress,
							Key:          &keyBig,
							Value:        state,
							BlockNumber:  &blockNumBig,
							Epoch:        uint32(task.Epoch),
						}

						if err := tp.ProcessTask(ctx, typesTask); err != nil {
							log.Printf("Failed to process task: %v", err)
							continue
						}
					}
				}
			}
		}
	}
}

// Stop gracefully stops the task processor
func (tp *TaskProcessor) Stop() {
	tp.logger.Printf("[TaskProcessor] Stopping task processor")
	
	// Clean up responses map
	tp.responsesMutex.Lock()
	tp.responses = make(map[string]map[string]*types.TaskResponse)
	tp.responsesMutex.Unlock()
	
	tp.logger.Printf("[TaskProcessor] Task processor stopped")
} 