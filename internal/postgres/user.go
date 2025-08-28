package postgres

import (
	"context"
	"database/sql"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

//go:generate mockgen -package mocks -source user.go -destination ./mocks/mock_user_repository.go UserRepository
type UserRepository interface {
	// Ping checks whether internal storage is up and running
	Ping(ctx context.Context) error
	// Create inserts new user data into database
	Create(ctx context.Context, ru *models.RegisterUserRequest) (int, error)
	// Get retrives user from database by login.Returns error If no user found by this login.
	Get(ctx context.Context, login string) (models.User, error)
	// Update updates user password
	UpdatePassword(ctx context.Context, userID int, password string) error
}

// userRepository implements UserRepository
type userRepository struct {
	db *sql.DB
}
