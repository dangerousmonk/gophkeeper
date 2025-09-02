package service

import (
	"context"
	"time"
)

func (s *UserService) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	return s.repo.Ping(ctx)
}
