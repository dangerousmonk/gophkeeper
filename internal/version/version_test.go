package version

import (
	"strings"
	"testing"
)

func TestGetVersionInfo(t *testing.T) {
	// Save original values to restore later
	originalVersion := Version
	originalBuildDate := BuildDate
	originalGitCommit := GitCommit
	originalGoVersion := GoVersion

	defer func() {
		Version = originalVersion
		BuildDate = originalBuildDate
		GitCommit = originalGitCommit
		GoVersion = originalGoVersion
	}()

	tests := []struct {
		name          string
		setVersion    string
		setBuildDate  string
		setGitCommit  string
		setGoVersion  string
		expectedParts []string
	}{
		{
			name:          "default",
			setVersion:    "0.1.0",
			setBuildDate:  "unknown",
			setGitCommit:  "unknown",
			setGoVersion:  "unknown",
			expectedParts: []string{"Version: 0.1.0", "Build Date: unknown", "Git Commit: unknown", "Go Version: unknown"},
		},
		{
			name:          "real_values",
			setVersion:    "1.2.3",
			setBuildDate:  "2023-12-25T15:30:45Z",
			setGitCommit:  "abc1234",
			setGoVersion:  "go1.21.5",
			expectedParts: []string{"Version: 1.2.3", "Build Date: 2023-12-25 15:30:45 UTC", "Git Commit: abc1234", "Go Version: go1.21.5"},
		},
		{
			name:          "empty",
			setVersion:    "",
			setBuildDate:  "",
			setGitCommit:  "",
			setGoVersion:  "",
			expectedParts: []string{"Version: ", "Build Date: Unknown", "Git Commit: ", "Go Version: "},
		},
		{
			name:          "special_characters",
			setVersion:    "v1.0.0-beta+exp.sha.5114f85",
			setBuildDate:  "2023-12-25T15:30:45+03:00",
			setGitCommit:  "feat/123-new-feature",
			setGoVersion:  "go1.21.5 linux/amd64",
			expectedParts: []string{"Version: v1.0.0-beta+exp.sha.5114f85", "Git Commit: feat/123-new-feature", "Go Version: go1.21.5 linux/amd64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the test values
			Version = tt.setVersion
			BuildDate = tt.setBuildDate
			GitCommit = tt.setGitCommit
			GoVersion = tt.setGoVersion

			result := GetVersionInfo()

			for _, part := range tt.expectedParts {
				if !strings.Contains(result, part) {
					t.Errorf("GetVersionInfo() should contain %q, got: %s", part, result)
				}
			}
		})
	}
}
