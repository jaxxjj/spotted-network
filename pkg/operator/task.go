package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
)

// ProcessTask processes a new task and broadcasts response
func (tp *TaskProcessor) ProcessTask(ctx context.Context, task *types.Task) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}
	log.Printf("[Task] Starting to process task %s", task.ID)

	// Check if we have already processed this task
	tp.responsesMutex.RLock()
	if responses, exists := tp.responses[task.ID]; exists {
		if _, processed := responses[tp.signer.Address()]; processed {
			tp.responsesMutex.RUnlock()
			log.Printf("[Task] Task %s already processed by this operator", task.ID)
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
		log.Printf("[Task] Task %s already processed and stored in database", task.ID)
		return nil
	}
	
	if task.BlockNumber == nil {
		return fmt.Errorf("task block number is nil")
	}

	// Validate block number
	blockUint64 := task.BlockNumber.Uint64()
	log.Printf("[Task] Processing task for block number: %d", blockUint64)

	// Get state client for chain
	stateClient, err := tp.node.chainClient.GetStateClient(int64(task.ChainID))
	if err != nil {
		return fmt.Errorf("failed to get state client: %v", err)
	}
	log.Printf("[Task] Got state client for chain %d", task.ChainID)

	if task.Key == nil {
		return fmt.Errorf("task key is nil")
	}

	// Get state from chain
	value, err := stateClient.GetStateAtBlock(
		ctx,
		ethcommon.HexToAddress(task.TargetAddress),
		task.Key,
		blockUint64,
	)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}
	log.Printf("[Task] Retrieved state value: %s", value.String())

	// Verify the state matches
	if value.Cmp(task.Value) != 0 {
		return fmt.Errorf("state value mismatch: expected %s, got %s", task.Value, value)
	}
	log.Printf("[Task] Verified state value matches expected value")

	// Sign the response with all required fields
	signParams := signer.TaskSignParams{
		User:        ethcommon.HexToAddress(task.TargetAddress),
		ChainID:     uint32(task.ChainID),
		BlockNumber: blockUint64,
		Key:         task.Key,
		Value:       value,
	}
	
	signature, err := tp.signer.SignTaskResponse(signParams)
	if err != nil {
		return fmt.Errorf("failed to sign response: %w", err)
	}
	log.Printf("[Task] Signed task response")

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
			log.Printf("[Task] Task %s already processed (duplicate key)", task.ID)
			return nil
		}
		return fmt.Errorf("failed to store response: %w", err)
	}
	log.Printf("[Task] Stored response in database")

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
	log.Printf("[Task] Stored own weight %s for task %s", weight.String(), task.ID)

	// Broadcast response
	if err := tp.broadcastResponse(response); err != nil {
		return fmt.Errorf("failed to broadcast response: %w", err)
	}
	log.Printf("[Task] Broadcasted response")

	return nil
}

// ProcessPendingTask processes a single pending task immediately
func (tp *TaskProcessor) ProcessPendingTask(ctx context.Context, task *tasks.Task) error {
	log.Printf("[Task] Processing pending task %s immediately", task.TaskID)
	
	// Get state client for the chain
	stateClient, err := tp.node.chainClient.GetStateClient(int64(task.ChainID))
	if err != nil {
		return fmt.Errorf("[Task] failed to get state client for chain %d: %w", task.ChainID, err)
	}

	// Convert block number to big.Int
	blockNumNumeric := task.BlockNumber
	if !blockNumNumeric.Valid {
		return fmt.Errorf("[Task] block number is null")
	}
	if blockNumNumeric.Int == nil {
		return fmt.Errorf("[Task] block number Int value is null")
	}

	// Handle numeric scale properly
	blockNumStr := common.NumericToString(blockNumNumeric)
	blockNum, _ := new(big.Int).SetString(blockNumStr, 10)

	log.Printf("[Task] Using block number: %s", blockNum.String())

	// Convert key to big.Int
	if !task.Key.Valid {
		return fmt.Errorf("[Task] key is null")
	}
	keyBig := new(big.Int)
	if task.Key.Int == nil {
		return fmt.Errorf("[Task] key Int value is null") 
	}
	if _, ok := keyBig.SetString(task.Key.Int.String(), 10); !ok {
		return fmt.Errorf("[Task] failed to parse key: %s", task.Key.Int.String())
	}
	log.Printf("[Task] Using key: %s", keyBig.String())

	// Get state at block with retries
	blockUint64 := blockNum.Uint64()
	targetAddr := ethcommon.HexToAddress(task.TargetAddress)
	log.Printf("[Task] Getting state for address %s at block %d", targetAddr.Hex(), blockUint64)
	
	state, err := tp.getStateWithRetries(ctx, stateClient, targetAddr, keyBig, blockUint64)
	if err != nil {
		return fmt.Errorf("[Task] failed to get state: %w", err)
	}
	log.Printf("[Task] Successfully retrieved state value: %s", state.String())

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
	log.Printf("[Task] Processing new task %s", task.TaskID)

	// Get state client for the chain
	stateClient, err := tp.node.chainClient.GetStateClient(int64(task.ChainID))
	if err != nil {
		return fmt.Errorf("[Task] failed to get state client for chain %d: %w", task.ChainID, err)
	}

	// Get latest block number
	latestBlock, err := stateClient.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("[Task] failed to get latest block number: %w", err)
	}

	// Get required confirmations
	var requiredConfirmations int32
	if err := task.RequiredConfirmations.Scan(&requiredConfirmations); err != nil {
		return fmt.Errorf("[Task] failed to scan required confirmations: %w", err)
	}

	// Get task block number
	if !task.BlockNumber.Valid || task.BlockNumber.Int == nil {
		return fmt.Errorf("[Task] block number is invalid")
	}
	taskBlock := task.BlockNumber.Int.Uint64()

	// If task has sufficient confirmations, process it immediately
	if latestBlock >= taskBlock {
		log.Printf("[Task] Task %s has sufficient confirmations (latest: %d >= task: %d), processing immediately", 
			task.TaskID, latestBlock, taskBlock)
		return tp.ProcessPendingTask(ctx, task)
	}

	log.Printf("[Task] Task %s needs more confirmations (latest: %d < task: %d), will process later", 
		task.TaskID, latestBlock, taskBlock)
	return nil
}