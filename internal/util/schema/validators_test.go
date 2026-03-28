package schema_test

import (
	"errors"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		value      int
		validators []schema.Validator[int]
		expValue   int
		expErr     bool
	}{
		{
			name:     "no validators",
			value:    42,
			expValue: 42,
			expErr:   false,
		},
		{
			name:       "all validators succeed",
			value:      10,
			validators: []schema.Validator[int]{schema.Positive(), schema.NonNegative()},
			expValue:   10,
			expErr:     false,
		},
		{
			name:       "fails positive validator",
			value:      0,
			validators: []schema.Validator[int]{schema.Positive()},
			expValue:   0,
			expErr:     true,
		},
		{
			name:  "fails first validator only",
			value: 5,
			validators: []schema.Validator[int]{
				func(int) error { return errors.New("fail") },
				func(int) error { return nil },
			},
			expValue: 5,
			expErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.value, tt.validators...)

			actValue, actErr := result.Get()
			assert.Equal(t, tt.expValue, actValue)

			if tt.expErr {
				assert.Error(t, actErr)
			} else {
				assert.NoError(t, actErr)
			}
		})
	}
}

func TestValidationResult_Optional(t *testing.T) {
	tests := []struct {
		name       string
		value      int
		validators []schema.Validator[int]
		fallback   int
		expected   int
	}{
		{
			name:     "no validation returns original",
			value:    7,
			fallback: 999,
			expected: 7,
		},
		{
			name:       "valid value returns original",
			value:      5,
			validators: []schema.Validator[int]{schema.Positive()},
			fallback:   100,
			expected:   5,
		},
		{
			name:       "invalid value returns fallback",
			value:      -1,
			validators: []schema.Validator[int]{schema.Positive()},
			fallback:   100,
			expected:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.value, tt.validators...)
			got := result.Optional(tt.fallback)
			require.Equal(t, tt.expected, got)
		})
	}
}

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

func TestValidURLValidator(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"https URL", "https://bitbucket.org", true},
		{"http URL", "http://localhost:8080", true},
		{"file URL", "file:///tmp/repos", true},
		{"URL with path", "https://bitbucket.org/workspace/repo", true},
		{"no scheme", "bitbucket.org/workspace", false},
		{"empty string", "", false},
		{"only path", "/workspace/repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := schema.ValidURL()(tt.input)
			if tt.valid {
				require.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, "expected valid URL")
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
