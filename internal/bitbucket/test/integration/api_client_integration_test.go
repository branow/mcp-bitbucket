//go:build integration

// # Running Integration Tests
//
// Integration tests connect to a real Bitbucket instance and require proper configuration.
// To run these tests, use the following command:
//
//	go test -tags=integration ./internal/bitbucket/test/integration
//
// # Required Environment Variables
//
// ## Basic Authentication Tests
//
// The following environment variables must be set to run basic auth tests:
//
//	TEST_BITBUCKET_URL               - Base URL of the Bitbucket instance (e.g., https://api.bitbucket.org/2.0)
//	TEST_BITBUCKET_EMAIL             - Email address for authentication
//	TEST_BITBUCKET_API_TOKEN         - API token or app password for authentication
//	TEST_BITBUCKET_WORKSPACE         - Workspace slug where test repositories will be created
//	TEST_BITBUCKET_PROJECT_KEY       - Project key where test repositories will be created
//	TEST_BITBUCKET_TIMEOUT           - Optional: Request timeout in seconds (default: 5)
//
// Example configuration for basic auth (or use .env file):
//
//	export TEST_BITBUCKET_URL="https://api.bitbucket.org/2.0"
//	export TEST_BITBUCKET_EMAIL="your-email@example.com"
//	export TEST_BITBUCKET_API_TOKEN="your-app-password"
//	export TEST_BITBUCKET_WORKSPACE="your-workspace"
//	export TEST_BITBUCKET_PROJECT_KEY="TEST"
//
// ## OAuth Tests
//
// To run OAuth integration tests, configure these additional variables:
//
//	TEST_BITBUCKET_CLIENT_ID         - OAuth client ID (from Bitbucket OAuth consumer)
//	TEST_BITBUCKET_CLIENT_SECRET     - OAuth client secret
//
// Example configuration for OAuth (in addition to basic auth vars):
//
//	export TEST_BITBUCKET_CLIENT_ID="your-oauth-client-id"
//	export TEST_BITBUCKET_CLIENT_SECRET="your-oauth-client-secret"
//
// Note: OAuth tests use the client credentials flow to obtain an access token automatically.
//
// # Test Cleanup
//
// All integration tests are designed to clean up after themselves:
//   - Test repositories are automatically deleted after each test using t.Cleanup()
//   - Temporary branches, pull requests, and other resources are created in test repositories
//   - Failed tests may leave artifacts; check your workspace if tests are interrupted
package bitbucket_integration_test

import (
	"context"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	intg "github.com/branow/mcp-bitbucket/internal/bitbucket/test/integration"
	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/branow/mcp-bitbucket/internal/util"
	sch "github.com/branow/mcp-bitbucket/internal/util/schema"
	"github.com/stretchr/testify/suite"
)

// ApiClientIntegrationTestSuite_BasicAuth is the test suite for Bitbucket client integration tests
type ApiClientIntegrationTestSuite_BasicAuth struct {
	suite.Suite
	bb        *bitbucket.ApiClient
	workspace string
	project   string
}

func TestApiClientIntegration_BasicAuth(t *testing.T) {
	suite.Run(t, new(ApiClientIntegrationTestSuite_BasicAuth))
}

func (s *ApiClientIntegrationTestSuite_BasicAuth) SetupSuite() {
	s.workspace = config.GetCrit("TEST_BITBUCKET_WORKSPACE", sch.String().Must(sch.NotBlank()).Critical())
	s.project = config.GetCrit("TEST_BITBUCKET_PROJECT_KEY", sch.String().Must(sch.NotBlank()).Critical())

	cfg := bitbucket.ApiConfig{
		Url:     config.GetOpt("TEST_BITBUCKET_URL", sch.String().Must(sch.NotBlank()).Optional("https://api.bitbucket.org/2.0")),
		Timeout: config.GetOpt("TEST_BITBUCKET_TIMEOUT", sch.Int().Must(sch.Positive()).Optional(5)),
	}

	username := config.GetCrit("TEST_BITBUCKET_EMAIL", sch.String().Must(sch.NotBlank()).Critical())
	password := config.GetCrit("TEST_BITBUCKET_API_TOKEN", sch.String().Must(sch.NotBlank()).Critical())
	authorizer := util.NewBasicAuthorizer(username, password)

	s.bb = bitbucket.NewApiClient(cfg, authorizer)
}

// TestRepositoryLifecycle verifies repository creation, retrieval, listing, and deletion.
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestRepositoryLifecycle() {
	t := s.T()
	t.Parallel()

	repoSlug, createdRepo, mainBranch := intg.StepCreateRepository(t, s.bb, s.workspace, s.project, "repo-lifecycle", true)
	s.NotEmpty(repoSlug, "Repository slug should not be empty")
	s.NotNil(createdRepo, "Repository creation failed")
	s.NotEmpty(mainBranch, "Main branch should not be empty")

	t.Cleanup(func() {
		intg.StepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	intg.StepVerifyGetRepository(t, s.bb, s.workspace, repoSlug, createdRepo)
	intg.StepVerifyListRepositories(t, s.bb, s.workspace, repoSlug, createdRepo)
}

// TestRepositorySourceTree verifies repository source tree structure with nested directories.
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestRepositorySourceTree() {
	t := s.T()
	t.Parallel()

	repoSlug, _, mainBranch := intg.StepCreateRepository(t, s.bb, s.workspace, s.project, "source-tree", true)
	s.NotEmpty(repoSlug, "Repository slug should not be empty")
	s.NotEmpty(mainBranch, "Main branch should not be empty")

	t.Cleanup(func() {
		intg.StepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"src/main/file.go":  "package main\n\nfunc main() {}\n",
		"docs/README.md":    "# Documentation\n\nThis is a test file.\n",
		"src/utils/util.go": "package utils\n\nfunc Helper() {}\n",
		"README.md":         "# Test Repository\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit with nested structure", "", "")

	intg.StepVerifySourceTree(t, s.bb, s.workspace, repoSlug, []string{"src", "docs", "README.md"})
	intg.StepVerifyDirectorySource(t, s.bb, s.workspace, repoSlug, mainBranch, "src", []string{"src/main", "src/utils"})
	intg.StepVerifyFileContent(t, s.bb, s.workspace, repoSlug, mainBranch, "src/main/file.go", []string{"package main", "func main()"})
}

// TestPullRequestBasics verifies pull request creation and retrieval.
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestPullRequestBasics() {
	t := s.T()
	t.Parallel()

	featureBranch := "feature-test-branch"

	repoSlug, _, mainBranch := intg.StepCreateRepository(t, s.bb, s.workspace, s.project, "pr-basics", true)
	t.Cleanup(func() {
		intg.StepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	commitHash := intg.StepGetLatestCommitHash(t, s.bb, s.workspace, repoSlug)
	s.NotEmpty(commitHash, "Commit hash should not be empty")

	branch := intg.StepCreateBranch(t, s.bb, s.workspace, repoSlug, featureBranch, commitHash)
	s.NotNil(branch, "Branch creation failed")

	updatedFiles := map[string]string{
		"README.md": "# Test Repository\n\nUpdated content from feature branch.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, updatedFiles, "Update README on feature branch", featureBranch, "")

	pr := intg.StepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Test Pull Request", "This is a test pull request", featureBranch, mainBranch)
	s.NotNil(pr, "Pull request creation failed")

	intg.StepVerifyGetPullRequest(t, s.bb, s.workspace, repoSlug, pr.ID, "Test Pull Request", featureBranch)
	intg.StepVerifyListPullRequests(t, s.bb, s.workspace, repoSlug, pr.ID, "Test Pull Request")
}

// TestPullRequestCommitsAndDiffAndComments verifies PR commits, diff, and comments with pagination.
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestPullRequestCommitsAndDiffAndComments() {
	t := s.T()
	t.Parallel()

	featureBranch := "feature-multiple-commits"

	repoSlug, _, mainBranch := intg.StepCreateRepository(t, s.bb, s.workspace, s.project, "pr-commits-diff-comments", true)
	t.Cleanup(func() {
		intg.StepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
		"file1.txt": "First file content.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	commitHash := intg.StepGetLatestCommitHash(t, s.bb, s.workspace, repoSlug)
	s.NotEmpty(commitHash, "Commit hash should not be empty")

	branch := intg.StepCreateBranch(t, s.bb, s.workspace, repoSlug, featureBranch, commitHash)
	s.NotNil(branch, "Branch creation failed")

	commit1Files := map[string]string{
		"README.md": "# Test Repository\n\nUpdated by commit 1.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, commit1Files, "Commit 1: Update README", featureBranch, "")

	commit2Files := map[string]string{
		"file1.txt": "First file updated in commit 2.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, commit2Files, "Commit 2: Update file1.txt", featureBranch, "")

	commit3Files := map[string]string{
		"file2.txt": "New file added in commit 3.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, commit3Files, "Commit 3: Add file2.txt", featureBranch, "")

	pr := intg.StepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Test PR with Multiple Commits", "This PR has 3 commits", featureBranch, mainBranch)
	s.NotNil(pr, "Pull request creation failed")

	intg.StepVerifyListPullRequestCommits(t, s.bb, s.workspace, repoSlug, pr.ID, []string{
		"Commit 1: Update README",
		"Commit 2: Update file1.txt",
		"Commit 3: Add file2.txt",
	})

	intg.StepVerifyGetPullRequestDiff(t, s.bb, s.workspace, repoSlug, pr.ID, []string{
		"README.md",
		"file1.txt",
		"file2.txt",
		"Updated by commit 1",
		"First file updated in commit 2",
		"New file added in commit 3",
	})

	comment1 := intg.StepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "First comment on this PR")
	s.NotNil(comment1, "First comment creation failed")

	comment2 := intg.StepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "Second comment for testing")
	s.NotNil(comment2, "Second comment creation failed")

	comment3 := intg.StepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "Third comment here")
	s.NotNil(comment3, "Third comment creation failed")

	comment4 := intg.StepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "Fourth comment to test pagination")
	s.NotNil(comment4, "Fourth comment creation failed")

	comment5 := intg.StepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "Fifth and final comment")
	s.NotNil(comment5, "Fifth comment creation failed")

	intg.StepVerifyListPullRequestComments(t, s.bb, s.workspace, repoSlug, pr.ID, 3, 1, 3, []string{
		"First comment on this PR",
		"Second comment for testing",
		"Third comment here",
	})

	intg.StepVerifyListPullRequestComments(t, s.bb, s.workspace, repoSlug, pr.ID, 3, 2, 2, []string{
		"Fourth comment to test pagination",
		"Fifth and final comment",
	})
}

// TestPullRequestStates verifies filtering pull requests by state (OPEN, MERGED, DECLINED).
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestPullRequestStates() {
	t := s.T()
	t.Parallel()

	repoSlug, _, mainBranch := intg.StepCreateRepository(t, s.bb, s.workspace, s.project, "pr-states", true)
	t.Cleanup(func() {
		intg.StepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	commitHash := intg.StepGetLatestCommitHash(t, s.bb, s.workspace, repoSlug)
	s.NotEmpty(commitHash, "Commit hash should not be empty")

	branch1 := "feature-open"
	intg.StepCreateBranch(t, s.bb, s.workspace, repoSlug, branch1, commitHash)
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, map[string]string{
		"file1.txt": "Content for open PR\n",
	}, "Add file1 for open PR", branch1, "")
	pr1 := intg.StepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Open Pull Request", "This PR will stay open", branch1, mainBranch)
	s.NotNil(pr1, "PR 1 creation failed")

	branch2 := "feature-merged"
	intg.StepCreateBranch(t, s.bb, s.workspace, repoSlug, branch2, commitHash)
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, map[string]string{
		"file2.txt": "Content for merged PR\n",
	}, "Add file2 for merged PR", branch2, "")
	pr2 := intg.StepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Merged Pull Request", "This PR will be merged", branch2, mainBranch)
	s.NotNil(pr2, "PR 2 creation failed")
	intg.StepMergePullRequest(t, s.bb, s.workspace, repoSlug, pr2.ID)

	branch3 := "feature-declined"
	intg.StepCreateBranch(t, s.bb, s.workspace, repoSlug, branch3, commitHash)
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, map[string]string{
		"file3.txt": "Content for declined PR\n",
	}, "Add file3 for declined PR", branch3, "")
	pr3 := intg.StepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Declined Pull Request", "This PR will be declined", branch3, mainBranch)
	s.NotNil(pr3, "PR 3 creation failed")
	intg.StepDeclinePullRequest(t, s.bb, s.workspace, repoSlug, pr3.ID)

	intg.StepVerifyListPullRequestsByState(t, s.bb, s.workspace, repoSlug, []string{"OPEN"}, []int{pr1.ID}, []int{pr2.ID, pr3.ID})
	intg.StepVerifyListPullRequestsByState(t, s.bb, s.workspace, repoSlug, []string{"MERGED"}, []int{pr2.ID}, []int{pr1.ID, pr3.ID})
	intg.StepVerifyListPullRequestsByState(t, s.bb, s.workspace, repoSlug, []string{"DECLINED"}, []int{pr3.ID}, []int{pr1.ID, pr2.ID})
}

// TestErrorNonExistentWorkspace verifies error handling for non-existent workspace.
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestErrorNonExistentWorkspace() {
	t := s.T()
	t.Parallel()

	nonExistentWorkspace := "non-existent-workspace-12345"

	t.Run("list repositories with non-existent workspace", func(t *testing.T) {
		_, err := s.bb.ListRepositories(context.Background(), nonExistentWorkspace, 10, 1)
		s.Error(err, "Should return error for non-existent workspace")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})

	t.Run("get repository with non-existent workspace", func(t *testing.T) {
		_, err := s.bb.GetRepository(context.Background(), nonExistentWorkspace, "any-repo")
		s.Error(err, "Should return error for non-existent workspace")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})
}

// TestErrorNonExistentRepository verifies error handling for non-existent repository.
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestErrorNonExistentRepository() {
	t := s.T()
	t.Parallel()

	nonExistentRepo := "non-existent-repo-12345"

	t.Run("get repository with non-existent repo", func(t *testing.T) {
		_, err := s.bb.GetRepository(context.Background(), s.workspace, nonExistentRepo)
		s.Error(err, "Should return error for non-existent repository")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})

	t.Run("get repository source with non-existent repo", func(t *testing.T) {
		_, err := s.bb.GetRepositorySource(context.Background(), s.workspace, nonExistentRepo)
		s.Error(err, "Should return error for non-existent repository")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})

	t.Run("list pull requests with non-existent repo", func(t *testing.T) {
		_, err := s.bb.ListPullRequests(context.Background(), s.workspace, nonExistentRepo, 10, 1, []string{"OPEN"})
		s.Error(err, "Should return error for non-existent repository")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})
}

// TestErrorNonExistentPullRequest verifies error handling for non-existent pull request.
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestErrorNonExistentPullRequest() {
	t := s.T()
	t.Parallel()

	repoSlug, _, _ := intg.StepCreateRepository(t, s.bb, s.workspace, s.project, "error-nonexistent-pr", true)
	t.Cleanup(func() {
		intg.StepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	nonExistentPRID := 99999

	t.Run("get pull request with non-existent PR ID", func(t *testing.T) {
		_, err := s.bb.GetPullRequest(context.Background(), s.workspace, repoSlug, nonExistentPRID)
		s.Error(err, "Should return error for non-existent pull request")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})

	t.Run("list pull request commits with non-existent PR ID", func(t *testing.T) {
		_, err := s.bb.ListPullRequestCommits(context.Background(), s.workspace, repoSlug, nonExistentPRID)
		s.Error(err, "Should return error for non-existent pull request")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})

	t.Run("list pull request comments with non-existent PR ID", func(t *testing.T) {
		_, err := s.bb.ListPullRequestComments(context.Background(), s.workspace, repoSlug, nonExistentPRID, 10, 1)
		s.Error(err, "Should return error for non-existent pull request")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})

	t.Run("get pull request diff with non-existent PR ID", func(t *testing.T) {
		_, err := s.bb.GetPullRequestDiff(context.Background(), s.workspace, repoSlug, nonExistentPRID)
		s.Error(err, "Should return error for non-existent pull request")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})
}

// TestErrorNonExistentFileDirectory verifies error handling for non-existent files and directories.
func (s *ApiClientIntegrationTestSuite_BasicAuth) TestErrorNonExistentFileDirectory() {
	t := s.T()
	t.Parallel()

	repoSlug, _, mainBranch := intg.StepCreateRepository(t, s.bb, s.workspace, s.project, "error-nonexistent-file", true)
	t.Cleanup(func() {
		intg.StepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	t.Run("get file source with non-existent file path", func(t *testing.T) {
		_, err := s.bb.GetFileSource(context.Background(), s.workspace, repoSlug, mainBranch, "non-existent-file.txt")
		s.Error(err, "Should return error for non-existent file")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})

	t.Run("get directory source with non-existent directory path", func(t *testing.T) {
		_, err := s.bb.GetDirectorySource(context.Background(), s.workspace, repoSlug, mainBranch, "non-existent-dir")
		s.Error(err, "Should return error for non-existent directory")
		util.AssertJsonRpcError(t, err, util.CodeResourceNotFoundErr, "Should be a ResourceNotFound error (404)")
	})
}

// ApiClientIntegrationTestSuite_OAuth is the test suite for ApiClient integration tests using OAuth
type ApiClientIntegrationTestSuite_OAuth struct {
	suite.Suite
	bb        *bitbucket.ApiClient
	workspace string
	project   string
	token     string
}

func TestApiClientIntegration_OAuth(t *testing.T) {
	suite.Run(t, new(ApiClientIntegrationTestSuite_OAuth))
}

func (s *ApiClientIntegrationTestSuite_OAuth) SetupSuite() {
	clientId := config.GetCrit("TEST_BITBUCKET_CLIENT_ID", sch.String().Must(sch.NotBlank()).Critical())
	clientSecret := config.GetCrit("TEST_BITBUCKET_CLIENT_SECRET", sch.String().Must(sch.NotBlank()).Critical())
	tokenUrl := config.GetOpt("TEST_BITBUCKET_ACCESS_TOKEN_URL", sch.String().Must(sch.NotBlank()).Optional("https://bitbucket.org/site/oauth2/access_token"))

	token, err := util.ObtainAccessToken(clientId, clientSecret, tokenUrl)
	s.Require().NoError(err)
	s.token = token

	s.workspace = config.GetCrit("TEST_BITBUCKET_WORKSPACE", sch.String().Must(sch.NotBlank()).Critical())
	s.project = config.GetCrit("TEST_BITBUCKET_PROJECT_KEY", sch.String().Must(sch.NotBlank()).Critical())

	cfg := bitbucket.ApiConfig{
		Url:     config.GetOpt("TEST_BITBUCKET_URL", sch.String().Must(sch.NotBlank()).Optional("https://api.bitbucket.org/2.0")),
		Timeout: config.GetOpt("TEST_BITBUCKET_TIMEOUT", sch.Int().Must(sch.Positive()).Optional(5)),
	}

	authorizer := util.NewOAuthAuthorizer(util.NewStaticTokenExtractor(token))
	s.bb = bitbucket.NewApiClient(cfg, authorizer)
}

// TestOAuthBasicOperations verifies that OAuth authentication works for basic Bitbucket operations
func (s *ApiClientIntegrationTestSuite_OAuth) TestOAuthBasicOperations() {
	t := s.T()
	t.Parallel()

	repoSlug, createdRepo, mainBranch := intg.StepCreateRepository(t, s.bb, s.workspace, s.project, "repo-lifecycle", true)
	s.NotEmpty(repoSlug, "Repository slug should not be empty")
	s.NotNil(createdRepo, "Repository creation failed")
	s.NotEmpty(mainBranch, "Main branch should not be empty")

	t.Cleanup(func() {
		intg.StepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	intg.StepVerifyGetRepository(t, s.bb, s.workspace, repoSlug, createdRepo)
	intg.StepVerifyListRepositories(t, s.bb, s.workspace, repoSlug, createdRepo)

	featureBranch := "feature-test-branch"

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	commitHash := intg.StepGetLatestCommitHash(t, s.bb, s.workspace, repoSlug)
	s.NotEmpty(commitHash, "Commit hash should not be empty")

	branch := intg.StepCreateBranch(t, s.bb, s.workspace, repoSlug, featureBranch, commitHash)
	s.NotNil(branch, "Branch creation failed")

	updatedFiles := map[string]string{
		"README.md": "# Test Repository\n\nUpdated content from feature branch.\n",
	}
	intg.StepCreateFiles(t, s.bb, s.workspace, repoSlug, updatedFiles, "Update README on feature branch", featureBranch, "")

	pr := intg.StepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Test Pull Request", "This is a test pull request", featureBranch, mainBranch)
	s.NotNil(pr, "Pull request creation failed")

	intg.StepVerifyGetPullRequest(t, s.bb, s.workspace, repoSlug, pr.ID, "Test Pull Request", featureBranch)
	intg.StepVerifyListPullRequests(t, s.bb, s.workspace, repoSlug, pr.ID, "Test Pull Request")
}
