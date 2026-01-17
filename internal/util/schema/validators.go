package schema

import (
	"fmt"
	"slices"
	"strings"
)

// ValidationResult represents the result of validating a value.
//
// It contains the original value and an error indicating whether
// validation succeeded or failed.
type ValidationResult[T any] struct {
	value T
	err   error
}

// Get returns the validated value and the validation error, if any.
//
// If validation failed, the returned error will be non-nil.
func (r *ValidationResult[T]) Get() (T, error) {
	return r.value, r.err
}

// Optional returns the validated value if validation succeeded,
// or the provided fallback value if validation failed.
//
// This is useful when a default value should be used in case of
// validation errors.
func (r *ValidationResult[T]) Optional(fallback T) T {
	if r.err != nil {
		return fallback
	}
	return r.value
}

// Validate applies the provided validators to the given value in order.
//
// Validation stops at the first validator that returns an error.
// The returned ValidationResult contains the original value and
// the first validation error encountered, if any.
func Validate[T any](value T, validators ...Validator[T]) *ValidationResult[T] {
	var err error
	for _, validator := range validators {
		if err = validator(value); err != nil {
			break
		}
	}
	return &ValidationResult[T]{value: value, err: err}
}

// Validator is a function that validates a value.
// It returns an error if the value is invalid, or nil if valid.
type Validator[T any] func(T) error

// Positive returns a Validator that checks if an integer is positive (> 0).
func Positive() Validator[int] {
	return func(i int) error {
		if i <= 0 {
			return fmt.Errorf("expected positive integer (> 0), got: %d", i)
		}
		return nil
	}
}

// NonNegative returns a Validator that checks if an integer is non-negative (>= 0).
func NonNegative() Validator[int] {
	return func(i int) error {
		if i < 0 {
			return fmt.Errorf("expected non-negative integer (>= 0), got: %d", i)
		}
		return nil
	}
}

// NotBlank returns a Validator that checks if a string is not blank.
// A string is considered blank if it is empty or contains only whitespace characters.
func NotBlank() Validator[string] {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("expected non-blank string, got: '%s'", s)
		}
		return nil
	}
}

// In returns a Validator that checks if a string or integer is one of the provided values.
// The comparison is exact - for strings, it is case-sensitive.
func In[T string | int](options ...T) Validator[T] {
	return func(val T) error {
		if !slices.Contains(options, val) {
			return fmt.Errorf("expected one of %v, got: %v", options, val)
		}
		return nil
	}
}

// NotEmpty returns a Validator that checks if a slice is not empty.
// This validator is generic and works with any slice type.
func NotEmpty[T any]() Validator[[]T] {
	return func(val []T) error {
		if len(val) == 0 {
			return fmt.Errorf("expected not empty list")
		}
		return nil
	}
}
