package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgtype"
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
		healthCheckInterval: 5 * time.Second,
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

	n.subscribersMu.RLock()
	defer n.subscribersMu.RUnlock()

	for peer, stream := range n.subscribers {
		if _, err := stream.Write(data); err != nil {
			log.Printf("Error sending update to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
		} else {
			log.Printf("Sent state update to %s - Type: %s, Operators: %d", 
				peer, updateType, len(operators))
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

func (n *Node) calculateEpochNumber(blockNumber uint64) uint32 {
	return uint32((blockNumber - GenesisBlock) / EpochPeriod)
}

func (n *Node) updateOperatorStates(ctx context.Context, currentEpoch uint32) error {
	// Get all operators in waitingActive state
	waitingActiveOps, err := n.db.ListOperatorsByStatus(ctx, "waitingActive")
	if err != nil {
		return fmt.Errorf("failed to get waiting active operators: %w", err)
	}

	// Process waiting active operators
	for _, op := range waitingActiveOps {
		currentEpochNumeric := pgtype.Numeric{
			Int:    new(big.Int).SetInt64(int64(currentEpoch)),
			Exp:    0,
			Valid:  true,
		}
		if op.ActiveEpoch.Int.Cmp(currentEpochNumeric.Int) == 0 {
			// Get operator weight from contract
			mainnetClient := n.chainClients.GetMainnetClient()
			weight, err := mainnetClient.GetOperatorWeight(ctx, common.HexToAddress(op.Address))
			if err != nil {
				log.Printf("[Registry] Failed to get weight for operator %s: %v", op.Address, err)
				continue
			}

			// Update operator state to active
			_, err = n.db.UpdateOperatorStatus(ctx, operators.UpdateOperatorStatusParams{
				Address: op.Address,
				Status: "active",
			})
			if err != nil {
				log.Printf("[Registry] Failed to update operator %s state: %v", op.Address, err)
				continue
			}

			log.Printf("[Registry] Activated operator %s with weight %s", op.Address, weight.String())
		}
	}

	// Get all operators in waitingExit state
	waitingExitOps, err := n.db.ListOperatorsByStatus(ctx, "waitingExit")
	if err != nil {
		return fmt.Errorf("failed to get waiting exit operators: %w", err)
	}

	// Process waiting exit operators
	for _, op := range waitingExitOps {
		currentEpochNumeric := pgtype.Numeric{
			Int:    new(big.Int).SetInt64(int64(currentEpoch)),
			Exp:    0,
			Valid:  true,
		}
		if op.ExitEpoch.Int.Cmp(currentEpochNumeric.Int) == 0 {
			// Update operator state to inactive
			_, err = n.db.UpdateOperatorStatus(ctx, operators.UpdateOperatorStatusParams{
				Address: op.Address,
				Status: "inactive",
			})
			if err != nil {
				log.Printf("[Registry] Failed to update operator %s state: %v", op.Address, err)
				continue
			}

			log.Printf("[Registry] Deactivated operator %s", op.Address)
		}
	}

	return nil
}

func (n *Node) monitorEpochUpdates(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastProcessedEpoch uint32

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
					log.Printf("[Registry] Failed to update operator states: %v", err)
					continue
				}
				lastProcessedEpoch = currentEpoch
			}
		}
	}
}