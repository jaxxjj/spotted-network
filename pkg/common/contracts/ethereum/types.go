package ethereum

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/galxe/spotted-network/pkg/common/contracts/bindings"
)

const (
	MainnetChainID = uint32(31337) // Ethereum mainnet chain ID
)

// Config 更明确的配置结构
type Config struct {
	ChainID             uint32
	EpochManagerAddress common.Address
	RegistryAddress     common.Address
	StateManagerAddress common.Address
	RPCEndpoint         string
}

// ChainClient represents a client for a specific Ethereum chain
type ChainClient struct {
	client       *ethclient.Client
	stateManager *bindings.StateManager
	epochManager *bindings.EpochManager
	registry     *bindings.ECDSAStakeRegistry
}

// ChainManager manages multiple chain instances
type ChainManager struct {
	chains map[uint32]*ChainClient
	mu     sync.RWMutex
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