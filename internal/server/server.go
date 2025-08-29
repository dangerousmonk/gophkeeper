package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

func (app *GophKeeperApp) Start() error {
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	go func() {
		if _, err := startgRPS(rootCtx, app); err != nil {
			app.Logger.Error("startgRPS failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-rootCtx.Done()
	app.Logger.Info("Received shutdown signal, shutting down...")
	return nil
}

// startgRPS starts grps server
func startgRPS(ctx context.Context, app *GophKeeperApp) (*grpc.Server, error) {
	address := fmt.Sprintf("%s:%s", app.Config.Server.Host, app.Config.Server.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on gRPC: %w", err)
	}

	go func() {
		if err := app.GRPCServer.Serve(lis); err != nil {
			app.Logger.Error("startgRPS:server failed to start", slog.Any("error", err))
			os.Exit(1)

		}
	}()
	app.Logger.Info("startgRPS:gRPC server started", slog.String("server_address", address))

	go func() {
		<-ctx.Done()
		app.GRPCServer.GracefulStop()
	}()

	return app.GRPCServer, nil
}
