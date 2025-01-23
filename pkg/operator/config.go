package operator

import (
	"github.com/galxe/spotted-network/pkg/config"
)

// Re-export config types for backward compatibility
type (
	Config        = config.Config
	ChainConfig   = config.ChainConfig
	ContractConfig = config.ContractConfig
	DatabaseConfig = config.DatabaseConfig
	P2PConfig     = config.P2PConfig
	HTTPConfig    = config.HTTPConfig
	LogConfig     = config.LogConfig
)

// LoadConfig loads the configuration from the given file path
func LoadConfig(path string) (*Config, error) {
	return config.LoadConfig(path)
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return config.DefaultConfig()
} 