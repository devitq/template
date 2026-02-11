package validator

import (
	"fmt"
	"reflect"
	"strings"

	"antifraud/internal/adapter/handler/middleware"
	"antifraud/internal/adapter/handler/response"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	// Use JSON tag names for error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: v,
	}
}

func (v *Validator) Validate(s any) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	if errs, ok := err.(validator.ValidationErrors); ok {
		var fieldErrors []response.FieldError
		for _, e := range errs {
			fieldErrors = append(fieldErrors, response.FieldError{
				Field:         e.Field(),
				Issue:         msgForTag(e.Tag(), e.Param()),
				RejectedValue: e.Value(),
			})
		}
		return &middleware.ValidationErrors{
			FieldErrors: fieldErrors,
		}
	}

	return err
}

func msgForTag(tag string, param string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s", param)
	case "max":
		return fmt.Sprintf("Must be at most %s", param)
	}
	return fmt.Sprintf("Validation failed on the %s tag", tag)
}
