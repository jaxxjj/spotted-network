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
	log.Printf("[TaskProcessor] Subscribed to response topic")

	// Wait for topic subscription to propagate
	log.Printf("[TaskProcessor] Waiting for topic subscription to propagate...")
	time.Sleep(5 * time.Second)

	// Check initial topic subscription status
	peers := responseTopic.ListPeers()
	log.Printf("[TaskProcessor] Initial topic subscription: 2 peers")
	for _, peer := range peers {
		log.Printf("[TaskProcessor] - Subscribed peer: %s", peer.String())
	}

	// Start goroutines
	go tp.handleResponses(sub)
	go tp.checkTimeouts(context.Background())
	go tp.checkConfirmations(context.Background())
	go tp.checkPendingTasks(context.Background())

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
	tp.logger.Printf("[ProcessTask] Stored response in database")

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
	tp.logger.Printf("[Broadcast] Starting to broadcast response for task %s", resp.TaskID)
	
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

				var lastCheckedBlock pgtype.Numeric
				if err := task.LastCheckedBlock.Scan(&lastCheckedBlock); err != nil {
					tp.logger.Printf("[Confirmation] Failed to scan last checked block: %v", err)
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
					
					var lastCheckedBlockNumeric pgtype.Numeric
					if err := lastCheckedBlockNumeric.Scan(fmt.Sprintf("%d", latestBlock)); err != nil {
						tp.logger.Printf("[Confirmation] Failed to scan last checked block: %v", err)
						continue
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
	tp.logger.Printf("[P2P] 3 peers subscribed to response topic:")
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