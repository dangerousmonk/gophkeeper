package postgres

import (
	"context"
	"time"
)

func (r *userRepository) UpdatePassword(ctx context.Context, userID int, password string) error {
	const query = `UPDATE users SET password=$1 WHERE id=$2`

	const timeout = time.Second * 2

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, password, userID)
	if err != nil {
		return err
	}

	return nil
}
