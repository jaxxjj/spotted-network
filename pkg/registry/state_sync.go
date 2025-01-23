package registry

import (
	"context"
	"log"
	"math/big"
	"sync"

	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

// StateSyncService handles operator state synchronization
type StateSyncService struct {
	host host.Host
	db   *operators.Queries
	
	// Subscribers for state updates
	subscribers   map[peer.ID]network.Stream
	subscribersMu sync.RWMutex
}

// NewStateSyncService creates a new state sync service
func NewStateSyncService(h host.Host, db *operators.Queries) *StateSyncService {
	return &StateSyncService{
		host:        h,
		db:          db,
		subscribers: make(map[peer.ID]network.Stream),
	}
}

// Start initializes the state sync service
func (s *StateSyncService) Start(ctx context.Context) error {
	// Set up protocol handler for state sync requests
	s.host.SetStreamHandler("/state-sync/1.0.0", s.handleStateSync)
	
	log.Printf("[StateSync] Service started successfully")
	return nil
}

// handleStateSync handles incoming state sync requests
func (s *StateSyncService) handleStateSync(stream network.Stream) {
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("[StateSync] New state sync request from %s", remotePeer)

	// Read message type
	msgType := make([]byte, 1)
	if _, err := stream.Read(msgType); err != nil {
		log.Printf("[StateSync] Error reading message type from %s: %v", remotePeer, err)
		stream.Reset()
		return
	}

	switch msgType[0] {
	case 0x01: // GetFullStateRequest
		s.handleGetFullState(stream)
	case 0x02: // SubscribeRequest
		s.Subscribe(remotePeer, stream)
		log.Printf("[StateSync] Subscribed %s to state updates", remotePeer)
	default:
		log.Printf("[StateSync] Unknown message type from %s: %x", remotePeer, msgType[0])
		stream.Reset()
	}
}

// handleGetFullState handles a request for full state
func (s *StateSyncService) handleGetFullState(stream network.Stream) {
	// Get all operators
	operators, err := s.db.ListOperatorsByStatus(context.Background(), "active")
	if err != nil {
		log.Printf("[StateSync] Failed to get operators: %v", err)
		stream.Reset()
		return
	}

	// Create response
	resp := &pb.GetFullStateResponse{
		Operators: make([]*pb.OperatorState, len(operators)),
	}
	for i, op := range operators {
		resp.Operators[i] = convertToProtoOperator(op)
	}

	// Marshal response
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("[StateSync] Failed to marshal response: %v", err)
		stream.Reset()
		return
	}

	// Write length prefix
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)

	if _, err := stream.Write(lengthBytes); err != nil {
		log.Printf("[StateSync] Failed to write length prefix: %v", err)
		stream.Reset()
		return
	}

	// Write response
	if _, err := stream.Write(data); err != nil {
		log.Printf("[StateSync] Failed to write response: %v", err)
		stream.Reset()
		return
	}

	log.Printf("[StateSync] Sent full state with %d operators", len(operators))
}

// Subscribe adds a new subscriber
func (s *StateSyncService) Subscribe(peerID peer.ID, stream network.Stream) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()

	// Close and delete existing stream if any
	if existingStream, ok := s.subscribers[peerID]; ok {
		existingStream.Reset()
		delete(s.subscribers, peerID)
	}

	s.subscribers[peerID] = stream
	log.Printf("[StateSync] New subscriber added: %s", peerID)
}

// Unsubscribe removes a subscriber
func (s *StateSyncService) Unsubscribe(peerID peer.ID) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()

	if stream, ok := s.subscribers[peerID]; ok {
		stream.Reset()
		delete(s.subscribers, peerID)
		log.Printf("[StateSync] Subscriber removed: %s", peerID)
	}
}

// BroadcastUpdate sends an update to all subscribed operators
func (s *StateSyncService) BroadcastUpdate(operators []*pb.OperatorState) {
	s.subscribersMu.RLock()
	defer s.subscribersMu.RUnlock()

	if len(operators) == 0 {
		log.Printf("[StateSync] No operators to broadcast")
		return
	}

	// Create update message
	update := &pb.OperatorStateUpdate{
		Operators: operators,
	}

	// Marshal update
	data, err := proto.Marshal(update)
	if err != nil {
		log.Printf("[StateSync] Failed to marshal state update: %v", err)
		return
	}

	// Write length prefix
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)

	// Broadcast to all subscribers
	for peer, stream := range s.subscribers {
		// Write length prefix
		if _, err := stream.Write(lengthBytes); err != nil {
			log.Printf("[StateSync] Error sending length prefix to %s: %v", peer, err)
			stream.Reset()
			delete(s.subscribers, peer)
			continue
		}

		// Write update data
		if _, err := stream.Write(data); err != nil {
			log.Printf("[StateSync] Error sending update to %s: %v", peer, err)
			stream.Reset()
			delete(s.subscribers, peer)
			continue
		}

		log.Printf("[StateSync] Sent state update to %s with %d operators", peer, len(operators))
	}
}

// Helper functions

func convertToProtoOperator(op operators.Operators) *pb.OperatorState {
	var exitEpoch *int32
	defaultExitEpoch := pgtype.Numeric{
		Int:    new(big.Int).SetUint64(4294967295),
		Exp:    0,
		Valid:  true,
	}
	if op.ExitEpoch.Int.Cmp(defaultExitEpoch.Int) != 0 { // Check if not default max value
		val := int32(op.ExitEpoch.Int.Int64())
		exitEpoch = &val
	}

	return &pb.OperatorState{
		Address:                op.Address,
		SigningKey:            op.SigningKey,
		RegisteredAtBlockNumber: op.RegisteredAtBlockNumber.Int.Int64(),
		RegisteredAtTimestamp:   op.RegisteredAtTimestamp.Int.Int64(),
		ActiveEpoch:           int32(op.ActiveEpoch.Int.Int64()),
		ExitEpoch:            exitEpoch,
		Status:               op.Status,
		Weight:               op.Weight.Int.String(),
	}
}

func readProtoMessage(stream network.Stream, msg proto.Message) error {
	buf := make([]byte, 1024*1024) // 1MB buffer
	n, err := stream.Read(buf)
	if err != nil {
		return err
	}
	return proto.Unmarshal(buf[:n], msg)
}

func writeProtoMessage(stream network.Stream, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = stream.Write(data)
	return err
}
