package proto

import (
	"bytes"
	context "context"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/auth"
	"github.com/dangerousmonk/gophkeeper/internal/config"
	"github.com/dangerousmonk/gophkeeper/internal/middleware"
	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/dangerousmonk/gophkeeper/internal/service"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

const (
	chunkSize = 1024 * 1024 // 1MB chunk
)

// GophKeepergRPCServer Supports all the service methods.
type GophKeepergRPCServer struct {
	UnimplementedGophKeeperServer
	userService   service.UserHandler
	vaultService  service.VaultHandler
	cfg           *config.Config
	authenticator auth.Authenticator
}

// NewGophKeepergRPCServer creates the ShortenerGRPCServer structure and returns a pointer to freshly created struct.
func NewGophKeepergRPCServer(
	userService service.UserHandler,
	vaultService service.VaultHandler,
	cfg *config.Config,
	authenticator auth.Authenticator,
) *GophKeepergRPCServer {
	return &GophKeepergRPCServer{
		cfg:           cfg,
		authenticator: authenticator,
		userService:   userService,
		vaultService:  vaultService,
	}
}

// Ping checks the service health.
func (srv GophKeepergRPCServer) Ping(ctx context.Context, _ *PingRequest) (*PingResponse, error) {
	err := srv.userService.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &PingResponse{}, nil
}

// RegisterUser is used to register new user.
func (srv GophKeepergRPCServer) RegisterUser(ctx context.Context, req *RegisterUserRequest) (*RegisterUserResponse, error) {
	registerReq := &models.RegisterUserRequest{Login: req.GetLogin(), Password: req.GetPassword()}

	res, err := srv.userService.Register(ctx, registerReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	token, err := srv.authenticator.CreateToken(res.ID, time.Hour*1)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := RegisterUserResponse_builder{Id: proto.Uint64(uint64(res.ID)), Login: &res.Login, Token: &token, Success: &res.Success}.Build()

	return resp, nil
}

// RegisterUser is used to register new user.
func (srv GophKeepergRPCServer) LoginUser(ctx context.Context, req *LoginUserRequest) (*LoginUserResponse, error) {
	registerReq := &models.LoginUserRequest{Login: req.GetLogin(), Password: req.GetPassword()}

	token, err := srv.userService.Login(ctx, registerReq.Login, registerReq.Password, srv.authenticator)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := LoginUserResponse_builder{Token: &token, Success: proto.Bool(true)}.Build()

	return resp, nil
}

// SaveVault saves data from client to vault.
func (srv GophKeepergRPCServer) SaveVault(ctx context.Context, req *SaveVaultRequest) (*SaveVaultResponse, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	v := models.Vault{
		UserID:        userID,
		Name:          req.GetName(),
		DataType:      models.DataType(req.GetDataType()),
		EncryptedData: req.GetEcryptedData(),
		MetaData:      req.GetMetaData().AsMap(),
	}

	_, err := srv.vaultService.Save(ctx, &v)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return SaveVaultResponse_builder{Success: proto.Bool(true)}.Build(), nil
}

// DeactivateVault is used to soft delete specific vault by using active flag.
func (srv GophKeepergRPCServer) DeactivateVault(ctx context.Context, req *DeactivateVaultRequest) (*DeactivateVaultResponse, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	err := srv.vaultService.Deactivate(ctx, userID, int(req.GetSecretId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return DeactivateVaultResponse_builder{Success: proto.Bool(true)}.Build(), nil
}

func (srv GophKeepergRPCServer) UploadFile(stream GophKeeper_UploadFileServer) error {
	userID, ok := middleware.UserIDFromContext(stream.Context())
	if !ok {
		slog.Warn("uploadFile:unauthorized failed", slog.Any("user_id", userID), slog.Any("context", stream.Context()))
		return status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	req, err := stream.Recv()
	if err != nil {
		slog.Warn("uploadFile:failed", slog.Any("error", err))
		return status.Error(codes.Unknown, err.Error())
	}

	fileData := bytes.Buffer{}
	fileSize := 0
	metaData := req.GetMetaData()

	slog.Info("uploadFile:received request", slog.Any("meta_data", metaData))

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			slog.Info("uploadFile:request", slog.String("message", "no more data"))
			break
		}

		if err != nil {
			slog.Warn("uploadFile:failed", slog.Any("error", err))
			return status.Error(codes.Unknown, err.Error())
		}

		chunk := req.GetChunkData()
		size := len(chunk)
		fileSize += size

		_, err = fileData.Write(chunk)
		if err != nil {
			slog.Warn("uploadFile:chank failed", slog.Any("error", err))
			return status.Error(codes.Internal, err.Error())
		}
	}

	v := models.Vault{
		UserID:        userID,
		Name:          req.GetFileName(),
		DataType:      models.Binary,
		EncryptedData: fileData.Bytes(),
		MetaData:      metaData.AsMap(),
	}

	res, err := srv.vaultService.Save(context.Background(), &v)
	if err != nil {
		slog.Warn("uploadFile:service save failed", slog.Any("error", err))
		return status.Error(codes.Internal, err.Error())
	}

	vItem := VaultItem_builder{
		Id:            proto.Int32(int32(res.ID)),
		UserId:        proto.Int32(int32(res.UserID)),
		DataType:      (*string)(&res.DataType),
		Name:          &res.Name,
		EncryptedData: res.EncryptedData,
		Version:       proto.Int32(int32(res.Version)),
		CreatedAt:     proto.String(res.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		UpdatedAt:     proto.String(res.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")),
		Active:        proto.Bool(res.Active),
	}.Build()

	err = stream.SendAndClose(vItem)
	if err != nil {
		slog.Warn("uploadFile:send and close failed", slog.Any("error", err))
		return status.Error(codes.Internal, err.Error())
	}

	slog.Info("uploadFile:success", slog.Int("file_size", fileSize))

	return nil
}

//nolint:funlen // gRPC stream logic
func (srv *GophKeepergRPCServer) GetSteamedVaults(_ *StreamVaultsRequest, stream GophKeeper_GetSteamedVaultsServer) error {
	ctx := stream.Context()
	userID, ok := middleware.UserIDFromContext(ctx)

	if !ok {
		slog.Warn("GetSteamedVaults:failed unauthorize", slog.Int("user_id", userID))
		return status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	vaults, err := srv.vaultService.GetByUser(ctx, userID)
	if err != nil {
		slog.Warn("GetSteamedVaults:service fetch failed", slog.Any("error", err))
		return status.Error(codes.Internal, err.Error())
	}

	slog.Info("GetSteamedVaults: received items from service", slog.Int("len", len(vaults)))

	var vaultItems []*VaultItem

	for i := range vaults {
		v := &vaults[i]

		pbMeta, err := structpb.NewStruct(v.MetaData)
		if err != nil {
			slog.Warn("GetSteamedVaults:error creating structpb.Struct", slog.Any("error", err))
			continue
		}

		vaultItems = append(vaultItems, VaultItem_builder{
			Id:            proto.Int32(int32(v.ID)),
			UserId:        proto.Int32(int32(v.UserID)),
			DataType:      (*string)(&v.DataType),
			Name:          &v.Name,
			EncryptedData: v.EncryptedData,
			MetaData:      pbMeta,
			Version:       proto.Int32(int32(v.Version)),
			CreatedAt:     proto.String(v.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
			UpdatedAt:     proto.String(v.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")),
			Active:        proto.Bool(v.Active),
		}.Build())
	}

	totalItems := len(vaultItems)
	if totalItems == 0 {
		return nil
	}

	// Stream each vault item with chunked encrypted data
	for itemIndex, item := range vaultItems {
		isFirstItem := itemIndex == 0
		isLastItem := itemIndex == totalItems-1

		// Send metadata for the current item
		metadata := StreamMetadata_builder{
			TotalItems:       proto.Int32(int32(totalItems)),
			CurrentItemIndex: proto.Int32(int32(itemIndex)),
			IsFirstItem:      proto.Bool(isFirstItem),
			IsLastItem:       proto.Bool(isLastItem),
		}.Build()

		if err := stream.Send(StreamVaultsResponse_builder{Metadata: metadata}.Build()); err != nil {
			slog.Warn("GetSteamedVaults:failed to send metadata", slog.Any("error", err))
			return status.Errorf(codes.Internal, "failed to send metadata: %v", err)
		}

		// Handle chunking of encrypted_data
		encryptedData := item.GetEncryptedData()
		totalChunks := (len(encryptedData) + chunkSize - 1) / chunkSize

		if len(encryptedData) == 0 {
			if err := srv.sendItemChunk(stream, item, nil, 0, 1, true, true); err != nil {
				slog.Warn("GetSteamedVaults: failed to send item without encrypted data chunks", slog.Any("error", err))
				return err
			}

			continue
		}

		// Stream encrypted data in chunks
		for chunkIndex := range totalChunks {
			start := chunkIndex * chunkSize
			end := min(start+chunkSize, len(encryptedData))

			chunk := encryptedData[start:end]
			isFirstChunk := chunkIndex == 0
			isLastChunk := chunkIndex == totalChunks-1

			if err := srv.sendItemChunk(stream, item, chunk, chunkIndex, totalChunks, isFirstChunk, isLastChunk); err != nil {
				slog.Warn("GetSteamedVaults: failed to sendItemChunk", slog.Any("error", err))
				return err
			}
		}
	}

	slog.Info("GetSteamedVaults:success")

	return nil
}

func (srv *GophKeepergRPCServer) sendItemChunk(
	stream GophKeeper_GetSteamedVaultsServer,
	item *VaultItem,
	chunk []byte,
	chunkIndex int,
	totalChunks int,
	isFirstChunk bool,
	isLastChunk bool,
) error {
	// Create a copy of the item without encrypted data for the chunk message
	itemWithoutData := VaultItem_builder{
		Id:            proto.Int32(item.GetId()),
		UserId:        proto.Int32(item.GetUserId()),
		Name:          proto.String(item.GetName()),
		DataType:      proto.String(item.GetDataType()),
		EncryptedData: nil, // Data is sent separately in chunks
		MetaData:      item.GetMetaData(),
		CreatedAt:     proto.String(item.GetCreatedAt()),
		UpdatedAt:     proto.String(item.GetUpdatedAt()),
		Active:        proto.Bool(item.GetActive()),
		Version:       proto.Int32(item.GetVersion()),
	}.Build()

	vaultChunk := VaultItemChunk_builder{
		Item:               itemWithoutData,
		EncryptedDataChunk: chunk,
		ChunkIndex:         proto.Int32(int32(chunkIndex)),
		TotalChunks:        proto.Int32(int32(totalChunks)),
		IsFirstChunk:       proto.Bool(isFirstChunk),
		IsLastChunk:        proto.Bool(isLastChunk),
	}.Build()

	return stream.Send(StreamVaultsResponse_builder{ItemChunk: vaultChunk}.Build())
}

// ChangePassword is used to change user password.
func (srv GophKeepergRPCServer) ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	changeReq := &models.ChangePasswordRequest{
		Login:           req.GetLogin(),
		CurrentPassword: req.GetCurrentPassword(),
		NewPassword:     req.GetNewPassword(),
	}

	_, err := srv.userService.ChangePassword(ctx, userID, changeReq)
	if err != nil {
		slog.Warn("ChangePassword:failed with error", slog.Any("error", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ChangePasswordResponse_builder{Success: proto.Bool(true)}.Build(), nil
}

// UpdateVault is used update specific vault with new data.
func (srv GophKeepergRPCServer) UpdateVault(ctx context.Context, req *UpdateVaultRequest) (*UpdateVaultResponse, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	err := srv.vaultService.Update(ctx, int(req.GetId()), userID, req.GetName(), req.GetEncryptedData())
	if err != nil {
		if errors.Is(err, service.ErrVaultOwnerMismatchUpdate) {
			return nil, status.Errorf(codes.Unauthenticated, "Access to forbidden data")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return UpdateVaultResponse_builder{Success: proto.Bool(true)}.Build(), nil
}
