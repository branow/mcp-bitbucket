package bitbucket

type BitBucketErrorResponse struct {
	Type  string             `json:"type"`
	Error BitBucketErrorBody `json:"error"`
}

type BitBucketErrorBody struct {
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
	Detail  string              `json:"detail,omitempty"`
	ID      string              `json:"id,omitempty"`
	Data    map[string]any      `json:"data,omitempty"`
}

type BitbucketApiResponse[T any] struct {
	Values   []T     `json:"values"`
	Pagelen  int     `json:"pagelen"`
	Size     *int    `json:"size,omitempty"`
	Page     *int    `json:"page,omitempty"`
	Next     *string `json:"next,omitempty"`
	Previous *string `json:"previous,omitempty"`
}

const (
	BitbcuketPullRequestStateOpen       = "OPEN"
	BitbcuketPullRequestStateMerged     = "MERGED"
	BitbcuketPullRequestStateDeclined   = "DECLINED"
	BitbcuketPullRequestStateSuperseded = "SUPERSEDED"
)

type BitbucketRepository struct {
	Type                  string                     `json:"type"`
	FullName              string                     `json:"full_name"`
	Links                 BitbucketRepositoryLinks   `json:"links"`
	Name                  string                     `json:"name"`
	Slug                  string                     `json:"slug"`
	Description           string                     `json:"description"`
	SCM                   string                     `json:"scm"`
	Website               *string                    `json:"website"`
	Owner                 BitbucketOwner             `json:"owner"`
	Workspace             BitbucketWorkspace         `json:"workspace"`
	IsPrivate             bool                       `json:"is_private"`
	Project               BitbucketProject           `json:"project"`
	ForkPolicy            string                     `json:"fork_policy"`
	CreatedOn             string                     `json:"created_on"`
	UpdatedOn             string                     `json:"updated_on"`
	Size                  int                        `json:"size"`
	Language              string                     `json:"language"`
	UUID                  string                     `json:"uuid"`
	MainBranch            BitbucketMainBranch        `json:"mainbranch"`
	OverrideSettings      BitbucketOverrideSettings  `json:"override_settings"`
	Parent                *BitbucketParentRepository `json:"parent"`
	EnforcedSignedCommits *bool                      `json:"enforced_signed_commits"`
	HasIssues             bool                       `json:"has_issues"`
	HasWiki               bool                       `json:"has_wiki"`
}

type BitbucketRepositoryLinks struct {
	Self         BitbucketLink        `json:"self"`
	HTML         BitbucketLink        `json:"html"`
	Avatar       BitbucketLink        `json:"avatar"`
	Pullrequests BitbucketLink        `json:"pullrequests"`
	Commits      BitbucketLink        `json:"commits"`
	Forks        BitbucketLink        `json:"forks"`
	Watchers     BitbucketLink        `json:"watchers"`
	Branches     BitbucketLink        `json:"branches"`
	Tags         BitbucketLink        `json:"tags"`
	Downloads    BitbucketLink        `json:"downloads"`
	Source       BitbucketLink        `json:"source"`
	Clone        []BitbucketCloneLink `json:"clone"`
	Hooks        BitbucketLink        `json:"hooks"`
}

type BitbucketLink struct {
	Href string  `json:"href"`
	Name *string `json:"name,omitempty"`
}

type BitbucketCloneLink struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type BitbucketCommonLinks struct {
	Self   BitbucketLink `json:"self"`
	HTML   BitbucketLink `json:"html"`
	Avatar BitbucketLink `json:"avatar"`
}

type BitbucketOwner struct {
	DisplayName string               `json:"display_name"`
	Links       BitbucketCommonLinks `json:"links"`
	Type        string               `json:"type"`
	UUID        string               `json:"uuid"`
	Username    string               `json:"username"`
}

type BitbucketWorkspace struct {
	Type  string               `json:"type"`
	UUID  string               `json:"uuid"`
	Name  string               `json:"name"`
	Slug  string               `json:"slug"`
	Links BitbucketCommonLinks `json:"links"`
}

type BitbucketProject struct {
	Type  string               `json:"type"`
	Key   string               `json:"key"`
	UUID  string               `json:"uuid"`
	Name  string               `json:"name"`
	Links BitbucketCommonLinks `json:"links"`
}

type BitbucketMainBranch struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type BitbucketOverrideSettings struct {
	DefaultMergeStrategy bool `json:"default_merge_strategy"`
	BranchingModel       bool `json:"branching_model"`
}

type BitbucketParentRepository struct {
	Type     string               `json:"type"`
	FullName string               `json:"full_name"`
	Links    BitbucketCommonLinks `json:"links"`
	Name     string               `json:"name"`
	UUID     string               `json:"uuid"`
}

type BitbucketSourceItem struct {
	Path        string                   `json:"path"`
	Type        string                   `json:"type"`
	Commit      BitbucketSourceCommit    `json:"commit"`
	Links       BitbucketSourceItemLinks `json:"links"`
	EscapedPath *string                  `json:"escaped_path,omitempty"`
	Size        *int                     `json:"size,omitempty"`
	Mimetype    *string                  `json:"mimetype"`
	Attributes  []string                 `json:"attributes,omitempty"`
}

type BitbucketSourceCommit struct {
	Hash  string                     `json:"hash"`
	Type  string                     `json:"type"`
	Links BitbucketSourceCommitLinks `json:"links"`
}

type BitbucketSourceCommitLinks struct {
	Self BitbucketLink `json:"self"`
	HTML BitbucketLink `json:"html"`
}

type BitbucketSourceItemLinks struct {
	Self    BitbucketLink  `json:"self"`
	Meta    BitbucketLink  `json:"meta"`
	History *BitbucketLink `json:"history,omitempty"`
}

type BitbucketPullRequest struct {
	CommentCount      int                               `json:"comment_count"`
	TaskCount         int                               `json:"task_count"`
	Type              string                            `json:"type"`
	ID                int                               `json:"id"`
	Title             string                            `json:"title"`
	Description       string                            `json:"description"`
	Rendered          *BitbucketPullRequestRendered     `json:"rendered,omitempty"`
	State             string                            `json:"state"`
	Draft             bool                              `json:"draft"`
	MergeCommit       *BitbucketPullRequestCommit       `json:"merge_commit"`
	CloseSourceBranch bool                              `json:"close_source_branch"`
	ClosedBy          *BitbucketUser                    `json:"closed_by"`
	Author            BitbucketUser                     `json:"author"`
	Reason            *string                           `json:"reason,omitempty"`
	CreatedOn         string                            `json:"created_on"`
	UpdatedOn         string                            `json:"updated_on"`
	Destination       BitbucketPullRequestBranch        `json:"destination"`
	Source            BitbucketPullRequestBranch        `json:"source"`
	Reviewers         []BitbucketUser                   `json:"reviewers,omitempty"`
	Participants      []BitbucketPullRequestParticipant `json:"participants,omitempty"`
	Links             BitbucketPullRequestLinks         `json:"links"`
	Summary           BitbucketPullRequestSummary       `json:"summary"`
}

type BitbucketUser struct {
	DisplayName string               `json:"display_name"`
	Links       BitbucketCommonLinks `json:"links"`
	Type        string               `json:"type"`
	UUID        string               `json:"uuid"`
	AccountID   string               `json:"account_id"`
	Nickname    string               `json:"nickname"`
	Username    string               `json:"username"`
}

type BitbucketPullRequestCommit struct {
	Hash  string               `json:"hash"`
	Links BitbucketCommonLinks `json:"links"`
	Type  string               `json:"type"`
}

type BitbucketPullRequestBranch struct {
	Branch     BitbucketBranchInfo            `json:"branch"`
	Commit     BitbucketPullRequestCommit     `json:"commit"`
	Repository BitbucketPullRequestRepository `json:"repository"`
}

type BitbucketBranchInfo struct {
	Name           string         `json:"name"`
	Links          map[string]any `json:"links"`
	SyncStrategies []string       `json:"sync_strategies,omitempty"`
}

type BitbucketPullRequestRepository struct {
	Type     string               `json:"type"`
	FullName string               `json:"full_name"`
	Links    BitbucketCommonLinks `json:"links"`
	Name     string               `json:"name"`
	UUID     string               `json:"uuid"`
}

type BitbucketPullRequestLinks struct {
	Self           BitbucketLink `json:"self"`
	HTML           BitbucketLink `json:"html"`
	Commits        BitbucketLink `json:"commits"`
	Approve        BitbucketLink `json:"approve"`
	RequestChanges BitbucketLink `json:"request-changes"`
	Diff           BitbucketLink `json:"diff"`
	Diffstat       BitbucketLink `json:"diffstat"`
	Comments       BitbucketLink `json:"comments"`
	Activity       BitbucketLink `json:"activity"`
	Merge          BitbucketLink `json:"merge"`
	Decline        BitbucketLink `json:"decline"`
	Statuses       BitbucketLink `json:"statuses"`
}

type BitbucketPullRequestSummary struct {
	Type   string `json:"type"`
	Raw    string `json:"raw"`
	Markup string `json:"markup"`
	HTML   string `json:"html"`
}

type BitbucketPullRequestRendered struct {
	Title       BitbucketPullRequestSummary  `json:"title"`
	Description BitbucketPullRequestSummary  `json:"description"`
	Reason      *BitbucketPullRequestSummary `json:"reason,omitempty"`
}

type BitbucketPullRequestParticipant struct {
	Type           string        `json:"type"`
	User           BitbucketUser `json:"user"`
	Role           string        `json:"role"`
	Approved       bool          `json:"approved"`
	State          *string       `json:"state"`
	ParticipatedOn *string       `json:"participated_on"`
}
