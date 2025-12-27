package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/branow/mcp-bitbucket/internal/util"
)

// Mime represents a MIME type (content type) for HTTP requests and responses.
type Mime string

const (
	MimeApplicationJson   Mime = "application/json"
	MimeTextPlain         Mime = "text/plain"
	MimeMultipartFormData Mime = "multipart/form-data"
	MimeOmit              Mime = "" // indicates no content type should be set (no body)
)

// RequestBuilder constructs HTTP requests with a specified body type and MIME type.
//
// Example:
//
//	builder := RequestBuilder[MyRequestBody]{
//	  Method: "POST",
//	  Url: UrlBuilder{
//	    BaseUrl: "https://api.example.com",
//	    Path:    []string{"users"},
//	  },
//	  Mime: MimeApplicationJson,
//	  Body: &MyRequestBody{Name: "John"},
//	}
//	req, err := builder.Build()
type RequestBuilder[T any] struct {
	// Method is the HTTP method (GET, POST, PUT, DELETE, etc.).
	Method string
	// Url is the URL builder for constructing the request URL.
	Url UrlBuilder
	// Mime is the MIME type for the request body.
	Mime Mime
	// Body is a pointer to the request body data.
	Body *T
}

// Build constructs an http.Request from the builder configuration.
//
// Returns an error if the URL cannot be built or the body cannot be written.
func (b RequestBuilder[T]) Build() (*http.Request, error) {
	url, err := b.Url.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequest(b.Method, url, nil)
	if err != nil {
		return nil, err
	}

	if b.Mime != MimeOmit {
		if err := WriteRequestBody(req, b.Mime, b.Body); err != nil {
			return nil, fmt.Errorf("failed to write request body: %w", err)
		}
	}

	return req, nil
}

// WriteRequestBody writes the body to the HTTP request based on the specified MIME type.
//
// Supported MIME types:
//   - MimeApplicationJson: Encodes body as JSON
//   - MimeTextPlain: Converts body to string
//   - MimeMultipartFormData: Writes body as multipart form (body must be *MultipartForm)
//   - MimeOmit: Skips writing body
//
// Returns an error if the MIME type is unsupported or the body cannot be written.
func WriteRequestBody[T any](req *http.Request, mime Mime, body *T) error {
	switch mime {
	case MimeOmit:
		return nil
	case MimeApplicationJson:
		return WriteRequestJson(req, body)
	case MimeTextPlain:
		WriteRequestText(req, body)
		return nil
	case MimeMultipartFormData:
		form, ok := any(body).(*MultipartForm)
		if !ok {
			return fmt.Errorf("multipart request body must be *MultipartForm, got %T", body)
		}
		WriteRequestFormMultipart(req, form)
		return nil
	default:
		return fmt.Errorf("unsupported request MIME type %s", mime)
	}
}

// WriteRequestJson encodes the body as JSON and writes it to the HTTP request.
// Sets the Content-Type header to "application/json".
//
// Returns an error if JSON encoding fails.
func WriteRequestJson[T any](req *http.Request, body *T) error {
	if body == nil {
		return nil
	}
	req.Header.Set("Content-Type", string(MimeApplicationJson))
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to encode json request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(jsonBody))
	req.ContentLength = int64(len(jsonBody))
	return nil
}

// WriteRequestFormMultipart writes a multipart form to the HTTP request.
// Sets the Content-Type header to "multipart/form-data" with the appropriate boundary.
//
// The form is written asynchronously using a pipe. The function returns immediately,
// and the form parts are written in a goroutine.
func WriteRequestFormMultipart(req *http.Request, form *MultipartForm) {
	if form == nil {
		return
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Body = pr
	req.ContentLength = -1

	go func() {
		defer pw.Close()
		defer writer.Close()

		for _, part := range form.Parts {
			if err := part.Write(writer); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()
}

// WriteRequestText converts the body to a string and writes it to the HTTP request.
// Sets the Content-Type header to "text/plain".
//
// The body is converted to a string using fmt.Sprintf("%v", body).
func WriteRequestText[T any](req *http.Request, body *T) {
	if body == nil {
		return
	}
	textBody := fmt.Sprintf("%v", *body)
	req.Header.Set("Content-Type", string(MimeTextPlain))
	req.Body = io.NopCloser(strings.NewReader(textBody))
	req.ContentLength = int64(len(textBody))
}

// ReadResponseBody reads the HTTP response body based on the specified MIME type.
//
// Supported MIME types:
//   - MimeApplicationJson: Decodes JSON response into body
//   - MimeTextPlain: Reads response as plain text (body must be *string)
//   - MimeOmit: Skips reading body
//
// Returns an error if the MIME type is unsupported or the body cannot be read.
func ReadResponseBody[T any](resp *http.Response, mime Mime, body *T) error {
	switch mime {
	case MimeOmit:
		return nil
	case MimeApplicationJson:
		return ReadResponseJson(resp, body)
	case MimeTextPlain:
		str, ok := any(body).(*string)
		if !ok {
			return fmt.Errorf("text response body must be *string, got %T", body)
		}
		return ReadResponseText(resp, str)
	default:
		return fmt.Errorf("unsupported response MIME type %s", mime)
	}
}

// ReadResponseJson decodes a JSON response body into the result pointer.
//
// Closes the response body after reading.
//
// Returns an error if JSON decoding fails.
func ReadResponseJson[T any](resp *http.Response, result *T) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if strictErr := dec.Decode(result); strictErr != nil {
		if err = json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to decode json response body: %w", err)
		}

		slog.Warn("JSON response contains unknown fields",
			util.NewLogArgsExtractor().AddError(strictErr).AddResponse(resp).Extract()...)
	}

	return nil
}

// ReadResponseText reads the response body as plain text.
//
// Closes the response body after reading.
//
// Returns an error if reading fails.
func ReadResponseText(resp *http.Response, result *string) error {
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	*result = string(data)
	return nil
}
