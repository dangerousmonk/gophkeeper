package postgres

import (
	"context"
	"time"
)

func (r *vaultRepository) Update(ctx context.Context, id int, name string, encryptedData []byte) error {
	const query = `UPDATE vault SET encrypted_data=$1, name=$2, version=version+1, updated_at=$3 WHERE id=$4`

	const timeout = time.Second * 2

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, encryptedData, name, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}
