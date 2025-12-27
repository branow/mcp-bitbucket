package web_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestRequest struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type TestResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func TestRequestBuilder_Build_Success(t *testing.T) {
	t.Run("GET request without body", func(t *testing.T) {
		builder := web.RequestBuilder[TestRequest]{
			Method: "GET",
			Url: web.UrlBuilder{
				BaseUrl: "http://example.com",
				Path:    []string{"api", "users"},
			},
			Mime: web.MimeOmit,
		}

		req, err := builder.Build()
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/api/users", req.URL.String())
	})

	t.Run("POST request with JSON body", func(t *testing.T) {
		builder := web.RequestBuilder[TestRequest]{
			Method: "POST",
			Url: web.UrlBuilder{
				BaseUrl: "http://example.com",
				Path:    []string{"api", "users"},
			},
			Mime: web.MimeApplicationJson,
			Body: &TestRequest{Name: "test", Value: 42},
		}

		req, err := builder.Build()
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/api/users", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

		bodyBytes, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"test","value":42}`, string(bodyBytes))
	})

	t.Run("PUT request with text body", func(t *testing.T) {
		builder := web.RequestBuilder[string]{
			Method: "PUT",
			Url: web.UrlBuilder{
				BaseUrl: "http://example.com",
				Path:    []string{"api", "data"},
			},
			Mime: web.MimeTextPlain,
			Body: stringPtr("hello world"),
		}

		req, err := builder.Build()
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/api/data", req.URL.String())
		assert.Equal(t, "text/plain", req.Header.Get("Content-Type"))

		bodyBytes, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(bodyBytes))
	})

	t.Run("PATCH request with invalid body", func(t *testing.T) {
		builder := web.RequestBuilder[TestRequest]{
			Method: "PATCH",
			Url: web.UrlBuilder{
				BaseUrl: "http://example.com",
				Path:    []string{"api", "users"},
			},
			Mime: web.MimeMultipartFormData,
			Body: &TestRequest{Name: "test", Value: 42},
		}

		req, err := builder.Build()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write request body")
		assert.Nil(t, req)
	})
}

func TestRequestBuilder_Build_InvalidURL(t *testing.T) {
	builder := web.RequestBuilder[TestRequest]{
		Method: "GET",
		Url: web.UrlBuilder{
			BaseUrl: "://invalid-url",
		},
		Mime: web.MimeOmit,
	}

	_, err := builder.Build()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build URL")
}

func TestWriteRequestBody_MimeOmit(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	body := &TestRequest{Name: "test", Value: 42}

	err := web.WriteRequestBody(req, web.MimeOmit, body)
	require.NoError(t, err)
	assert.Equal(t, http.NoBody, req.Body)
}

func TestWriteRequestBody_UnsupportedMime(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	body := &TestRequest{Name: "test", Value: 42}

	err := web.WriteRequestBody(req, web.Mime("unsupported/type"), body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported request MIME type")
}

func TestWriteRequestBody_MimeTextPlain(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)
	text := "hello world"

	web.WriteRequestBody(req, web.MimeTextPlain, &text)

	assert.Equal(t, "text/plain", req.Header.Get("Content-Type"))
	assert.Equal(t, int64(len(text)), req.ContentLength)

	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(bodyBytes))
}

func TestWriteRequestBody_MimeApplicationJson(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)
	body := &TestRequest{Name: "test", Value: 42}
	expected := `{"name":"test","value":42}`

	err := web.WriteRequestBody(req, web.MimeApplicationJson, body)
	require.NoError(t, err)

	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, int64(len(expected)), req.ContentLength)

	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.JSONEq(t, expected, string(bodyBytes))
}

func TestWriteRequestBody_MimeFormMultipart(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)

	form := &web.MultipartForm{
		Parts: []web.FormPart{
			&web.TextField{Name: "field1", Value: "value1"},
			&web.TextField{Name: "field2", Value: "value2"},
			&web.FileField{
				Name:     "file1",
				Filename: "test.txt",
				Reader:   strings.NewReader("file content"),
			},
		},
	}

	web.WriteRequestBody(req, web.MimeMultipartFormData, form)

	assert.Contains(t, req.Header.Get("Content-Type"), "multipart/form-data")
	assert.Contains(t, req.Header.Get("Content-Type"), "boundary=")
	assert.Equal(t, int64(-1), req.ContentLength)
	assert.NotNil(t, req.Body)

	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	bodyStr := string(bodyBytes)
	assert.Contains(t, bodyStr, "field1")
	assert.Contains(t, bodyStr, "value1")
	assert.Contains(t, bodyStr, "field2")
	assert.Contains(t, bodyStr, "value2")
	assert.Contains(t, bodyStr, "file1")
	assert.Contains(t, bodyStr, "test.txt")
	assert.Contains(t, bodyStr, "file content")
}

func TestWriteRequestBody_MultipartFormWrongType(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)
	body := &TestRequest{Name: "test", Value: 42}

	err := web.WriteRequestBody(req, web.MimeMultipartFormData, body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multipart request body must be *MultipartForm")
}

func TestWriteRequestJson_NilBody(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)

	err := web.WriteRequestJson[TestRequest](req, nil)
	require.NoError(t, err)
	assert.Equal(t, http.NoBody, req.Body)
}

func TestWriteRequestJson_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)

	// Create a type that cannot be marshaled to JSON
	body := &struct {
		Chan chan int `json:"chan"`
	}{
		Chan: make(chan int),
	}

	err := web.WriteRequestJson(req, body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encode json request body")
}

func TestWriteRequestText_NilBody(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)

	web.WriteRequestText[string](req, nil)

	assert.Empty(t, req.Header.Get("Content-Type"))
	assert.Equal(t, http.NoBody, req.Body)
}

func TestWriteRequestText_Integer(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)
	num := 42

	web.WriteRequestText(req, &num)

	assert.Equal(t, "text/plain", req.Header.Get("Content-Type"))

	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, "42", string(bodyBytes))
}

func TestWriteRequestFormMultipart_NilForm(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)

	web.WriteRequestFormMultipart(req, nil)

	assert.Empty(t, req.Header.Get("Content-Type"))
	assert.Equal(t, http.NoBody, req.Body)
}

func TestWriteRequestFormMultipart_EmptyForm(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", nil)

	form := &web.MultipartForm{
		Parts: []web.FormPart{},
	}

	web.WriteRequestFormMultipart(req, form)

	assert.Contains(t, req.Header.Get("Content-Type"), "multipart/form-data")
	assert.NotNil(t, req.Body)
}

func TestReadResponseBody_MimeOmit(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
	}

	var result TestResponse
	err := web.ReadResponseBody(resp, web.MimeOmit, &result)
	require.NoError(t, err)
	assert.Empty(t, result.Status)
}

func TestReadResponseBody_UnsupportedMime(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
	}

	var result TestResponse
	err := web.ReadResponseBody(resp, web.Mime("unsupported/type"), &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported response MIME type")
}

func TestReadResponseBody_MimeApplicationJson(t *testing.T) {
	jsonData := `{"status":"ok","message":"success"}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(jsonData)),
	}

	var result TestResponse
	err := web.ReadResponseBody(resp, web.MimeApplicationJson, &result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result.Status)
	assert.Equal(t, "success", result.Message)
}

func TestReadResponseBody_MimeTextPlain(t *testing.T) {
	textData := "hello world"
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(textData)),
	}

	var result string
	err := web.ReadResponseBody(resp, web.MimeTextPlain, &result)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestReadResponseBody_MimeTextPlain_WrongType(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("hello")),
	}

	var result TestResponse
	err := web.ReadResponseBody(resp, web.MimeTextPlain, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text response body must be *string")
}

func TestReadResponseJson_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`invalid json`)),
	}

	var result TestResponse
	err := web.ReadResponseJson(resp, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode json response body")
}

func TestReadResponseJson_UnknownFields(t *testing.T) {
	jsonData := `{"status":"ok","message":"success","unknown":"field"}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(jsonData)),
	}

	var result TestResponse
	err := web.ReadResponseBody(resp, web.MimeApplicationJson, &result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result.Status)
	assert.Equal(t, "success", result.Message)
}

func TestReadResponseText_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	var result string
	err := web.ReadResponseText(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func stringPtr(s string) *string {
	return &s
}
