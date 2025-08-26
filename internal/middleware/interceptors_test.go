package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockAuthenticator struct {
	validateTokenFunc func(token string) (*utils.Claims, error)
}

func (m *mockAuthenticator) CreateToken(userID int, duration time.Duration) (string, error) {
	return "", nil
}

func (m *mockAuthenticator) ValidateToken(token string) (*utils.Claims, error) {
	return m.validateTokenFunc(token)
}

func TestAuthUnaryInterceptor(t *testing.T) {
	validToken := "valid.token.here"
	validClaims := &utils.Claims{UserID: 123}

	tests := []struct {
		name         string
		fullMethod   string
		setupContext func() context.Context
		mockAuth     *mockAuthenticator
		wantUserID   interface{}
		wantErr      codes.Code
	}{
		{
			name:       "public_method_skip",
			fullMethod: "/server.GophKeeper/RegisterUser",
			setupContext: func() context.Context {
				return context.Background()
			},
			mockAuth:   &mockAuthenticator{},
			wantUserID: nil,
			wantErr:    codes.OK,
		},
		{
			name:       "missing_metadata",
			fullMethod: "/server.GophKeeper/SaveVault",
			setupContext: func() context.Context {
				return context.Background()
			},
			mockAuth:   &mockAuthenticator{},
			wantUserID: nil,
			wantErr:    codes.Unauthenticated,
		},
		{
			name:       "missing_authorization_header",
			fullMethod: "/server.GophKeeper/SaveVault",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			mockAuth:   &mockAuthenticator{},
			wantUserID: nil,
			wantErr:    codes.Unauthenticated,
		},
		{
			name:       "invalid_token_prefix",
			fullMethod: "/server.GophKeeper/SaveVault",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "InvalidPrefix token",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			mockAuth:   &mockAuthenticator{},
			wantUserID: nil,
			wantErr:    codes.Unauthenticated,
		},
		{
			name:       "invalid_token",
			fullMethod: "/server.GophKeeper/SaveVault",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer invalid.token.here",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			mockAuth: &mockAuthenticator{
				validateTokenFunc: func(token string) (*utils.Claims, error) {
					return nil, errors.New("invalid token")
				},
			},
			wantUserID: nil,
			wantErr:    codes.Unauthenticated,
		},
		{
			name:       "valid_token",
			fullMethod: "/server.GophKeeper/SaveVault",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer " + validToken,
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			mockAuth: &mockAuthenticator{
				validateTokenFunc: func(token string) (*utils.Claims, error) {
					if token == validToken {
						return validClaims, nil
					}
					return nil, errors.New("invalid token")
				},
			},
			wantUserID: validClaims.UserID,
			wantErr:    codes.OK,
		},
		{
			name:       "expired_token",
			fullMethod: "/server.GophKeeper/SaveVault",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer expired.token.here",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			mockAuth: &mockAuthenticator{
				validateTokenFunc: func(token string) (*utils.Claims, error) {
					return nil, utils.ErrExpiredToken
				},
			},
			wantUserID: nil,
			wantErr:    codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AuthUnaryInterceptor(tt.mockAuth)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				// Verify userID was set in context if expected
				if tt.wantUserID != nil {
					userID := ctx.Value(userIDContextKey)
					assert.Equal(t, tt.wantUserID, userID)
				}
				return "response", nil
			}

			resp, err := interceptor(
				tt.setupContext(),
				nil, // request
				&grpc.UnaryServerInfo{FullMethod: tt.fullMethod},
				handler,
			)

			if tt.wantErr != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantErr, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, "response", resp)
			}
		})
	}
}

func TestIsPublicMethod(t *testing.T) {
	tests := []struct {
		method   string
		isPublic bool
	}{
		{"/server.GophKeeper/RegisterUser", true},
		{"/server.GophKeeper/LoginUser", true},
		{"/server.GophKeeper/Ping", true},
		{"/server.GophKeeper/RegisterUser2", false},
		{"/other.Service/Method", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			assert.Equal(t, tt.isPublic, IsPublicMethod(tt.method))
		})
	}
}

func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID int
		wantOK bool
	}{
		{
			name:   "context_ok",
			ctx:    context.WithValue(context.Background(), userIDContextKey, 123),
			wantID: 123,
			wantOK: true,
		},
		{
			name:   "context_string",
			ctx:    context.WithValue(context.Background(), userIDContextKey, "user123"),
			wantID: -1,
			wantOK: false,
		},
		{
			name:   "context_with_different_key",
			ctx:    context.WithValue(context.Background(), "otherKey", 123),
			wantID: -1,
			wantOK: false,
		},
		{
			name:   "empty_context",
			ctx:    context.Background(),
			wantID: -1,
			wantOK: false,
		},
		{
			name:   "nil_context",
			ctx:    nil,
			wantID: -1,
			wantOK: false,
		},
		{
			name:   "context_with_nil_value",
			ctx:    context.WithValue(context.Background(), userIDContextKey, nil),
			wantID: -1,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := UserIDFromContext(tt.ctx)
			assert.Equal(t, tt.wantID, gotID)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}
