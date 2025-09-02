package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTAuthenticatorCreateToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		duration  time.Duration
		secretKey string
		wantErr   bool
		errType   error
	}{
		{
			name:      "valid_token_creation",
			userID:    123,
			duration:  time.Hour,
			secretKey: "dzb069fe533c433ab1f0c822dba31129",
			wantErr:   false,
		},
		{
			name:      "empty_secret_key",
			userID:    123,
			duration:  time.Hour,
			secretKey: "",
			wantErr:   true,
		},
		{
			name:      "short_secret_key",
			userID:    123,
			duration:  time.Hour,
			secretKey: "short",
			wantErr:   true,
		},
		{
			name:      "negative_duration",
			userID:    123,
			duration:  -time.Hour,
			secretKey: "dzb069fe533c433ab1f0c822dba31129",
			wantErr:   true,
		},
		{
			name:      "zero_duration",
			userID:    123,
			duration:  0,
			secretKey: "dzb069fe533c433ab1f0c822dba31129",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewJWTAuthenticator(tt.secretKey)
			if err != nil && !tt.wantErr {
				t.Fatalf("NewJWTAuthenticator() unexpected error = %v", err)
			}
			if err != nil && tt.wantErr {
				return
			}

			token, err := auth.CreateToken(tt.userID, tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("CreateToken() returned empty token")
			}
		})
	}
}

func TestJWTAuthenticatorValidateToken(t *testing.T) {
	validSecret := "dzb069fe533c433ab1f0c822dba31129"
	auth, err := NewJWTAuthenticator(validSecret)
	if err != nil {
		t.Fatalf("NewJWTAuthenticator() error = %v", err)
	}

	// Create a valid token
	validToken, err := auth.CreateToken(123, time.Hour)
	if err != nil {
		t.Fatalf("CreateToken() error = %v", err)
	}

	// Create authenticator with different secret
	diffAuth, err := NewJWTAuthenticator("different-32-char-secret-key-now")
	if err != nil {
		t.Fatalf("NewJWTAuthenticator() error = %v", err)
	}

	tests := []struct {
		name       string
		token      string
		auth       Authenticator
		wantErr    bool
		wantUserID int
	}{
		{
			name:       "valid_token",
			token:      validToken,
			auth:       auth,
			wantErr:    false,
			wantUserID: 123,
		},
		{
			name:    "empty_token",
			token:   "",
			auth:    auth,
			wantErr: true,
		},
		{
			name:    "malformed_token",
			token:   "malformed.token.here",
			auth:    auth,
			wantErr: true,
		},
		{
			name:    "token_with_different_secret",
			token:   validToken,
			auth:    diffAuth,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := tt.auth.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims.UserID != tt.wantUserID {
					t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, tt.wantUserID)
				}
			}
		})
	}
}

func TestNewClaims(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		duration time.Duration
		wantErr  bool
		errType  error
	}{
		{
			name:     "valid_positive_duration",
			userID:   123,
			duration: time.Hour,
			wantErr:  false,
		},
		{
			name:     "zero_duration",
			userID:   123,
			duration: 0,
			wantErr:  true,
			errType:  errNegativeDuration,
		},
		{
			name:     "negative_duration",
			userID:   123,
			duration: -time.Hour,
			wantErr:  true,
			errType:  errNegativeDuration,
		},
		{
			name:     "positive_small_duration",
			userID:   123,
			duration: time.Millisecond,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := NewClaims(tt.userID, tt.duration)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClaims() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != tt.errType {
					t.Errorf("NewClaims() error = %v, wantErr %v", err, tt.errType)
				}
			} else {
				if claims.UserID != tt.userID {
					t.Errorf("NewClaims() userID = %v, want %v", claims.UserID, tt.userID)
				}
				if claims.ExpiresAt == nil {
					t.Error("NewClaims() ExpiresAt should not be nil")
				}
			}
		})
	}
}

func TestNewJWTAuthenticator(t *testing.T) {
	tests := []struct {
		name       string
		secretKey  string
		wantErr    bool
		errMessage string
	}{
		{
			name:      "valid_32_len_key",
			secretKey: "dzb069fe533c433ab1f0c822dba31129",
			wantErr:   false,
		},
		{
			name:      "valid_longer_than_32",
			secretKey: "dzb069fe533c433ab1f0c822dba31129dzb069fe533c433ab1f0c822dba31129",
			wantErr:   false,
		},
		{
			name:       "empty_key",
			secretKey:  "",
			wantErr:    true,
			errMessage: "invalid secretKey len",
		},
		{
			name:       "short_key_len_31",
			secretKey:  "dzb069fe533c433ab1f0c822dba3112",
			wantErr:    true,
			errMessage: "invalid secretKey len",
		},
		{
			name:       "short_key",
			secretKey:  "a",
			wantErr:    true,
			errMessage: "invalid secretKey len",
		},
		{
			name:      "with_special_characters",
			secretKey: "32-char-key-with-!@#$%^&*()-_=+a",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewJWTAuthenticator(tt.secretKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewJWTAuthenticator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Error("NewJWTAuthenticator() expected error, got nil")
					return
				}
				if err.Error() != tt.errMessage {
					t.Errorf("NewJWTAuthenticator() error message = %v, want %v", err.Error(), tt.errMessage)
				}
			} else {
				if err != nil {
					t.Errorf("NewJWTAuthenticator() unexpected error = %v", err)
					return
				}
				if auth == nil {
					t.Error("NewJWTAuthenticator() returned nil authenticator")
				}

			}
		})
	}
}

func TestClaimsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		claims      *Claims
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				},
				UserID: 123,
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "expired_token",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
				},
				UserID: 123,
			},
			wantErr:     true,
			expectedErr: ErrExpiredToken,
		},
		{
			name: "token_expired_now",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now),
				},
				UserID: 123,
			},
			wantErr:     true,
			expectedErr: ErrExpiredToken,
		},
		{
			name: "token_with_nil_ExpiresAt",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: nil,
				},
				UserID: 123,
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "token_expiring_in_1_sec",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Second)),
				},
				UserID: 123,
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "token_expired_1_sec_ago",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(-time.Second)),
				},
				UserID: 123,
			},
			wantErr:     true,
			expectedErr: ErrExpiredToken,
		},
		{
			name: "token with_zero_time",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Time{}),
				},
				UserID: 123,
			},
			wantErr:     true,
			expectedErr: ErrExpiredToken,
		},
		{
			name: "token_only_user_id_no_claims",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{},
				UserID:           123,
			},
			wantErr:     false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.Valid()

			if (err != nil) != tt.wantErr {
				t.Errorf("Claims.Valid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != tt.expectedErr {
					t.Errorf("Claims.Valid() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("Claims.Valid() unexpected error = %v", err)
				}
			}
		})
	}
}
