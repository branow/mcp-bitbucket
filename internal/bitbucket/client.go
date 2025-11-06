package bitbucket

import (
	"net/http"
	"strconv"
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

func (c *Client) ListRepositories(namespaceSlug string, pagelen int, page int) (*BitbucketApiResponse[Repository], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[Repository]]{
		Body: &BitbucketApiResponse[Repository]{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", namespaceSlug},
			Query: map[string]string{
				"pagelen": strconv.Itoa(pagelen),
				"page":    strconv.Itoa(page),
			},
		}),
		resp,
	)

	return resp.Body, err
}

func (c *Client) GetRepository(namespaceSlug string, repoSlug string) (*Repository, error) {
	resp := &BitbucketResponse[Repository]{
		Body: &Repository{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", namespaceSlug, repoSlug},
		}),
		resp,
	)

	return resp.Body, err
}

func (c *Client) GetRepositorySource(namespaceSlug string, repoSlug string) (*BitbucketApiResponse[SourceItem], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[SourceItem]]{
		Body: &BitbucketApiResponse[SourceItem]{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", namespaceSlug, repoSlug, "src"},
		}),
		resp,
	)

	return resp.Body, err
}

func (c *Client) ListPullRequests(namespaceSlug string, repoSlug string, pagelen int, page int) (*BitbucketApiResponse[PullRequest], error) {
	resp := &BitbucketResponse[BitbucketApiResponse[PullRequest]]{
		Body: &BitbucketApiResponse[PullRequest]{},
	}

	err := Perform(
		c.prepare(&BitbucketRequest{
			Method: "GET",
			Path:   []string{"repositories", namespaceSlug, repoSlug, "pullrequests"},
			Query: map[string]string{
				"pagelen": strconv.Itoa(pagelen),
				"page":    strconv.Itoa(page),
			},
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
