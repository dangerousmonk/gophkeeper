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

// VaultHandler defines the contract for vault operations
type VaultHandler interface {
	// Save is used to insert new Vault record
	Save(ctx context.Context, req *models.Vault) (*models.Vault, error)
	// Deactivate is used to soft delete specific vault record
	Deactivate(ctx context.Context, userID int, id int) error
	// GetByUser retrives all active vault records saved by specific user
	GetByUser(ctx context.Context, userID int) ([]models.Vault, error)
}

var _ VaultHandler = (*VaultService)(nil)
