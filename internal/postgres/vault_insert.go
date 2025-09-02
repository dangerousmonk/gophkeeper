package postgres

import (
	"context"
	"time"

	"github.com/dangerousmonk/gophkeeper/internal/models"
)

func (r *vaultRepository) Insert(ctx context.Context, v *models.Vault) error {
	const query = `INSERT INTO vault (user_id, name, data_type, encrypted_data, meta_data) VALUES ($1, $2, $3, $4, $5)`
	const timeout = time.Second * 10 // for large data

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, v.UserID, v.Name, v.DataType, v.EncryptedData, v.MetaData)

	if err != nil {
		return err
	}

	return nil
}
