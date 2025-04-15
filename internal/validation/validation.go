package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// TranslateValidationErrors translates validation errors into a list of error messages
func TranslateValidationErrors(err error) []string {
	var validationErrors []string

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			switch e.Tag() {
			case "required":
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("%s is required", e.Field()),
				)
			case "email":
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("%s is not a valid email", e.Value()),
				)
			case "min":
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("%s must be at least %s characters long", e.Field(), e.Param()),
				)
			default:
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("%s is not valid", e.Field()),
				)
			}
		}
	} else {
		validationErrors = append(validationErrors, "Invalid request body")
	}

	return validationErrors
}
