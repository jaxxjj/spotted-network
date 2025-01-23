package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

type Node struct {
	host *p2p.Host
	
	// Connected operators
	operators map[peer.ID]*OperatorInfo
	operatorsMu sync.RWMutex
	
	// Health check interval
	healthCheckInterval time.Duration

	// Database connection
	db *operators.Queries

	// Event listener
	eventListener *EventListener

	// Chain clients manager
	chainClients *ethereum.ChainClients

	// State sync subscribers
	subscribers map[peer.ID]network.Stream
	subscribersMu sync.RWMutex
}

func NewNode(ctx context.Context, cfg *p2p.Config, db *operators.Queries, chainClients *ethereum.ChainClients) (*Node, error) {
	host, err := p2p.NewHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	node := &Node{
		host:               host,
		operators:          make(map[peer.ID]*OperatorInfo),
		healthCheckInterval: 100 * time.Second,
		db:                db,
		chainClients:      chainClients,
		subscribers:       make(map[peer.ID]network.Stream),
	}

	// Create and initialize event listener
	node.eventListener = NewEventListener(chainClients, db)

	// Start listening for chain events
	if err := node.eventListener.StartListening(ctx); err != nil {
		return nil, fmt.Errorf("failed to start event listener: %w", err)
	}

	// Set stream handler for operator connections
	node.SetupProtocols()

	// Start health check
	go node.startHealthCheck(ctx)

	log.Printf("Registry Node started. %s\n", host.GetHostInfo())
	return node, nil
}

func (n *Node) Stop() error {
	return n.host.Close()
}

// SetupProtocols sets up the protocol handlers for the registry node
func (n *Node) SetupProtocols() {
	// Set up join request handler
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleStream)
	
	// Set up state sync handler
	n.host.SetStreamHandler("/state-sync/1.0.0", n.handleStateSync)
	
	log.Printf("Protocol handlers set up")
}

// BroadcastStateUpdate sends a state update to all subscribers
func (n *Node) BroadcastStateUpdate(operators []*pb.OperatorState, updateType string) {
	update := &pb.OperatorStateUpdate{
		Type:      updateType,
		Operators: operators,
	}

	data, err := proto.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling state update: %v", err)
		return
	}

	// Prepare message type and length prefix
	msgType := []byte{0x02} // 0x02 for state update
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)

	n.subscribersMu.RLock()
	defer n.subscribersMu.RUnlock()

	for peer, stream := range n.subscribers {
		// Write message type
		if _, err := stream.Write(msgType); err != nil {
			log.Printf("Error sending message type to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
			continue
		}

		// Write length prefix
		if _, err := stream.Write(lengthBytes); err != nil {
			log.Printf("Error sending length prefix to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
			continue
		}

		// Write protobuf data
		if _, err := stream.Write(data); err != nil {
			log.Printf("Error sending update data to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
		} else {
			log.Printf("Sent state update to %s - Type: %s, Length: %d, Operators: %d", 
				peer, updateType, length, len(operators))
		}
	}
}

// GetOperatorInfo returns information about a connected operator
func (n *Node) GetOperatorInfo(id peer.ID) *OperatorInfo {
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()
	return n.operators[id]
}

func (n *Node) GetConnectedOperators() []peer.ID {
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()

	operators := make([]peer.ID, 0, len(n.operators))
	for id := range n.operators {
		operators = append(operators, id)
	}
	return operators
}

// GetHostID returns the node's libp2p host ID
func (n *Node) GetHostID() string {
	return n.host.ID().String()
}

func (n *Node) Start(ctx context.Context) error {
	// Start epoch monitoring
	go n.monitorEpochUpdates(ctx)

	return nil
}

func (n *Node) calculateEpochNumber(blockNumber uint64) uint32 {
	return uint32((blockNumber - GenesisBlock) / EpochPeriod)
}

func (n *Node) monitorEpochUpdates(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastProcessedEpoch uint32

	log.Printf("[Registry] Starting epoch monitoring...")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get latest block number
			mainnetClient := n.chainClients.GetMainnetClient()
			blockNumber, err := mainnetClient.GetLatestBlockNumber(ctx)
			if err != nil {
				log.Printf("[Registry] Failed to get latest block number: %v", err)
				continue
			}

			// Calculate current epoch
			currentEpoch := n.calculateEpochNumber(blockNumber)

			// If we're in a new epoch, update operator states
			if currentEpoch > lastProcessedEpoch {
				if err := n.updateOperatorStates(ctx, currentEpoch); err != nil {
					log.Printf("[Registry] Failed to update operator states for epoch %d: %v", currentEpoch, err)
					continue
				}
				lastProcessedEpoch = currentEpoch
				log.Printf("[Registry] Updated operator states for epoch %d", currentEpoch)
			}
		}
	}
}

// GetOperatorByAddress gets operator info from database
func (n *Node) GetOperatorByAddress(ctx context.Context, address string) (operators.Operators, error) {
	return n.db.GetOperatorByAddress(ctx, address)
}

// UpdateOperatorStatus updates operator status in database
func (n *Node) UpdateOperatorStatus(ctx context.Context, address string, status string) error {
	_, err := n.db.UpdateOperatorStatus(ctx, operators.UpdateOperatorStatusParams{
		Address: address,
		Status:  status,
	})
	return err
}

func (n *Node) updateOperatorStates(ctx context.Context, currentEpoch uint32) error {
	updatedOperators := make([]*pb.OperatorState, 0)

	// Get all operators in waitingActive state
	waitingActiveOps, err := n.db.ListOperatorsByStatus(ctx, "waitingActive")
	if err != nil {
		return fmt.Errorf("failed to get waiting active operators: %w", err)
	}

	log.Printf("[Registry] Processing %d waiting active operators for epoch %d", len(waitingActiveOps), currentEpoch)

	// Process waiting active operators
	for _, op := range waitingActiveOps {
		currentEpochNumeric := pgtype.Numeric{
			Int:    new(big.Int).SetInt64(int64(currentEpoch)),
			Exp:    0,
			Valid:  true,
		}
		if op.ActiveEpoch.Int.Cmp(currentEpochNumeric.Int) == 0 {
			// Update operator state to active
			updatedOp, err := n.db.UpdateOperatorStatus(ctx, operators.UpdateOperatorStatusParams{
				Address: op.Address,
				Status: "active",
			})
			if err != nil {
				log.Printf("[Registry] Failed to update operator %s state: %v", op.Address, err)
				continue
			}

			// Add to updated operators list without weight for now
			updatedOperators = append(updatedOperators, &pb.OperatorState{
				Address:                 updatedOp.Address,
				SigningKey:             updatedOp.SigningKey,
				RegisteredAtBlockNumber: updatedOp.RegisteredAtBlockNumber.Int.Int64(),
				RegisteredAtTimestamp:   updatedOp.RegisteredAtTimestamp.Int.Int64(),
				ActiveEpoch:            int32(updatedOp.ActiveEpoch.Int.Int64()),
				ExitEpoch:              nil,
				Status:                 updatedOp.Status,
			})

			log.Printf("[Registry] Activated operator %s at epoch %d", op.Address, currentEpoch)
		}
	}

	// Get all operators in waitingExit state
	waitingExitOps, err := n.db.ListOperatorsByStatus(ctx, "waitingExit")
	if err != nil {
		return fmt.Errorf("failed to get waiting exit operators: %w", err)
	}

	log.Printf("[Registry] Processing %d waiting exit operators for epoch %d", len(waitingExitOps), currentEpoch)

	// Process waiting exit operators
	for _, op := range waitingExitOps {
		currentEpochNumeric := pgtype.Numeric{
			Int:    new(big.Int).SetInt64(int64(currentEpoch)),
			Exp:    0,
			Valid:  true,
		}
		if op.ExitEpoch.Int.Cmp(currentEpochNumeric.Int) == 0 {
			// Update operator state to inactive
			updatedOp, err := n.db.UpdateOperatorStatus(ctx, operators.UpdateOperatorStatusParams{
				Address: op.Address,
				Status: "inactive",
			})
			if err != nil {
				log.Printf("[Registry] Failed to update operator %s state: %v", op.Address, err)
				continue
			}

			// Add to updated operators list
			exitEpoch := int32(updatedOp.ExitEpoch.Int.Int64())
			updatedOperators = append(updatedOperators, &pb.OperatorState{
				Address:                 updatedOp.Address,
				SigningKey:             updatedOp.SigningKey,
				RegisteredAtBlockNumber: updatedOp.RegisteredAtBlockNumber.Int.Int64(),
				RegisteredAtTimestamp:   updatedOp.RegisteredAtTimestamp.Int.Int64(),
				ActiveEpoch:            int32(updatedOp.ActiveEpoch.Int.Int64()),
				ExitEpoch:              &exitEpoch,
				Status:                 updatedOp.Status,
				Weight:                 common.NumericToString(updatedOp.Weight),
			})

			log.Printf("[Registry] Deactivated operator %s at epoch %d", op.Address, currentEpoch)
		}
	}

	// Get all active operators and update their weights
	activeOps, err := n.db.ListOperatorsByStatus(ctx, "active")
	if err != nil {
		return fmt.Errorf("failed to get active operators: %w", err)
	}

	log.Printf("[Registry] Updating weights for %d active operators", len(activeOps))

	mainnetClient := n.chainClients.GetMainnetClient()
	for _, op := range activeOps {
		weight, err := mainnetClient.GetOperatorWeight(ctx, ethcommon.HexToAddress(op.Address))
		if err != nil {
			log.Printf("[Registry] Failed to get weight for operator %s: %v", op.Address, err)
			continue
		}

		// Add to updated operators list with weight
		updatedOperators = append(updatedOperators, &pb.OperatorState{
			Address:                 op.Address,
			SigningKey:             op.SigningKey,
			RegisteredAtBlockNumber: op.RegisteredAtBlockNumber.Int.Int64(),
			RegisteredAtTimestamp:   op.RegisteredAtTimestamp.Int.Int64(),
			ActiveEpoch:            int32(op.ActiveEpoch.Int.Int64()),
			ExitEpoch:              nil,
			Status:                 op.Status,
			Weight:                 weight.String(),
		})

		log.Printf("[Registry] Updated weight for operator %s: %s", op.Address, weight.String())
	}

	// Broadcast updates if any operators were updated
	if len(updatedOperators) > 0 {
		n.BroadcastStateUpdate(updatedOperators, "state_update")
		log.Printf("[Registry] Broadcast state update for %d operators at epoch %d", len(updatedOperators), currentEpoch)
	} else {
		log.Printf("[Registry] No operator state changes needed for epoch %d", currentEpoch)
	}

	return nil
}