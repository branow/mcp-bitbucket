package web

import "net/url"

// UrlBuilder constructs URLs by combining a base URL with path segments
// and query parameters.
//
// Example:
//
//	builder := &UrlBuilder{
//	  BaseUrl:     "https://api.example.com",
//	  Path:        []string{"repositories", "owner", "repo"},
//	  QueryParams: map[string]string{"page": "1", "limit": "10"},
//	}
//	url, err := builder.Build()
//	// url = "https://api.example.com/repositories/owner/repo?limit=10&page=1"
type UrlBuilder struct {
	// BaseUrl is the base URL including scheme and host (e.g., "https://api.example.com").
	BaseUrl string
	// Path is a list of path segments to append to the base URL.
	Path []string
	// QueryParams is a map of query parameter names to values.
	QueryParams map[string]string
}

// Build constructs the final URL string from the base URL, path segments,
// and query parameters.
//
// Returns an error if the base URL cannot be parsed.
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
