package bitbucket

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Username string
	Password string
	BaseUrl  string
	Timeout  int
}

type Client struct {
	username string
	password string
	baseUrl  string
	client   *http.Client
}

func NewClient(config Config) *Client {
	return &Client{
		username: config.Username,
		password: config.Password,
		baseUrl:  config.BaseUrl,
		client:   &http.Client{Timeout: time.Duration(config.Timeout) * time.Second},
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

	return resp.Body, err
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

	return resp.Body, err
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

	return resp.Body, err
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

	return resp.Body, err
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

	return resp.Body, err
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

	return resp.Body, err
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

	return resp.Body, err
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

	return resp.Body, err
}

func (c *Client) prepare(req *BitbucketRequest) *BitbucketRequest {
	req.BaseUrl = c.baseUrl
	req.Username = c.username
	req.Password = c.password
	req.Client = c.client
	return req
}
