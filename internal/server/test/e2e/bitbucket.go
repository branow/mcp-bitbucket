package e2e

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func NewBitbucketServer(t *testing.T, auth Middleware) *httptest.Server {
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
