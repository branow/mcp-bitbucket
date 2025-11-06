package util

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type UrlBuilder struct {
	BaseUrl     string
	Path        []string
	QueryParams map[string]string
}

func (b *UrlBuilder) Build() (string, error) {
	u, err := url.Parse(b.BaseUrl)
	if err != nil {
		return "", err
	}

	u = u.JoinPath(b.Path...)

	q := u.Query()
	for key, value := range b.QueryParams {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func CreateRequest(method string, urlBuilder UrlBuilder, body []byte) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	url, err := urlBuilder.Build()
	if err != nil {
		slog.Error(
			"Failed to build URL",
			NewLogArgsExtractor().AddUrlBuilder(urlBuilder).AddError(err).Extract()...,
		)
		return nil, err
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		slog.Error(
			"Failed to create request",
			NewLogArgsExtractor().AddUrl(url).AddError(err).Extract()...,
		)
		return req, err
	}
	return req, nil
}

func DoRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(
			"Failed to perform request",
			NewLogArgsExtractor().AddRequest(req).AddError(err).Extract()...,
		)
		return resp, err
	}

	logArgs := NewLogArgsExtractor().AddResponse(resp).Extract()

	switch {
	case resp.StatusCode >= 500:
		slog.Error("HTTP request failed with server error", logArgs...)
	case resp.StatusCode >= 400:
		slog.Warn("HTTP request failed with client error", logArgs...)
	default:
		slog.Info("HTTP request completed", logArgs...)
	}

	return resp, nil
}

func ReadResponseBytes(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(
			"Failed to read response body",
			NewLogArgsExtractor().AddResponse(resp).AddError(err).Extract()...,
		)
		return nil, err
	}
	return body, nil
}

func ReadResponseJson[T any](resp *http.Response, result T) error {
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		slog.Error(
			"Failed to decode json response body",
			NewLogArgsExtractor().AddResponse(resp).AddError(err).Extract()...,
		)
		return err
	}
	return nil
}
