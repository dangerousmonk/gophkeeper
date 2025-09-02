package postgres

import (
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func ApplyMigrations(dsn string) error {
	slog.Info("ApplyMigrations start", slog.String("dsn", dsn))
	m, err := migrate.New("file://migrations/", dsn)
	if err != nil {
		slog.Error("ApplyMigrations failed init instance ", slog.Any("err", err))
		return err
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("ApplyMigrations no change")
			return nil
		}
		slog.Error("ApplyMigrations failed", slog.Any("err", err))
		return err
	}
	slog.Info("ApplyMigrations success")
	return nil
}
