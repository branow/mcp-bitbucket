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

func TestNotBlankValidator(t *testing.T) {
	schema := schema.String().Must(schema.NotBlank())

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"regular string", "hello", true},
		{"string with spaces", "hello world", true},
		{"string with leading spaces", "  hello", true},
		{"string with trailing spaces", "hello  ", true},
		{"empty string", "", false},
		{"only spaces", "   ", false},
		{"only tabs", "\t\t", false},
		{"only newlines", "\n\n", false},
		{"mixed whitespace", " \t\n ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidator(t, schema, tt.input, tt.valid, "expected non-blank string")
		})
	}
}

func TestInValidator_String(t *testing.T) {
	schema := schema.String().Must(schema.In("oauth", "basic"))

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid oauth", "oauth", true},
		{"valid basic", "basic", true},
		{"invalid uppercase", "OAuth", false},
		{"invalid other", "token", false},
		{"invalid empty", "", false},
		{"invalid partial", "oa", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidator(t, schema, tt.input, tt.valid, "expected one of")
		})
	}
}

func TestInValidator_Int(t *testing.T) {
	schema := schema.Int().Must(schema.In(1, 2, 3, 5, 8, 13))

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid 1", "1", true},
		{"valid 5", "5", true},
		{"valid 13", "13", true},
		{"invalid 0", "0", false},
		{"invalid 4", "4", false},
		{"invalid 10", "10", false},
		{"invalid negative", "-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidator(t, schema, tt.input, tt.valid, "expected one of")
		})
	}
}

func TestNotEmptyValidator(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"single item", "item", true},
		{"multiple items", "one;two;three", true},
		{"empty string becomes empty list item", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := schema.List(";").Must(schema.NotEmpty[string]())
			_, err := schema.Parse(tt.input)
			if tt.valid {
				require.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, "expected not empty list")
			}
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
