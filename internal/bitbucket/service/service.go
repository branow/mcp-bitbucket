// Package service provides a high-level service layer for interacting with the Bitbucket API.
//
// This package wraps the low-level Bitbucket API client and provides domain-specific
// methods with mapped types that are more suitable for application use.
package service

import (
	"context"
	"strings"

	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
	"golang.org/x/sync/errgroup"
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

// GetRepositoryOptions configures what additional data to fetch with the repository.
type GetRepositoryOptions struct {
	IncludeSource bool // Include the root-level source listing (1 level depth)
	IncludeReadme bool // Include the README file content if found in root
}

// GetRepository retrieves detailed information about a specific repository.
// It can optionally fetch the root-level source listing and README content in parallel.
//
// Parameters:
//   - ctx: Context for the request
//   - namespace: The workspace slug or username
//   - name: The repository name/slug
//   - options: Configuration for additional data to fetch
//
// Returns detailed repository information, or an error if the request fails.
func (s *Service) GetRepository(ctx context.Context, namespace string, name string, options GetRepositoryOptions) (*RepositoryDetails, error) {
	g, ctx := errgroup.WithContext(ctx)

	var repo *client.Repository
	var src *client.ApiResponse[client.SourceItem]
	var readmeSrc *client.SourceItem
	var readmeContent *string

	g.Go(func() error {
		var err error
		repo, err = s.client.GetRepository(ctx, namespace, name)
		return err
	})

	if options.IncludeSource || options.IncludeReadme {
		g.Go(func() error {
			var err error
			src, err = s.client.GetRepositorySource(ctx, namespace, name)
			if err != nil {
				return err
			}

			if options.IncludeReadme {
				if readmeSrc = findReadmeInSource(src.Values); readmeSrc != nil {
					readmeContent, err = s.client.GetFileSource(ctx, namespace, name, readmeSrc.Commit.Hash, readmeSrc.Path)
				}
			}

			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if !options.IncludeSource {
		src = nil
	}

	return MapRepositoryDetails(repo, src, readmeSrc, readmeContent), nil
}

func findReadmeInSource(items []client.SourceItem) *client.SourceItem {
	for i, item := range items {
		if strings.HasPrefix(strings.ToLower(item.Path), "readme.") {
			return &items[i]
		}
	}
	return nil
}

// GetPullRequestOptions configures what additional data to fetch with the pull request.
type GetPullRequestOptions struct {
	IncludeCommits  bool // Include the pull request commits
	IncludeDiff     bool // Include the pull request diff
	IncludeComments bool // Include the pull request comments
}

// GetPullRequest retrieves detailed information about a specific pull request.
// It can optionally fetch commits, diff, and comments in parallel.
//
// Parameters:
//   - ctx: Context for the request
//   - namespace: The workspace slug or username
//   - repoSlug: The repository name/slug
//   - pullRequestId: The pull request ID
//   - options: Configuration for additional data to fetch
//
// Returns detailed pull request information, or an error if the request fails.
func (s *Service) GetPullRequest(ctx context.Context, namespace string, repoSlug string, pullRequestId int, options GetPullRequestOptions) (*PullRequestDetails, error) {
	g, ctx := errgroup.WithContext(ctx)

	var pr *client.PullRequest
	var commits *client.ApiResponse[client.Commit]
	var diff *string
	var comments *client.ApiResponse[client.PullRequestComment]

	g.Go(func() error {
		var err error
		pr, err = s.client.GetPullRequest(ctx, namespace, repoSlug, pullRequestId)
		return err
	})

	if options.IncludeCommits {
		g.Go(func() error {
			var err error
			commits, err = s.client.ListPullRequestCommits(ctx, namespace, repoSlug, pullRequestId)
			return err
		})
	}

	if options.IncludeDiff {
		g.Go(func() error {
			var err error
			diff, err = s.client.GetPullRequestDiff(ctx, namespace, repoSlug, pullRequestId)
			return err
		})
	}

	if options.IncludeComments {
		g.Go(func() error {
			var err error
			comments, err = s.client.ListPullRequestComments(ctx, namespace, repoSlug, pullRequestId, 50, 1)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return MapPullRequestDetails(pr, commits, diff, comments), nil
}
