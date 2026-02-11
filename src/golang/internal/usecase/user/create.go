package user

import (
	"context"

	"antifraud/internal/domain/models"
)

func (s *service) Create(ctx context.Context, u *models.User) (*models.User, error) {
	return s.repo.Create(ctx, u)
}
