package user

import (
	"antifraud/internal/adapter/handler/http/dto"
	_ "antifraud/internal/adapter/handler/response"
	"antifraud/internal/domain/models"
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
//	@Failure		400		{object}	response.ValidationError	"Validation failed"
//	@Failure		409		{object}	response.ApiError	"User already exists"
//	@Failure		500		{object}	response.ApiError	"Internal server error"
//	@Router			/user [post]
func (h *handler) Create(c *fiber.Ctx) error {
	var body dto.UserCreateRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := h.validator.Validate(body); err != nil {
		return err
	}

	u := &models.User{
		Username: body.Username,
		Age:      body.Age,
	}
	user, err := h.service.Create(c.UserContext(), u)
	if err != nil {
		return err
	}

	resp := &dto.UserCreateResponse{
		ID:       user.ID,
		Username: user.Username,
		Age:      user.Age,
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}
