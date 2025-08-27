package postgres

import (
	"database/sql"
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
