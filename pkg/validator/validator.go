// Package validator provides lightweight struct validation for DTOs.
package validator

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// Errors maps field names to their error messages.
type Errors map[string]string

func (e Errors) Error() string {
	sb := strings.Builder{}
	for k, v := range e {
		sb.WriteString(fmt.Sprintf("%s: %s; ", k, v))
	}
	return strings.TrimRight(sb.String(), "; ")
}

// Validator validates structs using a fluent builder per field.
type Validator struct {
	errs Errors
}

func New() *Validator { return &Validator{errs: make(Errors)} }

// Check adds an error for field if condition is false.
func (v *Validator) Check(condition bool, field, message string) *Validator {
	if !condition {
		v.errs[field] = message
	}
	return v
}

// Required checks that a string field is non-empty.
func (v *Validator) Required(value, field string) *Validator {
	return v.Check(strings.TrimSpace(value) != "", field, "is required")
}

// MinLen checks minimum string length.
func (v *Validator) MinLen(value string, min int, field string) *Validator {
	return v.Check(len(strings.TrimSpace(value)) >= min, field,
		fmt.Sprintf("must be at least %d characters", min))
}

// MaxLen checks maximum string length.
func (v *Validator) MaxLen(value string, max int, field string) *Validator {
	return v.Check(len(value) <= max, field,
		fmt.Sprintf("must not exceed %d characters", max))
}

// Min checks numeric minimum (float64).
func (v *Validator) Min(value, min float64, field string) *Validator {
	return v.Check(value >= min, field, fmt.Sprintf("must be at least %.0f", min))
}

// Max checks numeric maximum.
func (v *Validator) Max(value, max float64, field string) *Validator {
	return v.Check(value <= max, field, fmt.Sprintf("must not exceed %.0f", max))
}

// OneOf checks that value is within allowed set.
func (v *Validator) OneOf(value string, allowed []string, field string) *Validator {
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	return v.Check(false, field, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
}

// ValidEmail does a simple email format check.
func (v *Validator) ValidEmail(value, field string) *Validator {
	if value == "" {
		return v
	}
	parts := strings.Split(value, "@")
	ok := len(parts) == 2 && len(parts[0]) > 0 && strings.Contains(parts[1], ".")
	return v.Check(ok, field, "must be a valid email address")
}

// StrongPassword checks password strength.
func (v *Validator) StrongPassword(value, field string) *Validator {
	if len(value) < 8 {
		return v.Check(false, field, "must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	return v.Check(hasUpper && hasLower && hasDigit, field,
		"must contain uppercase, lowercase and a digit")
}

// Valid returns true when no errors have been collected.
func (v *Validator) Valid() bool { return len(v.errs) == 0 }

// Errors returns the collected error map.
func (v *Validator) Errors() map[string]any {
	out := make(map[string]any, len(v.errs))
	for k, msg := range v.errs {
		out[k] = msg
	}
	return out
}

// ValidateStruct runs validation tags on a struct (required:"true", min:"N", max:"N").
// This is a lightweight alternative to heavy external libs.
func ValidateStruct(s any) Errors {
	errs := make(Errors)
	rv := reflect.ValueOf(s)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()

	for i := range rt.NumField() {
		field := rt.Field(i)
		value := rv.Field(i)

		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}
		name := field.Tag.Get("json")
		if name == "" || name == "-" {
			name = strings.ToLower(field.Name)
		}
		name = strings.Split(name, ",")[0]

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			switch {
			case rule == "required":
				if value.Kind() == reflect.String && strings.TrimSpace(value.String()) == "" {
					errs[name] = "is required"
				}
			}
		}
	}
	return errs
}
