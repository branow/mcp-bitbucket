//go:build integration

// # Running GitClient Integration Tests
//
// Integration tests connect to a real Bitbucket instance and require proper configuration.
// To run these tests, use the following command:
//
//	go test -tags=integration ./internal/bitbucket/test/integration
//
// # Required Environment Variables
//
//	TEST_BITBUCKET_GIT_URL       - Base URL for git clone operations (e.g., https://bitbucket.org)
//	TEST_BITBUCKET_URL           - Base URL of the Bitbucket API (e.g., https://api.bitbucket.org/2.0)
//	TEST_BITBUCKET_EMAIL         - Email address for authentication
//	TEST_BITBUCKET_API_TOKEN     - API token or app password for authentication
//	TEST_BITBUCKET_WORKSPACE     - Workspace slug where test repositories will be created
//	TEST_BITBUCKET_PROJECT_KEY   - Project key where test repositories will be created
//
// Example configuration (or use .env file):
//
//	export TEST_BITBUCKET_GIT_URL="https://bitbucket.org"
//	export TEST_BITBUCKET_URL="https://api.bitbucket.org/2.0"
//	export TEST_BITBUCKET_EMAIL="your-email@example.com"
//	export TEST_BITBUCKET_API_TOKEN="your-app-password"
//	export TEST_BITBUCKET_WORKSPACE="your-workspace"
//	export TEST_BITBUCKET_PROJECT_KEY="TEST"
package bitbucket_integration_test

import (
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	intg "github.com/branow/mcp-bitbucket/internal/bitbucket/test/integration"
	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/branow/mcp-bitbucket/internal/util"
	sch "github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/stretchr/testify/suite"
)

// GitClientIntegrationTestSuite is the test suite for GitClient integration tests.
type GitClientIntegrationTestSuite struct {
	suite.Suite
	api       *bitbucket.ApiClient
	git       *bitbucket.GitClient
	gitURL    string
	workspace string
	project   string
}

func TestGitClientIntegration(t *testing.T) {
	suite.Run(t, new(GitClientIntegrationTestSuite))
}

func (s *GitClientIntegrationTestSuite) SetupSuite() {
	s.gitURL = config.GetCrit("TEST_BITBUCKET_GIT_URL", sch.String().Must(sch.NotBlank(), sch.ValidURL()).Critical())
	s.workspace = config.GetCrit("TEST_BITBUCKET_WORKSPACE", sch.String().Must(sch.NotBlank()).Critical())
	s.project = config.GetCrit("TEST_BITBUCKET_PROJECT_KEY", sch.String().Must(sch.NotBlank()).Critical())

	apiURL := config.GetOpt("TEST_BITBUCKET_URL", sch.String().Must(sch.NotBlank()).Optional("https://api.bitbucket.org/2.0"))
	timeout := config.GetOpt("TEST_BITBUCKET_TIMEOUT", sch.Int().Must(sch.Positive()).Optional(5))
	username := config.GetCrit("TEST_BITBUCKET_EMAIL", sch.String().Must(sch.NotBlank()).Critical())
	password := config.GetCrit("TEST_BITBUCKET_API_TOKEN", sch.String().Must(sch.NotBlank()).Critical())

	s.api = bitbucket.NewApiClient(
		bitbucket.ApiConfig{Url: apiURL, Timeout: timeout},
		util.NewBasicAuthorizer(username, password),
	)
	s.git = bitbucket.NewGitClient(
		bitbucket.GitConfig{BaseURL: s.gitURL},
		util.NewStaticGitAuthorizer(password),
	)
}

// TestClone verifies that a repository can be cloned to a local directory.
func (s *GitClientIntegrationTestSuite) TestClone() {
	t := s.T()
	t.Parallel()

	repoSlug, _, _ := intg.StepCreateRepository(t, s.api, s.workspace, s.project, "git-clone", true)
	t.Cleanup(func() { intg.StepDeleteRepository(t, s.api, s.workspace, repoSlug) })

	intg.StepCreateFiles(t, s.api, s.workspace, repoSlug, map[string]string{
		"README.md": "# Test\n",
	}, "initial commit", "", "")
	intg.StepCreateFiles(t, s.api, s.workspace, repoSlug, map[string]string{
		"README.md": "# Updated\n",
	}, "second commit", "", "")

	targetDir := intg.StepCreateCloneDir(t)
	intg.StepCloneRepository(t, s.git, s.workspace, repoSlug, targetDir, 0, "")
	intg.StepVerifyCommitCount(t, targetDir, 2)
	intg.StepVerifyFileExists(t, targetDir, "README.md")
}

// TestClone_ShallowClone verifies that a shallow clone with depth=1 exposes exactly one commit.
func (s *GitClientIntegrationTestSuite) TestClone_ShallowClone() {
	t := s.T()
	t.Parallel()

	repoSlug, _, _ := intg.StepCreateRepository(t, s.api, s.workspace, s.project, "git-shallow", true)
	t.Cleanup(func() { intg.StepDeleteRepository(t, s.api, s.workspace, repoSlug) })

	intg.StepCreateFiles(t, s.api, s.workspace, repoSlug, map[string]string{
		"README.md": "# Test\n",
	}, "initial commit", "", "")
	intg.StepCreateFiles(t, s.api, s.workspace, repoSlug, map[string]string{
		"README.md": "# Updated\n",
	}, "second commit", "", "")

	targetDir := intg.StepCreateCloneDir(t)
	intg.StepCloneRepository(t, s.git, s.workspace, repoSlug, targetDir, 1, "")
	intg.StepVerifyCommitCount(t, targetDir, 1)
	intg.StepVerifyFileExists(t, targetDir, "README.md")
}

// TestClone_WithRef verifies that cloning with a specific branch ref checks out the correct branch.
// A feature branch is created with a unique file; after cloning, that file must be present.
func (s *GitClientIntegrationTestSuite) TestClone_WithRef() {
	t := s.T()
	t.Parallel()

	repoSlug, _, _ := intg.StepCreateRepository(t, s.api, s.workspace, s.project, "git-ref", true)
	t.Cleanup(func() { intg.StepDeleteRepository(t, s.api, s.workspace, repoSlug) })

	intg.StepCreateFiles(t, s.api, s.workspace, repoSlug, map[string]string{
		"README.md": "# Test\n",
	}, "initial commit", "", "")

	commitHash := intg.StepGetLatestCommitHash(t, s.api, s.workspace, repoSlug)
	featureBranch := "feature/test-ref-clone"
	intg.StepCreateBranch(t, s.api, s.workspace, repoSlug, featureBranch, commitHash)
	intg.StepCreateFiles(t, s.api, s.workspace, repoSlug, map[string]string{
		"FEATURE.md": "# Feature\n",
	}, "add feature file", featureBranch, "")

	targetDir := intg.StepCreateCloneDir(t)
	intg.StepCloneRepository(t, s.git, s.workspace, repoSlug, targetDir, 0, featureBranch)
	intg.StepVerifyFileExists(t, targetDir, "README.md")
	intg.StepVerifyFileExists(t, targetDir, "FEATURE.md")
}

// TestClone_InvalidCredentials verifies that cloning with a bad token returns an error.
func (s *GitClientIntegrationTestSuite) TestClone_InvalidCredentials() {
	t := s.T()
	t.Parallel()

	repoSlug, _, _ := intg.StepCreateRepository(t, s.api, s.workspace, s.project, "git-bad-creds", true)
	t.Cleanup(func() { intg.StepDeleteRepository(t, s.api, s.workspace, repoSlug) })

	intg.StepCreateFiles(t, s.api, s.workspace, repoSlug, map[string]string{
		"README.md": "# Test\n",
	}, "initial commit", "", "")

	badClient := bitbucket.NewGitClient(
		bitbucket.GitConfig{BaseURL: s.gitURL},
		util.NewStaticGitAuthorizer("invalid-token"),
	)

	targetDir := intg.StepCreateCloneDir(t)
	intg.StepCloneFails(t, badClient, s.workspace, repoSlug, targetDir, 0, "")
}
