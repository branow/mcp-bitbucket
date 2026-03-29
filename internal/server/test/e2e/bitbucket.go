package e2e

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stretchr/testify/require"
)

func NewBitbucketApiServer(t *testing.T, auth Middleware) *httptest.Server {
	mux := http.NewServeMux()
	newBitbucketRepositoriesHandler(t, mux)
	newBitbucketRepositoriesNotFoundHandler(t, mux)
	newBitbucketRepositoryHandler(t, mux)
	newBitbucketRepositoryWithoutReadmeHandler(t, mux)
	newBitbucketRepositoryNotFoundHandler(t, mux)
	newBitbucketRepositorySourceHandler(t, mux)
	newBitbucketRepositorySourceWithoutReadmeHandler(t, mux)
	newBitbucketRepositorySourceNotFoundHandler(t, mux)
	newBitbucketFileSourceReadmeHandler(t, mux)
	newBitbucketFileSourceServiceHandler(t, mux)
	newBitbucketFileSourceNestedReadmeHandler(t, mux)
	newBitbucketFileSourceNotFoundHandler(t, mux)
	newBitbucketDirectorySourceHandler(t, mux)
	newBitbucketDirectorySourceNotFoundHandler(t, mux)
	newBitbucketPullRequestsHandler(t, mux)
	newBitbucketPullRequestsNotFoundHandler(t, mux)
	newBitbucketPullRequestHandler(t, mux)
	newBitbucketPullRequestNotFoundHandler(t, mux)
	newBitbucketPullRequestCommitsHandler(t, mux)
	newBitbucketPullRequestCommitsNotFoundHandler(t, mux)
	newBitbucketPullRequestDiffHandler(t, mux)
	newBitbucketPullRequestDiffNotFoundHandler(t, mux)
	newBitbucketPullRequestCommentsHandler(t, mux)
	newBitbucketPullRequestCommentsNotFoundHandler(t, mux)
	return httptest.NewServer(auth(mux))
}

type Middleware func(http.Handler) http.Handler

func NewBasicAuthMiddleware(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			const prefix = "Basic "
			if !strings.HasPrefix(auth, prefix) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			payload, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(string(payload), ":", 2)
			if len(parts) != 2 || parts[0] != username || parts[1] != password {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewGitTokenMiddleware returns a Middleware that validates Basic auth where
// the password must equal token. This matches go-git's auth format where the
// git token is sent as the Basic auth password.
func NewGitTokenMiddleware(token string) Middleware {
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

func NewOpaqueTokenMiddleware(token string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			const prefix = "Bearer "

			if !strings.HasPrefix(auth, prefix) || auth[len(prefix):] != token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func newBitbucketRepositoriesHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "repositories.json"))
	})
}

func newBitbucketRepositoriesNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/invalid-workspace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "not-found.json"))
	})
}

func newBitbucketRepositoryHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "repository.json"))
	})
}

func newBitbucketRepositoryWithoutReadmeHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository-without-readme", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "repository.json"))
	})
}

func newBitbucketRepositoryNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/invalid-repository", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "not-found.json"))
	})
}

func newBitbucketRepositorySourceHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/src", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "repository-source.json"))
	})
}

func newBitbucketRepositorySourceWithoutReadmeHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository-without-readme/src", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "repository-source-without-readme.json"))
	})
}

func newBitbucketRepositorySourceNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/invalid-repository/src", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "not-found.json"))
	})
}

func newBitbucketFileSourceReadmeHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/abc123def456/README.md", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "file-source-readme.md"))
	})
}

func newBitbucketFileSourceServiceHandler(t *testing.T, mux *http.ServeMux) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "file-source-service.java"))
	}
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/feature-branch/src/main/java/com/example/service/UserService.java", handler)
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/main/src/main/java/com/example/service/UserService.java", handler)
}

func newBitbucketFileSourceNestedReadmeHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/abc123def456/docs/api/v2/endpoints/users/README.md", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "file-source-readme.md"))
	})
}

func newBitbucketFileSourceNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/abc123def456/src/main/java/com/example/nonexistent/Missing.java", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "pull-request-not-found.txt"))
	})
}

func newBitbucketDirectorySourceHandler(t *testing.T, mux *http.ServeMux) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(ReadBitbucketFile(t, "directory-source.json"))
	}
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/feature-branch/src/main/java/com/example/service", handler)
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/main/src/main/java/com/example/service", handler)
}

func newBitbucketDirectorySourceNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/src/abc123def456/src/main/java/com/example/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "directory-not-found.json"))
	})
}

func newBitbucketPullRequestsHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		state := r.URL.Query().Get("state")
		switch state {
		case "MERGED":
			w.Write(ReadBitbucketFile(t, "pull-requests-merged.json"))
		default:
			w.Write(ReadBitbucketFile(t, "pull-requests.json"))
		}
	})
}

func newBitbucketPullRequestsNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/invalid-repository/pullrequests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "not-found.json"))
	})
}

func newBitbucketPullRequestHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "pull-request.json"))
	})
}

func newBitbucketPullRequestNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/999", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "pull-request-not-found.txt"))
	})
}

func newBitbucketPullRequestCommitsHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/1/commits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "pull-request-commits.json"))
	})
}

func newBitbucketPullRequestCommitsNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/999/commits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "pull-request-not-found.txt"))
	})
}

func newBitbucketPullRequestDiffHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/1/diff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "pull-request-diff.txt"))
	})
}

func newBitbucketPullRequestDiffNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/999/diff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "pull-request-not-found.txt"))
	})
}

func newBitbucketPullRequestCommentsHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ReadBitbucketFile(t, "pull-request-comments.json"))
	})
}

func newBitbucketPullRequestCommentsNotFoundHandler(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("/repositories/test-workspace/test-repository/pullrequests/999/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(ReadBitbucketFile(t, "pull-request-not-found.txt"))
	})
}

// NewBitbucketGitServer creates an HTTP git server that serves the given repositories
// using git http-backend over CGI. Each repo must be in the form "workspace/repository".
// Each bare repo is pre-populated with an initial commit so shallow clones work.
// The provided auth middleware is applied to all requests.
// Returns an httptest.Server; use Server.URL as the BITBUCKET_GIT_URL value.
func NewBitbucketGitServer(t *testing.T, auth Middleware, repos ...string) *httptest.Server {
	t.Helper()
	server, _ := NewBitbucketGitServerWithBaseDir(t, auth, repos...)
	return server
}

// NewBitbucketGitServerWithBaseDir is like NewBitbucketGitServer but also returns
// the base directory so callers can push additional commits after setup.
func NewBitbucketGitServerWithBaseDir(t *testing.T, auth Middleware, repos ...string) (*httptest.Server, string) {
	t.Helper()

	baseDir, err := os.MkdirTemp("", "mcp-bitbucket-git-*")
	require.NoError(t, err, "failed to create git base dir")
	t.Cleanup(func() { os.RemoveAll(baseDir) })

	for _, repo := range repos {
		initBareRepo(t, baseDir, repo)
	}

	gitExec, err := exec.LookPath("git")
	require.NoError(t, err, "git executable not found in PATH")

	gitHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := &cgi.Handler{
			Path:       gitExec,
			Args:       []string{"http-backend"},
			Env:        []string{"GIT_PROJECT_ROOT=" + baseDir, "GIT_HTTP_EXPORT_ALL=1"},
			InheritEnv: []string{"HOME", "PATH", "USER", "TMPDIR", "GIT_EXEC_PATH"},
		}
		h.ServeHTTP(w, r)
	})

	server := httptest.NewServer(auth(gitHandler))
	t.Cleanup(server.Close)
	return server, baseDir
}

// NewGitRepos creates bare git repositories under a temp directory and returns
// the file:// base URL to use as BITBUCKET_GIT_URL.
//
// Each repo in repos must be in the form "workspace/repository".
// Each bare repo is pre-populated with an initial commit so shallow clones work.
func NewGitRepos(t *testing.T, repos ...string) string {
	t.Helper()
	baseURL, _ := NewGitReposWithBaseDir(t, repos...)
	return baseURL
}

// NewGitReposWithBaseDir is like NewGitRepos but also returns the base directory
// so callers can push additional commits after setup.
func NewGitReposWithBaseDir(t *testing.T, repos ...string) (gitBaseURL, baseDir string) {
	t.Helper()

	baseDir, err := os.MkdirTemp("", "mcp-bitbucket-git-*")
	require.NoError(t, err, "failed to create git base dir")
	t.Cleanup(func() { os.RemoveAll(baseDir) })

	for _, repo := range repos {
		initBareRepo(t, baseDir, repo)
	}

	return "file://" + filepath.ToSlash(baseDir), baseDir
}

// AddCommitToBareRepo pushes a new file commit directly into the bare repository
// at baseDir/repo. repo must be in the form "workspace/repository".
func AddCommitToBareRepo(t *testing.T, baseDir, repo, filename, content string) {
	t.Helper()

	bareDir := filepath.Join(baseDir, filepath.FromSlash(repo))
	workDir, err := os.MkdirTemp("", "mcp-bitbucket-work-*")
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v failed: %s", args, out)
	}

	run(workDir, "clone", bareDir, ".")
	run(workDir, "config", "user.email", "test@example.com")
	run(workDir, "config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, filename), []byte(content), 0644))
	run(workDir, "add", filename)
	run(workDir, "commit", "-m", "add "+filename)
	run(workDir, "push", "origin", "main")
}

// initBareRepo creates a bare git repository at baseDir/repo and populates it
// with an initial empty commit on the main branch.
// repo must be in the form "workspace/repository".
func initBareRepo(t *testing.T, baseDir, repo string) {
	t.Helper()

	bareDir := filepath.Join(baseDir, filepath.FromSlash(repo))
	require.NoError(t, os.MkdirAll(bareDir, 0755), "failed to create bare repo dir")

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v failed: %s", args, out)
	}

	run(bareDir, "init", "--bare")

	workDir, err := os.MkdirTemp("", "mcp-bitbucket-work-*")
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	run(workDir, "init", "-b", "main")
	run(workDir, "config", "user.email", "test@example.com")
	run(workDir, "config", "user.name", "Test")
	run(workDir, "commit", "--allow-empty", "-m", "initial commit")
	run(workDir, "remote", "add", "origin", bareDir)
	run(workDir, "push", "origin", "main")
}
