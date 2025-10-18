package bitbucket

import (
	"github.com/branow/mcp-bitbucket/internal/util"
)

type Client struct {
	authentification string
}

func NewClient(email string, apitoken string) *Client {
	return &Client{
		authentification: util.BasicAuth(email, apitoken),
	}
}
