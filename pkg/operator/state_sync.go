package operator

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

const (
	GenesisBlock = 0
	EpochPeriod  = 12
	epochMonitorInterval = 5 * time.Second

	MsgTypeGetFullState    byte = 0x01
	MsgTypeSubscribe       byte = 0x02
	MsgTypeStateUpdate     byte = 0x03
	MsgTypePeerSync        byte = 0x04  // New message type for peer sync
)

type ChainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	GetMinimumWeight(ctx context.Context) (*big.Int, error)
	GetThresholdWeight(ctx context.Context) (*big.Int, error)
	GetTotalWeight(ctx context.Context) (*big.Int, error)
	GetStateAtBlock(ctx context.Context, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error)
	GetCurrentEpoch(ctx context.Context) (uint32, error)
	Close() error
}	

func (node *Node) subscribeToStateUpdates() error {
	log.Printf("[StateSync] Opening state sync stream to registry...")
	stream, err := node.host.NewStream(context.Background(), node.registryID, "/state-sync/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open state sync stream: %w", err)
	}

	// Send message type
	msgType := []byte{MsgTypeSubscribe}
	if _, err := stream.Write(msgType); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Send subscribe request
	req := &pb.SubscribeRequest{}
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal subscribe request: %w", err)
	}

	// Send length prefix
	length := uint32(len(data))
	if err := commonHelpers.WriteLengthPrefix(stream, length); err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}

	log.Printf("[StateSync] Sending state sync subscribe request (type: 0x02, length: %d)...", length)
	bytesWritten, err := stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send subscribe request: %w", err)
	}
	log.Printf("[StateSync] Sent %d bytes to registry", bytesWritten)
	log.Printf("[StateSync] Successfully subscribed to state updates")
	log.Printf("[StateSync] Starting state update handler...")

	// Handle updates in background
	go node.handleStateUpdates(stream)
	return nil
}

func (node *Node) getFullState() error {
	log.Printf("[StateSync] Requesting full operator state from registry...")
	stream, err := node.host.NewStream(context.Background(), node.registryID, "/state-sync/1.0.0")
	if err != nil {
		return fmt.Errorf("[StateSync] failed to open state sync stream: %w", err)
	}
	defer stream.Close()

	// Send get full state request with message type
	msgType := []byte{MsgTypeGetFullState}
	if _, err := stream.Write(msgType); err != nil {
		return fmt.Errorf("[StateSync] failed to write message type: %w", err)
	}

	// Send get full state request
	req := &pb.GetFullStateRequest{}
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("[StateSync] failed to marshal get state request: %w", err)
	}

	// Send length-prefixed data
	if err := commonHelpers.WriteLengthPrefixedData(stream, data); err != nil {
		return fmt.Errorf("[StateSync] failed to write request: %w", err)
	}

	// Read response with deadline
	respData, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, 5*time.Second)
	if err != nil {
		return fmt.Errorf("[StateSync] failed to read response: %w", err)
	}

	var resp pb.GetFullStateResponse
	if err := proto.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("[StateSync] failed to unmarshal state response: %w", err)
	}

	log.Printf("Successfully unmarshaled response with %d operators", len(resp.Operators))
	
	// Update local state
	log.Printf("Updating local operator states...")
	node.updateOperatorStates(resp.Operators)
	
	// Print initial state table with header
	log.Printf("\n=== Initial Operator State Table ===")
	if len(resp.Operators) == 0 {
		log.Printf("No operators found in response")
	} else {
		log.Printf("Received operators:")
		for _, op := range resp.Operators {
			log.Printf("- Address: %s, Status: %s, ActiveEpoch: %d, Weight: %s", 
				op.Address, op.Status, op.ActiveEpoch, op.Weight)
		}
	}
	node.PrintOperatorStates()
	
	return nil
}

func (node *Node) handleStateUpdates(stream network.Stream) {
	defer func() {
		log.Printf("[StateSync] State update handler stopping, closing stream...")
		stream.Close()
	}()

	log.Printf("[StateSync] Started handling state updates from registry")
	log.Printf("[StateSync] Stream ID: %s", stream.ID())

	for {
		// Read message type first
		msgType := make([]byte, 1)
		if _, err := io.ReadFull(stream, msgType); err != nil {
			if err != io.EOF {
				log.Printf("[StateSync] Error reading message type: %v", err)
			} else {
				log.Printf("[StateSync] State update stream closed by registry")
			}
			return
		}

		// Handle different message types
		switch msgType[0] {
		case MsgTypeStateUpdate:
			if err := node.handleStateUpdate(stream); err != nil {
				log.Printf("[StateSync] Error handling state update: %v", err)
				return
			}
		case MsgTypePeerSync:
			if err := node.handlePeerSync(stream); err != nil {
				log.Printf("[StateSync] Error handling peer sync: %v", err)
				return
			}
		default:
			log.Printf("[StateSync] Invalid message type: 0x%02x", msgType[0])
			return
		}
	}
}

func (node *Node) handleStateUpdate(stream network.Stream) error {
	// Read length-prefixed data with deadline
	data, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, 5*time.Second)
	if err != nil {
		return fmt.Errorf("error reading state update: %w", err)
	}

	// Validate message length
	const maxMessageSize = 1024 * 1024 // 1MB
	if len(data) > maxMessageSize {
		return fmt.Errorf("message too large (%d bytes), max allowed size is %d bytes", len(data), maxMessageSize)
	}

	var update pb.OperatorStateUpdate
	if err := proto.Unmarshal(data, &update); err != nil {
		return fmt.Errorf("error unmarshaling state update: %w", err)
	}

	// Update local state
	node.updateOperatorStates(update.Operators)
	log.Printf("[StateSync] Successfully processed update with %d operators", len(update.Operators))
	return nil
}

func (node *Node) handlePeerSync(stream network.Stream) error {
	// Read length-prefixed data with deadline
	data, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, 5*time.Second)
	if err != nil {
		log.Printf("[StateSync] Failed to read peer sync data: %v", err)
		return err
	}

	// Parse peer sync message
	var msg pb.PeerSyncMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		log.Printf("[StateSync] Failed to unmarshal peer sync message: %v", err)
		return err
	}

	// Update local peer list
	for _, peer := range msg.Peers {
		log.Printf("[StateSync] Received peer info: %s with %d multiaddrs, status: %s", 
			peer.PeerId, len(peer.Multiaddrs), peer.Status)
	}

	return nil
}

// updateOperatorStates updates the states of multiple operators
func (n *Node) updateOperatorStates(operators []*pb.OperatorState) {
	n.operators.mu.Lock()
	defer n.operators.mu.Unlock()

	log.Printf("[StateSync] Updating operator states in memory...")
	log.Printf("[StateSync] Current operator count before update: %d", len(n.operators.byAddress))
	
	// Update memory state
	for _, op := range operators {
		prevState := n.operators.byAddress[op.Address]
		
		// Create new operator state
		newState := &OperatorState{
			Address:     op.Address,
			SigningKey:  op.SigningKey,
			Status:      op.Status,
			LastSeen:    time.Now(),
			Weight:      op.Weight,
			ActiveEpoch: op.ActiveEpoch,
		}
		
		// Update state in maps
		n.operators.byAddress[op.Address] = newState
		
		if prevState != nil {
			log.Printf("[StateSync] Updated operator state - Address: %s\n  Old: Status=%s, ActiveEpoch=%d, Weight=%s\n  New: Status=%s, ActiveEpoch=%d, Weight=%s", 
				op.Address, 
				prevState.Status, prevState.ActiveEpoch, prevState.Weight,
				op.Status, op.ActiveEpoch, op.Weight)
				
			// Keep peer info if it exists
			if prevState.PeerID != "" {
				newState.PeerID = prevState.PeerID
				newState.Multiaddrs = prevState.Multiaddrs
				newState.AddrInfo = prevState.AddrInfo
				n.operators.byPeerID[prevState.PeerID] = newState
			}
		} else {
			log.Printf("[StateSync] Added new operator - Address: %s, SigningKey: %s, Status: %s, ActiveEpoch: %d, Weight: %s", 
				op.Address, op.SigningKey, op.Status, op.ActiveEpoch, op.Weight)
		}
	}
	
	log.Printf("[StateSync] Memory state updated - Current operator count: %d", len(n.operators.byAddress))
}

// PrintOperatorStates prints all operator states stored in memory
func (n *Node) PrintOperatorStates() {
	n.operators.mu.RLock()
	defer n.operators.mu.RUnlock()

	log.Printf("\nOperator States (%d total):", len(n.operators.byAddress))
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
	log.Printf("| %-42s | %-12s | %-12s | %-10s |", "Address", "Status", "ActiveEpoch", "Weight")
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
	
	for _, op := range n.operators.byAddress {
		log.Printf("| %-42s | %-12s | %-12d | %-10s |", 
			op.Address,
			op.Status,
			op.ActiveEpoch,
			op.Weight,
		)
	}
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
}

// GetOperatorState returns the state of a specific operator
func (n *Node) GetOperatorState(address string) *OperatorState {
	n.operators.mu.RLock()
	defer n.operators.mu.RUnlock()
	return n.operators.byAddress[address]
}

// GetOperatorCount returns the number of operators in memory
func (n *Node) GetOperatorCount() int {
	n.operators.mu.RLock()
	defer n.operators.mu.RUnlock()
	return len(n.operators.byAddress)
}

func (n *Node) calculateEpochNumber(blockNumber uint64) uint32 {
	return uint32((blockNumber - GenesisBlock) / EpochPeriod)
}

func (n *Node) updateEpochState(ctx context.Context, epochNumber uint32) error {
	// Get mainnet client
	mainnetClient, err := n.chainManager.GetMainnetClient()
	if err != nil {
		return fmt.Errorf("failed to get mainnet client: %w", err)
	}
	
	// Get epoch state from contract
	minimumStake, err := mainnetClient.GetMinimumWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get minimum stake: %w", err)
	}
	
	totalWeight, err := mainnetClient.GetTotalWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total stake: %w", err)
	}
	
	thresholdWeight, err := mainnetClient.GetThresholdWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get threshold weight: %w", err)
	}
	
	_, err = n.epochStateQuerier.UpsertEpochState(ctx, epoch_states.UpsertEpochStateParams{
		EpochNumber: epochNumber,
		BlockNumber: uint64(epochNumber * EpochPeriod),
		MinimumWeight: commonHelpers.BigIntToNumeric(minimumStake),
		TotalWeight: commonHelpers.BigIntToNumeric(totalWeight),
		ThresholdWeight: commonHelpers.BigIntToNumeric(thresholdWeight),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to update epoch state: %w", err)
	}

	log.Printf("[Epoch] Updated epoch %d state (minimum stake: %s, total weight: %s, threshold weight: %s)", 
		epochNumber, minimumStake.String(), totalWeight.String(), thresholdWeight.String())
	return nil
}

func (n *Node) monitorEpochUpdates(ctx context.Context) {
	ticker := time.NewTicker(epochMonitorInterval)
	defer ticker.Stop()

	var lastProcessedEpoch uint32

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get latest block number
			mainnetClient, err := n.chainManager.GetMainnetClient()
			if err != nil {
				log.Printf("[Epoch] Failed to get mainnet client: %v", err)
				continue
			}
			blockNumber, err := mainnetClient.BlockNumber(ctx)
			if err != nil {
				log.Printf("[Epoch] Failed to get latest block number: %v", err)
				continue
			}

			// Calculate current epoch
			currentEpoch := n.calculateEpochNumber(blockNumber)

			// If we're in a new epoch, update the state
			if currentEpoch > lastProcessedEpoch {
				if err := n.updateEpochState(ctx, currentEpoch); err != nil {
					log.Printf("[Epoch] Failed to update epoch state: %v", err)
					continue
				}
				lastProcessedEpoch = currentEpoch
				log.Printf("[Epoch] Successfully updated epoch %d", currentEpoch)
			}	
		}
	}
}


