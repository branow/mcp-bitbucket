package schema_test

import (
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPositiveValidator(t *testing.T) {
	schema := schema.Int().Must(schema.Positive())

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"positive number", "42", true},
		{"one", "1", true},
		{"large positive", "999999", true},
		{"zero", "0", false},
		{"negative", "-5", false},
		{"large negative", "-999999", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidator(t, schema, tt.input, tt.valid, "expected positive integer")
		})
	}
}

func TestNonNegativeValidator(t *testing.T) {
	schema := schema.Int().Must(schema.NonNegative())

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"positive number", "42", true},
		{"zero", "0", true},
		{"large positive", "999999", true},
		{"negative", "-1", false},
		{"large negative", "-999999", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidator(t, schema, tt.input, tt.valid, "expected non-negative integer")
		})
	}
}

func testValidator[T comparable](t *testing.T, schema schema.Required[T], in string, valid bool, errorContains string) {
	t.Helper()
	_, err := schema.Parse(in)
	if valid {
		require.NoError(t, err)
	} else {
		assert.ErrorContains(t, err, errorContains)
	}
}
