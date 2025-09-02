package postgres

import (
	"context"
	"time"
)

func (r *userRepository) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := r.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
