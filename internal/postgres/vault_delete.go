package postgres

import (
	"context"
	"time"
)

func (r *vaultRepository) Deactivate(ctx context.Context, id int) error {
	const query = `UPDATE vault SET active=$1 WHERE id=$2`

	const timeout = time.Second * 2

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, false, id)
	if err != nil {
		return err
	}

	return nil
}
