package user

import (
	"errors"

	"antifraud/internal/errorz"
	"antifraud/internal/models"
	"antifraud/internal/transport/http/dto"
	"github.com/gofiber/fiber/v2"
)

// Create handles the creation of a user
//
//	@Summary		Create a new user
//	@Description	Creates a new user
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UserCreateRequest	true	"User data"
//	@Success		201		{object}	dto.UserCreateResponse
//	@Failure		400		{object}	dto.HTTPStatus	"Invalid request body"
//	@Failure		409		{object}	dto.HTTPStatus	"User already exists"
//	@Failure		500		{object}	dto.HTTPStatus	"Internal server error"
//	@Router			/user [post]
func (h *handler) Create(c *fiber.Ctx) error {
	var body dto.UserCreateRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&dto.HTTPStatus{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		})
	}

	if err := h.validator.Struct(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&dto.HTTPStatus{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		})
	}

	u := &models.User{
		Username: body.Username,
		Age:      body.Age,
	}
	user, err := h.service.Create(c.UserContext(), u)
	if err != nil {
		if errors.Is(err, errorz.UserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(&dto.HTTPStatus{
				Code:    fiber.StatusConflict,
				Message: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(&dto.HTTPStatus{
			Code:    fiber.StatusInternalServerError,
			Message: errorz.InternalServerError.Error(),
		})
	}

	resp := &dto.UserCreateResponse{
		ID:       user.ID,
		Username: user.Username,
		Age:      user.Age,
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}
