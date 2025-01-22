package contracts

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

// Client defines the interface for interacting with all contracts
type Client interface {
	EpochClient
	RegistryClient
	StateClient

	// Close closes the client connection
	Close() error

	// StateClient returns the state client implementation
	StateClient() StateClient
}

// EpochClient defines interactions with the EpochManager contract
type EpochClient interface {
	// Get current epoch number
	GetCurrentEpoch(ctx context.Context) (uint32, error)
	
	// Get epoch interval info
	GetEpochInterval(ctx context.Context, epoch uint32) (startBlock, graceBlock, endBlock uint64, err error)
	
	// Check if can advance epoch
	CanAdvanceEpoch(ctx context.Context) (bool, error)
	
	// Get blocks until next epoch
	BlocksUntilNextEpoch(ctx context.Context) (uint64, error)
	
	// Check if in grace period
	IsInGracePeriod(ctx context.Context) (bool, error)

	// Get effective epoch for a specific block number
	GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error)
}

// RegistryClient defines interactions with the ECDSAStakeRegistry contract
type RegistryClient interface {
	EpochClient // Embed EpochClient to ensure RegistryClient implements all EpochClient methods

	// Get operator stake weight
	GetOperatorWeight(ctx context.Context, operator common.Address) (*big.Int, error)
	
	// Get total stake weight
	GetTotalWeight(ctx context.Context) (*big.Int, error)
	
	// Get minimum stake required
	GetMinimumStake(ctx context.Context) (*big.Int, error)
	
	// Get threshold stake required
	GetThresholdWeight(ctx context.Context) (*big.Int, error)
	
	// Check if operator is registered
	IsOperatorRegistered(ctx context.Context, operator common.Address) (bool, error)

	// Watch for operator registered events
	WatchOperatorRegistered(opts *bind.FilterOpts, sink chan<- *OperatorRegisteredEvent) (event.Subscription, error)

	// Watch for operator deregistered events
	WatchOperatorDeregistered(opts *bind.FilterOpts, sink chan<- *OperatorDeregisteredEvent) (event.Subscription, error)
}

// StateClient defines interactions with the StateManager contract
type StateClient interface {
	// Get state at block
	GetStateAtBlock(ctx context.Context, target common.Address, key *big.Int, blockNumber uint64) (*big.Int, error)
	
	// Get state at timestamp
	GetStateAtTimestamp(ctx context.Context, target common.Address, key *big.Int, timestamp uint64) (*big.Int, error)
	
	// Get latest state
	GetLatestState(ctx context.Context, target common.Address, key *big.Int) (*big.Int, error)

	// Get latest block number
	GetLatestBlockNumber(ctx context.Context) (uint64, error)

	// Close closes the client connection
	Close() error
}

// OperatorRegisteredEvent represents an operator registered event
type OperatorRegisteredEvent struct {
	Operator    common.Address
	BlockNumber *big.Int
	SigningKey  common.Address
	Timestamp   *big.Int
	AVS         common.Address
} 

type OperatorDeregisteredEvent struct {
	Operator    common.Address
	BlockNumber *big.Int
	AVS         common.Address
}