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

Signing Key Options:
1. Use keystore file:
   --signing-key-path /path/to/keystore.json --password yourpassword

2. Use private key directly:
   --signing-key-priv 0x123...abc`,
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
