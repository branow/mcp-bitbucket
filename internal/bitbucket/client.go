package bitbucket

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Config interface {
	BitbucketUrl() string
	BitbucketEmail() string
	BitbucketApiToken() string
  BitbucketTimeout() int
}

type Client struct {
	username string
	password string
	baseUrl  string
	client   *http.Client
}

func NewClient(config Config) *Client {
	return &Client{
		username: config.BitbucketEmail(),
		password: config.BitbucketApiToken(),
		baseUrl:  config.BitbucketUrl(),
		client:   &http.Client{Timeout: time.Duration(config.BitbucketTimeout()) * time.Second},
	}
}

func (c *Client) ListRepositories(workspaceSlug string, pagelen int, page int) (*BitbucketApiResponse[BitbucketRepository], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[BitbucketRepository]]{
		Body: &BitbucketApiResponse[BitbucketRepository]{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug},
			Query: map[string]string{
				"pagelen": strconv.Itoa(pagelen),
				"page":    strconv.Itoa(page),
			},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) GetRepository(workspaceSlug string, repoSlug string) (*BitbucketRepository, error) {
	resp := &BitbucketResponse[BitbucketRepository]{
		Body: &BitbucketRepository{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) GetRepositorySource(workspaceSlug string, repoSlug string) (*BitbucketApiResponse[BitbucketSourceItem], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[BitbucketSourceItem]]{
		Body: &BitbucketApiResponse[BitbucketSourceItem]{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug, "src"},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) ListPullRequests(workspaceSlug string, repoSlug string, pagelen int, page int, states []string) (*BitbucketApiResponse[BitbucketPullRequest], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[BitbucketPullRequest]]{
		Body: &BitbucketApiResponse[BitbucketPullRequest]{},
	}

	query := map[string]string{
		"pagelen": strconv.Itoa(pagelen),
		"page":    strconv.Itoa(page),
	}

	if len(states) > 0 {
		query["state"] = strings.Join(states, ",")
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests"},
			Query:  query,
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) GetPullRequest(workspaceSlug string, repoSlug string, pullRequestId int) (*BitbucketPullRequest, error) {
	resp := &BitbucketResponse[BitbucketPullRequest]{
		Body: &BitbucketPullRequest{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId)},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) ListPullRequestCommits(workspaceSlug string, repoSlug string, pullRequestId int) (*BitbucketApiResponse[BitbucketCommit], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[BitbucketCommit]]{
		Body: &BitbucketApiResponse[BitbucketCommit]{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "commits"},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) ListPullRequestComments(workspaceSlug string, repoSlug string, pullRequestId int, pagelen int, page int) (*BitbucketApiResponse[BitbucketPullRequestComment], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[BitbucketPullRequestComment]]{
		Body: &BitbucketApiResponse[BitbucketPullRequestComment]{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "comments"},
			Query: map[string]string{
				"pagelen": strconv.Itoa(pagelen),
				"page":    strconv.Itoa(page),
			},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) GetPullRequestDiff(workspaceSlug string, repoSlug string, pullRequestId int) (*string, error) {
	resp := &BitbucketResponse[string]{
		Body: new(string),
	}

	err := PerformText(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug, "pullrequests", strconv.Itoa(pullRequestId), "diff"},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) GetFileSource(workspaceSlug string, repoSlug string, commit string, path string) (*string, error) {
	resp := &BitbucketResponse[string]{
		Body: new(string),
	}

	err := PerformText(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug, "src", commit, path},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) GetDirectorySource(workspaceSlug string, repoSlug string, commit string, path string) (*BitbucketApiResponse[BitbucketSourceItem], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[BitbucketSourceItem]]{
		Body: &BitbucketApiResponse[BitbucketSourceItem]{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", workspaceSlug, repoSlug, "src", commit, path},
		}),
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) prepare(req *BitbucketRequest) *BitbucketRequest {
	req.BaseUrl = c.baseUrl
	req.Username = c.username
	req.Password = c.password
	req.Client = c.client
	return req
}
