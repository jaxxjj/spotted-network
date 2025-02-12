package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/galxe/spotted-network/cmd/operator/app"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the operator node",
	Long: `Start the operator node with the specified configuration.
	
This command will:
1. Load configuration from the specified file
2. Initialize all required components
3. Start the operator node
4. Handle graceful shutdown on interrupt`,
	PreRunE: validateStartFlags,
	RunE:    runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

// validateStartFlags checks if all required flags are provided
func validateStartFlags(cmd *cobra.Command, args []string) error {
	// Check if config file exists
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s. Run 'operator init' first", cfgFile)
	}

	// Check required flags
	if signingKey == "" {
		return fmt.Errorf("signing-key is required")
	}
	if p2pKey == "" {
		return fmt.Errorf("p2p-key-64 is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Check if signing key file exists
	if _, err := os.Stat(signingKey); os.IsNotExist(err) {
		return fmt.Errorf("signing key file not found: %s", signingKey)
	}

	return nil
}

func runStart(cmd *cobra.Command, args []string) error {
	// Create signal channel for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create app instance
	application := app.New(cmd.Context())

	// Start application in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := application.Run(
			signingKey,
			p2pKey,
			password,
		); err != nil {
			errChan <- fmt.Errorf("application error: %w", err)
		}
	}()

	// Wait for interrupt or error
	select {
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal. Shutting down...")
		// Trigger graceful shutdown
		if err := application.Shutdown(); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}
