package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/postgres"
	"github.com/dangerousmonk/gophkeeper/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoUserWithLogin    = errors.New("no user with this login")
)

// Login checks if user with given credentials exists in the system and returns access token.
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (s *UserService) Login(
	ctx context.Context,
	login string,
	password string,
	auth utils.Authenticator,
) (string, error) {
	const op = "UserService:Login"
	user, err := s.repo.Get(ctx, login)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			slog.Warn(op, slog.Any("error", err))
			return "", fmt.Errorf("%s %w", op, ErrNoUserWithLogin)
		}
		slog.Error(op, slog.Any("error", err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := s.encryptor.CheckPassword(password, user.PasswordHash); err != nil {
		slog.Warn(op, slog.Any("error", err))
		return "", fmt.Errorf("%s %w", op, ErrInvalidCredentials)
	}

	token, err := auth.CreateToken(user.ID, time.Hour*1)
	if err != nil {
		slog.Warn(op, slog.Any("error", err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	slog.Info(op + " user logged in successfully")

	return token, nil
}
