package p2p

import (
	"encoding/binary"
	"fmt"
	"io"

	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

const (
	// Protocol message types
	MsgTypeJoinRequest byte = 0x01
	MsgTypeJoinResponse byte = 0x02
)

// WriteLengthPrefixed writes a message with type and length prefix
func WriteLengthPrefixed(stream network.Stream, msgType byte, data []byte) error {
	// Write message type (1 byte)
	if _, err := stream.Write([]byte{msgType}); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Write length prefix (4 bytes, big endian)
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))
	if _, err := stream.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}

	// Write data
	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// ReadLengthPrefixed reads a message with type and length prefix
func ReadLengthPrefixed(stream network.Stream) (byte, []byte, error) {
	// Read message type
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		return 0, nil, fmt.Errorf("failed to read message type: %w", err)
	}

	// Read length prefix
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(stream, lengthBuf); err != nil {
		return 0, nil, fmt.Errorf("failed to read length prefix: %w", err)
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// Read data
	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		return 0, nil, fmt.Errorf("failed to read data: %w", err)
	}

	return msgType[0], data, nil
}

// ReadJoinRequest reads a JoinRequest from a stream
func ReadJoinRequest(stream network.Stream) (*pb.JoinRequest, error) {
	// Read message with length prefix
	msgType, data, err := ReadLengthPrefixed(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	// Verify message type
	if msgType != MsgTypeJoinRequest {
		return nil, fmt.Errorf("unexpected message type: %d", msgType)
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
	WriteLengthPrefixed(stream, MsgTypeJoinResponse, data)
}

// SendSuccess sends a success response on the stream
func SendSuccess(stream network.Stream, activeOperators []*pb.ActiveOperator) {
	response := &pb.JoinResponse{
		Success: true,
		ActiveOperators: activeOperators,
	}
	data, _ := proto.Marshal(response)
	WriteLengthPrefixed(stream, MsgTypeJoinResponse, data)
} 