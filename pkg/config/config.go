package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Chains     map[uint32]*ChainConfig  `yaml:"chains"`
	Database   DatabaseConfig           `yaml:"database"`
	P2P        P2PConfig               `yaml:"p2p"`
	HTTP       HTTPConfig              `yaml:"http"`
	Logging    LoggingConfig           `yaml:"logging"`
	RegistryID string                  `yaml:"registry_id"`
	Metric     MetricConfig            `yaml:"metric"`
}

// ChainConfig represents configuration for a specific chain
type ChainConfig struct {
	RPC                  string          `yaml:"rpc"`
	Contracts           ContractsConfig  `yaml:"contracts"`
	RequiredConfirmations uint16         `yaml:"required_confirmations"`
	AverageBlockTime     float64         `yaml:"average_block_time"`
}

// ContractsConfig holds addresses for deployed contracts
type ContractsConfig struct {
	Registry     string `yaml:"registry"`
	EpochManager string `yaml:"epochManager"`
	StateManager string `yaml:"stateManager"`
}

// DatabaseConfig represents database connection configuration
type DatabaseConfig struct {
	URL             string        `yaml:"url"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// P2PConfig represents P2P network configuration
type P2PConfig struct {
	Port            int      `yaml:"port"`
	BootstrapNodes []string `yaml:"bootstrap_nodes"`
	ExternalIP     string   `yaml:"external_ip"`
}

// HTTPConfig represents HTTP server configuration
type HTTPConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// MetricConfig represents metrics server configuration
type MetricConfig struct {
	Port int    `yaml:"port"`
}

var (
	config *Config
)

// LoadConfig loads configuration from the specified file path
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Log raw config for debugging
	log.Printf("Raw config data: %s", string(data))

	// Log environment variables
	log.Printf("Environment variables:")
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "DATABASE_URL=") ||
		   strings.HasPrefix(env, "REGISTRY_ADDRESS=") ||
		   strings.HasPrefix(env, "EPOCH_MANAGER_ADDRESS=") ||
		   strings.HasPrefix(env, "STATE_MANAGER_ADDRESS=") {
			log.Printf("%s", env)
		}
	}

	// Expand environment variables in the config
	expandedData := os.ExpandEnv(string(data))
	log.Printf("Config after env var expansion: %s", expandedData)

	if err := yaml.Unmarshal([]byte(expandedData), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate the configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Log final config for debugging
	log.Printf("Final loaded config: %+v", cfg)

	return &cfg, nil
}

// validate performs validation of the configuration
func (c *Config) validate() error {
	if len(c.Chains) == 0 {
		return fmt.Errorf("at least one chain configuration is required")
	}

	for chainID, chainCfg := range c.Chains {
		if chainCfg.RPC == "" {
			return fmt.Errorf("rpc url is required for chain %d", chainID)
		}
		if chainCfg.Contracts.Registry == "" {
			return fmt.Errorf("registry address is required for chain %d", chainID)
		}
		if chainCfg.Contracts.EpochManager == "" {
			return fmt.Errorf("epoch manager address is required for chain %d", chainID)
		}
		if chainCfg.Contracts.StateManager == "" {
			return fmt.Errorf("state manager address is required for chain %d", chainID)
		}
		if chainCfg.RequiredConfirmations <= 0 {
			return fmt.Errorf("required confirmations must be positive for chain %d", chainID)
		}
		if chainCfg.AverageBlockTime <= 0 {
			return fmt.Errorf("average block time must be positive for chain %d", chainID)
		}
	}

	return nil
}

func DefaultConfig() *Config {
	return &Config{
		Chains: map[uint32]*ChainConfig{
			1: { // Ethereum mainnet
				RPC: "http://localhost:8545",
				Contracts: ContractsConfig{
					Registry:     "",
					EpochManager: "",
					StateManager: "",
				},
				RequiredConfirmations: 12,
				AverageBlockTime: 12.5,
			},
		},
		Database: DatabaseConfig{
			URL:             "postgres://spotted:spotted@localhost:5432/operator1?sslmode=disable",
			MaxOpenConns:    20,
			MaxIdleConns:    5,
			ConnMaxLifetime: 1 * time.Hour,
		},
		P2P: P2PConfig{
			Port:           0, // Random port
			BootstrapNodes: []string{},
			ExternalIP:     "0.0.0.0",
		},
		HTTP: HTTPConfig{
			Port: 8001,
			Host: "0.0.0.0",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Metric: MetricConfig{
			Port: 8080,
		},
	}
}

// GetConfig returns the singleton config instance
func GetConfig() *Config {
	if config == nil {
		panic("config not loaded")
	}
	return config
} 