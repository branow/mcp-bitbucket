package service

import "github.com/branow/mcp-bitbucket/internal/bitbucket/client"

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
func MapPage[T, U any](resp *client.ApiResponse[T], mapper func(*T) *U) *Page[U] {
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

// MapRepository converts a Bitbucket API Repository to the domain Repository type.
// Returns nil if the input repository is nil.
func MapRepository(repository *client.Repository) *Repository {
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

// MapProject converts a Bitbucket API Project to the domain Project type.
// Returns nil if the input project is nil.
func MapProject(project *client.Project) *Project {
	if project == nil {
		return nil
	}
	return &Project{
		Key:  project.Key,
		UUID: project.UUID,
		Name: project.Name,
	}
}

// MapParentRepository converts a Bitbucket API ParentRepository to the domain ParentRepository type.
// Returns nil if the input parent repository is nil.
func MapParentRepository(parent *client.ParentRepository) *ParentRepository {
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
func MapOwner(owner *client.Owner) *Owner {
	if owner == nil {
		return nil
	}
	return &Owner{
		DisplayName: owner.DisplayName,
		UUID:        owner.UUID,
		Username:    owner.Username,
	}
}

// MapWorkspace converts a Bitbucket API Workspace to the domain Workspace type.
// Returns nil if the input workspace is nil.
func MapWorkspace(workspace *client.Workspace) *Workspace {
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
