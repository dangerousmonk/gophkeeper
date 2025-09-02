package components

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/dangerousmonk/gophkeeper/internal/client/messages"
	"github.com/dangerousmonk/gophkeeper/internal/encryption"
	"github.com/dangerousmonk/gophkeeper/internal/server/proto"
	"github.com/dangerousmonk/gophkeeper/internal/utils"
)

func contextWithToken(token string, ctx context.Context) context.Context {
	if token != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}
	return ctx
}

func registerUser(client proto.GophKeeperClient, login, password string) tea.Cmd {
	return func() tea.Msg {
		req := &proto.RegisterUserRequest{
			Login:    login,
			Password: password,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	return func() tea.Msg {
		req := &proto.LoginUserRequest{
			Login:    login,
			Password: password,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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

func saveVault(
	client proto.GophKeeperClient,
	token, password string,
	sType secretType,
	formData map[string]string,
	title string,
) tea.Cmd {
	return func() tea.Msg {
		var secretData map[string]string
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		switch sType {
		case secretTypeCredential:
			secretData = map[string]string{
				"service":  formData["Service"],
				"username": formData["Username"],
				"password": formData["Password"],
				"url":      formData["URL"],
			}
		case secretTypeBankCard:
			secretData = map[string]string{
				"card_name":   formData["Card Name"],
				"card_number": formData["Card Number"],
				"expiry":      formData["Expiry"],
				"cvv":         formData["CVV"],
				"cardholder":  formData["Cardholder"],
			}
		case secretTypeText:
			secretData = map[string]string{
				"title":   formData["Title"],
				"content": formData["Content"],
			}
		case secretTypeBinary:
			fPath := formData["File Path"]
			fName := formData["File Name"]
			encryptedData, err := encryption.EncryptFile(fPath, password)
			if err != nil {
				return messages.SaveVaultResultMsg{
					Err:     err,
					Success: false,
				}
			}
			metaData, err := utils.GetFileMetadata(fPath)
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

			ctx = contextWithToken(token, ctx)
			slog.Info("SaveVault:uploadFile started", slog.String("file_name", fName))

			err = uploadFile(ctx, client, fName, encryptedData, metaDataStruct)
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

		// Convert to JSON
		jsonData, err := json.Marshal(secretData)
		if err != nil {
			return messages.SaveVaultResultMsg{
				Err:     fmt.Errorf("failed to marshal data: %w", err),
				Success: false,
			}
		}

		encryptedData, err := encryption.EncryptData(jsonData, password)
		if err != nil {
			return messages.SaveVaultResultMsg{
				Err:     fmt.Errorf("failed to encrypt data: %w", err),
				Success: false,
			}
		}

		req := &proto.SaveVaultRequest{
			Name:         title,
			DataType:     string(sType),
			EcryptedData: encryptedData,
		}

		ctx = contextWithToken(token, ctx)
		resp, err := client.SaveVault(ctx, req)
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
	return func() tea.Msg {
		if vault == nil {
			return messages.DeactivateVaultResultMsg{
				Err: fmt.Errorf("no vault selected"),
			}
		}

		req := &proto.DeactivateVaultRequest{
			SecretId: vault.Id,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ctx = contextWithToken(token, ctx)

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
	byteReader := bytes.NewReader(encData)
	reader := bufio.NewReader(byteReader)
	buffer := make([]byte, 1024)

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
		if err == io.EOF {
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

// vaultItemWithData represents a fully reconstructed VaultItem
type vaultItemWithData struct {
	*proto.VaultItem
	ReconstructedData []byte
}

// getVaultsStream retrieves vault items via streaming with automatic chunk reassembly
func getVaultsStream(client proto.GophKeeperClient, token, password string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		slog.Info("GetVaultsStream:started")
		ctx = contextWithToken(token, ctx)

		stream, err := client.GetSteamedVaults(ctx, &emptypb.Empty{})
		if err != nil {
			return messages.GetVaultsResultMsg{
				Err:    fmt.Errorf("failed to create stream: %v", err),
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
					Err:    fmt.Errorf("context canceled: %v", ctx.Err()),
					Vaults: nil,
				}
			default:
			}

			response, err := stream.Recv()
			if err == io.EOF {
				break // Stream completed successfully
			}
			if err != nil {
				return messages.GetVaultsResultMsg{
					Err:    fmt.Errorf("stream receive error: %v", err),
					Vaults: nil,
				}
			}

			switch payload := response.Payload.(type) {
			case *proto.StreamVaultsResponse_Metadata:
				mu.Lock()
				currentMeta = payload.Metadata

				// If we have a completed item from previous metadata, add it to results
				if currentItem != nil && len(currentChunks) > 0 {
					reconstructed := utils.MergeChunks(currentChunks)
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
					reconstructed := utils.MergeChunks(currentChunks)
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
				decryptedData, err := encryption.DecryptData(vault.ReconstructedData, password)
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

func changePassword(client proto.GophKeeperClient, login, token, currentPassword, newPassword string) tea.Cmd {
	return func() tea.Msg {
		req := &proto.ChangePasswordRequest{
			CurrentPassword: currentPassword,
			NewPassword:     newPassword,
			Login:           login,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ctx = contextWithToken(token, ctx)
		resp, err := client.ChangePassword(ctx, req)
		if err != nil {
			return messages.ChangePasswordResultMsg{
				Err:    fmt.Errorf("gRPC call failed: %w", err),
				Sucess: false,
			}
		}

		return messages.ChangePasswordResultMsg{
			Sucess: resp.Success,
			Err:    nil,
		}
	}
}
