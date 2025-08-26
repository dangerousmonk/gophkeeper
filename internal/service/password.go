package service

import (
	"context"
	"errors"

	"github.com/dangerousmonk/gophkeeper/internal/encryption"
	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/go-playground/validator/v10"
)

var (
	ErrPasswordNotChanged       = errors.New("userService:new password equals old password")
	ErrPasswordEncryptionFailed = errors.New("userService:failed to encrypt password")
)

func (s *UserService) ChangePassword(ctx context.Context, userID int, req *models.ChangePasswordRequest) (*models.ChangePasswordResponse, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(req)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		return &models.ChangePasswordResponse{Success: false}, errors
	}

	if req.CurrentPassword == req.NewPassword {
		return &models.ChangePasswordResponse{Success: false}, ErrPasswordNotChanged
	}

	hashedPassword, err := encryption.HashPassword(req.NewPassword)
	if err != nil {
		return &models.ChangePasswordResponse{Success: false}, ErrPasswordEncryptionFailed
	}

	err = s.repo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		return &models.ChangePasswordResponse{Success: false}, err
	}
	return &models.ChangePasswordResponse{Success: true}, nil
}
