package bitbucket

import (
	"context"
	"strings"

	"github.com/branow/mcp-bitbucket/internal/util"
	sch "github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"golang.org/x/sync/errgroup"
)

// Service provides high-level operations for interacting with Bitbucket.
// It wraps the API client and the git client, delegating to each accordingly.
type Service struct {
	api *ApiClient
	git *GitClient
}

// NewService creates a new Bitbucket service with the given API client and git client.
func NewService(client *ApiClient, gitClient *GitClient) *Service {
	return &Service{api: client, git: gitClient}
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

	resp, err := s.api.ListRepositories(ctx, workspace, page, size)
	if err != nil {
		return nil, err
	}
	return MapPage(resp, MapRepository), nil
}

// ListPullRequests retrieves a paginated list of pull requests from the
// specified repository with optional filtering by state. By default returns
// only open pull requests.
//
// Parameters:
//   - ctx: Context for the request
//   - options: Configuration for the list operation
//
// Returns a Page containing PullRequestSummary items, or an error if the request fails.
func (s *Service) ListPullRequests(
	ctx context.Context,
	options ListPullRequestsOptions,
) (*Page[PullRequestSummary], error) {

	workspace, err := sch.Validate(options.Workspace, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("workspace: " + err.Error())
	}
	repository, err := sch.Validate(options.Repository, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("repository: " + err.Error())
	}
	page := sch.Validate(options.Page, sch.Positive()).Optional(1)
	size := sch.Validate(options.PageSize, sch.Positive()).Optional(25)
	states := options.State

	resp, err := s.api.ListPullRequests(ctx, workspace, repository, size, page, states)
	if err != nil {
		return nil, err
	}

	return MapPage(resp, MapPullRequestSummary), nil
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
		repo, err = s.api.GetRepository(ctx, workspace, repository)
		return err
	})

	if options.IncludeSource || options.IncludeReadme {
		g.Go(func() error {
			var err error
			src, err = s.api.GetRepositorySource(ctx, workspace, repository)
			if err != nil {
				return err
			}

			if options.IncludeReadme {
				if readmeSrc = findReadmeInSource(src.Values); readmeSrc != nil {
					readmeContent, err = s.api.GetFileSource(ctx, workspace, repository, readmeSrc.Commit.Hash, readmeSrc.Path)
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

// GetDirectorySource retrieves the contents of a directory at a specific commit
// or branch in a repository.
//
// Parameters:
//   - ctx: Context for the request
//   - options: Configuration for the operation including workspace, repository,
//     commit/branch, and directory path
//
// Returns a paginated list of source items (files and subdirectories),
// or an error if the request fails.
func (s *Service) GetDirectorySource(
	ctx context.Context,
	options GetDirectorySourceOptions,
) (*Page[SourceItem], error) {

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
		repo, err := s.api.GetRepository(ctx, workspace, repository)
		if err != nil {
			return nil, err
		}
		ref = repo.MainBranch.Name
	}

	resp, err := s.api.GetDirectorySource(ctx, workspace, repository, ref, path)
	if err != nil {
		return nil, err
	}
	return MapPage(resp, MapSourceItem), nil
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
		pr, err = s.api.GetPullRequest(ctx, workspace, repository, id)
		return err
	})

	if options.IncludeCommits {
		g.Go(func() error {
			var err error
			commits, err = s.api.ListPullRequestCommits(ctx, workspace, repository, id)
			return err
		})
	}

	if options.IncludeDiff {
		g.Go(func() error {
			var err error
			diff, err = s.api.GetPullRequestDiff(ctx, workspace, repository, id)
			return err
		})
	}

	if options.IncludeComments {
		g.Go(func() error {
			var err error
			comments, err = s.api.ListPullRequestComments(ctx, workspace, repository, id, 50, 1)
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
		repo, err := s.api.GetRepository(ctx, workspace, repository)
		if err != nil {
			return nil, err
		}
		ref = repo.MainBranch.Name
	}

	// Fetch file content
	content, err := s.api.GetFileSource(ctx, workspace, repository, ref, path)
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

// CloneRepository clones a Bitbucket repository to a local path.
// Validation is performed on the input options before delegating to the GitClient.
// Authentication is handled by the GitClient — the token is never exposed to the agent.
//
// Parameters:
//   - ctx: The request context
//   - options: Clone parameters (workspace, repository, target path, depth, ref)
//
// Returns the resolved absolute path of the cloned repository, or an error if
// validation fails or the clone operation does not succeed.
func (s *Service) CloneRepository(
	ctx context.Context,
	options CloneRepositoryOptions,
) (*CloneRepositoryResult, error) {

	workspace, err := sch.Validate(options.Workspace, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("workspace: " + err.Error())
	}

	repository, err := sch.Validate(options.Repository, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("repository: " + err.Error())
	}

	targetPath, err := sch.Validate(options.TargetPath, sch.NotBlank()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("target_path: " + err.Error())
	}

	// depth=0 means full clone; only validate if the caller explicitly set a depth.
	depth, err := sch.Validate(options.Depth, sch.NonNegative()).Get()
	if err != nil {
		return nil, util.NewInvalidParamsError("depth: " + err.Error())
	}

	if options.PullIfExists && IsRepository(targetPath) {
		cloneURL, err := (&web.UrlBuilder{
			BaseUrl: s.git.cfg.BaseURL,
			Path:    []string{workspace, repository},
		}).Build()
		if err != nil {
			return nil, util.NewInvalidParamsError("failed to build clone URL: " + err.Error())
		}

		absPath, err := s.git.Pull(ctx, targetPath, cloneURL, options.Ref)
		if err != nil {
			return nil, err
		}
		return &CloneRepositoryResult{Path: absPath}, nil
	}

	absPath, err := s.git.Clone(ctx, workspace, repository, targetPath, depth, options.Ref)
	if err != nil {
		return nil, err
	}

	return &CloneRepositoryResult{Path: absPath}, nil
}
