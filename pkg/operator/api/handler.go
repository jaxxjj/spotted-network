package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"log"

	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
)

type TaskFinalResponse struct {
	TaskID              string            `json:"task_id"`
	Epoch               uint32            `json:"epoch"`
	Status             string            `json:"status"`
	Value              string            `json:"value"`
	BlockNumber        uint64            `json:"block_number"`
	ChainID            uint64            `json:"chain_id"`
	TargetAddress      string            `json:"target_address"`
	Key                string            `json:"key"`
	OperatorSignatures map[string][]byte `json:"operator_signatures"`
	TotalWeight        string            `json:"total_weight"`
	ConsensusReachedAt time.Time         `json:"consensus_reached_at"`
}

// Handler handles HTTP requests
type Handler struct {
	taskQueries *tasks.Queries
	chainClient *ethereum.ChainClients
	consensusDB *consensus_responses.Queries
	taskProcessor interface {
		ProcessTask(ctx context.Context, task *types.Task) error
		ProcessPendingTask(ctx context.Context, task *tasks.Task) error
	}
	config *config.Config
}

// NewHandler creates a new handler
func NewHandler(taskQueries *tasks.Queries, chainClient *ethereum.ChainClients, consensusDB *consensus_responses.Queries, taskProcessor interface {
	ProcessTask(ctx context.Context, task *types.Task) error
	ProcessPendingTask(ctx context.Context, task *tasks.Task) error
}, config *config.Config) *Handler {
	return &Handler{
		taskQueries: taskQueries,
		chainClient: chainClient,
		consensusDB: consensusDB,
		taskProcessor: taskProcessor,
		config: config,
	}
}

type SendRequestParams struct {
	ChainID       int64  `json:"chain_id"`
	TargetAddress string `json:"target_address"`
	Key           string `json:"key"`
	BlockNumber   *int64 `json:"block_number,omitempty"`
	Timestamp     *int64 `json:"timestamp,omitempty"`
	WaitFinality  bool   `json:"wait_finality"`
}

type SendRequestResponse struct {
	TaskID string `json:"task_id"`
}

func (h *Handler) SendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params SendRequestParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request parameters
	if err := h.validateRequest(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get mainnet client for epoch
	mainnetClient := h.chainClient.GetMainnetClient()
	
	// Get current epoch
	currentEpoch, err := mainnetClient.GetCurrentEpoch(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get current epoch: %v", err), http.StatusInternalServerError)
		return
	}

	// Get state client for chain
	stateClient, err := h.chainClient.GetStateClient(params.ChainID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get state client: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert timestamp to block number if provided
	var blockNumber uint64
	var timestamp *int64
	if params.Timestamp != nil {
		// Convert timestamp to block number for task ID generation
		blockNumber, err = commonHelpers.TimestampToBlockNumber(r.Context(), stateClient, params.ChainID, *params.Timestamp)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to convert timestamp to block number: %v", err), http.StatusInternalServerError)
			return
		}
		// Update params.BlockNumber for task ID generation
		blockNumberInt64 := int64(blockNumber)
		params.BlockNumber = &blockNumberInt64
		timestamp = params.Timestamp
		params.Timestamp = nil
	} else {
		// Use provided block number
		blockNumber = uint64(*params.BlockNumber)
		// Convert block number to timestamp for storage
		timestampInt64, err := commonHelpers.BlockNumberToTimestamp(r.Context(), stateClient, params.ChainID, blockNumber)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to convert block number to timestamp: %v", err), http.StatusInternalServerError)
			return
		}
		timestamp = &timestampInt64
	}

	// Generate task ID using block number
	taskID := h.generateTaskID(&params, int64(currentEpoch), "0")

	// Create task params
	blockNumberNumeric := pgtype.Numeric{
		Int:   new(big.Int).SetUint64(blockNumber),
		Valid: true,
		Exp:   0,
	}

	timestampNumeric := pgtype.Numeric{
		Int:   new(big.Int).SetInt64(*timestamp),
		Valid: true,
		Exp:   0,
	}

	// Convert key to big.Int
	keyBig := new(big.Int)
	keyBig.SetString(params.Key, 0)

	// Convert to pgtype.Numeric
	keyNum := pgtype.Numeric{
		Int:   keyBig,
		Valid: true,
	}

	// Check if task already exists
	existingTask, err := h.taskQueries.GetTaskByID(r.Context(), taskID)
	if err == nil {
		// Task exists, return its current status
		response := struct {
			TaskID string `json:"task_id"`
			Status string `json:"status"`
			RequiredConfirmations int32 `json:"required_confirmations,omitempty"`
		}{
			TaskID: existingTask.TaskID,
			Status: existingTask.Status,
			RequiredConfirmations: existingTask.RequiredConfirmations.Int32,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Determine required confirmations
	requiredConfirmations := h.getRequiredConfirmations(params.ChainID)

	// Get latest block number for confirmation check
	latestBlock, err := stateClient.GetLatestBlockNumber(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get latest block: %v", err), http.StatusInternalServerError)
		return
	}

	// Determine initial status based on block confirmations
	status := "pending"
	if params.BlockNumber != nil {
		// Check if task needs confirmations
		if latestBlock < uint64(*params.BlockNumber) + uint64(requiredConfirmations) {
			status = "confirming"
			log.Printf("[API] Task requires %d confirmations, waiting for block %d (current: %d)", 
				requiredConfirmations, uint64(*params.BlockNumber) + uint64(requiredConfirmations), latestBlock)
		} else {
			log.Printf("[API] Task already has sufficient confirmations (target: %d, current: %d)", 
				uint64(*params.BlockNumber) + uint64(requiredConfirmations), latestBlock)
		}
	}

	// Create task
	task, err := h.taskQueries.CreateTask(r.Context(), tasks.CreateTaskParams{
		TaskID:        taskID,
		TargetAddress: params.TargetAddress,
		ChainID:       int32(params.ChainID),
		BlockNumber:   blockNumberNumeric,
		Timestamp:     timestampNumeric,
		Epoch:        int32(currentEpoch),
		Key:          keyNum,
		Status:       status,
		RequiredConfirmations: pgtype.Int4{Int32: int32(requiredConfirmations), Valid: true},
		Value:        pgtype.Numeric{}, // Empty value for now
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[API] Created new task %s with status %s", task.TaskID, task.Status)
	if task.Status == "confirming" {
		log.Printf("[API] Task %s requires %d block confirmations", task.TaskID, requiredConfirmations)
	} else if task.Status == "pending" {
		// If task is pending, process it immediately
		log.Printf("[API] Task %s is pending, processing immediately", task.TaskID)
		
		// Use ProcessPendingTask directly with task pointer
		if err := h.taskProcessor.ProcessPendingTask(r.Context(), &task); err != nil {
			log.Printf("[API] Failed to process pending task %s: %v", task.TaskID, err)
			// Don't return error here, as the task is already created
		}
	}

	// Return response
	response := struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
		RequiredConfirmations int32 `json:"required_confirmations,omitempty"`
	}{
		TaskID: task.TaskID,
		Status: task.Status,
		RequiredConfirmations: task.RequiredConfirmations.Int32,
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
		supportedChains := make([]int64, 0, len(h.config.Chains))
		for chainID := range h.config.Chains {
			supportedChains = append(supportedChains, chainID)
		}
		return fmt.Errorf("unsupported chain ID %d. supported chains: %v", params.ChainID, supportedChains)
	}

	// 2. Validate target address
	if !common.IsHexAddress(params.TargetAddress) {
		return fmt.Errorf("invalid target address: must be valid Ethereum address")
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
	if params.BlockNumber == nil && params.Timestamp == nil {
		return fmt.Errorf("either block number or timestamp must be provided")
	}

	if params.BlockNumber != nil && params.Timestamp != nil {
		return fmt.Errorf("only one of block number or timestamp should be provided")
	}

	// 5. If block number provided, validate it
	if params.BlockNumber != nil {
		// Get state client for target chain
		stateClient, err := h.chainClient.GetStateClient(int64(params.ChainID))
		if err != nil {
			return fmt.Errorf("failed to get state client: %w", err)
		}

		// Get latest block number
		latestBlock, err := stateClient.GetLatestBlockNumber(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get latest block: %w", err)
		}


		// Validate block number range
		blockNum := uint64(*params.BlockNumber)
		if blockNum > latestBlock {
			return fmt.Errorf("block number %d is in the future (latest: %d)", blockNum, latestBlock)
		}
	}

	// 6. If timestamp provided, validate it
	if params.Timestamp != nil {
		// Get current time
		now := time.Now().Unix()

		// Validate timestamp is not in future
		if *params.Timestamp > now {
			return fmt.Errorf("timestamp is in the future")
		}
	}

	return nil
}

func (h *Handler) generateTaskID(params *SendRequestParams, epoch int64, value string) string {
	// Generate task ID using keccak256(abi.encodePacked(targetAddress, chainId, blockNumber, epoch, key, value))
	data := []byte{}
	data = append(data, common.HexToAddress(params.TargetAddress).Bytes()...)
	data = append(data, byte(params.ChainID))
	data = append(data, byte(*params.BlockNumber))
	data = append(data, byte(epoch))
	data = append(data, []byte(params.Key)...)
	data = append(data, []byte(value)...)

	hash := crypto.Keccak256(data)
	return common.Bytes2Hex(hash)
}

// GetTaskConsensus returns the consensus result for a task
func (h *Handler) GetTaskConsensus(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	consensus, err := h.consensusDB.GetConsensusResponse(r.Context(), taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get consensus: %v", err), http.StatusInternalServerError)
		return
	}

	// Get task details
	task, err := h.taskQueries.GetTaskByID(r.Context(), taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get task: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert numeric values
	var value, blockNumber, key string
	if err := task.Value.Scan(&value); err != nil {
		value = "0"
	}
	if err := task.BlockNumber.Scan(&blockNumber); err != nil {
		blockNumber = "0"
	}
	if err := task.Key.Scan(&key); err != nil {
		key = "0"
	}

	// Parse operator signatures to calculate total weight
	var operatorSigs map[string]map[string]interface{}
	totalWeight := big.NewInt(0)
	if err := json.Unmarshal(consensus.OperatorSignatures, &operatorSigs); err == nil {
		for _, sigData := range operatorSigs {
			if weightStr, ok := sigData["weight"].(string); ok {
				weight := new(big.Int)
				if _, ok := weight.SetString(weightStr, 10); ok {
					totalWeight.Add(totalWeight, weight)
				}
			}
		}
	}

	// Format response
	response := struct {
		TaskID              string          `json:"task_id"`
		Epoch              int32           `json:"epoch"`
		Value              string          `json:"value"`
		BlockNumber        string          `json:"block_number"`
		ChainID            int32           `json:"chain_id"`
		TargetAddress      string          `json:"target_address"`
		Key                string          `json:"key"`
		OperatorSignatures json.RawMessage `json:"operator_signatures"`
		TotalWeight        string          `json:"total_weight"`
		ConsensusReachedAt *time.Time     `json:"consensus_reached_at,omitempty"`
	}{
		TaskID:              consensus.TaskID,
		Epoch:              consensus.Epoch,
		Value:              value,
		BlockNumber:        blockNumber,
		ChainID:           task.ChainID,
		TargetAddress:     task.TargetAddress,
		Key:               key,
		OperatorSignatures: consensus.OperatorSignatures,
		TotalWeight:       totalWeight.String(),
		ConsensusReachedAt: &consensus.ConsensusReachedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetTaskFinalResponse returns the final response for a task
func (h *Handler) GetTaskFinalResponse(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	// Get consensus response
	consensus, err := h.consensusDB.GetConsensusResponse(r.Context(), taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get consensus: %v", err), http.StatusInternalServerError)
		return
	}

	// Get task details
	task, err := h.taskQueries.GetTaskByID(r.Context(), taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert numeric values
	var value, key string
	var blockNumber uint64
	if err := task.Value.Scan(&value); err != nil {
		value = "0"
	}
	if err := task.BlockNumber.Scan(&blockNumber); err != nil {
		blockNumber = 0
	}
	if err := task.Key.Scan(&key); err != nil {
		key = "0"
	}

	// Parse operator signatures and calculate total weight
	var operatorSigs map[string]map[string]interface{}
	totalWeight := big.NewInt(0)
	if err := json.Unmarshal(consensus.OperatorSignatures, &operatorSigs); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse operator signatures: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert signatures and calculate total weight
	finalOperatorSigs := make(map[string][]byte)
	for addr, sigData := range operatorSigs {
		// Add weight to total
		if weightStr, ok := sigData["weight"].(string); ok {
			weight := new(big.Int)
			if _, ok := weight.SetString(weightStr, 10); ok {
				totalWeight.Add(totalWeight, weight)
			}
		}
		// Convert signature
		if sigStr, ok := sigData["signature"].(string); ok {
			sig, err := hex.DecodeString(sigStr)
			if err != nil {
				continue
			}
			finalOperatorSigs[addr] = sig
		}
	}

	// Build response
	response := &TaskFinalResponse{
		TaskID:             taskID,
		Epoch:             uint32(consensus.Epoch),
		Value:             value,
		BlockNumber:       blockNumber,
		ChainID:          uint64(task.ChainID),
		TargetAddress:     task.TargetAddress,
		Key:              key,
		OperatorSignatures: finalOperatorSigs,
		TotalWeight:       totalWeight.String(),
		ConsensusReachedAt: consensus.ConsensusReachedAt.Time,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// getRequiredConfirmations returns the required confirmations for a chain
func (h *Handler) getRequiredConfirmations(chainID int64) int32 {
	if chainConfig, ok := h.config.Chains[chainID]; ok {
		return int32(chainConfig.RequiredConfirmations)
	}
	return 12 // Default to 12 confirmations for unknown chains
} 