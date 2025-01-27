package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/config"
)

// NewChainClientManager creates a new chain client manager
func NewChainClientManager(cfg *config.Config) (*ChainClientManager, error) {
	manager := &ChainClientManager{
		clients: make(map[int64]*ChainClient),
	}

	// Initialize clients for each chain
	for chainID, chainCfg := range cfg.Chains {
		clientCfg := &Config{
			EpochManagerAddress: common.HexToAddress(chainCfg.Contracts.EpochManager),
			RegistryAddress:    common.HexToAddress(chainCfg.Contracts.Registry),
			StateManagerAddress: common.HexToAddress(chainCfg.Contracts.StateManager),
			RPCEndpoint:       chainCfg.RPC,
		}

		client, err := NewChainClient(clientCfg)
		if err != nil {
			manager.Close()
			return nil, fmt.Errorf("failed to create client for chain %d: %w", chainID, err)
		}
		manager.clients[chainID] = client
	}

	// Ensure mainnet client exists
	if _, exists := manager.clients[MainnetChainID]; !exists {
		return nil, fmt.Errorf("ethereum mainnet client (chain ID %d) is required", MainnetChainID)
	}

	return manager, nil
}

// GetMainnetClient returns the mainnet client
func (c *ChainClientManager) GetMainnetClient() (*ChainClient, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client, exists := c.clients[MainnetChainID]
	if !exists {
		return nil, fmt.Errorf("mainnet client not initialized")
	}
	return client, nil
}

// GetClientByChainId returns the appropriate client for a given chain ID
func (c *ChainClientManager) GetClientByChainId(chainID int64) (*ChainClient, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client, exists := c.clients[chainID]
	if !exists {
		return nil, fmt.Errorf("no client found for chain ID: %d", chainID)
	}
	return client, nil
}

// AddClient adds a new client for a chain
func (c *ChainClientManager) AddClient(chainID int64, client *ChainClient) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.clients[chainID]; exists {
		return fmt.Errorf("client already exists for chain ID: %d", chainID)
	}

	c.clients[chainID] = client
	return nil
}

// RemoveClient removes a client for a chain
func (c *ChainClientManager) RemoveClient(chainID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chainID == MainnetChainID {
		return fmt.Errorf("cannot remove mainnet client (chain ID %d)", MainnetChainID)
	}

	if client, exists := c.clients[chainID]; exists {
		if err := client.Close(); err != nil {
			return fmt.Errorf("failed to close client: %w", err)
		}
		delete(c.clients, chainID)
		return nil
	}

	return fmt.Errorf("no client found for chain ID: %d", chainID)
}

// Close closes all clients
func (c *ChainClientManager) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	for chainID, client := range c.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close client for chain %d: %w", chainID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}
	return nil
} 