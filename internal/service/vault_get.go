package service

import (
	"context"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

func (s *VaultService) GetByUser(ctx context.Context, userID int) ([]models.Vault, error) {
	vaults, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return []models.Vault{}, err
	}
	return vaults, nil
}
