package bitbucket

import (
	"context"
	"os"
	"path/filepath"

	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/branow/mcp-bitbucket/internal/util/web"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// GitConfig provides configuration for creating a Bitbucket git client.
type GitConfig struct {
	// BaseURL is the base URL for git clone operations (e.g., "https://bitbucket.org")
	BaseURL string
}

// GitClient is a Bitbucket git client that performs git operations such as cloning.
// Authentication is provided via a GitAuthorizer — the token is never embedded in
// URLs or exposed to the process list.
type GitClient struct {
	cfg        GitConfig
	authorizer util.GitAuthorizer
}

// NewGitClient creates a new GitClient with the given configuration and authorizer.
func NewGitClient(cfg GitConfig, authorizer util.GitAuthorizer) *GitClient {
	return &GitClient{cfg: cfg, authorizer: authorizer}
}

// Clone clones a Bitbucket repository to a local path.
//
// Parameters:
//   - ctx: The request context
//   - workspace: Bitbucket workspace slug
//   - repository: Repository slug
//   - targetPath: Local filesystem path to clone into (created if absent)
//   - depth: Shallow clone depth; 0 means a full clone
//   - ref: Branch, tag, or full reference to clone (e.g. "main", "refs/tags/v1.0");
//     defaults to the repository's default branch when empty
//
// Returns the resolved absolute local path of the cloned repository, or an error if
// the clone operation fails.
func (c *GitClient) Clone(
	ctx context.Context,
	workspace, repository, targetPath string,
	depth int,
	ref string,
) (string, error) {
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", util.NewInvalidParamsError("failed to resolve target path: " + err.Error())
	}

	cloneURL, err := (&web.UrlBuilder{
		BaseUrl: c.cfg.BaseURL,
		Path:    []string{workspace, repository},
	}).Build()
	if err != nil {
		return "", util.NewInvalidParamsError("failed to build clone URL: " + err.Error())
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return "", util.NewInvalidParamsError("failed to create target directory: " + err.Error())
	}

	username, token, err := c.authorizer.Authorize(ctx)
	if err != nil {
		return "", util.NewInvalidParamsError("failed to obtain git credentials: " + err.Error())
	}

	cloneOpts := &gogit.CloneOptions{
		URL: cloneURL,
		Auth: &githttp.BasicAuth{
			Username: username,
			Password: token,
		},
		Depth: depth,
	}

	if ref != "" {
		// Accept full reference names (refs/heads/main, refs/tags/v1.0) as-is;
		// treat short names as branch names (refs/heads/<ref>).
		r := plumbing.ReferenceName(ref)
		if r.IsBranch() || r.IsTag() {
			cloneOpts.ReferenceName = r
		} else {
			cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(ref)
		}
	}

	if _, err := gogit.PlainCloneContext(ctx, absPath, false, cloneOpts); err != nil {
		return "", util.NewInvalidParamsError("git clone failed: " + err.Error())
	}

	return absPath, nil
}
