package ethereum

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/config"
)

// NewManager creates a new chain manager from application config
func NewManager(cfg *config.Config) (*ChainManager, error) {
	chainConfigs := make(map[uint32]*Config)
	for chainID, chainCfg := range cfg.Chains {
		chainConfigs[chainID] = &Config{
			ChainID:            chainID,
			RPCEndpoint:        chainCfg.RPC,
			StateManagerAddress: common.HexToAddress(chainCfg.Contracts.StateManager),
			EpochManagerAddress: common.HexToAddress(chainCfg.Contracts.EpochManager),
			RegistryAddress:    common.HexToAddress(chainCfg.Contracts.Registry),
		}
	}
	
	return NewChainManager(chainConfigs)
}

// NewChainManager creates a new chain client manager
func NewChainManager(configs map[uint32]*Config) (*ChainManager, error) {
	m := &ChainManager{
		chains: make(map[uint32]*ChainClient),
		mu:     sync.RWMutex{},
	}

	// Initialize clients for each chain
	for chainID, config := range configs {
		chain, err := NewChainClient(config)
		if err != nil {
			m.Close()
			return nil, fmt.Errorf("failed to initialize chain %d: %w", chainID, err)
		}
		m.chains[chainID] = chain
	}

	return m, nil
}

// GetChain returns the chain instance for a given chain ID
func (m *ChainManager) GetClientByChainId(chainID uint32) (*ChainClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chain, exists := m.chains[chainID]
	if !exists {
		return nil, fmt.Errorf("chain not found: %d", chainID)
	}
	return chain, nil
}

// GetMainnetChain returns the mainnet chain instance
func (m *ChainManager) GetMainnetClient() (*ChainClient, error) {
	return m.GetClientByChainId(MainnetChainID)
}

// AddChain adds a new chain
func (m *ChainManager) AddChain(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.chains[config.ChainID]; exists {
		return fmt.Errorf("chain already exists: %d", config.ChainID)
	}

	chain, err := NewChainClient(config)
	if err != nil {
		return fmt.Errorf("failed to create chain: %w", err)
	}

	m.chains[config.ChainID] = chain
	return nil
}

// RemoveChain removes a chain
func (m *ChainManager) RemoveChain(chainID uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	chain, exists := m.chains[chainID]
	if !exists {
		return fmt.Errorf("chain not found: %d", chainID)
	}

	if err := chain.Close(); err != nil {
		return fmt.Errorf("failed to close chain: %w", err)
	}

	delete(m.chains, chainID)
	return nil
}

// Close closes all chains
func (m *ChainManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for chainID, chain := range m.chains {
		if err := chain.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close chain %d: %w", chainID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing chains: %v", errs)
	}
	return nil
}

// ListChains returns all available chain IDs
func (m *ChainManager) ListChains() []uint32 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chainIDs := make([]uint32, 0, len(m.chains))
	for chainID := range m.chains {
		chainIDs = append(chainIDs, chainID)
	}
	return chainIDs
} 