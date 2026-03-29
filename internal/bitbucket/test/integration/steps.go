package bitbucket_integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func StepCreateRepository(t *testing.T, bb *bitbucket.ApiClient, workspace, testProjectKey, testName string, isPrivate bool) (string, *bitbucket.ApiRepository, string) {
	t.Helper()
	var repo *bitbucket.ApiRepository
	var mainBranch string

	repoSlug := generateRepoSlug(testName)
	description := fmt.Sprintf("Test repository for %s", testName)

	t.Run("create repository", func(t *testing.T) {
		createReq := &bitbucket.ApiCreateRepositoryRequest{
			SCM:         "git",
			IsPrivate:   &isPrivate,
			Description: description,
			Project: &bitbucket.ApiCreateRepositoryProjectRef{
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

func StepDeleteRepository(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string) {
	t.Helper()
	if err := bb.DeleteRepository(context.Background(), workspace, repoSlug); err != nil {
		t.Logf("Warning: Failed to delete repository %s: %v", repoSlug, err)
	}
}

func StepCreateFiles(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, files map[string]string, message string, branch string, parents string) {
	t.Helper()

	StepName := "create files on main branch"
	if branch != "" {
		StepName = fmt.Sprintf("create files on %s branch", branch)
	}

	t.Run(StepName, func(t *testing.T) {
		createFilesReq := &bitbucket.ApiCreateFilesRequest{
			Branch:  branch,
			Message: message,
			Files:   files,
			Parents: parents,
		}

		err := bb.CreateOrUpdateFiles(context.Background(), workspace, repoSlug, createFilesReq)
		require.NoError(t, err, "Failed to create files")
	})
}

func StepGetLatestCommitHash(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string) string {
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

func StepCreateBranch(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, branchName string, commitHash string) *bitbucket.ApiBranch {
	t.Helper()
	var branch *bitbucket.ApiBranch

	t.Run(fmt.Sprintf("create branch '%s'", branchName), func(t *testing.T) {
		createBranchReq := &bitbucket.ApiCreateBranchRequest{
			Name: branchName,
			Target: bitbucket.ApiCreateBranchTarget{
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

func StepCreatePullRequest(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, title string, description string, sourceBranch string, destBranch string) *bitbucket.ApiPullRequest {
	t.Helper()
	var pr *bitbucket.ApiPullRequest

	t.Run(fmt.Sprintf("create pull request '%s'", title), func(t *testing.T) {
		createPRReq := &bitbucket.ApiCreatePullRequestRequest{
			Title:       title,
			Description: description,
			Source: bitbucket.ApiCreatePullRequestBranch{
				Branch: bitbucket.ApiCreatePullRequestBranchName{
					Name: sourceBranch,
				},
			},
			Destination: &bitbucket.ApiCreatePullRequestBranch{
				Branch: bitbucket.ApiCreatePullRequestBranchName{
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

func StepVerifyGetRepository(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, expectedRepo *bitbucket.ApiRepository) {
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

func StepVerifyListRepositories(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, expectedRepo *bitbucket.ApiRepository) {
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

func StepVerifySourceTree(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, expectedPaths []string) {
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

func StepVerifyDirectorySource(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, mainBranch string, dirPath string, expectedPaths []string) {
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

func StepVerifyFileContent(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, mainBranch string, filePath string, expectedSubstrings []string) {
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

func StepVerifyGetPullRequest(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, prID int, expectedTitle string, expectedSourceBranch string) {
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

func StepVerifyListPullRequests(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, prID int, expectedTitle string) {
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

func StepCreatePullRequestComment(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, prID int, commentText string) *bitbucket.ApiPullRequestComment {
	t.Helper()
	var comment *bitbucket.ApiPullRequestComment

	t.Run(fmt.Sprintf("create comment '%s'", commentText), func(t *testing.T) {
		createCommentReq := &bitbucket.ApiCreatePullRequestCommentRequest{
			Content: bitbucket.ApiCreatePullRequestCommentContent{
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

func StepVerifyListPullRequestCommits(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, prID int, expectedCommitMessages []string) {
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

func StepVerifyGetPullRequestDiff(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, prID int, expectedSubstrings []string) {
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

func StepVerifyListPullRequestComments(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, prID int, pagelen int, page int, expectedCount int, expectedCommentTexts []string) {
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

func StepMergePullRequest(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, prID int) {
	t.Helper()

	t.Run(fmt.Sprintf("merge pull request #%d", prID), func(t *testing.T) {
		mergeReq := &bitbucket.ApiMergePullRequestRequest{
			Type:    "merge_commit",
			Message: fmt.Sprintf("Merge pull request #%d", prID),
		}

		pr, err := bb.MergePullRequest(context.Background(), workspace, repoSlug, prID, mergeReq)
		require.NoError(t, err, "Failed to merge pull request")
		require.NotNil(t, pr, "Merged pull request should not be nil")
		require.Equal(t, "MERGED", pr.State, "Pull request state should be MERGED")
	})
}

func StepDeclinePullRequest(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, prID int) {
	t.Helper()

	t.Run(fmt.Sprintf("decline pull request #%d", prID), func(t *testing.T) {
		pr, err := bb.DeclinePullRequest(context.Background(), workspace, repoSlug, prID)
		require.NoError(t, err, "Failed to decline pull request")
		require.NotNil(t, pr, "Declined pull request should not be nil")
		require.Equal(t, "DECLINED", pr.State, "Pull request state should be DECLINED")
	})
}

func StepVerifyListPullRequestsByState(t *testing.T, bb *bitbucket.ApiClient, workspace, repoSlug string, states []string, expectedPRIDs []int, unexpectedPRIDs []int) {
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

// StepCreateCloneDir creates a temporary directory for a clone target and registers cleanup.
func StepCreateCloneDir(t *testing.T) string {
	t.Helper()
	targetDir, err := os.MkdirTemp("", "mcp-bitbucket-clone-*")
	require.NoError(t, err, "failed to create temp clone directory")
	t.Cleanup(func() { os.RemoveAll(targetDir) })
	return targetDir
}

// StepCloneRepository clones repoSlug into targetDir and verifies the .git directory exists.
func StepCloneRepository(t *testing.T, git *bitbucket.GitClient, workspace, repoSlug, targetDir string, depth int, ref string) {
	t.Helper()

	StepName := fmt.Sprintf("clone %s/%s", workspace, repoSlug)
	if depth > 0 {
		StepName += fmt.Sprintf(" depth=%d", depth)
	}
	if ref != "" {
		StepName += " ref=" + ref
	}

	t.Run(StepName, func(t *testing.T) {
		absPath, err := git.Clone(context.Background(), workspace, repoSlug, targetDir, depth, ref)
		require.NoError(t, err, "Clone should succeed")
		require.NotEmpty(t, absPath)

		_, statErr := os.Stat(filepath.Join(absPath, ".git"))
		require.NoError(t, statErr, ".git directory should exist after clone")
	})
}

// StepCloneFails clones repoSlug into targetDir and verifies the operation returns an error.
func StepCloneFails(t *testing.T, git *bitbucket.GitClient, workspace, repoSlug, targetDir string, depth int, ref string) {
	t.Helper()
	t.Run(fmt.Sprintf("clone %s/%s should fail", workspace, repoSlug), func(t *testing.T) {
		_, err := git.Clone(context.Background(), workspace, repoSlug, targetDir, depth, ref)
		require.Error(t, err, "Clone should fail")
	})
}

// StepVerifyCommitCount opens the git repository at repoDir and asserts exactly expectedCount
// commits are visible in its history.
func StepVerifyCommitCount(t *testing.T, repoDir string, expectedCount int) {
	t.Helper()
	t.Run(fmt.Sprintf("verify commit count = %d", expectedCount), func(t *testing.T) {
		repo, err := gogit.PlainOpen(repoDir)
		require.NoError(t, err, "should be able to open cloned repo")

		iter, err := repo.Log(&gogit.LogOptions{})
		require.NoError(t, err)

		count := 0
		err = iter.ForEach(func(_ *object.Commit) error {
			count++
			return nil
		})
		// Shallow clones hit ErrObjectNotFound at the shallow boundary — treat it as end of history.
		if err != nil && !errors.Is(err, plumbing.ErrObjectNotFound) {
			require.NoError(t, err, "unexpected error iterating commits")
		}
		require.Equal(t, expectedCount, count, "unexpected commit count in cloned repo")
	})
}

// StepVerifyFileExists asserts that file (relative to repoDir) exists on disk.
func StepVerifyFileExists(t *testing.T, repoDir, file string) {
	t.Helper()
	t.Run(fmt.Sprintf("verify file %s exists", file), func(t *testing.T) {
		_, err := os.Stat(filepath.Join(repoDir, file))
		require.NoError(t, err, "expected file %q to be present in cloned repo", file)
	})
}

// StepPullRepository pulls latest changes into repoDir and verifies the operation succeeds.
func StepPullRepository(t *testing.T, git *bitbucket.GitClient, repoDir, remoteURL, ref string) {
	t.Helper()
	stepName := "pull " + remoteURL
	if ref != "" {
		stepName += " ref=" + ref
	}
	t.Run(stepName, func(t *testing.T) {
		absPath, err := git.Pull(context.Background(), repoDir, remoteURL, ref)
		require.NoError(t, err, "Pull should succeed")
		require.NotEmpty(t, absPath)
	})
}

// StepPullFails attempts a pull and verifies the operation returns an error.
func StepPullFails(t *testing.T, git *bitbucket.GitClient, repoDir, remoteURL, ref string) {
	t.Helper()
	t.Run("pull "+remoteURL+" should fail", func(t *testing.T) {
		_, err := git.Pull(context.Background(), repoDir, remoteURL, ref)
		require.Error(t, err, "Pull should fail")
	})
}

func generateRepoSlug(testName string) string {
	return fmt.Sprintf("%s-%d", testName, time.Now().UnixNano())
}
