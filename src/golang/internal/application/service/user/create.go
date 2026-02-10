package user

import (
	"context"

	"antifraud/internal/models"
)

func (s *service) Create(ctx context.Context, u *models.User) (*models.User, error) {
	return s.repo.Create(ctx, u)
}
