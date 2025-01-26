package ethereum

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts"
	"github.com/galxe/spotted-network/pkg/common/contracts/bindings"
)

// Client implements the contracts.Client interface for Ethereum
type Client struct {
	ethClient  *ethclient.Client
	epoch      *EpochClient
	registry   *RegistryClient
	state      *StateClient
}

// Config contains Ethereum client configuration
type Config struct {
	EpochManagerAddress      common.Address
	RegistryAddress         common.Address
	StateManagerAddress     common.Address
	RPCEndpoint            string
}

// NewClient creates a new Ethereum contract client
func NewClient(cfg *Config) (contracts.Client, error) {
	// Connect to Ethereum node
	ethClient, err := ethclient.Dial(cfg.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Create contract bindings
	epochManager, err := bindings.NewEpochManager(cfg.EpochManagerAddress, ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create epoch manager binding: %w", err)
	}

	registry, err := bindings.NewECDSAStakeRegistry(cfg.RegistryAddress, ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry binding: %w", err)
	}

	stateManager, err := bindings.NewStateManager(cfg.StateManagerAddress, ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager binding: %w", err)
	}

	// Create epoch client first since registry needs it
	epochClient := NewEpochClient(epochManager)

	// Create client instance
	client := &Client{
		ethClient:  ethClient,
		epoch:      epochClient,
		registry:   NewRegistryClient(registry, epochClient),
		state:      NewStateClient(stateManager, ethClient),
	}

	return client, nil
}

// CallOpts creates a new CallOpts instance for read operations
func (c *Client) CallOpts(ctx context.Context) *bind.CallOpts {
	return &bind.CallOpts{Context: ctx}
}

// Close closes the Ethereum client connection
func (c *Client) Close() error {
	c.ethClient.Close()
	return nil
}

// EpochClient returns the epoch client implementation
func (c *Client) EpochClient() contracts.EpochClient {
	return c.epoch
}

// RegistryClient returns the registry client implementation
func (c *Client) RegistryClient() contracts.RegistryClient {
	return c.registry
}

// StateClient returns the state client implementation
func (c *Client) StateClient() contracts.StateClient {
	return c.state
}

// BlocksUntilNextEpoch returns the number of blocks until the next epoch
func (c *Client) BlocksUntilNextEpoch(ctx context.Context) (uint64, error) {
	return c.epoch.BlocksUntilNextEpoch(ctx)
}

// GetCurrentEpoch returns the current epoch number
func (c *Client) GetCurrentEpoch(ctx context.Context) (uint32, error) {
	return c.epoch.GetCurrentEpoch(ctx)
}

// CanAdvanceEpoch checks if the epoch can be advanced
func (c *Client) CanAdvanceEpoch(ctx context.Context) (bool, error) {
	return c.epoch.CanAdvanceEpoch(ctx)
}

// GetEpochInterval returns the epoch interval information
func (c *Client) GetEpochInterval(ctx context.Context, epoch uint32) (startBlock, graceBlock, endBlock uint64, err error) {
	return c.epoch.GetEpochInterval(ctx, epoch)
}

// GetLatestState returns the latest state
func (c *Client) GetLatestState(ctx context.Context, target common.Address, key *big.Int) (*big.Int, error) {
	return c.state.GetLatestState(ctx, target, key)
}

// GetMinimumStake returns the minimum stake required
func (c *Client) GetMinimumStake(ctx context.Context) (*big.Int, error) {
	return c.registry.GetMinimumStake(ctx)
}

// GetOperatorWeight returns the operator stake weight
func (c *Client) GetOperatorWeight(ctx context.Context, operator common.Address) (*big.Int, error) {
	return c.registry.GetOperatorWeight(ctx, operator)
}

// GetTotalWeight returns the total stake weight
func (c *Client) GetTotalWeight(ctx context.Context) (*big.Int, error) {
	return c.registry.GetTotalWeight(ctx)
}

// GetThresholdStake returns the threshold stake required
func (c *Client) GetThresholdWeight(ctx context.Context) (*big.Int, error) {
	return c.registry.GetThresholdWeight(ctx)
}

// IsOperatorRegistered checks if an operator is registered
func (c *Client) IsOperatorRegistered(ctx context.Context, operator common.Address) (bool, error) {
	return c.registry.IsOperatorRegistered(ctx, operator)
}

// IsInGracePeriod checks if currently in grace period
func (c *Client) IsInGracePeriod(ctx context.Context) (bool, error) {
	return c.epoch.IsInGracePeriod(ctx)
}

// GetStateAtBlock returns the state at a specific block
func (c *Client) GetStateAtBlock(ctx context.Context, target common.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	return c.state.GetStateAtBlock(ctx, target, key, blockNumber)
}

// GetStateAtTimestamp returns the state at a specific timestamp
func (c *Client) GetStateAtTimestamp(ctx context.Context, target common.Address, key *big.Int, timestamp uint64) (*big.Int, error) {
	return c.state.GetStateAtTimestamp(ctx, target, key, timestamp)
}

// GetLatestBlockNumber returns the latest block number from the chain
func (c *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	return c.state.GetLatestBlockNumber(ctx)
}

// GetEffectiveEpochForBlock returns the effective epoch for a specific block number
func (c *Client) GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error) {
	return c.epoch.GetEffectiveEpochForBlock(ctx, blockNumber)
}

// WatchOperatorRegistered forwards to registry client
func (c *Client) WatchOperatorRegistered(opts *bind.FilterOpts, sink chan<- *contracts.OperatorRegisteredEvent) (event.Subscription, error) {
	return c.registry.WatchOperatorRegistered(opts, sink)
}

// WatchOperatorDeregistered forwards to registry client
func (c *Client) WatchOperatorDeregistered(opts *bind.FilterOpts, sink chan<- *contracts.OperatorDeregisteredEvent) (event.Subscription, error) {
	return c.registry.WatchOperatorDeregistered(opts, sink)
}
