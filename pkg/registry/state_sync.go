package registry

import (
	"context"
	"sync"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
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
	subscribers   map[peer.ID]chan *pb.OperatorStateUpdate
	subscribersMu sync.RWMutex
}

func NewStateSyncService(h host.Host, db *operators.Queries) *StateSyncService {
	return &StateSyncService{
		host:        h,
		db:          db,
		subscribers: make(map[peer.ID]chan *pb.OperatorStateUpdate),
	}
}

// Start initializes the state sync service
func (s *StateSyncService) Start(ctx context.Context) error {
	// Set up protocol handler for state sync requests
	s.host.SetStreamHandler("/state-sync/1.0.0", s.handleStateSync)
	
	// Start broadcasting state updates
	go s.broadcastStateUpdates(ctx)
	
	return nil
}

// handleStateSync handles incoming state sync requests
func (s *StateSyncService) handleStateSync(stream network.Stream) {
	defer stream.Close()
	
	// Read request
	var req pb.GetFullStateRequest
	if err := readProtoMessage(stream, &req); err != nil {
		return
	}
	
	// Get all operators from database
	operators, err := s.db.ListOperatorsByStatus(context.Background(), "active")
	if err != nil {
		return
	}
	
	// Convert to proto message
	resp := &pb.GetFullStateResponse{
		Operators: make([]*pb.OperatorState, len(operators)),
	}
	for i, op := range operators {
		resp.Operators[i] = convertToProtoOperator(op)
	}
	
	// Send response
	if err := writeProtoMessage(stream, resp); err != nil {
		return
	}
}

// Subscribe adds a new subscriber for state updates
func (s *StateSyncService) Subscribe(peerID peer.ID) chan *pb.OperatorStateUpdate {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()
	
	ch := make(chan *pb.OperatorStateUpdate, 100)
	s.subscribers[peerID] = ch
	return ch
}

// Unsubscribe removes a subscriber
func (s *StateSyncService) Unsubscribe(peerID peer.ID) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()
	
	if ch, ok := s.subscribers[peerID]; ok {
		close(ch)
		delete(s.subscribers, peerID)
	}
}

// BroadcastUpdate sends an update to all subscribers
func (s *StateSyncService) BroadcastUpdate(update *pb.OperatorStateUpdate) {
	s.subscribersMu.RLock()
	defer s.subscribersMu.RUnlock()
	
	for _, ch := range s.subscribers {
		select {
		case ch <- update:
		default:
			// Skip if channel is full
		}
	}
}

// broadcastStateUpdates periodically broadcasts full state
func (s *StateSyncService) broadcastStateUpdates(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get all operators
			operators, err := s.db.ListOperatorsByStatus(ctx, "active")
			if err != nil {
				continue
			}
			
			// Create update message
			update := &pb.OperatorStateUpdate{
				Type:      "FULL",
				Operators: make([]*pb.OperatorState, len(operators)),
			}
			for i, op := range operators {
				update.Operators[i] = convertToProtoOperator(op)
			}
			
			// Broadcast to all subscribers
			s.BroadcastUpdate(update)
		}
	}
}

// Helper functions

func convertToProtoOperator(op operators.Operators) *pb.OperatorState {
	var exitEpoch *int32
	if op.ExitEpoch != 4294967295 { // Check if not default max value
		val := int32(op.ExitEpoch)
		exitEpoch = &val
	}

	return &pb.OperatorState{
		Address:                op.Address,
		SigningKey:            op.SigningKey,
		RegisteredAtBlockNumber: int64(op.RegisteredAtBlockNumber),
		RegisteredAtTimestamp:   int64(op.RegisteredAtTimestamp),
		ActiveEpoch:           int32(op.ActiveEpoch),
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
