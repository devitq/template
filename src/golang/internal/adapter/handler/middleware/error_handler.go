package middleware

import (
	"errors"
	"time"

	"antifraud/internal/adapter/handler/response"
	"antifraud/internal/domain/errorz"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type AppError struct {
	Code    string
	Message string
	Status  int
	Details map[string]any
}

func (e *AppError) Error() string {
	return e.Message
}

type ValidationErrors struct {
	FieldErrors []response.FieldError
}

func (e *ValidationErrors) Error() string {
	return "validation failed"
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	errorCode := "INTERNAL_SERVER_ERROR"
	message := "Internal server error"
	var details map[string]any

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		message = fiberErr.Message
	}

	var appErr *AppError
	var valErrs *ValidationErrors

	// Try to get traceId from locals (set by some middleware) or OTEL or generate new one
	traceId := uuid.New().String()
	if traceIdRaw := c.Locals("traceId"); traceIdRaw != nil {
		if tid, ok := traceIdRaw.(string); ok {
			traceId = tid
		}
	} else {
		span := trace.SpanFromContext(c.UserContext())
		if span.SpanContext().IsValid() {
			traceId = span.SpanContext().TraceID().String()
		}
	}

	if errors.As(err, &valErrs) {
		return c.Status(fiber.StatusBadRequest).JSON(response.ValidationError{
			Code:        "VALIDATION_FAILED",
			Message:     "Validation failed",
			TraceId:     traceId,
			Timestamp:   time.Now().UTC(),
			Path:        c.Path(),
			FieldErrors: valErrs.FieldErrors,
		})
	}

	if errors.As(err, &appErr) {
		code = appErr.Status
		errorCode = appErr.Code
		message = appErr.Message
		details = appErr.Details
	} else if errors.Is(err, errorz.UserAlreadyExists) {
		code = fiber.StatusConflict
		errorCode = "USER_ALREADY_EXISTS"
		message = err.Error()
	}

	return c.Status(code).JSON(response.ApiError{
		Code:      errorCode,
		Message:   message,
		TraceId:   traceId,
		Timestamp: time.Now().UTC(),
		Path:      c.Path(),
		Details:   details,
	})
}
