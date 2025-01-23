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

// monitorEpochUpdates monitors for epoch changes and updates operator states
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
			currentEpoch := calculateEpochNumber(blockNumber)

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

func (n *Node) updateOperatorStates(ctx context.Context, currentEpoch uint32) error {
	updatedOperators := make([]*pb.OperatorState, 0)
	mainnetClient := n.chainClients.GetMainnetClient()
	currentBlock, err := mainnetClient.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// Get all operators
	allOps, err := n.db.ListAllOperators(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all operators: %w", err)
	}

	log.Printf("[Registry] Processing %d operators for epoch %d (block %d)", len(allOps), currentEpoch, currentBlock)

	for _, op := range allOps {
		activeEpoch := uint32(op.ActiveEpoch.Int.Int64())
		exitEpoch := uint32(op.ExitEpoch.Int.Int64())
		
		// Use helper function to determine status
		newStatus, logMsg := DetermineOperatorStatus(currentBlock, activeEpoch, exitEpoch)
		log.Printf("[Node] %s", logMsg)

		// Only update if status has changed
		if newStatus != op.Status {
			// Update operator state in database
			updatedOp, err := n.db.UpdateOperatorStatus(ctx, operators.UpdateOperatorStatusParams{
				Address: op.Address,
				Status:  newStatus,
			})
			if err != nil {
				log.Printf("[Registry] Failed to update operator %s state: %v", op.Address, err)
				continue
			}

			// Get weight if status is active
			var weight *big.Int
			if newStatus == "active" {
				weight, err = mainnetClient.GetOperatorWeight(ctx, ethcommon.HexToAddress(op.Address))
				if err != nil {
					log.Printf("[Registry] Failed to get weight for operator %s: %v", op.Address, err)
					continue
				}
			} else {
				weight = big.NewInt(0)
			}

			// Add to updated operators list
			var exitEpochPtr *int32
			if exitEpoch != 4294967295 {
				exitEpochInt32 := int32(exitEpoch)
				exitEpochPtr = &exitEpochInt32
			}

			updatedOperators = append(updatedOperators, &pb.OperatorState{
				Address:                 updatedOp.Address,
				SigningKey:             updatedOp.SigningKey,
				RegisteredAtBlockNumber: updatedOp.RegisteredAtBlockNumber.Int.Int64(),
				RegisteredAtTimestamp:   updatedOp.RegisteredAtTimestamp.Int.Int64(),
				ActiveEpoch:            int32(activeEpoch),
				ExitEpoch:              exitEpochPtr,
				Status:                 newStatus,
				Weight:                 weight.String(),
			})

			log.Printf("[Registry] Updated operator %s status from %s to %s at epoch %d (block %d)", 
				op.Address, op.Status, newStatus, currentEpoch, currentBlock)
		}
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