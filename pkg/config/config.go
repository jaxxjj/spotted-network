package config

import (
	"fmt"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Chains   map[uint32]*ChainConfig `yaml:"chains"`
	Database DatabaseConfig          `yaml:"database"`
	P2P      P2PConfig               `yaml:"p2p"`
	HTTP     HTTPConfig              `yaml:"http"`
	Logging  LoggingConfig           `yaml:"logging"`
	Metric   MetricConfig            `yaml:"metric"`
	Redis    RedisConfig             `yaml:"redis"`
}

// ChainConfig represents configuration for a specific chain
type ChainConfig struct {
	RPC                   string          `yaml:"rpc"`
	Contracts             ContractsConfig `yaml:"contracts"`
	RequiredConfirmations uint16          `yaml:"required_confirmations"`
	AverageBlockTime      float64         `yaml:"average_block_time"`
}

// ContractsConfig holds addresses for deployed contracts
type ContractsConfig struct {
	Registry     string `yaml:"registry"`
	EpochManager string `yaml:"epochManager"`
	StateManager string `yaml:"stateManager"`
}

// DatabaseConfig represents database connection configuration
type DatabaseConfig struct {
	AppName          string        `yaml:"app_name"`
	Username         string        `yaml:"username"`
	Password         string        `yaml:"password"`
	Host             string        `yaml:"host"`
	Port             int           `yaml:"port"`
	DBName           string        `yaml:"dbname"`
	MaxConns         int           `yaml:"max_conns"`
	MinConns         int           `yaml:"min_conns"`
	MaxConnLifetime  time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime  time.Duration `yaml:"max_conn_idle_time"`
	IsProxy          bool          `yaml:"is_proxy"`
	EnablePrometheus bool          `yaml:"enable_prometheus"`
	EnableTracing    bool          `yaml:"enable_tracing"`
	ReplicaPrefixes  []string      `yaml:"replica_prefixes"`
}

// P2PConfig represents P2P network configuration
type P2PConfig struct {
	Port           int      `yaml:"port"`
	BootstrapPeers []string `yaml:"bootstrap_peers"`
	Rendezvous     string   `yaml:"rendezvous"`
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
	Port int `yaml:"port"`
}

// RedisConfig represents Redis connection configuration
type RedisConfig struct {
	Host                string        `yaml:"host"`
	Port                int           `yaml:"port"`
	Password            string        `yaml:"password"`
	IsFailover          bool          `yaml:"is_failover"`
	IsElasticache       bool          `yaml:"is_elasticache"`
	IsClusterMode       bool          `yaml:"is_cluster_mode"`
	ClusterAddrs        []string      `yaml:"cluster_addrs"`
	ClusterMaxRedirects int           `yaml:"cluster_max_redirects"`
	ReadTimeout         time.Duration `yaml:"read_timeout"`
	PoolSize            int           `yaml:"pool_size"`
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

	// Expand environment variables in the config
	expandedData := os.ExpandEnv(string(data))

	if err := yaml.Unmarshal([]byte(expandedData), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate the configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// validate performs validation of the configuration
func (c *Config) validate() error {
	if len(c.Chains) == 0 {
		return fmt.Errorf("at least one chain configuration is required")
	}

	for chainID, chainCfg := range c.Chains {
		if chainID == 1 {
			if chainCfg.Contracts.Registry == "" {
				return fmt.Errorf("registry address is required for mainnet")
			}
			if chainCfg.Contracts.EpochManager == "" {
				return fmt.Errorf("epoch manager address is required for mainnet")
			}
		}
		if chainCfg.RPC == "" {
			return fmt.Errorf("rpc url is required for chain %d", chainID)
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

	// Validate P2P configuration
	if c.P2P.Port < 0 {
		return fmt.Errorf("p2p port must be non-negative")
	}
	if c.P2P.Rendezvous == "" {
		return fmt.Errorf("p2p rendezvous string is required")
	}

	// Validate bootstrap peers if provided
	if len(c.P2P.BootstrapPeers) > 0 {
		if _, err := c.P2P.GetBootstrapPeers(); err != nil {
			return fmt.Errorf("invalid bootstrap peers: %w", err)
		}
	}

	return nil
}

// GetConfig returns the singleton config instance
func GetConfig() *Config {
	if config == nil {
		panic("config not loaded")
	}
	return config
}

// GetBootstrapPeers returns a list of valid bootstrap peer AddrInfo
func (c *P2PConfig) GetBootstrapPeers() ([]peer.AddrInfo, error) {
	var peers []peer.AddrInfo

	for _, addr := range c.BootstrapPeers {
		// Skip empty addresses
		if addr == "" {
			continue
		}

		// Parse the multiaddr
		pi, err := peer.AddrInfoFromString(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap peer address %s: %w", addr, err)
		}
		peers = append(peers, *pi)
	}

	return peers, nil
}
