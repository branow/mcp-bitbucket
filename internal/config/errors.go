package config

import "fmt"

type MissingEnvVarError struct {
	Key string
}

func (e MissingEnvVarError) Error() string {
	return fmt.Sprintf("environment variable %q is missing", e.Key)
}
