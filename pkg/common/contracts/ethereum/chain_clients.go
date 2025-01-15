package ethereum

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/internal/avs/config"
	"github.com/galxe/spotted-network/pkg/common/contracts"
)

// ChainClients manages multiple chain clients
type ChainClients struct {
	mainnetClient contracts.Client                    // 主网客户端，包含所有合约
	stateClients  map[string]contracts.StateClient   // 其他链的状态客户端
	mu            sync.RWMutex
}

// NewChainClients creates clients for all configured chains
func NewChainClients(cfg *config.Config) (*ChainClients, error) {
	chainClients := &ChainClients{
		stateClients: make(map[string]contracts.StateClient),
	}

	// 初始化每个链的客户端
	for chainName, chainCfg := range cfg.Chains {
		if chainName == "ethereum" {
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
				return nil, fmt.Errorf("failed to create state client for chain %s: %w", chainName, err)
			}
			chainClients.stateClients[chainName] = stateClient
		}
	}

	if chainClients.mainnetClient == nil {
		return nil, fmt.Errorf("ethereum mainnet client is required")
	}

	return chainClients, nil
}

// GetMainnetClient returns the ethereum mainnet client
func (c *ChainClients) GetMainnetClient() contracts.Client {
	return c.mainnetClient
}

// GetStateClient returns the state client for a specific chain
func (c *ChainClients) GetStateClient(chainName string) (contracts.StateClient, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if chainName == "ethereum" {
		return c.mainnetClient.StateClient(), nil
	}

	client, exists := c.stateClients[chainName]
	if !exists {
		return nil, fmt.Errorf("no state client found for chain: %s", chainName)
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
	for chainName, client := range c.stateClients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close state client for chain %s: %w", chainName, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}
	return nil
}

// AddStateClient adds a new state client for a chain
func (c *ChainClients) AddStateClient(chainName string, rpc string, stateManagerAddr common.Address) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainName == "ethereum" {
		return fmt.Errorf("cannot add ethereum as state-only client")
	}

	if _, exists := c.stateClients[chainName]; exists {
		return fmt.Errorf("client already exists for chain: %s", chainName)
	}

	client, err := NewStateOnlyClient(rpc, stateManagerAddr)
	if err != nil {
		return fmt.Errorf("failed to create state client for chain %s: %w", chainName, err)
	}

	c.stateClients[chainName] = client
	return nil
}

// RemoveStateClient removes and closes a state client
func (c *ChainClients) RemoveStateClient(chainName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainName == "ethereum" {
		return fmt.Errorf("cannot remove ethereum mainnet client")
	}

	client, exists := c.stateClients[chainName]
	if !exists {
		return fmt.Errorf("no client found for chain: %s", chainName)
	}

	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close client for chain %s: %w", chainName, err)
	}

	delete(c.stateClients, chainName)
	return nil
} 