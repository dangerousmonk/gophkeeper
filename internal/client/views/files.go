package views

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dangerousmonk/gophkeeper/internal/client/messages"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
)

func (m *Model) downloadFile(path string) tea.Cmd {
	return func() tea.Msg {
		vault := m.SelectedVault
		if vault == nil || vault.DataType != "binary" {
			return messages.DownloadResultMsg{
				Err:     fmt.Errorf("no file selected for download"),
				Success: false,
			}
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return messages.DownloadResultMsg{
				Err:     fmt.Errorf("failed to create directory: %w", err),
				Success: false,
			}
		}

		// Check if file already exists and create unique name
		finalPath := path
		counter := 1
		for {
			if _, err := os.Stat(finalPath); os.IsNotExist(err) {
				break
			}
			ext := filepath.Ext(path)
			name := path[:len(path)-len(ext)]
			finalPath = fmt.Sprintf("%s_%d%s", name, counter, ext)
			counter++
		}

		// Write file to disk
		if err := os.WriteFile(finalPath, m.SelectedVault.EncryptedData, 0644); err != nil {
			return messages.DownloadResultMsg{
				Err:     fmt.Errorf("failed to write file: %w", err),
				Success: false,
			}
		}

		return messages.DownloadResultMsg{
			Success: true,
			Message: fmt.Sprintf("File downloaded to: %s", finalPath),
		}
	}
}

func getDefaultDownloadPath(vault *proto.VaultItem) string {
	fileName := "downloaded_file"
	if vault.MetaData == nil || vault.MetaData.Fields == nil {
		return filepath.Join("./downloads/", fileName)
	}

	if nameVal, exists := vault.MetaData.Fields["file_name"]; exists {
		fileName = nameVal.GetStringValue()
	}
	return filepath.Join("./downloads/", fileName)
}
