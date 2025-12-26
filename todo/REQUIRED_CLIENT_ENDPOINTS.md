# Required Client Endpoints for Integration Tests

## Current Endpoints (Already Implemented)
- `ListRepositories(workspace, pagelen, page)` - GET /repositories/{workspace}
- `GetRepository(workspace, repoSlug)` - GET /repositories/{workspace}/{repo}
- `GetRepositorySource(workspace, repoSlug)` - GET /repositories/{workspace}/{repo}/src
- `ListPullRequests(workspace, repoSlug, pagelen, page, states)` - GET /repositories/{workspace}/{repo}/pullrequests
- `GetPullRequest(workspace, repoSlug, prId)` - GET /repositories/{workspace}/{repo}/pullrequests/{prId}
- `ListPullRequestCommits(workspace, repoSlug, prId)` - GET /repositories/{workspace}/{repo}/pullrequests/{prId}/commits
- `ListPullRequestComments(workspace, repoSlug, prId, pagelen, page)` - GET /repositories/{workspace}/{repo}/pullrequests/{prId}/comments
- `GetPullRequestDiff(workspace, repoSlug, prId)` - GET /repositories/{workspace}/{repo}/pullrequests/{prId}/diff
- `GetFileSource(workspace, repoSlug, commit, path)` - GET /repositories/{workspace}/{repo}/src/{commit}/{path}
- `GetDirectorySource(workspace, repoSlug, commit, path)` - GET /repositories/{workspace}/{repo}/src/{commit}/{path}

## New Endpoints Required

### Repository Management
**Needed for**: Scenarios 1-17 (creating/deleting test repositories)

1. `CreateRepository(workspace, repoSlug, isPrivate, description)` - POST /repositories/{workspace}/{repo}
   - Request body: `{ "scm": "git", "is_private": bool, "description": string }`
   - Returns: `*Repository`

2. `DeleteRepository(workspace, repoSlug)` - DELETE /repositories/{workspace}/{repo}
   - Returns: `error` (204 on success)

### File/Content Management
**Needed for**: Scenarios 2-4, 5-9, 10, 11, 14-17 (creating files and commits)

3. `CreateOrUpdateFile(workspace, repoSlug, branch, filePath, content, message, parentCommit)` - POST /repositories/{workspace}/{repo}/src
   - Request body: Form data with file content and commit message
   - Can create multiple files in a single commit
   - Returns: `error` (or commit hash)

   Alternative approach using lower-level API:
   - Use the `/src` endpoint with POST and form data
   - Need to handle multipart/form-data encoding

### Branch Management
**Needed for**: Scenarios 5-9, 11 (creating feature branches)

4. `CreateBranch(workspace, repoSlug, branchName, target)` - POST /repositories/{workspace}/{repo}/refs/branches
   - Request body: `{ "name": string, "target": { "hash": string } }`
   - Returns: `*Branch` (need to define Branch type)

### Pull Request Management
**Needed for**: Scenarios 5-9, 11, 14, 17 (creating, merging, declining PRs)

5. `CreatePullRequest(workspace, repoSlug, title, description, sourceBranch, destBranch)` - POST /repositories/{workspace}/{repo}/pullrequests
   - Request body: `{ "title": string, "description": string, "source": { "branch": { "name": string } }, "destination": { "branch": { "name": string } } }`
   - Returns: `*PullRequest`

6. `MergePullRequest(workspace, repoSlug, prId)` - POST /repositories/{workspace}/{repo}/pullrequests/{prId}/merge
   - Request body: `{ "type": "merge", "message": string }` (optional)
   - Returns: `*PullRequest`

7. `DeclinePullRequest(workspace, repoSlug, prId)` - POST /repositories/{workspace}/{repo}/pullrequests/{prId}/decline
   - Returns: `*PullRequest`

### Pull Request Comment Management
**Needed for**: Scenarios 8, 11 (creating comments)

8. `CreatePullRequestComment(workspace, repoSlug, prId, content)` - POST /repositories/{workspace}/{repo}/pullrequests/{prId}/comments
   - Request body: `{ "content": { "raw": string } }`
   - Returns: `*PullRequestComment`

## Summary

**Total New Endpoints**: 8

**Breakdown by Category**:
- Repository management: 2 (Create, Delete)
- File/Content management: 1 (CreateOrUpdateFile)
- Branch management: 1 (CreateBranch)
- Pull Request management: 3 (Create, Merge, Decline)
- Comment management: 1 (Create)

## Implementation Priority

1. **High Priority** (Required for all test scenarios):
   - `CreateRepository`
   - `DeleteRepository`
   - `CreateOrUpdateFile`

2. **Medium Priority** (Required for PR-related tests):
   - `CreateBranch`
   - `CreatePullRequest`

3. **Low Priority** (Required for specific PR state tests):
   - `MergePullRequest`
   - `DeclinePullRequest`
   - `CreatePullRequestComment`

## Notes

- The Bitbucket API uses POST for creating resources and DELETE for deleting
- Some endpoints require specific request body formats (JSON vs form-data)
- Creating files in Bitbucket can be done via the `/src` endpoint with POST and form-data
- Branch creation might alternatively be handled implicitly by creating commits on a new branch name
- Error scenarios (12-17) don't require new endpoints, they test existing endpoints with invalid data
