package postgres

import (
	"database/sql"
	"errors"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

// PostgresRepositories bundles all repositories that share same database connection
type PostgresRepositories struct {
	User  UserRepository
	Vault VaultRepository
	db    *sql.DB // or connection pool
}

// NewPostgresRepositories creates new repository instances with shared db connection
func NewPostgresRepositories(db *sql.DB) *PostgresRepositories {
	return &PostgresRepositories{
		User:  &userRepository{db: db},
		Vault: &vaultRepository{db: db},
		db:    db,
	}
}
