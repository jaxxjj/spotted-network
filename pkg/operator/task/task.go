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
)
type OperatorSigner interface{
	SignTaskResponse(params signer.TaskSignParams) ([]byte, error)
	VerifyTaskResponse(params signer.TaskSignParams, signature []byte, signerAddr string) error
	AggregateSignatures(sigs map[string][]byte) []byte
	GetOperatorAddress() ethcommon.Address
	GetSigningAddress() ethcommon.Address
	Sign(message []byte) ([]byte, error)
}

// ProcessTask processes a new task and broadcasts response
func (tp *TaskProcessor) ProcessTask(ctx context.Context, task *tasks.Tasks) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}
	log.Printf("[Task] Starting to process task %s", task.TaskID)

	blockNumber := task.BlockNumber
	if blockNumber == 0 {
		return fmt.Errorf("task block number is nil")
	}

	log.Printf("[Task] Processing task for block number: %d", blockNumber)

	// Get state client for chain
	stateClient, err := tp.chainManager.GetClientByChainId(task.ChainID)
	if err != nil {
		return fmt.Errorf("failed to get state client: %v", err)
	}
	log.Printf("[Task] Got state client for chain %d", task.ChainID)
	KeyBig, err := utils.NumericToBigInt(task.Key)
	if err != nil {
		return fmt.Errorf("failed to convert key to big.Int: %v", err)
	}
	// Get state from chain
	value, err := tp.getStateWithRetries(
		ctx,
		stateClient,
		ethcommon.HexToAddress(task.TargetAddress),
		KeyBig,
		blockNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}
	log.Printf("[Task] Retrieved state value: %s", value.String())
	taskValueBig, err := utils.NumericToBigInt(task.Value)
	if err != nil {
		return fmt.Errorf("failed to convert task value to big.Int: %v", err)
	}
	// Verify the state matches
	if value.Cmp(taskValueBig) != 0 {
		return fmt.Errorf("state value mismatch: expected %s, got %s", utils.NumericToString(task.Value), value)
	}
	log.Printf("[Task] Verified state value matches expected value")

	// Sign the response with all required fields
	signParams := signer.TaskSignParams{
		User:        ethcommon.HexToAddress(task.TargetAddress),
		ChainID:     uint32(task.ChainID),
		BlockNumber: blockNumber,
		Key:         KeyBig,
		Value:       value,
	}
	
	signature, err := tp.signer.SignTaskResponse(signParams)
	if err != nil {
		return fmt.Errorf("failed to sign response: %w", err)
	}
	log.Printf("[Task] Signed task response")

	// Create response
	response := &task_responses.TaskResponses{
		TaskID:        task.TaskID,
		OperatorAddress:  tp.signer.GetOperatorAddress().Hex(),
		SigningKey:    tp.signer.GetSigningAddress().Hex(),
		Signature:     signature,
		Value:         utils.BigIntToNumeric(value),
		BlockNumber:   task.BlockNumber,
		ChainID:       task.ChainID,
		TargetAddress: task.TargetAddress,
		Key:          task.Key,
		Epoch:        task.Epoch,
		Timestamp:    task.Timestamp,
		SubmittedAt:  time.Now(),
	}

	// Store response in database
	if err := tp.storeResponse(ctx, response); err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			log.Printf("[Task] Task %s already processed (duplicate key)", task.TaskID)
			return nil
		}
		return fmt.Errorf("failed to store response: %w", err)
	}
	log.Printf("[Task] Stored response in database")

	// Store in local map
	tp.responsesMutex.Lock()
	if _, exists := tp.responses[task.TaskID]; !exists {
		tp.responses[task.TaskID] = make(map[string]*task_responses.TaskResponses)
	}
	tp.responses[task.TaskID][tp.signer.GetOperatorAddress().Hex()] = response
	tp.responsesMutex.Unlock()

	// Get and store our own weight
	weight, err := tp.node.getOperatorWeight(tp.node.getHostID())
	if err != nil {
		return fmt.Errorf("failed to get own weight: %w", err)
	}
	tp.weightsMutex.Lock()
	if _, exists := tp.taskWeights[task.TaskID]; !exists {
		tp.taskWeights[task.TaskID] = make(map[string]*big.Int)
	}
	tp.taskWeights[task.TaskID][tp.signer.GetOperatorAddress().Hex()] = weight
	tp.weightsMutex.Unlock()
	log.Printf("[Task] Stored own weight %s for task %s", weight.String(), task.TaskID)

	// Broadcast response
	if err := tp.broadcastResponse(response); err != nil {
		return fmt.Errorf("failed to broadcast response: %w", err)
	}
	log.Printf("[Task] Broadcasted response")

	return nil
}

// getStateWithRetries attempts to get state with retries
func (tp *TaskProcessor) getStateWithRetries(ctx context.Context, chainClient ChainClient, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		// Get latest block for validation
		latestBlock, err := chainClient.BlockNumber(ctx)
		if err != nil {
			log.Printf("[StateCheck] Failed to get latest block: %v", err)
			time.Sleep(retryDelay)
			continue
		}

		// Validate block number
		if blockNumber > latestBlock {
			return nil, fmt.Errorf("block number %d is in the future (latest: %d)", blockNumber, latestBlock)
		}

		// Attempt to get state
		state, err := chainClient.GetStateAtBlock(ctx, target, key, blockNumber)
		if err != nil {
			// Check for specific contract errors
			if strings.Contains(err.Error(), "0x7c44ec9a") { // StateManager__BlockNotFound
				return nil, fmt.Errorf("block %d not found in state history", blockNumber)
			}
			if strings.Contains(err.Error(), "StateManager__KeyNotFound") {
				return nil, fmt.Errorf("key %s not found for address %s", key.String(), target.Hex())
			}
			if strings.Contains(err.Error(), "StateManager__NoHistoryFound") {
				return nil, fmt.Errorf("no state history found for block %d and key %s", blockNumber, key.String())
			}

			log.Printf("[StateCheck] Attempt %d failed: %v", i+1, err)
			time.Sleep(retryDelay)
			continue
		}

		return state, nil
	}

	return nil, fmt.Errorf("failed to get state after %d retries", maxRetries)
}