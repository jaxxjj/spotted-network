package ethereum

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/contracts"
	"github.com/galxe/spotted-network/pkg/config"
)

const (
	MainnetChainID = int64(31337) // Ethereum mainnet chain ID
)

// ChainClients manages multiple chain clients
type ChainClients struct {
	mainnetClient contracts.Client                    // 主网客户端，包含所有合约
	stateClients  map[int64]contracts.StateClient    // 其他链的状态客户端
	mu            sync.RWMutex
}

// NewChainClients creates clients for all configured chains
func NewChainClients(cfg *config.Config) (*ChainClients, error) {
	chainClients := &ChainClients{
		stateClients: make(map[int64]contracts.StateClient),
	}

	// 初始化每个链的客户端
	for chainID, chainCfg := range cfg.Chains {
		if chainID == MainnetChainID {
			// 为主网创建完整客户端
			clientCfg := &Config{
				EpochManagerAddress:      common.HexToAddress(chainCfg.Contracts.EpochManager),
				RegistryAddress:         common.HexToAddress(chainCfg.Contracts.Registry),
				StateManagerAddress:     common.HexToAddress(chainCfg.Contracts.StateManager),
				RPCEndpoint:            chainCfg.RPC,
			}

			client, err := NewClient(clientCfg)
			if err != nil {
				chainClients.Close()
				return nil, fmt.Errorf("failed to create mainnet client: %w", err)
			}
			chainClients.mainnetClient = client
		} else {
			// 为其他链只创建 StateManager 客户端
			stateClient, err := NewStateOnlyClient(chainCfg.RPC, common.HexToAddress(chainCfg.Contracts.StateManager))
			if err != nil {
				chainClients.Close()
				return nil, fmt.Errorf("failed to create state client for chain %d: %w", chainID, err)
			}
			chainClients.stateClients[chainID] = stateClient
		}
	}

	if chainClients.mainnetClient == nil {
		return nil, fmt.Errorf("ethereum mainnet client (chain ID %d) is required", MainnetChainID)
	}

	return chainClients, nil
}

// GetMainnetClient returns the ethereum mainnet client
func (c *ChainClients) GetMainnetClient() contracts.Client {
	return c.mainnetClient
}

// GetStateClient returns the state client for a specific chain
func (c *ChainClients) GetStateClient(chainID int64) (contracts.StateClient, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if chainID == MainnetChainID {
		return c.mainnetClient.StateClient(), nil
	}

	client, exists := c.stateClients[chainID]
	if !exists {
		return nil, fmt.Errorf("no state client found for chain ID: %d", chainID)
	}
	return client, nil
}

// Close closes all chain clients
func (c *ChainClients) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	// Close mainnet client
	if c.mainnetClient != nil {
		if err := c.mainnetClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close mainnet client: %w", err))
		}
	}

	// Close state clients
	for chainID, client := range c.stateClients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close state client for chain %d: %w", chainID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}
	return nil
}

// AddStateClient adds a new state client for a chain
func (c *ChainClients) AddStateClient(chainID int64, rpc string, stateManagerAddr common.Address) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainID == MainnetChainID {
		return fmt.Errorf("cannot add ethereum mainnet (chain ID %d) as state-only client", MainnetChainID)
	}

	if _, exists := c.stateClients[chainID]; exists {
		return fmt.Errorf("client already exists for chain ID: %d", chainID)
	}

	client, err := NewStateOnlyClient(rpc, stateManagerAddr)
	if err != nil {
		return fmt.Errorf("failed to create state client for chain %d: %w", chainID, err)
	}

	c.stateClients[chainID] = client
	return nil
}

// RemoveStateClient removes and closes a state client
func (c *ChainClients) RemoveStateClient(chainID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainID == MainnetChainID {
		return fmt.Errorf("cannot remove ethereum mainnet client (chain ID %d)", MainnetChainID)
	}

	client, exists := c.stateClients[chainID]
	if !exists {
		return fmt.Errorf("no client found for chain ID: %d", chainID)
	}

	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close client for chain %d: %w", chainID, err)
	}

	delete(c.stateClients, chainID)
	return nil
} 