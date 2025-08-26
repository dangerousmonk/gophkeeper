package postgres

import (
	"context"
	"log/slog"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

func (r *vaultRepository) Insert(ctx context.Context, v *models.Vault) error {
	slog.Info("repo:insert", slog.Any("user_id", v.UserID))
	const query = `INSERT INTO vault (user_id, name, data_type, encrypted_data, meta_data) VALUES ($1, $2, $3, $4, $5)`

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, query, v.UserID, v.Name, v.DataType, v.EncryptedData, v.MetaData)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
