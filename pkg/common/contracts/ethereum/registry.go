package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts"
	"github.com/galxe/spotted-network/pkg/common/contracts/bindings"
)

// RegistryClient implements the contracts.RegistryClient interface
type RegistryClient struct {
	contract     *bindings.ECDSAStakeRegistry
	epochManager *EpochClient
}

// NewRegistryClient creates a new registry client instance
func NewRegistryClient(contract *bindings.ECDSAStakeRegistry, epochManager *EpochClient) *RegistryClient {
	return &RegistryClient{
		contract:     contract,
		epochManager: epochManager,
	}
}

// GetOperatorWeight gets the stake weight for an operator
func (c *RegistryClient) GetOperatorWeight(ctx context.Context, operator common.Address) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current epoch: %w", err)
	}
	weight, err := c.contract.GetOperatorWeightAtEpoch(opts, operator, epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator weight: %w", err)
	}
	return weight, nil
}

// GetTotalWeight gets the total stake weight
func (c *RegistryClient) GetTotalWeight(ctx context.Context) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current epoch: %w", err)
	}
	weight, err := c.contract.GetTotalWeightAtEpoch(opts, epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to get total weight: %w", err)
	}
	return weight, nil
}

// GetMinimumStake gets the minimum stake required
func (c *RegistryClient) GetMinimumStake(ctx context.Context) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	stake, err := c.contract.MinimumWeight(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get minimum stake: %w", err)
	}
	return stake, nil
}

// GetThresholdWeight gets the threshold weight required
func (c *RegistryClient) GetThresholdWeight(ctx context.Context) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current epoch: %w", err)
	}
	weight, err := c.contract.GetThresholdWeightAtEpoch(opts, epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to get threshold weight: %w", err)
	}
	return weight, nil
}

// IsOperatorRegistered checks if an operator is registered
func (c *RegistryClient) IsOperatorRegistered(ctx context.Context, operator common.Address) (bool, error) {
	opts := &bind.CallOpts{Context: ctx}
	registered, err := c.contract.OperatorRegistered(opts, operator)
	if err != nil {
		return false, fmt.Errorf("failed to check if operator is registered: %w", err)
	}
	return registered, nil
}

// WatchOperatorRegistered watches for operator registered events
func (c *RegistryClient) WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *contracts.OperatorRegisteredEvent) (event.Subscription, error) {
	// Create a subscription
	sub := event.NewSubscription(func(quit <-chan struct{}) error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		var lastBlock uint64
		if filterOpts.Start > 0 {
			lastBlock = filterOpts.Start - 1 // Start from the block before to ensure we don't miss any events
		}

		for {
			select {
			case <-quit:
				return nil
			case <-ticker.C:
				// Get events from last processed block
				events, err := c.contract.FilterOperatorRegistered(&bind.FilterOpts{
					Start:   lastBlock + 1,
					Context: filterOpts.Context,
				}, nil, nil, nil)
				if err != nil {
					continue
				}

				// Process events
				for events.Next() {
					event := events.Event
					sink <- &contracts.OperatorRegisteredEvent{
						Operator:    event.Operator,
						BlockNumber: event.BlockNumber,
						SigningKey:  event.SigningKey,
						Timestamp:   event.Timestamp,
						AVS:        event.Avs,
					}
					// Update last block to the event's block number
					if event.BlockNumber.Uint64() > lastBlock {
						lastBlock = event.BlockNumber.Uint64()
					}
				}
			}
		}
	})

	return sub, nil
}

// Forward EpochClient methods to the epoch manager

func (c *RegistryClient) GetCurrentEpoch(ctx context.Context) (uint32, error) {
	return c.epochManager.GetCurrentEpoch(ctx)
}

func (c *RegistryClient) GetEpochInterval(ctx context.Context, epoch uint32) (startBlock, graceBlock, endBlock uint64, err error) {
	return c.epochManager.GetEpochInterval(ctx, epoch)
}

func (c *RegistryClient) CanAdvanceEpoch(ctx context.Context) (bool, error) {
	return c.epochManager.CanAdvanceEpoch(ctx)
}

func (c *RegistryClient) BlocksUntilNextEpoch(ctx context.Context) (uint64, error) {
	return c.epochManager.BlocksUntilNextEpoch(ctx)
}

func (c *RegistryClient) IsInGracePeriod(ctx context.Context) (bool, error) {
	return c.epochManager.IsInGracePeriod(ctx)
}

func (c *RegistryClient) GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error) {
	return c.epochManager.GetEffectiveEpochForBlock(ctx, blockNumber)
} 