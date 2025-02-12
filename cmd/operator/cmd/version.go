package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Print detailed version information about the operator node.
This includes version number, build time, commit hash, and Go version.`,
	Run: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf(`Spotted Network Operator
Version:    %s
Build Time: %s
Git Commit: %s
Go Version: %s
OS/Arch:    %s/%s
`,
		Version,
		BuildTime,
		CommitSHA,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}
