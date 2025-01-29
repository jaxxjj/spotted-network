package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"log"
	"strconv"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	commonTypes "github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
)

// StateClient defines the interface for state-related operations
// that the API handler needs
type ChainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	GetStateAtBlock(ctx context.Context, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error)
	Close()
	GetCurrentEpoch(ctx context.Context) (uint32, error)
}

type TaskProcessor interface {
	ProcessTask(ctx context.Context, task *tasks.Tasks) error
}

// ChainManager defines the interface for managing chain clients
// that the API handler needs
type ChainManager interface {
	// GetMainnetClient returns the mainnet client
	GetMainnetClient() (*ethereum.ChainClient, error)
	// GetClientByChainId returns the appropriate client for a given chain ID
	GetClientByChainId(chainID uint32) (*ethereum.ChainClient, error)
}

// TaskQuerier defines the interface for task database operations needed by the handler
type TaskQuerier interface {
	CreateTask(ctx context.Context, arg tasks.CreateTaskParams) (tasks.Tasks, error)
	GetTaskByID(ctx context.Context, taskID string) (tasks.Tasks, error)
}

// ConsensusResponseQuerier defines the interface for consensus response database operations
type ConsensusResponseQuerier interface {
	GetConsensusResponseByTaskId(ctx context.Context, taskID string) (consensus_responses.ConsensusResponse, error)
	GetConsensusResponseByRequest(ctx context.Context, arg consensus_responses.GetConsensusResponseByRequestParams) (consensus_responses.ConsensusResponse, error)
}

type ConsensusResponse struct {
	TaskID              string            `json:"task_id"`
	Epoch              uint32            `json:"epoch"`
	Status             string            `json:"status"`
	Value              string            `json:"value"`
	BlockNumber        uint64            `json:"block_number"`
	ChainID            uint32            `json:"chain_id"`
	TargetAddress      string            `json:"target_address"`
	Key                string            `json:"key"`
	OperatorSignatures []byte 			 `json:"operator_signatures"`
	TotalWeight        string            `json:"total_weight"`
	ConsensusReachedAt time.Time         `json:"consensus_reached_at"`
}

// Handler handles HTTP requests
type Handler struct {
	taskQueries    TaskQuerier
	chainManager   ChainManager
	consensusDB    ConsensusResponseQuerier
	taskProcessor  TaskProcessor
	config        *config.Config
}

type SendRequestParams struct {
	ChainID       uint32 `json:"chain_id"`
	TargetAddress string `json:"target_address"`
	Key           string `json:"key"`
	BlockNumber   uint64 `json:"block_number,omitempty"`
	Timestamp     uint64 `json:"timestamp,omitempty"`
	WaitFinality  bool   `json:"wait_finality"`
}

type SendRequestResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
	RequiredConfirmations uint16 `json:"required_confirmations,omitempty"`
}

// NewHandler creates a new handler
func NewHandler(
	taskQueries TaskQuerier,
	chainManager ChainManager,
	consensusDB ConsensusResponseQuerier,
	taskProcessor TaskProcessor,
	config *config.Config,
) *Handler {
	return &Handler{
		taskQueries:   taskQueries,
		chainManager:  chainManager,
		consensusDB:   consensusDB,
		taskProcessor: taskProcessor,
		config:       config,
	}
}


func (h *Handler) SendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "[API] Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params SendRequestParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "[API] Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request parameters
	if err := h.validateRequest(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get current epoch from mainnet client
	mainnetClient, err := h.chainManager.GetMainnetClient()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mainnet client: %v", err), http.StatusInternalServerError)
		return
	}
	currentEpoch, err := mainnetClient.GetCurrentEpoch(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get current epoch: %v", err), http.StatusInternalServerError)
		return
	}

	// Get state client for chain
	stateClient, err := h.chainManager.GetClientByChainId(params.ChainID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get state client for chain %d: %v", params.ChainID, err), http.StatusBadRequest)
		return
	}

	// Convert timestamp to block number if provided
	var blockNumber uint64
	var timestamp uint64
	targetAddress := ethcommon.HexToAddress(params.TargetAddress)
	if params.Timestamp != 0 {
		// Convert timestamp to block number for task ID generation
		blockNumber, err = commonHelpers.TimestampToBlockNumber(r.Context(), stateClient, params.ChainID, params.Timestamp)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to convert timestamp to block number: %v", err), http.StatusInternalServerError)
			return
		}
		// Update params.BlockNumber for task ID generation
		params.BlockNumber = blockNumber
		timestamp = params.Timestamp
		params.Timestamp = 0
	} else {
		// Use provided block number
		blockNumber = params.BlockNumber
		// Convert block number to timestamp for storage
		timestamp, err = commonHelpers.BlockNumberToTimestamp(r.Context(), stateClient, params.ChainID, blockNumber)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to convert block number to timestamp: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Get state value
	keyBig := commonHelpers.StringToBigInt(params.Key)
	value, err := stateClient.GetStateAtBlock(r.Context(), targetAddress, keyBig, blockNumber)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get state at block: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate task ID using block number
	taskID := h.generateTaskID(&params, value.String())

	// Check if task already exists
	existingTask, err := h.taskQueries.GetTaskByID(r.Context(), taskID)
	if err == nil {
		// Task exists, return its current status
		response := SendRequestResponse{
			TaskID: existingTask.TaskID,
			Status: string(existingTask.Status),
			RequiredConfirmations: uint16(existingTask.RequiredConfirmations),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Determine required confirmations
	requiredConfirmations := h.getRequiredConfirmations(params.ChainID)

	// Get latest block number for confirmation check
	latestBlock, err := stateClient.BlockNumber(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get latest block: %v", err), http.StatusInternalServerError)
		return
	}

	// Determine initial status based on block confirmations
	status := commonTypes.TaskStatusPending
	if params.BlockNumber != 0 {
		// Check if task needs confirmations
		if latestBlock < params.BlockNumber + uint64(requiredConfirmations) {
			status = commonTypes.TaskStatusConfirming
			log.Printf("[API] Task requires %d confirmations, waiting for block %d (current: %d)", 
				requiredConfirmations, params.BlockNumber + uint64(requiredConfirmations), latestBlock)
		} else {
			log.Printf("[API] Task already has sufficient confirmations (target: %d, current: %d)", 
				params.BlockNumber + uint64(requiredConfirmations), latestBlock)
		}
	}

	// Create task
	task, err := h.taskQueries.CreateTask(r.Context(), tasks.CreateTaskParams{
		TaskID:        taskID,
		TargetAddress: params.TargetAddress,
		ChainID:       params.ChainID,
		BlockNumber:   blockNumber,
		Timestamp:     timestamp,
		Epoch:        currentEpoch,
		Key:          commonHelpers.BigIntToNumeric(keyBig),
		Value:        commonHelpers.BigIntToNumeric(value),
		Status:       status,
		RequiredConfirmations: requiredConfirmations,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[API] Created new task %s with status %s", task.TaskID, task.Status)
	if task.Status == commonTypes.TaskStatusConfirming {
		log.Printf("[API] Task %s requires %d block confirmations", task.TaskID, requiredConfirmations)
	} else if task.Status == commonTypes.TaskStatusPending {
		// If task is pending, process it immediately
		log.Printf("[API] Task %s is pending, processing immediately", task.TaskID)
		
		// Use ProcessPendingTask directly with task pointer
		if err := h.taskProcessor.ProcessTask(r.Context(), &task); err != nil {
			log.Printf("[API] Failed to process pending task %s: %v", task.TaskID, err)
			// Don't return error here, as the task is already created
		}
	}

	// Return response
	response := SendRequestResponse{
		TaskID: task.TaskID,
		Status: string(task.Status),
		RequiredConfirmations: task.RequiredConfirmations,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) validateRequest(params *SendRequestParams) error {
	// 1. Validate chain ID (must be positive integer and supported in config)
	if params.ChainID <= 0 {
		return fmt.Errorf("invalid chain ID: must be positive integer")
	}

	// Check if chain ID is supported in config
	if _, ok := h.config.Chains[params.ChainID]; !ok {
		// Get list of supported chains for better error message
		supportedChains := make([]uint32, 0, len(h.config.Chains))
		for chainID := range h.config.Chains {
			supportedChains = append(supportedChains, chainID)
		}
		return fmt.Errorf("unsupported chain ID %d. supported chains: %v", params.ChainID, supportedChains)
	}

	// 2. Validate target address
	if !ethcommon.IsHexAddress(params.TargetAddress) {
		return fmt.Errorf("invalid target address: must be valid EVM address")
	}

	// 3. Validate key (must be valid uint256)
	keyBig := new(big.Int)
	if _, ok := keyBig.SetString(params.Key, 0); !ok {
		return fmt.Errorf("invalid key: must be valid uint256")
	}
	// Check if key is within uint256 range
	if keyBig.Cmp(big.NewInt(0)) < 0 || keyBig.Cmp(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))) > 0 {
		return fmt.Errorf("invalid key: must be within uint256 range")
	}

	// 4. Validate block number and timestamp
	if params.BlockNumber == 0 && params.Timestamp == 0 {
		return fmt.Errorf("either block number or timestamp must be provided")
	}

	if params.BlockNumber != 0 && params.Timestamp != 0 {
		return fmt.Errorf("only one of block number or timestamp should be provided")
	}

	// 5. If block number provided, validate it
	if params.BlockNumber != 0 {
		// Get state client for target chain
		chainClient, err := h.chainManager.GetClientByChainId(params.ChainID)
		if err != nil {
			return fmt.Errorf("failed to get state client: %w", err)
		}

		// Get latest block number
		latestBlock, err := chainClient.BlockNumber(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get latest block: %w", err)
		}

		// Validate block number range
		blockNum := params.BlockNumber
		if blockNum > latestBlock {
			return fmt.Errorf("block number %d is in the future (latest: %d)", blockNum, latestBlock)
		}
	}

	// 6. If timestamp provided, validate it
	if params.Timestamp != 0 {
		// Get current time
		now := time.Now().Unix()

		// Validate timestamp is not in future
		if params.Timestamp > uint64(now) {
			return fmt.Errorf("timestamp is in the future")
		}
	}

	return nil
}

func (h *Handler) generateTaskID(params *SendRequestParams, value string) string {
	// Generate task ID using keccak256(abi.encodePacked(targetAddress, chainId, blockNumber, epoch, key, value))
	data := []byte{}
	data = append(data, ethcommon.HexToAddress(params.TargetAddress).Bytes()...)
	data = append(data, byte(params.ChainID))
	data = append(data, byte(params.BlockNumber))
	data = append(data, []byte(params.Key)...)
	data = append(data, []byte(value)...)

	hash := crypto.Keccak256(data)
	return ethcommon.Bytes2Hex(hash)
}

// GetTaskConsensus returns the consensus result for a task
func (h *Handler) GetTaskConsensusByTaskID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	consensus, err := h.consensusDB.GetConsensusResponseByTaskId(r.Context(), taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get consensus: %v", err), http.StatusInternalServerError)
		return
	}

	// Format response
	response := ConsensusResponse{
		TaskID:              consensus.TaskID,
		Epoch:              consensus.Epoch,
		Value:              commonHelpers.NumericToString(consensus.Value),
		BlockNumber:        consensus.BlockNumber,
		ChainID:           consensus.ChainID,
		TargetAddress:     consensus.TargetAddress,
		Key:               commonHelpers.NumericToString(consensus.Key),
		OperatorSignatures: consensus.AggregatedSignatures,
		TotalWeight:       commonHelpers.NumericToString(consensus.TotalWeight),
		ConsensusReachedAt: consensus.ConsensusReachedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetConsensusResponseByRequest returns the consensus result by request parameters
func (h *Handler) GetConsensusResponseByRequest(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	targetAddress := r.URL.Query().Get("target_address")
	if !ethcommon.IsHexAddress(targetAddress) {
		http.Error(w, "invalid target address", http.StatusBadRequest)
		return
	}

	chainID, err := strconv.ParseUint(r.URL.Query().Get("chain_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid chain_id", http.StatusBadRequest)
		return
	}

	blockNumber, err := strconv.ParseUint(r.URL.Query().Get("block_number"), 10, 64)
	if err != nil {
		http.Error(w, "invalid block_number", http.StatusBadRequest)
		return
	}

	// Parse and validate key
	keyStr := r.URL.Query().Get("key")
	keyBig := commonHelpers.StringToBigInt(keyStr)

	keyNum := pgtype.Numeric{
		Int:   keyBig,
		Valid: true,
	}

	// Get consensus response
	consensus, err := h.consensusDB.GetConsensusResponseByRequest(r.Context(), consensus_responses.GetConsensusResponseByRequestParams{
		TargetAddress: targetAddress,
		ChainID:      uint32(chainID),
		BlockNumber:  blockNumber,
		Key:         keyNum,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get consensus: %v", err), http.StatusInternalServerError)
		return
	}

	// Format response
	response := ConsensusResponse{
		TaskID:              consensus.TaskID,
		Epoch:              consensus.Epoch,
		Value:              commonHelpers.NumericToString(consensus.Value),
		BlockNumber:        consensus.BlockNumber,
		ChainID:           consensus.ChainID,
		TargetAddress:     consensus.TargetAddress,
		Key:               commonHelpers.NumericToString(consensus.Key),
		OperatorSignatures: consensus.AggregatedSignatures,
		TotalWeight:       commonHelpers.NumericToString(consensus.TotalWeight),
		ConsensusReachedAt: consensus.ConsensusReachedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// getRequiredConfirmations returns the required confirmations for a chain
func (h *Handler) getRequiredConfirmations(chainID uint32) uint16 {
	if chainConfig, ok := h.config.Chains[chainID]; ok {
		return chainConfig.RequiredConfirmations
	}
	return 12 // Default to 12 confirmations for unknown chains
} 