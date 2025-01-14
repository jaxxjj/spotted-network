package p2p

import (
	"fmt"
	"io"

	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

// ReadJoinRequest reads a JoinRequest from a stream
func ReadJoinRequest(stream network.Stream) (*pb.JoinRequest, error) {
	// Read the request
	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	// Parse the request using protobuf
	var req pb.JoinRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return &req, nil
}

// SendError sends an error response on the stream
func SendError(stream network.Stream, err error) {
	response := &pb.JoinResponse{
		Success: false,
		Error: err.Error(),
	}
	data, _ := proto.Marshal(response)
	stream.Write(data)
}

// SendSuccess sends a success response on the stream
func SendSuccess(stream network.Stream) {
	response := &pb.JoinResponse{
		Success: true,
	}
	data, _ := proto.Marshal(response)
	stream.Write(data)
} 