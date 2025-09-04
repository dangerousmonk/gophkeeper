package auth

import (
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	secretKeySize = 32
)

var (
	ErrExpiredToken     = errors.New("token: has expired")
	errInvalidToken     = errors.New("token: is invalid")
	errInvalidClaims    = errors.New("claims: failed to initialize")
	errNegativeDuration = errors.New("claims: duration is less or equal zero")
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

func newClaims(userID int, duration time.Duration) (*Claims, error) {
	if duration <= 0 {
		return &Claims{}, errNegativeDuration
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
		UserID: userID,
	}

	return claims, nil
}

type jwtAuthenticator struct {
	secretKey string
}

func (claims *Claims) Valid() error {
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return ErrExpiredToken
	}

	return nil
}

//go:generate mockgen -package mocks -source auth.go -destination ./mocks/mock_auth.go Authenticator
type Authenticator interface {
	CreateToken(userID int, duration time.Duration) (string, error)
	ValidateToken(token string) (*Claims, error)
}

// NewJWTAuthenticator is used to initialize new jwtAuthenticator that imlements Authenticator interface.
func NewJWTAuthenticator(secretKey string) (Authenticator, error) {
	if len(secretKey) < secretKeySize {
		return nil, errors.New("invalid secretKey len")
	}

	return &jwtAuthenticator{secretKey}, nil
}

// CreateToken issues new token for user using JWT, for given duration. Returns signed string or an error.
func (auth *jwtAuthenticator) CreateToken(userID int, duration time.Duration) (string, error) {
	claims, err := newClaims(userID, duration)
	if err != nil {
		slog.Error("CreateToken failed create claims", slog.Any("err", err))
		return "", errInvalidClaims
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(auth.secretKey))
}

// ValidateToken validates provided token using JWT. Returns claims or an error.
func (auth *jwtAuthenticator) ValidateToken(token string) (*Claims, error) {
	claims := &Claims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			slog.Error("ValidateToken jwt method is not valid", slog.Bool("ok", ok))
			return nil, errInvalidToken
		}

		return []byte(auth.secretKey), nil
	}

	_, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
