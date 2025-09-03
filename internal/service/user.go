package service

import (
	"context"

	"github.com/dangerousmonk/gophkeeper/internal/auth"
	"github.com/dangerousmonk/gophkeeper/internal/encryption"
	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/dangerousmonk/gophkeeper/internal/postgres"
)

// UserHandler defines the contract for user operations.
type UserHandler interface {
	Register(ctx context.Context, req *models.RegisterUserRequest) (*models.RegisterUserResponse, error)
	Login(ctx context.Context, login, password string, auth auth.Authenticator) (string, error)
	ChangePassword(ctx context.Context, userID int, req *models.ChangePasswordRequest) (*models.ChangePasswordResponse, error)
}

type UserService struct {
	repo      postgres.UserRepository
	encryptor encryption.PasswordEncryptor
}

func NewUserService(repo postgres.UserRepository, encryptor encryption.PasswordEncryptor) *UserService {
	return &UserService{
		repo:      repo,
		encryptor: encryptor,
	}
}

var _ UserHandler = (*UserService)(nil)
