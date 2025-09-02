package version

import (
	"fmt"

	"github.com/dangerousmonk/gophkeeper/internal/utils"
)

var (
	Version   = "0.1.0"
	BuildDate = "unknown"
	GitCommit = "unknown"
	GoVersion = "unknown"
)

// GetVersionInfo returns formatted information about TUI client
func GetVersionInfo() string {
	return fmt.Sprintf("Version: %s\nBuild Date: %s\nGit Commit: %s\nGo Version: %s",
		Version, utils.FormatDate(BuildDate), GitCommit, GoVersion)
}
