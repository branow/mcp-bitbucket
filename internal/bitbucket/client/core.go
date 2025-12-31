// Package client provides core functionality for making HTTP requests to the Bitbucket API.
// This file contains shared request/response handling logic used by the Client.
package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/web"
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
	// Authorizer handles authentication for the request (basic auth or OAuth)
	Authorizer util.Authorizer
	// Context is the request context containing authentication tokens and cancellation
	Context context.Context
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
//   - The request cannot be built (returns util.NewInternalError)
//   - The HTTP request fails (returns util.NewInternalError)
//   - The API returns a 5xx error (returns util.NewResourceUnavailableError)
//   - The API returns a 404 error (returns util.NewResourceNotFoundError)
//   - The API returns other 4xx errors (returns util.NewInvalidParamsError)
//   - The response cannot be deserialized (returns util.NewInternalError)
func Perform[T, U any](bbReq *BitbucketRequest[T], bbResp *BitbucketResponse[U]) error {
	req, err := buildRequest(bbReq)
	if err != nil {
		return err
	}

	resp, err := bbReq.Client.Do(req)
	if err != nil {
		slog.Error("Failed to perform request", util.NewLogArgsExtractor().AddError(err).AddRequest(req).Extract()...)
		return util.NewInternalError()
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
		return nil, util.NewInternalError()
	}

	if bbReq.Authorizer.Authorize(bbReq.Context, req) != nil {
		slog.Error("Authorization failed", util.NewLogArgsExtractor().AddError(err).AddRequest(req).Extract()...)
		return nil, util.NewInternalError()
	}

	return req, nil
}

func readResponse[T any](resp *http.Response, bbResp *BitbucketResponse[T]) error {
	switch {
	case resp.StatusCode >= 500:
		return util.NewResourceUnavailableError(fmt.Sprintf("Bitbucket service unavailable (status %d)", resp.StatusCode))
	case resp.StatusCode >= 400:
		errResp := &ErrorResponse{}
		err := web.ReadResponseJson(resp, errResp)

		if resp.StatusCode == 404 {
			message := fmt.Sprintf("Resource not found at %s", resp.Request.URL.String())
			if err == nil && errResp.Error.Message != "" {
				message = errResp.Error.Message
			}
			return util.NewResourceNotFoundError(message)
		}

		message := fmt.Sprintf("Bitbucket API error (status %d)", resp.StatusCode)
		if err == nil && errResp.Error.Message != "" {
			message = errResp.Error.Message
		}
		return util.NewInvalidParamsError(message)
	}

	if err := web.ReadResponseBody(resp, bbResp.Mime, bbResp.Body); err != nil {
		slog.Error("Failed to read response", util.NewLogArgsExtractor().AddError(err).AddResponse(resp).Extract()...)
		return util.NewInternalError()
	}

	return nil
}
