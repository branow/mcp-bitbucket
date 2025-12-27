package util

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// UriParams holds extracted parameters from a URI.
// Path contains parameters extracted from the URI path segments.
// Query contains parameters extracted from the URI query string.
type UriParams struct {
	Path  map[string]string
	Query map[string]string
}

// ParseUriParams extracts parameters from a URI based on a template.
// This is a convenience function that creates a parser and parses the URI in one call.
//
// Template format:
//   - Path parameters: {paramName} in path segments, e.g., "/users/{id}/posts/{postId}"
//   - Query parameters: {paramName} in query values, e.g., "?page={page}&size={size}"
//
// Important limitations and behaviors:
//   - Placeholders are NOT supported in scheme, host, or port portions of the template
//   - URL decoding is automatically applied by Go's url.Parse:
//   - Query parameter "+" characters are decoded to spaces
//   - Percent-encoded characters (e.g., %20, %2B) are decoded
//   - Query parameters that are missing from the URI will be included in the result with empty string values
//   - Scheme comparison is case-insensitive (HTTP and http are treated as the same)
//   - Host comparison is exact and case-sensitive (localhost != 127.0.0.1, ports must match exactly)
//   - Path comparison is case-sensitive
//   - Empty templates or URIs will result in validation errors
//
// Example:
//
//	params, err := ParseUriParams(
//	  "https://api.example.com/repos/{owner}/{repo}/issues?state={state}&page={page}",
//	  "https://api.example.com/repos/golang/go/issues?state=open&page=3",
//	)
//	// params.Path = {"owner": "golang", "repo": "go"}
//	// params.Query = {"state": "open", "page": "3"}
func ParseUriParams(template, uri string) (*UriParams, error) {
	parser, err := NewUriTemplateParser(template)
	if err != nil {
		return nil, err
	}
	return parser.Parse(uri)
}

// UriTemplateParser parses URIs against a template to extract parameters.
type UriTemplateParser struct {
	Template *url.URL
}

// NewUriTemplateParser creates a new URI template parser.
// The template is parsed and validated immediately. If the template is invalid,
// an error is returned.
//
// Template syntax:
//   - Use {paramName} in path segments to define path parameters
//   - Use {paramName} in query values to define query parameters
//   - Placeholders in scheme, host, or port are NOT supported and will cause parse errors
//
// Returns an error if:
//   - The template cannot be parsed as a valid URI
//   - The template contains invalid characters
//   - The template uses placeholders in unsupported positions (host, port, scheme)
func NewUriTemplateParser(template string) (*UriTemplateParser, error) {
	templateUrl, err := url.Parse(template)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}
	return &UriTemplateParser{Template: templateUrl}, nil
}

// Parse extracts parameters from the given URI based on the template.
//
// The URI must match the template in:
//   - Scheme (case-insensitive, HTTP and http are the same)
//   - Host (exact match including case and port if present)
//   - Path structure (same number of segments, literals must match)
//
// Path parameters are extracted from segments marked with {paramName} in the template.
// Query parameters are extracted from query values marked with {paramName} in the template.
//
// Important behaviors:
//   - URL decoding is applied automatically (spaces, percent-encoding)
//   - Query parameters missing from the URI will have empty string values in the result
//   - Extra query parameters in the URI that aren't in the template are ignored
//   - Path segments are compared after URL decoding
//
// Returns an error if:
//   - The URI cannot be parsed
//   - Scheme or host doesn't match
//   - Path segment count doesn't match
//   - Path literal segments don't match
func (p *UriTemplateParser) Parse(uri string) (*UriParams, error) {
	actualUrl, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	if p.Template.Scheme != actualUrl.Scheme {
		return nil, fmt.Errorf("scheme mismatch: expected %s, got %s", p.Template.Scheme, actualUrl.Scheme)
	}
	if p.Template.Host != actualUrl.Host {
		return nil, fmt.Errorf("host mismatch: expected %s, got %s", p.Template.Host, actualUrl.Host)
	}

	pathParams, err := extractPathParams(p.Template.Path, actualUrl.Path)
	if err != nil {
		return nil, err
	}

	queryParams, err := extractQueryParams(p.Template.RawQuery, actualUrl.Query())
	if err != nil {
		return nil, err
	}

	return &UriParams{
		Path:  pathParams,
		Query: queryParams,
	}, nil
}

func extractPathParams(templatePath, actualPath string) (map[string]string, error) {
	params := make(map[string]string)
	templateSegments := strings.Split(strings.Trim(templatePath, "/"), "/")
	actualSegments := strings.Split(strings.Trim(actualPath, "/"), "/")

	if len(templateSegments) != len(actualSegments) {
		return params, fmt.Errorf("path segment count mismatch: expected %d, got %d", len(templateSegments), len(actualSegments))
	}

	paramRegex := regexp.MustCompile(`^\{([^}]+)\}$`)

	for i, templateSegment := range templateSegments {
		if matches := paramRegex.FindStringSubmatch(templateSegment); matches != nil {
			paramName := matches[1]
			params[paramName] = actualSegments[i]
		} else {
			if templateSegment != actualSegments[i] {
				return params, fmt.Errorf("path segment mismatch at position %d: expected %s, got %s", i, templateSegment, actualSegments[i])
			}
		}
	}

	return params, nil
}

func extractQueryParams(templateQuery string, actualQuery url.Values) (map[string]string, error) {
	params := make(map[string]string)

	if templateQuery == "" {
		return params, nil
	}

	paramRegex := regexp.MustCompile(`\{([^}]+)\}`)
	templatePairs := strings.Split(templateQuery, "&")

	for _, pair := range templatePairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		queryKey := parts[0]
		templateValue := parts[1]

		if matches := paramRegex.FindStringSubmatch(templateValue); matches != nil {
			paramName := matches[1]
			params[paramName] = actualQuery.Get(queryKey)
		}
	}

	return params, nil
}
