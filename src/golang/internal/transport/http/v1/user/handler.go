package user

import (
	"context"

	"antifraud/internal/models"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type service interface {
	Create(ctx context.Context, u *models.User) (*models.User, error)
}

type handler struct {
	service   service
	validator *validator.Validate
}

func New(s service) *handler {
	return &handler{
		service:   s,
		validator: validator.New(),
	}
}

func (h *handler) Setup(router fiber.Router) {
	router.Post("", h.Create)
}
