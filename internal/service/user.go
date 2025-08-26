package service

import (
	"context"

	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/dangerousmonk/gophkeeper/internal/postgres"
	"github.com/dangerousmonk/gophkeeper/internal/utils"
)

// UserHandler defines the contract for user operations
type UserHandler interface {
	Register(ctx context.Context, req *models.RegisterUserRequest) (*models.RegisterUserResponse, error)
	Login(ctx context.Context, login string, password string, auth utils.Authenticator) (string, error)
	ChangePassword(ctx context.Context, userID int, req *models.ChangePasswordRequest) (*models.ChangePasswordResponse, error)
}

type UserService struct {
	repo postgres.UserRepository
}

func NewUserService(repo postgres.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

var _ UserHandler = (*UserService)(nil)
