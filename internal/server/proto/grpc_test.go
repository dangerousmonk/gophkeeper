package proto

import (
	"context"
	"errors"
	"testing"
	"time"

	authm "github.com/dangerousmonk/gophkeeper/internal/auth/mocks"
	"github.com/dangerousmonk/gophkeeper/internal/middleware"
	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/dangerousmonk/gophkeeper/internal/service"
	"github.com/dangerousmonk/gophkeeper/internal/service/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withUserIDCtx(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, middleware.UserIDContextKey, userID)
}

func TestUserRegister(t *testing.T) {
	expectedUserID := 123
	expectedLogin := "registerUser"
	expectedToken := "registerToken"

	cases := []struct {
		name      string
		req       RegisterUserRequest
		userStub  func(s *mocks.MockUserHandler)
		authStub  func(s *authm.MockAuthenticator)
		wantError bool
	}{
		{
			name: "success",
			req:  RegisterUserRequest{Login: expectedLogin, Password: "foobar"},
			userStub: func(r *mocks.MockUserHandler) {
				r.EXPECT().
					Register(gomock.Any(), &models.RegisterUserRequest{Login: expectedLogin, Password: "foobar"}).
					Times(1).
					Return(&models.RegisterUserResponse{Login: expectedLogin, Token: expectedToken, ID: expectedUserID, Success: true}, nil)
			},
			authStub: func(r *authm.MockAuthenticator) {
				r.EXPECT().
					CreateToken(expectedUserID, time.Hour*1).Times(1).
					Return(expectedToken, nil)
			},
			wantError: false,
		},
		{
			name: "service_error",
			req:  RegisterUserRequest{Login: expectedLogin, Password: "foobar"},
			userStub: func(r *mocks.MockUserHandler) {
				r.EXPECT().
					Register(gomock.Any(), &models.RegisterUserRequest{Login: expectedLogin, Password: "foobar"}).
					Times(1).
					Return(&models.RegisterUserResponse{Success: false}, errors.New("some error"))
			},
			authStub: func(r *authm.MockAuthenticator) {
				r.EXPECT().
					CreateToken(expectedUserID, time.Hour*1).Times(0)
			},
			wantError: true,
		},
		{
			name: "authenticator_error",
			req:  RegisterUserRequest{Login: expectedLogin, Password: "foobar"},
			userStub: func(r *mocks.MockUserHandler) {
				r.EXPECT().
					Register(gomock.Any(), &models.RegisterUserRequest{Login: expectedLogin, Password: "foobar"}).
					Times(1).
					Return(&models.RegisterUserResponse{Login: expectedLogin, Token: expectedToken, ID: expectedUserID, Success: true}, nil)
			},
			authStub: func(r *authm.MockAuthenticator) {
				r.EXPECT().
					CreateToken(expectedUserID, time.Hour*1).Times(1).
					Return("", errors.New("some error"))
			},
			wantError: true,
		},
	}

	for i := range cases {
		t.Run(cases[i].name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockUserHandler(ctrl)
			authenticator := authm.NewMockAuthenticator(ctrl)

			srv := GophKeepergRPCServer{
				userService:   handler,
				authenticator: authenticator,
			}

			cases[i].userStub(handler)
			cases[i].authStub(authenticator)

			resp, err := srv.RegisterUser(context.Background(), &cases[i].req)

			if cases[i].wantError {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.True(t, resp.Success)

				assert.Equal(t, uint64(expectedUserID), resp.Id)
				assert.Equal(t, expectedLogin, resp.Login)
				assert.Equal(t, expectedToken, resp.Token)
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	expectedLogin := "testuser"
	expectedToken := "test-token-123"

	cases := []struct {
		name      string
		req       LoginUserRequest
		userStub  func(s *mocks.MockUserHandler)
		wantError bool
	}{
		{
			name: "success",
			req:  LoginUserRequest{Login: expectedLogin, Password: "foobar"},
			userStub: func(r *mocks.MockUserHandler) {
				r.EXPECT().
					Login(gomock.Any(), expectedLogin, "foobar", gomock.Any()).
					Times(1).
					Return(expectedToken, nil)
			},
			wantError: false,
		},
		{
			name: "service_error",
			req:  LoginUserRequest{Login: expectedLogin, Password: "foobar"},
			userStub: func(r *mocks.MockUserHandler) {
				r.EXPECT().
					Login(gomock.Any(), expectedLogin, "foobar", gomock.Any()).
					Times(1).
					Return("", errors.New("some error"))
			},
			wantError: true,
		},
	}

	for i := range cases {
		t.Run(cases[i].name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockUserHandler(ctrl)
			authenticator := authm.NewMockAuthenticator(ctrl)

			srv := GophKeepergRPCServer{
				userService:   handler,
				authenticator: authenticator,
			}

			cases[i].userStub(handler)

			resp, err := srv.LoginUser(context.Background(), &cases[i].req)

			if cases[i].wantError {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.True(t, resp.Success)
				assert.Equal(t, expectedToken, resp.Token)
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	login := "admin"
	oldPass := "foobar"
	newPass := "testtest"
	userID := 123

	cases := []struct {
		name      string
		req       ChangePasswordRequest
		userStub  func(s *mocks.MockUserHandler)
		withCtx   bool
		wantError bool
	}{
		{
			name: "success",
			req:  ChangePasswordRequest{Login: login, CurrentPassword: oldPass, NewPassword: newPass},
			userStub: func(r *mocks.MockUserHandler) {
				r.EXPECT().
					ChangePassword(gomock.Any(), userID, &models.ChangePasswordRequest{Login: login, CurrentPassword: oldPass, NewPassword: newPass}).
					Times(1).
					Return(&models.ChangePasswordResponse{Success: true}, nil)
			},
			withCtx:   true,
			wantError: false,
		},
		{
			name: "no_user",
			req:  ChangePasswordRequest{Login: login, CurrentPassword: oldPass, NewPassword: newPass},
			userStub: func(r *mocks.MockUserHandler) {
				r.EXPECT().
					ChangePassword(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			withCtx:   false,
			wantError: true,
		},
		{
			name: "service_error",
			req:  ChangePasswordRequest{Login: login, CurrentPassword: oldPass, NewPassword: newPass},
			userStub: func(r *mocks.MockUserHandler) {
				r.EXPECT().
					ChangePassword(gomock.Any(), userID, &models.ChangePasswordRequest{Login: login, CurrentPassword: oldPass, NewPassword: newPass}).
					Times(1).
					Return(&models.ChangePasswordResponse{Success: false}, errors.New("some error"))
			},
			withCtx:   true,
			wantError: true,
		},
	}

	for i := range cases {
		t.Run(cases[i].name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockUserHandler(ctrl)
			authenticator := authm.NewMockAuthenticator(ctrl)

			srv := GophKeepergRPCServer{
				userService:   handler,
				authenticator: authenticator,
			}

			cases[i].userStub(handler)

			ctx := context.Background()

			if cases[i].withCtx {
				ctx = withUserIDCtx(context.Background(), userID)
			}

			resp, err := srv.ChangePassword(ctx, &cases[i].req)

			if cases[i].wantError {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestUpdateVault(t *testing.T) {
	vaultID := 123
	name := "new name"
	encryptedData := []byte("\xc96b541e13643d2b4c4df43")
	userID := 123

	cases := []struct {
		name      string
		req       UpdateVaultRequest
		vaultStub func(s *mocks.MockVaultHandler)
		withCtx   bool
		wantError bool
	}{
		{
			name: "success",
			req:  UpdateVaultRequest{Id: int32(vaultID), Name: name, EncryptedData: encryptedData},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Update(gomock.Any(), vaultID, userID, name, encryptedData).
					Times(1).
					Return(nil)
			},
			withCtx:   true,
			wantError: false,
		},
		{
			name: "no_user",
			req:  UpdateVaultRequest{Id: int32(vaultID), Name: name, EncryptedData: encryptedData},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Update(gomock.Any(), vaultID, userID, name, encryptedData).Times(0)
			},
			withCtx:   false,
			wantError: true,
		},
		{
			name: "owner_missmatch",
			req:  UpdateVaultRequest{Id: int32(vaultID), Name: name, EncryptedData: encryptedData},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Update(gomock.Any(), vaultID, userID, name, encryptedData).
					Times(1).
					Return(service.ErrVaultOwnerMismatchUpdate)
			},
			withCtx:   true,
			wantError: true,
		},
		{
			name: "service_error",
			req:  UpdateVaultRequest{Id: int32(vaultID), Name: name, EncryptedData: encryptedData},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Update(gomock.Any(), vaultID, userID, name, encryptedData).
					Times(1).
					Return(errors.New("some error"))
			},
			withCtx:   true,
			wantError: true,
		},
	}

	for i := range cases {
		t.Run(cases[i].name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockVaultHandler(ctrl)
			authenticator := authm.NewMockAuthenticator(ctrl)

			srv := GophKeepergRPCServer{
				vaultService:  handler,
				authenticator: authenticator,
			}

			cases[i].vaultStub(handler)

			ctx := context.Background()

			if cases[i].withCtx {
				ctx = withUserIDCtx(context.Background(), userID)
			}

			resp, err := srv.UpdateVault(ctx, &cases[i].req)

			if cases[i].wantError {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestDeactivateVault(t *testing.T) {
	vaultID := 99
	userID := 12345

	cases := []struct {
		name      string
		req       DeactivateVaultRequest
		vaultStub func(s *mocks.MockVaultHandler)
		withCtx   bool
		wantError bool
	}{
		{
			name: "success",
			req:  DeactivateVaultRequest{SecretId: int32(vaultID)},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Deactivate(gomock.Any(), userID, vaultID).
					Times(1).
					Return(nil)
			},
			withCtx:   true,
			wantError: false,
		},
		{
			name: "no_user",
			req:  DeactivateVaultRequest{SecretId: int32(vaultID)},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Deactivate(gomock.Any(), userID, vaultID).Times(0)
			},
			withCtx:   false,
			wantError: true,
		},
		{
			name: "service_error",
			req:  DeactivateVaultRequest{SecretId: int32(vaultID)},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Deactivate(gomock.Any(), userID, vaultID).
					Times(1).
					Return(errors.New("some error"))
			},
			withCtx:   true,
			wantError: true,
		},
	}

	for i := range cases {
		t.Run(cases[i].name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockVaultHandler(ctrl)
			authenticator := authm.NewMockAuthenticator(ctrl)

			srv := GophKeepergRPCServer{
				vaultService:  handler,
				authenticator: authenticator,
			}

			cases[i].vaultStub(handler)

			ctx := context.Background()

			if cases[i].withCtx {
				ctx = withUserIDCtx(context.Background(), userID)
			}

			resp, err := srv.DeactivateVault(ctx, &cases[i].req)

			if cases[i].wantError {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestSaveVault(t *testing.T) {
	name := "first vault"
	encryptedData := []byte("\xc96b541e13643d2b4c4df43")
	userID := 123

	cases := []struct {
		name      string
		req       SaveVaultRequest
		vaultStub func(s *mocks.MockVaultHandler)
		withCtx   bool
		wantError bool
	}{
		{
			name: "success",
			req:  SaveVaultRequest{Name: name, DataType: "credentials", EcryptedData: encryptedData},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, vault *models.Vault) (*models.Vault, error) {
						assert.Equal(t, userID, vault.UserID)
						assert.Equal(t, name, vault.Name)
						assert.Equal(t, models.DataType("credentials"), vault.DataType)
						assert.Equal(t, encryptedData, vault.EncryptedData)
						assert.NotNil(t, vault.MetaData)

						return vault, nil
					})
			},
			withCtx:   true,
			wantError: false,
		},
		{
			name: "no_user",
			req:  SaveVaultRequest{Name: name, DataType: "credentials", EcryptedData: encryptedData},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(0)
			},
			withCtx:   false,
			wantError: true,
		},
		{
			name: "service_error",
			req:  SaveVaultRequest{Name: name, DataType: "credentials", EcryptedData: encryptedData},
			vaultStub: func(r *mocks.MockVaultHandler) {
				r.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&models.Vault{}, errors.New("some error"))
			},
			withCtx:   true,
			wantError: true,
		},
	}

	for i := range cases {
		t.Run(cases[i].name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockVaultHandler(ctrl)
			authenticator := authm.NewMockAuthenticator(ctrl)

			srv := GophKeepergRPCServer{
				vaultService:  handler,
				authenticator: authenticator,
			}

			cases[i].vaultStub(handler)

			ctx := context.Background()

			if cases[i].withCtx {
				ctx = withUserIDCtx(context.Background(), userID)
			}

			resp, err := srv.SaveVault(ctx, &cases[i].req)

			if cases[i].wantError {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.True(t, resp.Success)
			}
		})
	}
}
