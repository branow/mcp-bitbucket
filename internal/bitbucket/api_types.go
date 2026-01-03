// All types in this file represent Bitbucket API request and response structures.
// These types closely mirror the JSON structures returned by the Bitbucket REST API v2.0.
// For detailed field descriptions, refer to the official Bitbucket API documentation at:
// https://developer.atlassian.com/cloud/bitbucket/rest/
package bitbucket

type ApiErrorResponse struct {
	Type  string       `json:"type"`
	Error ApiErrorBody `json:"error"`
}

type ApiErrorBody struct {
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

type ApiRepository struct {
	Type                  string               `json:"type"`
	FullName              string               `json:"full_name"`
	Links                 ApiRepositoryLinks   `json:"links"`
	Name                  string               `json:"name"`
	Slug                  string               `json:"slug"`
	Description           string               `json:"description"`
	SCM                   string               `json:"scm"`
	Website               *string              `json:"website"`
	Owner                 ApiOwner             `json:"owner"`
	Workspace             ApiWorkspace         `json:"workspace"`
	IsPrivate             bool                 `json:"is_private"`
	Project               ApiProject           `json:"project"`
	ForkPolicy            string               `json:"fork_policy"`
	CreatedOn             string               `json:"created_on"`
	UpdatedOn             string               `json:"updated_on"`
	Size                  int                  `json:"size"`
	Language              string               `json:"language"`
	UUID                  string               `json:"uuid"`
	MainBranch            ApiMainBranch        `json:"mainbranch"`
	OverrideSettings      ApiOverrideSettings  `json:"override_settings"`
	Parent                *ApiParentRepository `json:"parent"`
	EnforcedSignedCommits *bool                `json:"enforced_signed_commits"`
	HasIssues             bool                 `json:"has_issues"`
	HasWiki               bool                 `json:"has_wiki"`
}

type ApiRepositoryLinks struct {
	Self         ApiLink        `json:"self"`
	HTML         ApiLink        `json:"html"`
	Avatar       ApiLink        `json:"avatar"`
	Pullrequests ApiLink        `json:"pullrequests"`
	Commits      ApiLink        `json:"commits"`
	Forks        ApiLink        `json:"forks"`
	Watchers     ApiLink        `json:"watchers"`
	Branches     ApiLink        `json:"branches"`
	Tags         ApiLink        `json:"tags"`
	Downloads    ApiLink        `json:"downloads"`
	Source       ApiLink        `json:"source"`
	Clone        []ApiCloneLink `json:"clone"`
	Hooks        ApiLink        `json:"hooks"`
}

type ApiLink struct {
	Href string  `json:"href"`
	Name *string `json:"name,omitempty"`
}

type ApiCloneLink struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type ApiCommonLinks struct {
	Self   ApiLink `json:"self"`
	HTML   ApiLink `json:"html"`
	Avatar ApiLink `json:"avatar"`
}

type ApiOwner struct {
	DisplayName string         `json:"display_name"`
	Links       ApiCommonLinks `json:"links"`
	Type        string         `json:"type"`
	UUID        string         `json:"uuid"`
	Username    string         `json:"username"`
}

type ApiWorkspace struct {
	Type  string         `json:"type"`
	UUID  string         `json:"uuid"`
	Name  string         `json:"name"`
	Slug  string         `json:"slug"`
	Links ApiCommonLinks `json:"links"`
}

type ApiProject struct {
	Type  string         `json:"type"`
	Key   string         `json:"key"`
	UUID  string         `json:"uuid"`
	Name  string         `json:"name"`
	Links ApiCommonLinks `json:"links"`
}

type ApiMainBranch struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ApiOverrideSettings struct {
	DefaultMergeStrategy bool `json:"default_merge_strategy"`
	BranchingModel       bool `json:"branching_model"`
}

type ApiParentRepository struct {
	Type     string         `json:"type"`
	FullName string         `json:"full_name"`
	Links    ApiCommonLinks `json:"links"`
	Name     string         `json:"name"`
	UUID     string         `json:"uuid"`
}

type ApiSourceItem struct {
	Path        string             `json:"path"`
	Type        string             `json:"type"`
	Commit      ApiSourceCommit    `json:"commit"`
	Links       ApiSourceItemLinks `json:"links"`
	EscapedPath *string            `json:"escaped_path,omitempty"`
	Size        *int               `json:"size,omitempty"`
	Mimetype    *string            `json:"mimetype"`
	Attributes  []string           `json:"attributes,omitempty"`
}

type ApiSourceCommit struct {
	Hash  string               `json:"hash"`
	Type  string               `json:"type"`
	Links ApiSourceCommitLinks `json:"links"`
}

type ApiSourceCommitLinks struct {
	Self ApiLink `json:"self"`
	HTML ApiLink `json:"html"`
}

type ApiSourceItemLinks struct {
	Self    ApiLink  `json:"self"`
	Meta    ApiLink  `json:"meta"`
	History *ApiLink `json:"history,omitempty"`
}

type ApiPullRequest struct {
	CommentCount      int                         `json:"comment_count"`
	TaskCount         int                         `json:"task_count"`
	Type              string                      `json:"type"`
	ID                int                         `json:"id"`
	Title             string                      `json:"title"`
	Description       string                      `json:"description"`
	Rendered          *ApiPullRequestRendered     `json:"rendered,omitempty"`
	State             string                      `json:"state"`
	Draft             bool                        `json:"draft"`
	MergeCommit       *ApiPullRequestCommit       `json:"merge_commit"`
	CloseSourceBranch bool                        `json:"close_source_branch"`
	ClosedBy          *ApiUser                    `json:"closed_by"`
	Author            ApiUser                     `json:"author"`
	Reason            *string                     `json:"reason,omitempty"`
	CreatedOn         string                      `json:"created_on"`
	UpdatedOn         string                      `json:"updated_on"`
	ClosedOn          *string                     `json:"closed_on,omitempty"`
	Destination       ApiPullRequestBranch        `json:"destination"`
	Source            ApiPullRequestBranch        `json:"source"`
	Reviewers         []ApiUser                   `json:"reviewers,omitempty"`
	Participants      []ApiPullRequestParticipant `json:"participants,omitempty"`
	Links             ApiPullRequestLinks         `json:"links"`
	Summary           ApiPullRequestSummary       `json:"summary"`
}

type ApiUser struct {
	DisplayName string         `json:"display_name"`
	Links       ApiCommonLinks `json:"links"`
	Type        string         `json:"type"`
	UUID        string         `json:"uuid"`
	AccountID   string         `json:"account_id"`
	Nickname    string         `json:"nickname"`
	Username    string         `json:"username"`
}

type ApiPullRequestCommit struct {
	Hash       string                 `json:"hash"`
	Links      ApiCommonLinks         `json:"links"`
	Type       string                 `json:"type"`
	Date       *string                `json:"date,omitempty"`
	Author     *ApiCommitAuthor       `json:"author,omitempty"`
	Message    *string                `json:"message,omitempty"`
	Summary    *ApiPullRequestSummary `json:"summary,omitempty"`
	Parents    []ApiCommitParent      `json:"parents,omitempty"`
	Repository *ApiCommitRepository   `json:"repository,omitempty"`
}

type ApiPullRequestBranch struct {
	Branch     ApiBranchInfo            `json:"branch"`
	Commit     ApiPullRequestCommit     `json:"commit"`
	Repository ApiPullRequestRepository `json:"repository"`
}

type ApiBranchInfo struct {
	Name           string         `json:"name"`
	Links          map[string]any `json:"links"`
	SyncStrategies []string       `json:"sync_strategies,omitempty"`
}

type ApiPullRequestRepository struct {
	Type     string         `json:"type"`
	FullName string         `json:"full_name"`
	Links    ApiCommonLinks `json:"links"`
	Name     string         `json:"name"`
	UUID     string         `json:"uuid"`
}

type ApiPullRequestLinks struct {
	Self           ApiLink `json:"self"`
	HTML           ApiLink `json:"html"`
	Commits        ApiLink `json:"commits"`
	Approve        ApiLink `json:"approve"`
	RequestChanges ApiLink `json:"request-changes"`
	Diff           ApiLink `json:"diff"`
	Diffstat       ApiLink `json:"diffstat"`
	Comments       ApiLink `json:"comments"`
	Activity       ApiLink `json:"activity"`
	Merge          ApiLink `json:"merge"`
	Decline        ApiLink `json:"decline"`
	Statuses       ApiLink `json:"statuses"`
}

type ApiPullRequestSummary struct {
	Type   string `json:"type"`
	Raw    string `json:"raw"`
	Markup string `json:"markup"`
	HTML   string `json:"html"`
}

type ApiPullRequestRendered struct {
	Title       ApiPullRequestSummary  `json:"title"`
	Description ApiPullRequestSummary  `json:"description"`
	Reason      *ApiPullRequestSummary `json:"reason,omitempty"`
}

type ApiPullRequestParticipant struct {
	Type           string  `json:"type"`
	User           ApiUser `json:"user"`
	Role           string  `json:"role"`
	Approved       bool    `json:"approved"`
	State          *string `json:"state"`
	ParticipatedOn *string `json:"participated_on"`
}

type ApiCommit struct {
	Type       string                `json:"type"`
	Hash       string                `json:"hash"`
	Date       string                `json:"date"`
	Author     ApiCommitAuthor       `json:"author"`
	Message    string                `json:"message"`
	Summary    ApiPullRequestSummary `json:"summary"`
	Links      ApiCommitLinks        `json:"links"`
	Parents    []ApiCommitParent     `json:"parents"`
	Repository ApiCommitRepository   `json:"repository"`
}

type ApiCommitAuthor struct {
	Type string  `json:"type"`
	Raw  string  `json:"raw"`
	User ApiUser `json:"user"`
}

type ApiCommitParent struct {
	Hash  string               `json:"hash"`
	Links ApiSourceCommitLinks `json:"links"`
	Type  string               `json:"type"`
}

type ApiCommitLinks struct {
	Self     ApiLink `json:"self"`
	HTML     ApiLink `json:"html"`
	Diff     ApiLink `json:"diff"`
	Approve  ApiLink `json:"approve"`
	Comments ApiLink `json:"comments"`
	Statuses ApiLink `json:"statuses"`
	Patch    ApiLink `json:"patch"`
}

type ApiCommitRepository struct {
	Type     string         `json:"type"`
	FullName string         `json:"full_name"`
	Links    ApiCommonLinks `json:"links"`
	Name     string         `json:"name"`
	UUID     string         `json:"uuid"`
}

type ApiPullRequestComment struct {
	ID          int                              `json:"id"`
	CreatedOn   string                           `json:"created_on"`
	UpdatedOn   string                           `json:"updated_on"`
	Content     ApiPullRequestCommentContent     `json:"content"`
	User        ApiUser                          `json:"user"`
	Deleted     bool                             `json:"deleted"`
	Inline      *ApiPullRequestCommentInline     `json:"inline,omitempty"`
	Parent      *ApiPullRequestCommentParent     `json:"parent,omitempty"`
	Pending     bool                             `json:"pending"`
	Type        string                           `json:"type"`
	Links       ApiPullRequestCommentLinks       `json:"links"`
	PullRequest ApiPullRequestCommentPullRequest `json:"pullrequest"`
	Resolution  *ApiPullRequestCommentResolution `json:"resolution,omitempty"`
}

type ApiPullRequestCommentContent struct {
	Type   string `json:"type"`
	Raw    string `json:"raw"`
	Markup string `json:"markup"`
	HTML   string `json:"html"`
}

type ApiPullRequestCommentInline struct {
	From      *int   `json:"from"`
	To        *int   `json:"to"`
	Path      string `json:"path"`
	StartFrom *int   `json:"start_from"`
	StartTo   *int   `json:"start_to"`
}

type ApiPullRequestCommentLinks struct {
	Self ApiLink  `json:"self"`
	HTML ApiLink  `json:"html"`
	Code *ApiLink `json:"code,omitempty"`
}

type ApiPullRequestCommentPullRequest struct {
	Type  string     `json:"type"`
	ID    int        `json:"id"`
	Title string     `json:"title"`
	Draft bool       `json:"draft"`
	Links ApiPRLinks `json:"links"`
}

type ApiPRLinks struct {
	Self ApiLink `json:"self"`
	HTML ApiLink `json:"html"`
}

type ApiPullRequestCommentParent struct {
	ID    int                              `json:"id"`
	Links ApiPullRequestCommentParentLinks `json:"links"`
}

type ApiPullRequestCommentParentLinks struct {
	Self ApiLink `json:"self"`
	HTML ApiLink `json:"html"`
}

type ApiPullRequestCommentResolution struct {
	Type *string `json:"type,omitempty"`
}

type ApiCreateRepositoryRequest struct {
	SCM         string                         `json:"scm"`
	IsPrivate   *bool                          `json:"is_private,omitempty"`
	Description string                         `json:"description,omitempty"`
	ForkPolicy  string                         `json:"fork_policy,omitempty"`
	Language    string                         `json:"language,omitempty"`
	HasIssues   *bool                          `json:"has_issues,omitempty"`
	HasWiki     *bool                          `json:"has_wiki,omitempty"`
	Project     *ApiCreateRepositoryProjectRef `json:"project,omitempty"`
}

type ApiCreateRepositoryProjectRef struct {
	Key string `json:"key,omitempty"`
}

type ApiCreateFilesRequest struct {
	Branch  string
	Message string
	Files   map[string]string
	Parents string
	Author  string
}

type ApiBranch struct {
	Type                 string          `json:"type"`
	Links                ApiBranchLinks  `json:"links"`
	Name                 string          `json:"name"`
	Target               ApiBranchTarget `json:"target"`
	MergeStrategies      []string        `json:"merge_strategies,omitempty"`
	DefaultMergeStrategy string          `json:"default_merge_strategy,omitempty"`
	SyncStrategies       []string        `json:"sync_strategies,omitempty"`
}

type ApiBranchTarget struct {
	Type       string               `json:"type"`
	Hash       string               `json:"hash"`
	Date       string               `json:"date,omitempty"`
	Author     *ApiCommitAuthor     `json:"author,omitempty"`
	Message    string               `json:"message,omitempty"`
	Repository *ApiCommitRepository `json:"repository,omitempty"`
	Parents    []ApiCommitParent    `json:"parents,omitempty"`
	Links      *ApiCommitLinks      `json:"links,omitempty"`
}

type ApiBranchLinks struct {
	Self              ApiLink  `json:"self"`
	Commits           ApiLink  `json:"commits"`
	HTML              ApiLink  `json:"html"`
	PullrequestCreate *ApiLink `json:"pullrequest_create,omitempty"`
}

type ApiCreateBranchRequest struct {
	Name   string                `json:"name"`
	Target ApiCreateBranchTarget `json:"target"`
}

type ApiCreateBranchTarget struct {
	Hash string `json:"hash"`
}

type ApiCreatePullRequestRequest struct {
	Title             string                         `json:"title"`
	Description       string                         `json:"description,omitempty"`
	Source            ApiCreatePullRequestBranch     `json:"source"`
	Destination       *ApiCreatePullRequestBranch    `json:"destination,omitempty"`
	CloseSourceBranch *bool                          `json:"close_source_branch,omitempty"`
	Draft             *bool                          `json:"draft,omitempty"`
	Reviewers         []ApiCreatePullRequestReviewer `json:"reviewers,omitempty"`
}

type ApiCreatePullRequestBranch struct {
	Branch ApiCreatePullRequestBranchName `json:"branch"`
}

type ApiCreatePullRequestBranchName struct {
	Name string `json:"name"`
}

type ApiCreatePullRequestReviewer struct {
	UUID string `json:"uuid"`
}

type ApiCreatePullRequestCommentRequest struct {
	Content ApiCreatePullRequestCommentContent `json:"content"`
	Inline  *ApiPullRequestCommentInline       `json:"inline,omitempty"`
}

type ApiCreatePullRequestCommentContent struct {
	Raw string `json:"raw"`
}

type ApiMergePullRequestRequest struct {
	Type              string `json:"type"`
	Message           string `json:"message,omitempty"`
	CloseSourceBranch *bool  `json:"close_source_branch,omitempty"`
	MergeStrategy     string `json:"merge_strategy,omitempty"`
}
