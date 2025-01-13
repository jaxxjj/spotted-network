package p2p

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/network"
)

// JoinRequest represents a request from an operator to join the network
type JoinRequest struct {
	OperatorAddress string `json:"operator_address"`
	SigningKey      string `json:"signing_key"`
	Message         []byte `json:"message"`
	Signature       []byte `json:"signature"`
}

// ReadJoinRequest reads a JoinRequest from a stream
func ReadJoinRequest(stream network.Stream) (*JoinRequest, error) {
	// Read the request
	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	// Parse the request
	var req JoinRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return &req, nil
}

// SendError sends an error response on the stream
func SendError(stream network.Stream, err error) {
	response := map[string]string{
		"status": "error",
		"error":  err.Error(),
	}
	data, _ := json.Marshal(response)
	stream.Write(data)
}

// SendSuccess sends a success response on the stream
func SendSuccess(stream network.Stream) {
	response := map[string]string{
		"status": "success",
	}
	data, _ := json.Marshal(response)
	stream.Write(data)
} 