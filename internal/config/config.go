package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// LoadConfig loads configuration from multiple sources with the following precedence:
// 1. Flags
// 2. Environment variables
// 3. Config file
// 4. Default values
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("LoadConfig:error=%v", err)
	}

	// Set default values
	setDefaults(v)

	// Bind environment variables
	bindEnv(v)

	// Bind command line flags
	bindFlags(v)

	// Read config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("SERVER_HOST", "localhost")
	v.SetDefault("SERVER_PORT", "8099")

	// Database defaults
	v.SetDefault("DB_USER", "gophkeeper")
	v.SetDefault("DB_NAME", "gophkeeper_db")
	v.SetDefault("DB_PASSWORD", "password")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_SSL_MODE", "disable")

	// Other defaults
	v.SetDefault("JWT_SECRET", "default-secret")
	v.SetDefault("ENV", "local")
}

func bindEnv(v *viper.Viper) {
	// Automatic env binds to all keys in uppercase with underscores
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func bindFlags(v *viper.Viper) {
	var _ *string = pflag.StringP("dsn", "d", "", "Database DSN")
	var _ *string = pflag.StringP("db_user", "u", "", "Database user name")
	var _ *string = pflag.StringP("db_name", "n", "", "Database name")
	var _ *string = pflag.StringP("db_pass", "p", "", "Database password")
	var _ *string = pflag.StringP("env", "e", "", "Environment to use")

	// Parse flags
	pflag.Parse()

	// Bind flags to viper
	v.BindPFlag("DB_USER", pflag.Lookup("db_user"))
	v.BindPFlag("DB_NAME", pflag.Lookup("db_name"))
	v.BindPFlag("DB_PASSWORD", pflag.Lookup("db_pass"))
	v.BindPFlag("ENV", pflag.Lookup("env"))
}
