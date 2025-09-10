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
	sproto "github.com/dangerousmonk/gophkeeper/internal/server/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// vaultItemWithData represents a fully reconstructed VaultItem.
type vaultItemWithData struct {
	*sproto.VaultItem
	ReconstructedData []byte
}

func contextWithToken(ctx context.Context, token string) context.Context {
	if token != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}

	return ctx
}

func registerUser(client sproto.GophKeeperClient, login, password string) tea.Cmd {
	const timeout = 3 * time.Second

	return func() tea.Msg {
		req := sproto.RegisterUserRequest_builder{
			Login:    &login,
			Password: &password,
		}.Build()

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
			Success: resp.GetSuccess(),
			Message: "Registered successfully",
			Err:     nil,
			Token:   resp.GetToken(),
			Login:   login,
		}
	}
}

func loginUser(client sproto.GophKeeperClient, login, password string) tea.Cmd {
	const timeout = 3 * time.Second

	return func() tea.Msg {
		req := sproto.LoginUserRequest_builder{
			Login:    &login,
			Password: &password,
		}.Build()

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
			Success: resp.GetSuccess(),
			Message: "Logged in successfully",
			Err:     nil,
			Token:   resp.GetToken(),
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

		req := sproto.SaveVaultRequest_builder{
			Name:         &title,
			DataType:     (*string)(&m.SecretType),
			EcryptedData: encryptedData,
		}.Build()

		ctx = contextWithToken(ctx, m.Token)

		resp, err := m.client.SaveVault(ctx, req)
		if err != nil {
			return messages.SaveVaultResultMsg{
				Err:     fmt.Errorf("gRPC call failed: %w", err),
				Success: false,
			}
		}

		return messages.SaveVaultResultMsg{
			Success: resp.GetSuccess(),
			Err:     nil,
		}
	}
}

func deactivateVaultGrpc(client sproto.GophKeeperClient, token string, vault *sproto.VaultItem) tea.Cmd {
	const timeout = 3 * time.Second

	return func() tea.Msg {
		if vault == nil {
			return messages.DeactivateVaultResultMsg{
				Err: fmt.Errorf("no vault selected"),
			}
		}

		req := sproto.DeactivateVaultRequest_builder{
			SecretId: proto.Int32(vault.GetId()),
		}.Build()

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
			Success: resp.GetSuccess(),
			Err:     nil,
		}
	}
}

func uploadFile(ctx context.Context, c sproto.GophKeeperClient, fname string, encData []byte, metaData *structpb.Struct) error {
	const chunkSize = 1024

	byteReader := bytes.NewReader(encData)
	reader := bufio.NewReader(byteReader)
	buffer := make([]byte, chunkSize)

	stream, err := c.UploadFile(ctx)
	if err != nil {
		slog.Warn("uploadFile:failed create stream", slog.Any("error", err))
		return err
	}

	req := sproto.UploadFileRequest_builder{
		FileName: &fname,
		MetaData: metaData,
	}.Build()

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

		req := sproto.UploadFileRequest_builder{
			FileName:  &fname,
			ChunkData: buffer[:n],
		}.Build()

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
func getVaultsStream(client sproto.GophKeeperClient, token, encryptionKey string) tea.Cmd {
	const timeout = 15 * time.Second

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		defer cancel()

		slog.Info("GetVaultsStream:started")

		ctx = contextWithToken(ctx, token)

		stream, err := client.GetSteamedVaults(ctx, &sproto.StreamVaultsRequest{})
		if err != nil {
			return messages.GetVaultsResultMsg{
				Err:    fmt.Errorf("failed to create stream: %w", err),
				Vaults: nil,
			}
		}

		var (
			currentItem    *vaultItemWithData
			currentChunks  [][]byte
			currentMeta    *sproto.StreamMetadata
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

			switch response.WhichPayload() {
			case sproto.StreamVaultsResponse_Metadata_case:
				mu.Lock()

				currentMeta = response.GetMetadata()

				// If we have a completed item from previous metadata, add it to results
				if currentItem != nil && len(currentChunks) > 0 {
					reconstructed := files.MergeChunks(currentChunks)
					currentItem.ReconstructedData = reconstructed
					collectedItems = append(collectedItems, currentItem)
					currentChunks = nil
					currentItem = nil
				}

				mu.Unlock()

			case sproto.StreamVaultsResponse_ItemChunk_case:
				chunk := response.GetItemChunk()

				mu.Lock()

				// Initialize new item if this is the first chunk
				if chunk.GetIsFirstChunk() {
					currentItem = &vaultItemWithData{
						VaultItem: chunk.GetItem(),
					}
					currentChunks = make([][]byte, chunk.GetTotalChunks())
				}

				// Store chunk in correct position
				if int(chunk.GetChunkIndex()) < len(currentChunks) {
					currentChunks[chunk.GetChunkIndex()] = chunk.GetEncryptedDataChunk()
				}

				// If this is the last chunk and we have metadata indicating last item,
				// process the completed item immediately
				if chunk.GetIsLastChunk() && currentMeta != nil && currentMeta.GetIsLastItem() {
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

		decryptedVaults := make([]*sproto.VaultItem, 0, len(collectedItems))
		for _, vault := range collectedItems {
			decryptedVault := sproto.VaultItem_builder{
				Id:            proto.Int32(vault.GetId()),
				UserId:        proto.Int32(vault.GetUserId()),
				Name:          proto.String(vault.GetName()),
				DataType:      proto.String(vault.GetDataType()),
				EncryptedData: nil,
				MetaData:      vault.GetMetaData(),
				CreatedAt:     proto.String(vault.GetCreatedAt()),
				UpdatedAt:     proto.String(vault.GetUpdatedAt()),
				Active:        proto.Bool(vault.GetActive()),
				Version:       proto.Int32(vault.GetVersion()),
			}.Build()

			if len(vault.ReconstructedData) > 0 {
				decryptedData, err := encryption.DecryptData(vault.ReconstructedData, encryptionKey)
				if err != nil {
					slog.Warn("GetVaultsStream:decryption error", slog.Any("error", err))
				} else {
					decryptedVault.SetEncryptedData(decryptedData)
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
func changePassword(client sproto.GophKeeperClient, login, token, currentPassword, newPassword string) tea.Cmd {
	const timeout = 3 * time.Second

	return func() tea.Msg {
		req := sproto.ChangePasswordRequest_builder{
			CurrentPassword: &currentPassword,
			NewPassword:     &newPassword,
			Login:           &login,
		}.Build()

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
			Success: resp.GetSuccess(),
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

		switch m.SelectedVault.GetDataType() {
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

		req := sproto.UpdateVaultRequest_builder{
			Id:            proto.Int32(m.SelectedVault.GetId()),
			EncryptedData: encryptedData,
			Name:          &title,
		}.Build()

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
			Success: resp.GetSuccess(),
			Err:     nil,
		}
	}
}
