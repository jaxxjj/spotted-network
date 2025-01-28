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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

const (
	GenesisBlock = 0
	EpochPeriod  = 12
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
	msgType := []byte{0x02} // 0x02 for SubscribeRequest
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
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)
	
	if _, err := stream.Write(lengthBytes); err != nil {
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
	msgType := []byte{0x01} // 0x01 for GetFullStateRequest
	if _, err := stream.Write(msgType); err != nil {
		return fmt.Errorf("[StateSync] failed to write message type: %w", err)
	}

	// Send get full state request
	req := &pb.GetFullStateRequest{}
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("[StateSync] failed to marshal get state request: %w", err)
	}

	// Send length prefix
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)
	
	if _, err := stream.Write(lengthBytes); err != nil {
		return fmt.Errorf("[StateSync] failed to write length prefix: %w", err)
	}

	log.Printf("[StateSync] Sending GetFullState request to registry (type: 0x01, length: %d)...", length)
	bytesWritten, err := stream.Write(data)
	if err != nil {
		return fmt.Errorf("[StateSync] failed to send get state request: %w", err)
	}
	log.Printf("[StateSync] Sent %d bytes to registry", bytesWritten)

	// Read response
	log.Printf("[StateSync] Waiting for response from registry...")
	
	// Read length prefix first
	respLengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(stream, respLengthBytes); err != nil {
		return fmt.Errorf("failed to read response length: %w", err)
	}
	
	respLength := uint32(respLengthBytes[0])<<24 | 
		uint32(respLengthBytes[1])<<16 | 
		uint32(respLengthBytes[2])<<8 | 
		uint32(respLengthBytes[3])
		
	log.Printf("Response length: %d bytes", respLength)
	
	// Read the full response
	respData := make([]byte, respLength)
	if _, err := io.ReadFull(stream, respData); err != nil {
		return fmt.Errorf("[StateSync] failed to read response data: %w", err)
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

		// Validate message type
		if msgType[0] != 0x02 {
			log.Printf("[StateSync] Invalid message type: 0x%02x, expected 0x02", msgType[0])
			return
		}
		
		// Read length prefix
		lengthBytes := make([]byte, 4)
		if _, err := io.ReadFull(stream, lengthBytes); err != nil {
			if err != io.EOF {
				log.Printf("[StateSync] Error reading update length: %v", err)
			}
			return
		}
		
		length := uint32(lengthBytes[0])<<24 | 
			uint32(lengthBytes[1])<<16 | 
			uint32(lengthBytes[2])<<8 | 
			uint32(lengthBytes[3])
			
		// Validate message length (max 1MB)
		const maxMessageSize = 1024 * 1024 // 1MB
		if length > maxMessageSize {
			log.Printf("[StateSync] Message too large (%d bytes), max allowed size is %d bytes", length, maxMessageSize)
			return
		}

		log.Printf("[StateSync] Received state update - Type: 0x%02x, Length: %d bytes", msgType[0], length)
		
		// Read the full update with timeout
		data := make([]byte, length)
		if err := stream.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			log.Printf("[StateSync] Failed to set read deadline: %v", err)
			return
		}
		
		if _, err := io.ReadFull(stream, data); err != nil {
			log.Printf("[StateSync] Error reading state update data: %v", err)
			return
		}

		// Reset deadline
		if err := stream.SetReadDeadline(time.Time{}); err != nil {
			log.Printf("[StateSync] Failed to reset read deadline: %v", err)
			return
		}

		var update pb.OperatorStateUpdate
		if err := proto.Unmarshal(data, &update); err != nil {
			log.Printf("[StateSync] Error unmarshaling state update: %v", err)
			continue
		}

		log.Printf("[StateSync] Successfully unmarshaled update with %d operators", len(update.Operators))
		
		// Print detailed operator states before update
		log.Printf("\n[StateSync] === Current Operator States Before Update ===")
		node.PrintOperatorStates()
		
		// Update states
		node.updateOperatorStates(update.Operators)
		
		// Print detailed operator states after update
		log.Printf("\n[StateSync] === Updated Operator States ===")
		node.PrintOperatorStates()
		
		log.Printf("[StateSync] State update processed successfully")
	}
}

func (node *Node) updateOperatorStates(operators []*pb.OperatorState) {
	node.statesMu.Lock()
	defer node.statesMu.Unlock()

	log.Printf("[StateSync] Updating operator states in memory...")
	log.Printf("[StateSync] Current operator count before update: %d", len(node.operatorStates))
	
	// Update memory state
	for _, op := range operators {
		prevState := node.operatorStates[op.Address]
		node.operatorStates[op.Address] = op
		
		if prevState != nil {
			log.Printf("[StateSync] Updated operator state - Address: %s\n  Old: Status=%s, ActiveEpoch=%d, Weight=%s\n  New: Status=%s, ActiveEpoch=%d, Weight=%s", 
				op.Address, 
				prevState.Status, prevState.ActiveEpoch, prevState.Weight,
				op.Status, op.ActiveEpoch, op.Weight)
		} else {
			log.Printf("[StateSync] Added new operator - Address: %s, Status: %s, ActiveEpoch: %d, Weight: %s", 
				op.Address, op.Status, op.ActiveEpoch, op.Weight)
		}
	}
	
	log.Printf("[StateSync] Memory state updated - Current operator count: %d", len(node.operatorStates))
}

// PrintOperatorStates prints all operator states stored in memory
func (node *Node) PrintOperatorStates() {
	node.statesMu.RLock()
	defer node.statesMu.RUnlock()

	log.Printf("\nOperator States (%d total):", len(node.operatorStates))
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
	log.Printf("| %-42s | %-12s | %-12s | %-10s |", "Address", "Status", "ActiveEpoch", "Weight")
	log.Printf("+-%-42s-+-%-12s-+-%-12s-+-%-10s-+", strings.Repeat("-", 42), strings.Repeat("-", 12), strings.Repeat("-", 12), strings.Repeat("-", 10))
	
	for _, op := range node.operatorStates {
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
func (node *Node) GetOperatorState(address string) *pb.OperatorState {
	node.statesMu.RLock()
	defer node.statesMu.RUnlock()
	return node.operatorStates[address]
}

// GetOperatorCount returns the number of operators in memory
func (node *Node) GetOperatorCount() int {
	node.statesMu.RLock()
	defer node.statesMu.RUnlock()
	return len(node.operatorStates)
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

	// Update epoch state in database
	now := time.Now()
	
	updatedAt := pgtype.Timestamptz{}
	updatedAt.Scan(now)
	
	_, err = n.epochState.UpsertEpochState(ctx, epoch_states.UpsertEpochStateParams{
		EpochNumber: epochNumber,
		BlockNumber: uint64(epochNumber * EpochPeriod),
		MinimumWeight: commonHelpers.BigIntToNumeric(minimumStake),
		TotalWeight: commonHelpers.BigIntToNumeric(totalWeight),
		ThresholdWeight: commonHelpers.BigIntToNumeric(thresholdWeight),
		UpdatedAt: updatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to update epoch state: %w", err)
	}

	log.Printf("[Epoch] Updated epoch %d state (minimum stake: %s, total weight: %s, threshold weight: %s)", 
		epochNumber, minimumStake.String(), totalWeight.String(), thresholdWeight.String())
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