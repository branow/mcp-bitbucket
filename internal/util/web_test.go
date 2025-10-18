package util_test

import (
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{"empty username and password", "", "", "Basic Og=="},
		{"empty password", "john@gmail.com", "", "Basic am9obkBnbWFpbC5jb206"},
		{"filled username password", "john@gmail.com", "test123$", "Basic am9obkBnbWFpbC5jb206dGVzdDEyMyQ="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := util.BasicAuth(tt.username, tt.password)
			assert.Equal(t, tt.expected, actual, "BasicAuth(%q, %q)", tt.username, tt.password)
		})
	}
}
