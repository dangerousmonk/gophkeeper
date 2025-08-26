package service

import (
	"context"

	"github.com/dangerousmonk/gophkeeper/internal/encryption"
	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/go-playground/validator/v10"
)

func (s *UserService) Register(ctx context.Context, req *models.RegisterUserRequest) (*models.RegisterUserResponse, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(req)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		return &models.RegisterUserResponse{}, errors
	}

	hashedPassword, err := encryption.HashPassword(req.Password)
	if err != nil {
		return &models.RegisterUserResponse{}, err
	}
	req.HashedPassword = hashedPassword

	userID, err := s.repo.Create(ctx, req)
	if err != nil {
		return &models.RegisterUserResponse{}, err
	}
	return &models.RegisterUserResponse{ID: userID, Login: req.Login, Sucess: true}, nil
}
