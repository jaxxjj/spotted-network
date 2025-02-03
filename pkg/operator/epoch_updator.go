package operator

import (
	"context"
	"fmt"
	"log"
	"time"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
)

func (n *Node) updateEpochState(ctx context.Context, epochNumber uint32) error {
	// Get epoch state from contract
	mainnetClient, err := n.chainManager.GetMainnetClient()
	if err != nil {
		return fmt.Errorf("failed to get mainnet client: %w", err)
	}
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
		MinimumWeight: utils.BigIntToNumeric(minimumStake),
		TotalWeight: utils.BigIntToNumeric(totalWeight),
		ThresholdWeight: utils.BigIntToNumeric(thresholdWeight),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to update epoch state: %w", err)
	}

	log.Printf("[Epoch] Updated epoch %d state (minimum stake: %s, total weight: %s, threshold weight: %s)", 
		epochNumber, minimumStake.String(), totalWeight.String(), thresholdWeight.String())
	return nil
}


