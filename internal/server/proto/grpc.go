package proto

import (
	"bytes"
	context "context"
	"io"
	"log/slog"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/config"
	"github.com/dangerousmonk/gophkeeper/internal/middleware"
	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/dangerousmonk/gophkeeper/internal/service"
	"github.com/dangerousmonk/gophkeeper/internal/utils"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

const (
	chunkSize = 1024 * 1024 // 1MB chunk
)

// GophKeepergRPCServer Supports all the service methods
type GophKeepergRPCServer struct {
	UnimplementedGophKeeperServer
	userService  *service.UserService
	vaultService *service.VaultService
	cfg          *config.Config
	auth         utils.Authenticator
}

// NewGophKeepergRPCServer creates the ShortenerGRPCServer structure and returns a pointer to freshly created struct.
func NewGophKeepergRPCServer(
	userService *service.UserService,
	vaultService *service.VaultService,
	cfg *config.Config,
	auth utils.Authenticator,
) *GophKeepergRPCServer {
	return &GophKeepergRPCServer{
		cfg:          cfg,
		auth:         auth,
		userService:  userService,
		vaultService: vaultService,
	}
}

// Ping checks the service health
func (srv GophKeepergRPCServer) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := srv.userService.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// RegisterUser is used to register new user
func (srv GophKeepergRPCServer) RegisterUser(ctx context.Context, req *RegisterUserRequest) (*RegisterUserResponse, error) {
	registerReq := &models.RegisterUserRequest{Login: req.Login, Password: req.Password}
	res, err := srv.userService.Register(ctx, registerReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	token, err := srv.auth.CreateToken(res.ID, time.Hour*1)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := RegisterUserResponse{Id: uint64(res.ID), Login: res.Login, Token: token, Success: res.Sucess}
	return &resp, nil
}

// RegisterUser is used to register new user
func (srv GophKeepergRPCServer) LoginUser(ctx context.Context, req *LoginUserRequest) (*LoginUserResponse, error) {
	registerReq := &models.LoginUserRequest{Login: req.Login, Password: req.Password}
	token, err := srv.userService.Login(ctx, registerReq.Login, registerReq.Password, srv.auth)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := LoginUserResponse{Token: token, Success: true}
	return &resp, nil
}

// SaveVault saves data from client to vault
func (srv GophKeepergRPCServer) SaveVault(ctx context.Context, req *SaveVaultRequest) (*SaveVaultResponse, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}
	v := models.Vault{
		UserID:        userID,
		Name:          req.Name,
		DataType:      models.DataType(req.DataType),
		EncryptedData: req.EcryptedData,
		MetaData:      req.MetaData.AsMap(),
	}
	_, err := srv.vaultService.Save(ctx, &v)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &SaveVaultResponse{Success: true}, nil
}

// GetVaults retrives all vaults saved by user
func (srv GophKeepergRPCServer) GetVaults(ctx context.Context, _ *emptypb.Empty) (*GetUserVaultsResponse, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}
	vaults, err := srv.vaultService.GetByUser(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var resp GetUserVaultsResponse
	for _, v := range vaults {
		pbMeta, err := structpb.NewStruct(v.MetaData)
		if err != nil {
			slog.Warn("GetVaults:failedcreating structpb.Struct", slog.Any("error", err))
			continue
		}

		resp.Vaults = append(resp.Vaults, &VaultItem{
			Id:            int32(v.ID),
			UserId:        int32(v.UserID),
			DataType:      string(v.DataType),
			Name:          v.Name,
			EncryptedData: v.EncryptedData,
			MetaData:      pbMeta,
			Version:       int32(v.Version),
			CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     v.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Active:        v.Active,
		})
	}
	return &resp, nil
}

// SaveVault saves data from client to vault
func (srv GophKeepergRPCServer) DeactivateVault(ctx context.Context, req *DeactivateVaultRequest) (*DeactivateVaultResponse, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}
	err := srv.vaultService.Deactivate(ctx, userID, int(req.SecretId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &DeactivateVaultResponse{Success: true}, nil
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
		if err == io.EOF {
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
	vItem := &VaultItem{
		Id:            int32(res.ID),
		UserId:        int32(res.UserID),
		DataType:      string(res.DataType),
		Name:          res.Name,
		EncryptedData: res.EncryptedData,
		Version:       int32(res.Version),
		CreatedAt:     res.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     res.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Active:        res.Active,
	}

	err = stream.SendAndClose(vItem)
	if err != nil {
		slog.Warn("uploadFile:send and close failed", slog.Any("error", err))
		return status.Error(codes.Internal, err.Error())
	}

	slog.Info("uploadFile:success", slog.Int("file_size", fileSize))
	return nil
}

func (srv *GophKeepergRPCServer) GetSteamedVaults(req *emptypb.Empty, stream GophKeeper_GetSteamedVaultsServer) error {
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
	for _, v := range vaults {
		pbMeta, err := structpb.NewStruct(v.MetaData)
		if err != nil {
			slog.Warn("GetSteamedVaults:error creating structpb.Struct", slog.Any("error", err))
			continue
		}

		vaultItems = append(vaultItems, &VaultItem{
			Id:            int32(v.ID),
			UserId:        int32(v.UserID),
			DataType:      string(v.DataType),
			Name:          v.Name,
			EncryptedData: v.EncryptedData,
			MetaData:      pbMeta,
			Version:       int32(v.Version),
			CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     v.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Active:        v.Active,
		})
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
		metadata := &StreamMetadata{
			TotalItems:       int32(totalItems),
			CurrentItemIndex: int32(itemIndex),
			IsFirstItem:      isFirstItem,
			IsLastItem:       isLastItem,
		}

		if err := stream.Send(&StreamVaultsResponse{
			Payload: &StreamVaultsResponse_Metadata{Metadata: metadata},
		}); err != nil {
			slog.Warn("GetSteamedVaults:failed to send metadata", slog.Any("error", err))
			return status.Errorf(codes.Internal, "failed to send metadata: %v", err)
		}

		// Handle chunking of encrypted_data
		encryptedData := item.EncryptedData
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
	itemWithoutData := &VaultItem{
		Id:            item.Id,
		UserId:        item.UserId,
		Name:          item.Name,
		DataType:      item.DataType,
		EncryptedData: nil, // Data is sent separately in chunks
		MetaData:      item.MetaData,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
		Active:        item.Active,
		Version:       item.Version,
	}

	vaultChunk := &VaultItemChunk{
		Item:               itemWithoutData,
		EncryptedDataChunk: chunk,
		ChunkIndex:         int32(chunkIndex),
		TotalChunks:        int32(totalChunks),
		IsFirstChunk:       isFirstChunk,
		IsLastChunk:        isLastChunk,
	}
	return stream.Send(&StreamVaultsResponse{
		Payload: &StreamVaultsResponse_ItemChunk{ItemChunk: vaultChunk},
	})
}

// ChangePassword is used to change user password
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
	return &ChangePasswordResponse{Success: true}, nil
}
