package util_test

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrlBuilder_Build_Success(t *testing.T) {
	tests := []struct {
		name     string
		builder  util.UrlBuilder
		expected string
	}{
		{
			"base URL only without path and query params",
			util.UrlBuilder{BaseUrl: "http://foo.com"},
			"http://foo.com/",
		},
		{
			"base URL with path segments",
			util.UrlBuilder{
				BaseUrl: "http://foo.com",
				Path:    []string{"", "/query/", "/search"},
			},
			"http://foo.com/query/search",
		},
		{
			"base URL with query params",
			util.UrlBuilder{
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
	builder := util.UrlBuilder{BaseUrl: "://invalid"}
	_, err := builder.Build()
	require.Error(t, err)
}

func TestCreateRequest_Success(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		urlBuilder util.UrlBuilder
		body       []byte
	}{
		{
			"without body",
			"GET",
			util.UrlBuilder{BaseUrl: "https://www.example.com", Path: []string{"v1", "api"}},
			nil,
		},
		{
			"with body",
			"GET",
			util.UrlBuilder{BaseUrl: "https://www.goolge.com"},
			[]byte("Hello, World!"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := util.CreateRequest(tt.method, tt.urlBuilder, tt.body)
			require.NoError(t, err)
			assert.Equal(t, tt.method, req.Method)

			expectedUrl, err := tt.urlBuilder.Build()
			require.NoError(t, err)
			assert.Equal(t, expectedUrl, req.URL.String())

			if tt.body != nil {
				defer req.Body.Close()
				bodyBytes, err := io.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.body, bodyBytes)
			} else {
				assert.Nil(t, req.Body)
			}
		})
	}
}

func TestCreateRequest_Failure(t *testing.T) {
	_, err := util.CreateRequest("GET", util.UrlBuilder{BaseUrl: "://invalid"}, nil)
	require.Error(t, err)
}

func TestDoRequest_Success(t *testing.T) {
	var gotReq *http.Request
	var gotResp *http.Response

	req := &http.Request{URL: &url.URL{Scheme: "https", Host: "example.com"}}
	resp := &http.Response{}

	client := NewMockClient(func(r *http.Request) (*http.Response, error) {
		gotReq = r
		return resp, nil
	})

	gotResp, err := util.DoRequest(client, req)
	require.NoError(t, err)
	assert.Equal(t, req.URL.String(), gotReq.URL.String(), "expected request URL to match")
	assert.Same(t, resp, gotResp, "expected response to be the same")
}

func TestDoRequest_Failure(t *testing.T) {
	causeErr := errors.New("test error")
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "https", Host: "example.com"},
	}

	client := NewMockClient(func(r *http.Request) (*http.Response, error) {
		return nil, causeErr
	})

	_, err := util.DoRequest(client, req)
	require.ErrorIs(t, err, causeErr)
}

func ReadResponse_Success(t *testing.T) {
	expected := "test"
	res := &http.Response{Body: io.NopCloser(strings.NewReader(expected))}
	actual, err := util.ReadResponseBytes(res)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func ReadResponse_Failure(t *testing.T) {
	res := &http.Response{
		Body: io.NopCloser(errReader{}),
		Request: &http.Request{
			Method: "GET",
			URL:    &url.URL{Scheme: "https", Host: "example.com"},
		},
	}
	_, err := util.ReadResponseBytes(res)
	require.Error(t, err)
	prefix := `failed to read response to "GET" "https://example.com": `
	hasPrefix := strings.HasPrefix(err.Error(), prefix)
	assert.Truef(t, hasPrefix, "expected error to start with %q, got %q", prefix, err.Error())
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewMockClient returns an *http.Client that uses the provided function
// to handle HTTP requests. Useful for testing code that depends on http.Client.
func NewMockClient(process func(req *http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(process),
	}
}
