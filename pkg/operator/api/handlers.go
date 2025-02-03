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
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	commonTypes "github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
)

// chain client interface
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

// ChainManager helps to manage chain clients
type ChainManager interface {
	// GetMainnetClient returns the mainnet client
	GetMainnetClient() (*ethereum.ChainClient, error)
	// GetClientByChainId returns the appropriate client for a given chain ID
	GetClientByChainId(chainID uint32) (*ethereum.ChainClient, error)
}

// TaskQuerier defines the interface for task database operations needed by the handler
type TasksQuerier interface {
	CreateTask(ctx context.Context, arg tasks.CreateTaskParams) (*tasks.Tasks, error) 
	GetTaskByID(ctx context.Context, taskID string) (*tasks.Tasks, error)
}

// ConsensusResponseQuerier defines the interface for consensus response database operations
type ConsensusResponseQuerier interface {
	GetConsensusResponseByTaskId(ctx context.Context, taskID string) (*consensus_responses.ConsensusResponse, error)
	GetConsensusResponseByRequest(ctx context.Context, arg consensus_responses.GetConsensusResponseByRequestParams) (*consensus_responses.ConsensusResponse, error)
}

// Handler handles HTTP requests
type Handler struct {
	tasks    TasksQuerier
	chainManager   ChainManager
	consensusResponses    ConsensusResponseQuerier
	taskProcessor  TaskProcessor
	config        *config.Config
}

// user send request params
type SendRequestParams struct {
	ChainID       uint32 `json:"chain_id"`
	TargetAddress string `json:"target_address"`
	Key           string `json:"key"`
	BlockNumber   uint64 `json:"block_number,omitempty"`
	Timestamp     uint64 `json:"timestamp,omitempty"`
	WaitFinality  bool   `json:"wait_finality"`
}

// send request response
type SendRequestResponse struct {
	TaskID               string `json:"task_id"`
	Status               string `json:"status"`
	RequiredConfirmations uint16 `json:"required_confirmations,omitempty"`
	Message              string `json:"message,omitempty"`
	Error               string `json:"error,omitempty"`
}

// consensus response wrapper
type ConsensusResponseWrapper struct {
	Data    *consensus_responses.ConsensusResponse `json:"data,omitempty"`
	Message string                                `json:"message,omitempty"`
	Error   string                                `json:"error,omitempty"`
}

// NewHandler creates a new handler
func NewHandler(
	tasks TasksQuerier,
	chainManager ChainManager,
	consensusResponses ConsensusResponseQuerier,
	taskProcessor TaskProcessor,
	config *config.Config,
) *Handler {
	if taskProcessor == nil {
		log.Fatal("[API] Task processor not initialized")
	}
	if chainManager == nil {
		log.Fatal("[API] Chain manager not initialized")
	}

	if tasks == nil {
		log.Fatal("[API] Tasks querier not initialized")
	}
	if consensusResponses == nil {
		log.Fatal("[API] Consensus responses querier not initialized")
	}
	return &Handler{
		tasks:   tasks,
		chainManager:  chainManager,
		consensusResponses:   consensusResponses,
		taskProcessor: taskProcessor,
		config:       config,
	}
}

// send request handler
func (h *Handler) SendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var params SendRequestParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validateRequest(&params); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	mainnetClient, err := h.chainManager.GetMainnetClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get mainnet client: %v", err))
		return
	}
	if mainnetClient == nil {
		log.Fatal("[API] Mainnet client not initialized")
	}

	currentEpoch, err := mainnetClient.GetCurrentEpoch(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get current epoch: %v", err))
		return
	}

	stateClient, err := h.chainManager.GetClientByChainId(params.ChainID)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to get state client for chain %d: %v", params.ChainID, err))
		return
	}
	if stateClient == nil {
		log.Printf("[API] State client is nil for chain %d", params.ChainID)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var blockNumber uint64
	var timestamp uint64
	targetAddress := ethcommon.HexToAddress(params.TargetAddress)
	if params.Timestamp != 0 {
		blockNumber, err = utils.TimestampToBlockNumber(r.Context(), stateClient, params.ChainID, params.Timestamp)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to convert timestamp to block number: %v", err))
			return
		}
		params.BlockNumber = blockNumber
		timestamp = params.Timestamp
		params.Timestamp = 0
	} else {
		blockNumber = params.BlockNumber
		timestamp, err = utils.BlockNumberToTimestamp(r.Context(), stateClient, params.ChainID, blockNumber)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to convert block number to timestamp: %v", err))
			return
		}
	}

	keyBig := utils.StringToBigInt(params.Key)
	value, err := stateClient.GetStateAtBlock(r.Context(), targetAddress, keyBig, blockNumber)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get state at block: %v", err))
		return
	}

	taskID := h.generateTaskID(&params, value.String())

	// check consensus reached
	consensus, err := h.consensusResponses.GetConsensusResponseByTaskId(r.Context(), taskID)
	if err != nil {
		log.Printf("[API] Failed to query consensus: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	
	if consensus != nil {
		writeSuccess(w, 
			taskID,
			string(commonTypes.TaskStatusCompleted),
			0, 
			fmt.Sprintf("Task already completed with consensus reached at %s", 
				consensus.ConsensusReachedAt.Format(time.RFC3339)),
		)
		return
	}

	// check task exists
	existingTask, err := h.tasks.GetTaskByID(r.Context(), taskID)
	if err != nil {
		log.Printf("[API] Failed to query task: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if existingTask != nil {
		// task exists
		writeSuccess(w, 
			existingTask.TaskID,
			string(existingTask.Status),
			uint16(existingTask.RequiredConfirmations),
			fmt.Sprintf("Task already exists with status: %s", existingTask.Status),
		)
		return
	}

	requiredConfirmations := h.getRequiredConfirmations(params.ChainID)

	latestBlock, err := stateClient.BlockNumber(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get latest block: %v", err))
		return
	}

	status := commonTypes.TaskStatusPending
	var statusMessage string
	if params.BlockNumber != 0 {
		if latestBlock < params.BlockNumber+uint64(requiredConfirmations) {
			status = commonTypes.TaskStatusConfirming
			statusMessage = fmt.Sprintf("Waiting for %d block confirmations", requiredConfirmations)
			log.Printf("[API] Task requires %d confirmations, waiting for block %d (current: %d)",
				requiredConfirmations, params.BlockNumber+uint64(requiredConfirmations), latestBlock)
		} else {
			statusMessage = "Task has sufficient confirmations"
			log.Printf("[API] Task already has sufficient confirmations (target: %d, current: %d)",
				params.BlockNumber+uint64(requiredConfirmations), latestBlock)
		}
	} else {
		statusMessage = "Task is pending processing"
	}

	task, err := h.tasks.CreateTask(r.Context(), tasks.CreateTaskParams{
		TaskID:               taskID,
		TargetAddress:        params.TargetAddress,
		ChainID:             params.ChainID,
		BlockNumber:         blockNumber,
		Timestamp:           timestamp,
		Epoch:               currentEpoch,
		Key:                 utils.BigIntToNumeric(keyBig),
		Value:               utils.BigIntToNumeric(value),
		Status:              status,
		RequiredConfirmations: requiredConfirmations,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create task: %v", err))
		return
	}

	if task == nil {
		log.Printf("[API] CreateTask returned nil task without error")
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Printf("[API] Created new task %s with status %s", task.TaskID, task.Status)
	if task.Status == commonTypes.TaskStatusConfirming {
		log.Printf("[API] Task %s requires %d block confirmations", task.TaskID, requiredConfirmations)
	} else if task.Status == commonTypes.TaskStatusPending {
		log.Printf("[API] Task %s is pending, processing immediately", task.TaskID)
		if err := h.taskProcessor.ProcessTask(r.Context(), task); err != nil {
			log.Printf("[API] Failed to process pending task %s: %v", task.TaskID, err)
			// Don't return error here, as the task is already created
			statusMessage = fmt.Sprintf("Task created but processing failed: %v", err)
		}
	}

	writeSuccess(w, 
		task.TaskID,
		string(task.Status),
		task.RequiredConfirmations,
		statusMessage,
	)
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

// GetTaskConsensusByTaskID returns the consensus result for a task
func (h *Handler) GetTaskConsensusByTaskID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	consensus, err := h.consensusResponses.GetConsensusResponseByTaskId(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get consensus: %v", err))
		return
	}
	if consensus == nil {
		writeError(w, http.StatusNotFound, "Consensus not found")
		return
	}

	response := ConsensusResponseWrapper{
		Data:    consensus,
		Message: "Consensus found successfully",
	}
	
	writeJSON(w, http.StatusOK, response)
}

// GetConsensusResponseByRequest returns the consensus result by request parameters
func (h *Handler) GetConsensusResponseByRequest(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	targetAddress := r.URL.Query().Get("target_address")
	if !ethcommon.IsHexAddress(targetAddress) {
		writeError(w, http.StatusBadRequest, "Invalid target address")
		return
	}

	chainID, err := strconv.ParseUint(r.URL.Query().Get("chain_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid chain_id")
		return
	}

	blockNumber, err := strconv.ParseUint(r.URL.Query().Get("block_number"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid block_number")
		return
	}

	// Parse and validate key
	keyStr := r.URL.Query().Get("key")
	keyBig := utils.StringToBigInt(keyStr)

	keyNum := pgtype.Numeric{
		Int:   keyBig,
		Valid: true,
	}

	// Get consensus response
	consensus, err := h.consensusResponses.GetConsensusResponseByRequest(r.Context(), consensus_responses.GetConsensusResponseByRequestParams{
		TargetAddress: targetAddress,
		ChainID:      uint32(chainID),
		BlockNumber:  blockNumber,
		Key:         keyNum,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get consensus: %v", err))
		return
	}
	if consensus == nil {
		writeError(w, http.StatusNotFound, "Consensus not found")
		return
	}

	response := ConsensusResponseWrapper{
		Data:    consensus,
		Message: "Consensus found successfully",
	}
	
	writeJSON(w, http.StatusOK, response)
}


