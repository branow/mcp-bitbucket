package util

import "fmt"

type ErrorBuilder struct {
	Action string
}

func (b *ErrorBuilder) Wrap(err error) error {
	return fmt.Errorf("failed to %s: %w", b.Action, err)
}

func WrapErrorFunc[T any](action string, f func() (T, error)) (T, error) {
	result, err := f()
	if err != nil {
		err = WrapError(action, err)
	}
	return result, err
}

func WrapError(action string, err error) error {
	return fmt.Errorf("failed to %s: %w", action, err)
}
