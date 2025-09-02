package version

import (
	"fmt"

	"github.com/dangerousmonk/gophkeeper/internal/dates"
)

var (
	version   = "0.1.0"
	buildDate = "unknown"
	gitCommit = "unknown"
	goVersion = "unknown"
)

// GetVersionInfo returns formatted information about TUI client
func GetVersionInfo() string {
	return fmt.Sprintf("Version: %s\nBuild Date: %s\nGit Commit: %s\nGo Version: %s",
		version, dates.FormatDate(buildDate), gitCommit, goVersion)
}
