package bitbucket

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/branow/mcp-bitbucket/internal/util"
)

var (
	ErrEmptyNamespace = errors.New("namespace is empty")
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

func (c *Client) ListRepositories(namespace string) (string, error) {
	return util.WrapErrorFunc("list repositories", func() (string, error) {
		namespace = strings.TrimSpace(namespace)
		if namespace == "" {
			return "", ErrEmptyNamespace
		}

		resp, err := c.request("GET", util.JoinUrlPath("repositories", namespace))
		if err != nil {
			return "", err
		}
		return resp, nil
	})
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
