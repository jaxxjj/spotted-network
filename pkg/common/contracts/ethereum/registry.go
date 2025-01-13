package ethereum

import (
	"context"
	"fmt"
	"math/big"

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

// GetThresholdStake gets the threshold stake required
func (c *RegistryClient) GetThresholdStake(ctx context.Context) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current epoch: %w", err)
	}
	stake, err := c.contract.GetTotalWeightAtEpoch(opts, epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to get threshold stake: %w", err)
	}
	return stake, nil
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
	// Create a channel for the contract events
	contractEvents := make(chan *bindings.ECDSAStakeRegistryOperatorRegistered)
	
	// Convert FilterOpts to WatchOpts
	var start *uint64
	if filterOpts.Start > 0 {
		startVal := filterOpts.Start
		start = &startVal
	}
	
	watchOpts := &bind.WatchOpts{
		Start:   start,
		Context: filterOpts.Context,
	}
	
	// Watch for events from the contract
	sub, err := c.contract.WatchOperatorRegistered(watchOpts, contractEvents, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to watch for operator registered events: %w", err)
	}

	// Start a goroutine to convert contract events to our event type
	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case event := <-contractEvents:
				sink <- &contracts.OperatorRegisteredEvent{
					Operator:    event.Operator,
					BlockNumber: event.BlockNumber,
					SigningKey:  event.SigningKey,
					Timestamp:   event.Timestamp,
					AVS:         event.Avs,
				}
			case <-sub.Err():
				return
			}
		}
	}()

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