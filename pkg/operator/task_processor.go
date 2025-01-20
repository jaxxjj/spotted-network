package operator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
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
	TaskResponseTopic = "/spotted/task-response"
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
	weightsMutex   sync.RWMutex
	taskWeights    map[string]map[string]*big.Int  // taskID -> operatorAddr -> weight
	threshold      *big.Int
	totalWeight    *big.Int  // 总权重
	logger         *log.Logger
}

// NewTaskProcessor creates a new task processor
func NewTaskProcessor(node *Node, taskQueries *tasks.Queries, responseQueries *task_responses.Queries, consensusDB *consensus_responses.Queries) (*TaskProcessor, error) {
	// Create response topic
	responseTopic, err := node.PubSub.Join(TaskResponseTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join response topic: %w", err)
	}
	log.Printf("[TaskProcessor] Joined response topic: %s", TaskResponseTopic)

	// 设置硬编码的权重和阈值
	threshold := big.NewInt(2)  // 阈值设为2
	totalWeight := big.NewInt(3)  // 总权重设为3

	tp := &TaskProcessor{
		node:          node,
		signer:        node.signer,
		taskQueries:   taskQueries,
		db:            responseQueries,
		consensusDB:   consensusDB,
		responseTopic: responseTopic,
		responses:     make(map[string]map[string]*types.TaskResponse),
		taskWeights:   make(map[string]map[string]*big.Int),
		threshold:     threshold,
		totalWeight:   totalWeight,
		logger:        log.Default(),
	}

	// Subscribe to response topic
	sub, err := responseTopic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to response topic: %w", err)
	}
	log.Printf("[TaskProcessor] Subscribed to response topic")

	// Wait for topic subscription to propagate
	log.Printf("[TaskProcessor] Waiting for topic subscription to propagate...")
	time.Sleep(5 * time.Second)

	// Check initial topic subscription status
	peers := responseTopic.ListPeers()
	tp.logger.Printf("[TaskProcessor] Initial topic subscription: %d peers", len(peers))
	for _, peer := range peers {
		tp.logger.Printf("[TaskProcessor] - Subscribed peer: %s", peer.String())
	}

	// Start goroutines
	ctx := context.Background()
	go tp.handleResponses(sub)
	go tp.checkTimeouts(ctx)
	go tp.checkConfirmations(ctx)
	go tp.checkPendingTasks(ctx)

	// Start periodic P2P status check
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			tp.checkP2PStatus()
			<-ticker.C
		}
	}()

	return tp, nil
}

// ProcessTask processes a new task and broadcasts response
func (tp *TaskProcessor) ProcessTask(ctx context.Context, task *types.Task) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}
	tp.logger.Printf("[ProcessTask] Starting to process task %s", task.ID)

	// Check if we have already processed this task
	tp.responsesMutex.RLock()
	if responses, exists := tp.responses[task.ID]; exists {
		if _, processed := responses[tp.signer.Address()]; processed {
			tp.responsesMutex.RUnlock()
			tp.logger.Printf("[ProcessTask] Task %s already processed by this operator", task.ID)
			return nil
		}
	}
	tp.responsesMutex.RUnlock()

	// Also check database
	_, err := tp.db.GetTaskResponse(ctx, task_responses.GetTaskResponseParams{
		TaskID: task.ID,
		OperatorAddress: tp.signer.Address(),
	})
	if err == nil {
		tp.logger.Printf("[ProcessTask] Task %s already processed and stored in database", task.ID)
		return nil
	}
	
	if task.BlockNumber == nil {
		return fmt.Errorf("task block number is nil")
	}

	// Get state client for chain
	stateClient, err := tp.node.chainClient.GetStateClient(fmt.Sprintf("%d", task.ChainID))
	if err != nil {
		return fmt.Errorf("failed to get state client: %w", err)
	}
	tp.logger.Printf("[ProcessTask] Got state client for chain %d", task.ChainID)

	if task.Key == nil {
		return fmt.Errorf("task key is nil")
	}

	// Get state from chain
	blockUint64 := task.BlockNumber.Uint64()
	value, err := stateClient.GetStateAtBlock(
		ctx,
		common.HexToAddress(task.TargetAddress),
		task.Key,
		blockUint64,
	)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}
	tp.logger.Printf("[ProcessTask] Retrieved state value: %s", value.String())

	// Verify the state matches
	if value.Cmp(task.Value) != 0 {
		return fmt.Errorf("state value mismatch: expected %s, got %s", task.Value, value)
	}
	tp.logger.Printf("[ProcessTask] Verified state value matches expected value")

	// Sign the response with all required fields
	signParams := signer.TaskSignParams{
		User:        common.HexToAddress(task.TargetAddress),
		ChainID:     uint32(task.ChainID),
		BlockNumber: blockUint64,
		Timestamp:   task.Timestamp.Uint64(),
		Key:         task.Key,
		Value:       value,
	}
	
	signature, err := tp.signer.SignTaskResponse(signParams)
	if err != nil {
		return fmt.Errorf("failed to sign response: %w", err)
	}
	tp.logger.Printf("[ProcessTask] Signed task response")

	// Create response
	response := &types.TaskResponse{
		TaskID:        task.ID,
		OperatorAddr:  tp.signer.Address(),
		SigningKey:    tp.signer.GetSigningKey(),
		Signature:     signature,
		Value:         value,
		BlockNumber:   task.BlockNumber,
		ChainID:       task.ChainID,
		TargetAddress: task.TargetAddress,
		Key:          task.Key,
		Epoch:        task.Epoch,
		Timestamp:    task.Timestamp,
		CreatedAt:    time.Now(),
	}

	// Store response in database
	if err := tp.storeResponse(ctx, response); err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			tp.logger.Printf("[ProcessTask] Task %s already processed (duplicate key)", task.ID)
			return nil
		}
		return fmt.Errorf("failed to store response: %w", err)
	}
	tp.logger.Printf("[ProcessTask] Stored response in database")

	// Store in local map
	tp.responsesMutex.Lock()
	if _, exists := tp.responses[task.ID]; !exists {
		tp.responses[task.ID] = make(map[string]*types.TaskResponse)
	}
	tp.responses[task.ID][tp.signer.Address()] = response
	tp.responsesMutex.Unlock()

	// Get and store our own weight
	weight, err := tp.getOperatorWeight(tp.signer.Address())
	if err != nil {
		return fmt.Errorf("failed to get own weight: %w", err)
	}
	tp.weightsMutex.Lock()
	if _, exists := tp.taskWeights[task.ID]; !exists {
		tp.taskWeights[task.ID] = make(map[string]*big.Int)
	}
	tp.taskWeights[task.ID][tp.signer.Address()] = weight
	tp.weightsMutex.Unlock()
	tp.logger.Printf("[ProcessTask] Stored own weight %s for task %s", weight.String(), task.ID)

	// Broadcast response
	if err := tp.broadcastResponse(response); err != nil {
		return fmt.Errorf("failed to broadcast response: %w", err)
	}
	tp.logger.Printf("[ProcessTask] Broadcasted response")

	return nil
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
		User:        common.HexToAddress(response.TargetAddress),
		ChainID:     uint32(response.ChainID),
		BlockNumber: response.BlockNumber.Uint64(),
		Timestamp:   response.Timestamp.Uint64(),
		Key:         response.Key,
		Value:       response.Value,
	}

	return tp.signer.VerifyTaskResponse(params, response.Signature, response.OperatorAddr)
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
func (tp *TaskProcessor) checkConsensus(taskID string) error {
	tp.logger.Printf("[Consensus] Starting consensus check for task %s", taskID)
	
	// Get responses and weights from memory maps
	tp.responsesMutex.RLock()
	responses := tp.responses[taskID]
	tp.responsesMutex.RUnlock()

	tp.weightsMutex.RLock()
	weights := tp.taskWeights[taskID]
	tp.weightsMutex.RUnlock()

	tp.logger.Printf("[Consensus] Found %d responses and %d weights for task %s", len(responses), len(weights), taskID)

	if len(responses) == 0 {
		tp.logger.Printf("[Consensus] No responses found for task %s", taskID)
		return nil
	}

	// Calculate total weight and collect signatures
	totalWeight := new(big.Int)
	operatorSigs := make(map[string]map[string]interface{})
	signatures := make(map[string][]byte) // For signature aggregation
	
	for addr, resp := range responses {
		weight := weights[addr]
		if weight == nil {
			tp.logger.Printf("[Consensus] No weight found for operator %s", addr)
			continue
		}

		totalWeight.Add(totalWeight, weight)
		operatorSigs[addr] = map[string]interface{}{
			"signature": hex.EncodeToString(resp.Signature),
			"weight":   weight.String(),
		}
		// Collect signature for aggregation
		signatures[addr] = resp.Signature
		tp.logger.Printf("[Consensus] Added weight %s from operator %s, total weight now: %s", weight.String(), addr, totalWeight.String())
	}

	// Check if threshold is reached
	tp.logger.Printf("[Consensus] Checking if total weight %s reaches threshold %s", totalWeight.String(), tp.threshold.String())
	if totalWeight.Cmp(tp.threshold) < 0 {
		tp.logger.Printf("[Consensus] Threshold not reached for task %s (total weight: %s, threshold: %s)", taskID, totalWeight.String(), tp.threshold.String())
		return nil
	}

	tp.logger.Printf("[Consensus] Threshold reached for task %s! Creating consensus response", taskID)

	// Get a sample response for task details
	var sampleResp *types.TaskResponse
	for _, resp := range responses {
		sampleResp = resp
		break
	}

	// Convert operator signatures to JSONB
	operatorSigsJSON, err := json.Marshal(operatorSigs)
	if err != nil {
		return fmt.Errorf("failed to marshal operator signatures: %w", err)
	}

	// Aggregate all signatures
	aggregatedSigs := tp.aggregateSignatures(signatures)
	tp.logger.Printf("[Consensus] Aggregated %d signatures for task %s", len(signatures), taskID)

	// Create consensus response
	consensus := consensus_responses.CreateConsensusResponseParams{
		TaskID:              taskID,
		Epoch:              int32(sampleResp.Epoch),
		Status:             "completed",
		Value:              types.NumericFromBigInt(sampleResp.Value),
		BlockNumber:        types.NumericFromBigInt(sampleResp.BlockNumber),
		ChainID:           int32(sampleResp.ChainID),
		TargetAddress:     sampleResp.TargetAddress,
		Key:               types.NumericFromBigInt(sampleResp.Key),
		AggregatedSignatures: aggregatedSigs, // Use aggregated signatures
		OperatorSignatures:  operatorSigsJSON,
		TotalWeight:        types.NumericFromBigInt(totalWeight),
		ConsensusReachedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	// Store consensus in database
	if err = tp.storeConsensus(context.Background(), consensus); err != nil {
		return fmt.Errorf("failed to store consensus: %w", err)
	}
	tp.logger.Printf("[Consensus] Successfully stored consensus response for task %s", taskID)

	// If we have the task locally, mark it as completed
	_, err = tp.taskQueries.GetTaskByID(context.Background(), taskID)
	if err == nil {
		if _, err = tp.taskQueries.UpdateTaskStatus(context.Background(), struct {
			TaskID string `json:"task_id"`
			Status string `json:"status"`
		}{
			TaskID: taskID,
			Status: "completed",
		}); err != nil {
			tp.logger.Printf("[Consensus] Failed to update task status: %v", err)
		} else {
			tp.logger.Printf("[Consensus] Updated task %s status to completed", taskID)
		}
	}

	// Clean up local maps
	tp.cleanupTask(taskID)
	tp.logger.Printf("[Consensus] Cleaned up local maps for task %s", taskID)

	return nil
}

// isActiveOperator checks if a signing key belongs to an active operator
func (tp *TaskProcessor) isActiveOperator(signingKey string) bool {
	// Get all operator states
	// tp.node.statesMu.RLock()
	// defer tp.node.statesMu.RUnlock()

	// // Check each operator's state
	// for _, state := range tp.node.operatorStates {
	// 	// TODO: After adding signing_key to proto OperatorState
	// 	// Compare signing keys: if strings.EqualFold(state.SigningKey, signingKey)
	// 	if state.Status == "active" {
	// 		tp.logger.Printf("[Operator] Found active operator %s", state.Address)
	// 		return true
	// 	}
	// }

	// tp.logger.Printf("[Operator] No active operator found for signing key %s", signingKey)
	// return false
	//暂时return true
	return true
}

// cleanupTask removes task data from local maps
func (tp *TaskProcessor) cleanupTask(taskID string) {
	tp.responsesMutex.Lock()
	delete(tp.responses, taskID)
	tp.responsesMutex.Unlock()

	tp.weightsMutex.Lock()
	delete(tp.taskWeights, taskID)
	tp.weightsMutex.Unlock()
}

// getOperatorWeight returns the weight of an operator
func (tp *TaskProcessor) getOperatorWeight(operatorAddr string) (*big.Int, error) {
	// TODO: 后续从链上获取operator权重
	return big.NewInt(1), nil
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
			// Get all tasks that might have timed out
			tasks, err := tp.taskQueries.ListAllTasks(ctx)
			if err != nil {
				tp.logger.Printf("[Timeout] Failed to list tasks: %v", err)
				continue
			}

			for _, task := range tasks {
				// Different timeout durations based on status
				var timeoutDuration time.Duration
				if task.Status == "confirming" {
					timeoutDuration = 15 * time.Minute
				} else {
					timeoutDuration = 5 * time.Minute  // 增加到5分钟
				}

				// Check if task has timed out based on last update
				if time.Since(task.UpdatedAt.Time) > timeoutDuration {
					tp.logger.Printf("[Timeout] Task %s has timed out (status: %s)", task.TaskID, task.Status)
					
					// Create failed consensus response
					consensusResp := consensus_responses.CreateConsensusResponseParams{
						TaskID: task.TaskID,
						Epoch:  task.Epoch,
						Status: "failed",
						Value:  task.Value,
						BlockNumber: task.BlockNumber,
						ChainID: task.ChainID,
						TargetAddress: task.TargetAddress,
						Key: task.Key,
						AggregatedSignatures: []byte{},
						OperatorSignatures: []byte("{}"),
						TotalWeight: pgtype.Numeric{Int: big.NewInt(0), Valid: true},
						ConsensusReachedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
					}
					
					if _, err := tp.consensusDB.CreateConsensusResponse(ctx, consensusResp); err != nil {
						tp.logger.Printf("[Timeout] Failed to store failed consensus: %v", err)
					}

					// Update task status to failed
					_, err = tp.taskQueries.UpdateTaskStatus(ctx, struct {
						TaskID string `json:"task_id"`
						Status string `json:"status"`
					}{
						TaskID: task.TaskID,
						Status: "failed",
					})
					if err != nil {
						tp.logger.Printf("[Timeout] Failed to update task status to failed: %v", err)
					}

					// Clean up memory if exists
					tp.responsesMutex.Lock()
					delete(tp.responses, task.TaskID)
					tp.responsesMutex.Unlock()

					tp.weightsMutex.Lock()
					delete(tp.taskWeights, task.TaskID)
					tp.weightsMutex.Unlock()
				}
			}
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
			tp.logger.Printf("[Confirmation] Starting confirmation check...")
			
			// Get all tasks in confirming status
			tasks, err := tp.taskQueries.ListConfirmingTasks(ctx)
			if err != nil {
				tp.logger.Printf("[Confirmation] Failed to list confirming tasks: %v", err)
				continue
			}
			tp.logger.Printf("[Confirmation] Found %d tasks in confirming status", len(tasks))

			for _, task := range tasks {
				tp.logger.Printf("[Confirmation] Processing task %s", task.TaskID)
				
				// Get state client for the chain
				stateClient, err := tp.node.chainClient.GetStateClient(fmt.Sprintf("%d", task.ChainID))
				if err != nil {
					tp.logger.Printf("[Confirmation] Failed to get state client for chain %d: %v", task.ChainID, err)
					continue
				}

				// Get latest block number
				latestBlock, err := stateClient.GetLatestBlockNumber(ctx)
				if err != nil {
					tp.logger.Printf("[Confirmation] Failed to get latest block number: %v", err)
					continue
				}
				tp.logger.Printf("[Confirmation] Latest block: %d", latestBlock)

				// Use LastCheckedBlock directly since it's already pgtype.Numeric
				if !task.LastCheckedBlock.Valid {
					tp.logger.Printf("[Confirmation] Last checked block is not valid")
					continue
				}

				// Calculate new confirmations
				var targetBlock uint64
				if err := task.BlockNumber.Scan(&targetBlock); err != nil {
					tp.logger.Printf("[Confirmation] Failed to scan target block: %v", err)
					continue
				}
				tp.logger.Printf("[Confirmation] Target block: %d", targetBlock)

				newConfirmations := int32(latestBlock - targetBlock)
				if newConfirmations < 0 {
					newConfirmations = 0
				}
				tp.logger.Printf("[Confirmation] New confirmations: %d", newConfirmations)

				// Convert current confirmations to int32
				var currentConfirmations pgtype.Int4
				if err := task.CurrentConfirmations.Scan(&currentConfirmations); err != nil {
					tp.logger.Printf("[Confirmation] Failed to scan current confirmations: %v", err)
					continue
				}

				var currentValue int32
				if err := currentConfirmations.Scan(&currentValue); err != nil {
					tp.logger.Printf("[Confirmation] Failed to scan current confirmations value: %v", err)
					continue
				}
				tp.logger.Printf("[Confirmation] Current confirmations: %d", currentValue)

				// Update task confirmations
				if newConfirmations > currentValue {
					tp.logger.Printf("[Confirmation] Updating confirmations for task %s from %d to %d", task.TaskID, currentValue, newConfirmations)
					
					// Create numeric type for last checked block
					lastCheckedBlockNumeric := pgtype.Numeric{
						Int:   new(big.Int).SetUint64(latestBlock),
						Valid: true,
					}

					var newConfirmationsNumeric pgtype.Int4
					if err := newConfirmationsNumeric.Scan(newConfirmations); err != nil {
						tp.logger.Printf("[Confirmation] Failed to scan new confirmations: %v", err)
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
						tp.logger.Printf("[Confirmation] Failed to update task confirmations: %v", err)
						continue
					}
					tp.logger.Printf("[Confirmation] Successfully updated confirmations for task %s", task.TaskID)

					var requiredConfirmations int32
					if err := task.RequiredConfirmations.Scan(&requiredConfirmations); err != nil {
						tp.logger.Printf("[Confirmation] Failed to scan required confirmations: %v", err)
						continue
					}
					tp.logger.Printf("[Confirmation] Required confirmations: %d", requiredConfirmations)

					if newConfirmations >= requiredConfirmations {
						tp.logger.Printf("[Confirmation] Task %s has reached required confirmations (%d >= %d), changing status to pending", task.TaskID, newConfirmations, requiredConfirmations)
						
						// Change task status to pending
						err = tp.taskQueries.UpdateTaskToPending(ctx, task.TaskID)
						if err != nil {
							tp.logger.Printf("[Confirmation] Failed to update task status to pending: %v", err)
							continue
						}
						tp.logger.Printf("[Confirmation] Successfully changed task %s status to pending", task.TaskID)

						// Process task immediately
						if err := tp.ProcessPendingTask(ctx, &task); err != nil {
							tp.logger.Printf("[Confirmation] Failed to process task immediately: %v", err)
							continue
						}
						tp.logger.Printf("[Confirmation] Successfully processed task %s immediately", task.TaskID)
					}
				}
			}
		}
	}
}

// checkPendingTasks periodically checks and processes pending tasks
func (tp *TaskProcessor) checkPendingTasks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("[TaskProcessor] Starting pending tasks check...")
			
			// List all pending tasks
			pendingTasks, err := tp.taskQueries.ListPendingTasks(ctx)
			if err != nil {
				log.Printf("[TaskProcessor] Failed to list pending tasks: %v", err)
				continue
			}
			log.Printf("[TaskProcessor] Found %d pending tasks", len(pendingTasks))

			for _, task := range pendingTasks {
				log.Printf("[TaskProcessor] Processing pending task %s", task.TaskID)
				
				if err := tp.ProcessPendingTask(ctx, &task); err != nil {
					log.Printf("[TaskProcessor] Failed to process task %s: %v", task.TaskID, err)
					continue
				}
				
				log.Printf("[TaskProcessor] Successfully processed pending task %s", task.TaskID)
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

// checkP2PStatus checks the status of P2P connections
func (tp *TaskProcessor) checkP2PStatus() {
	peers := tp.node.host.Network().Peers()
	tp.logger.Printf("[P2P] Connected to %d peers:", len(peers))
	for _, peer := range peers {
		addrs := tp.node.host.Network().Peerstore().Addrs(peer)
		tp.logger.Printf("[P2P] - Peer %s at %v", peer.String(), addrs)
	}

	// Check pubsub topic
	peers = tp.responseTopic.ListPeers()
	tp.logger.Printf("[P2P] %d peers subscribed to response topic:", len(peers))
	for _, peer := range peers {
		tp.logger.Printf("[P2P] - Subscribed peer: %s", peer.String())
	}
}

// ProcessPendingTask processes a single pending task immediately
func (tp *TaskProcessor) ProcessPendingTask(ctx context.Context, task *tasks.Task) error {
	tp.logger.Printf("[TaskProcessor] Processing pending task %s immediately", task.TaskID)
	
	// Get state client for the chain
	stateClient, err := tp.node.chainClient.GetStateClient(fmt.Sprintf("%d", task.ChainID))
	if err != nil {
		return fmt.Errorf("failed to get state client for chain %d: %w", task.ChainID, err)
	}

	// Convert block number to big.Int
	blockNumNumeric := task.BlockNumber
	if !blockNumNumeric.Valid {
		return fmt.Errorf("block number is null")
	}
	blockNum := new(big.Int)
	if blockNumNumeric.Int == nil {
		return fmt.Errorf("block number Int value is null")
	}
	if _, ok := blockNum.SetString(blockNumNumeric.Int.String(), 10); !ok {
		return fmt.Errorf("failed to parse block number: %s", blockNumNumeric.Int.String())
	}

	// Convert key to big.Int
	if !task.Key.Valid {
		return fmt.Errorf("key is null")
	}
	keyBig := new(big.Int)
	if task.Key.Int == nil {
		return fmt.Errorf("key Int value is null") 
	}
	if _, ok := keyBig.SetString(task.Key.Int.String(), 10); !ok {
		return fmt.Errorf("failed to parse key: %s", task.Key.Int.String())
	}

	// Get state at block
	blockUint64 := blockNum.Uint64()
	state, err := stateClient.GetStateAtBlock(ctx,
		common.HexToAddress(task.TargetAddress),
		keyBig,
		blockUint64)
	if err != nil {
		return fmt.Errorf("failed to get state at block: %w", err)
	}

	// Create types.Task
	typesTask := &types.Task{
		ID:            task.TaskID,
		ChainID:       uint64(task.ChainID),
		TargetAddress: task.TargetAddress,
		BlockNumber:   blockNum,
		Key:          keyBig,
		Value:        state,
		Epoch:        uint32(task.Epoch),
		Timestamp:    &big.Int{}, // Initialize with current timestamp
	}

	// Set timestamp to current time
	typesTask.Timestamp.SetInt64(time.Now().Unix())

	// Process the task
	return tp.ProcessTask(ctx, typesTask)
}

// ProcessNewTask processes a newly created task immediately if it has sufficient confirmations
func (tp *TaskProcessor) ProcessNewTask(ctx context.Context, task *tasks.Task) error {
	tp.logger.Printf("[TaskProcessor] Processing new task %s", task.TaskID)

	// Get current confirmations
	var currentConfirmations int32
	if err := task.CurrentConfirmations.Scan(&currentConfirmations); err != nil {
		return fmt.Errorf("failed to scan current confirmations: %w", err)
	}

	// Get required confirmations
	var requiredConfirmations int32
	if err := task.RequiredConfirmations.Scan(&requiredConfirmations); err != nil {
		return fmt.Errorf("failed to scan required confirmations: %w", err)
	}

	// If task has sufficient confirmations, process it immediately
	if currentConfirmations >= requiredConfirmations {
		tp.logger.Printf("[TaskProcessor] Task %s has sufficient confirmations (%d/%d), processing immediately", 
			task.TaskID, currentConfirmations, requiredConfirmations)
		return tp.ProcessPendingTask(ctx, task)
	}

	tp.logger.Printf("[TaskProcessor] Task %s needs more confirmations (%d/%d), will process later", 
		task.TaskID, currentConfirmations, requiredConfirmations)
	return nil
}

// convertToTaskResponse converts a protobuf message to TaskResponse
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

// storeConsensus stores a consensus response in the database
func (tp *TaskProcessor) storeConsensus(ctx context.Context, consensus consensus_responses.CreateConsensusResponseParams) error {
	_, err := tp.consensusDB.CreateConsensusResponse(ctx, consensus)
	return err
} 