package service

import (
	"context"

	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/dangerousmonk/gophkeeper/internal/postgres"
)

type VaultService struct {
	repo postgres.VaultRepository
}

func NewVaultService(repo postgres.VaultRepository) *VaultService {
	return &VaultService{
		repo: repo,
	}
}

// VaultHandler defines the contract for vault operations.
//
//go:generate mockgen -package mocks -source vault.go -destination ./mocks/mock_vault_handler.go VaultHandler
type VaultHandler interface {
	// Save is used to insert new Vault record
	Save(ctx context.Context, req *models.Vault) (*models.Vault, error)
	// Deactivate is used to soft delete specific vault record
	Deactivate(ctx context.Context, userID, id int) error
	// GetByUser retrives all active vault records saved by specific user
	GetByUser(ctx context.Context, userID int) ([]models.Vault, error)
	// Update is uses to update existing active vault record with new data
	Update(ctx context.Context, id, userID int, name string, encryptedData []byte) error
}

var _ VaultHandler = (*VaultService)(nil)
