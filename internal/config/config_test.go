package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		configFile     string
		envVars        map[string]string
		flags          []string
		expectedConfig *Config
		expectError    bool
	}{
		{
			name:       "default_values_only",
			configFile: "",
			envVars:    nil,
			flags:      nil,
			expectedConfig: &Config{
				JWTSecret: "default-secret",
				Server: ServerConfig{
					Host: "localhost",
					Port: "8099",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Name:     "gophkeeper_db",
					User:     "gophkeeper",
					Password: "password",
					Port:     "5432",
					SSLMode:  "disable",
				},
				Environment: "local",
			},
			expectError: false,
		},
		{
			name:       "env_variables_override_defaults",
			configFile: "",
			envVars: map[string]string{
				"SERVER_HOST": "http://some.com",
				"SERVER_PORT": "8081",
				"DB_USER":     "env-user",
				"DB_PASSWORD": "env-pass",
				"DB_SSL_MODE": "enable",
				"JWT_SECRET":  "env-secret",
			},
			flags: nil,
			expectedConfig: &Config{
				JWTSecret: "env-secret",
				Server: ServerConfig{
					Host: "http://some.com",
					Port: "8081",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Name:     "gophkeeper_db",
					User:     "env-user",
					Password: "env-pass",
					Port:     "5432",
					SSLMode:  "enable",
				},
				Environment: "local",
			},
			expectError: false,
		},
		{
			name:       "flags_override_env_and_defaults",
			configFile: "",
			envVars: map[string]string{
				"DB_USER": "env-user",
			},
			flags: []string{"--db_user", "flag-user"},
			expectedConfig: &Config{
				JWTSecret: "default-secret",
				Server: ServerConfig{
					Host: "localhost",
					Port: "8099",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Name:     "gophkeeper_db",
					User:     "flag-user",
					Password: "password",
					Port:     "5432",
					SSLMode:  "disable",
				},
				Environment: "local",
			},
			expectError: false,
		},
		{
			name:           "invalid_config_file_path",
			configFile:     "nonexistent.yaml",
			envVars:        nil,
			flags:          nil,
			expectedConfig: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			// Reset flags for each test
			pflag.CommandLine = pflag.NewFlagSet(t.Name(), pflag.ContinueOnError)

			// Set flags if provided
			if tt.flags != nil {
				os.Args = append([]string{"test"}, tt.flags...)
			} else {
				os.Args = []string{"test"}
			}

			cfg, err := LoadConfig(tt.configFile)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedConfig, cfg)
		})
	}
}
