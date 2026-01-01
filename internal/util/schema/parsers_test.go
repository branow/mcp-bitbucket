package schema_test

import (
	"fmt"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringParser(t *testing.T) {
	schema := schema.String()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"usual string", "123 hello world", "123 hello world"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testParser(t, schema, tt.input, true, tt.expected, "")
		})
	}
}

func TestIntParser(t *testing.T) {
	schema := schema.Int()

	tests := []struct {
		name     string
		input    string
		valid    bool
		expected int
	}{
		{"positive integer", "42", true, 42},
		{"negative integer", "-10", true, -10},
		{"zero", "0", true, 0},
		{"large number", "999999", true, 999999},
		{"overflow", "99999999999999999999", false, 0},
		{"invalid letters", "abc", false, 0},
		{"invalid float", "12.34", false, 0},
		{"invalid mixed", "12abc", false, 0},
		{"empty string", "", false, 0},
		{"only spaces", "   ", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testParser(t, schema, tt.input, tt.valid, tt.expected, fmt.Sprintf("expected valid integer, got: '%v'", tt.input))
		})
	}
}

func TestBoolParser(t *testing.T) {
	schema := schema.Bool()

	tests := []struct {
		name     string
		input    string
		valid    bool
		expected bool
	}{
		{"true lowercase", "true", true, true},
		{"true uppercase", "TRUE", true, true},
		{"true mixed case", "True", true, true},
		{"false lowercase", "false", true, false},
		{"false uppercase", "FALSE", true, false},
		{"false mixed case", "False", true, false},
		{"true as 1", "1", true, true},
		{"false as 0", "0", true, false},
		{"true as t", "t", true, true},
		{"true as T", "T", true, true},
		{"false as f", "f", true, false},
		{"false as F", "F", true, false},
		{"invalid number", "42", false, false},
		{"invalid letters", "abc", false, false},
		{"invalid yes", "yes", false, false},
		{"invalid no", "no", false, false},
		{"empty string", "", false, false},
		{"only spaces", "   ", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testParser(t, schema, tt.input, tt.valid, tt.expected, fmt.Sprintf("expected valid boolean, got: '%v'", tt.input))
		})
	}
}

func TestListParser(t *testing.T) {
	tests := []struct {
		name      string
		delimiter string
		input     string
		expected  []string
	}{
		{"single item", ";", "item", []string{"item"}},
		{"multiple items semicolon", ";", "one;two;three", []string{"one", "two", "three"}},
		{"multiple items comma", ",", "apple,banana,orange", []string{"apple", "banana", "orange"}},
		{"with spaces", ";", "foo ; bar ; baz", []string{"foo ", " bar ", " baz"}},
		{"empty string", ";", "", []string{}},
		{"empty items", ";", ";;", []string{"", "", ""}},
		{"single delimiter", ";", ";", []string{"", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := schema.List(tt.delimiter)
			actual, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func testParser[T comparable](t *testing.T, schema schema.Required[T], in string, valid bool, expected T, errorContains string) {
	t.Helper()
	actual, err := schema.Parse(in)
	if valid {
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	} else {
		assert.ErrorContains(t, err, errorContains)
	}
}
