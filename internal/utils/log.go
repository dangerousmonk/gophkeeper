package utils

import (
	"io"
	"log/slog"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// InitLogger create new slog.Logger with specified writer, and configures handler and log level based on environment
func InitLogger(env string, w io.Writer) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
