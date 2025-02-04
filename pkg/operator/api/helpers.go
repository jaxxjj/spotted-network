// pkg/operator/api/response.go

package api

import (
	"encoding/json"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func writeJSON(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, err string) {
	response := SendRequestResponse{
		Status: "error",
		Error:  err,
	}
	writeJSON(w, status, response)
}

func writeSuccess(w http.ResponseWriter, taskID string, status string, confirmations uint16, message string) {
	response := SendRequestResponse{
		TaskID:               taskID,
		Status:               status,
		RequiredConfirmations: confirmations,
		Message:              message,
	}
	writeJSON(w, http.StatusOK, response)
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

// getRequiredConfirmations returns the required confirmations for a chain
func (h *Handler) getRequiredConfirmations(chainID uint32) uint16 {
	if chainConfig, ok := h.config.Chains[chainID]; ok {
		return chainConfig.RequiredConfirmations
	}
	return 12 // Default to 12 confirmations for unknown chains
}

