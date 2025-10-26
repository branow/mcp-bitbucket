package util

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func CreateRequest(method string, url string, body []byte) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return req, fmt.Errorf("failed to create %q request to %q: %w", method, url, err)
	}
	return req, nil
}

func DoRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("failed to make %q request to %q: %w", req.Method, req.URL, err)
	}

	slog.Info("HTTP request completed",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
	)

	return resp, nil
}

func ReadResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q request to %q: %w", resp.Request.Method, resp.Request.URL, err)
	}
	return body, nil
}

func JoinUrlPath(segments ...string) string {
	var parts []string
	for _, s := range segments {
		s = strings.Trim(s, "/")
		if s != "" {
			parts = append(parts, s)
		}
	}
	return "/" + strings.Join(parts, "/")
}
