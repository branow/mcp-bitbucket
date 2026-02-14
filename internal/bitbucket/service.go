package bitbucket

import (
	"context"
	"strings"

	"github.com/branow/mcp-bitbucket/internal/util"
	sch "github.com/branow/mcp-bitbucket/internal/util/schema"
	"golang.org/x/sync/errgroup"
)

// Service provides high-level operations for interacting with Bitbucket.
// It wraps the Bitbucket API client and handles mapping between API types
// and domain types.
type Service struct {
	client *Client
}

// NewService creates a new Bitbucket service with the given client.
func NewService(client *Client) *Service {
	return &Service{client: client}
}

// ListRepositories retrieves a paginated list of repositories from the
// specified workspace.
//
// Parameters:
//   - ctx: Context for the request
//   - options: Configuration for the list operation
//
// Returns a Page containing Repository items, or an error if the request fails.
func (s *Service) ListRepositories(
	ctx context.Context,
	options ListRepositoriesOptions,
) (*Page[Repository], error) {

	workspace, err := sch.Validate(options.Workspace, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("workspace: " + err.Error())
	}
	page := sch.Validate(options.Page, sch.Positive()).Optional(1)
	size := sch.Validate(options.PageSize, sch.Positive()).Optional(50)

	resp, err := s.client.ListRepositories(ctx, workspace, page, size)
	if err != nil {
		return nil, err
	}
	return MapPage(resp, MapRepository), nil
}

// GetRepository retrieves detailed information about a specific repository.
// It can optionally fetch the root-level source listing and README content
// in parallel.
//
// Parameters:
//   - ctx: Context for the request
//   - options: Configuration for the operation and additional data to fetch
//
// Returns detailed repository information, or an error if the request fails.
func (s *Service) GetRepository(
	ctx context.Context,
	options GetRepositoryOptions,
) (*RepositoryDetails, error) {

	workspace, err := sch.Validate(options.Workspace, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("workspace: " + err.Error())
	}
	repository, err := sch.Validate(options.Repository, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("repository: " + err.Error())
	}

	g, ctx := errgroup.WithContext(ctx)

	var repo *ApiRepository
	var src *ApiResponse[ApiSourceItem]
	var readmeSrc *ApiSourceItem
	var readmeContent *string

	g.Go(func() error {
		var err error
		repo, err = s.client.GetRepository(ctx, workspace, repository)
		return err
	})

	if options.IncludeSource || options.IncludeReadme {
		g.Go(func() error {
			var err error
			src, err = s.client.GetRepositorySource(ctx, workspace, repository)
			if err != nil {
				return err
			}

			if options.IncludeReadme {
				if readmeSrc = findReadmeInSource(src.Values); readmeSrc != nil {
					readmeContent, err = s.client.GetFileSource(ctx, workspace, repository, readmeSrc.Commit.Hash, readmeSrc.Path)
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

func findReadmeInSource(items []ApiSourceItem) *ApiSourceItem {
	for i, item := range items {
		if strings.HasPrefix(strings.ToLower(item.Path), "readme.") {
			return &items[i]
		}
	}
	return nil
}

// GetPullRequest retrieves detailed information about a specific pull request.
// It can optionally fetch commits, diff, and comments in parallel.
//
// Parameters:
//   - ctx: Context for the request
//   - options: Configuration for the operation and additional data to fetch
//
// Returns detailed pull request information, or an error if the request fails.
func (s *Service) GetPullRequest(
	ctx context.Context,
	options GetPullRequestOptions,
) (*PullRequestDetails, error) {

	workspace, err := sch.Validate(options.Workspace, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("workspace: " + err.Error())
	}
	repository, err := sch.Validate(options.Repository, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("repository: " + err.Error())
	}
	id, err := sch.Validate(options.Id, sch.Positive()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("id: " + err.Error())
	}

	g, ctx := errgroup.WithContext(ctx)

	var pr *ApiPullRequest
	var commits *ApiResponse[ApiCommit]
	var diff *string
	var comments *ApiResponse[ApiPullRequestComment]

	g.Go(func() error {
		var err error
		pr, err = s.client.GetPullRequest(ctx, workspace, repository, id)
		return err
	})

	if options.IncludeCommits {
		g.Go(func() error {
			var err error
			commits, err = s.client.ListPullRequestCommits(ctx, workspace, repository, id)
			return err
		})
	}

	if options.IncludeDiff {
		g.Go(func() error {
			var err error
			diff, err = s.client.GetPullRequestDiff(ctx, workspace, repository, id)
			return err
		})
	}

	if options.IncludeComments {
		g.Go(func() error {
			var err error
			comments, err = s.client.ListPullRequestComments(ctx, workspace, repository, id, 50, 1)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return MapPullRequestDetails(pr, commits, diff, comments), nil
}

// GetFileContent retrieves the content of a file from a repository.
//
// Parameters:
//   - ctx: Context for the request
//   - options: Configuration for the file content retrieval
//
// Returns detailed file content with metadata, or an error if the request fails.
func (s *Service) GetFileContent(
	ctx context.Context,
	options GetFileContentOptions,
) (*FileContent, error) {

	workspace, err := sch.Validate(options.Workspace, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("workspace: " + err.Error())
	}
	repository, err := sch.Validate(options.Repository, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("repository: " + err.Error())
	}
	path, err := sch.Validate(options.Path, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("path: " + err.Error())
	}

	// Resolve ref: if empty, fetch repository to get main branch
	ref := options.Ref
	if ref == "" {
		repo, err := s.client.GetRepository(ctx, workspace, repository)
		if err != nil {
			return nil, err
		}
		ref = repo.MainBranch.Name
	}

	// Fetch file content
	content, err := s.client.GetFileSource(ctx, workspace, repository, ref, path)
	if err != nil {
		return nil, err
	}

	if content == nil {
		return nil, util.NewInternalError()
	}

	size := len(*content)

	result := &FileContent{
		Path:    path,
		Size:    size,
		Commit:  ref,
		Content: content,
	}

	return result, nil
}
