package components

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dangerousmonk/gophkeeper/internal/client/messages"
	"github.com/dangerousmonk/gophkeeper/internal/encryption"
	"github.com/dangerousmonk/gophkeeper/internal/files"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

// vaultItemWithData represents a fully reconstructed VaultItem.
type vaultItemWithData struct {
	*proto.VaultItem
	ReconstructedData []byte
}

func contextWithToken(ctx context.Context, token string) context.Context {
	if token != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}

	return ctx
}

func registerUser(client proto.GophKeeperClient, login, password string) tea.Cmd {
	const timeout = 3 * time.Second

	return func() tea.Msg {
		req := &proto.RegisterUserRequest{
			Login:    login,
			Password: password,
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		resp, err := client.RegisterUser(ctx, req)
		if err != nil {
			return messages.RegistrationResultMsg{
				Err:     fmt.Errorf("gRPC call failed: %w", err),
				Success: false,
			}
		}

		return messages.RegistrationResultMsg{
			Success: resp.Success,
			Message: "Registered successfully",
			Err:     nil,
			Token:   resp.Token,
			Login:   login,
		}
	}
}

func loginUser(client proto.GophKeeperClient, login, password string) tea.Cmd {
	const timeout = 3 * time.Second

	return func() tea.Msg {
		req := &proto.LoginUserRequest{
			Login:    login,
			Password: password,
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		resp, err := client.LoginUser(ctx, req)
		if err != nil {
			return messages.LoginResultMsg{
				Err:     fmt.Errorf("gRPC call failed: %w", err),
				Success: false,
			}
		}

		return messages.LoginResultMsg{
			Success: resp.Success,
			Message: "Logged in successfully",
			Err:     nil,
			Token:   resp.Token,
			Pasword: password,
			Login:   login,
		}
	}
}

//nolint:funlen // gRPC save stream logic
func saveVault(
	m *Model,
	title string,
) tea.Cmd {
	const timeout = 15 * time.Second

	return func() tea.Msg {
		var secretData map[string]string

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		switch m.SecretType {
		case secretTypeCredential:
			secretData = map[string]string{
				"service":  m.FormData["Service"],
				"username": m.FormData["Username"],
				"password": m.FormData["Password"],
				"url":      m.FormData["URL"],
			}
		case secretTypeBankCard:
			secretData = map[string]string{
				"card_name":   m.FormData["Card Name"],
				"card_number": m.FormData["Card Number"],
				"expiry":      m.FormData["Expiry"],
				"cvv":         m.FormData["CVV"],
				"cardholder":  m.FormData["Cardholder"],
			}
		case secretTypeText:
			secretData = map[string]string{
				"title":   m.FormData["Title"],
				"content": m.FormData["Content"],
			}
		case secretTypeBinary:
			fPath := m.FormData["File Path"]
			fName := m.FormData["File Name"]

			encryptedData, err := encryption.EncryptFile(fPath, m.encryptionKey)
			if err != nil {
				return messages.SaveVaultResultMsg{
					Err:     err,
					Success: false,
				}
			}

			metaData, err := files.GetFileMetadata(fPath)
			if err != nil {
				return messages.SaveVaultResultMsg{
					Err:     fmt.Errorf("failed to read file metadata %w", err),
					Success: false,
				}
			}

			metaDataStruct, err := structpb.NewStruct(metaData)
			if err != nil {
				return messages.SaveVaultResultMsg{
					Err:     fmt.Errorf("failed to create metadata struct: %w", err),
					Success: false,
				}
			}

			_, err = json.Marshal(secretData)
			if err != nil {
				return messages.SaveVaultResultMsg{
					Err:     fmt.Errorf("failed to marshal data: %w", err),
					Success: false,
				}
			}

			ctx = contextWithToken(ctx, m.Token)

			slog.Info("SaveVault:uploadFile started", slog.String("file_name", fName))

			err = uploadFile(ctx, m.client, fName, encryptedData, metaDataStruct)
			if err != nil {
				return messages.SaveVaultResultMsg{
					Err:     fmt.Errorf("SaveVault:gRPC call failed: %w", err),
					Success: false,
				}
			}

			return messages.SaveVaultResultMsg{
				Success: true,
				Err:     nil,
			}
		}

		jsonData, err := json.Marshal(secretData)
		if err != nil {
			return messages.SaveVaultResultMsg{
				Err:     fmt.Errorf("failed to marshal data: %w", err),
				Success: false,
			}
		}

		encryptedData, err := encryption.EncryptData(jsonData, m.encryptionKey)
		if err != nil {
			return messages.SaveVaultResultMsg{
				Err:     fmt.Errorf("failed to encrypt data: %w", err),
				Success: false,
			}
		}

		req := &proto.SaveVaultRequest{
			Name:         title,
			DataType:     string(m.SecretType),
			EcryptedData: encryptedData,
		}

		ctx = contextWithToken(ctx, m.Token)

		resp, err := m.client.SaveVault(ctx, req)
		if err != nil {
			return messages.SaveVaultResultMsg{
				Err:     fmt.Errorf("gRPC call failed: %w", err),
				Success: false,
			}
		}

		return messages.SaveVaultResultMsg{
			Success: resp.Success,
			Err:     nil,
		}
	}
}

func deactivateVaultGrpc(client proto.GophKeeperClient, token string, vault *proto.VaultItem) tea.Cmd {
	const timeout = 3 * time.Second

	return func() tea.Msg {
		if vault == nil {
			return messages.DeactivateVaultResultMsg{
				Err: fmt.Errorf("no vault selected"),
			}
		}

		req := &proto.DeactivateVaultRequest{
			SecretId: vault.Id,
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ctx = contextWithToken(ctx, token)

		resp, err := client.DeactivateVault(ctx, req)
		if err != nil {
			return messages.DeactivateVaultResultMsg{
				Err:     fmt.Errorf("gRPC call failed: %w", err),
				Success: false,
			}
		}

		return messages.DeactivateVaultResultMsg{
			Success: resp.Success,
			Err:     nil,
		}
	}
}

func uploadFile(ctx context.Context, c proto.GophKeeperClient, fname string, encData []byte, metaData *structpb.Struct) error {
	const chunkSize = 1024

	byteReader := bytes.NewReader(encData)
	reader := bufio.NewReader(byteReader)
	buffer := make([]byte, chunkSize)

	stream, err := c.UploadFile(ctx)
	if err != nil {
		slog.Warn("uploadFile:failed create stream", slog.Any("error", err))
		return err
	}

	req := &proto.UploadFileRequest{
		FileName: fname,
		Data:     &proto.UploadFileRequest_MetaData{MetaData: metaData},
	}

	err = stream.Send(req)
	if err != nil {
		slog.Warn("uploadFile:failed send metadata", slog.Any("error", err))
		return err
	}

	for {
		n, err := reader.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			slog.Warn("uploadFile:cannot read chunk to buffer", slog.Any("error", err))
			return err
		}

		req := &proto.UploadFileRequest{
			FileName: fname,
			Data:     &proto.UploadFileRequest_ChunkData{ChunkData: buffer[:n]},
		}

		err = stream.Send(req)
		if err != nil {
			slog.Warn("uploadFile:failed to send chunk", slog.Any("error", err))
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		slog.Warn("uploadFile:failed to close and recv", slog.Any("error", err))
		return err
	}

	slog.Info("uploadFile:finished", slog.String("file_name", fname))

	return nil
}

// getVaultsStream retrieves vault items via streaming with automatic chunk reassembly.
//
//nolint:funlen // gRPC get items stream logic
func getVaultsStream(client proto.GophKeeperClient, token, encryptionKey string) tea.Cmd {
	const timeout = 15 * time.Second

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		defer cancel()

		slog.Info("GetVaultsStream:started")

		ctx = contextWithToken(ctx, token)

		stream, err := client.GetSteamedVaults(ctx, &proto.StreamVaultsRequest{})
		if err != nil {
			return messages.GetVaultsResultMsg{
				Err:    fmt.Errorf("failed to create stream: %w", err),
				Vaults: nil,
			}
		}

		var (
			currentItem    *vaultItemWithData
			currentChunks  [][]byte
			currentMeta    *proto.StreamMetadata
			mu             sync.Mutex
			collectedItems []*vaultItemWithData
		)

		for {
			select {
			case <-ctx.Done():
				return messages.GetVaultsResultMsg{
					Err:    fmt.Errorf("context canceled: %w", ctx.Err()),
					Vaults: nil,
				}
			default:
			}

			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break // Stream completed successfully
			}

			if err != nil {
				return messages.GetVaultsResultMsg{
					Err:    fmt.Errorf("stream receive error: %w", err),
					Vaults: nil,
				}
			}

			switch payload := response.Payload.(type) {
			case *proto.StreamVaultsResponse_Metadata:
				mu.Lock()

				currentMeta = payload.Metadata

				// If we have a completed item from previous metadata, add it to results
				if currentItem != nil && len(currentChunks) > 0 {
					reconstructed := files.MergeChunks(currentChunks)
					currentItem.ReconstructedData = reconstructed
					collectedItems = append(collectedItems, currentItem)
					currentChunks = nil
					currentItem = nil
				}

				mu.Unlock()

			case *proto.StreamVaultsResponse_ItemChunk:
				chunk := payload.ItemChunk

				mu.Lock()

				// Initialize new item if this is the first chunk
				if chunk.IsFirstChunk {
					currentItem = &vaultItemWithData{
						VaultItem: chunk.Item,
					}
					currentChunks = make([][]byte, chunk.TotalChunks)
				}

				// Store chunk in correct position
				if int(chunk.ChunkIndex) < len(currentChunks) {
					currentChunks[chunk.ChunkIndex] = chunk.EncryptedDataChunk
				}

				// If this is the last chunk and we have metadata indicating last item,
				// process the completed item immediately
				if chunk.IsLastChunk && currentMeta != nil && currentMeta.IsLastItem {
					reconstructed := files.MergeChunks(currentChunks)
					currentItem.ReconstructedData = reconstructed
					collectedItems = append(collectedItems, currentItem)
					currentChunks = nil
					currentItem = nil
				}

				mu.Unlock()
			}
		}

		slog.Info("GetVaultsStream: done collecting", slog.Int("len", len(collectedItems)))

		decryptedVaults := make([]*proto.VaultItem, 0, len(collectedItems))
		for _, vault := range collectedItems {
			decryptedVault := &proto.VaultItem{
				Id:            vault.Id,
				UserId:        vault.UserId,
				Name:          vault.Name,
				DataType:      vault.DataType,
				EncryptedData: nil,
				MetaData:      vault.MetaData,
				CreatedAt:     vault.CreatedAt,
				UpdatedAt:     vault.UpdatedAt,
				Active:        vault.Active,
				Version:       vault.Version,
			}

			if len(vault.ReconstructedData) > 0 {
				decryptedData, err := encryption.DecryptData(vault.ReconstructedData, encryptionKey)
				if err != nil {
					slog.Warn("GetVaultsStream:decryption error", slog.Any("error", err))
				} else {
					decryptedVault.EncryptedData = decryptedData
				}
			}

			decryptedVaults = append(decryptedVaults, decryptedVault)
		}

		return messages.GetVaultsResultMsg{
			Vaults: decryptedVaults,
			Err:    nil,
		}
	}
}

// changePassword func sends gRPC request to the server to update user's password.
func changePassword(client proto.GophKeeperClient, login, token, currentPassword, newPassword string) tea.Cmd {
	const timeout = 3 * time.Second

	return func() tea.Msg {
		req := &proto.ChangePasswordRequest{
			CurrentPassword: currentPassword,
			NewPassword:     newPassword,
			Login:           login,
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ctx = contextWithToken(ctx, token)

		resp, err := client.ChangePassword(ctx, req)
		if err != nil {
			return messages.ChangePasswordResultMsg{
				Err:     fmt.Errorf("gRPC call failed: %w", err),
				Success: false,
			}
		}

		return messages.ChangePasswordResultMsg{
			Success: resp.Success,
			Err:     nil,
		}
	}
}

// updateVault func sends gRPC request to the server to update specific vault record with new data.
func updateVault(m *Model) tea.Cmd {
	const timeout = 5 * time.Second

	return func() tea.Msg {
		if m.SelectedVault == nil {
			return messages.UpdateVaultResultMsg{
				Err: fmt.Errorf("no vault selected"),
			}
		}

		var secretData map[string]string

		title := m.FormData[m.CurrentForm.Fields[0].Name]

		switch m.SelectedVault.DataType {
		case secretTypeCredential:
			secretData = map[string]string{
				"service":  m.FormData["Service"],
				"username": m.FormData["Username"],
				"password": m.FormData["Password"],
				"url":      m.FormData["URL"],
			}
		case secretTypeBankCard:
			secretData = map[string]string{
				"card_name":   m.FormData["Card Name"],
				"card_number": m.FormData["Card Number"],
				"expiry":      m.FormData["Expiry"],
				"cvv":         m.FormData["CVV"],
				"cardholder":  m.FormData["Cardholder"],
			}
		case secretTypeText:
			secretData = map[string]string{
				"title":   m.FormData["Title"],
				"content": m.FormData["Content"],
			}
		}

		jsonData, err := json.Marshal(secretData)
		if err != nil {
			return messages.UpdateVaultResultMsg{
				Err:     fmt.Errorf("failed to marshal data: %w", err),
				Success: false,
			}
		}

		encryptedData, err := encryption.EncryptData(jsonData, m.encryptionKey)
		if err != nil {
			return messages.UpdateVaultResultMsg{
				Err:     fmt.Errorf("failed to encrypt data: %w", err),
				Success: false,
			}
		}

		req := &proto.UpdateVaultRequest{
			Id:            m.SelectedVault.Id,
			EncryptedData: encryptedData,
			Name:          title,
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ctx = contextWithToken(ctx, m.Token)

		resp, err := m.client.UpdateVault(ctx, req)
		if err != nil {
			return messages.UpdateVaultResultMsg{
				Err:     fmt.Errorf("gRPC call failed: %w", err),
				Success: false,
			}
		}

		return messages.UpdateVaultResultMsg{
			Success: resp.Success,
			Err:     nil,
		}
	}
}
