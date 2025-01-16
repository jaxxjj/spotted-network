package registry

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ChainConfig struct {
	// RPC endpoint for the chain
	RPC string `yaml:"rpc"`

	// Contract addresses
	Contracts ContractConfig `yaml:"contracts"`
}

type ContractConfig struct {
	// EpochManager contract address
	EpochManager string `yaml:"epoch_manager"`

	// Registry contract address (ECDSAStakeRegistry)
	Registry string `yaml:"registry"`

	// StateManager contract address
	StateManager string `yaml:"state_manager"`
}

type DatabaseConfig struct {
	URL             string `yaml:"url"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
}

type P2PConfig struct {
	Port           int      `yaml:"port"`
	BootstrapNodes []string `yaml:"bootstrap_nodes"`
	ExternalIP     string   `yaml:"external_ip"`
}

type HTTPConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type Config struct {
	// Chain configuration
	Chain struct {
		ID        int    `yaml:"id"`
		RPCURL    string `yaml:"rpc_url"`
		StartBlock int64  `yaml:"start_block"`
	} `yaml:"chain"`

	// Chains configuration (multiple chain support)
	Chains map[string]*ChainConfig `yaml:"chains"`

	// Database configuration
	Database DatabaseConfig `yaml:"database"`

	// P2P configuration
	P2P P2PConfig `yaml:"p2p"`

	// HTTP configuration
	HTTP HTTPConfig `yaml:"http"`

	// Logging configuration
	Logging LogConfig `yaml:"logging"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		Chain: struct {
			ID        int    `yaml:"id"`
			RPCURL    string `yaml:"rpc_url"`
			StartBlock int64  `yaml:"start_block"`
		}{
			ID:        31337, // Anvil chain ID
			RPCURL:    "http://localhost:8545",
			StartBlock: 0,
		},
		Chains: map[string]*ChainConfig{
			"ethereum": {
				RPC: "http://localhost:8545",
				Contracts: ContractConfig{
					EpochManager: "",
					Registry:    "",
					StateManager: "",
				},
			},
		},
		Database: DatabaseConfig{
			URL:             "postgres://postgres:postgres@localhost:5432/spotted?sslmode=disable",
			MaxOpenConns:    20,
			MaxIdleConns:    5,
			ConnMaxLifetime: "1h",
		},
		P2P: P2PConfig{
			Port:           8000,
			BootstrapNodes: []string{},
			ExternalIP:     "0.0.0.0",
		},
		HTTP: HTTPConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Logging: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
} 