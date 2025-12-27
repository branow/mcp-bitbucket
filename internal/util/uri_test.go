package util_test

import (
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUriParams_Success(t *testing.T) {
	tests := []struct {
		name     string
		template string
		uri      string
		expected *util.UriParams
	}{
		// Single path parameter tests
		{
			name:     "single path parameter at end",
			template: "https://api.example.com/users/{id}",
			uri:      "https://api.example.com/users/12345",
			expected: &util.UriParams{
				Path:  map[string]string{"id": "12345"},
				Query: map[string]string{},
			},
		},
		{
			name:     "single path parameter in middle",
			template: "https://api.example.com/users/{id}/profile",
			uri:      "https://api.example.com/users/abc-def/profile",
			expected: &util.UriParams{
				Path:  map[string]string{"id": "abc-def"},
				Query: map[string]string{},
			},
		},
		{
			name:     "single path parameter at start",
			template: "file:///{drive}/documents/file.txt",
			uri:      "file:///C/documents/file.txt",
			expected: &util.UriParams{
				Path:  map[string]string{"drive": "C"},
				Query: map[string]string{},
			},
		},

		// Multiple path parameters tests
		{
			name:     "two consecutive path parameters",
			template: "http://localhost:8080/api/{version}/{resource}",
			uri:      "http://localhost:8080/api/v2/products",
			expected: &util.UriParams{
				Path:  map[string]string{"version": "v2", "resource": "products"},
				Query: map[string]string{},
			},
		},
		{
			name:     "three path parameters with literals between",
			template: "https://cdn.example.com/{region}/static/{category}/assets/{filename}",
			uri:      "https://cdn.example.com/us-west/static/images/assets/logo.png",
			expected: &util.UriParams{
				Path:  map[string]string{"region": "us-west", "category": "images", "filename": "logo.png"},
				Query: map[string]string{},
			},
		},
		{
			name:     "four path parameters",
			template: "custom://example.com/{org}/{project}/{repo}/{branch}",
			uri:      "custom://example.com/acme/web/frontend/main",
			expected: &util.UriParams{
				Path:  map[string]string{"org": "acme", "project": "web", "repo": "frontend", "branch": "main"},
				Query: map[string]string{},
			},
		},

		// Query parameter tests
		{
			name:     "single query parameter",
			template: "http://search.example.com/search?q={query}",
			uri:      "http://search.example.com/search?q=golang+tutorial",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"query": "golang tutorial"}, // + is decoded to space
			},
		},
		{
			name:     "three query parameters",
			template: "https://api.weather.com/forecast?city={city}&units={units}&lang={lang}",
			uri:      "https://api.weather.com/forecast?city=London&units=metric&lang=en",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"city": "London", "units": "metric", "lang": "en"},
			},
		},
		{
			name:     "query parameter with empty value",
			template: "http://example.com/api?filter={filter}",
			uri:      "http://example.com/api?filter=",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"filter": ""},
			},
		},

		// Mixed path and query parameters
		{
			name:     "path and query parameters combined",
			template: "https://api.github.com/repos/{owner}/{repo}/issues?state={state}&page={page}",
			uri:      "https://api.github.com/repos/golang/go/issues?state=open&page=3",
			expected: &util.UriParams{
				Path:  map[string]string{"owner": "golang", "repo": "go"},
				Query: map[string]string{"state": "open", "page": "3"},
			},
		},
		{
			name:     "complex path and query mix",
			template: "ftp://files.example.com:21/{directory}/{file}?mode={mode}&type={type}",
			uri:      "ftp://files.example.com:21/uploads/document.pdf?mode=binary&type=ascii",
			expected: &util.UriParams{
				Path:  map[string]string{"directory": "uploads", "file": "document.pdf"},
				Query: map[string]string{"mode": "binary", "type": "ascii"},
			},
		},

		// No parameters tests
		{
			name:     "no parameters with path",
			template: "https://example.com/static/page.html",
			uri:      "https://example.com/static/page.html",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{},
			},
		},
		{
			name:     "no parameters with query string",
			template: "http://example.com/page?foo=bar&baz=qux",
			uri:      "http://example.com/page?foo=bar&baz=qux",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{},
			},
		},
		{
			name:     "root path only",
			template: "https://example.com",
			uri:      "https://example.com",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{},
			},
		},
		{
			name:     "root with trailing slash",
			template: "https://example.com/",
			uri:      "https://example.com/",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{},
			},
		},

		// Optional/missing query parameters
		{
			name:     "some query parameters provided",
			template: "http://api.example.com/data?limit={limit}&offset={offset}&sort={sort}",
			uri:      "http://api.example.com/data?limit=10&sort=asc",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"limit": "10", "offset": "", "sort": "asc"}, // missing params have empty values
			},
		},
		{
			name:     "no query parameters provided",
			template: "http://api.example.com/data?filter={filter}",
			uri:      "http://api.example.com/data",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"filter": ""}, // missing param has empty value
			},
		},

		// Extra parameters in URI (not in template)
		{
			name:     "extra query parameters ignored",
			template: "https://api.example.com/search?q={query}",
			uri:      "https://api.example.com/search?q=test&extra1=val1&extra2=val2",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"query": "test"},
			},
		},

		// URL encoding tests
		{
			name:     "URL-encoded space in path parameter",
			template: "https://files.example.com/{folder}/{filename}",
			uri:      "https://files.example.com/my%20documents/report%202024.pdf",
			expected: &util.UriParams{
				Path:  map[string]string{"folder": "my documents", "filename": "report 2024.pdf"}, // %20 decoded to space
				Query: map[string]string{},
			},
		},
		{
			name:     "URL-encoded special characters in query",
			template: "http://search.example.com/?q={query}",
			uri:      "http://search.example.com/?q=a%2Bb%3Dc%26d%3De",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"query": "a+b=c&d=e"},
			},
		},
		{
			name:     "URL-encoded unicode in path",
			template: "https://example.com/{category}/{item}",
			uri:      "https://example.com/%E6%97%A5%E6%9C%AC/%E6%9D%B1%E4%BA%AC",
			expected: &util.UriParams{
				Path:  map[string]string{"category": "日本", "item": "東京"}, // UTF-8 decoded
				Query: map[string]string{},
			},
		},

		// Different schemes
		{
			name:     "http scheme",
			template: "http://api.example.com/{version}/users",
			uri:      "http://api.example.com/v1/users",
			expected: &util.UriParams{
				Path:  map[string]string{"version": "v1"},
				Query: map[string]string{},
			},
		},
		{
			name:     "https scheme",
			template: "https://secure.example.com/{resource}",
			uri:      "https://secure.example.com/account",
			expected: &util.UriParams{
				Path:  map[string]string{"resource": "account"},
				Query: map[string]string{},
			},
		},
		{
			name:     "custom scheme",
			template: "myapp://app/{feature}/{action}",
			uri:      "myapp://app/editor/open",
			expected: &util.UriParams{
				Path:  map[string]string{"feature": "editor", "action": "open"},
				Query: map[string]string{},
			},
		},
		{
			name:     "file scheme",
			template: "file:///{path}/{subpath}/{filename}",
			uri:      "file:///home/user/document.txt",
			expected: &util.UriParams{
				Path:  map[string]string{"path": "home", "subpath": "user", "filename": "document.txt"},
				Query: map[string]string{},
			},
		},

		// Port numbers
		{
			name:     "with port number in host",
			template: "http://localhost:3000/api/{endpoint}",
			uri:      "http://localhost:3000/api/health",
			expected: &util.UriParams{
				Path:  map[string]string{"endpoint": "health"},
				Query: map[string]string{},
			},
		},
		{
			name:     "different port numbers",
			template: "http://server.local:8443/{service}/{method}",
			uri:      "http://server.local:8443/auth/login",
			expected: &util.UriParams{
				Path:  map[string]string{"service": "auth", "method": "login"},
				Query: map[string]string{},
			},
		},

		// Special values
		{
			name:     "numeric values in path",
			template: "https://api.example.com/orders/{orderId}/items/{itemId}",
			uri:      "https://api.example.com/orders/98765/items/12345",
			expected: &util.UriParams{
				Path:  map[string]string{"orderId": "98765", "itemId": "12345"},
				Query: map[string]string{},
			},
		},
		{
			name:     "uuid in path",
			template: "https://api.example.com/entities/{uuid}",
			uri:      "https://api.example.com/entities/550e8400-e29b-41d4-a716-446655440000",
			expected: &util.UriParams{
				Path:  map[string]string{"uuid": "550e8400-e29b-41d4-a716-446655440000"},
				Query: map[string]string{},
			},
		},
		{
			name:     "boolean-like values in query",
			template: "http://api.example.com/data?active={active}&verbose={verbose}",
			uri:      "http://api.example.com/data?active=true&verbose=false",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"active": "true", "verbose": "false"},
			},
		},
		{
			name:     "negative numbers in query",
			template: "http://api.example.com/data?offset={offset}&limit={limit}",
			uri:      "http://api.example.com/data?offset=-10&limit=100",
			expected: &util.UriParams{
				Path:  map[string]string{},
				Query: map[string]string{"offset": "-10", "limit": "100"},
			},
		},

		// Edge cases with slashes
		{
			name:     "deep nested path",
			template: "https://storage.example.com/{bucket}/{folder1}/{folder2}/{folder3}/{file}",
			uri:      "https://storage.example.com/mybucket/2024/01/15/data.json",
			expected: &util.UriParams{
				Path:  map[string]string{"bucket": "mybucket", "folder1": "2024", "folder2": "01", "folder3": "15", "file": "data.json"},
				Query: map[string]string{},
			},
		},
		{
			name:     "single segment path",
			template: "https://example.com/{page}",
			uri:      "https://example.com/about",
			expected: &util.UriParams{
				Path:  map[string]string{"page": "about"},
				Query: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := util.ParseUriParams(tt.template, tt.uri)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseUriParams_Failure(t *testing.T) {
	tests := []struct {
		name     string
		template string
		uri      string
		errorMsg string
	}{
		// Invalid template/URI format
		{
			name:     "invalid template - missing scheme",
			template: "://invalid",
			uri:      "http://example.com/path",
			errorMsg: "invalid template",
		},
		{
			name:     "invalid template - malformed",
			template: "ht!tp://example.com",
			uri:      "http://example.com",
			errorMsg: "invalid template",
		},
		{
			name:     "invalid URI - missing scheme",
			template: "http://example.com/path",
			uri:      "://invalid",
			errorMsg: "invalid URI",
		},
		{
			name:     "invalid URI - malformed",
			template: "https://api.example.com",
			uri:      "ht!tp://invalid",
			errorMsg: "invalid URI",
		},

		// Scheme mismatches
		{
			name:     "scheme mismatch - http vs https",
			template: "http://example.com/api",
			uri:      "https://example.com/api",
			errorMsg: "scheme mismatch",
		},
		{
			name:     "scheme mismatch - custom vs http",
			template: "custom://app/resource",
			uri:      "http://app/resource",
			errorMsg: "scheme mismatch",
		},
		{
			name:     "scheme mismatch - file vs http",
			template: "file:///path/to/file",
			uri:      "http:///path/to/file",
			errorMsg: "scheme mismatch",
		},
		{
			name:     "scheme mismatch - ftp vs http",
			template: "ftp://example.com",
			uri:      "http://example.com",
			errorMsg: "scheme mismatch",
		},

		// Host mismatches
		{
			name:     "host mismatch - different domains",
			template: "https://example.com/api",
			uri:      "https://different.com/api",
			errorMsg: "host mismatch",
		},
		{
			name:     "host mismatch - subdomain difference",
			template: "https://api.example.com/data",
			uri:      "https://app.example.com/data",
			errorMsg: "host mismatch",
		},
		{
			name:     "host mismatch - with port vs without",
			template: "http://localhost/api",
			uri:      "http://localhost:8080/api",
			errorMsg: "host mismatch",
		},
		{
			name:     "host mismatch - different ports",
			template: "http://example.com:8080/api",
			uri:      "http://example.com:3000/api",
			errorMsg: "host mismatch",
		},
		{
			name:     "host mismatch - localhost vs 127.0.0.1",
			template: "http://localhost/api",
			uri:      "http://127.0.0.1/api",
			errorMsg: "host mismatch",
		},

		// Path segment count mismatches
		{
			name:     "path too short - missing segments",
			template: "https://api.example.com/{version}/users/{id}",
			uri:      "https://api.example.com/v1/users",
			errorMsg: "path segment count mismatch",
		},
		{
			name:     "path too long - extra segments",
			template: "https://api.example.com/{version}/users",
			uri:      "https://api.example.com/v1/users/extra",
			errorMsg: "path segment count mismatch",
		},
		{
			name:     "path count mismatch - root vs path",
			template: "https://example.com/",
			uri:      "https://example.com/path",
			errorMsg: "path segment mismatch",
		},
		{
			name:     "path count mismatch - one vs two segments",
			template: "https://api.example.com/{resource}",
			uri:      "https://api.example.com/users/profile",
			errorMsg: "path segment count mismatch",
		},

		// Path segment literal mismatches
		{
			name:     "path literal mismatch - first segment",
			template: "https://api.example.com/users/{id}",
			uri:      "https://api.example.com/posts/123",
			errorMsg: "path segment mismatch",
		},
		{
			name:     "path literal mismatch - middle segment",
			template: "https://api.example.com/{version}/users/profile",
			uri:      "https://api.example.com/v1/posts/profile",
			errorMsg: "path segment mismatch",
		},
		{
			name:     "path literal mismatch - last segment",
			template: "https://api.example.com/api/{resource}/list",
			uri:      "https://api.example.com/api/users/details",
			errorMsg: "path segment mismatch",
		},
		{
			name:     "path literal mismatch - case sensitive",
			template: "https://api.example.com/Users/{id}",
			uri:      "https://api.example.com/users/123",
			errorMsg: "path segment mismatch",
		},
		{
			name:     "path literal mismatch - plural vs singular",
			template: "https://api.example.com/user/{id}",
			uri:      "https://api.example.com/users/123",
			errorMsg: "path segment mismatch",
		},

		// Complex mismatch scenarios
		{
			name:     "multiple mismatches - scheme and path",
			template: "http://api.example.com/{version}/users",
			uri:      "https://api.example.com/v1/posts",
			errorMsg: "scheme mismatch",
		},
		{
			name:     "multiple mismatches - host and path count",
			template: "https://api1.example.com/{resource}",
			uri:      "https://api2.example.com/users/profile",
			errorMsg: "host mismatch",
		},

		// Edge cases
		{
			name:     "empty template",
			template: "",
			uri:      "https://example.com",
			errorMsg: "scheme mismatch",
		},
		{
			name:     "empty URI",
			template: "https://example.com",
			uri:      "",
			errorMsg: "scheme mismatch",
		},
		{
			name:     "whitespace in template host",
			template: "https://example.com /api",
			uri:      "https://example.com/api",
			errorMsg: "invalid template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := util.ParseUriParams(tt.template, tt.uri)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}
