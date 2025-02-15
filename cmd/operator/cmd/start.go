package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/galxe/spotted-network/cmd/operator/app"
	"github.com/galxe/spotted-network/pkg/config"
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
		return fmt.Errorf("config file not found: %s. Run 'spotted init' first", cfgFile)
	}

	// Validate signing key options
	if signingKeyPath == "" && signingKeyPriv == "" {
		return fmt.Errorf("either --signing-key-path or --signing-key-priv must be provided")
	}
	if signingKeyPath != "" && signingKeyPriv != "" {
		return fmt.Errorf("cannot use both --signing-key-path and --signing-key-priv at the same time")
	}

	// If using keystore file, validate password and file existence
	if signingKeyPath != "" {
		if password == "" {
			return fmt.Errorf("--password is required when using --signing-key-path")
		}
		if _, err := os.Stat(signingKeyPath); os.IsNotExist(err) {
			return fmt.Errorf("signing key file not found: %s", signingKeyPath)
		}
	}

	// Validate p2p key
	if p2pKey == "" {
		return fmt.Errorf("p2p-key-64 is required")
	}

	return nil
}

func runStart(cmd *cobra.Command, args []string) error {
	// load config file
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// set database environment variables
	dbCfg := cfg.Database
	os.Setenv("POSTGRES_APPNAME", dbCfg.AppName)
	os.Setenv("POSTGRES_USERNAME", dbCfg.Username)
	os.Setenv("POSTGRES_PASSWORD", dbCfg.Password)
	os.Setenv("POSTGRES_HOST", dbCfg.Host)
	os.Setenv("POSTGRES_PORT", strconv.Itoa(dbCfg.Port))
	os.Setenv("POSTGRES_DBNAME", dbCfg.DBName)
	os.Setenv("POSTGRES_MAXCONNS", strconv.Itoa(dbCfg.MaxConns))
	os.Setenv("POSTGRES_MINCONNS", strconv.Itoa(dbCfg.MinConns))
	os.Setenv("POSTGRES_MAXCONNLIFETIME", dbCfg.MaxConnLifetime.String())
	os.Setenv("POSTGRES_MAXCONNIDLETIME", dbCfg.MaxConnIdleTime.String())
	os.Setenv("POSTGRES_ISPROXY", strconv.FormatBool(dbCfg.IsProxy))
	os.Setenv("POSTGRES_ENABLEPROMETHEUS", strconv.FormatBool(dbCfg.EnablePrometheus))
	os.Setenv("POSTGRES_ENABLETRACING", strconv.FormatBool(dbCfg.EnableTracing))

	// set redis environment variables
	redisCfg := cfg.Redis
	os.Setenv("REDIS_HOST", redisCfg.Host)
	os.Setenv("REDIS_PORT", strconv.Itoa(redisCfg.Port))
	os.Setenv("REDIS_PASSWORD", redisCfg.Password)
	os.Setenv("REDIS_IS_FAILOVER", strconv.FormatBool(redisCfg.IsFailover))
	os.Setenv("REDIS_IS_ELASTICACHE", strconv.FormatBool(redisCfg.IsElasticache))
	os.Setenv("REDIS_IS_CLUSTER_MODE", strconv.FormatBool(redisCfg.IsClusterMode))
	os.Setenv("REDIS_CLUSTER_MAX_REDIRECTS", strconv.Itoa(redisCfg.ClusterMaxRedirects))
	os.Setenv("REDIS_READ_TIMEOUT", redisCfg.ReadTimeout.String())
	os.Setenv("REDIS_POOL_SIZE", strconv.Itoa(redisCfg.PoolSize))

	os.Setenv("METRIC_PORT", strconv.Itoa(cfg.Metric.Port))

	// create signal channel for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// create app instance
	application := app.New(cmd.Context())

	// start app in a goroutine
	errChan := make(chan error, 1)
	go func() {
		var err error
		isKeyPath := signingKeyPath != ""
		signingKey := signingKeyPriv
		if isKeyPath {
			signingKey = signingKeyPath
		}
		err = application.Run(
			isKeyPath,
			signingKey,
			p2pKey,
			password,
		)
		if err != nil {
			errChan <- fmt.Errorf("application error: %w", err)
		}
	}()

	// wait for interrupt signal or error
	select {
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal. Shutting down...")
		if err := application.Shutdown(); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}
