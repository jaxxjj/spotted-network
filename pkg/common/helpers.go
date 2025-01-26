package common

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/galxe/spotted-network/pkg/common/contracts"
	"github.com/jackc/pgx/v5/pgtype"
)

// NumericToString converts pgtype.Numeric to string with proper handling of exponent
func NumericToString(n pgtype.Numeric) string {
	if !n.Valid || n.Int == nil {
		return "0"
	}
	weight := new(big.Int).Set(n.Int)
	if n.Exp > 0 {
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)
		weight.Mul(weight, multiplier)
	} else if n.Exp < 0 {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)
		weight.Div(weight, divisor)
	}
	return weight.String()
}

// BlockNumberToTimestamp converts a block number to its corresponding timestamp
// for a specific chain
func BlockNumberToTimestamp(ctx context.Context, client contracts.StateClient, chainID int64, blockNumber uint64) (int64, error) {
	// Get ethclient from state client
	ethClient := client.(interface{ GetEthClient() *ethclient.Client }).GetEthClient()

	// Get block by number
	block, err := ethClient.BlockByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		return 0, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
	}

	// Return block timestamp
	return int64(block.Time()), nil
}

// TimestampToBlockNumber converts a timestamp to its nearest block number
// for a specific chain. It uses a more efficient algorithm that takes into
// account the average block time for the chain.
func TimestampToBlockNumber(ctx context.Context, client contracts.StateClient, chainID int64, timestamp int64) (uint64, error) {
	// Get latest block
	currentBlockNumber, err := client.GetLatestBlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block: %w", err)
	}

	// Get ethclient from state client
	ethClient := client.(interface{ GetEthClient() *ethclient.Client }).GetEthClient()

	// Get current block
	currentBlock, err := ethClient.BlockByNumber(ctx, new(big.Int).SetUint64(currentBlockNumber))
	if err != nil {
		return 0, fmt.Errorf("failed to get current block: %w", err)
	}

	// Use default average block time based on chain ID
	var averageBlockTime float64
	switch chainID {
	case 1: // Ethereum mainnet
		averageBlockTime = 12.5
	case 137: // Polygon
		averageBlockTime = 2.0
	case 56: // BSC
		averageBlockTime = 3.0
	default:
		averageBlockTime = 12.5 // Default to Ethereum-like chains
	}

	blockNumber := currentBlockNumber
	block := currentBlock

	// First pass: use average block time to get close to target
	for block.Time() > uint64(timestamp) {
		// Calculate number of blocks to go back
		decreaseBlocks := float64(block.Time()-uint64(timestamp)) / averageBlockTime
		decreaseBlocks = math.Floor(decreaseBlocks)

		if decreaseBlocks < 1 {
			break
		}

		// Don't go below block 0
		if blockNumber <= uint64(decreaseBlocks) {
			blockNumber = 0
			break
		}

		blockNumber -= uint64(decreaseBlocks)

		// Get new block
		block, err = ethClient.BlockByNumber(ctx, new(big.Int).SetUint64(blockNumber))
		if err != nil {
			return 0, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
		}
	}

	// Second pass: fine tune by walking one block at a time
	// If we overshot (block time < target), walk forward
	for block.Time() < uint64(timestamp) {
		nextBlockNumber := blockNumber + 1
		if nextBlockNumber > currentBlockNumber {
			break
		}

		nextBlock, err := ethClient.BlockByNumber(ctx, new(big.Int).SetUint64(nextBlockNumber))
		if err != nil {
			break
		}

		// If next block would overshoot, stop here
		if nextBlock.Time() > uint64(timestamp) {
			// Return the closest block
			if uint64(timestamp)-block.Time() < nextBlock.Time()-uint64(timestamp) {
				return blockNumber, nil
			}
			return nextBlockNumber, nil
		}

		block = nextBlock
		blockNumber = nextBlockNumber
	}

	// If we undershot (block time > target), walk backward
	for block.Time() > uint64(timestamp) {
		if blockNumber == 0 {
			break
		}

		prevBlockNumber := blockNumber - 1
		prevBlock, err := ethClient.BlockByNumber(ctx, new(big.Int).SetUint64(prevBlockNumber))
		if err != nil {
			break
		}

		// If previous block would undershoot, stop here
		if prevBlock.Time() < uint64(timestamp) {
			// Return the closest block
			if uint64(timestamp)-prevBlock.Time() < block.Time()-uint64(timestamp) {
				return prevBlockNumber, nil
			}
			return blockNumber, nil
		}

		block = prevBlock
		blockNumber = prevBlockNumber
	}

	return blockNumber, nil
}

// ValidateBlockNumberAndTimestamp validates that exactly one of blockNumber or timestamp is provided
// and converts between them as needed
func ValidateBlockNumberAndTimestamp(ctx context.Context, client contracts.StateClient, chainID int64, blockNumber *int64, timestamp *int64) (uint64, int64, error) {
	// Check that exactly one is provided
	if (blockNumber == nil && timestamp == nil) || (blockNumber != nil && timestamp != nil) {
		return 0, 0, fmt.Errorf("exactly one of block_number or timestamp must be provided")
	}

	var resultBlockNumber uint64
	var resultTimestamp int64
	var err error

	if blockNumber != nil {
		// Convert block number to timestamp
		resultBlockNumber = uint64(*blockNumber)
		resultTimestamp, err = BlockNumberToTimestamp(ctx, client, chainID, resultBlockNumber)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to convert block number to timestamp: %w", err)
		}
	} else {
		// Convert timestamp to block number
		resultTimestamp = *timestamp
		resultBlockNumber, err = TimestampToBlockNumber(ctx, client, chainID, resultTimestamp)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to convert timestamp to block number: %w", err)
		}
	}

	return resultBlockNumber, resultTimestamp, nil
}
