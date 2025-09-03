package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dangerousmonk/gophkeeper/internal/client/components"
	"github.com/dangerousmonk/gophkeeper/internal/config"
	"github.com/dangerousmonk/gophkeeper/internal/logger"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("main:LoadConfig failed=%v", err)
	}

	const logFileMode = 0o666 // Read and write for owner, group, and others

	f, err := os.OpenFile("client.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, logFileMode)
	if err != nil {
		log.Fatalf("main:OpenFile failed error=%v", err)
	}
	defer f.Close()

	appLog := logger.InitLogger(cfg.Environment, f)
	slog.SetDefault(appLog)

	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		appLog.Error("grpc:connection failed", slog.Any("error", err))
		return
	}

	defer conn.Close()

	client := proto.NewGophKeeperClient(conn)

	p := tea.NewProgram(components.NewModel(conn, client, appLog, cfg.EncryptionKey), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		appLog.Error("main:tea run failed", slog.Any("error", err))
		return
	}
}
