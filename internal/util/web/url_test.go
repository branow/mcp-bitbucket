package web_test

import (
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrlBuilder_Build_Success(t *testing.T) {
	tests := []struct {
		name     string
		builder  web.UrlBuilder
		expected string
	}{
		{
			"base URL only without path and query params",
			web.UrlBuilder{BaseUrl: "http://foo.com"},
			"http://foo.com/",
		},
		{
			"base URL with path segments",
			web.UrlBuilder{
				BaseUrl: "http://foo.com",
				Path:    []string{"", "/query/", "/search"},
			},
			"http://foo.com/query/search",
		},
		{
			"base URL with query params",
			web.UrlBuilder{
				BaseUrl:     "http://foo.com",
				QueryParams: map[string]string{"p1": "v1", "p2": "", "": "v3"},
			},
			"http://foo.com/?=v3&p1=v1&p2=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := tt.builder.Build()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestUrlBuilder_Build_Failure(t *testing.T) {
	builder := web.UrlBuilder{BaseUrl: "://invalid"}
	_, err := builder.Build()
	require.Error(t, err)
}
