//go:build integration

// Package client_test provides integration tests for the Bitbucket API client.
//
// # Running Integration Tests
//
// Integration tests connect to a real Bitbucket instance and require proper configuration.
// To run these tests, use the following command:
//
//	go test -tags=integration ./internal/bitbucket/client
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
package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/branow/mcp-bitbucket/internal/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite_BasicAuth is the test suite for Bitbucket client integration tests
type IntegrationTestSuite_BasicAuth struct {
	suite.Suite
	bb        *client.Client
	workspace string
	project   string
}

func TestIntegration_BasicAuth(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite_BasicAuth))
}

func (s *IntegrationTestSuite_BasicAuth) SetupSuite() {
	s.workspace = config.GetString("TEST_BITBUCKET_WORKSPACE", "")
	s.project = config.GetString("TEST_BITBUCKET_PROJECT_KEY", "")

	cfg := client.BitbucketConfig{
		Url:     config.GetString("TEST_BITBUCKET_URL", "https://api.bitbucket.org/2.0"),
		Timeout: config.GetInt("TEST_BITBUCKET_TIMEOUT", 5),
	}

	username := config.GetString("TEST_BITBUCKET_EMAIL", "")
	password := config.GetString("TEST_BITBUCKET_API_TOKEN", "")
	authorizer := util.NewBasicAuthorizer(username, password)

	s.bb = client.NewClient(cfg, authorizer)
}

// TestRepositoryLifecycle verifies repository creation, retrieval, listing, and deletion.
func (s *IntegrationTestSuite_BasicAuth) TestRepositoryLifecycle() {
	t := s.T()
	t.Parallel()

	repoSlug, createdRepo, mainBranch := stepCreateRepository(t, s.bb, s.workspace, s.project, "repo-lifecycle", true)
	s.NotEmpty(repoSlug, "Repository slug should not be empty")
	s.NotNil(createdRepo, "Repository creation failed")
	s.NotEmpty(mainBranch, "Main branch should not be empty")

	t.Cleanup(func() {
		stepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	stepVerifyGetRepository(t, s.bb, s.workspace, repoSlug, createdRepo)
	stepVerifyListRepositories(t, s.bb, s.workspace, repoSlug, createdRepo)
}

// TestRepositorySourceTree verifies repository source tree structure with nested directories.
func (s *IntegrationTestSuite_BasicAuth) TestRepositorySourceTree() {
	t := s.T()
	t.Parallel()

	repoSlug, _, mainBranch := stepCreateRepository(t, s.bb, s.workspace, s.project, "source-tree", true)
	s.NotEmpty(repoSlug, "Repository slug should not be empty")
	s.NotEmpty(mainBranch, "Main branch should not be empty")

	t.Cleanup(func() {
		stepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"src/main/file.go":  "package main\n\nfunc main() {}\n",
		"docs/README.md":    "# Documentation\n\nThis is a test file.\n",
		"src/utils/util.go": "package utils\n\nfunc Helper() {}\n",
		"README.md":         "# Test Repository\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit with nested structure", "", "")

	stepVerifySourceTree(t, s.bb, s.workspace, repoSlug, []string{"src", "docs", "README.md"})
	stepVerifyDirectorySource(t, s.bb, s.workspace, repoSlug, mainBranch, "src", []string{"src/main", "src/utils"})
	stepVerifyFileContent(t, s.bb, s.workspace, repoSlug, mainBranch, "src/main/file.go", []string{"package main", "func main()"})
}

// TestPullRequestBasics verifies pull request creation and retrieval.
func (s *IntegrationTestSuite_BasicAuth) TestPullRequestBasics() {
	t := s.T()
	t.Parallel()

	featureBranch := "feature-test-branch"

	repoSlug, _, mainBranch := stepCreateRepository(t, s.bb, s.workspace, s.project, "pr-basics", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	commitHash := stepGetLatestCommitHash(t, s.bb, s.workspace, repoSlug)
	s.NotEmpty(commitHash, "Commit hash should not be empty")

	branch := stepCreateBranch(t, s.bb, s.workspace, repoSlug, featureBranch, commitHash)
	s.NotNil(branch, "Branch creation failed")

	updatedFiles := map[string]string{
		"README.md": "# Test Repository\n\nUpdated content from feature branch.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, updatedFiles, "Update README on feature branch", featureBranch, "")

	pr := stepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Test Pull Request", "This is a test pull request", featureBranch, mainBranch)
	s.NotNil(pr, "Pull request creation failed")

	stepVerifyGetPullRequest(t, s.bb, s.workspace, repoSlug, pr.ID, "Test Pull Request", featureBranch)
	stepVerifyListPullRequests(t, s.bb, s.workspace, repoSlug, pr.ID, "Test Pull Request")
}

// TestPullRequestCommitsAndDiffAndComments verifies PR commits, diff, and comments with pagination.
func (s *IntegrationTestSuite_BasicAuth) TestPullRequestCommitsAndDiffAndComments() {
	t := s.T()
	t.Parallel()

	featureBranch := "feature-multiple-commits"

	repoSlug, _, mainBranch := stepCreateRepository(t, s.bb, s.workspace, s.project, "pr-commits-diff-comments", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
		"file1.txt": "First file content.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	commitHash := stepGetLatestCommitHash(t, s.bb, s.workspace, repoSlug)
	s.NotEmpty(commitHash, "Commit hash should not be empty")

	branch := stepCreateBranch(t, s.bb, s.workspace, repoSlug, featureBranch, commitHash)
	s.NotNil(branch, "Branch creation failed")

	commit1Files := map[string]string{
		"README.md": "# Test Repository\n\nUpdated by commit 1.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, commit1Files, "Commit 1: Update README", featureBranch, "")

	commit2Files := map[string]string{
		"file1.txt": "First file updated in commit 2.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, commit2Files, "Commit 2: Update file1.txt", featureBranch, "")

	commit3Files := map[string]string{
		"file2.txt": "New file added in commit 3.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, commit3Files, "Commit 3: Add file2.txt", featureBranch, "")

	pr := stepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Test PR with Multiple Commits", "This PR has 3 commits", featureBranch, mainBranch)
	s.NotNil(pr, "Pull request creation failed")

	stepVerifyListPullRequestCommits(t, s.bb, s.workspace, repoSlug, pr.ID, []string{
		"Commit 1: Update README",
		"Commit 2: Update file1.txt",
		"Commit 3: Add file2.txt",
	})

	stepVerifyGetPullRequestDiff(t, s.bb, s.workspace, repoSlug, pr.ID, []string{
		"README.md",
		"file1.txt",
		"file2.txt",
		"Updated by commit 1",
		"First file updated in commit 2",
		"New file added in commit 3",
	})

	comment1 := stepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "First comment on this PR")
	s.NotNil(comment1, "First comment creation failed")

	comment2 := stepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "Second comment for testing")
	s.NotNil(comment2, "Second comment creation failed")

	comment3 := stepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "Third comment here")
	s.NotNil(comment3, "Third comment creation failed")

	comment4 := stepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "Fourth comment to test pagination")
	s.NotNil(comment4, "Fourth comment creation failed")

	comment5 := stepCreatePullRequestComment(t, s.bb, s.workspace, repoSlug, pr.ID, "Fifth and final comment")
	s.NotNil(comment5, "Fifth comment creation failed")

	stepVerifyListPullRequestComments(t, s.bb, s.workspace, repoSlug, pr.ID, 3, 1, 3, []string{
		"First comment on this PR",
		"Second comment for testing",
		"Third comment here",
	})

	stepVerifyListPullRequestComments(t, s.bb, s.workspace, repoSlug, pr.ID, 3, 2, 2, []string{
		"Fourth comment to test pagination",
		"Fifth and final comment",
	})
}

// TestPullRequestStates verifies filtering pull requests by state (OPEN, MERGED, DECLINED).
func (s *IntegrationTestSuite_BasicAuth) TestPullRequestStates() {
	t := s.T()
	t.Parallel()

	repoSlug, _, mainBranch := stepCreateRepository(t, s.bb, s.workspace, s.project, "pr-states", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	commitHash := stepGetLatestCommitHash(t, s.bb, s.workspace, repoSlug)
	s.NotEmpty(commitHash, "Commit hash should not be empty")

	branch1 := "feature-open"
	stepCreateBranch(t, s.bb, s.workspace, repoSlug, branch1, commitHash)
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, map[string]string{
		"file1.txt": "Content for open PR\n",
	}, "Add file1 for open PR", branch1, "")
	pr1 := stepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Open Pull Request", "This PR will stay open", branch1, mainBranch)
	s.NotNil(pr1, "PR 1 creation failed")

	branch2 := "feature-merged"
	stepCreateBranch(t, s.bb, s.workspace, repoSlug, branch2, commitHash)
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, map[string]string{
		"file2.txt": "Content for merged PR\n",
	}, "Add file2 for merged PR", branch2, "")
	pr2 := stepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Merged Pull Request", "This PR will be merged", branch2, mainBranch)
	s.NotNil(pr2, "PR 2 creation failed")
	stepMergePullRequest(t, s.bb, s.workspace, repoSlug, pr2.ID)

	branch3 := "feature-declined"
	stepCreateBranch(t, s.bb, s.workspace, repoSlug, branch3, commitHash)
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, map[string]string{
		"file3.txt": "Content for declined PR\n",
	}, "Add file3 for declined PR", branch3, "")
	pr3 := stepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Declined Pull Request", "This PR will be declined", branch3, mainBranch)
	s.NotNil(pr3, "PR 3 creation failed")
	stepDeclinePullRequest(t, s.bb, s.workspace, repoSlug, pr3.ID)

	stepVerifyListPullRequestsByState(t, s.bb, s.workspace, repoSlug, []string{"OPEN"}, []int{pr1.ID}, []int{pr2.ID, pr3.ID})
	stepVerifyListPullRequestsByState(t, s.bb, s.workspace, repoSlug, []string{"MERGED"}, []int{pr2.ID}, []int{pr1.ID, pr3.ID})
	stepVerifyListPullRequestsByState(t, s.bb, s.workspace, repoSlug, []string{"DECLINED"}, []int{pr3.ID}, []int{pr1.ID, pr2.ID})
}

// TestErrorNonExistentWorkspace verifies error handling for non-existent workspace.
func (s *IntegrationTestSuite_BasicAuth) TestErrorNonExistentWorkspace() {
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
func (s *IntegrationTestSuite_BasicAuth) TestErrorNonExistentRepository() {
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
func (s *IntegrationTestSuite_BasicAuth) TestErrorNonExistentPullRequest() {
	t := s.T()
	t.Parallel()

	repoSlug, _, _ := stepCreateRepository(t, s.bb, s.workspace, s.project, "error-nonexistent-pr", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, s.bb, s.workspace, repoSlug)
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
func (s *IntegrationTestSuite_BasicAuth) TestErrorNonExistentFileDirectory() {
	t := s.T()
	t.Parallel()

	repoSlug, _, mainBranch := stepCreateRepository(t, s.bb, s.workspace, s.project, "error-nonexistent-file", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

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

// IntegrationTestSuite_OAuth is the test suite for Bitbucket client integration tests using OAuth
type IntegrationTestSuite_OAuth struct {
	suite.Suite
	bb        *client.Client
	workspace string
	project   string
	token     string
}

func TestIntegration_OAuth(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite_OAuth))
}

func (s *IntegrationTestSuite_OAuth) SetupSuite() {
	clientId := config.GetString("TEST_BITBUCKET_CLIENT_ID", "")
	clientSecret := config.GetString("TEST_BITBUCKET_CLIENT_SECRET", "")
	tokenUrl := config.GetString("TEST_BITBUCKET_ACCESS_TOKEN_URL", "https://bitbucket.org/site/oauth2/access_token")

	token, err := util.ObtainAccessToken(clientId, clientSecret, tokenUrl)
	s.Require().NoError(err)
	s.token = token

	s.workspace = config.GetString("TEST_BITBUCKET_WORKSPACE", "")
	s.project = config.GetString("TEST_BITBUCKET_PROJECT_KEY", "")

	cfg := client.BitbucketConfig{
		Url:     config.GetString("TEST_BITBUCKET_URL", "https://api.bitbucket.org/2.0"),
		Timeout: config.GetInt("TEST_BITBUCKET_TIMEOUT", 5),
	}

	authorizer := util.NewOAuthAuthorizer(util.NewStaticTokenExtractor(token))
	s.bb = client.NewClient(cfg, authorizer)
}

// TestOAuthBasicOperations verifies that OAuth authentication works for basic Bitbucket operations
func (s *IntegrationTestSuite_OAuth) TestOAuthBasicOperations() {
	t := s.T()
	t.Parallel()

	repoSlug, createdRepo, mainBranch := stepCreateRepository(t, s.bb, s.workspace, s.project, "repo-lifecycle", true)
	s.NotEmpty(repoSlug, "Repository slug should not be empty")
	s.NotNil(createdRepo, "Repository creation failed")
	s.NotEmpty(mainBranch, "Main branch should not be empty")

	t.Cleanup(func() {
		stepDeleteRepository(t, s.bb, s.workspace, repoSlug)
	})

	stepVerifyGetRepository(t, s.bb, s.workspace, repoSlug, createdRepo)
	stepVerifyListRepositories(t, s.bb, s.workspace, repoSlug, createdRepo)

	featureBranch := "feature-test-branch"

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, files, "Initial commit", "", "")

	commitHash := stepGetLatestCommitHash(t, s.bb, s.workspace, repoSlug)
	s.NotEmpty(commitHash, "Commit hash should not be empty")

	branch := stepCreateBranch(t, s.bb, s.workspace, repoSlug, featureBranch, commitHash)
	s.NotNil(branch, "Branch creation failed")

	updatedFiles := map[string]string{
		"README.md": "# Test Repository\n\nUpdated content from feature branch.\n",
	}
	stepCreateFiles(t, s.bb, s.workspace, repoSlug, updatedFiles, "Update README on feature branch", featureBranch, "")

	pr := stepCreatePullRequest(t, s.bb, s.workspace, repoSlug, "Test Pull Request", "This is a test pull request", featureBranch, mainBranch)
	s.NotNil(pr, "Pull request creation failed")

	stepVerifyGetPullRequest(t, s.bb, s.workspace, repoSlug, pr.ID, "Test Pull Request", featureBranch)
	stepVerifyListPullRequests(t, s.bb, s.workspace, repoSlug, pr.ID, "Test Pull Request")
}

func generateRepoSlug(testName string) string {
	return fmt.Sprintf("%s-%d", testName, time.Now().UnixNano())
}

func stepCreateRepository(t *testing.T, bb *client.Client, workspace, testProjectKey, testName string, isPrivate bool) (string, *client.Repository, string) {
	t.Helper()
	var repo *client.Repository
	var mainBranch string

	repoSlug := generateRepoSlug(testName)
	description := fmt.Sprintf("Test repository for %s", testName)

	t.Run("create repository", func(t *testing.T) {
		createReq := &client.CreateRepositoryRequest{
			SCM:         "git",
			IsPrivate:   &isPrivate,
			Description: description,
			Project: &client.CreateRepositoryProjectRef{
				Key: testProjectKey,
			},
		}

		var err error
		repo, err = bb.CreateRepository(context.Background(), workspace, repoSlug, createReq)
		require.NoError(t, err, "Failed to create repository")
		require.NotNil(t, repo, "Created repository should not be nil")
		require.Equal(t, repoSlug, repo.Slug, "Repository slug should match")
		require.Equal(t, description, repo.Description, "Repository description should match")
		require.Equal(t, isPrivate, repo.IsPrivate, "Repository privacy should match")
		require.Equal(t, "git", repo.SCM, "Repository SCM should be git")

		mainBranch = repo.MainBranch.Name
		require.NotEmpty(t, mainBranch, "Main branch name should not be empty")
	})

	return repoSlug, repo, mainBranch
}

func stepDeleteRepository(t *testing.T, bb *client.Client, workspace, repoSlug string) {
	t.Helper()
	if err := bb.DeleteRepository(context.Background(), workspace, repoSlug); err != nil {
		t.Logf("Warning: Failed to delete repository %s: %v", repoSlug, err)
	}
}

func stepCreateFiles(t *testing.T, bb *client.Client, workspace, repoSlug string, files map[string]string, message string, branch string, parents string) {
	t.Helper()

	stepName := "create files on main branch"
	if branch != "" {
		stepName = fmt.Sprintf("create files on %s branch", branch)
	}

	t.Run(stepName, func(t *testing.T) {
		createFilesReq := &client.CreateFilesRequest{
			Branch:  branch,
			Message: message,
			Files:   files,
			Parents: parents,
		}

		err := bb.CreateOrUpdateFiles(context.Background(), workspace, repoSlug, createFilesReq)
		require.NoError(t, err, "Failed to create files")
	})
}

func stepGetLatestCommitHash(t *testing.T, bb *client.Client, workspace, repoSlug string) string {
	t.Helper()
	var commitHash string

	t.Run("get latest commit hash", func(t *testing.T) {
		sourceTree, err := bb.GetRepositorySource(context.Background(), workspace, repoSlug)
		require.NoError(t, err, "Failed to get repository source")
		require.NotEmpty(t, sourceTree.Values, "Source tree should not be empty")

		commitHash = sourceTree.Values[0].Commit.Hash
		require.NotEmpty(t, commitHash, "Commit hash should not be empty")
	})

	return commitHash
}

func stepCreateBranch(t *testing.T, bb *client.Client, workspace, repoSlug string, branchName string, commitHash string) *client.Branch {
	t.Helper()
	var branch *client.Branch

	t.Run(fmt.Sprintf("create branch '%s'", branchName), func(t *testing.T) {
		createBranchReq := &client.CreateBranchRequest{
			Name: branchName,
			Target: client.CreateBranchTarget{
				Hash: commitHash,
			},
		}

		var err error
		branch, err = bb.CreateBranch(context.Background(), workspace, repoSlug, createBranchReq)
		require.NoError(t, err, "Failed to create branch")
		require.NotNil(t, branch, "Created branch should not be nil")
		require.Equal(t, branchName, branch.Name, "Branch name should match")
		require.Equal(t, commitHash, branch.Target.Hash, "Branch target hash should match")
	})

	return branch
}

func stepCreatePullRequest(t *testing.T, bb *client.Client, workspace, repoSlug string, title string, description string, sourceBranch string, destBranch string) *client.PullRequest {
	t.Helper()
	var pr *client.PullRequest

	t.Run(fmt.Sprintf("create pull request '%s'", title), func(t *testing.T) {
		createPRReq := &client.CreatePullRequestRequest{
			Title:       title,
			Description: description,
			Source: client.CreatePullRequestBranch{
				Branch: client.CreatePullRequestBranchName{
					Name: sourceBranch,
				},
			},
			Destination: &client.CreatePullRequestBranch{
				Branch: client.CreatePullRequestBranchName{
					Name: destBranch,
				},
			},
		}

		var err error
		pr, err = bb.CreatePullRequest(context.Background(), workspace, repoSlug, createPRReq)
		require.NoError(t, err, "Failed to create pull request")
		require.NotNil(t, pr, "Pull request should not be nil")
		require.Equal(t, title, pr.Title, "PR title should match")
		require.Equal(t, description, pr.Description, "PR description should match")
		require.Equal(t, sourceBranch, pr.Source.Branch.Name, "PR source branch should match")
		require.Equal(t, destBranch, pr.Destination.Branch.Name, "PR destination branch should match")
	})

	return pr
}

func stepVerifyGetRepository(t *testing.T, bb *client.Client, workspace, repoSlug string, expectedRepo *client.Repository) {
	t.Helper()

	t.Run("verify get repository", func(t *testing.T) {
		fetchedRepo, err := bb.GetRepository(context.Background(), workspace, repoSlug)
		require.NoError(t, err, "Failed to get repository")
		require.NotNil(t, fetchedRepo, "Fetched repository should not be nil")
		require.Equal(t, repoSlug, fetchedRepo.Slug, "Fetched repository slug should match")
		require.Equal(t, expectedRepo.Description, fetchedRepo.Description, "Fetched repository description should match")
		require.Equal(t, expectedRepo.IsPrivate, fetchedRepo.IsPrivate, "Fetched repository privacy should match")
		require.Equal(t, expectedRepo.UUID, fetchedRepo.UUID, "Repository UUID should match")
	})
}

func stepVerifyListRepositories(t *testing.T, bb *client.Client, workspace, repoSlug string, expectedRepo *client.Repository) {
	t.Helper()

	t.Run("verify list repositories includes created repo", func(t *testing.T) {
		repoList, err := bb.ListRepositories(context.Background(), workspace, 50, 1)
		require.NoError(t, err, "Failed to list repositories")
		require.NotNil(t, repoList, "Repository list should not be nil")

		found := false
		for _, repo := range repoList.Values {
			if repo.Slug == repoSlug {
				found = true
				require.Equal(t, expectedRepo.Description, repo.Description, "Listed repository description should match")
				require.Equal(t, expectedRepo.IsPrivate, repo.IsPrivate, "Listed repository privacy should match")
				break
			}
		}
		require.True(t, found, "Created repository should appear in repository list")
	})
}

func stepVerifySourceTree(t *testing.T, bb *client.Client, workspace, repoSlug string, expectedPaths []string) {
	t.Helper()

	t.Run("verify source tree structure", func(t *testing.T) {
		sourceTree, err := bb.GetRepositorySource(context.Background(), workspace, repoSlug)
		require.NoError(t, err, "Failed to get repository source")
		require.NotNil(t, sourceTree, "Source tree should not be nil")

		pathsFound := make(map[string]bool)
		for _, path := range expectedPaths {
			pathsFound[path] = false
		}

		for _, item := range sourceTree.Values {
			if _, exists := pathsFound[item.Path]; exists {
				pathsFound[item.Path] = true
			}
		}

		for path, found := range pathsFound {
			require.True(t, found, "Expected path %s not found in source tree", path)
		}
	})
}

func stepVerifyDirectorySource(t *testing.T, bb *client.Client, workspace, repoSlug string, mainBranch string, dirPath string, expectedPaths []string) {
	t.Helper()

	t.Run(fmt.Sprintf("verify directory '%s' structure", dirPath), func(t *testing.T) {
		dirTree, err := bb.GetDirectorySource(context.Background(), workspace, repoSlug, mainBranch, dirPath)
		require.NoError(t, err, "Failed to get directory source")
		require.NotNil(t, dirTree, "Directory tree should not be nil")

		pathsFound := make(map[string]bool)
		for _, path := range expectedPaths {
			pathsFound[path] = false
		}

		for _, item := range dirTree.Values {
			if _, exists := pathsFound[item.Path]; exists {
				pathsFound[item.Path] = true
			}
		}

		for path, found := range pathsFound {
			require.True(t, found, "Expected path %s not found in directory", path)
		}
	})
}

func stepVerifyFileContent(t *testing.T, bb *client.Client, workspace, repoSlug string, mainBranch string, filePath string, expectedSubstrings []string) {
	t.Helper()

	t.Run(fmt.Sprintf("verify file '%s' content", filePath), func(t *testing.T) {
		fileContent, err := bb.GetFileSource(context.Background(), workspace, repoSlug, mainBranch, filePath)
		require.NoError(t, err, "Failed to get file source")
		require.NotNil(t, fileContent, "File content should not be nil")

		for _, substring := range expectedSubstrings {
			require.Contains(t, *fileContent, substring, "File content should contain '%s'", substring)
		}
	})
}

func stepVerifyGetPullRequest(t *testing.T, bb *client.Client, workspace, repoSlug string, prID int, expectedTitle string, expectedSourceBranch string) {
	t.Helper()

	t.Run("verify get pull request", func(t *testing.T) {
		pr, err := bb.GetPullRequest(context.Background(), workspace, repoSlug, prID)
		require.NoError(t, err, "Failed to get pull request")
		require.NotNil(t, pr, "Pull request should not be nil")
		require.Equal(t, prID, pr.ID, "PR ID should match")
		require.Equal(t, expectedTitle, pr.Title, "PR title should match")
		require.Equal(t, expectedSourceBranch, pr.Source.Branch.Name, "PR source branch should match")
	})
}

func stepVerifyListPullRequests(t *testing.T, bb *client.Client, workspace, repoSlug string, prID int, expectedTitle string) {
	t.Helper()

	t.Run("verify list pull requests", func(t *testing.T) {
		prList, err := bb.ListPullRequests(context.Background(), workspace, repoSlug, 50, 1, []string{"OPEN"})
		require.NoError(t, err, "Failed to list pull requests")
		require.NotNil(t, prList, "Pull request list should not be nil")

		found := false
		for _, pr := range prList.Values {
			if pr.ID == prID {
				found = true
				require.Equal(t, expectedTitle, pr.Title, "Listed PR title should match")
				break
			}
		}
		require.True(t, found, "Created pull request should appear in pull request list")
	})
}

func stepCreatePullRequestComment(t *testing.T, bb *client.Client, workspace, repoSlug string, prID int, commentText string) *client.PullRequestComment {
	t.Helper()
	var comment *client.PullRequestComment

	t.Run(fmt.Sprintf("create comment '%s'", commentText), func(t *testing.T) {
		createCommentReq := &client.CreatePullRequestCommentRequest{
			Content: client.CreatePullRequestCommentContent{
				Raw: commentText,
			},
		}

		var err error
		comment, err = bb.CreatePullRequestComment(context.Background(), workspace, repoSlug, prID, createCommentReq)
		require.NoError(t, err, "Failed to create comment")
		require.NotNil(t, comment, "Created comment should not be nil")
		require.Equal(t, commentText, comment.Content.Raw, "Comment text should match")
	})

	return comment
}

func stepVerifyListPullRequestCommits(t *testing.T, bb *client.Client, workspace, repoSlug string, prID int, expectedCommitMessages []string) {
	t.Helper()

	t.Run("verify list pull request commits", func(t *testing.T) {
		commits, err := bb.ListPullRequestCommits(context.Background(), workspace, repoSlug, prID)
		require.NoError(t, err, "Failed to list pull request commits")
		require.NotNil(t, commits, "Commits list should not be nil")
		require.GreaterOrEqual(t, len(commits.Values), len(expectedCommitMessages), "Should have at least %d commits", len(expectedCommitMessages))

		// Verify each expected commit message appears in the commits list
		for _, expectedMsg := range expectedCommitMessages {
			found := false
			for _, commit := range commits.Values {
				if commit.Message == expectedMsg {
					found = true
					break
				}
			}
			require.True(t, found, "Expected commit message '%s' not found in commits", expectedMsg)
		}
	})
}

func stepVerifyGetPullRequestDiff(t *testing.T, bb *client.Client, workspace, repoSlug string, prID int, expectedSubstrings []string) {
	t.Helper()

	t.Run("verify get pull request diff", func(t *testing.T) {
		diff, err := bb.GetPullRequestDiff(context.Background(), workspace, repoSlug, prID)
		require.NoError(t, err, "Failed to get pull request diff")
		require.NotNil(t, diff, "Diff should not be nil")
		require.NotEmpty(t, *diff, "Diff content should not be empty")

		// Verify each expected substring appears in the diff
		for _, substring := range expectedSubstrings {
			require.Contains(t, *diff, substring, "Diff should contain '%s'", substring)
		}
	})
}

func stepVerifyListPullRequestComments(t *testing.T, bb *client.Client, workspace, repoSlug string, prID int, pagelen int, page int, expectedCount int, expectedCommentTexts []string) {
	t.Helper()

	t.Run(fmt.Sprintf("verify list pull request comments (page %d, pagelen %d)", page, pagelen), func(t *testing.T) {
		comments, err := bb.ListPullRequestComments(context.Background(), workspace, repoSlug, prID, pagelen, page)
		require.NoError(t, err, "Failed to list pull request comments")
		require.NotNil(t, comments, "Comments list should not be nil")
		require.Equal(t, expectedCount, len(comments.Values), "Should have %d comments on page %d", expectedCount, page)

		// Verify each expected comment text appears in the comments list
		for _, expectedText := range expectedCommentTexts {
			found := false
			for _, comment := range comments.Values {
				if comment.Content.Raw == expectedText {
					found = true
					break
				}
			}
			require.True(t, found, "Expected comment '%s' not found on page %d", expectedText, page)
		}
	})
}

func stepMergePullRequest(t *testing.T, bb *client.Client, workspace, repoSlug string, prID int) {
	t.Helper()

	t.Run(fmt.Sprintf("merge pull request #%d", prID), func(t *testing.T) {
		mergeReq := &client.MergePullRequestRequest{
			Type:    "merge_commit",
			Message: fmt.Sprintf("Merge pull request #%d", prID),
		}

		pr, err := bb.MergePullRequest(context.Background(), workspace, repoSlug, prID, mergeReq)
		require.NoError(t, err, "Failed to merge pull request")
		require.NotNil(t, pr, "Merged pull request should not be nil")
		require.Equal(t, "MERGED", pr.State, "Pull request state should be MERGED")
	})
}

func stepDeclinePullRequest(t *testing.T, bb *client.Client, workspace, repoSlug string, prID int) {
	t.Helper()

	t.Run(fmt.Sprintf("decline pull request #%d", prID), func(t *testing.T) {
		pr, err := bb.DeclinePullRequest(context.Background(), workspace, repoSlug, prID)
		require.NoError(t, err, "Failed to decline pull request")
		require.NotNil(t, pr, "Declined pull request should not be nil")
		require.Equal(t, "DECLINED", pr.State, "Pull request state should be DECLINED")
	})
}

func stepVerifyListPullRequestsByState(t *testing.T, bb *client.Client, workspace, repoSlug string, states []string, expectedPRIDs []int, unexpectedPRIDs []int) {
	t.Helper()

	stateStr := states[0]
	t.Run(fmt.Sprintf("verify list pull requests with state filter [%s]", stateStr), func(t *testing.T) {
		prList, err := bb.ListPullRequests(context.Background(), workspace, repoSlug, 50, 1, states)
		require.NoError(t, err, "Failed to list pull requests with state filter %v", states)
		require.NotNil(t, prList, "Pull request list should not be nil")

		foundPRIDs := make(map[int]bool)
		for _, pr := range prList.Values {
			foundPRIDs[pr.ID] = true
			require.Contains(t, states, pr.State, "PR #%d has unexpected state %s, expected one of %v", pr.ID, pr.State, states)
		}

		for _, prID := range expectedPRIDs {
			require.True(t, foundPRIDs[prID], "Expected PR #%d with state %s not found in results", prID, stateStr)
		}

		for _, prID := range unexpectedPRIDs {
			require.False(t, foundPRIDs[prID], "Unexpected PR #%d found in results for state filter %s", prID, stateStr)
		}
	})
}
