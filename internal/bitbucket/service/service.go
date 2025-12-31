// Package service provides a high-level service layer for interacting with the Bitbucket API.
//
// This package wraps the low-level Bitbucket API client and provides domain-specific
// methods with mapped types that are more suitable for application use.
package service

import (
	"context"

	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
)

// Service provides high-level operations for interacting with Bitbucket.
// It wraps the Bitbucket API client and handles mapping between API types and domain types.
type Service struct {
	client *client.Client
}

// NewService creates a new Bitbucket service with the given client.
func NewService(client *client.Client) *Service {
	return &Service{client: client}
}

// ListRepositories retrieves a paginated list of repositories from the specified namespace.
// It returns the repositories mapped to the domain Repository type.
//
// Parameters:
//   - ctx: Context for the request
//   - namespace: The workspace slug or username
//   - page: The page number (1-based)
//   - size: The number of items per page
//
// Returns a Page containing Repository items, or an error if the request fails.
func (s *Service) ListRepositories(ctx context.Context, namespace string, page, size int) (*Page[Repository], error) {
	resp, err := s.client.ListRepositories(ctx, namespace, page, size)
	if err != nil {
		return nil, err
	}
	return MapPage(resp, MapRepository), nil
}
