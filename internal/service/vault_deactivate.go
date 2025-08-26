package service

import (
	"context"
	"errors"
	"log/slog"
)

var (
	ErrVaultOwnerMismatch = errors.New("vaultService:this vault belongs to other user")
)

func (s *VaultService) Deactivate(ctx context.Context, userID, id int) error {
	vault, err := s.repo.Get(ctx, id)
	if err != nil {
		slog.Warn("service:Deactivate", slog.Any("error", err))
		return err
	}

	if vault.UserID != userID {
		return ErrVaultOwnerMismatch
	}
	err = s.repo.Deactivate(ctx, id)
	if err != nil {
		return err
	}
	slog.Warn("service:Deactivate success deactivated", slog.Int("id", id))
	return nil
}
