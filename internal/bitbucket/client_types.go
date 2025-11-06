package bitbucket

type BitBucketErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type BitbucketApiResponse[T any] struct {
	Values  []T    `json:"values"`
	Pagelen int    `json:"pagelen"`
	Size    int    `json:"size"`
	Page    int    `json:"page"`
	Next    string `json:"next"`
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
	Href string `json:"href"`
}

type CloneLink struct {
	Name string `json:"name"`
	Href string `json:"href"`
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
	CommentCount      int                `json:"comment_count"`
	TaskCount         int                `json:"task_count"`
	Type              string             `json:"type"`
	ID                int                `json:"id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	State             string             `json:"state"`
	Draft             bool               `json:"draft"`
	MergeCommit       *PullRequestCommit `json:"merge_commit"`
	CloseSourceBranch bool               `json:"close_source_branch"`
	ClosedBy          *User              `json:"closed_by"`
	Author            User               `json:"author"`
	Reason            string             `json:"reason"`
	CreatedOn         string             `json:"created_on"`
	UpdatedOn         string             `json:"updated_on"`
	Destination       PullRequestBranch  `json:"destination"`
	Source            PullRequestBranch  `json:"source"`
	Links             PullRequestLinks   `json:"links"`
	Summary           PullRequestSummary `json:"summary"`
}

type User struct {
	DisplayName string      `json:"display_name"`
	Links       CommonLinks `json:"links"`
	Type        string      `json:"type"`
	UUID        string      `json:"uuid"`
	AccountID   string      `json:"account_id"`
	Nickname    string      `json:"nickname"`
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
	Name  string                 `json:"name"`
	Links map[string]interface{} `json:"links"`
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
