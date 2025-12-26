// Package client defines types for the Bitbucket API client.
//
// All types in this file represent Bitbucket API request and response structures.
// These types closely mirror the JSON structures returned by the Bitbucket REST API v2.0.
// For detailed field descriptions, refer to the official Bitbucket API documentation at:
// https://developer.atlassian.com/cloud/bitbucket/rest/
package client

type ErrorResponse struct {
	Type  string    `json:"type"`
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
	Detail  string              `json:"detail,omitempty"`
	ID      string              `json:"id,omitempty"`
	Data    map[string]any      `json:"data,omitempty"`
}

type ApiResponse[T any] struct {
	Values   []T     `json:"values"`
	Pagelen  int     `json:"pagelen"`
	Size     *int    `json:"size,omitempty"`
	Page     *int    `json:"page,omitempty"`
	Next     *string `json:"next,omitempty"`
	Previous *string `json:"previous,omitempty"`
}

type Repository struct {
	Type                  string            `json:"type"`
	FullName              string            `json:"full_name"`
	Links                 RepositoryLinks   `json:"links"`
	Name                  string            `json:"name"`
	Slug                  string            `json:"slug"`
	Description           string            `json:"description"`
	SCM                   string            `json:"scm"`
	Website               *string           `json:"website"`
	Owner                 Owner             `json:"owner"`
	Workspace             Workspace         `json:"workspace"`
	IsPrivate             bool              `json:"is_private"`
	Project               Project           `json:"project"`
	ForkPolicy            string            `json:"fork_policy"`
	CreatedOn             string            `json:"created_on"`
	UpdatedOn             string            `json:"updated_on"`
	Size                  int               `json:"size"`
	Language              string            `json:"language"`
	UUID                  string            `json:"uuid"`
	MainBranch            MainBranch        `json:"mainbranch"`
	OverrideSettings      OverrideSettings  `json:"override_settings"`
	Parent                *ParentRepository `json:"parent"`
	EnforcedSignedCommits *bool             `json:"enforced_signed_commits"`
	HasIssues             bool              `json:"has_issues"`
	HasWiki               bool              `json:"has_wiki"`
}

type RepositoryLinks struct {
	Self         Link        `json:"self"`
	HTML         Link        `json:"html"`
	Avatar       Link        `json:"avatar"`
	Pullrequests Link        `json:"pullrequests"`
	Commits      Link        `json:"commits"`
	Forks        Link        `json:"forks"`
	Watchers     Link        `json:"watchers"`
	Branches     Link        `json:"branches"`
	Tags         Link        `json:"tags"`
	Downloads    Link        `json:"downloads"`
	Source       Link        `json:"source"`
	Clone        []CloneLink `json:"clone"`
	Hooks        Link        `json:"hooks"`
}

type Link struct {
	Href string  `json:"href"`
	Name *string `json:"name,omitempty"`
}

type CloneLink struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type CommonLinks struct {
	Self   Link `json:"self"`
	HTML   Link `json:"html"`
	Avatar Link `json:"avatar"`
}

type Owner struct {
	DisplayName string      `json:"display_name"`
	Links       CommonLinks `json:"links"`
	Type        string      `json:"type"`
	UUID        string      `json:"uuid"`
	Username    string      `json:"username"`
}

type Workspace struct {
	Type  string      `json:"type"`
	UUID  string      `json:"uuid"`
	Name  string      `json:"name"`
	Slug  string      `json:"slug"`
	Links CommonLinks `json:"links"`
}

type Project struct {
	Type  string      `json:"type"`
	Key   string      `json:"key"`
	UUID  string      `json:"uuid"`
	Name  string      `json:"name"`
	Links CommonLinks `json:"links"`
}

type MainBranch struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type OverrideSettings struct {
	DefaultMergeStrategy bool `json:"default_merge_strategy"`
	BranchingModel       bool `json:"branching_model"`
}

type ParentRepository struct {
	Type     string      `json:"type"`
	FullName string      `json:"full_name"`
	Links    CommonLinks `json:"links"`
	Name     string      `json:"name"`
	UUID     string      `json:"uuid"`
}

type SourceItem struct {
	Path        string          `json:"path"`
	Type        string          `json:"type"`
	Commit      SourceCommit    `json:"commit"`
	Links       SourceItemLinks `json:"links"`
	EscapedPath *string         `json:"escaped_path,omitempty"`
	Size        *int            `json:"size,omitempty"`
	Mimetype    *string         `json:"mimetype"`
	Attributes  []string        `json:"attributes,omitempty"`
}

type SourceCommit struct {
	Hash  string            `json:"hash"`
	Type  string            `json:"type"`
	Links SourceCommitLinks `json:"links"`
}

type SourceCommitLinks struct {
	Self Link `json:"self"`
	HTML Link `json:"html"`
}

type SourceItemLinks struct {
	Self    Link  `json:"self"`
	Meta    Link  `json:"meta"`
	History *Link `json:"history,omitempty"`
}

type PullRequest struct {
	CommentCount      int                      `json:"comment_count"`
	TaskCount         int                      `json:"task_count"`
	Type              string                   `json:"type"`
	ID                int                      `json:"id"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description"`
	Rendered          *PullRequestRendered     `json:"rendered,omitempty"`
	State             string                   `json:"state"`
	Draft             bool                     `json:"draft"`
	MergeCommit       *PullRequestCommit       `json:"merge_commit"`
	CloseSourceBranch bool                     `json:"close_source_branch"`
	ClosedBy          *User                    `json:"closed_by"`
	Author            User                     `json:"author"`
	Reason            *string                  `json:"reason,omitempty"`
	CreatedOn         string                   `json:"created_on"`
	UpdatedOn         string                   `json:"updated_on"`
	Destination       PullRequestBranch        `json:"destination"`
	Source            PullRequestBranch        `json:"source"`
	Reviewers         []User                   `json:"reviewers,omitempty"`
	Participants      []PullRequestParticipant `json:"participants,omitempty"`
	Links             PullRequestLinks         `json:"links"`
	Summary           PullRequestSummary       `json:"summary"`
}

type User struct {
	DisplayName string      `json:"display_name"`
	Links       CommonLinks `json:"links"`
	Type        string      `json:"type"`
	UUID        string      `json:"uuid"`
	AccountID   string      `json:"account_id"`
	Nickname    string      `json:"nickname"`
	Username    string      `json:"username"`
}

type PullRequestCommit struct {
	Hash  string      `json:"hash"`
	Links CommonLinks `json:"links"`
	Type  string      `json:"type"`
}

type PullRequestBranch struct {
	Branch     BranchInfo            `json:"branch"`
	Commit     PullRequestCommit     `json:"commit"`
	Repository PullRequestRepository `json:"repository"`
}

type BranchInfo struct {
	Name           string         `json:"name"`
	Links          map[string]any `json:"links"`
	SyncStrategies []string       `json:"sync_strategies,omitempty"`
}

type PullRequestRepository struct {
	Type     string      `json:"type"`
	FullName string      `json:"full_name"`
	Links    CommonLinks `json:"links"`
	Name     string      `json:"name"`
	UUID     string      `json:"uuid"`
}

type PullRequestLinks struct {
	Self           Link `json:"self"`
	HTML           Link `json:"html"`
	Commits        Link `json:"commits"`
	Approve        Link `json:"approve"`
	RequestChanges Link `json:"request-changes"`
	Diff           Link `json:"diff"`
	Diffstat       Link `json:"diffstat"`
	Comments       Link `json:"comments"`
	Activity       Link `json:"activity"`
	Merge          Link `json:"merge"`
	Decline        Link `json:"decline"`
	Statuses       Link `json:"statuses"`
}

type PullRequestSummary struct {
	Type   string `json:"type"`
	Raw    string `json:"raw"`
	Markup string `json:"markup"`
	HTML   string `json:"html"`
}

type PullRequestRendered struct {
	Title       PullRequestSummary  `json:"title"`
	Description PullRequestSummary  `json:"description"`
	Reason      *PullRequestSummary `json:"reason,omitempty"`
}

type PullRequestParticipant struct {
	Type           string  `json:"type"`
	User           User    `json:"user"`
	Role           string  `json:"role"`
	Approved       bool    `json:"approved"`
	State          *string `json:"state"`
	ParticipatedOn *string `json:"participated_on"`
}

type Commit struct {
	Type       string             `json:"type"`
	Hash       string             `json:"hash"`
	Date       string             `json:"date"`
	Author     CommitAuthor       `json:"author"`
	Message    string             `json:"message"`
	Summary    PullRequestSummary `json:"summary"`
	Links      CommitLinks        `json:"links"`
	Parents    []CommitParent     `json:"parents"`
	Repository CommitRepository   `json:"repository"`
}

type CommitAuthor struct {
	Type string           `json:"type"`
	Raw  string           `json:"raw"`
	User CommitAuthorUser `json:"user"`
}

type CommitAuthorUser struct {
	DisplayName string      `json:"display_name"`
	Links       CommonLinks `json:"links"`
	Type        string      `json:"type"`
	UUID        string      `json:"uuid"`
	AccountID   string      `json:"account_id"`
	Nickname    string      `json:"nickname"`
}

type CommitParent struct {
	Hash  string            `json:"hash"`
	Links SourceCommitLinks `json:"links"`
	Type  string            `json:"type"`
}

type CommitLinks struct {
	Self     Link `json:"self"`
	HTML     Link `json:"html"`
	Diff     Link `json:"diff"`
	Approve  Link `json:"approve"`
	Comments Link `json:"comments"`
	Statuses Link `json:"statuses"`
	Patch    Link `json:"patch"`
}

type CommitRepository struct {
	Type     string      `json:"type"`
	FullName string      `json:"full_name"`
	Links    CommonLinks `json:"links"`
	Name     string      `json:"name"`
	UUID     string      `json:"uuid"`
}

type PullRequestComment struct {
	ID          int                           `json:"id"`
	CreatedOn   string                        `json:"created_on"`
	UpdatedOn   string                        `json:"updated_on"`
	Content     PullRequestCommentContent     `json:"content"`
	User        PullRequestCommentUser        `json:"user"`
	Deleted     bool                          `json:"deleted"`
	Inline      *PullRequestCommentInline     `json:"inline,omitempty"`
	Parent      *PullRequestCommentParent     `json:"parent,omitempty"`
	Pending     bool                          `json:"pending"`
	Type        string                        `json:"type"`
	Links       PullRequestCommentLinks       `json:"links"`
	PullRequest PullRequestCommentPullRequest `json:"pullrequest"`
	Resolution  *PullRequestCommentResolution `json:"resolution,omitempty"`
}

type PullRequestCommentContent struct {
	Type   string `json:"type"`
	Raw    string `json:"raw"`
	Markup string `json:"markup"`
	HTML   string `json:"html"`
}

type PullRequestCommentUser struct {
	DisplayName string      `json:"display_name"`
	Links       CommonLinks `json:"links"`
	Type        string      `json:"type"`
	UUID        string      `json:"uuid"`
	Username    *string     `json:"username,omitempty"`
	AccountID   *string     `json:"account_id,omitempty"`
	Nickname    *string     `json:"nickname,omitempty"`
}

type PullRequestCommentInline struct {
	From      *int   `json:"from"`
	To        *int   `json:"to"`
	Path      string `json:"path"`
	StartFrom *int   `json:"start_from"`
	StartTo   *int   `json:"start_to"`
}

type PullRequestCommentLinks struct {
	Self Link  `json:"self"`
	HTML Link  `json:"html"`
	Code *Link `json:"code,omitempty"`
}

type PullRequestCommentPullRequest struct {
	Type  string  `json:"type"`
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Draft bool    `json:"draft"`
	Links PRLinks `json:"links"`
}

type PRLinks struct {
	Self Link `json:"self"`
	HTML Link `json:"html"`
}

type PullRequestCommentParent struct {
	ID    int                           `json:"id"`
	Links PullRequestCommentParentLinks `json:"links"`
}

type PullRequestCommentParentLinks struct {
	Self Link `json:"self"`
	HTML Link `json:"html"`
}

type PullRequestCommentResolution struct {
	Type *string `json:"type,omitempty"`
}

type CreateRepositoryRequest struct {
	SCM         string                      `json:"scm"`
	IsPrivate   *bool                       `json:"is_private,omitempty"`
	Description string                      `json:"description,omitempty"`
	ForkPolicy  string                      `json:"fork_policy,omitempty"`
	Language    string                      `json:"language,omitempty"`
	HasIssues   *bool                       `json:"has_issues,omitempty"`
	HasWiki     *bool                       `json:"has_wiki,omitempty"`
	Project     *CreateRepositoryProjectRef `json:"project,omitempty"`
}

type CreateRepositoryProjectRef struct {
	Key string `json:"key,omitempty"`
}
