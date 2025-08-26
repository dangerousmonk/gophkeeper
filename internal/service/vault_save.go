package service

import (
	"context"
	"log/slog"

	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/go-playground/validator/v10"
)

func (s *VaultService) Save(ctx context.Context, req *models.Vault) (*models.Vault, error) {
	slog.Info("service:Save", slog.Any("user_id", req.UserID))
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(req)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		return &models.Vault{}, errors
	}

	err = s.repo.Insert(ctx, req)
	if err != nil {
		return &models.Vault{}, err
	}
	return &models.Vault{}, nil
}
