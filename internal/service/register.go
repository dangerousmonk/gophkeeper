package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/go-playground/validator/v10"
)

func (s *UserService) Register(ctx context.Context, req *models.RegisterUserRequest) (*models.RegisterUserResponse, error) {
	const op = "UserService:Register"

	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(req)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return &models.RegisterUserResponse{Success: false}, validationErrors
		}

		return &models.RegisterUserResponse{Success: false}, fmt.Errorf("validation failed: %w", err)
	}

	hashedPassword, err := s.encryptor.HashPassword(req.Password)
	if err != nil {
		slog.Warn(op, slog.Any("error", err))
		return &models.RegisterUserResponse{}, ErrPasswordEncryptionFailed
	}

	req.HashedPassword = hashedPassword

	userID, err := s.repo.Create(ctx, req)
	if err != nil {
		slog.Warn(op, slog.Any("error", err))
		return &models.RegisterUserResponse{}, err
	}

	slog.Info(op+"user registered", slog.Int("user_id", userID))

	return &models.RegisterUserResponse{ID: userID, Login: req.Login, Success: true}, nil
}
