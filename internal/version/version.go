package version

import (
	"fmt"
	"time"
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
		Version, formatBuildDate(BuildDate), GitCommit, GoVersion)
}

func formatBuildDate(dateStr string) string {
	if dateStr == "unknown" {
		return dateStr
	}

	parsedDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr // Return original if parsing fails
	}

	return parsedDate.Format("2006-01-02 15:04:05 MST")
}
