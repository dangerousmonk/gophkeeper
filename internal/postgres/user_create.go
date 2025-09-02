package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

const loginConstraint = "users_login_key"

var (
	ErrUserExists = errors.New("user already exists")
)

func (r *userRepository) Create(ctx context.Context, ru *models.RegisterUserRequest) (int, error) {
	const query = `INSERT INTO users (login, password, last_login_at) VALUES ($1, $2, $3) RETURNING id`
	const timeout = time.Second * 2

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var userID int

	err := r.db.QueryRowContext(ctx, query, ru.Login, ru.HashedPassword, time.Now()).Scan(&userID)

	if err != nil {
		if isUniqueViolation(err, loginConstraint) {
			return -1, ErrUserExists
		}
		return -1, err
	}

	return userID, nil
}
