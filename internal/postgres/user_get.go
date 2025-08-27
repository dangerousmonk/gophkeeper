package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// User returns user by login If found, error otherwise
func (r *userRepository) Get(ctx context.Context, login string) (models.User, error) {
	const op = "Repository:UserGet"
	const timeout = time.Second * 2

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	stmt, err := r.db.Prepare("SELECT id, login, password FROM users WHERE login = $1 AND active IS true")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, login)

	var user models.User
	err = row.Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
