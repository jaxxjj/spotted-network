package config

// Config represents the AVS configuration
type Config struct {
	// Chains contains configuration for each blockchain
	Chains map[string]*ChainConfig
}

// ChainConfig represents configuration for a specific blockchain
type ChainConfig struct {
	// RPC endpoint for the chain
	RPC string

	// Contract addresses
	Contracts ContractConfig
}

// ContractConfig contains addresses for deployed contracts
type ContractConfig struct {
	// EpochManager contract address
	EpochManager string

	// Registry contract address (ECDSAStakeRegistry)
	Registry string

	// StateManager contract address
	StateManager string
} 