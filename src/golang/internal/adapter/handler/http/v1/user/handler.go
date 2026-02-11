package user

import (
	"context"
	"log/slog"

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
	logger    *slog.Logger
}

func New(s service, v validator, l *slog.Logger) *handler {
	return &handler{
		service:   s,
		validator: v,
		logger:    l,
	}
}

func (h *handler) Setup(router fiber.Router) {
	router.Post("", h.Create)
}
