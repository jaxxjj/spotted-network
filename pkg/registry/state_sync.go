package registry

import (
	"context"
	"log"
	"math/big"
	"sync"
	"time"

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
	subscribers   map[peer.ID]chan *pb.OperatorStateUpdate
	subscribersMu sync.RWMutex
}

func NewStateSyncService(h host.Host, db *operators.Queries) *StateSyncService {
	log.Printf("[StateSync] Initializing state sync service")
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
	
	log.Printf("[StateSync] Service started successfully")
	return nil
}

// handleStateSync handles incoming state sync requests
func (s *StateSyncService) handleStateSync(stream network.Stream) {
	defer stream.Close()
	
	peer := stream.Conn().RemotePeer()
	log.Printf("[StateSync] Received state sync request from peer %s", peer.String())
	
	// Read request
	var req pb.GetFullStateRequest
	if err := readProtoMessage(stream, &req); err != nil {
		log.Printf("[StateSync] Failed to read request from peer %s: %v", peer.String(), err)
		return
	}
	
	// Get all operators from database
	operators, err := s.db.ListOperatorsByStatus(context.Background(), "active")
	if err != nil {
		log.Printf("[StateSync] Failed to get operators from database: %v", err)
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
		log.Printf("[StateSync] Failed to send response to peer %s: %v", peer.String(), err)
		return
	}
	
	log.Printf("[StateSync] Successfully sent state response to peer %s with %d operators", peer.String(), len(operators))
}

// Subscribe adds a new subscriber for state updates
func (s *StateSyncService) Subscribe(peerID peer.ID) chan *pb.OperatorStateUpdate {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()
	
	ch := make(chan *pb.OperatorStateUpdate, 100)
	s.subscribers[peerID] = ch
	log.Printf("[StateSync] Peer %s subscribed to state updates. Total subscribers: %d", peerID.String(), len(s.subscribers))
	return ch
}

// Unsubscribe removes a subscriber
func (s *StateSyncService) Unsubscribe(peerID peer.ID) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()
	
	if ch, ok := s.subscribers[peerID]; ok {
		close(ch)
		delete(s.subscribers, peerID)
		log.Printf("[StateSync] Peer %s unsubscribed from state updates. Remaining subscribers: %d", peerID.String(), len(s.subscribers))
	}
}

// BroadcastUpdate sends an update to all subscribers
func (s *StateSyncService) BroadcastUpdate(update *pb.OperatorStateUpdate) {
	s.subscribersMu.RLock()
	defer s.subscribersMu.RUnlock()
	
	successCount := 0
	for peerID, ch := range s.subscribers {
		select {
		case ch <- update:
			successCount++
		default:
			log.Printf("[StateSync] Failed to send update to peer %s: channel full", peerID.String())
		}
	}
	
	log.Printf("[StateSync] Broadcasted state update type %s to %d/%d subscribers", update.Type, successCount, len(s.subscribers))
}

// broadcastStateUpdates periodically broadcasts full state
func (s *StateSyncService) broadcastStateUpdates(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	log.Printf("[StateSync] Starting periodic state broadcast (interval: 5m)")
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("[StateSync] Stopping periodic state broadcast")
			return
		case <-ticker.C:
			// Get all operators
			operators, err := s.db.ListOperatorsByStatus(ctx, "active")
			if err != nil {
				log.Printf("[StateSync] Failed to get operators for periodic broadcast: %v", err)
				continue
			}
			
			log.Printf("[StateSync] Preparing periodic state broadcast with %d operators", len(operators))
			
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
