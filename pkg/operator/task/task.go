package task

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	pb "github.com/galxe/spotted-network/proto"
	"google.golang.org/protobuf/proto"
)

type OperatorRepo interface {
	GetOperatorBySigningKey(ctx context.Context, signingKey string) (*operators.Operators, error)
	GetOperatorByP2PKey(ctx context.Context, p2pKey string) (*operators.Operators, error)
}

type ChainClient interface {
	GetStateAtBlock(ctx context.Context, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
}
type ChainManager interface {
	// GetMainnetClient returns the mainnet client
	GetMainnetClient() (ethereum.ChainClient, error)
	// GetClientByChainId returns the appropriate client for a given chain ID
	GetClientByChainId(chainID uint32) (ethereum.ChainClient, error)
	Close() error
}
type OperatorSigner interface {
	SignTaskResponse(params signer.TaskSignParams) ([]byte, error)
	VerifyTaskResponse(params signer.TaskSignParams, signature []byte, signerAddr string) error
	GetSigningAddress() ethcommon.Address
}

// ProcessTask processes a new task and broadcasts response
func (tp *taskProcessor) ProcessTask(ctx context.Context, task *tasks.Tasks) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}
	if tp.alreadyProcessed(task.TaskID, tp.signer.GetSigningAddress().Hex()) {
		log.Printf("[Task] Task %s already processed by signing key %s", task.TaskID, tp.signer.GetSigningAddress().Hex())
		return nil
	}
	log.Printf("[Task] Starting to process task %s", task.TaskID)

	blockNumber := task.BlockNumber
	if blockNumber == 0 {
		return fmt.Errorf("task block number is nil")
	}

	log.Printf("[Task] Processing task for block number: %d", blockNumber)

	// Get state client for chain
	chainClient, err := tp.chainManager.GetClientByChainId(task.ChainID)
	if err != nil {
		return fmt.Errorf("failed to get state client: %v", err)
	}
	log.Printf("[Task] Got state client for chain %d", task.ChainID)
	KeyBig, err := utils.NumericToBigInt(task.Key)
	if err != nil {
		return fmt.Errorf("failed to convert key to big.Int: %v", err)
	}
	// Get state from chain
	value, err := getStateWithRetries(
		ctx,
		chainClient,
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
	response := taskResponse{
		taskID:        task.TaskID,
		signature:     signature,
		value:         value.String(),
		blockNumber:   task.BlockNumber,
		chainID:       task.ChainID,
		targetAddress: task.TargetAddress,
		key:           utils.NumericToString(task.Key),
		epoch:         task.Epoch,
	}

	operatorSigningKey := tp.signer.GetSigningAddress().Hex()

	// Get and store our own weight
	operator, err := tp.operatorRepo.GetOperatorBySigningKey(ctx, operatorSigningKey)
	if err != nil {
		return fmt.Errorf("failed to get operator: %w", err)
	}

	weight, err := utils.NumericToBigInt(operator.Weight)
	if err != nil {
		return fmt.Errorf("failed to convert operator weight to big int: %w", err)
	}

	// Store in local map with proper initialization
	tp.taskResponseTrack.mu.Lock()
	if _, exists := tp.taskResponseTrack.responses[task.TaskID]; !exists {
		tp.taskResponseTrack.responses[task.TaskID] = make(map[string]taskResponse)
		tp.taskResponseTrack.weights[task.TaskID] = make(map[string]*big.Int)
	}
	tp.taskResponseTrack.responses[task.TaskID][operatorSigningKey] = response
	tp.taskResponseTrack.weights[task.TaskID][operatorSigningKey] = weight
	tp.taskResponseTrack.mu.Unlock()

	log.Printf("[Task] Stored own response and weight %s for task %s with key %s",
		weight, task.TaskID, operatorSigningKey)

	// Broadcast response
	if err := tp.broadcastResponse(response); err != nil {
		return fmt.Errorf("failed to broadcast response: %w", err)
	}
	log.Printf("[Task] Broadcasted response")
	tp.checkConsensus(ctx, response)
	return nil
}

// broadcastResponse broadcasts a task response to other operators
func (tp *taskProcessor) broadcastResponse(response taskResponse) error {
	log.Printf("[Response] Starting to broadcast response for task %s", response.taskID)

	// Create protobuf message
	msg := &pb.TaskResponseMessage{
		TaskId:        response.taskID,
		Signature:     response.signature,
		Value:         response.value,
		BlockNumber:   response.blockNumber,
		ChainId:       response.chainID,
		TargetAddress: response.targetAddress,
		Key:           response.key,
		Epoch:         response.epoch,
	}

	// Marshal message
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	log.Printf("[Broadcast] Marshaled response message for task %s", response.taskID)

	// Publish to topic
	if err := tp.responseTopic.Publish(context.Background(), data); err != nil {
		log.Printf("[Response] Failed to publish response: %v", err)
		return err
	}
	log.Printf("[Response] Successfully published response for task %s to topic %s", response.taskID, tp.responseTopic.String())

	return nil
}

// getStateWithRetries attempts to get chain state with retries
func getStateWithRetries(ctx context.Context, chainClient ChainClient, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	maxRetries := 3
	retryDelay := 100 * time.Millisecond

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
