package postgres

import (
	"database/sql"
)

// Repositories bundles all repositories that share same database connection.
type Repositories struct {
	User  UserRepository
	Vault VaultRepository
	db    *sql.DB
}

// NewPostgresRepositories creates new repository instances with shared db connection.
func NewPostgresRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:  &userRepository{db: db},
		Vault: &vaultRepository{db: db},
		db:    db,
	}
}
