package bitbucket_test

import (
	"context"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gitTestToken = "valid-token"

type gitClientCloneTestCase struct {
	name        string
	depth       int
	ref         string
	clientToken string // empty triggers authorizer error
	baseURL     string // override base URL; empty = use test server
	errContains string
}

func TestGitClient_Clone(t *testing.T) {
	t.Parallel()

	tests := []gitClientCloneTestCase{
		{
			name:        "success",
			clientToken: gitTestToken,
		},
		{
			name:        "shallow clone",
			depth:       1,
			clientToken: gitTestToken,
		},
		{
			name:        "with ref",
			ref:         "main",
			clientToken: gitTestToken,
		},
		{
			name:        "with full ref",
			ref:         "refs/heads/main",
			clientToken: gitTestToken,
		},
		{
			name:        "invalid token",
			clientToken: "wrong-token",
			errContains: "git clone failed",
		},
		{
			name:        "invalid URL",
			baseURL:     "http://127.0.0.1:1",
			clientToken: "token",
			errContains: "git clone failed",
		},
		{
			name:        "authorizer error",
			clientToken: "",
			errContains: "failed to obtain git credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			baseURL := tt.baseURL
			if baseURL == "" {
				baseURL = newGitTestServer(t, tokenAuth(gitTestToken), "ws/repo")
			}

			client := bitbucket.NewGitClient(
				bitbucket.GitConfig{BaseURL: baseURL},
				util.NewStaticGitAuthorizer(tt.clientToken),
			)
			absPath, err := client.Clone(context.Background(), "ws", "repo", t.TempDir(), tt.depth, tt.ref)

			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, absPath)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, absPath)
				_, statErr := os.Stat(filepath.Join(absPath, ".git"))
				assert.NoError(t, statErr, ".git directory should exist after clone")
			}
		})
	}
}

func TestIsRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
	}{
		{
			name: "valid git repository",
			setup: func(t *testing.T) string {
				baseDir := t.TempDir()
				initBareRepo(t, baseDir, "ws/repo")
				return filepath.Join(baseDir, "ws", "repo")
			},
			expected: true,
		},
		{
			name: "empty directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			expected: false,
		},
		{
			name: "non-existent path",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := tt.setup(t)
			assert.Equal(t, tt.expected, bitbucket.IsRepository(path))
		})
	}
}

type gitClientPullTestCase struct {
	name        string
	preClone    bool   // clone the repo into repoDir before calling Pull
	addCommit   bool   // push an extra commit to the bare repo
	repoPath    string // server-relative remote path; defaults to "/ws/repo"
	ref         string
	clientToken string // empty triggers authorizer error
	errContains string
}

func TestGitClient_Pull(t *testing.T) {
	t.Parallel()

	tests := []gitClientPullTestCase{
		{
			name:        "pulls new commit",
			preClone:    true,
			addCommit:   true,
			clientToken: gitTestToken,
		},
		{
			name:        "already up to date",
			preClone:    true,
			clientToken: gitTestToken,
		},
		{
			name:        "pull with ref",
			preClone:    true,
			addCommit:   true,
			ref:         "main",
			clientToken: gitTestToken,
		},
		{
			name:        "invalid token",
			preClone:    true,
			addCommit:   true,
			clientToken: "wrong-token",
			errContains: "git pull failed",
		},
		{
			name:        "authorizer error",
			preClone:    true,
			clientToken: "",
			errContains: "failed to obtain git credentials",
		},
		{
			name:        "not a git repository",
			clientToken: gitTestToken,
			errContains: "failed to open repository",
		},
		{
			name:        "remote URL mismatch",
			preClone:    true,
			repoPath:    "/other-ws/repo",
			clientToken: gitTestToken,
			errContains: "does not match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			baseDir := t.TempDir()
			initBareRepo(t, baseDir, "ws/repo")
			serverURL := newGitTestServerAt(t, tokenAuth(gitTestToken), baseDir)

			repoDir := t.TempDir()
			if tt.preClone {
				cloneRepoWithToken(t, serverURL+"/ws/repo", gitTestToken, repoDir)
			}
			if tt.addCommit {
				addCommitToBareRepo(t, baseDir, "ws/repo")
			}

			repoPath := tt.repoPath
			if repoPath == "" {
				repoPath = "/ws/repo"
			}

			client := bitbucket.NewGitClient(
				bitbucket.GitConfig{BaseURL: serverURL},
				util.NewStaticGitAuthorizer(tt.clientToken),
			)
			absPath, err := client.Pull(context.Background(), repoDir, serverURL+repoPath, tt.ref)

			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, absPath)
			} else {
				require.NoError(t, err)
				assert.Equal(t, repoDir, absPath)
			}
		})
	}
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
	return newGitTestServerAt(t, auth, baseDir)
}

// newGitTestServerAt creates a git HTTP server rooted at an existing baseDir.
func newGitTestServerAt(t *testing.T, auth func(http.Handler) http.Handler, baseDir string) string {
	t.Helper()

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

	runGit(t, bareDir, "init", "--bare")
	workDir := t.TempDir()
	runGit(t, workDir, "init", "-b", "main")
	runGit(t, workDir, "config", "user.email", "test@example.com")
	runGit(t, workDir, "config", "user.name", "Test")
	runGit(t, workDir, "commit", "--allow-empty", "-m", "initial commit")
	runGit(t, workDir, "remote", "add", "origin", bareDir)
	runGit(t, workDir, "push", "origin", "main")
}

// addCommitToBareRepo pushes a new file commit directly into the bare repository at baseDir/repo.
func addCommitToBareRepo(t *testing.T, baseDir, repo string) {
	t.Helper()

	bareDir := filepath.Join(baseDir, filepath.FromSlash(repo))
	workDir := t.TempDir()

	runGit(t, workDir, "clone", bareDir, ".")
	runGit(t, workDir, "config", "user.email", "test@example.com")
	runGit(t, workDir, "config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "update.md"), []byte("update"), 0644))
	runGit(t, workDir, "add", "update.md")
	runGit(t, workDir, "commit", "-m", "add update.md")
	runGit(t, workDir, "push", "origin", "main")
}

// cloneRepoWithToken clones repoURL into targetDir embedding token credentials in the URL.
func cloneRepoWithToken(t *testing.T, repoURL, token, targetDir string) {
	t.Helper()

	u, err := url.Parse(repoURL)
	require.NoError(t, err)
	u.User = url.UserPassword("x-token-auth", token)

	out, err := exec.Command("git", "clone", u.String(), targetDir).CombinedOutput()
	require.NoErrorf(t, err, "setup clone failed: %s", out)
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

// runGit executes a git command in dir and fails the test if it errors.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %v failed: %s", args, out)
}
