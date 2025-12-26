// Package client provides core functionality for making HTTP requests to the Bitbucket API.
// This file contains shared request/response handling logic used by the Client.
package client

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/web"
)

var (
	// ErrInternal indicates an internal error occurred while building or sending the request.
	ErrInternal = errors.New("Failed to make request to bitbucket")
	// ErrServerBitbucket indicates the Bitbucket API returned a 5xx server error.
	ErrServerBitbucket = errors.New("Bitbucket service is currently unavailable")
	// ErrClientBitbucket indicates the Bitbucket API returned a 4xx client error.
	ErrClientBitbucket = errors.New("Bitbucket failed to process request")
)

// BitbucketRequest represents an HTTP request to the Bitbucket API.
// It encapsulates all parameters needed to build and execute the request.
type BitbucketRequest[T any] struct {
	// Method is the HTTP method (GET, POST, PUT, DELETE, etc.)
	Method string
	// BaseUrl is the base URL of the Bitbucket API (e.g., "https://api.bitbucket.org/2.0")
	BaseUrl string
	// Path contains URL path segments that will be joined with the base URL
	Path []string
	// Query contains URL query parameters
	Query map[string]string
	// Body is the request body to be serialized (nil for GET requests)
	Body *T
	// Mime specifies the Content-Type for the request body
	Mime web.Mime
	// Username is used for HTTP basic authentication
	Username string
	// Password is the API token used for HTTP basic authentication
	Password string
	// Client is the HTTP client used to execute the request
	Client *http.Client
}

// BitbucketResponse represents an HTTP response from the Bitbucket API.
// It specifies how to deserialize the response body.
type BitbucketResponse[T any] struct {
	// Body is a pointer to the structure where the response will be deserialized
	Body *T
	// Mime specifies the expected Content-Type of the response
	Mime web.Mime
}

// Perform executes a Bitbucket API request and deserializes the response.
//
// It builds the HTTP request from the provided BitbucketRequest, executes it,
// and processes the response according to the BitbucketResponse specification.
//
// Returns an error if:
//   - The request cannot be built (returns ErrInternal)
//   - The HTTP request fails (returns ErrInternal)
//   - The API returns a 5xx error (returns ErrServerBitbucket)
//   - The API returns a 4xx error (returns ErrClientBitbucket with details)
//   - The response cannot be deserialized (returns ErrInternal)
func Perform[T, U any](bbReq *BitbucketRequest[T], bbResp *BitbucketResponse[U]) error {
	req, err := buildRequest(bbReq)
	if err != nil {
		return err
	}

	resp, err := bbReq.Client.Do(req)
	if err != nil {
		slog.Error("Failed to perform request", util.NewLogArgsExtractor().AddError(err).AddRequest(req).Extract()...)
		return ErrInternal
	}

	return readResponse(resp, bbResp)
}

func buildRequest[T any](bbReq *BitbucketRequest[T]) (*http.Request, error) {
	url := web.UrlBuilder{
		BaseUrl:     bbReq.BaseUrl,
		Path:        bbReq.Path,
		QueryParams: bbReq.Query,
	}

	req, err := (&web.RequestBuilder[T]{
		Method: bbReq.Method,
		Url:    url,
		Mime:   web.Mime(bbReq.Mime),
		Body:   bbReq.Body,
	}).Build()

	if err != nil {
		slog.Error("Failed to build request", util.NewLogArgsExtractor().AddError(err).AddRequest(req).Extract()...)
		return nil, ErrInternal
	}

	req.SetBasicAuth(bbReq.Username, bbReq.Password)

	return req, nil
}

func readResponse[T any](resp *http.Response, bbResp *BitbucketResponse[T]) error {
	switch {
	case resp.StatusCode >= 500:
		return ErrServerBitbucket
	case resp.StatusCode >= 400:
		errResp := &ErrorResponse{}
		err := web.ReadResponseJson(resp, errResp)
		if err != nil {
			return fmt.Errorf("%w: %d", ErrClientBitbucket, resp.StatusCode)
		}
		if errResp.Error.Message != "" {
			return fmt.Errorf("%w: %s", ErrClientBitbucket, errResp.Error.Message)
		}
		return ErrClientBitbucket
	}

	if err := web.ReadResponseBody(resp, bbResp.Mime, bbResp.Body); err != nil {
		slog.Error("Failed to read response", util.NewLogArgsExtractor().AddError(err).AddResponse(resp).Extract()...)
		return ErrInternal
	}

	return nil
}
