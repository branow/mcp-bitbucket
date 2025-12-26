# Integration Test Scenarios

## Scenario 1: Repository Operations
**Step 1 (Create)**: Create a new repository with name, description, and privacy settings
**Step 2 (Check)**:
- Call `GetRepository` and verify repo details match
- Call `ListRepositories` and verify repo appears in the list
**Step 3 (Delete)**: Delete the created repository

## Scenario 2: Repository Source Tree
**Step 1 (Create)**:
- Create a repository
- Create multiple files in nested directory structure (e.g., `/src/main/file.go`, `/docs/README.md`)
**Step 2 (Check)**: Call `GetRepositorySource` and verify the directory tree structure
**Step 3 (Delete)**: Delete the repository

## Scenario 3: File Source
**Step 1 (Create)**:
- Create a repository
- Create a file with specific content on main branch
**Step 2 (Check)**: Call `GetFileSource` and verify the file content matches
**Step 3 (Delete)**: Delete the repository

## Scenario 4: Directory Source
**Step 1 (Create)**:
- Create a repository
- Create multiple files within a specific directory (e.g., `/src/utils/file1.go`, `/src/utils/file2.go`)
**Step 2 (Check)**: Call `GetDirectorySource` on `/src/utils/` and verify all files appear
**Step 3 (Delete)**: Delete the repository

## Scenario 5: Pull Request Basics
**Step 1 (Create)**:
- Create a repository with a file on main branch
- Create a feature branch
- Modify/add file on feature branch
- Create a pull request from feature to main
**Step 2 (Check)**:
- Call `GetPullRequest` and verify PR details
- Call `ListPullRequests` and verify PR appears
**Step 3 (Delete)**: Delete the repository

## Scenario 6: Pull Request Commits
**Step 1 (Create)**:
- Create a repository with pull request (from Scenario 5)
- Ensure feature branch has 2-3 commits
**Step 2 (Check)**: Call `ListPullRequestCommits` and verify all commits appear
**Step 3 (Delete)**: Delete the repository

## Scenario 7: Pull Request Diff
**Step 1 (Create)**:
- Create a repository with pull request (from Scenario 5)
- Feature branch has file changes
**Step 2 (Check)**: Call `GetPullRequestDiff` and verify diff contains expected changes
**Step 3 (Delete)**: Delete the repository

## Scenario 8: Pull Request Comments
**Step 1 (Create)**:
- Create a repository with pull request (from Scenario 5)
- Add 3-5 comments to the pull request
**Step 2 (Check)**: Call `ListPullRequestComments` with pagination and verify all comments appear
**Step 3 (Delete)**: Delete the repository

## Scenario 9: Pull Request States
**Step 1 (Create)**:
- Create a repository
- Create 3 pull requests: one OPEN, one MERGED, one DECLINED
**Step 2 (Check)**:
- Call `ListPullRequests` with state filter `["OPEN"]` and verify only open PR appears
- Call `ListPullRequests` with state filter `["MERGED"]` and verify only merged PR appears
- Call `ListPullRequests` with state filter `["DECLINED"]` and verify only declined PR appears
**Step 3 (Delete)**: Delete the repository

## Scenario 10: Pagination - Repositories
**Step 1 (Create)**: Create 15 repositories in the test workspace
**Step 2 (Check)**:
- Call `ListRepositories` with pagelen=5, page=1 and verify 5 repos returned
- Call `ListRepositories` with pagelen=5, page=2 and verify next 5 repos returned
- Verify pagination metadata (page, size, next link)
**Step 3 (Delete)**: Delete all 15 repositories

## Scenario 11: Pagination - Pull Request Comments
**Step 1 (Create)**:
- Create a repository with pull request
- Add 12 comments to the pull request
**Step 2 (Check)**:
- Call `ListPullRequestComments` with pagelen=5, page=1 and verify 5 comments returned
- Call `ListPullRequestComments` with pagelen=5, page=2 and verify next 5 comments returned
- Call `ListPullRequestComments` with pagelen=5, page=3 and verify remaining 2 comments returned
**Step 3 (Delete)**: Delete the repository

## Scenario 12: Error - Non-existent Workspace
**Step 1 (Create)**: Nothing
**Step 2 (Check)**:
- Call `ListRepositories` with non-existent workspace slug and verify error (404)
- Call `GetRepository` with non-existent workspace slug and verify error (404)
**Step 3 (Delete)**: Nothing

## Scenario 13: Error - Non-existent Repository
**Step 1 (Create)**: Nothing
**Step 2 (Check)**:
- Call `GetRepository` with non-existent repository slug and verify error (404)
- Call `GetRepositorySource` with non-existent repository slug and verify error (404)
- Call `ListPullRequests` with non-existent repository slug and verify error (404)
**Step 3 (Delete)**: Nothing

## Scenario 14: Error - Non-existent Pull Request
**Step 1 (Create)**: Create a repository (without any pull requests)
**Step 2 (Check)**:
- Call `GetPullRequest` with non-existent PR ID (e.g., 99999) and verify error (404)
- Call `ListPullRequestCommits` with non-existent PR ID and verify error (404)
- Call `ListPullRequestComments` with non-existent PR ID and verify error (404)
- Call `GetPullRequestDiff` with non-existent PR ID and verify error (404)
**Step 3 (Delete)**: Delete the repository

## Scenario 15: Error - Non-existent File/Directory
**Step 1 (Create)**: Create a repository with some files
**Step 2 (Check)**:
- Call `GetFileSource` with non-existent file path and verify error (404)
- Call `GetDirectorySource` with non-existent directory path and verify error (404)
**Step 3 (Delete)**: Delete the repository

## Scenario 16: Error - Invalid Pagination Parameters
**Step 1 (Create)**: Create a repository
**Step 2 (Check)**:
- Call `ListRepositories` with pagelen=0 and verify error (400)
- Call `ListRepositories` with page=0 and verify error (400)
- Call `ListRepositories` with pagelen=-1 and verify error (400)
- Call `ListPullRequestComments` with pagelen=0 and verify error (400)
**Step 3 (Delete)**: Delete the repository

## Scenario 17: Error - Invalid Pull Request State Filter
**Step 1 (Create)**: Create a repository with pull requests
**Step 2 (Check)**: Call `ListPullRequests` with invalid state (e.g., `["INVALID_STATE"]`) and verify error (400)
**Step 3 (Delete)**: Delete the repository
