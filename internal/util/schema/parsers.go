package schema

import (
	"fmt"
	"strconv"
	"strings"
)

// String creates a Required for string values.
// All strings are considered valid, including empty strings.
func String() Required[string] {
	return NewSchema(func(s string) (string, error) { return s, nil })
}

// Int creates a Required for integer values.
// The input must be a valid integer string that can be parsed by strconv.Atoi.
func Int() Required[int] {
	return NewSchema(func(s string) (int, error) {
		value, err := strconv.Atoi(s)
		if err != nil {
			return value, fmt.Errorf("expected valid integer, got: '%s'", s)
		}
		return value, nil
	})
}

// List creates a Required for string slice values.
// The input is split by the specified delimiter to create a slice of strings.
// If the input is an empty string, it returns an empty slice instead of a slice with one empty string.
func List(delimiter string) Required[[]string] {
	return NewSchema(func(s string) ([]string, error) {
		if s == "" {
			return []string{}, nil
		}
		return strings.Split(s, delimiter), nil
	})
}
