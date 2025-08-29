package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dangerousmonk/gophkeeper/internal/client/components"
	"github.com/dangerousmonk/gophkeeper/internal/config"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
	"github.com/dangerousmonk/gophkeeper/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("main:LoadConfig failed=%v", err)
	}

	f, err := os.OpenFile("client.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		os.Exit(1)
	}
	defer f.Close()

	logger := utils.InitLogger(cfg.Environment, f)
	slog.SetDefault(logger)

	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("grpc:connection failed", slog.Any("error", err))
		os.Exit(1)
	}

	defer conn.Close()

	client := proto.NewGophKeeperClient(conn)

	logger.Info("main:setup completed")

	p := tea.NewProgram(components.NewModel(conn, &client, logger), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logger.Error("main:tea run failed", slog.Any("error", err))
		os.Exit(1)
	}
}
