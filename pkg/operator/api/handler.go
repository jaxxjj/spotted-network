package api

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	taskQueries *tasks.Queries
	chainClient *ethereum.ChainClients
}

func NewHandler(taskQueries *tasks.Queries, chainClient *ethereum.ChainClients) *Handler {
	return &Handler{
		taskQueries: taskQueries,
		chainClient: chainClient,
	}
}

type SendRequestParams struct {
	ChainID       int64  `json:"chain_id"`
	TargetAddress string `json:"target_address"`
	Key           string `json:"key"`
	BlockNumber   *int64 `json:"block_number,omitempty"`
	Timestamp     *int64 `json:"timestamp,omitempty"`
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

	// Get state client for target chain
	stateClient, err := h.chainClient.GetStateClient(fmt.Sprintf("%d", params.ChainID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get state client: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert key to big.Int
	keyBig := new(big.Int)
	keyBig.SetString(params.Key, 0)

	// Query state based on block number or timestamp
	var value *big.Int
	if params.BlockNumber != nil {
		value, err = stateClient.GetStateAtBlock(r.Context(), common.HexToAddress(params.TargetAddress), keyBig, uint64(*params.BlockNumber))
	} else {
		value, err = stateClient.GetStateAtTimestamp(r.Context(), common.HexToAddress(params.TargetAddress), keyBig, uint64(*params.Timestamp))
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query state: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate task ID
	taskID := h.generateTaskID(&params, int64(currentEpoch), value.String())

	// Convert numeric values
	var blockNumber, timestamp pgtype.Numeric
	if params.BlockNumber != nil {
		if err := blockNumber.Scan(fmt.Sprintf("%d", *params.BlockNumber)); err != nil {
			http.Error(w, "Invalid block number", http.StatusBadRequest)
			return
		}
	}
	if params.Timestamp != nil {
		if err := timestamp.Scan(fmt.Sprintf("%d", *params.Timestamp)); err != nil {
			http.Error(w, "Invalid timestamp", http.StatusBadRequest)
			return
		}
	}

	// Convert key and value to Numeric
	var keyNum, valueNum pgtype.Numeric
	if err := keyNum.Scan(keyBig.String()); err != nil {
		http.Error(w, "Invalid key value", http.StatusInternalServerError)
		return
	}
	if err := valueNum.Scan(value.String()); err != nil {
		http.Error(w, "Invalid state value", http.StatusInternalServerError)
		return
	}

	// Convert expiration time
	var expireTime pgtype.Timestamp
	if err := expireTime.Scan(time.Now().Add(15 * time.Second)); err != nil {
		http.Error(w, "Failed to set expiration time", http.StatusInternalServerError)
		return
	}

	// Create task
	task, err := h.taskQueries.CreateTask(r.Context(), tasks.CreateTaskParams{
		TaskID:        taskID,
		TargetAddress: params.TargetAddress,
		ChainID:       int32(params.ChainID),
		BlockNumber:   blockNumber,
		Timestamp:     timestamp,
		Epoch:        int32(currentEpoch),
		Key:          keyNum,
		Value:        valueNum,
		ExpireTime:   expireTime,
		Status:       "pending",
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SendRequestResponse{
		TaskID: task.TaskID,
	})
}

func (h *Handler) validateRequest(params *SendRequestParams) error {
	if params.ChainID <= 0 {
		return fmt.Errorf("invalid chain ID")
	}

	if !common.IsHexAddress(params.TargetAddress) {
		return fmt.Errorf("invalid target address")
	}

	if params.BlockNumber == nil && params.Timestamp == nil {
		return fmt.Errorf("either block number or timestamp must be provided")
	}

	if params.BlockNumber != nil && params.Timestamp != nil {
		return fmt.Errorf("only one of block number or timestamp should be provided")
	}

	return nil
}

func (h *Handler) generateTaskID(params *SendRequestParams, epoch int64, value string) string {
	// Generate task ID using keccak256(abi.encodePacked(targetAddress, chainId, blockNumber, timestamp, epoch, key, value))
	data := []byte{}
	data = append(data, common.HexToAddress(params.TargetAddress).Bytes()...)
	data = append(data, byte(params.ChainID))
	if params.BlockNumber != nil {
		data = append(data, byte(*params.BlockNumber))
	}
	if params.Timestamp != nil {
		data = append(data, byte(*params.Timestamp))
	}
	data = append(data, byte(epoch))
	data = append(data, []byte(params.Key)...)
	data = append(data, []byte(value)...)

	hash := crypto.Keccak256(data)
	return common.Bytes2Hex(hash)
} 