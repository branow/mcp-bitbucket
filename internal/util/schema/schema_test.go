package schema_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMust_SingleValidator(t *testing.T) {
	schema := schema.Int().Must(schema.Positive())

	t.Run("valid input", func(t *testing.T) {
		result, err := schema.Parse("42")
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("invalid by parser", func(t *testing.T) {
		_, err := schema.Parse("abc")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected valid integer")
	})

	t.Run("invalid by validator", func(t *testing.T) {
		_, err := schema.Parse("-5")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected positive integer")
	})
}

func TestMust_MultipleValidators(t *testing.T) {
	evenValidator := func(i int) error {
		if i%2 != 0 {
			return fmt.Errorf("expected even number, got: %d", i)
		}
		return nil
	}

	schema := schema.Int().Must(schema.Positive(), evenValidator)

	tests := []struct {
		name      string
		input     string
		shouldErr bool
		errorMsg  string
	}{
		{"valid even positive", "42", false, ""},
		{"invalid zero", "0", true, "expected positive integer"},
		{"invalid negative even", "-4", true, "expected positive integer"},
		{"invalid positive odd", "5", true, "expected even number"},
		{"invalid parser", "abc", true, "expected valid integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := schema.Parse(tt.input)
			if tt.shouldErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMust_ValidatorOrder(t *testing.T) {
	firstValidator := func(i int) error {
		if i < 10 {
			return fmt.Errorf("first: must be >= 10")
		}
		return nil
	}

	secondValidator := func(i int) error {
		if i > 100 {
			return fmt.Errorf("second: must be <= 100")
		}
		return nil
	}

	schema := schema.Int().Must(firstValidator, secondValidator)

	tests := []struct {
		name     string
		input    string
		errorMsg string
	}{
		{"below range fails first", "5", "first: must be >= 10"},
		{"above range fails second", "150", "second: must be <= 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := schema.Parse(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestSchema_Conversions(t *testing.T) {
	t.Run("Required to Optional", func(t *testing.T) {
		required := schema.Int()
		optional := required.Optional(99)

		assert.Equal(t, 42, optional.Parse("42"))
		assert.Equal(t, 99, optional.Parse("invalid"))
	})

	t.Run("Required to Critical", func(t *testing.T) {
		required := schema.Int()
		critical := required.Critical()

		assert.Equal(t, 42, critical.Parse("42"))
		assert.Panics(t, func() {
			critical.Parse("invalid")
		})
	})

	t.Run("Optional to Required", func(t *testing.T) {
		optional := schema.Int().Optional(99)
		required := optional.Required()

		result, err := required.Parse("42")
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		_, err = required.Parse("invalid")
		require.Error(t, err)
	})

	t.Run("Optional to Critical", func(t *testing.T) {
		optional := schema.Int().Optional(99)
		critical := optional.Critical()

		assert.Equal(t, 42, critical.Parse("42"))
		assert.Panics(t, func() {
			critical.Parse("invalid")
		})
	})

	t.Run("Critical to Optional", func(t *testing.T) {
		critical := schema.Int().Critical()
		optional := critical.Optional(99)

		assert.Equal(t, 42, optional.Parse("42"))
		assert.Equal(t, 99, optional.Parse("invalid"))
	})

	t.Run("Critical to Required", func(t *testing.T) {
		critical := schema.Int().Critical()
		required := critical.Required()

		result, err := required.Parse("42")
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		_, err = required.Parse("invalid")
		require.Error(t, err)
	})
}

func TestComplexExpression_MultipleValidatorsWithConversions(t *testing.T) {
	evenValidator := func(i int) error {
		if i%2 != 0 {
			return fmt.Errorf("must be even")
		}
		return nil
	}

	lessThan100 := func(i int) error {
		if i >= 100 {
			return fmt.Errorf("must be < 100")
		}
		return nil
	}

	schema := schema.Int().
		Must(schema.Positive(), evenValidator, lessThan100).
		Optional(50)

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"valid", "42", 42},
		{"parser fails", "abc", 50},
		{"not positive", "0", 50},
		{"not even", "5", 50},
		{"too large", "100", 50},
		{"negative even", "-4", 50},
		{"valid edge case", "2", 2},
		{"valid large", "98", 98},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Parse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComplexExpression_CustomParserWithValidators(t *testing.T) {
	sumParser := func(s string) (int, error) {
		parts := strings.Split(s, ",")
		sum := 0
		for _, part := range parts {
			var num int
			_, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &num)
			if err != nil {
				return 0, fmt.Errorf("invalid format: %s", part)
			}
			sum += num
		}
		return sum, nil
	}

	schema := schema.NewSchema(sumParser).
		Must(schema.Positive()).
		Optional(0)

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"single number", "42", 42},
		{"multiple numbers", "10,20,30", 60},
		{"with spaces", "5, 10, 15", 30},
		{"invalid format", "1,abc,3", 0},
		{"empty string", "", 0},
		{"zero sum", "10,-10", 0},
		{"negative sum", "-5,-10", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Parse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComplexExpression_NestedConversions(t *testing.T) {
	base := schema.Int().Must(schema.Positive())
	critical1 := base.Critical()
	optional := critical1.Optional(100)
	required := optional.Required()
	critical2 := required.Critical()

	t.Run("valid input", func(t *testing.T) {
		result := critical2.Parse("42")
		assert.Equal(t, 42, result)
	})

	t.Run("invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			critical2.Parse("abc")
		})
	})

	t.Run("validator fails panics", func(t *testing.T) {
		assert.Panics(t, func() {
			critical2.Parse("0")
		})
	})
}

func TestOnFallback(t *testing.T) {
	t.Run("fallback listener called on parse error", func(t *testing.T) {
		var called bool
		var capturedFallback int
		var capturedError error

		listener := func(fallback int, err error) {
			called = true
			capturedFallback = fallback
			capturedError = err
		}

		s := schema.Int().Optional(99).OnFallback(listener)
		result := s.Parse("invalid")

		assert.Equal(t, 99, result)
		assert.True(t, called, "listener should be called")
		assert.Equal(t, 99, capturedFallback)
		assert.Error(t, capturedError)
		assert.Contains(t, capturedError.Error(), "expected valid integer")
	})

	t.Run("fallback listener called on validation error", func(t *testing.T) {
		var called bool
		var capturedFallback int
		var capturedError error

		listener := func(fallback int, err error) {
			called = true
			capturedFallback = fallback
			capturedError = err
		}

		s := schema.Int().Must(schema.Positive()).Optional(99).OnFallback(listener)
		result := s.Parse("-5")

		assert.Equal(t, 99, result)
		assert.True(t, called, "listener should be called")
		assert.Equal(t, 99, capturedFallback)
		assert.Error(t, capturedError)
		assert.Contains(t, capturedError.Error(), "expected positive integer")
	})

	t.Run("fallback listener not called on success", func(t *testing.T) {
		var called bool

		listener := func(fallback int, err error) {
			called = true
		}

		s := schema.Int().Optional(99).OnFallback(listener)
		result := s.Parse("42")

		assert.Equal(t, 42, result)
		assert.False(t, called, "listener should not be called on success")
	})

	t.Run("multiple fallback listeners", func(t *testing.T) {
		var called1, called2 bool

		listener1 := func(fallback int, err error) {
			called1 = true
		}

		listener2 := func(fallback int, err error) {
			called2 = true
		}

		s := schema.Int().Optional(99).OnFallback(listener1).OnFallback(listener2)
		s.Parse("invalid")

		assert.True(t, called1, "first listener should be called")
		assert.True(t, called2, "second listener should be called")
	})
}
