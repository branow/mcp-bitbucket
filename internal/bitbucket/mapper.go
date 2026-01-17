package bitbucket

// MapRepositoryDetails converts Bitbucket API data to domain RepositoryDetails type.
// Returns nil if the input repository is nil.
func MapRepositoryDetails(repository *ApiRepository, src *ApiResponse[ApiSourceItem], readmeSrc *ApiSourceItem, readmeContent *string) *RepositoryDetails {
	if repository == nil {
		return nil
	}

	return &RepositoryDetails{
		Repository: MapRepository(repository),
		Source:     MapPage(src, MapSourceItem),
		Readme:     MapSourceFile(readmeSrc, readmeContent),
	}
}

// MapList converts a slice of items from source type to target type.
// It applies the provided mapper function to each item in the slice.
//
// Type parameters:
//   - T: The source type
//   - U: The target type
//
// Parameters:
//   - items: The slice of items to map
//   - mapper: Function to convert each item from type T to type U
//
// Returns a slice containing the mapped items.
func MapList[T, U any](items []T, mapper func(*T) *U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = *mapper(&item)
	}
	return result
}

// MapPage converts a Bitbucket API response to a domain Page type.
// It applies the provided mapper function to each item in the response.
//
// Type parameters:
//   - T: The source type from the Bitbucket API
//   - U: The target domain type
//
// Parameters:
//   - resp: The API response containing a list of items
//   - mapper: Function to convert each item from type T to type U
//
// Returns a Page containing the mapped items with pagination metadata.
// Returns nil if the input API response is nil.
func MapPage[T, U any](resp *ApiResponse[T], mapper func(*T) *U) *Page[U] {
	if resp == nil {
		return nil
	}

	content := make([]U, len(resp.Values))
	for i, value := range resp.Values {
		content[i] = *mapper(&value)
	}

	page := 1
	if resp.Page != nil {
		page = *resp.Page
	}

	size := 0
	if resp.Size != nil {
		size = *resp.Size
	}

	return &Page[U]{
		PageSize: resp.Pagelen,
		Size:     size,
		Page:     page,
		Items:    content,
	}
}

// MapSourceFile converts a Bitbucket API ApiSourceItem to domain SourceFile type with content.
// Returns nil if the input source is nil.
func MapSourceFile(src *ApiSourceItem, content *string) *SourceFile {
	if src == nil {
		return nil
	}

	return &SourceFile{
		Path:        src.Path,
		Commit:      src.Commit.Hash,
		EscapedPath: src.EscapedPath,
		Size:        src.Size,
		Mimetype:    src.Mimetype,
		Content:     content,
	}
}

// MapSourceItem converts a Bitbucket API ApiSourceItem to domain SourceItem type.
// Returns nil if the input source is nil.
func MapSourceItem(src *ApiSourceItem) *SourceItem {
	if src == nil {
		return nil
	}

	return &SourceItem{
		Path:        src.Path,
		Type:        src.Type,
		Commit:      src.Commit.Hash,
		EscapedPath: src.EscapedPath,
		Size:        src.Size,
		Mimetype:    src.Mimetype,
	}
}

// MapRepository converts a Bitbucket API ApiRepository to the domain Repository type.
// Returns nil if the input repository is nil.
func MapRepository(repository *ApiRepository) *Repository {
	if repository == nil {
		return nil
	}
	return &Repository{
		FullName:    repository.FullName,
		Name:        repository.Name,
		Slug:        repository.Slug,
		Description: repository.Description,
		Website:     MapStringPointer(repository.Website),
		IsPrivate:   repository.IsPrivate,
		ForkPolicy:  repository.ForkPolicy,
		CreatedOn:   repository.CreatedOn,
		UpdatedOn:   repository.UpdatedOn,
		Size:        repository.Size,
		Language:    repository.Language,
		UUID:        repository.UUID,
		SCM:         repository.SCM,
		MainBranch:  repository.MainBranch.Name,
		HasIssues:   repository.HasIssues,
		HasWiki:     repository.HasWiki,
		Parent:      MapParentRepository(repository.Parent),
		Project:     MapProject(&repository.Project),
		Workspace:   MapWorkspace(&repository.Workspace),
		Owner:       MapOwner(&repository.Owner),
		OverrideSettings: &OverrideSettings{
			DefaultMergeStrategy: repository.OverrideSettings.DefaultMergeStrategy,
			BranchingModel:       repository.OverrideSettings.BranchingModel,
		},
	}
}

// MapProject converts a Bitbucket API ApiProject to the domain Project type.
// Returns nil if the input project is nil.
func MapProject(project *ApiProject) *Project {
	if project == nil {
		return nil
	}
	return &Project{
		Key:  project.Key,
		UUID: project.UUID,
		Name: project.Name,
	}
}

// MapParentRepository converts a Bitbucket API ApiParentRepository to the domain ParentRepository type.
// Returns nil if the input parent repository is nil.
func MapParentRepository(parent *ApiParentRepository) *ParentRepository {
	if parent == nil {
		return nil
	}
	return &ParentRepository{
		FullName: parent.FullName,
		Name:     parent.Name,
		UUID:     parent.UUID,
	}
}

// MapOwner converts a Bitbucket API Owner to the domain Owner type.
// Returns nil if the input owner is nil.
func MapOwner(owner *ApiOwner) *Owner {
	if owner == nil {
		return nil
	}
	return &Owner{
		DisplayName: owner.DisplayName,
		UUID:        owner.UUID,
		Username:    owner.Username,
	}
}

// MapWorkspace converts a Bitbucket API ApiWorkspace to the domain Workspace type.
// Returns nil if the input workspace is nil.
func MapWorkspace(workspace *ApiWorkspace) *Workspace {
	if workspace == nil {
		return nil
	}
	return &Workspace{
		Name: workspace.Name,
		UUID: workspace.UUID,
		Slug: workspace.Slug,
	}
}

// MapStringPointer safely dereferences a string pointer, returning an empty string if nil.
func MapStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// MapPullRequestDetails converts Bitbucket API data to domain PullRequestDetails type.
// Returns nil if the input pull request is nil.
func MapPullRequestDetails(pr *ApiPullRequest, commits *ApiResponse[ApiCommit], diff *string, comments *ApiResponse[ApiPullRequestComment]) *PullRequestDetails {
	if pr == nil {
		return nil
	}

	return &PullRequestDetails{
		PullRequest: MapPullRequest(pr),
		Commits:     MapPage(commits, MapPullRequestCommit),
		Diff:        diff,
		Comments:    MapPage(comments, MapPullRequestComment),
	}
}

// MapPullRequest converts a Bitbucket API ApiPullRequest to the domain PullRequest type.
// Returns nil if the input pull request is nil.
func MapPullRequest(pr *ApiPullRequest) *PullRequest {
	if pr == nil {
		return nil
	}

	return &PullRequest{
		ID:                pr.ID,
		Title:             pr.Title,
		Description:       pr.Description,
		State:             pr.State,
		Draft:             pr.Draft,
		Author:            MapUser(&pr.Author),
		CreatedOn:         pr.CreatedOn,
		UpdatedOn:         pr.UpdatedOn,
		ClosedOn:          pr.ClosedOn,
		ClosedBy:          MapUser(pr.ClosedBy),
		Reason:            pr.Reason,
		MergeCommit:       MapMergeCommit(pr.MergeCommit),
		CloseSourceBranch: pr.CloseSourceBranch,
		CommentCount:      pr.CommentCount,
		TaskCount:         pr.TaskCount,
		Source:            MapPullRequestBranch(&pr.Source),
		Destination:       MapPullRequestBranch(&pr.Destination),
		Reviewers:         MapList(pr.Reviewers, MapUser),
		Participants:      MapList(pr.Participants, MapParticipant),
	}
}

// MapMergeCommit extracts the commit hash from a merge commit.
// Returns nil if the commit is nil.
func MapMergeCommit(commit *ApiPullRequestCommit) *string {
	if commit == nil {
		return nil
	}
	return &commit.Hash
}

// MapPullRequestBranch converts a Bitbucket API ApiPullRequestBranch to domain PullRequestBranch type.
// Returns nil if the input branch is nil.
func MapPullRequestBranch(branch *ApiPullRequestBranch) *PullRequestBranch {
	if branch == nil {
		return nil
	}

	return &PullRequestBranch{
		Name:       branch.Branch.Name,
		Hash:       branch.Commit.Hash,
		Repository: MapBranchRepository(&branch.Repository),
	}
}

// MapBranchRepository converts a Bitbucket API ApiRepository to domain BranchRepository type.
// Returns nil if the input repository is nil.
func MapBranchRepository(repo *ApiPullRequestRepository) *PullRequestRepository {
	if repo == nil {
		return nil
	}

	return &PullRequestRepository{
		FullName: repo.FullName,
		Name:     repo.Name,
		UUID:     repo.UUID,
	}
}

// MapParticipant converts a Bitbucket API ApiParticipant to domain Participant type.
// Returns nil if the input participant is nil.
func MapParticipant(participant *ApiPullRequestParticipant) *Participant {
	if participant == nil {
		return nil
	}

	return &Participant{
		User:           MapUser(&participant.User),
		Role:           participant.Role,
		Approved:       participant.Approved,
		State:          participant.State,
		ParticipatedOn: participant.ParticipatedOn,
	}
}

// MapUser converts a Bitbucket API ApiUser to domain User type.
// Returns nil if the input user is nil.
func MapUser(user *ApiUser) *User {
	if user == nil {
		return nil
	}

	return &User{
		DisplayName: user.DisplayName,
		UUID:        user.UUID,
		AccountId:   user.AccountID,
		Nickname:    user.Nickname,
		Username:    user.Username,
	}
}

// MapPullRequestCommit converts a Bitbucket API ApiCommit to domain PullRequestCommit type.
// Returns nil if the input commit is nil.
func MapPullRequestCommit(commit *ApiCommit) *PullRequestCommit {
	if commit == nil {
		return nil
	}

	var parent string
	if len(commit.Parents) > 0 {
		parent = commit.Parents[0].Hash
	}

	return &PullRequestCommit{
		Hash:    commit.Hash,
		Date:    commit.Date,
		Author:  MapUser(&commit.Author.User),
		Message: commit.Message,
		Parent:  parent,
	}
}

// MapPullRequestComment converts a Bitbucket API ApiPullRequestComment to domain PullRequestComment type.
// Returns nil if the input comment is nil.
func MapPullRequestComment(comment *ApiPullRequestComment) *PullRequestComment {
	if comment == nil {
		return nil
	}

	return &PullRequestComment{
		ID:        comment.ID,
		CreatedOn: comment.CreatedOn,
		UpdatedOn: comment.UpdatedOn,
		Content:   comment.Content.Raw,
		User:      MapUser(&comment.User),
		Deleted:   comment.Deleted,
		Pending:   comment.Pending,
		Inline:    MapInline(comment.Inline),
	}
}

// MapInline converts a Bitbucket API ApiPullRequestCommentInline to domain Inline type.
// Returns nil if the input inline is nil.
func MapInline(inline *ApiPullRequestCommentInline) *Inline {
	if inline == nil {
		return nil
	}

	return &Inline{
		Path: inline.Path,
		To:   inline.To,
		From: inline.From,
	}
}
