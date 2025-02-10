package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// ChainManager defines the interface for chain management
type ChainManager interface {
	// GetMainnetClient returns the mainnet client
	GetMainnetClient() (ChainClient, error)
	// GetClientByChainId returns the appropriate client for a given chain ID
	GetClientByChainId(chainID uint32) (ChainClient, error)
	// Close closes all chains
	Close() error
}

// ChainClient defines the interface for chain operations
type ChainClient interface {
	// Basic methods
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	Close() error

	// State methods
	GetLatestState(ctx context.Context, target common.Address, key *big.Int) (*big.Int, error)
	GetStateAtBlock(ctx context.Context, target common.Address, key *big.Int, blockNumber uint64) (*big.Int, error)
	GetStateAtTimestamp(ctx context.Context, target common.Address, key *big.Int, timestamp uint64) (*big.Int, error)

	// Epoch methods
	GetCurrentEpoch(ctx context.Context) (uint32, error)
	GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error)

	// Registry methods
	GetOperatorWeight(ctx context.Context, operator common.Address) (*big.Int, error)
	GetTotalWeight(ctx context.Context) (*big.Int, error)
	GetMinimumWeight(ctx context.Context) (*big.Int, error)
	GetThresholdWeight(ctx context.Context) (*big.Int, error)
	IsOperatorRegistered(ctx context.Context, operator common.Address) (bool, error)
	GetOperatorSigningKey(ctx context.Context, operator common.Address, epoch uint32) (common.Address, error)
	GetOperatorP2PKey(ctx context.Context, operator common.Address, epoch uint32) (common.Address, error)

	// Watch methods
	WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *OperatorRegisteredEvent) (event.Subscription, error)
	WatchOperatorDeregistered(filterOpts *bind.FilterOpts, sink chan<- *OperatorDeregisteredEvent) (event.Subscription, error)
}

// Event types that are part of the public interface
type OperatorRegisteredEvent struct {
	Operator    common.Address
	BlockNumber *big.Int
	SigningKey  common.Address
	P2PKey      common.Address
	AVS         common.Address
}

type OperatorDeregisteredEvent struct {
	Operator    common.Address
	BlockNumber *big.Int
	AVS         common.Address
}
