package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/go-playground/validator/v10"
)

func (s *VaultService) Save(ctx context.Context, req *models.Vault) (*models.Vault, error) {
	const op = "VaultService:Save"
	slog.Info(op, slog.Any("user_id", req.UserID))

	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(req)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return &models.Vault{}, validationErrors
		}

		return &models.Vault{}, fmt.Errorf("validation failed: %w", err)
	}

	err = s.repo.Insert(ctx, req)
	if err != nil {
		slog.Warn(op, slog.Any("error", err))
		return &models.Vault{}, err
	}

	return &models.Vault{}, nil
}
