// Package utils internal/utils/validator.go
package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/oklog/ulid/v2"
)

var (
	validate      *validator.Validate
	phoneRegex    = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
)

func init() {
	validate = validator.New()

	// Custom validators
	validate.RegisterValidation("ulid", validateULID)
	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("username", validateUsername)
}

func validateULID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, err := ulid.Parse(value)
	return err == nil
}

func validatePhone(fl validator.FieldLevel) bool {
	return phoneRegex.MatchString(fl.Field().String())
}

func validateUsername(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func FormatValidationErrors(err error) map[string]interface{} {
	errors := make(map[string]interface{})

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errors[field] = fmt.Sprintf("%s is required", field)
			case "min":
				errors[field] = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
			case "max":
				errors[field] = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
			case "email":
				errors[field] = fmt.Sprintf("%s must be a valid email", field)
			case "ulid":
				errors[field] = fmt.Sprintf("%s must be a valid ULID", field)
			case "phone":
				errors[field] = fmt.Sprintf("%s must be a valid phone number", field)
			case "username":
				errors[field] = fmt.Sprintf("%s must be 3-32 alphanumeric characters", field)
			default:
				errors[field] = fmt.Sprintf("%s failed validation", field)
			}
		}
	}

	return errors
}
