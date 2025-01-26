package operator

import (
	"github.com/galxe/spotted-network/pkg/config"
)

// Config re-exports config.Config
type Config = config.Config

// LoadConfig loads configuration from the specified file path
func LoadConfig(path string) (*Config, error) {
	return config.LoadConfig(path)
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return config.DefaultConfig()
} 