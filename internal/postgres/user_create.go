package postgres

import (
	"context"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

func (r *userRepository) Create(ctx context.Context, ru *models.RegisterUserRequest) (int, error) {
	const query = `INSERT INTO users (login, password, last_login_at) VALUES ($1, $2, $3) RETURNING id`
	const timeout = 2

	var userID int
	ctx, cancel := context.WithTimeout(ctx, time.Second*timeout)
	defer cancel()

	tx, err := r.db.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, query, ru.Login, ru.HashedPassword, time.Now()).Scan(&userID)

	if err != nil {
		return -1, err
	}

	if err := tx.Commit(); err != nil {
		return -1, err
	}

	return userID, nil
}
