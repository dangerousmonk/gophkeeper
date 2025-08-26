package postgres

import (
	"context"
	"time"
)

func (r *vaultRepository) Deactivate(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE vault SET active=$1 WHERE id=$2`

	_, err = tx.ExecContext(ctx, query, false, id)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
