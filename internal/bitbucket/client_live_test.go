// Package bitbucket_live_test contains live integration tests that hit the real Bitbucket API
// to verify that the client implementation actually works against real endpoints.
//
// To run these tests, you need to:
// 1. Create a test data file at testdata/live/bitbucket.json with structure matching the TestData struct
// 2. Set up environment variables with real Bitbucket credentials to access the Bitbucket API
//
// In most cases, you don't need to run these tests. They will be automatically skipped
// if the test data file is not present, so don't worry if you see them as skipped.
package bitbucket_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/branow/mcp-bitbucket/internal/bitbucket"
	"github.com/branow/mcp-bitbucket/internal/config"
	"github.com/stretchr/testify/require"
)

var cfg = bitbucket.Config{
	Username: config.BitBucketEmail(),
	Password: config.BitBucketApiToken(),
	BaseUrl:  config.BitBucketUrl(),
	Timeout:  config.BitBucketTimeout(),
}

var client = bitbucket.NewClient(cfg)
var tests = loadTestData()

const testdataDir = "testdata/live"

// func TestListRepositories(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		t.Run(fmt.Sprintf("list repositories %s", workspace.Slug), func(t *testing.T) {
// 			resp, err := client.ListRepositories(workspace.Slug, 10, 1)
// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 			saveJson(t, "repository-list.json", resp)
// 		})
// 	}
// }

// func TestGetRepository(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		for _, repository := range workspace.Repositories {
// 			t.Run(fmt.Sprintf("get repository %s", repository.Slug), func(t *testing.T) {
// 				resp, err := client.GetRepository(workspace.Slug, repository.Slug)
// 				require.NoError(t, err)
// 				require.NotNil(t, resp)
// 				saveJson(t, fmt.Sprintf("repository-%s.json", repository.Slug), resp)
// 			})
// 		}
// 	}
// }

// func TestGetRepositorySource(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		for _, repository := range workspace.Repositories {
// 			t.Run(fmt.Sprintf("get repository source %s", repository.Slug), func(t *testing.T) {
// 				resp, err := client.GetRepositorySource(workspace.Slug, repository.Slug)
// 				require.NoError(t, err)
// 				require.NotNil(t, resp)
// 				saveJson(t, fmt.Sprintf("repository-src-%s.json", repository.Slug), resp)
// 			})
// 		}
// 	}
// }

// func TestListPullRequests(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		for _, repository := range workspace.Repositories {
// 			t.Run(fmt.Sprintf("list pull requests %s", repository.Slug), func(t *testing.T) {
// 				resp, err := client.ListPullRequests(workspace.Slug, repository.Slug, 10, 1, repository.PullRequestStates)
// 				require.NoError(t, err)
// 				require.NotNil(t, resp)
// 				saveJson(t, fmt.Sprintf("pull-requests-%s.json", repository.Slug), resp)
// 			})
// 		}
// 	}
// }

// func TestGetPullRequest(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		for _, repository := range workspace.Repositories {
// 			for _, pr := range repository.PullRequests {
// 				t.Run(fmt.Sprintf("get pull request %s-%d", repository.Slug, pr.Id), func(t *testing.T) {
// 					resp, err := client.GetPullRequest(workspace.Slug, repository.Slug, pr.Id)
// 					require.NoError(t, err)
// 					require.NotNil(t, resp)
// 					saveJson(t, fmt.Sprintf("pull-request-%s-%d.json", repository.Slug, pr.Id), resp)
// 				})
// 			}
// 		}
// 	}
// }

// func TestListPullRequestCommits(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		for _, repository := range workspace.Repositories {
// 			for _, pr := range repository.PullRequests {
// 				t.Run(fmt.Sprintf("list pull request commits %s-%d", repository.Slug, pr.Id), func(t *testing.T) {
// 					resp, err := client.ListPullRequestCommits(workspace.Slug, repository.Slug, pr.Id)
// 					require.NoError(t, err)
// 					require.NotNil(t, resp)
// 					saveJson(t, fmt.Sprintf("pull-request-commits-%s-%d.json", repository.Slug, pr.Id), resp)
// 				})
// 			}
// 		}
// 	}
// }

// func TestListPullRequestComments(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		for _, repository := range workspace.Repositories {
// 			for _, pr := range repository.PullRequests {
// 				t.Run(fmt.Sprintf("list pull request comments %s-%d", repository.Slug, pr.Id), func(t *testing.T) {
// 					resp, err := client.ListPullRequestComments(workspace.Slug, repository.Slug, pr.Id, 10, 1)
// 					require.NoError(t, err)
// 					require.NotNil(t, resp)
// 					saveJson(t, fmt.Sprintf("pull-request-comments-%s-%d.json", repository.Slug, pr.Id), resp)
// 				})
// 			}
// 		}
// 	}
// }

// func TestGetPullRequestDiff(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		for _, repository := range workspace.Repositories {
// 			for _, pr := range repository.PullRequests {
// 				t.Run(fmt.Sprintf("get pull request diff %s-%d", repository.Slug, pr.Id), func(t *testing.T) {
// 					resp, err := client.GetPullRequestDiff(workspace.Slug, repository.Slug, pr.Id)
// 					require.NoError(t, err)
// 					require.NotNil(t, resp)
// 					saveText(t, fmt.Sprintf("pull-request-diff-%s-%d.txt", repository.Slug, pr.Id), *resp)
// 				})
// 			}
// 		}
// 	}
// }

// func TestGetFileSource(t *testing.T) {
// 	skipIfNoTestData(t)
// 	for _, workspace := range tests {
// 		for _, repository := range workspace.Repositories {
// 			for _, file := range repository.Files {
// 				t.Run(fmt.Sprintf("get file source %s-%s", repository.Slug, file.Path), func(t *testing.T) {
// 					resp, err := client.GetFileSource(workspace.Slug, repository.Slug, file.Commit, file.Path)
// 					require.NoError(t, err)
// 					require.NotNil(t, resp)
// 					saveText(t, fmt.Sprintf("file-source-%s-%s.txt", repository.Slug, file.Commit), *resp)
// 				})
// 			}
// 		}
// 	}
// }

func TestGetDirectorySource(t *testing.T) {
	skipIfNoTestData(t)
	for _, workspace := range tests {
		for _, repository := range workspace.Repositories {
			for _, dir := range repository.Directories {
				t.Run(fmt.Sprintf("get directory source %s-%s", repository.Slug, dir.Path), func(t *testing.T) {
					resp, err := client.GetDirectorySource(workspace.Slug, repository.Slug, dir.Commit, dir.Path)
					require.NoError(t, err)
					require.NotNil(t, resp)
					saveJson(t, fmt.Sprintf("directory-source-%s-%s.json", repository.Slug, dir.Commit), resp)
				})
			}
		}
	}
}

type TestData struct {
	Slug         string `json:"slug"`
	Repositories []struct {
		Slug         string `json:"slug"`
		PullRequests []struct {
			Id int `json:"id"`
		} `json:"pull_requests"`
		PullRequestStates []string `json:"pull_request_states"`
		Files             []struct {
			Commit string `json:"commit"`
			Path   string `json:"path"`
		} `json:"files"`
		Directories []struct {
			Commit string `json:"commit"`
			Path   string `json:"path"`
		} `json:"directories"`
	} `json:"repositories"`
}

func loadTestData() []TestData {
	filepath, err := getFilePath("bitbucket.json")
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil
	}

	var tests []TestData
	err = json.Unmarshal(data, &tests)
	if err != nil {
		return nil
	}

	return tests
}

func skipIfNoTestData(t *testing.T) {
	t.Helper()
	if len(tests) == 0 {
		t.Skipf("Skipping %s: bitbucket.json not found or empty", t.Name())
	}
}

func saveJson(t *testing.T, filename string, data any) {
	t.Helper()

	filepath, err := getFilePath(filename)
	require.NoError(t, err)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(filepath, jsonData, 0644)
	require.NoError(t, err)

	t.Logf("Response saved to %s", filepath)
}

func saveText(t *testing.T, filename string, data string) {
	t.Helper()

	filepath, err := getFilePath(filename)
	require.NoError(t, err)

	err = os.WriteFile(filepath, []byte(data), 0644)
	require.NoError(t, err)

	t.Logf("Response saved to %s", filepath)
}

func getFilePath(path string) (string, error) {
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(testdataDir, path), nil
}
