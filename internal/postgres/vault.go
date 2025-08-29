package postgres

import (
	"context"
	"database/sql"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

//go:generate mockgen -package mocks -source vault.go -destination ./mocks/mock_vault_repository.go VaultRepository
type VaultRepository interface {
	// Insert inserts new record into vault table
	Insert(ctx context.Context, v *models.Vault) error
	// GetByUserID retrieves all active vault records saved by certain user
	GetByUserID(ctx context.Context, userID int) ([]models.Vault, error)
	// Deactivate is used to soft delete specific vault record
	Deactivate(ctx context.Context, id int) error
	// Get retrives specific vault record by id
	Get(ctx context.Context, id int) (models.Vault, error)
}

// vaultRepository implements VaultRepository
type vaultRepository struct {
	db *sql.DB
}
