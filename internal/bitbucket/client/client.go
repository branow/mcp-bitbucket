// Package client provides a Bitbucket API client for interacting with repositories,
// pull requests, and source code.
package client

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/web"
)

// BitbucketConfig provides configuration for creating a Bitbucket API client.
type BitbucketConfig struct {
	// Url is the base URL of the Bitbucket API (e.g., "https://api.bitbucket.org/2.0")
	Url string
	// Timeout is the HTTP request timeout in seconds
	Timeout int
}

// Client is a Bitbucket API client that provides methods for accessing
// repositories, pull requests, and source code.
type Client struct {
	cfg        BitbucketConfig
	authorizer util.Authorizer
	client     *http.Client
}

// NewClient creates a new Bitbucket API client with the provided configuration.
// The client uses provided authorizer for request authentication.
// The timeout specified in the config is applied to all HTTP requests.
func NewClient(config BitbucketConfig, authorizer util.Authorizer) *Client {
	return &Client{
		cfg:        config,
		authorizer: authorizer,
		client:     &http.Client{Timeout: time.Duration(config.Timeout) * time.Second},
	}
}

// CreateRepository creates a new repository in the specified workspace.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier where the repository will be created
//   - repoSlug: The desired repository slug identifier (URL-friendly name)
//   - body: Repository configuration including SCM type, privacy settings, and optional metadata
//
// Returns the created repository details including generated metadata and links.
func (c *Client) CreateRepository(ctx context.Context, workspaceSlug string, repoSlug string, body *CreateRepositoryRequest) (*Repository, error) {
	resp := &BitbucketResponse[Repository]{
		Body: &Repository{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[CreateRepositoryRequest]{
		Method: "POST",
		Path:   []string{"repositories", workspaceSlug, repoSlug},
		Body:   body,
		Mime:   web.MimeApplicationJson,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// DeleteRepository permanently deletes a repository from the specified workspace.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//
// Returns an error if the deletion fails. This operation is irreversible.
func (c *Client) DeleteRepository(ctx context.Context, workspaceSlug string, repoSlug string) error {
	resp := &BitbucketResponse[Repository]{
		Mime: web.MimeOmit,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "DELETE",
		Path:   []string{"repositories", workspaceSlug, repoSlug},
		Mime:   web.MimeOmit,
	})

	return Perform(req, resp)
}

// ListRepositories retrieves a paginated list of repositories for the specified workspace.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - pagelen: Number of items per page
//   - page: Page number to retrieve (1-indexed)
//
// Returns the API response containing the list of repositories and pagination metadata.
func (c *Client) ListRepositories(ctx context.Context, workspaceSlug string, pagelen int, page int) (*ApiResponse[Repository], error) {
	resp := &BitbucketResponse[ApiResponse[Repository]]{
		Body: &ApiResponse[Repository]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug},
		Query: map[string]string{
			"pagelen": strconv.Itoa(pagelen),
			"page":    strconv.Itoa(page),
		},
		Mime: web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetRepository retrieves detailed information about a specific repository.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//
// Returns the repository details including metadata, links, and configuration.
func (c *Client) GetRepository(ctx context.Context, workspaceSlug string, repoSlug string) (*Repository, error) {
	resp := &BitbucketResponse[Repository]{
		Body: &Repository{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetRepositorySource retrieves the source code tree of a repository at the default branch.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//
// Returns the API response containing the list of files and directories at the repository root.
func (c *Client) GetRepositorySource(ctx context.Context, workspaceSlug string, repoSlug string) (*ApiResponse[SourceItem], error) {
	resp := &BitbucketResponse[ApiResponse[SourceItem]]{
		Body: &ApiResponse[SourceItem]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "src"},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ListPullRequests retrieves a paginated list of pull requests for the specified repository.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pagelen: Number of items per page (maximum 50)
//   - page: Page number to retrieve (1-indexed)
//   - states: Filter by pull request states (e.g., "OPEN", "MERGED", "DECLINED"). Empty slice returns all states.
//
// Returns the API response containing the list of pull requests and pagination metadata.
func (c *Client) ListPullRequests(ctx context.Context, workspaceSlug string, repoSlug string, pagelen int, page int, states []string) (*ApiResponse[PullRequest], error) {
	resp := &BitbucketResponse[ApiResponse[PullRequest]]{
		Body: &ApiResponse[PullRequest]{},
		Mime: web.MimeApplicationJson,
	}

	query := map[string]string{
		"pagelen": strconv.Itoa(pagelen),
		"page":    strconv.Itoa(page),
	}

	if len(states) > 0 {
		query["state"] = strings.Join(states, ",")
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests"},
		Query:  query,
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetPullRequest retrieves detailed information about a specific pull request.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//
// Returns the pull request details including title, description, state, author, reviewers, and metadata.
func (c *Client) GetPullRequest(ctx context.Context, workspaceSlug string, repoSlug string, pullRequestId int) (*PullRequest, error) {
	resp := &BitbucketResponse[PullRequest]{
		Body: &PullRequest{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId)},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ListPullRequestCommits retrieves all commits included in a specific pull request.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//
// Returns the API response containing the list of commits with their hash, message, author, and metadata.
func (c *Client) ListPullRequestCommits(ctx context.Context, workspaceSlug string, repoSlug string, pullRequestId int) (*ApiResponse[Commit], error) {
	resp := &BitbucketResponse[ApiResponse[Commit]]{
		Body: &ApiResponse[Commit]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "commits"},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ListPullRequestComments retrieves a paginated list of comments on a specific pull request.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//   - pagelen: Number of items per page
//   - page: Page number to retrieve (1-indexed)
//
// Returns the API response containing the list of comments with their content, author, and inline code references.
func (c *Client) ListPullRequestComments(ctx context.Context, workspaceSlug string, repoSlug string, pullRequestId int, pagelen int, page int) (*ApiResponse[PullRequestComment], error) {
	resp := &BitbucketResponse[ApiResponse[PullRequestComment]]{
		Body: &ApiResponse[PullRequestComment]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "comments"},
		Query: map[string]string{
			"pagelen": strconv.Itoa(pagelen),
			"page":    strconv.Itoa(page),
		},
		Mime: web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetPullRequestDiff retrieves the unified diff for a specific pull request.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//
// Returns the diff content as a plain text string in unified diff format.
func (c *Client) GetPullRequestDiff(ctx context.Context, workspaceSlug string, repoSlug string, pullRequestId int) (*string, error) {
	resp := &BitbucketResponse[string]{
		Body: new(string),
		Mime: web.MimeTextPlain,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "diff"},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetFileSource retrieves the raw content of a file at a specific commit.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - commit: The commit hash or branch name
//   - path: The file path relative to the repository root
//
// Returns the file content as a plain text string.
func (c *Client) GetFileSource(ctx context.Context, workspaceSlug string, repoSlug string, commit string, path string) (*string, error) {
	resp := &BitbucketResponse[string]{
		Body: new(string),
		Mime: web.MimeTextPlain,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "src", commit, path},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetDirectorySource retrieves the contents of a directory at a specific commit.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - commit: The commit hash or branch name
//   - path: The directory path relative to the repository root
//
// Returns the API response containing the list of files and subdirectories.
func (c *Client) GetDirectorySource(ctx context.Context, workspaceSlug string, repoSlug string, commit string, path string) (*ApiResponse[SourceItem], error) {
	resp := &BitbucketResponse[ApiResponse[SourceItem]]{
		Body: &ApiResponse[SourceItem]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "src", commit, path},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// CreateOrUpdateFiles creates or updates multiple files in a repository in a single commit.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - body: Request configuration including files, commit message, branch, and optional metadata
//
// This method uses the Bitbucket API's /src endpoint to create multiple files atomically
// in a single commit. File paths are used as form field names and should use forward slashes
// for nested directories. The API interprets paths as absolute from the repository root.
//
// The API returns 201 Created on success with no response body.
//
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-source/#api-repositories-workspace-repo-slug-src-post
func (c *Client) CreateOrUpdateFiles(ctx context.Context, workspaceSlug string, repoSlug string, body *CreateFilesRequest) error {
	form := &web.MultipartForm{
		Parts: []web.FormPart{},
	}

	// Add metadata as form fields (not query parameters!)
	if body.Message != "" {
		form.Parts = append(form.Parts, &web.TextField{
			Name:  "message",
			Value: body.Message,
		})
	}

	if body.Branch != "" {
		form.Parts = append(form.Parts, &web.TextField{
			Name:  "branch",
			Value: body.Branch,
		})
	}

	if body.Parents != "" {
		form.Parts = append(form.Parts, &web.TextField{
			Name:  "parents",
			Value: body.Parents,
		})
	}

	if body.Author != "" {
		form.Parts = append(form.Parts, &web.TextField{
			Name:  "author",
			Value: body.Author,
		})
	}

	// Add files as file fields
	for filePath, content := range body.Files {
		if !strings.HasPrefix(filePath, "/") {
			filePath = "/" + filePath
		}
		form.Parts = append(form.Parts, &web.FileField{
			Name:     filePath,
			Filename: strings.TrimPrefix(filePath, "/"),
			Reader:   strings.NewReader(content),
		})
	}

	resp := &BitbucketResponse[any]{
		Mime: web.MimeOmit,
	}

	req := prepare(c, ctx, &BitbucketRequest[web.MultipartForm]{
		Method: "POST",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "src"},
		Body:   form,
		Mime:   web.MimeMultipartFormData,
	})

	return Perform(req, resp)
}

// CreateBranch creates a new branch in the specified repository.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - body: Request configuration including branch name and target commit hash
//
// The branch name should not include any prefixes (e.g. refs/heads).
// The target hash can be a full commit hash or "default" to use the default branch tip.
// Using a full commit hash is the preferred approach.
//
// Returns the created branch object with details including links and merge strategies.
//
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-refs/#api-repositories-workspace-repo-slug-refs-branches-post
func (c *Client) CreateBranch(ctx context.Context, workspaceSlug string, repoSlug string, body *CreateBranchRequest) (*Branch, error) {
	resp := &BitbucketResponse[Branch]{
		Body: &Branch{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[CreateBranchRequest]{
		Method: "POST",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "refs", "branches"},
		Body:   body,
		Mime:   web.MimeApplicationJson,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetBranch retrieves information about a specific branch.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - branchName: The name of the branch to retrieve
//
// Returns detailed branch information including the target commit hash, merge strategies,
// and sync strategies.
func (c *Client) GetBranch(ctx context.Context, workspaceSlug string, repoSlug string, branchName string) (*Branch, error) {
	resp := &BitbucketResponse[Branch]{
		Body: &Branch{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "refs", "branches", branchName},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// CreatePullRequest creates a new pull request in the specified repository.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - body: Request configuration including title, source branch, and optional fields
//
// The minimum required fields are title and source branch name.
// If destination is not specified, it defaults to the repository's main branch.
// Optional fields include description, close_source_branch, draft, and reviewers.
//
// Returns the created pull request object with full details.
//
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-post
func (c *Client) CreatePullRequest(ctx context.Context, workspaceSlug string, repoSlug string, body *CreatePullRequestRequest) (*PullRequest, error) {
	resp := &BitbucketResponse[PullRequest]{
		Body: &PullRequest{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[CreatePullRequestRequest]{
		Method: "POST",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests"},
		Body:   body,
		Mime:   web.MimeApplicationJson,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// CreatePullRequestComment creates a new comment on a specific pull request.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//   - body: Request configuration including comment content and optional inline/parent references
//
// The minimum required field is content.raw with the comment text.
// Optional fields include inline (for code line comments) and parent (for reply comments).
//
// Returns the created comment object with full details.
//
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-post
func (c *Client) CreatePullRequestComment(ctx context.Context, workspaceSlug string, repoSlug string, pullRequestId int, body *CreatePullRequestCommentRequest) (*PullRequestComment, error) {
	resp := &BitbucketResponse[PullRequestComment]{
		Body: &PullRequestComment{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[CreatePullRequestCommentRequest]{
		Method: "POST",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "comments"},
		Body:   body,
		Mime:   web.MimeApplicationJson,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// MergePullRequest merges a pull request.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//   - body: Request configuration including type, message, close_source_branch, and merge_strategy
//
// The type field is required. Other fields are optional.
// Returns the updated pull request object with state changed to "MERGED".
//
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-merge-post
func (c *Client) MergePullRequest(ctx context.Context, workspaceSlug string, repoSlug string, pullRequestId int, body *MergePullRequestRequest) (*PullRequest, error) {
	resp := &BitbucketResponse[PullRequest]{
		Body: &PullRequest{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[MergePullRequestRequest]{
		Method: "POST",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "merge"},
		Body:   body,
		Mime:   web.MimeApplicationJson,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// DeclinePullRequest declines a pull request.
//
// Parameters:
//   - ctx: Context for the request
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//
// Returns the updated pull request object with state changed to "DECLINED".
//
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-decline-post
func (c *Client) DeclinePullRequest(ctx context.Context, workspaceSlug string, repoSlug string, pullRequestId int) (*PullRequest, error) {
	resp := &BitbucketResponse[PullRequest]{
		Body: &PullRequest{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, ctx, &BitbucketRequest[any]{
		Method: "POST",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "decline"},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// prepare populates a BitbucketRequest with client configuration and authentication.
// It sets the base URL, HTTP client, and determines which authentication method to use.
// BearerAuth takes precedence over BasicAuth if both are configured.
func prepare[T any](c *Client, ctx context.Context, req *BitbucketRequest[T]) *BitbucketRequest[T] {
	req.Context = ctx
	req.BaseUrl = c.cfg.Url
	req.Client = c.client
	req.Authorizer = c.authorizer
	return req
}
