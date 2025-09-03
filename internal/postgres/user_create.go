package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const loginConstraint = "users_login_key"

var ErrUserExists = errors.New("user already exists")

// isUniqueViolation checks if an error is a unique constraint error for given constraint name.
func isUniqueViolation(err error, constraint string) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.Code == pgerrcode.UniqueViolation && pgError.ConstraintName == constraint
	}

	return false
}

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
