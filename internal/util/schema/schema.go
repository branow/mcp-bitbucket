package schema

// Optional represents a schema that always succeeds by returning
// a fallback value when parsing or validation fails.
//
// Use Optional when you want to provide default values for invalid input.
//
// Example:
//
//	schema := Int().Must(Positive()).Optional(10)
//	value := schema.Parse("invalid") // returns 10
//	value = schema.Parse("42")        // returns 42
//
// IMPORTANT: Schemas are mutable. Calling Must() modifies the underlying schema.
// Do not reuse base schemas if you want independent validation chains:
//
//	base := Int()
//	schema1 := base.Must(Positive())    // modifies base!
//	schema2 := base.Must(NonNegative()) // also modifies base!
type Optional[T any] interface {
	// Parse parses and validates the input string.
	// Returns the parsed value on success, or the fallback value on failure.
	Parse(input string) T

	// Must adds validators that will be checked after parsing.
	// Returns the same schema for method chaining.
	// WARNING: This mutates the underlying schema.
	Must(...Validator[T]) Optional[T]

	// Required converts this to a Required that returns errors.
	Required() Required[T]

	// Critical converts this to a Critical that panics on errors.
	Critical() Critical[T]
}

// Required represents a schema that returns an error when parsing
// or validation fails.
//
// Use Required for explicit error handling where you want to handle
// validation failures yourself.
//
// Example:
//
//	schema := Int().Must(Positive())
//	value, err := schema.Parse("42")
//	if err != nil {
//	  // handle error
//	}
//
// IMPORTANT: Schemas are mutable. Calling Must() modifies the underlying schema.
// Do not reuse base schemas if you want independent validation chains:
//
//	base := Int()
//	schema1 := base.Must(Positive())    // modifies base!
//	schema2 := base.Must(NonNegative()) // also modifies base!
type Required[T any] interface {
	// Parse parses and validates the input string.
	// Returns the parsed value and an error if parsing or validation fails.
	Parse(input string) (T, error)

	// Must adds validators that will be checked after parsing.
	// Returns the same schema for method chaining.
	// WARNING: This mutates the underlying schema.
	Must(...Validator[T]) Required[T]

	// Optional converts this to an Optional with a fallback value.
	// The fallback value is returned when parsing or validation fails.
	Optional(fallback T) Optional[T]

	// Critical converts this to a Critical that panics on errors.
	Critical() Critical[T]
}

// Critical represents a schema that panics when parsing or validation fails.
//
// Use Critical for configuration values that must be valid for the
// application to function. This is useful during startup when invalid config
// should prevent the application from running.
//
// Example:
//
//	schema := Int().Must(Positive()).Critical()
//	port := schema.Parse(os.Getenv("PORT")) // panics if invalid
//
// IMPORTANT: Schemas are mutable. Calling Must() modifies the underlying schema.
// Do not reuse base schemas if you want independent validation chains:
//
//	base := Int()
//	schema1 := base.Must(Positive())    // modifies base!
//	schema2 := base.Must(NonNegative()) // also modifies base!
type Critical[T any] interface {
	// Parse parses and validates the input string.
	// Returns the parsed value on success, or panics on failure.
	Parse(input string) T

	// Must adds validators that will be checked after parsing.
	// Returns the same schema for method chaining.
	// WARNING: This mutates the underlying schema.
	Must(...Validator[T]) Critical[T]

	// Optional converts this to an Optional with a fallback value.
	Optional(fallback T) Optional[T]

	// Required converts this to a Required that returns errors.
	Required() Required[T]
}

// NewSchema creates a new Required with a custom parser function.
func NewSchema[T any](parser func(string) (T, error)) Required[T] {
	return reqView[T]{s: &schema[T]{parser: parser}}
}

type schema[T any] struct {
	parser     func(string) (T, error)
	validators []Validator[T]
}

func (s *schema[T]) parse(input string) (T, error) {
	value, err := s.parser(input)
	if err != nil {
		return value, err
	}
	for _, validator := range s.validators {
		if err := validator(value); err != nil {
			return value, err
		}
	}
	return value, nil
}

type reqView[T any] struct{ s *schema[T] }

func (v reqView[T]) Parse(in string) (T, error) { return v.s.parse(in) }
func (v reqView[T]) Must(validators ...Validator[T]) Required[T] {
	v.s.validators = append(v.s.validators, validators...)
	return v
}
func (v reqView[T]) Optional(fb T) Optional[T] { return optView[T]{v.s, fb} }
func (v reqView[T]) Critical() Critical[T]     { return critView[T]{v.s} }

type optView[T any] struct {
	s        *schema[T]
	fallback T
}

func (v optView[T]) Parse(in string) T {
	val, err := v.s.parse(in)
	if err != nil {
		return v.fallback
	}
	return val
}
func (v optView[T]) Must(validators ...Validator[T]) Optional[T] {
	v.s.validators = append(v.s.validators, validators...)
	return v
}
func (v optView[T]) Required() Required[T] { return reqView[T]{v.s} }
func (v optView[T]) Critical() Critical[T] { return critView[T]{v.s} }

type critView[T any] struct{ s *schema[T] }

func (v critView[T]) Parse(in string) T {
	val, err := v.s.parse(in)
	if err != nil {
		panic(err)
	}
	return val
}
func (v critView[T]) Must(validators ...Validator[T]) Critical[T] {
	v.s.validators = append(v.s.validators, validators...)
	return v
}
func (v critView[T]) Required() Required[T]     { return reqView[T]{v.s} }
func (v critView[T]) Optional(fb T) Optional[T] { return optView[T]{v.s, fb} }
