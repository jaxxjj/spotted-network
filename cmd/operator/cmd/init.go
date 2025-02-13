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
	HTTPPort int    `yaml:"port"`
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

	// New field for deployment mode
	IsDockerMode bool `yaml:"is_docker_mode"`
}

// 预定义支持的链和对应的配置
var chainConfigs = map[uint32]struct {
	registryAddr          string
	epochMgrAddr          string
	stateMgrAddr          string
	requiredConfirmations uint16
	averageBlockTime      float64
}{
	31337: { // Mainnet
		registryAddr:          "0x5FbDB2315678afecb367f032d93F642f64180aa3",
		epochMgrAddr:          "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512",
		stateMgrAddr:          "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707",
		requiredConfirmations: 12,
		averageBlockTime:      12.5,
	},
	11155111: { // Sepolia
		registryAddr:          "", // 非 mainnet 不需要
		epochMgrAddr:          "", // 非 mainnet 不需要
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

		contracts := config.ContractsConfig{
			StateManager: chainInfo.stateMgrAddr,
		}

		// 只有 mainnet 需要 registry 和 epochManager
		if chainID == 31337 {
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

// 添加新函数返回默认配置
func getDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"chains": map[uint32]interface{}{
			31337: map[string]interface{}{
				"rpc": "http://localhost:8545",
				"contracts": map[string]interface{}{
					"registry":     "0x5FbDB2315678afecb367f032d93F642f64180aa3",
					"epochManager": "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512",
					"stateManager": "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707",
				},
				"required_confirmations": uint16(12),
				"average_block_time":     12.5,
			},
			11155111: map[string]interface{}{
				"rpc": "http://localhost:8545",
				"contracts": map[string]interface{}{
					"stateManager": "0x3333333333333333333333333333333333333333",
				},
				"required_confirmations": uint16(6),
				"average_block_time":     15.0,
			},
		},
		"p2p": map[string]interface{}{
			"port":            10000,
			"rendezvous":      "spotted-network", // 固定值
			"bootstrap_peers": []string{},
		},
		"http": map[string]interface{}{
			"port": 8080,
			"host": "0.0.0.0",
		},
		"logging": map[string]interface{}{
			"level":  "info",
			"format": "json",
		},
		"database": map[string]interface{}{
			"app_name":           "operator",
			"username":           "spotted",
			"password":           "spotted",
			"host":               "localhost",
			"port":               5432,
			"dbname":             "spotted",
			"max_conns":          100,
			"min_conns":          0,
			"max_conn_lifetime":  "6h",
			"max_conn_idle_time": "1m",
			"is_proxy":           false,
			"enable_prometheus":  true,
			"enable_tracing":     true,
			"replica_prefixes":   []string{},
		},
		"redis": map[string]interface{}{
			"host":                  "127.0.0.1",
			"port":                  6379,
			"password":              "",
			"is_failover":           false,
			"is_elasticache":        false,
			"is_cluster_mode":       false,
			"cluster_addrs":         []string{},
			"cluster_max_redirects": 3,
			"read_timeout":          "3s",
			"pool_size":             50,
		},
		"metric": map[string]interface{}{
			"port": 4014,
		},
	}
}

// 将函数移到文件顶部的函数定义区域
func setDockerDefaults(answers *ConfigAnswers) {
	// 数据库默认配置
	answers.DBHost = "postgres"
	answers.DBPort = 5432
	answers.DBUsername = "spotted"
	answers.DBPassword = "spotted"
	answers.DBName = "operator1"
	answers.DBMaxConns = 100
	answers.DBMinConns = 0
	answers.DBMaxConnLife = 6 * time.Hour
	answers.DBMaxConnIdle = time.Minute
	answers.DBIsProxy = false
	answers.DBEnableMetrics = true
	answers.DBEnableTracing = true
	answers.DBAppName = "operator1"

	// Redis默认配置
	answers.RedisHost = "redis"
	answers.RedisPort = 6379
	answers.RedisPassword = ""
	answers.RedisIsFailover = false
	answers.RedisIsElasticache = false
	answers.RedisIsClusterMode = false
	answers.RedisClusterMaxRedirs = 3
	answers.RedisReadTimeout = 3 * time.Second
	answers.RedisPoolSize = 50
}

func runInit(cmd *cobra.Command, args []string) error {
	var answers ConfigAnswers
	var useDefault bool

	// 添加模式选择
	modePrompt := &survey.Select{
		Message: "Choose deployment mode:",
		Options: []string{"Docker Mode", "Local Mode"},
		Default: "Docker Mode",
	}
	var mode string
	if err := survey.AskOne(modePrompt, &mode); err != nil {
		return fmt.Errorf("failed to select mode: %w", err)
	}
	answers.IsDockerMode = mode == "Docker Mode"

	// 如果是Docker模式,设置默认的database和redis配置
	if answers.IsDockerMode {
		setDockerDefaults(&answers)
	}

	// 收集链配置 - 无论是否使用default都需要
	chainConfigs, err := collectChainConfigs()
	if err != nil {
		return err
	}

	defaultPrompt := &survey.Confirm{
		Message: "Use default configuration?",
		Default: false,
	}
	if err := survey.AskOne(defaultPrompt, &useDefault); err != nil {
		return fmt.Errorf("failed to confirm default config: %w", err)
	}

	// 获取基础配置
	config := getDefaultConfig()

	// 更新chains配置
	config["chains"] = chainConfigs

	if !useDefault {
		// 只在非Docker模式下收集数据库配置
		if !answers.IsDockerMode {
			if err := survey.Ask([]*survey.Question{
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
			}, &answers); err != nil {
				return fmt.Errorf("failed to collect database config: %w", err)
			}

			// 收集Redis配置
			if err := survey.Ask([]*survey.Question{
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
			}, &answers); err != nil {
				return fmt.Errorf("failed to collect Redis config: %w", err)
			}
		}

		// 继续收集其他配置(对两种模式都需要)
		if err := survey.Ask([]*survey.Question{
			{
				Name: "P2PPort",
				Prompt: &survey.Input{
					Message: "Enter P2P Port:",
					Default: "10000",
				},
				Validate: survey.Required,
			},
			{
				Name: "HTTPPort",
				Prompt: &survey.Input{
					Message: "Enter HTTP Port:",
					Default: "8080",
				},
			},
		}, &answers); err != nil {
			return fmt.Errorf("failed to collect config: %w", err)
		}

		// 更新配置
		config["p2p"] = map[string]interface{}{
			"port":            answers.P2PPort,
			"rendezvous":      "spotted-network", // 固定值
			"bootstrap_peers": answers.BootstrapPeers,
		}
		config["http"] = map[string]interface{}{
			"port": answers.HTTPPort,
			"host": "0.0.0.0",
		}
		config["database"] = map[string]interface{}{
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
			"replica_prefixes":   []string{},
		}
		config["redis"] = map[string]interface{}{
			"host":                  answers.RedisHost,
			"port":                  answers.RedisPort,
			"password":              answers.RedisPassword,
			"is_failover":           false,
			"is_elasticache":        false,
			"is_cluster_mode":       answers.RedisIsClusterMode,
			"cluster_addrs":         []string{},
			"cluster_max_redirects": 3,
			"read_timeout":          "3s",
			"pool_size":             50,
		}
		config["metric"] = map[string]interface{}{
			"port": "${METRIC_PORT:-4014}",
		}
	}

	// 处理bootstrap peers
	var isFirstNode bool
	isFirstNodePrompt := &survey.Confirm{
		Message: "Is this the first node in the network?",
		Default: false,
	}
	if err := survey.AskOne(isFirstNodePrompt, &isFirstNode); err != nil {
		return fmt.Errorf("failed to confirm first node status: %w", err)
	}

	if !isFirstNode {
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

		p2pConfig := config["p2p"].(map[string]interface{})
		p2pConfig["bootstrap_peers"] = bootstrapPeers
	}

	// 创建配置目录
	if configPath == "" {
		configPath = "./config/operator.yaml"
	}
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 写入配置文件
	configData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 显示成功信息和下一步操作
	fmt.Println("\nConfiguration initialized successfully!")
	if answers.IsDockerMode {
		fmt.Println("\nNext steps:")
		fmt.Println("1. Review the configuration in config/operator.yaml")
		fmt.Println("2. Start the services:")
		fmt.Println("   cd ~/.spotted && docker-compose up -d")
		fmt.Println("3. Check service status:")
		fmt.Println("   docker-compose ps")
	} else {
		fmt.Println("\nNext steps:")
		fmt.Println("1. Review the configuration in config/operator.yaml")
		fmt.Println("2. Start the operator:")
		fmt.Println("   spotted start [flags]")
	}

	return nil
}
