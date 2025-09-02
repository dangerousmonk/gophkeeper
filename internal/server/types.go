package server

import (
	"log/slog"

	"github.com/dangerousmonk/gophkeeper/internal/auth"
	"github.com/dangerousmonk/gophkeeper/internal/config"
	"github.com/dangerousmonk/gophkeeper/internal/middleware"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
	"github.com/dangerousmonk/gophkeeper/internal/service"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// GophKeeperApp is a structure to represent gophkeeper server side app and its main components
type GophKeeperApp struct {
	Config        *config.Config
	Logger        *slog.Logger
	GRPCServer    *grpc.Server
	Authenticator *auth.Authenticator
}

func NewGophKeeperApp(
	cfg *config.Config,
	l *slog.Logger,
	userService *service.UserService,
	vaultService *service.VaultService,
	authenticator *auth.Authenticator,
) *GophKeeperApp {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			l.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}
	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		middleware.AuthUnaryInterceptor(*authenticator),
	),
		grpc.ChainStreamInterceptor(middleware.StreamAuthInterceptor(*authenticator)),
	)

	proto.RegisterGophKeeperServer(gRPCServer, proto.NewGophKeepergRPCServer(userService, vaultService, cfg, *authenticator))
	reflection.Register(gRPCServer)

	return &GophKeeperApp{
		Config:        cfg,
		Logger:        l,
		GRPCServer:    gRPCServer,
		Authenticator: authenticator,
	}
}
