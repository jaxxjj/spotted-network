package ethereum

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/galxe/spotted-network/pkg/common/contracts/bindings"
)

// StateClient implements the contracts.StateClient interface
type StateClient struct {
	contract *bindings.StateManager
	ethClient *ethclient.Client
}

// creates a new state client instance
func NewStateClient(contract *bindings.StateManager, ethClient *ethclient.Client) *StateClient {
	return &StateClient{
		contract: contract,
		ethClient: ethClient,
	}
}

// gets state at a specific block number
func (c *StateClient) GetStateAtBlock(ctx context.Context, target common.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	history, err := c.contract.GetHistoryAt(opts, target, key, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get state at block: %w", err)
	}
	return history.Value, nil
}

// gets state at a specific timestamp
func (c *StateClient) GetStateAtTimestamp(ctx context.Context, target common.Address, key *big.Int, timestamp uint64) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	// Get all history after timestamp and take the first entry
	histories, err := c.contract.GetHistoryAfterTimestamp(opts, target, key, big.NewInt(int64(timestamp)))
	if err != nil {
		return nil, fmt.Errorf("failed to get state at timestamp: %w", err)
	}
	if len(histories) == 0 {
		return nil, fmt.Errorf("no state found at timestamp %d", timestamp)
	}
	return histories[0].Value, nil
}

// gets the latest state
func (c *StateClient) GetLatestState(ctx context.Context, target common.Address, key *big.Int) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	return c.contract.GetCurrentValue(opts, target, key)
}

// gets the latest block number
func (c *StateClient) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.ethClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %w", err)
	}
	return blockNumber, nil
}

// Close closes the client connection
func (c *StateClient) Close() error {
	return nil
} 