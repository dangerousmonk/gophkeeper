package service

import (
	"context"
	"errors"
	"log/slog"
)

var ErrVaultOwnerMismatchUpdate = errors.New("vaultService:this vault belongs to other user")

func (s *VaultService) Update(ctx context.Context, id, userID int, name string, encryptedData []byte) error {
	const op = "VaultService:Update"
	slog.Info(op, slog.Any("id", id))

	vault, err := s.repo.Get(ctx, id)
	if err != nil {
		slog.Warn(op, slog.Any("error", err))
		return err
	}

	if vault.UserID != userID {
		return ErrVaultOwnerMismatchUpdate
	}

	err = s.repo.Update(ctx, id, name, encryptedData)
	if err != nil {
		slog.Warn(op, slog.Any("error", err))
		return err
	}

	return nil
}
