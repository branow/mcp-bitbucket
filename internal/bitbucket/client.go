package bitbucket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/branow/mcp-bitbucket/internal/util"
)

type Config struct {
	Username  string
	Password  string
	BaseUrl   string
	Namespace string
	Timeout   int
}

type Client struct {
	username  string
	password  string
	baseUrl   string
	namespace string
	client    *http.Client
}

func NewClient(config Config) *Client {
	return &Client{
		username:  config.Username,
		password:  config.Password,
		baseUrl:   config.BaseUrl,
		namespace: config.Namespace,
		client:    &http.Client{Timeout: time.Duration(config.Timeout) * time.Second},
	}
}

func (c *Client) ListRepositories() (string, error) {
	resp, err := c.request("GET", util.JoinUrlPath("repositories", c.namespace))
	if err != nil {
		return "", fmt.Errorf("failed to list repositories: %w", err)
	}
	return resp, nil
}

func (c *Client) request(method string, path string) (string, error) {
	url := c.buildUrl(path)
	req, err := util.CreateRequest(method, url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.username, c.password)
	resp, err := util.DoRequest(c.client, req)
	if err != nil {
		return "", err
	}
	body, err := util.ReadResponse(resp)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *Client) buildUrl(segments ...string) string {
	return fmt.Sprintf("%s%s", c.baseUrl, util.JoinUrlPath(segments...))
}
