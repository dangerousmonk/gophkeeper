package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/dangerousmonk/gophkeeper/internal/auth"
	"github.com/dangerousmonk/gophkeeper/internal/config"
	"github.com/dangerousmonk/gophkeeper/internal/encryption"
	"github.com/dangerousmonk/gophkeeper/internal/logger"
	"github.com/dangerousmonk/gophkeeper/internal/postgres"
	"github.com/dangerousmonk/gophkeeper/internal/server"
	"github.com/dangerousmonk/gophkeeper/internal/service"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("main:LoadConfig failed error=%v", err)
	}

	appLog := logger.InitLogger(cfg.Environment, os.Stdout)
	slog.SetDefault(appLog)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dsn := postgres.GetDSN(cfg)

	db, err := postgres.InitDB(ctx, dsn)
	if err != nil {
		appLog.Error("main:InitDB failed", slog.Any("error", err))
		return
	}

	defer db.Close()

	err = postgres.ApplyMigrations(dsn)
	if err != nil {
		appLog.Error("main:ApplyMigrations failed", slog.Any("error", err))
		return
	}

	jwtAuthenticator, err := auth.NewJWTAuthenticator(cfg.JWTSecret)
	if err != nil {
		appLog.Error("main:NewJWTAuthenticator failed", slog.Any("error", err))
		return
	}

	repos := postgres.NewPostgresRepositories(db)
	passEncryptor := encryption.NewPaswordEncryptor()
	userService := service.NewUserService(repos.User, passEncryptor)
	vaultService := service.NewVaultService(repos.Vault)

	app := server.NewGophKeeperApp(cfg, appLog, userService, vaultService, jwtAuthenticator)

	err = app.Start()
	if err != nil {
		appLog.Error("main:failed start application", slog.Any("error", err))
		return
	}
}
