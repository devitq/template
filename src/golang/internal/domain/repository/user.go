package repository

import (
	"context"

	"antifraud/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) (*models.User, error)
}
