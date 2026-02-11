package user

import (
	"context"

	"antifraud/internal/infra/database/sqlc/storage"
	"antifraud/internal/domain/errorz"
	"antifraud/internal/domain/models"
	"antifraud/pkg/dberrors"
)

func (r *repo) Create(ctx context.Context, u *models.User) (*models.User, error) {
	newUser, err := r.queries.CreateUser(ctx, &storage.CreateUserParams{
		Username: u.Username,
		Age:      u.Age,
	})
	if err != nil {
		if dberrors.IsUniqueViolation(err) {
			return nil, errorz.UserAlreadyExists
		}
		return nil, err
	}

	return &models.User{
		ID:       newUser.ID,
		Username: newUser.Username,
		Age:      newUser.Age,
	}, nil
}
