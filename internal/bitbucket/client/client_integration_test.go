//go:build integration

// Package client_test provides integration tests for the Bitbucket API client.
//
// # Running Integration Tests
//
// Integration tests connect to a real Bitbucket instance and require proper configuration.
// To run these tests, use the following command:
//
//  go test -tags=integration ./internal/bitbucket/client
//
// # Required Environment Variables
//
// The following environment variables must be set to connect to Bitbucket:
//
//  BITBUCKET_URL               - Base URL of the Bitbucket instance (e.g., https://api.bitbucket.org/2.0)
//  BITBUCKET_EMAIL             - Email address for authentication
//  BITBUCKET_API_TOKEN         - API token or app password for authentication
//  BITBUCKET_TEST_WORKSPACE    - Workspace slug where test repositories will be created
//  BITBUCKET_TEST_PROJECT_KEY  - Project key where test repositories will be created
//  BITBUCKET_TIMEOUT           - Optional: Request timeout in seconds (default: 5)
//
// Example configuration (or use .env file):
//
//  export BITBUCKET_URL="https://api.bitbucket.org/2.0"
//  export BITBUCKET_EMAIL="your-email@example.com"
//  export BITBUCKET_API_TOKEN="your-app-password"
//  export BITBUCKET_TEST_WORKSPACE="your-workspace"
//  export BITBUCKET_TEST_PROJECT_KEY="TEST"
//
// # Test Cleanup
//
// All integration tests are designed to clean up after themselves:
//  - Test repositories are automatically deleted after each test using t.Cleanup()
//  - Temporary branches, pull requests, and other resources are created in test repositories
//  - Failed tests may leave artifacts; check your workspace if tests are interrupted
package client_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/stretchr/testify/require"
)

var cfg = &itConfig{}
var bb = client.NewClient(cfg)
var testWorkspace = cfg.BitbucketTestWorkspace()
var testProjectKey = cfg.BitbucketTestProjectKey()

// TestRepositoryLifecycle verifies repository creation, retrieval, listing, and deletion.
func TestRepositoryLifecycle(t *testing.T) {
	t.Parallel()

	repoSlug, createdRepo, mainBranch := stepCreateRepository(t, "repo-lifecycle", true)
	require.NotEmpty(t, repoSlug, "Repository slug should not be empty")
	require.NotNil(t, createdRepo, "Repository creation failed")
	require.NotEmpty(t, mainBranch, "Main branch should not be empty")

	t.Cleanup(func() {
		stepDeleteRepository(t, repoSlug)
	})

	stepVerifyGetRepository(t, repoSlug, createdRepo)
	stepVerifyListRepositories(t, repoSlug, createdRepo)
}

// TestRepositorySourceTree verifies repository source tree structure with nested directories.
func TestRepositorySourceTree(t *testing.T) {
	t.Parallel()

	repoSlug, _, mainBranch := stepCreateRepository(t, "source-tree", true)
	require.NotEmpty(t, repoSlug, "Repository slug should not be empty")
	require.NotEmpty(t, mainBranch, "Main branch should not be empty")

	t.Cleanup(func() {
		stepDeleteRepository(t, repoSlug)
	})

	files := map[string]string{
		"src/main/file.go":  "package main\n\nfunc main() {}\n",
		"docs/README.md":    "# Documentation\n\nThis is a test file.\n",
		"src/utils/util.go": "package utils\n\nfunc Helper() {}\n",
		"README.md":         "# Test Repository\n",
	}
	stepCreateFiles(t, repoSlug, files, "Initial commit with nested structure", "", "")

	stepVerifySourceTree(t, repoSlug, []string{"src", "docs", "README.md"})
	stepVerifyDirectorySource(t, repoSlug, mainBranch, "src", []string{"src/main", "src/utils"})
	stepVerifyFileContent(t, repoSlug, mainBranch, "src/main/file.go", []string{"package main", "func main()"})
}

// TestPullRequestBasics verifies pull request creation and retrieval.
func TestPullRequestBasics(t *testing.T) {
	t.Parallel()

	featureBranch := "feature-test-branch"

	repoSlug, _, mainBranch := stepCreateRepository(t, "pr-basics", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
	}
	stepCreateFiles(t, repoSlug, files, "Initial commit", "", "")

	commitHash := stepGetLatestCommitHash(t, repoSlug)
	require.NotEmpty(t, commitHash, "Commit hash should not be empty")

	branch := stepCreateBranch(t, repoSlug, featureBranch, commitHash)
	require.NotNil(t, branch, "Branch creation failed")

	updatedFiles := map[string]string{
		"README.md": "# Test Repository\n\nUpdated content from feature branch.\n",
	}
	stepCreateFiles(t, repoSlug, updatedFiles, "Update README on feature branch", featureBranch, "")

	pr := stepCreatePullRequest(t, repoSlug, "Test Pull Request", "This is a test pull request", featureBranch, mainBranch)
	require.NotNil(t, pr, "Pull request creation failed")

	stepVerifyGetPullRequest(t, repoSlug, pr.ID, "Test Pull Request", featureBranch)
	stepVerifyListPullRequests(t, repoSlug, pr.ID, "Test Pull Request")
}

// TestPullRequestCommitsAndDiffAndComments verifies PR commits, diff, and comments with pagination.
func TestPullRequestCommitsAndDiffAndComments(t *testing.T) {
	t.Parallel()

	featureBranch := "feature-multiple-commits"

	repoSlug, _, mainBranch := stepCreateRepository(t, "pr-commits-diff-comments", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
		"file1.txt": "First file content.\n",
	}
	stepCreateFiles(t, repoSlug, files, "Initial commit", "", "")

	commitHash := stepGetLatestCommitHash(t, repoSlug)
	require.NotEmpty(t, commitHash, "Commit hash should not be empty")

	branch := stepCreateBranch(t, repoSlug, featureBranch, commitHash)
	require.NotNil(t, branch, "Branch creation failed")

	commit1Files := map[string]string{
		"README.md": "# Test Repository\n\nUpdated by commit 1.\n",
	}
	stepCreateFiles(t, repoSlug, commit1Files, "Commit 1: Update README", featureBranch, "")

	commit2Files := map[string]string{
		"file1.txt": "First file updated in commit 2.\n",
	}
	stepCreateFiles(t, repoSlug, commit2Files, "Commit 2: Update file1.txt", featureBranch, "")

	commit3Files := map[string]string{
		"file2.txt": "New file added in commit 3.\n",
	}
	stepCreateFiles(t, repoSlug, commit3Files, "Commit 3: Add file2.txt", featureBranch, "")

	pr := stepCreatePullRequest(t, repoSlug, "Test PR with Multiple Commits", "This PR has 3 commits", featureBranch, mainBranch)
	require.NotNil(t, pr, "Pull request creation failed")

	stepVerifyListPullRequestCommits(t, repoSlug, pr.ID, []string{
		"Commit 1: Update README",
		"Commit 2: Update file1.txt",
		"Commit 3: Add file2.txt",
	})

	stepVerifyGetPullRequestDiff(t, repoSlug, pr.ID, []string{
		"README.md",
		"file1.txt",
		"file2.txt",
		"Updated by commit 1",
		"First file updated in commit 2",
		"New file added in commit 3",
	})

	comment1 := stepCreatePullRequestComment(t, repoSlug, pr.ID, "First comment on this PR")
	require.NotNil(t, comment1, "First comment creation failed")

	comment2 := stepCreatePullRequestComment(t, repoSlug, pr.ID, "Second comment for testing")
	require.NotNil(t, comment2, "Second comment creation failed")

	comment3 := stepCreatePullRequestComment(t, repoSlug, pr.ID, "Third comment here")
	require.NotNil(t, comment3, "Third comment creation failed")

	comment4 := stepCreatePullRequestComment(t, repoSlug, pr.ID, "Fourth comment to test pagination")
	require.NotNil(t, comment4, "Fourth comment creation failed")

	comment5 := stepCreatePullRequestComment(t, repoSlug, pr.ID, "Fifth and final comment")
	require.NotNil(t, comment5, "Fifth comment creation failed")

	stepVerifyListPullRequestComments(t, repoSlug, pr.ID, 3, 1, 3, []string{
		"First comment on this PR",
		"Second comment for testing",
		"Third comment here",
	})

	stepVerifyListPullRequestComments(t, repoSlug, pr.ID, 3, 2, 2, []string{
		"Fourth comment to test pagination",
		"Fifth and final comment",
	})
}

// TestPullRequestStates verifies filtering pull requests by state (OPEN, MERGED, DECLINED).
func TestPullRequestStates(t *testing.T) {
	t.Parallel()

	repoSlug, _, mainBranch := stepCreateRepository(t, "pr-states", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n\nInitial content.\n",
	}
	stepCreateFiles(t, repoSlug, files, "Initial commit", "", "")

	commitHash := stepGetLatestCommitHash(t, repoSlug)
	require.NotEmpty(t, commitHash, "Commit hash should not be empty")

	branch1 := "feature-open"
	stepCreateBranch(t, repoSlug, branch1, commitHash)
	stepCreateFiles(t, repoSlug, map[string]string{
		"file1.txt": "Content for open PR\n",
	}, "Add file1 for open PR", branch1, "")
	pr1 := stepCreatePullRequest(t, repoSlug, "Open Pull Request", "This PR will stay open", branch1, mainBranch)
	require.NotNil(t, pr1, "PR 1 creation failed")

	branch2 := "feature-merged"
	stepCreateBranch(t, repoSlug, branch2, commitHash)
	stepCreateFiles(t, repoSlug, map[string]string{
		"file2.txt": "Content for merged PR\n",
	}, "Add file2 for merged PR", branch2, "")
	pr2 := stepCreatePullRequest(t, repoSlug, "Merged Pull Request", "This PR will be merged", branch2, mainBranch)
	require.NotNil(t, pr2, "PR 2 creation failed")
	stepMergePullRequest(t, repoSlug, pr2.ID)

	branch3 := "feature-declined"
	stepCreateBranch(t, repoSlug, branch3, commitHash)
	stepCreateFiles(t, repoSlug, map[string]string{
		"file3.txt": "Content for declined PR\n",
	}, "Add file3 for declined PR", branch3, "")
	pr3 := stepCreatePullRequest(t, repoSlug, "Declined Pull Request", "This PR will be declined", branch3, mainBranch)
	require.NotNil(t, pr3, "PR 3 creation failed")
	stepDeclinePullRequest(t, repoSlug, pr3.ID)

	stepVerifyListPullRequestsByState(t, repoSlug, []string{"OPEN"}, []int{pr1.ID}, []int{pr2.ID, pr3.ID})
	stepVerifyListPullRequestsByState(t, repoSlug, []string{"MERGED"}, []int{pr2.ID}, []int{pr1.ID, pr3.ID})
	stepVerifyListPullRequestsByState(t, repoSlug, []string{"DECLINED"}, []int{pr3.ID}, []int{pr1.ID, pr2.ID})
}

// TestErrorNonExistentWorkspace verifies error handling for non-existent workspace.
func TestErrorNonExistentWorkspace(t *testing.T) {
	t.Parallel()

	nonExistentWorkspace := "non-existent-workspace-12345"

	t.Run("list repositories with non-existent workspace", func(t *testing.T) {
		_, err := bb.ListRepositories(nonExistentWorkspace, 10, 1)
		require.Error(t, err, "Should return error for non-existent workspace")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})

	t.Run("get repository with non-existent workspace", func(t *testing.T) {
		_, err := bb.GetRepository(nonExistentWorkspace, "any-repo")
		require.Error(t, err, "Should return error for non-existent workspace")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})
}

// TestErrorNonExistentRepository verifies error handling for non-existent repository.
func TestErrorNonExistentRepository(t *testing.T) {
	t.Parallel()

	nonExistentRepo := "non-existent-repo-12345"

	t.Run("get repository with non-existent repo", func(t *testing.T) {
		_, err := bb.GetRepository(testWorkspace, nonExistentRepo)
		require.Error(t, err, "Should return error for non-existent repository")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})

	t.Run("get repository source with non-existent repo", func(t *testing.T) {
		_, err := bb.GetRepositorySource(testWorkspace, nonExistentRepo)
		require.Error(t, err, "Should return error for non-existent repository")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})

	t.Run("list pull requests with non-existent repo", func(t *testing.T) {
		_, err := bb.ListPullRequests(testWorkspace, nonExistentRepo, 10, 1, []string{"OPEN"})
		require.Error(t, err, "Should return error for non-existent repository")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})
}

// TestErrorNonExistentPullRequest verifies error handling for non-existent pull request.
func TestErrorNonExistentPullRequest(t *testing.T) {
	t.Parallel()

	repoSlug, _, _ := stepCreateRepository(t, "error-nonexistent-pr", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, repoSlug)
	})

	nonExistentPRID := 99999

	t.Run("get pull request with non-existent PR ID", func(t *testing.T) {
		_, err := bb.GetPullRequest(testWorkspace, repoSlug, nonExistentPRID)
		require.Error(t, err, "Should return error for non-existent pull request")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})

	t.Run("list pull request commits with non-existent PR ID", func(t *testing.T) {
		_, err := bb.ListPullRequestCommits(testWorkspace, repoSlug, nonExistentPRID)
		require.Error(t, err, "Should return error for non-existent pull request")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})

	t.Run("list pull request comments with non-existent PR ID", func(t *testing.T) {
		_, err := bb.ListPullRequestComments(testWorkspace, repoSlug, nonExistentPRID, 10, 1)
		require.Error(t, err, "Should return error for non-existent pull request")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})

	t.Run("get pull request diff with non-existent PR ID", func(t *testing.T) {
		_, err := bb.GetPullRequestDiff(testWorkspace, repoSlug, nonExistentPRID)
		require.Error(t, err, "Should return error for non-existent pull request")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})
}

// TestErrorNonExistentFileDirectory verifies error handling for non-existent files and directories.
func TestErrorNonExistentFileDirectory(t *testing.T) {
	t.Parallel()

	repoSlug, _, mainBranch := stepCreateRepository(t, "error-nonexistent-file", true)
	t.Cleanup(func() {
		stepDeleteRepository(t, repoSlug)
	})

	files := map[string]string{
		"README.md": "# Test Repository\n",
	}
	stepCreateFiles(t, repoSlug, files, "Initial commit", "", "")

	t.Run("get file source with non-existent file path", func(t *testing.T) {
		_, err := bb.GetFileSource(testWorkspace, repoSlug, mainBranch, "non-existent-file.txt")
		require.Error(t, err, "Should return error for non-existent file")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})

	t.Run("get directory source with non-existent directory path", func(t *testing.T) {
		_, err := bb.GetDirectorySource(testWorkspace, repoSlug, mainBranch, "non-existent-dir")
		require.Error(t, err, "Should return error for non-existent directory")
		require.ErrorIs(t, err, client.ErrClientBitbucket, "Should be a client error (404)")
	})
}

func generateRepoSlug(testName string) string {
	return fmt.Sprintf("%s-%d", testName, time.Now().UnixNano())
}

func stepCreateRepository(t *testing.T, testName string, isPrivate bool) (string, *client.Repository, string) {
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
		repo, err = bb.CreateRepository(testWorkspace, repoSlug, createReq)
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

func stepDeleteRepository(t *testing.T, repoSlug string) {
	t.Helper()
	if err := bb.DeleteRepository(testWorkspace, repoSlug); err != nil {
		t.Logf("Warning: Failed to delete repository %s: %v", repoSlug, err)
	}
}

func stepCreateFiles(t *testing.T, repoSlug string, files map[string]string, message string, branch string, parents string) {
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

		err := bb.CreateOrUpdateFiles(testWorkspace, repoSlug, createFilesReq)
		require.NoError(t, err, "Failed to create files")
	})
}

func stepGetLatestCommitHash(t *testing.T, repoSlug string) string {
	t.Helper()
	var commitHash string

	t.Run("get latest commit hash", func(t *testing.T) {
		sourceTree, err := bb.GetRepositorySource(testWorkspace, repoSlug)
		require.NoError(t, err, "Failed to get repository source")
		require.NotEmpty(t, sourceTree.Values, "Source tree should not be empty")

		commitHash = sourceTree.Values[0].Commit.Hash
		require.NotEmpty(t, commitHash, "Commit hash should not be empty")
	})

	return commitHash
}

func stepCreateBranch(t *testing.T, repoSlug string, branchName string, commitHash string) *client.Branch {
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
		branch, err = bb.CreateBranch(testWorkspace, repoSlug, createBranchReq)
		require.NoError(t, err, "Failed to create branch")
		require.NotNil(t, branch, "Created branch should not be nil")
		require.Equal(t, branchName, branch.Name, "Branch name should match")
		require.Equal(t, commitHash, branch.Target.Hash, "Branch target hash should match")
	})

	return branch
}

func stepCreatePullRequest(t *testing.T, repoSlug string, title string, description string, sourceBranch string, destBranch string) *client.PullRequest {
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
		pr, err = bb.CreatePullRequest(testWorkspace, repoSlug, createPRReq)
		require.NoError(t, err, "Failed to create pull request")
		require.NotNil(t, pr, "Pull request should not be nil")
		require.Equal(t, title, pr.Title, "PR title should match")
		require.Equal(t, description, pr.Description, "PR description should match")
		require.Equal(t, sourceBranch, pr.Source.Branch.Name, "PR source branch should match")
		require.Equal(t, destBranch, pr.Destination.Branch.Name, "PR destination branch should match")
	})

	return pr
}

func stepVerifyGetRepository(t *testing.T, repoSlug string, expectedRepo *client.Repository) {
	t.Helper()

	t.Run("verify get repository", func(t *testing.T) {
		fetchedRepo, err := bb.GetRepository(testWorkspace, repoSlug)
		require.NoError(t, err, "Failed to get repository")
		require.NotNil(t, fetchedRepo, "Fetched repository should not be nil")
		require.Equal(t, repoSlug, fetchedRepo.Slug, "Fetched repository slug should match")
		require.Equal(t, expectedRepo.Description, fetchedRepo.Description, "Fetched repository description should match")
		require.Equal(t, expectedRepo.IsPrivate, fetchedRepo.IsPrivate, "Fetched repository privacy should match")
		require.Equal(t, expectedRepo.UUID, fetchedRepo.UUID, "Repository UUID should match")
	})
}

func stepVerifyListRepositories(t *testing.T, repoSlug string, expectedRepo *client.Repository) {
	t.Helper()

	t.Run("verify list repositories includes created repo", func(t *testing.T) {
		repoList, err := bb.ListRepositories(testWorkspace, 50, 1)
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

func stepVerifySourceTree(t *testing.T, repoSlug string, expectedPaths []string) {
	t.Helper()

	t.Run("verify source tree structure", func(t *testing.T) {
		sourceTree, err := bb.GetRepositorySource(testWorkspace, repoSlug)
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

func stepVerifyDirectorySource(t *testing.T, repoSlug string, mainBranch string, dirPath string, expectedPaths []string) {
	t.Helper()

	t.Run(fmt.Sprintf("verify directory '%s' structure", dirPath), func(t *testing.T) {
		dirTree, err := bb.GetDirectorySource(testWorkspace, repoSlug, mainBranch, dirPath)
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

func stepVerifyFileContent(t *testing.T, repoSlug string, mainBranch string, filePath string, expectedSubstrings []string) {
	t.Helper()

	t.Run(fmt.Sprintf("verify file '%s' content", filePath), func(t *testing.T) {
		fileContent, err := bb.GetFileSource(testWorkspace, repoSlug, mainBranch, filePath)
		require.NoError(t, err, "Failed to get file source")
		require.NotNil(t, fileContent, "File content should not be nil")

		for _, substring := range expectedSubstrings {
			require.Contains(t, *fileContent, substring, "File content should contain '%s'", substring)
		}
	})
}

func stepVerifyGetPullRequest(t *testing.T, repoSlug string, prID int, expectedTitle string, expectedSourceBranch string) {
	t.Helper()

	t.Run("verify get pull request", func(t *testing.T) {
		pr, err := bb.GetPullRequest(testWorkspace, repoSlug, prID)
		require.NoError(t, err, "Failed to get pull request")
		require.NotNil(t, pr, "Pull request should not be nil")
		require.Equal(t, prID, pr.ID, "PR ID should match")
		require.Equal(t, expectedTitle, pr.Title, "PR title should match")
		require.Equal(t, expectedSourceBranch, pr.Source.Branch.Name, "PR source branch should match")
	})
}

func stepVerifyListPullRequests(t *testing.T, repoSlug string, prID int, expectedTitle string) {
	t.Helper()

	t.Run("verify list pull requests", func(t *testing.T) {
		prList, err := bb.ListPullRequests(testWorkspace, repoSlug, 50, 1, []string{"OPEN"})
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

func stepCreatePullRequestComment(t *testing.T, repoSlug string, prID int, commentText string) *client.PullRequestComment {
	t.Helper()
	var comment *client.PullRequestComment

	t.Run(fmt.Sprintf("create comment '%s'", commentText), func(t *testing.T) {
		createCommentReq := &client.CreatePullRequestCommentRequest{
			Content: client.CreatePullRequestCommentContent{
				Raw: commentText,
			},
		}

		var err error
		comment, err = bb.CreatePullRequestComment(testWorkspace, repoSlug, prID, createCommentReq)
		require.NoError(t, err, "Failed to create comment")
		require.NotNil(t, comment, "Created comment should not be nil")
		require.Equal(t, commentText, comment.Content.Raw, "Comment text should match")
	})

	return comment
}

func stepVerifyListPullRequestCommits(t *testing.T, repoSlug string, prID int, expectedCommitMessages []string) {
	t.Helper()

	t.Run("verify list pull request commits", func(t *testing.T) {
		commits, err := bb.ListPullRequestCommits(testWorkspace, repoSlug, prID)
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

func stepVerifyGetPullRequestDiff(t *testing.T, repoSlug string, prID int, expectedSubstrings []string) {
	t.Helper()

	t.Run("verify get pull request diff", func(t *testing.T) {
		diff, err := bb.GetPullRequestDiff(testWorkspace, repoSlug, prID)
		require.NoError(t, err, "Failed to get pull request diff")
		require.NotNil(t, diff, "Diff should not be nil")
		require.NotEmpty(t, *diff, "Diff content should not be empty")

		// Verify each expected substring appears in the diff
		for _, substring := range expectedSubstrings {
			require.Contains(t, *diff, substring, "Diff should contain '%s'", substring)
		}
	})
}

func stepVerifyListPullRequestComments(t *testing.T, repoSlug string, prID int, pagelen int, page int, expectedCount int, expectedCommentTexts []string) {
	t.Helper()

	t.Run(fmt.Sprintf("verify list pull request comments (page %d, pagelen %d)", page, pagelen), func(t *testing.T) {
		comments, err := bb.ListPullRequestComments(testWorkspace, repoSlug, prID, pagelen, page)
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

func stepMergePullRequest(t *testing.T, repoSlug string, prID int) {
	t.Helper()

	t.Run(fmt.Sprintf("merge pull request #%d", prID), func(t *testing.T) {
		mergeReq := &client.MergePullRequestRequest{
			Type:    "merge_commit",
			Message: fmt.Sprintf("Merge pull request #%d", prID),
		}

		pr, err := bb.MergePullRequest(testWorkspace, repoSlug, prID, mergeReq)
		require.NoError(t, err, "Failed to merge pull request")
		require.NotNil(t, pr, "Merged pull request should not be nil")
		require.Equal(t, "MERGED", pr.State, "Pull request state should be MERGED")
	})
}

func stepDeclinePullRequest(t *testing.T, repoSlug string, prID int) {
	t.Helper()

	t.Run(fmt.Sprintf("decline pull request #%d", prID), func(t *testing.T) {
		pr, err := bb.DeclinePullRequest(testWorkspace, repoSlug, prID)
		require.NoError(t, err, "Failed to decline pull request")
		require.NotNil(t, pr, "Declined pull request should not be nil")
		require.Equal(t, "DECLINED", pr.State, "Pull request state should be DECLINED")
	})
}

func stepVerifyListPullRequestsByState(t *testing.T, repoSlug string, states []string, expectedPRIDs []int, unexpectedPRIDs []int) {
	t.Helper()

	stateStr := states[0]
	t.Run(fmt.Sprintf("verify list pull requests with state filter [%s]", stateStr), func(t *testing.T) {
		prList, err := bb.ListPullRequests(testWorkspace, repoSlug, 50, 1, states)
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

type itConfig struct{}

func (c *itConfig) BitbucketTestWorkspace() string {
	return config.GetString("BITBUCKET_TEST_WORKSPACE", "")
}

func (c *itConfig) BitbucketTestProjectKey() string {
	return config.GetString("BITBUCKET_TEST_PROJECT_KEY", "")
}

func (c *itConfig) BitbucketUrl() string {
	return config.GetString("BITBUCKET_URL", "")
}

func (c *itConfig) BitbucketEmail() string {
	return config.GetString("BITBUCKET_EMAIL", "")
}

func (c *itConfig) BitbucketApiToken() string {
	return config.GetString("BITBUCKET_API_TOKEN", "")
}

func (c *itConfig) BitbucketTimeout() int {
	return config.GetInt("BITBUCKET_TIMEOUT", 5)
}
