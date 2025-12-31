package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestBody struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestPerform_Success(t *testing.T) {
	t.Parallel()

	t.Run("with JSON response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/test", r.URL.Path)
			assert.Equal(t, "1", r.URL.Query().Get("page"))

			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "user", username)
			assert.Equal(t, "pass", password)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"result","value":99}`))
		}))
		defer server.Close()

		authorizer := util.NewBasicAuthorizer("user", "pass")

		req := &client.BitbucketRequest[TestBody]{
			Method:     "POST",
			BaseUrl:    server.URL,
			Path:       []string{"api", "test"},
			Query:      map[string]string{"page": "1"},
			Body:       &TestBody{Name: "test", Value: 42},
			Mime:       web.MimeApplicationJson,
			Authorizer: authorizer,
			Context:    context.Background(),
			Client:     server.Client(),
		}

		resp := &client.BitbucketResponse[TestBody]{
			Body: &TestBody{},
			Mime: web.MimeApplicationJson,
		}

		err := client.Perform(req, resp)
		require.NoError(t, err)
		assert.Equal(t, "result", resp.Body.Name)
		assert.Equal(t, 99, resp.Body.Value)
	})

	t.Run("without response body", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		authorizer := util.NewBasicAuthorizer("user", "pass")

		req := &client.BitbucketRequest[TestBody]{
			Method:     "DELETE",
			BaseUrl:    server.URL,
			Path:       []string{"api", "test"},
			Mime:       web.MimeOmit,
			Authorizer: authorizer,
			Context:    context.Background(),
			Client:     server.Client(),
		}

		resp := &client.BitbucketResponse[TestBody]{
			Mime: web.MimeOmit,
		}

		err := client.Perform(req, resp)
		require.NoError(t, err)
	})
}

func TestPerform_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	authorizer := util.NewBasicAuthorizer("user", "pass")

	req := &client.BitbucketRequest[TestBody]{
		Method:     "GET",
		BaseUrl:    server.URL,
		Path:       []string{"api"},
		Mime:       web.MimeOmit,
		Authorizer: authorizer,
		Context:    context.Background(),
		Client:     server.Client(),
	}

	resp := &client.BitbucketResponse[TestBody]{
		Body: &TestBody{},
		Mime: web.MimeApplicationJson,
	}

	err := client.Perform(req, resp)
	require.Error(t, err)
	util.AssertJsonRpcError(t, err, util.CodeResourceUnavailableErr)
}

func TestPerform_ClientError(t *testing.T) {
	t.Parallel()

	t.Run("with error message", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"type":"error","error":{"message":"Invalid parameter"}}`))
		}))
		defer server.Close()

		authorizer := util.NewBasicAuthorizer("user", "pass")

		req := &client.BitbucketRequest[TestBody]{
			Method:     "POST",
			BaseUrl:    server.URL,
			Path:       []string{"api"},
			Mime:       web.MimeOmit,
			Authorizer: authorizer,
			Context:    context.Background(),
			Client:     server.Client(),
		}

		resp := &client.BitbucketResponse[TestBody]{
			Body: &TestBody{},
			Mime: web.MimeApplicationJson,
		}

		err := client.Perform(req, resp)
		require.Error(t, err)
		util.AssertJsonRpcError(t, err, util.CodeInvalidParamsErr)
		assert.Contains(t, err.Error(), "Invalid parameter")
	})

	t.Run("without error message", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		authorizer := util.NewBasicAuthorizer("user", "pass")

		req := &client.BitbucketRequest[TestBody]{
			Method:     "GET",
			BaseUrl:    server.URL,
			Path:       []string{"api"},
			Mime:       web.MimeOmit,
			Authorizer: authorizer,
			Context:    context.Background(),
			Client:     server.Client(),
		}

		resp := &client.BitbucketResponse[TestBody]{
			Body: &TestBody{},
			Mime: web.MimeApplicationJson,
		}

		err := client.Perform(req, resp)
		require.Error(t, err)
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestPerform_RequestBuildError(t *testing.T) {
	t.Parallel()

	authorizer := util.NewBasicAuthorizer("user", "pass")

	req := &client.BitbucketRequest[TestBody]{
		Method:     "GET",
		BaseUrl:    "://invalid",
		Path:       []string{"api"},
		Mime:       web.MimeOmit,
		Authorizer: authorizer,
		Context:    context.Background(),
		Client:     &http.Client{},
	}

	resp := &client.BitbucketResponse[TestBody]{
		Body: &TestBody{},
		Mime: web.MimeApplicationJson,
	}

	err := client.Perform(req, resp)
	require.Error(t, err)
	util.AssertJsonRpcError(t, err, util.CodeInternalErr)
}

func TestPerform_NetworkError(t *testing.T) {
	t.Parallel()

	authorizer := util.NewBasicAuthorizer("user", "pass")

	req := &client.BitbucketRequest[TestBody]{
		Method:     "GET",
		BaseUrl:    "http://invalid-domain-that-does-not-exist.local",
		Path:       []string{"api"},
		Mime:       web.MimeOmit,
		Authorizer: authorizer,
		Context:    context.Background(),
		Client:     &http.Client{Timeout: time.Duration(100) * time.Millisecond},
	}

	resp := &client.BitbucketResponse[TestBody]{
		Body: &TestBody{},
		Mime: web.MimeApplicationJson,
	}

	err := client.Perform(req, resp)
	require.Error(t, err)
	util.AssertJsonRpcError(t, err, util.CodeInternalErr)
}

func TestPerform_InvalidResponseBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	authorizer := util.NewBasicAuthorizer("user", "pass")

	req := &client.BitbucketRequest[TestBody]{
		Method:     "GET",
		BaseUrl:    server.URL,
		Path:       []string{"api"},
		Mime:       web.MimeOmit,
		Authorizer: authorizer,
		Context:    context.Background(),
		Client:     server.Client(),
	}

	resp := &client.BitbucketResponse[TestBody]{
		Body: &TestBody{},
		Mime: web.MimeApplicationJson,
	}

	err := client.Perform(req, resp)
	require.Error(t, err)
	util.AssertJsonRpcError(t, err, util.CodeInternalErr)
}
