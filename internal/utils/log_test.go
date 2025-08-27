package utils

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name        string
		env         string
		wantHandler string // "text" or "json"
		wantLevel   slog.Level
	}{
		{
			name:        "local",
			env:         envLocal,
			wantHandler: "text",
			wantLevel:   slog.LevelDebug,
		},
		{
			name:        "dev",
			env:         envDev,
			wantHandler: "json",
			wantLevel:   slog.LevelDebug,
		},
		{
			name:        "prod",
			env:         envProd,
			wantHandler: "json",
			wantLevel:   slog.LevelInfo,
		},
		{
			name:        "unknown",
			env:         "staging",
			wantHandler: "json",
			wantLevel:   slog.LevelInfo,
		},
		{
			name:        "empty",
			env:         "",
			wantHandler: "json",
			wantLevel:   slog.LevelInfo,
		},
		{
			name:        "case_insensitive",
			env:         "LOCAL",
			wantHandler: "json", // Should default to JSON since case doesn't match
			wantLevel:   slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := InitLogger(tt.env, &buf)

			if logger == nil {
				t.Fatal("InitLogger() returned nil logger")
			}

			logger.Debug("debug test message")
			logger.Info("info test message")
			logger.Warn("warn test message")

			output := buf.String()

			switch tt.wantHandler {
			case "text":
				if !strings.Contains(output, "level=") || !strings.Contains(output, "msg=") {
					t.Errorf("Expected text format output, got: %s", output)
				}
			case "json":
				if !strings.Contains(output, `"level"`) || !strings.Contains(output, `"msg"`) {
					t.Errorf("Expected JSON format output, got: %s", output)
				}
			}

			// Verify log level by checking if debug messages are present
			hasDebug := strings.Contains(output, "debug test message") || strings.Contains(output, `"msg":"debug test message"`)

			if tt.wantLevel == slog.LevelDebug {
				if !hasDebug {
					t.Error("Debug level should include debug messages, but none found")
				}
			} else {
				if hasDebug {
					t.Error("Info level should not include debug messages, but debug messages found")
				}
			}

			hasInfo := strings.Contains(output, "info test message") || strings.Contains(output, `"msg":"info test message"`)
			if !hasInfo {
				t.Error("Info message should be present in output")
			}
		})
	}
}
