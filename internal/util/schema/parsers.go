package schema

import (
	"fmt"
	"strconv"
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
