package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
)
const (
	StateSyncProtocol = "/spotted/state-sync/1.0.0"
)

type EpochUpdatorQuerier interface {
	UpsertEpochState(ctx context.Context, arg epoch_states.UpsertEpochStateParams) (*epoch_states.EpochState, error)
}

type EpochUpdatorChainClient interface {
	GetMinimumWeight(ctx context.Context) (*big.Int, error)
	GetTotalWeight(ctx context.Context) (*big.Int, error)
	GetThresholdWeight(ctx context.Context) (*big.Int, error)
}

type EpochUpdator struct {
	node *Node
	epochQuerier EpochUpdatorQuerier
	chainClient EpochUpdatorChainClient
	lastProcessedEpoch uint32
	pubsub PubSubService
	epochUpdateTopic ResponseTopic
}

func NewEpochUpdator(ctx context.Context, node *Node, epochQuerier EpochUpdatorQuerier, chainClient EpochUpdatorChainClient) (*EpochUpdator, error) {
	// Create instance first
	e := &EpochUpdator{
		node: node,
		epochQuerier: epochQuerier,
		chainClient: chainClient,
	}
	
	log.Printf("[Epoch] Epoch monitoring started")
	
	return e, nil
}


func (e *EpochUpdator) updateEpochState(ctx context.Context, epochNumber uint32) error {
	// Get epoch state from contract
	minimumStake, err := e.chainClient.GetMinimumWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get minimum stake: %w", err)
	}
	
	totalWeight, err := e.chainClient.GetTotalWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total stake: %w", err)
	}
	
	thresholdWeight, err := e.chainClient.GetThresholdWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get threshold weight: %w", err)
	}
	
	_, err = e.epochQuerier.UpsertEpochState(ctx, epoch_states.UpsertEpochStateParams{
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



func (e *EpochUpdator) calculateEpochNumber(blockNumber uint64) uint32 {
	return uint32((blockNumber - GenesisBlock) / EpochPeriod)
}

