package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type ChainConfig struct {
	// RPC endpoint for the chain
	RPC string `yaml:"rpc"`

	// Contract addresses
	Contracts ContractConfig `yaml:"contracts"`

	// Number of block confirmations required before processing tasks
	RequiredConfirmations int32 `yaml:"required_confirmations"`
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
		ID             int    `yaml:"id"`
		RPCURL         string `yaml:"rpc_url"`
		RegistryAddr   string `yaml:"registry_address"`
		StartBlock     int64  `yaml:"start_block"`
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

// LoadConfig loads the configuration from the given file path
func LoadConfig(path string) (*Config, error) {
	log.Printf("Loading config from file: %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	log.Printf("Raw config data: %s", string(data))
	
	// Log all relevant environment variables
	log.Printf("Environment variables:")
	log.Printf("DATABASE_URL=%s", os.Getenv("DATABASE_URL"))
	
	// Replace environment variables in the config file
	content := os.ExpandEnv(string(data))
	log.Printf("Config after env var expansion: %s", content)

	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(content), cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// If DATABASE_URL is set in environment, override the config
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		log.Printf("Overriding database URL from environment: %s", dbURL)
		cfg.Database.URL = dbURL
	}

	log.Printf("Final loaded config: %+v", cfg)
	log.Printf("Final database config: %+v", cfg.Database)

	return cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		Chain: struct {
			ID             int    `yaml:"id"`
			RPCURL         string `yaml:"rpc_url"`
			RegistryAddr   string `yaml:"registry_address"`
			StartBlock     int64  `yaml:"start_block"`
		}{
			ID:           31337, // Anvil chain ID
			RPCURL:       "http://localhost:8545",
			RegistryAddr: "",
			StartBlock:   0,
		},
		Chains: map[string]*ChainConfig{
			"ethereum": {
				RPC: "http://localhost:8545",
				Contracts: ContractConfig{
					EpochManager: "",
					Registry:    "",
					StateManager: "",
				},
				RequiredConfirmations: 12, // Default for Ethereum
			},
		},
		Database: DatabaseConfig{
			URL:             "postgres://spotted:spotted@localhost:5432/operator1?sslmode=disable",
			MaxOpenConns:    20,
			MaxIdleConns:    5,
			ConnMaxLifetime: "1h",
		},
		P2P: P2PConfig{
			Port:           0, // Random port
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