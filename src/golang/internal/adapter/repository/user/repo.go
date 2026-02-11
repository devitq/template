package user

import (
	"antifraud/internal/domain/repository"
	"antifraud/internal/infra/database/sqlc/storage"
)

var _ repository.UserRepository = (*repo)(nil)

type repo struct {
	queries *storage.Queries
}

func New(queries *storage.Queries) *repo {
	return &repo{
		queries: queries,
	}
}
