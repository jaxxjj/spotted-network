package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

// configPath is the path to the config file
var configPath string

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path (default is ./config/operator.yaml)")
}

// configAnswers holds all configuration answers
type ConfigAnswers struct {
	DeployMode string `yaml:"deploy_mode"` // "docker" or "docker-compose"

	ChainConfigs map[uint32]*config.ChainConfig `yaml:"chains"`

	// node configuration
	IsFirstNode    bool     `yaml:"is_first_node"`
	BootstrapPeers []string `yaml:"bootstrap_peers,omitempty"`

	// basic connection configuration for docker mode
	DBHost     string `yaml:"db_host,omitempty"`
	DBPort     int    `yaml:"db_port,omitempty"`
	DBUser     string `yaml:"db_user,omitempty"`
	DBPassword string `yaml:"db_password,omitempty"`
	DBName     string `yaml:"db_name,omitempty"`

	RedisHost     string `yaml:"redis_host,omitempty"`
	RedisPort     int    `yaml:"redis_port,omitempty"`
	RedisPassword string `yaml:"redis_password,omitempty"`
}

// chainConfigs defines supported chains and their default configurations
var chainConfigs = map[uint32]struct {
	rpcURL                string
	registryAddr          string
	epochMgrAddr          string
	stateMgrAddr          string
	requiredConfirmations uint16
	averageBlockTime      float64
}{
	11155111: { // Sepolia mainnet
		rpcURL:                "wss://ethereum-sepolia-rpc.publicnode.com",
		registryAddr:          "0xB6dE44d8F1425752CAc1103D99e59eD329F65aCF",
		epochMgrAddr:          "0x5bFB7609a51F8577D90e8576DE6e85BC7fBf08F7",
		stateMgrAddr:          "0xcc6Db3c0389128bad36796079aB336B3AfC1cF19",
		requiredConfirmations: 2,
		averageBlockTime:      12.0,
	},
	84532: { // Base Sepolia
		rpcURL:                "https://base-sepolia-rpc.publicnode.com",
		registryAddr:          "",
		epochMgrAddr:          "",
		stateMgrAddr:          "0xe8Cbc41961125A1B0F86465Ff9a6666e39104E9e",
		requiredConfirmations: 2,
		averageBlockTime:      2.0,
	},
	421614: { // Arbitrum Sepolia
		rpcURL:                "https://arbitrum-sepolia.gateway.tenderly.co",
		registryAddr:          "",
		epochMgrAddr:          "",
		stateMgrAddr:          "0xe3Ed30610de2914b45d848718d6837eF14361C41",
		requiredConfirmations: 2,
		averageBlockTime:      2.0,
	},
}

func collectChainConfigs() (map[uint32]*config.ChainConfig, error) {
	configs := make(map[uint32]*config.ChainConfig)

	for chainID, chainInfo := range chainConfigs {
		var rpcURL string
		prompt := &survey.Input{
			Message: fmt.Sprintf("Enter RPC URL for Chain ID %d:", chainID),
			Default: chainInfo.rpcURL,
			Help:    "The RPC endpoint for this blockchain node",
		}

		if err := survey.AskOne(prompt, &rpcURL); err != nil {
			return nil, err
		}

		contracts := config.ContractsConfig{
			StateManager: chainInfo.stateMgrAddr,
		}

		// only mainnet need registry and epochManager
		if chainID == 11155111 {
			contracts.Registry = chainInfo.registryAddr
			contracts.EpochManager = chainInfo.epochMgrAddr
		}

		configs[chainID] = &config.ChainConfig{
			RPC:                   rpcURL,
			Contracts:             contracts,
			RequiredConfirmations: chainInfo.requiredConfirmations,
			AverageBlockTime:      chainInfo.averageBlockTime,
		}
	}

	return configs, nil
}

func runInit(cmd *cobra.Command, args []string) error {
	answers := &ConfigAnswers{}

	// 1. collect chain configurations
	chainConfigs, err := collectChainConfigs()
	if err != nil {
		return fmt.Errorf("failed to collect chain configs: %w", err)
	}
	answers.ChainConfigs = chainConfigs

	// 2. first node configuration
	isFirstNodePrompt := &survey.Confirm{
		Message: "Is this the first node in the network?",
		Default: false,
	}
	if err := survey.AskOne(isFirstNodePrompt, &answers.IsFirstNode); err != nil {
		return fmt.Errorf("failed to confirm first node status: %w", err)
	}

	// 3. if not first node, collect bootstrap peers
	if !answers.IsFirstNode {
		bootstrapPeers := []string{}
		continueAdding := true

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

	// 4. choose deployment mode
	modePrompt := &survey.Select{
		Message: "Choose deployment mode:",
		Options: []string{"docker", "docker-compose"},
		Default: "docker-compose",
	}
	if err := survey.AskOne(modePrompt, &answers.DeployMode); err != nil {
		return fmt.Errorf("failed to get deployment mode: %w", err)
	}

	// 5. if docker mode, collect external service configuration
	if answers.DeployMode == "docker" {
		// database configuration
		dbQuestions := []*survey.Question{
			{
				Name: "DBHost",
				Prompt: &survey.Input{
					Message: "Enter PostgreSQL host:",
					Default: "localhost",
				},
			},
			{
				Name: "DBPort",
				Prompt: &survey.Input{
					Message: "Enter PostgreSQL port:",
					Default: "5432",
				},
			},
			{
				Name: "DBUser",
				Prompt: &survey.Input{
					Message: "Enter PostgreSQL user:",
					Default: "spotted",
				},
			},
			{
				Name: "DBPassword",
				Prompt: &survey.Password{
					Message: "Enter PostgreSQL password:",
				},
			},
			{
				Name: "DBName",
				Prompt: &survey.Input{
					Message: "Enter PostgreSQL database name:",
					Default: "spotted",
				},
			},
		}
		if err := survey.Ask(dbQuestions, answers); err != nil {
			return fmt.Errorf("failed to get database config: %w", err)
		}

		// Redis configuration
		redisQuestions := []*survey.Question{
			{
				Name: "RedisHost",
				Prompt: &survey.Input{
					Message: "Enter Redis host:",
					Default: "localhost",
				},
			},
			{
				Name: "RedisPort",
				Prompt: &survey.Input{
					Message: "Enter Redis port:",
					Default: "6379",
				},
			},
			{
				Name: "RedisPassword",
				Prompt: &survey.Password{
					Message: "Enter Redis password (optional):",
				},
			},
		}
		if err := survey.Ask(redisQuestions, answers); err != nil {
			return fmt.Errorf("failed to get redis config: %w", err)
		}
	}

	// 6. generate final configuration
	config := generateConfig(answers)

	// 7. create config directory
	configDir := "config"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 8. specify config file path
	configPath := filepath.Join(configDir, "operator.yaml")

	// 9. write config to file
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration has been written to %s\n", configPath)
	return nil
}

func generateConfig(answers *ConfigAnswers) map[string]interface{} {
	config := make(map[string]interface{})

	// 1. chain configurations (not dependent on deployment mode)
	config["chains"] = answers.ChainConfigs

	// 2. set database and redis configuration based on deployment mode
	if answers.DeployMode == "docker-compose" {
		// docker-compose mode uses default configuration
		config["database"] = map[string]interface{}{
			"host":           "postgres",
			"port":           5432,
			"username":       "spotted",
			"password":       "spotted",
			"dbname":         "spotted",
			"max_conns":      100,
			"min_conns":      0,
			"max_conn_life":  "6h",
			"max_conn_idle":  "1m",
			"is_proxy":       false,
			"enable_metrics": true,
			"enable_tracing": true,
			"app_name":       "spotted",
		}

		config["redis"] = map[string]interface{}{
			"host":                  "redis",
			"port":                  6379,
			"password":              "",
			"is_failover":           false,
			"is_elasticache":        false,
			"is_cluster_mode":       false,
			"cluster_max_redirects": 3,
			"read_timeout":          "3s",
			"pool_size":             50,
		}
	} else {
		// docker mode uses user-provided configuration
		config["database"] = map[string]interface{}{
			"host":     answers.DBHost,
			"port":     answers.DBPort,
			"username": answers.DBUser,
			"password": answers.DBPassword,
			"dbname":   answers.DBName,
		}

		config["redis"] = map[string]interface{}{
			"host":     answers.RedisHost,
			"port":     answers.RedisPort,
			"password": answers.RedisPassword,
		}
	}

	// 3. hardcoded port configuration
	config["metric"] = map[string]interface{}{
		"port": 4014,
	}

	config["http"] = map[string]interface{}{
		"port": 8080,
	}

	config["p2p"] = map[string]interface{}{
		"port":            10000,
		"rendezvous":      "spotted-network",
		"bootstrap_peers": answers.BootstrapPeers,
	}

	return config
}
