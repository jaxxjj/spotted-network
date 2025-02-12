package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize operator configuration",
	Long: `Initialize operator configuration with interactive prompts.
Default values will be used if you press enter without input.
The configuration will be saved to config/operator.yaml.`,
	RunE: runInit,
}

// 添加配置文件路径的flag
var configPath string

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path (default is ./config/operator.yaml)")
}

// ConfigAnswers holds all configuration answers
type ConfigAnswers struct {
	// Chains Configuration
	ChainRPC              string  `yaml:"rpc"`
	RegistryContract      string  `yaml:"registry"`
	EpochManagerContract  string  `yaml:"epoch_manager"`
	StateManagerContract  string  `yaml:"state_manager"`
	RequiredConfirmations int     `yaml:"required_confirmations"`
	AverageBlockTime      float64 `yaml:"average_block_time"`

	// P2P Configuration
	P2PPort        int      `yaml:"port"`
	P2PRendezvous  string   `yaml:"rendezvous"`
	BootstrapPeers []string `yaml:"bootstrap_peers"`

	// HTTP Configuration
	HTTPPort string `yaml:"port"`
	HTTPHost string `yaml:"host"`

	// Database Configuration
	DBUsername      string        `yaml:"username"`
	DBPassword      string        `yaml:"password"`
	DBHost          string        `yaml:"host"`
	DBPort          int           `yaml:"port"`
	DBName          string        `yaml:"dbname"`
	DBMaxConns      int           `yaml:"max_conns"`
	DBMinConns      int           `yaml:"min_conns"`
	DBMaxConnLife   time.Duration `yaml:"max_conn_lifetime"`
	DBMaxConnIdle   time.Duration `yaml:"max_conn_idle_time"`
	DBIsProxy       bool          `yaml:"is_proxy"`
	DBEnableMetrics bool          `yaml:"enable_prometheus"`
	DBEnableTracing bool          `yaml:"enable_tracing"`
	DBAppName       string        `yaml:"app_name"`

	// Redis Configuration
	RedisHost             string        `yaml:"host"`
	RedisPort             int           `yaml:"port"`
	RedisPassword         string        `yaml:"password"`
	RedisIsFailover       bool          `yaml:"is_failover"`
	RedisIsElasticache    bool          `yaml:"is_elasticache"`
	RedisIsClusterMode    bool          `yaml:"is_cluster_mode"`
	RedisClusterAddrs     []string      `yaml:"cluster_addrs"`
	RedisClusterMaxRedirs int           `yaml:"cluster_max_redirects"`
	RedisReadTimeout      time.Duration `yaml:"read_timeout"`
	RedisPoolSize         int           `yaml:"pool_size"`

	// Metric Configuration
	MetricPort int `yaml:"port"`

	// New fields for P2P Configuration
	IsFirstNode bool `yaml:"is_first_node"`
}

// 预定义支持的链和对应的配置
var chainConfigs = map[uint32]struct {
	registryAddr          string
	epochMgrAddr          string
	stateMgrAddr          string
	requiredConfirmations uint16
	averageBlockTime      float64
}{
	31337: {
		registryAddr:          "0x5FbDB2315678afecb367f032d93F642f64180aa3",
		epochMgrAddr:          "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512",
		stateMgrAddr:          "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707",
		requiredConfirmations: 12,
		averageBlockTime:      12.5,
	},
	11155111: {
		registryAddr:          "0x1111111111111111111111111111111111111111",
		epochMgrAddr:          "0x2222222222222222222222222222222222222222",
		stateMgrAddr:          "0x3333333333333333333333333333333333333333",
		requiredConfirmations: 6,
		averageBlockTime:      15.0,
	},
	// 可以添加更多链的配置
}

func collectChainConfigs() (map[uint32]*config.ChainConfig, error) {
	configs := make(map[uint32]*config.ChainConfig)

	for chainID, chainInfo := range chainConfigs {
		var rpcURL string
		prompt := &survey.Input{
			Message: fmt.Sprintf("Enter RPC URL for Chain ID %d:", chainID),
			Default: "http://localhost:8545",
			Help:    "The RPC endpoint for this blockchain node",
		}

		if err := survey.AskOne(prompt, &rpcURL); err != nil {
			return nil, err
		}

		configs[chainID] = &config.ChainConfig{
			RPC: rpcURL,
			Contracts: config.ContractsConfig{
				Registry:     chainInfo.registryAddr,
				EpochManager: chainInfo.epochMgrAddr,
				StateManager: chainInfo.stateMgrAddr,
			},
			RequiredConfirmations: chainInfo.requiredConfirmations,
			AverageBlockTime:      chainInfo.averageBlockTime,
		}
	}

	return configs, nil
}

func runInit(cmd *cobra.Command, args []string) error {
	// 如果没有指定配置文件路径,使用默认路径
	if configPath == "" {
		configPath = filepath.Join("config", "operator.yaml")
	}

	// 检查配置文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		// 文件已存在,询问是否覆盖
		var overwrite bool
		overwritePrompt := &survey.Confirm{
			Message: fmt.Sprintf("Config file %s already exists. Do you want to overwrite it?", configPath),
			Default: false,
		}
		if err := survey.AskOne(overwritePrompt, &overwrite); err != nil {
			return fmt.Errorf("failed to confirm overwrite: %w", err)
		}
		if !overwrite {
			return fmt.Errorf("aborted: config file already exists")
		}
	}

	answers := &ConfigAnswers{}

	// 获取链配置
	chainConfigs, err := collectChainConfigs()
	if err != nil {
		return fmt.Errorf("failed to collect chain configurations: %w", err)
	}

	// P2P Configuration Questions
	p2pQuestions := []*survey.Question{
		{
			Name: "P2PPort",
			Prompt: &survey.Input{
				Message: "Enter P2P Port:",
				Default: "10000",
			},
			Validate: survey.Required,
		},
		{
			Name: "P2PRendezvous",
			Prompt: &survey.Input{
				Message: "Enter P2P Rendezvous String:",
				Default: "spotted-network",
			},
		},
		{
			Name: "IsFirstNode",
			Prompt: &survey.Confirm{
				Message: "Is this the first node in the network?",
				Default: false,
			},
		},
	}

	// Database Configuration Questions
	dbQuestions := []*survey.Question{
		{
			Name: "DBHost",
			Prompt: &survey.Input{
				Message: "Enter Database Host:",
				Default: "localhost",
			},
		},
		{
			Name: "DBPort",
			Prompt: &survey.Input{
				Message: "Enter Database Port:",
				Default: "5432",
			},
		},
		{
			Name: "DBUsername",
			Prompt: &survey.Input{
				Message: "Enter Database Username:",
				Default: "postgres",
			},
		},
		{
			Name: "DBPassword",
			Prompt: &survey.Password{
				Message: "Enter Database Password:",
			},
		},
		{
			Name: "DBName",
			Prompt: &survey.Input{
				Message: "Enter Database Name:",
				Default: "spotted",
			},
		},
		{
			Name: "DBIsProxy",
			Prompt: &survey.Confirm{
				Message: "Is this a proxy connection?",
				Default: false,
			},
		},
	}

	// Redis Configuration Questions
	redisQuestions := []*survey.Question{
		{
			Name: "RedisHost",
			Prompt: &survey.Input{
				Message: "Enter Redis Host:",
				Default: "127.0.0.1",
			},
		},
		{
			Name: "RedisPort",
			Prompt: &survey.Input{
				Message: "Enter Redis Port:",
				Default: "6379",
			},
		},
		{
			Name: "RedisPassword",
			Prompt: &survey.Password{
				Message: "Enter Redis Password (optional):",
			},
		},
		{
			Name: "RedisIsClusterMode",
			Prompt: &survey.Confirm{
				Message: "Enable Redis Cluster Mode?",
				Default: false,
			},
		},
	}

	// Ask questions by category
	fmt.Println("\n=== P2P Configuration ===")
	if err := survey.Ask(p2pQuestions, answers); err != nil {
		return fmt.Errorf("failed to get P2P configuration: %w", err)
	}

	// Handle bootstrap peers based on whether it's first node
	if !answers.IsFirstNode {
		bootstrapPeers := []string{}
		continueAdding := true

		// Must have at least one bootstrap peer
		fmt.Println("\n=== Bootstrap Peers Configuration ===")
		fmt.Println("You need at least one bootstrap peer (max 5)")

		for len(bootstrapPeers) < 5 && continueAdding {
			var peer string
			peerPrompt := &survey.Input{
				Message: fmt.Sprintf("Enter bootstrap peer #%d:", len(bootstrapPeers)+1),
				Help:    "Format: /ip4/1.2.3.4/tcp/4001/p2p/QmPeerID123",
			}

			if err := survey.AskOne(peerPrompt, &peer); err != nil {
				return fmt.Errorf("failed to get bootstrap peer: %w", err)
			}

			if peer != "" {
				bootstrapPeers = append(bootstrapPeers, peer)

				// If we have at least one peer and less than 5, ask if want to add more
				if len(bootstrapPeers) < 5 {
					addMore := false
					addMorePrompt := &survey.Confirm{
						Message: "Do you want to add another bootstrap peer?",
						Default: false,
					}

					if err := survey.AskOne(addMorePrompt, &addMore); err != nil {
						return fmt.Errorf("failed to confirm adding more peers: %w", err)
					}

					continueAdding = addMore
				}
			}
		}

		if len(bootstrapPeers) == 0 {
			return fmt.Errorf("at least one bootstrap peer is required for non-first nodes")
		}

		answers.BootstrapPeers = bootstrapPeers
	}

	fmt.Println("\n=== Database Configuration ===")
	if err := survey.Ask(dbQuestions, answers); err != nil {
		return fmt.Errorf("failed to get database configuration: %w", err)
	}

	fmt.Println("\n=== Redis Configuration ===")
	if err := survey.Ask(redisQuestions, answers); err != nil {
		return fmt.Errorf("failed to get redis configuration: %w", err)
	}

	// Convert answers to configuration structure
	config := map[string]interface{}{
		"chains": chainConfigs,
		"p2p": map[string]interface{}{
			"port":            answers.P2PPort,
			"rendezvous":      answers.P2PRendezvous,
			"bootstrap_peers": answers.BootstrapPeers,
		},
		"http": map[string]interface{}{
			"port": 8080,
			"host": "0.0.0.0",
		},
		"database": map[string]interface{}{
			"username":           answers.DBUsername,
			"password":           answers.DBPassword,
			"host":               answers.DBHost,
			"port":               answers.DBPort,
			"dbname":             answers.DBName,
			"max_conns":          100,
			"min_conns":          0,
			"max_conn_lifetime":  "6h",
			"max_conn_idle_time": "1m",
			"is_proxy":           answers.DBIsProxy,
			"enable_prometheus":  true,
			"enable_tracing":     true,
			"app_name":           "operator",
		},
		"redis": map[string]interface{}{
			"host":                  answers.RedisHost,
			"port":                  answers.RedisPort,
			"password":              answers.RedisPassword,
			"is_cluster_mode":       answers.RedisIsClusterMode,
			"cluster_max_redirects": 3,
			"read_timeout":          "3s",
			"pool_size":             50,
		},
		"metric": map[string]interface{}{
			"port": 4014,
		},
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal configuration to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write configuration file
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("\nConfiguration has been saved to %s\n", configPath)
	fmt.Println("You can now start the operator with:")
	if configPath != "./config/operator.yaml" {
		fmt.Printf("  operator start --config %s\n", configPath)
	} else {
		fmt.Println("  operator start")
	}

	return nil
}
