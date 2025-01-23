package ethereum

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/galxe/spotted-network/pkg/common/contracts/bindings"
)

// EpochClient implements the contracts.EpochClient interface
type EpochClient struct {
	contract *bindings.EpochManager
}

// NewEpochClient creates a new epoch client instance
func NewEpochClient(contract *bindings.EpochManager) *EpochClient {
	return &EpochClient{
		contract: contract,
	}
}

// GetCurrentEpoch gets the current epoch number
func (c *EpochClient) GetCurrentEpoch(ctx context.Context) (uint32, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.contract.GetCurrentEpoch(opts)
	if err != nil {
		return 0, fmt.Errorf("[EpochClient] failed to get current epoch: %w", err)
	}
	return epoch, nil
}

// GetEpochInterval gets the start, grace and end blocks for an epoch
func (c *EpochClient) GetEpochInterval(ctx context.Context, epoch uint32) (startBlock, graceBlock, endBlock uint64, err error) {
	opts := &bind.CallOpts{Context: ctx}
	interval, err := c.contract.GetEpochInterval(opts, epoch)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("[EpochClient] failed to get epoch interval: %w", err)
	}
	return interval.StartBlock, interval.GraceBlock, interval.EndBlock, nil
}

// CanAdvanceEpoch checks if the epoch can be advanced
func (c *EpochClient) CanAdvanceEpoch(ctx context.Context) (bool, error) {
	opts := &bind.CallOpts{Context: ctx}
	can, err := c.contract.CanAdvanceEpoch(opts)
	if err != nil {
		return false, fmt.Errorf("[EpochClient] failed to check if can advance epoch: %w", err)
	}
	return can, nil
}

// BlocksUntilNextEpoch gets the number of blocks until the next epoch
func (c *EpochClient) BlocksUntilNextEpoch(ctx context.Context) (uint64, error) {
	opts := &bind.CallOpts{Context: ctx}
	blocks, err := c.contract.BlocksUntilNextEpoch(opts)
	if err != nil {
		return 0, fmt.Errorf("[EpochClient] failed to get blocks until next epoch: %w", err)
	}
	return blocks, nil
}

// IsInGracePeriod checks if currently in grace period
func (c *EpochClient) IsInGracePeriod(ctx context.Context) (bool, error) {
	opts := &bind.CallOpts{Context: ctx}
	inGrace, err := c.contract.IsInGracePeriod(opts)
	if err != nil {
		return false, fmt.Errorf("[EpochClient] failed to check if in grace period: %w", err)
	}
	return inGrace, nil
}

// GetEffectiveEpochForBlock returns the effective epoch for a specific block number
func (c *EpochClient) GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error) {
	opts := &bind.CallOpts{Context: ctx}
	epoch, err := c.contract.GetEffectiveEpochForBlock(opts, blockNumber)
	if err != nil {
		return 0, fmt.Errorf("[EpochClient] failed to get effective epoch for block: %w", err)
	}
	return epoch, nil
} 