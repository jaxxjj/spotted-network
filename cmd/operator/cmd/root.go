package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version   = "1.0.0"
	CommitSHA = "unknown"
	BuildTime = "unknown"

	// Global flags
	cfgFile        string
	signingKeyPath string
	signingKeyPriv string
	p2pKey         string
	password       string
	debugMode      bool

	// Docker related flags
	dockerMode  bool
	dataDir     string
	metricsAddr string
	p2pPort     int
	httpPort    int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spotted",
	Short: "Spotted Network Operator Node",
	Long: `Spotted Network Operator Node is a decentralized network participant
that helps maintain consensus and process tasks.

This application provides the following features:
- Consensus participation
- Task processing
- P2P networking
- Chain monitoring
- API endpoints

Docker Mode Options:
1. Run with Docker:
   --docker-mode --data-dir /data --metrics 0.0.0.0:4014

2. Run with volume:
   --docker-mode --data-dir /data --p2p-port 10000 --http-port 8080`,
	Version: fmt.Sprintf("%s (Build: %s, Commit: %s)", Version, BuildTime, CommitSHA),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags - available to all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is ./config/operator.yaml)")
	rootCmd.PersistentFlags().StringVar(&signingKeyPath, "signing-key-path", "",
		"path to signing keystore file (required if --signing-key-priv is not set)")
	rootCmd.PersistentFlags().StringVar(&signingKeyPriv, "signing-key-priv", "",
		"ECDSA private key in hex format (required if --signing-key-path is not set)")
	rootCmd.PersistentFlags().StringVar(&p2pKey, "p2p-key-64", "",
		"p2p key in base64 format")
	rootCmd.PersistentFlags().StringVar(&password, "password", "",
		"password for keystore (required only when using --signing-key-path)")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false,
		"enable debug mode")

	// Docker related flags
	rootCmd.PersistentFlags().BoolVar(&dockerMode, "docker-mode", false,
		"run in docker mode")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "",
		"data directory for storing files (required in docker mode)")
	rootCmd.PersistentFlags().StringVar(&metricsAddr, "metrics", "",
		"metrics listen address (e.g. 0.0.0.0:4014)")
	rootCmd.PersistentFlags().IntVar(&p2pPort, "p2p-port", 10000,
		"p2p listen port")
	rootCmd.PersistentFlags().IntVar(&httpPort, "http-port", 8080,
		"http api listen port")

	// Add version template
	rootCmd.SetVersionTemplate(`Version: {{.Version}}
`)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			fmt.Printf("Config file not found: %s\n", cfgFile)
			os.Exit(1)
		}
	} else {
		// Search for default config in ./config/operator.yaml
		defaultConfig := "./config/operator.yaml"
		if _, err := os.Stat(defaultConfig); os.IsNotExist(err) {
			fmt.Printf("Default config file not found: %s\n", defaultConfig)
			fmt.Println("Run 'operator init' to create a new configuration file")
			return
		}
		cfgFile = defaultConfig
	}

	fmt.Printf("Using config file: %s\n", cfgFile)
}
