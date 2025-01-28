package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
)

type EpochUpdator struct {
	node *Node
	lastProcessedEpoch uint32
}


func NewEpochUpdator(node *Node) *EpochUpdator {
	return &EpochUpdator{
		node: node,
	}
}

// Start starts monitoring epoch updates
func (e *EpochUpdator) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	log.Printf("[Epoch] Starting epoch monitoring...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Get latest block number
			blockNumber, err := e.node.mainnetClient.BlockNumber(ctx)
			if err != nil {
				log.Printf("[Epoch] Failed to get latest block number: %v", err)
				continue
			}

			// Calculate current epoch
			currentEpoch := calculateEpochNumber(blockNumber)

			// If we're in a new epoch, update operator states
			if currentEpoch > e.lastProcessedEpoch {
				if err := e.updateOperatorStates(ctx, currentEpoch); err != nil {
					log.Printf("[Epoch] Failed to update operator states for epoch %d: %v", currentEpoch, err)
					continue
				}
				e.lastProcessedEpoch = currentEpoch
				log.Printf("[Epoch] Updated operator states for epoch %d", currentEpoch)
			}
		}
	}
}

func (e *EpochUpdator) updateOperatorStates(ctx context.Context, currentEpoch uint32) error {
	updatedOperators := make([]*pb.OperatorState, 0)
	currentBlock, err := e.node.mainnetClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// Get all operators
	allOperators, err := e.node.operators.ListAllOperators(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all operators: %w", err)
	}

	log.Printf("[Epoch] Processing %d operators for epoch %d (block %d)", len(allOperators), currentEpoch, currentBlock)

	for _, operator := range allOperators {
		activeEpoch := operator.ActiveEpoch
		exitEpoch := operator.ExitEpoch
		
		// Use helper function to determine status
		newStatus, logMsg := DetermineOperatorStatus(currentBlock, activeEpoch, exitEpoch)
		log.Printf("[Node] %s", logMsg)

		// Only update if status has changed
		if newStatus != operator.Status {
			// Update operator state in database
			updatedOp, err := e.node.operators.UpdateOperatorState(ctx, operators.UpdateOperatorStateParams{
				Address: operator.Address,
				Status:  newStatus,
			})
			if err != nil {
				log.Printf("[Epoch] Failed to update operator %s state: %v", operator.Address, err)
				continue
			}

			// Get weight if status is active
			var weight *big.Int
			if newStatus == "active" {
				weight, err = e.node.mainnetClient.GetOperatorWeight(ctx, ethcommon.HexToAddress(operator.Address))
				if err != nil {
					log.Printf("[Epoch] Failed to get weight for operator %s: %v", operator.Address, err)
					continue
				}
			} else {
				weight = big.NewInt(0)
			}

			// Add to updated operators list
			var exitEpochPtr *uint32
			if exitEpoch != 4294967295 {
				exitEpochCopy := exitEpoch
				exitEpochPtr = &exitEpochCopy
			}

			updatedOperators = append(updatedOperators, &pb.OperatorState{
				Address:                 updatedOp.Address,
				SigningKey:             updatedOp.SigningKey,
				RegisteredAtBlockNumber: updatedOp.RegisteredAtBlockNumber,
				RegisteredAtTimestamp:   updatedOp.RegisteredAtTimestamp,
				ActiveEpoch:            activeEpoch,
				ExitEpoch:              exitEpochPtr,
				Status:                 string(newStatus),
				Weight:                 weight.String(),
			})

			log.Printf("[Epoch] Updated operator %s status from %s to %s at epoch %d (block %d)", 
				operator.Address, operator.Status, newStatus, currentEpoch, currentBlock)
		}
	}

	// Broadcast updates if any operators were updated
	if len(updatedOperators) > 0 {
		e.node.BroadcastStateUpdate(updatedOperators, "state_update")
		log.Printf("[Epoch] Broadcast state update for %d operators at epoch %d", len(updatedOperators), currentEpoch)
	} else {
		log.Printf("[Epoch] No operator state changes needed for epoch %d", currentEpoch)
	}

	return nil
}