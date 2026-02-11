package user

import (
	"antifraud/internal/domain/repository"
)

type service struct {
	repo repository.UserRepository
}

func New(repo repository.UserRepository) *service {
	return &service{repo}
}
