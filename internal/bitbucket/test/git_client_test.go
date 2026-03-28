package bitbucket_test

import (
	"context"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitClient_Clone(t *testing.T) {
	t.Parallel()
	const token = "valid-token"
	baseURL := newGitTestServer(t, tokenAuth(token), "ws/repo")
	client := bitbucket.NewGitClient(bitbucket.GitConfig{BaseURL: baseURL}, util.NewStaticGitAuthorizer(token))

	absPath, err := client.Clone(context.Background(), "ws", "repo", t.TempDir(), 0, "")

	require.NoError(t, err)
	assert.NotEmpty(t, absPath)
	_, statErr := os.Stat(filepath.Join(absPath, ".git"))
	assert.NoError(t, statErr, ".git directory should exist after clone")
}

func TestGitClient_Clone_ShallowClone(t *testing.T) {
	t.Parallel()
	const token = "valid-token"
	baseURL := newGitTestServer(t, tokenAuth(token), "ws/repo")
	client := bitbucket.NewGitClient(bitbucket.GitConfig{BaseURL: baseURL}, util.NewStaticGitAuthorizer(token))

	absPath, err := client.Clone(context.Background(), "ws", "repo", t.TempDir(), 1, "")

	require.NoError(t, err)
	assert.NotEmpty(t, absPath)
	_, statErr := os.Stat(filepath.Join(absPath, ".git"))
	assert.NoError(t, statErr, ".git directory should exist after shallow clone")
}

func TestGitClient_Clone_WithRef(t *testing.T) {
	t.Parallel()
	const token = "valid-token"
	baseURL := newGitTestServer(t, tokenAuth(token), "ws/repo")
	client := bitbucket.NewGitClient(bitbucket.GitConfig{BaseURL: baseURL}, util.NewStaticGitAuthorizer(token))

	absPath, err := client.Clone(context.Background(), "ws", "repo", t.TempDir(), 0, "main")

	require.NoError(t, err)
	assert.NotEmpty(t, absPath)
}

func TestGitClient_Clone_WithFullRef(t *testing.T) {
	t.Parallel()
	const token = "valid-token"
	baseURL := newGitTestServer(t, tokenAuth(token), "ws/repo")
	client := bitbucket.NewGitClient(bitbucket.GitConfig{BaseURL: baseURL}, util.NewStaticGitAuthorizer(token))

	absPath, err := client.Clone(context.Background(), "ws", "repo", t.TempDir(), 0, "refs/heads/main")

	require.NoError(t, err)
	assert.NotEmpty(t, absPath)
}

func TestGitClient_Clone_InvalidToken(t *testing.T) {
	t.Parallel()
	const validToken = "valid-token"
	baseURL := newGitTestServer(t, tokenAuth(validToken), "ws/repo")
	client := bitbucket.NewGitClient(bitbucket.GitConfig{BaseURL: baseURL}, util.NewStaticGitAuthorizer("wrong-token"))

	_, err := client.Clone(context.Background(), "ws", "repo", t.TempDir(), 0, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")
}

func TestGitClient_Clone_InvalidURL(t *testing.T) {
	t.Parallel()

	client := bitbucket.NewGitClient(bitbucket.GitConfig{BaseURL: "http://127.0.0.1:1"}, util.NewStaticGitAuthorizer("token"))

	_, err := client.Clone(context.Background(), "ws", "repo", t.TempDir(), 0, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")
}

func TestGitClient_Clone_AuthorizerError(t *testing.T) {
	t.Parallel()
	const token = "valid-token"
	baseURL := newGitTestServer(t, tokenAuth(token), "ws/repo")
	// StaticGitAuthorizer with empty token returns an error from Authorize.
	client := bitbucket.NewGitClient(bitbucket.GitConfig{BaseURL: baseURL}, util.NewStaticGitAuthorizer(""))

	_, err := client.Clone(context.Background(), "ws", "repo", t.TempDir(), 0, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to obtain git credentials")
}

// newGitTestServer creates a local git HTTP server backed by git http-backend.
// repos must be in "workspace/repository" form; each is pre-populated with an
// initial commit on the main branch. The auth middleware is applied to all requests.
// Returns the server base URL.
func newGitTestServer(t *testing.T, auth func(http.Handler) http.Handler, repos ...string) string {
	t.Helper()

	baseDir := t.TempDir()
	for _, repo := range repos {
		initBareRepo(t, baseDir, repo)
	}

	gitExec, err := exec.LookPath("git")
	require.NoError(t, err, "git must be installed")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := &cgi.Handler{
			Path:       gitExec,
			Args:       []string{"http-backend"},
			Env:        []string{"GIT_PROJECT_ROOT=" + baseDir, "GIT_HTTP_EXPORT_ALL=1"},
			InheritEnv: []string{"HOME", "PATH", "USER", "TMPDIR", "GIT_EXEC_PATH"},
		}
		h.ServeHTTP(w, r)
	})

	server := httptest.NewServer(auth(handler))
	t.Cleanup(server.Close)
	return server.URL
}

// initBareRepo creates a bare git repository at baseDir/repo pre-populated with
// an initial empty commit on the main branch. repo must be in "workspace/repository" form.
func initBareRepo(t *testing.T, baseDir, repo string) {
	t.Helper()

	bareDir := filepath.Join(baseDir, filepath.FromSlash(repo))
	require.NoError(t, os.MkdirAll(bareDir, 0755))

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v failed: %s", args, out)
	}

	run(bareDir, "init", "--bare")
	workDir := t.TempDir()
	run(workDir, "init", "-b", "main")
	run(workDir, "config", "user.email", "test@example.com")
	run(workDir, "config", "user.name", "Test")
	run(workDir, "commit", "--allow-empty", "-m", "initial commit")
	run(workDir, "remote", "add", "origin", bareDir)
	run(workDir, "push", "origin", "main")
}

// tokenAuth returns a middleware that requires Basic auth with the given token as password.
// Any other credentials receive HTTP 401.
func tokenAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, pass, ok := r.BasicAuth()
			if !ok || pass != token {
				w.Header().Set("WWW-Authenticate", `Basic realm="git"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
