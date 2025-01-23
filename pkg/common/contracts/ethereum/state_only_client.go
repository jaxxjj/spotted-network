package ethereum

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/galxe/spotted-network/pkg/common/contracts"
	"github.com/galxe/spotted-network/pkg/common/contracts/bindings"
)

// StateOnlyClient implements the contracts.StateClient interface
type StateOnlyClient struct {
	ethClient *ethclient.Client
	contract  *bindings.StateManager
}

// NewStateOnlyClient creates a new state-only client
func NewStateOnlyClient(rpcEndpoint string, stateManagerAddr common.Address) (contracts.StateClient, error) {
	// Connect to Ethereum node
	ethClient, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Create StateManager contract binding
	contract, err := bindings.NewStateManager(stateManagerAddr, ethClient)
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("failed to create state manager binding: %w", err)
	}

	return &StateOnlyClient{
		ethClient: ethClient,
		contract:  contract,
	}, nil
}

// GetStateAtBlock gets state at a specific block number
func (c *StateOnlyClient) GetStateAtBlock(ctx context.Context, target common.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	history, err := c.contract.GetHistoryAtBlock(opts, target, key, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get state at block: %w", err)
	}
	return history.Value, nil
}

// GetStateAtTimestamp gets state at a specific timestamp
func (c *StateOnlyClient) GetStateAtTimestamp(ctx context.Context, target common.Address, key *big.Int, timestamp uint64) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	history, err := c.contract.GetHistoryAfterTimestamp(opts, target, key, big.NewInt(int64(timestamp)))
	if err != nil {
		return nil, fmt.Errorf("failed to get state at timestamp: %w", err)
	}
	if len(history) == 0 {
		return nil, fmt.Errorf("no state found after timestamp")
	}
	return history[0].Value, nil
}

// GetLatestState gets the latest state
func (c *StateOnlyClient) GetLatestState(ctx context.Context, target common.Address, key *big.Int) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	value, err := c.contract.GetCurrentValue(opts, target, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest state: %w", err)
	}
	return value, nil
}

// GetLatestBlockNumber gets the latest block number
func (c *StateOnlyClient) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.ethClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %w", err)
	}
	return blockNumber, nil
}

// GetEthClient returns the underlying ethclient.Client
func (c *StateOnlyClient) GetEthClient() *ethclient.Client {
	return c.ethClient
}

// Close closes the client connection
func (c *StateOnlyClient) Close() error {
	c.ethClient.Close()
	return nil
} 