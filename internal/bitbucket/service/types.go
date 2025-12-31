package service

// Page represents a paginated response containing a list of items.
type Page[T any] struct {
	PageSize int `json:"pagelen"` // Number of items per page
	Size     int `json:"size"`    // Total number of items available
	Page     int `json:"page"`    // Current page number (1-based)
	Items    []T `json:"items"`   // Items in the current page
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
