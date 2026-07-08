// Package validator wraps go-playground/validator with custom rules.
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator is a singleton wrapper around the go-playground validator instance.
type Validator struct {
	v *validator.Validate
}

// New creates and returns a configured Validator.
func New() *Validator {
	v := validator.New()

	// Register custom validators here.
	// e.g., v.RegisterValidation("hexcolor", validateHexColor)

	return &Validator{v: v}
}

// Validate validates any struct and returns a slice of human-readable errors.
func (val *Validator) Validate(s interface{}) []string {
	err := val.v.Struct(s)
	if err == nil {
		return nil
	}

	var errs []string
	for _, e := range err.(validator.ValidationErrors) {
		errs = append(errs, fieldError(e))
	}
	return errs
}

// ValidateVar validates a single variable against a tag string.
func (val *Validator) ValidateVar(field interface{}, tag string) error {
	return val.v.Var(field, tag)
}

// fieldError turns a single validation error into a readable message.
func fieldError(e validator.FieldError) string {
	field := strings.ToLower(e.Field())
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, e.Tag())
	}
}
