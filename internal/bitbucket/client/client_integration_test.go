// Package bitbucket_live_test contains live integration tests that hit the real Bitbucket API
// to verify that the client implementation actually works against real endpoints.
//
// To run these tests, you need to:
// 1. Create a test data file at testdata/live/bitbucket.json with structure matching the TestData struct
// 2. Set up environment variables with real Bitbucket credentials to access the Bitbucket API
//
// In most cases, you don't need to run these tests. They will be automatically skipped
// if the test data file is not present, so don't worry if you see them as skipped.
package client_test

import (
	"github.com/branow/mcp-bitbucket/internal/bitbucket/client"
	"github.com/branow/mcp-bitbucket/internal/config"
)

var cfg = config.NewGlobal()
var bb = client.NewClient(cfg)
