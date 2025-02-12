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
	cfgFile    string
	signingKey string
	p2pKey     string
	password   string
	debugMode  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "operator",
	Short: "Spotted Network Operator Node",
	Long: `Spotted Network Operator Node is a decentralized network participant
that helps maintain consensus and process tasks.

This application provides the following features:
- Consensus participation
- Task processing
- P2P networking
- Chain monitoring
- API endpoints`,
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
	rootCmd.PersistentFlags().StringVar(&signingKey, "signing-key", "",
		"path to signing keystore file")
	rootCmd.PersistentFlags().StringVar(&p2pKey, "p2p-key-64", "",
		"p2p key in base64 format")
	rootCmd.PersistentFlags().StringVar(&password, "password", "",
		"password for keystore")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false,
		"enable debug mode")

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

	// Note: Actual config loading is handled by the app package
	fmt.Printf("Using config file: %s\n", cfgFile)
}
