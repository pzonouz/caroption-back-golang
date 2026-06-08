package utils

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewValidate() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	// Register custom validation
	validate.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()

		return strings.TrimSpace(s) != ""
	})

	return validate
}
