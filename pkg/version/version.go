package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of spotted network
	Version = "v0.1.0"

	// GitCommit is the git commit hash of the build
	GitCommit = "unknown"

	// BuildTime is the time when the binary was built
	BuildTime = "unknown"
)

// Info returns version information
func Info() string {
	return fmt.Sprintf("Version: %s\nGit Commit: %s\nBuild Time: %s\nGo Version: %s\nOS/Arch: %s/%s",
		Version,
		GitCommit,
		BuildTime,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}
