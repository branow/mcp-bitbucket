// Package client provides a Bitbucket API client for interacting with repositories,
// pull requests, and source code.
package client

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/branow/mcp-bitbucket/internal/util/web"
)

// Config provides configuration parameters for the Bitbucket client.
type Config interface {
	// BitbucketUrl returns the base URL of the Bitbucket API (e.g., "https://api.bitbucket.org/2.0").
	BitbucketUrl() string
	// BitbucketEmail returns the email address for authentication.
	BitbucketEmail() string
	// BitbucketApiToken returns the API token for authentication.
	BitbucketApiToken() string
	// BitbucketTimeout returns the HTTP request timeout in seconds.
	BitbucketTimeout() int
}

// Client is a Bitbucket API client that provides methods for accessing
// repositories, pull requests, and source code.
type Client struct {
	username string
	password string
	baseUrl  string
	client   *http.Client
}

// NewClient creates a new Bitbucket API client with the provided configuration.
// The client uses HTTP basic authentication with the email and API token from the config.
// The timeout specified in the config is applied to all HTTP requests.
func NewClient(config Config) *Client {
	return &Client{
		username: config.BitbucketEmail(),
		password: config.BitbucketApiToken(),
		baseUrl:  config.BitbucketUrl(),
		client:   &http.Client{Timeout: time.Duration(config.BitbucketTimeout()) * time.Second},
	}
}

// ListRepositories retrieves a paginated list of repositories for the specified workspace.
//
// Parameters:
//   - workspaceSlug: The workspace slug identifier
//   - pagelen: Number of items per page
//   - page: Page number to retrieve (1-indexed)
//
// Returns the API response containing the list of repositories and pagination metadata.
func (c *Client) ListRepositories(workspaceSlug string, pagelen int, page int) (*ApiResponse[Repository], error) {
	resp := &BitbucketResponse[ApiResponse[Repository]]{
		Body: &ApiResponse[Repository]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//
// Returns the repository details including metadata, links, and configuration.
func (c *Client) GetRepository(workspaceSlug string, repoSlug string) (*Repository, error) {
	resp := &BitbucketResponse[Repository]{
		Body: &Repository{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//
// Returns the API response containing the list of files and directories at the repository root.
func (c *Client) GetRepositorySource(workspaceSlug string, repoSlug string) (*ApiResponse[SourceItem], error) {
	resp := &BitbucketResponse[ApiResponse[SourceItem]]{
		Body: &ApiResponse[SourceItem]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pagelen: Number of items per page
//   - page: Page number to retrieve (1-indexed)
//   - states: Filter by pull request states (e.g., "OPEN", "MERGED", "DECLINED"). Empty slice returns all states.
//
// Returns the API response containing the list of pull requests and pagination metadata.
func (c *Client) ListPullRequests(workspaceSlug string, repoSlug string, pagelen int, page int, states []string) (*ApiResponse[PullRequest], error) {
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

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//
// Returns the pull request details including title, description, state, author, reviewers, and metadata.
func (c *Client) GetPullRequest(workspaceSlug string, repoSlug string, pullRequestId int) (*PullRequest, error) {
	resp := &BitbucketResponse[PullRequest]{
		Body: &PullRequest{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//
// Returns the API response containing the list of commits with their hash, message, author, and metadata.
func (c *Client) ListPullRequestCommits(workspaceSlug string, repoSlug string, pullRequestId int) (*ApiResponse[Commit], error) {
	resp := &BitbucketResponse[ApiResponse[Commit]]{
		Body: &ApiResponse[Commit]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//   - pagelen: Number of items per page
//   - page: Page number to retrieve (1-indexed)
//
// Returns the API response containing the list of comments with their content, author, and inline code references.
func (c *Client) ListPullRequestComments(workspaceSlug string, repoSlug string, pullRequestId int, pagelen int, page int) (*ApiResponse[PullRequestComment], error) {
	resp := &BitbucketResponse[ApiResponse[PullRequestComment]]{
		Body: &ApiResponse[PullRequestComment]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - pullRequestId: The pull request ID number
//
// Returns the diff content as a plain text string in unified diff format.
func (c *Client) GetPullRequestDiff(workspaceSlug string, repoSlug string, pullRequestId int) (*string, error) {
	resp := &BitbucketResponse[string]{
		Body: new(string),
		Mime: web.MimeTextPlain,
	}

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - commit: The commit hash or branch name
//   - path: The file path relative to the repository root
//
// Returns the file content as a plain text string.
func (c *Client) GetFileSource(workspaceSlug string, repoSlug string, commit string, path string) (*string, error) {
	resp := &BitbucketResponse[string]{
		Body: new(string),
		Mime: web.MimeTextPlain,
	}

	req := prepare(c, &BitbucketRequest[any]{
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
//   - workspaceSlug: The workspace slug identifier
//   - repoSlug: The repository slug identifier
//   - commit: The commit hash or branch name
//   - path: The directory path relative to the repository root
//
// Returns the API response containing the list of files and subdirectories.
func (c *Client) GetDirectorySource(workspaceSlug string, repoSlug string, commit string, path string) (*ApiResponse[SourceItem], error) {
	resp := &BitbucketResponse[ApiResponse[SourceItem]]{
		Body: &ApiResponse[SourceItem]{},
		Mime: web.MimeApplicationJson,
	}

	req := prepare(c, &BitbucketRequest[any]{
		Method: "GET",
		Path:   []string{"repositories", workspaceSlug, repoSlug, "src", commit, path},
		Mime:   web.MimeOmit,
	})

	if err := Perform(req, resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func prepare[T any](c *Client, req *BitbucketRequest[T]) *BitbucketRequest[T] {
	req.BaseUrl = c.baseUrl
	req.Username = c.username
	req.Password = c.password
	req.Client = c.client
	return req
}
