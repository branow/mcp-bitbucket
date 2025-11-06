package bitbucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/branow/mcp-bitbucket/internal/util"
)

var (
	ErrEmptyNamespace  = errors.New("namespace is empty")
	ErrInternal        = errors.New("failed to make request to bitbucket")
	ErrServerBitbucket = errors.New("bitbucket service is currently unavailable")
	ErrClientBitbucket = errors.New("bitbucket failed to process request")
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

func (c *Client) ListRepositories(namespace string, pagelen int, page int) (*BitBucketResponse[Repository], error) {
	result := &BitBucketResponse[Repository]{}

	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return result, ErrEmptyNamespace
	}

	query := map[string]string{
		"pagelen": strconv.Itoa(pagelen),
		"page":    strconv.Itoa(page),
	}

	resp, err := c.request("GET", []string{"repositories", namespace}, query)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(resp, result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) request(method string, path []string, query map[string]string) ([]byte, error) {
	url := util.UrlBuilder{
		BaseUrl:     c.baseUrl,
		Path:        path,
		QueryParams: query,
	}
	req, err := util.CreateRequest(method, url, nil)
	if err != nil {
		return []byte{}, ErrInternal
	}
	req.SetBasicAuth(c.username, c.password)
	resp, err := util.DoRequest(c.client, req)
	if err != nil {
		return []byte{}, ErrInternal
	}

	switch {
	case resp.StatusCode >= 500:
		return []byte{}, ErrServerBitbucket
	case resp.StatusCode >= 400:
		errResp := &BitBucketErrorResponse{}
		err := util.ReadResponseJson(resp, errResp)
		if err != nil {
			return []byte{}, ErrClientBitbucket
		}
		if errResp.Error.Message != "" {
			return []byte{}, fmt.Errorf("%w: %s", ErrClientBitbucket, errResp.Error.Message)
		}
		return []byte{}, ErrClientBitbucket
	default:
		result, err := util.ReadResponseBytes(resp)
		if err != nil {
			return []byte{}, ErrInternal
		}
		return result, nil
	}

}
