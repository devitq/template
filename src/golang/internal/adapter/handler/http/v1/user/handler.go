package user

import (
	"context"

	"antifraud/internal/domain/models"
	"github.com/gofiber/fiber/v2"
)

type service interface {
	Create(ctx context.Context, u *models.User) (*models.User, error)
}

type validator interface {
	Validate(any) error
}

type handler struct {
	service   service
	validator validator
}

func New(s service, v validator) *handler {
	return &handler{
		service:   s,
		validator: v,
	}
}

func (h *handler) Setup(router fiber.Router) {
	router.Post("", h.Create)
}
