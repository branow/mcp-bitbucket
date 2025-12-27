package schema

import (
	"fmt"
	"strings"
)

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
