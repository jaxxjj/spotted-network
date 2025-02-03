package common

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-msgio"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
)

const (
	GenesisBlock = 0
	EpochPeriod = 12
)

// ChainClient defines the interface for getting blocks that the helper functions need
type ChainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

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

// NumericToBigInt converts pgtype.Numeric to *big.Int with proper handling of exponent
func NumericToBigInt(n pgtype.Numeric) (*big.Int, error) {
	if n.Exp == 0 {
		return n.Int, nil
    }
    big10 := big.NewInt(10)
    big0 := big.NewInt(0)
    num := &big.Int{}
    num.Set(n.Int)
    if n.Exp > 0 {
        mul := &big.Int{}
        mul.Exp(big10, big.NewInt(int64(n.Exp)), nil)
        num.Mul(num, mul)
        return num, nil
    }

    div := &big.Int{}
    div.Exp(big10, big.NewInt(int64(-n.Exp)), nil)
    remainder := &big.Int{}
    num.DivMod(num, div, remainder)
    if remainder.Cmp(big0) != 0 {
        return nil, fmt.Errorf("cannot convert %v to integer", n)
    }
    return num, nil
}

// NumericToInt64 converts a pgtype.Numeric to int64
func NumericToInt64(num pgtype.Numeric) int64 {
    if !num.Valid {
        return 0
    }
    fl, _ := num.Float64Value()
    return int64(fl.Float64)
}

// NumericToUint64 converts a pgtype.Numeric to uint64
func NumericToUint64(num pgtype.Numeric) uint64 {
    if !num.Valid {
        return 0
    }
    fl, _ := num.Float64Value()
    return uint64(fl.Float64)
}

func StringToBigInt(s string) *big.Int {
	if s == "" {
		return big.NewInt(0)
	}

	x, _ := new(big.Int).SetString(s, 10)

	return x
}

// BigIntToNumeric converts a *big.Int to a pgtype.Numeric
func BigIntToNumeric(x *big.Int) pgtype.Numeric {
	return pgtype.Numeric{Int: x, Valid: true}
}

// StringToNumeric converts a string to a pgtype.Numeric
func StringToNumeric(s string) pgtype.Numeric {
	return BigIntToNumeric(StringToBigInt(s))
}

func Uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}


// BlockNumberToTimestamp converts a block number to its corresponding timestamp
func BlockNumberToTimestamp(ctx context.Context, client ChainClient, chainID uint32, blockNumber uint64) (uint64, error) {
	// Get block by number
	block, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		return 0, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
	}

	// Return block timestamp
	return block.Time(), nil
}

// TimestampToBlockNumber converts a timestamp to its nearest block number
// for a specific chain. It uses a more efficient algorithm that takes into
// account the average block time for the chain.
func TimestampToBlockNumber(ctx context.Context, client ChainClient, chainID uint32, timestamp uint64) (uint64, error) {
	// Get latest block
	currentBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block: %w", err)
	}

	// Get current block
	currentBlock, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(currentBlockNumber))
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
	for block.Time() > timestamp {
		// Calculate number of blocks to go back
		decreaseBlocks := float64(block.Time()-timestamp) / averageBlockTime
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
		block, err = client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNumber))
		if err != nil {
			return 0, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
		}
	}

	// Second pass: fine tune by walking one block at a time
	// If we overshot (block time < target), walk forward
	for block.Time() < timestamp {
		nextBlockNumber := blockNumber + 1
		if nextBlockNumber > currentBlockNumber {
			break
		}

		nextBlock, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(nextBlockNumber))
		if err != nil {
			break
		}

		// If next block would overshoot, stop here
		if nextBlock.Time() > timestamp {
			// Return the closest block
			if timestamp-block.Time() < nextBlock.Time()-timestamp {
				return blockNumber, nil
			}
			return nextBlockNumber, nil
		}

		block = nextBlock
		blockNumber = nextBlockNumber
	}

	// If we undershot (block time > target), walk backward
	for block.Time() > timestamp {
		if blockNumber == 0 {
			break
		}

		prevBlockNumber := blockNumber - 1
		prevBlock, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(prevBlockNumber))
		if err != nil {
			break
		}

		// If previous block would undershoot, stop here
		if prevBlock.Time() < timestamp {
			// Return the closest block
			if timestamp-prevBlock.Time() < block.Time()-timestamp {
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
func ValidateBlockNumberAndTimestamp(ctx context.Context, client ChainClient, chainID uint32, blockNumber *uint64, timestamp *uint64) (uint64, uint64, error) {
	// Check that exactly one is provided
	if (blockNumber == nil && timestamp == nil) || (blockNumber != nil && timestamp != nil) {
		return 0, 0, fmt.Errorf("exactly one of block_number or timestamp must be provided")
	}

	var resultBlockNumber uint64
	var resultTimestamp uint64
	var err error

	if blockNumber != nil {
		// Convert block number to timestamp
		resultBlockNumber = *blockNumber
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

func CalculateEpochNumber(blockNumber uint64) uint32 {
	return uint32((blockNumber - GenesisBlock) / EpochPeriod)
}

// writeLengthPrefix writes a 4-byte length prefix to the stream
func WriteLengthPrefix(stream network.Stream, length uint32) error {
    lengthBuf := make([]byte, 4)
    binary.BigEndian.PutUint32(lengthBuf, length)
    
    if _, err := stream.Write(lengthBuf); err != nil {
        return fmt.Errorf("failed to write length prefix: %w", err)
    }
    return nil
}

// readLengthPrefix reads a 4-byte length prefix from the stream
func ReadLengthPrefix(stream network.Stream) (uint32, error) {
    lengthBuf := make([]byte, 4)
    if _, err := io.ReadFull(stream, lengthBuf); err != nil {
        return 0, fmt.Errorf("failed to read length prefix: %w", err)
    }
    
    length := binary.BigEndian.Uint32(lengthBuf)
    return length, nil
}

// WriteLengthPrefixedData writes length-prefixed data to a stream
func WriteLengthPrefixedData(stream network.Stream, data []byte) error {
    if err := WriteLengthPrefix(stream, uint32(len(data))); err != nil {
        return err
    }
    
    if _, err := stream.Write(data); err != nil {
        return fmt.Errorf("failed to write data: %w", err)
    }
    
    return nil
}

func MultiaddrsToStrings(multiaddrs []multiaddr.Multiaddr) []string {
	strs := make([]string, len(multiaddrs))
	for i, addr := range multiaddrs {
		strs[i] = addr.String()
	}
	return strs
}

func StringsToMultiaddrs(addrs []string) ([]multiaddr.Multiaddr, error) {
    multiaddrs := make([]multiaddr.Multiaddr, 0, len(addrs))
    
    for _, addr := range addrs {
        maddr, err := multiaddr.NewMultiaddr(addr)
        if err != nil {
            return nil, fmt.Errorf("invalid multiaddr %s: %w", addr, err)
        }
        multiaddrs = append(multiaddrs, maddr)
    }
    
    return multiaddrs, nil
}


// readMessage reads a protobuf message from a stream
func ReadStreamMessage(stream network.Stream, msg proto.Message, maxMessageSize int) error {
	reader := msgio.NewVarintReaderSize(stream, maxMessageSize)
	bytes, err := reader.ReadMsg()
	if err != nil {
		return err
	}
	defer reader.ReleaseMsg(bytes)
	
	return proto.Unmarshal(bytes, msg)
}

// writeMessage writes a protobuf message to a stream
func WriteStreamMessage(stream network.Stream, msg proto.Message) error {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	
	writer := msgio.NewVarintWriter(stream)
	return writer.WriteMsg(bytes)
}