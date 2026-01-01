package service

// Page represents a paginated response containing a list of items.
type Page[T any] struct {
	PageSize int `json:"pagelen"` // Number of items per page
	Size     int `json:"size"`    // Total number of items available
	Page     int `json:"page"`    // Current page number (1-based)
	Items    []T `json:"items"`   // Items in the current page
}

// RepositoryDetails represents detailed information about a repository including optional source listing and README.
type RepositoryDetails struct {
	Repository *Repository       `json:"repository"`
	Readme     *SourceFile       `json:"readme,omitempty"`
	Source     *Page[SourceItem] `json:"source,omitempty"`
}

// Repository represents a Bitbucket repository with simplified fields for domain use.
type Repository struct {
	FullName         string            `json:"full_name"`
	Name             string            `json:"name"`
	Slug             string            `json:"slug"`
	Description      string            `json:"description"`
	Website          string            `json:"website"`
	IsPrivate        bool              `json:"is_private"`
	Project          *Project          `json:"project"`
	ForkPolicy       string            `json:"fork_policy"`
	CreatedOn        string            `json:"created_on"`
	UpdatedOn        string            `json:"updated_on"`
	Size             int               `json:"size"`
	Language         string            `json:"language"`
	UUID             string            `json:"uuid"`
	SCM              string            `json:"scm"`
	MainBranch       string            `json:"mainbranch"`
	OverrideSettings *OverrideSettings `json:"override_settings"`
	Parent           *ParentRepository `json:"parent"`
	HasIssues        bool              `json:"has_issues"`
	HasWiki          bool              `json:"has_wiki"`
	Owner            *Owner            `json:"owner"`
	Workspace        *Workspace        `json:"workspace"`
}

// Project represents a Bitbucket project containing repositories.
type Project struct {
	Key  string `json:"key"`
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// OverrideSettings represents repository settings that override workspace defaults.
type OverrideSettings struct {
	DefaultMergeStrategy bool `json:"default_merge_strategy"`
	BranchingModel       bool `json:"branching_model"`
}

// ParentRepository represents the parent repository if this repository is a fork.
type ParentRepository struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
}

// Owner represents the owner (user or team) of a Bitbucket repository.
type Owner struct {
	DisplayName string `json:"display_name"`
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
}

// Workspace represents a Bitbucket workspace containing projects and repositories.
type Workspace struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// SourceFile represents a file from the repository source with its content.
type SourceFile struct {
	Path        string  `json:"path"`
	Commit      string  `json:"commit"`
	EscapedPath *string `json:"escaped_path,omitempty"`
	Size        *int    `json:"size,omitempty"`
	Mimetype    *string `json:"mimetype,omitempty"`
	Content     *string `json:"content"`
}

// SourceItem represents a file or directory entry in the repository source listing.
type SourceItem struct {
	Path        string  `json:"path"`
	Type        string  `json:"type"`
	Commit      string  `json:"commit"`
	EscapedPath *string `json:"escaped_path,omitempty"`
	Size        *int    `json:"size,omitempty"`
	Mimetype    *string `json:"mimetype,omitempty"`
}

// PullRequestDetails represents detailed information about a pull request including optional commits, diff, and comments.
type PullRequestDetails struct {
	PullRequest *PullRequest              `json:"pullRequest"`
	Commits     *Page[PullRequestCommit]  `json:"commits,omitempty"`
	Diff        *string                   `json:"diff,omitempty"`
	Comments    *Page[PullRequestComment] `json:"comments,omitempty"`
}

// PullRequest represents a Bitbucket pull request with simplified fields for domain use.
type PullRequest struct {
	ID                int                `json:"id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	State             string             `json:"state"`
	Draft             bool               `json:"draft"`
	Author            *User              `json:"author"`
	CreatedOn         string             `json:"created_on"`
	UpdatedOn         string             `json:"updated_on"`
	ClosedOn          *string            `json:"closed_on,omitempty"`
	ClosedBy          *User              `json:"closed_by,omitempty"`
	Reason            *string            `json:"reason,omitempty"`
	MergeCommit       *string            `json:"merge_commit,omitempty"`
	CloseSourceBranch bool               `json:"close_source_branch"`
	CommentCount      int                `json:"comment_count"`
	TaskCount         int                `json:"task_count"`
	Source            *PullRequestBranch `json:"source"`
	Destination       *PullRequestBranch `json:"destination"`
	Reviewers         []User             `json:"reviewers,omitempty"`
	Participants      []Participant      `json:"participants,omitempty"`
}

// PullRequestBranch represents a source or destination branch in a pull request.
type PullRequestBranch struct {
	Name       string                 `json:"name"`
	Hash       string                 `json:"hash"`
	Repository *PullRequestRepository `json:"repository"`
}

// PullRequestRepository represents a minimal repository reference in a pull request.
type PullRequestRepository struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
}

// Participant represents a user who has participated in a pull request.
type Participant struct {
	User           *User   `json:"user"`
	Role           string  `json:"role"`
	Approved       bool    `json:"approved"`
	State          *string `json:"state,omitempty"`
	ParticipatedOn *string `json:"participated_on,omitempty"`
}

// User represents a Bitbucket user.
type User struct {
	DisplayName string `json:"display_name"`
	UUID        string `json:"uuid"`
	AccountId   string `json:"account_id"`
	Nickname    string `json:"nickname,omitempty"`
	Username    string `json:"username,omitempty"`
}

// Commit represents a commit in a pull request.
type PullRequestCommit struct {
	Hash    string `json:"hash"`
	Date    string `json:"date"`
	Author  *User  `json:"author"`
	Message string `json:"message"`
	Parent  string `json:"parent"`
}

// PullRequestComment represents a comment on a pull request.
type PullRequestComment struct {
	ID        int     `json:"id"`
	CreatedOn string  `json:"created_on"`
	UpdatedOn string  `json:"updated_on"`
	Content   string  `json:"content"`
	User      *User   `json:"user"`
	Deleted   bool    `json:"deleted"`
	Pending   bool    `json:"pending"`
	Inline    *Inline `json:"inline,omitempty"`
}

// Inline represents inline comment information (file path and line number).
type Inline struct {
	Path string `json:"path"`
	To   *int   `json:"to,omitempty"`
	From *int   `json:"from,omitempty"`
}
