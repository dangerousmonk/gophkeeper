package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/dangerousmonk/gophkeeper/internal/config"
)

// GetDSN returns database connection string filled from config
func GetDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
}

// InitDB function is used to initialize new DB
func InitDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	slog.Info("InitDB success", slog.String("dsn", dsn))
	return db, nil
}
