package bitbucket

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/branow/mcp-bitbucket/internal/util"
)

var (
	ErrInternal        = errors.New("failed to make request to bitbucket")
	ErrServerBitbucket = errors.New("bitbucket service is currently unavailable")
	ErrClientBitbucket = errors.New("bitbucket failed to process request")
)

type BitbucketRequest struct {
	Method   string
	BaseUrl  string
	Path     []string
	Query    map[string]string
	Username string
	Password string
	Client   *http.Client
}

type BitbucketResponse[T any] struct {
	Body *T
}

func Perform[T any](bbReq *BitbucketRequest, bbRes *BitbucketResponse[T]) error {
	url := util.UrlBuilder{
		BaseUrl:     bbReq.BaseUrl,
		Path:        bbReq.Path,
		QueryParams: bbReq.Query,
	}
	req, err := util.CreateRequest(bbReq.Method, url, nil)
	if err != nil {
		return ErrInternal
	}
	req.SetBasicAuth(bbReq.Username, bbReq.Password)
	resp, err := util.DoRequest(bbReq.Client, req)
	if err != nil {
		return ErrInternal
	}

	switch {
	case resp.StatusCode >= 500:
		return ErrServerBitbucket
	case resp.StatusCode >= 400:
		errResp := &BitBucketErrorResponse{}
		err := util.ReadResponseJson(resp, errResp)
		if err != nil {
			return ErrClientBitbucket
		}
		if errResp.Error.Message != "" {
			return fmt.Errorf("%w: %s", ErrClientBitbucket, errResp.Error.Message)
		}
		return ErrClientBitbucket
	default:
		return util.ReadResponseJson(resp, bbRes.Body)
	}
}
