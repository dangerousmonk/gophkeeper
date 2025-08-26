package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/dangerousmonk/gophkeeper/internal/config"
	"github.com/dangerousmonk/gophkeeper/internal/postgres"
	"github.com/dangerousmonk/gophkeeper/internal/server"
	"github.com/dangerousmonk/gophkeeper/internal/service"
	"github.com/dangerousmonk/gophkeeper/internal/utils"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("main:LoadConfig failed error=%v", err)
	}

	logger := utils.InitLogger(cfg.Environment, os.Stdout)
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dsn := postgres.GetDSN(cfg)
	db, err := postgres.InitDB(ctx, dsn)
	if err != nil {
		logger.Error("main:InitDB failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	err = postgres.ApplyMigrations(dsn)
	if err != nil {
		logger.Error("main:ApplyMigrations failed", slog.Any("error", err))
		os.Exit(1)
	}

	jwtAuthenticator, err := utils.NewJWTAuthenticator(cfg.JWTSecret)
	if err != nil {
		logger.Error("main:NewJWTAuthenticator failed", slog.Any("error", err))
		os.Exit(1)
	}

	repos := postgres.NewPostgresRepositories(db)
	userService := service.NewUserService(repos.User)
	vaultService := service.NewVaultService(repos.Vault)
	app := server.NewGophKeeperApp(cfg, logger, userService, vaultService, &jwtAuthenticator)
	err = app.Start()
	if err != nil {
		logger.Error("main:failed start application", slog.Any("error", err))
		os.Exit(1)
	}
}
