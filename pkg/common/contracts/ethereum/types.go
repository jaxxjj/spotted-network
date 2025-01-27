package ethereum

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/galxe/spotted-network/pkg/common/contracts/bindings"
)

const (
	MainnetChainID = int64(31337) // Ethereum mainnet chain ID
)

// Config contains Ethereum client configuration
type Config struct {
	EpochManagerAddress  common.Address
	RegistryAddress      common.Address
	StateManagerAddress  common.Address
	RPCEndpoint         string
}

// ChainClient is the Ethereum client implementation that can interact with
// state manager, epoch manager, and registry contracts
type ChainClient struct {
	ethclient.Client
	stateManager  *bindings.StateManager
	epochManager  *bindings.EpochManager
	registry      *bindings.ECDSAStakeRegistry
}

// ChainClientManager manages multiple chain clients
type ChainClientManager struct {
	clients map[int64]*ChainClient
	mu      sync.RWMutex
}


// OperatorRegisteredEvent represents an operator registered event
type OperatorRegisteredEvent struct {
	Operator    common.Address
	BlockNumber *big.Int
	SigningKey  common.Address
	Timestamp   *big.Int
	AVS         common.Address
}

// OperatorDeregisteredEvent represents an operator deregistered event
type OperatorDeregisteredEvent struct {
	Operator    common.Address
	BlockNumber *big.Int
	AVS         common.Address
} 