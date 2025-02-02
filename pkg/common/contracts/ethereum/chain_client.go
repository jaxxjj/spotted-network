package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts/bindings"
)

// NewChainClient creates a new Ethereum client for a specific chain
func NewChainClient(cfg *Config) (*ChainClient, error) {
	ethClient, err := ethclient.Dial(cfg.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to connect to Ethereum node: %w", err)
	}

	stateManager, err := bindings.NewStateManager(cfg.StateManagerAddress, ethClient)
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("[ChainClient] failed to create state manager binding: %w", err)
	}

	epochManager, err := bindings.NewEpochManager(cfg.EpochManagerAddress, ethClient)
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("[ChainClient] failed to create epoch manager binding: %w", err)
	}

	registry, err := bindings.NewECDSAStakeRegistry(cfg.RegistryAddress, ethClient)
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("[ChainClient] failed to create registry binding: %w", err)
	}

	return &ChainClient{
		client:         ethClient,
		stateManager: stateManager,
		epochManager: epochManager,
		registry:     registry,
	}, nil
}

// Close implements contracts.ChainClient
func (c *ChainClient) Close() error {
	c.client.Close()
	return nil
}

// State methods
func (c *ChainClient) GetLatestState(ctx context.Context, target common.Address, key *big.Int) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	value, err := c.stateManager.GetCurrentValue(opts, target, key)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get latest state: %w", err)
	}
	return value, nil
}

func (c *ChainClient) GetStateAtBlock(ctx context.Context, target common.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	history, err := c.stateManager.GetHistoryAtBlock(opts, target, key, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get state at block: %w", err)
	}
	return history.Value, nil
}

func (c *ChainClient) GetStateAtTimestamp(ctx context.Context, target common.Address, key *big.Int, timestamp uint64) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	history, err := c.stateManager.GetHistoryAfterTimestamp(opts, target, key, big.NewInt(int64(timestamp)))
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get state at timestamp: %w", err)
	}
	if len(history) == 0 {
		return nil, fmt.Errorf("[ChainClient] no state found after timestamp")
	}
	return history[0].Value, nil
}

// Epoch methods
func (c *ChainClient) GetCurrentEpoch(ctx context.Context) (uint32, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetCurrentEpoch(opts)
	if err != nil {
		return 0, fmt.Errorf("[ChainClient] failed to get current epoch: %w", err)
	}
	return epoch, nil
}

func (c *ChainClient) GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetEffectiveEpochForBlock(opts, blockNumber)
	if err != nil {
		return 0, fmt.Errorf("[ChainClient] failed to get effective epoch for block: %w", err)
	}
	return epoch, nil
}

// Registry methods
// GetOperatorWeight gets the stake weight for an operator
func (c *ChainClient) GetOperatorWeight(ctx context.Context, operator common.Address) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetCurrentEpoch(opts)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get current epoch: %w", err)
	}
	weight, err := c.registry.GetOperatorWeightAtEpoch(opts, operator, epoch)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get operator weight: %w", err)
	}
	return weight, nil
}

// GetTotalWeight gets the total stake weight
func (c *ChainClient) GetTotalWeight(ctx context.Context) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetCurrentEpoch(opts)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get current epoch: %w", err)
	}
	weight, err := c.registry.GetTotalWeightAtEpoch(opts, epoch)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get total weight: %w", err)
	}
	return weight, nil
}

// gets the minimum stake required
func (c *ChainClient) GetMinimumWeight(ctx context.Context) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	stake, err := c.registry.MinimumWeight(opts)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get minimum stake: %w", err)
	}
	return stake, nil
}

// GetThresholdWeight gets the threshold weight required
func (c *ChainClient) GetThresholdWeight(ctx context.Context) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.epochManager.GetCurrentEpoch(opts)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get current epoch: %w", err)
	}
	weight, err := c.registry.GetThresholdWeightAtEpoch(opts, epoch)
	if err != nil {
		return nil, fmt.Errorf("[ChainClient] failed to get threshold weight: %w", err)
	}
	return weight, nil
}

// IsOperatorRegistered checks if an operator is registered
func (c *ChainClient) IsOperatorRegistered(ctx context.Context, operator common.Address) (bool, error) {
	opts := &bind.CallOpts{Context: ctx}
	registered, err := c.registry.OperatorRegistered(opts, operator)
	if err != nil {
		return false, fmt.Errorf("[ChainClient] failed to check if operator is registered: %w", err)
	}
	return registered, nil
}

// BlockNumber returns the latest block number
func (c *ChainClient) BlockNumber(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}

func (c *ChainClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
    return c.client.BlockByNumber(ctx, number)
}

// WatchOperatorRegistered watches for operator registered events
func (c *ChainClient) WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *OperatorRegisteredEvent) (event.Subscription, error) {
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
				events, err := c.registry.FilterOperatorRegistered(&bind.FilterOpts{
					Start:   lastBlock + 1,
					Context: filterOpts.Context,
				}, nil, nil, nil)
				if err != nil {
					continue
				}

				// Process events
				for events.Next() {
					event := events.Event
					sink <- &OperatorRegisteredEvent{
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

// WatchOperatorDeregistered watches for operator deregistration events
func (c *ChainClient) WatchOperatorDeregistered(filterOpts *bind.FilterOpts, sink chan<- *OperatorDeregisteredEvent) (event.Subscription, error) {
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
				events, err := c.registry.FilterOperatorDeregistered(&bind.FilterOpts{
					Start:   lastBlock + 1,
					Context: filterOpts.Context,
				}, nil, nil, nil)
				if err != nil {
					continue
				}

				// Process events
				for events.Next() {
					event := events.Event
					sink <- &OperatorDeregisteredEvent{
						Operator:    event.Operator,
						BlockNumber: event.BlockNumber,
						AVS:         event.Avs,
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

// GetOperatorSigningKey gets the signing key for an operator at a specific epoch
func (c *ChainClient) GetOperatorSigningKey(ctx context.Context, operator common.Address, epoch uint32) (common.Address, error) {
	opts := &bind.CallOpts{Context: ctx}
	signingKey, err := c.registry.GetOperatorSigningKeyAtEpoch(opts, operator, epoch)
	if err != nil {
		return common.Address{}, fmt.Errorf("[ChainClient] failed to get operator signing key at epoch %d: %w", epoch, err)
	}
	return signingKey, nil
}