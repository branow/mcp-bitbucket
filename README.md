<p align="center">
  <h1 align="center">MCP Bitbucket</h1>
  <p align="center">
    A Model Context Protocol server for Bitbucket Cloud — browse repos, review PRs, and read source code from Claude.
  </p>
</p>

<p align="center">
  <a href="https://modelcontextprotocol.io"><img src="https://badge.mcpx.dev?type=server&features=tools,resources" alt="MCP Server" /></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT" /></a>
  <a href="https://github.com/branow/mcp-bitbucket/pkgs/container/mcp-bitbucket"><img src="https://img.shields.io/badge/GHCR-latest-blue?logo=docker" alt="Docker" /></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white" alt="Go 1.24" /></a>
</p>

---

## Features

- **Repository browsing** — list repos, view details, explore directory trees, read file content
- **Pull request workflows** — list PRs, get full details with commits, diffs, and comments in a single call
- **Resource templates** — URI-based access via the MCP resource protocol
- **Dual auth** — Basic Auth for quick setup, OAuth 2.0 for multi-user deployments
- **Docker-ready** — lightweight Alpine image (~15 MB), published to GHCR

### Benchmark

In a [benchmark](benchmark/REPORT.md) comparing 4 open-source Bitbucket MCP servers on a PR review task (Claude Opus 4.6), this server completed reviews at **7–11x lower cost** and in **2 turns** vs 5–21 for competitors — thanks to bundled tool responses and a minimal tool surface. See the [full report](benchmark/REPORT.md) for details.

## Quick Start

### 1. Start the server

```bash
docker run -d --restart unless-stopped \
  -p 9847:8080 \
  -e BITBUCKET_AUTH=basic \
  -e BITBUCKET_EMAIL=your@email.com \
  -e BITBUCKET_API_TOKEN=your-token \
  --name mcp-bitbucket \
  ghcr.io/branow/mcp-bitbucket:latest
```

### 2. Connect your client

<details>
<summary><strong>Claude Code (CLI)</strong></summary>

```bash
claude mcp add --transport http bitbucket http://localhost:9847/mcp
```

</details>

<details>
<summary><strong>Claude Desktop</strong></summary>

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "bitbucket": {
      "url": "http://localhost:9847/mcp"
    }
  }
}
```

</details>

<details>
<summary><strong>VS Code / Cursor</strong></summary>

Add to `.vscode/mcp.json` in your workspace:

```json
{
  "servers": {
    "bitbucket": {
      "url": "http://localhost:9847/mcp"
    }
  }
}
```

</details>

That's it. The container auto-restarts across reboots.

## Available Tools

<details>
<summary><strong>list_repositories</strong> — List repositories in a workspace</summary>

| Parameter | Type | Required | Description |
|---|---|---|---|
| `workspace` | string | Yes | Workspace slug or username |
| `page` | int | No | Page number (1-based). Default: 1 |
| `page_size` | int | No | Items per page. Default: 50 |

</details>

<details>
<summary><strong>get_repository</strong> — Get repository details with optional source listing and README</summary>

| Parameter | Type | Required | Description |
|---|---|---|---|
| `workspace` | string | Yes | Workspace slug or username |
| `repository` | string | Yes | Repository name/slug |
| `include_source` | bool | No | Include root-level source listing |
| `include_readme` | bool | No | Include README content if found |

</details>

<details>
<summary><strong>list_pull_requests</strong> — List pull requests with state filtering</summary>

| Parameter | Type | Required | Description |
|---|---|---|---|
| `workspace` | string | Yes | Workspace slug or username |
| `repository` | string | Yes | Repository name/slug |
| `state` | []string | No | Filter: OPEN, MERGED, DECLINED. Default: OPEN |
| `page` | int | No | Page number (1-based). Default: 1 |
| `page_size` | int | No | Items per page. Default: 25 |

</details>

<details>
<summary><strong>get_pull_request</strong> — Get PR details with optional commits, diff, and comments</summary>

| Parameter | Type | Required | Description |
|---|---|---|---|
| `workspace` | string | Yes | Workspace slug or username |
| `repository` | string | Yes | Repository name/slug |
| `id` | int | Yes | Pull request ID |
| `include_commits` | bool | No | Include commits |
| `include_diff` | bool | No | Include diff |
| `include_comments` | bool | No | Include comments |

</details>

<details>
<summary><strong>get_file_content</strong> — Get file content at a specific branch/commit</summary>

| Parameter | Type | Required | Description |
|---|---|---|---|
| `workspace` | string | Yes | Workspace slug or username |
| `repository` | string | Yes | Repository name/slug |
| `path` | string | Yes | File path relative to repository root |
| `ref` | string | No | Branch, tag, or commit hash. Default: main branch |

</details>

<details>
<summary><strong>get_directory_source</strong> — List files and subdirectories at a path</summary>

| Parameter | Type | Required | Description |
|---|---|---|---|
| `workspace` | string | Yes | Workspace slug or username |
| `repository` | string | Yes | Repository name/slug |
| `path` | string | Yes | Directory path relative to repository root |
| `ref` | string | No | Branch, tag, or commit hash. Default: main branch |

</details>

## Resource Templates

The server also exposes [MCP resource templates](https://modelcontextprotocol.io/docs/concepts/resources) for URI-based access:

| Template | URI Pattern |
|---|---|
| List Repositories | `bitbucket://api/{workspace}/repositories{?page,pageSize}` |
| Get Repository | `bitbucket://api/{workspace}/repositories/{repository}{?src,readme}` |
| List Pull Requests | `bitbucket://api/{workspace}/repositories/{repository}/pullrequests{?state,page,pageSize}` |
| Get Pull Request | `bitbucket://api/{workspace}/repositories/{repository}/pullrequests/{id}{?commits,diff,comments}` |
| File Content | `bitbucket://api/{workspace}/repositories/{repository}/src/{+path}{?ref}` |
| Directory Source | `bitbucket://api/{workspace}/repositories/{repository}/src/dir/{+path}{?ref}` |

## Configuration

### Authentication

The server supports two authentication methods:

| Method | When to use | Variables needed |
|---|---|---|
| **Basic Auth** | Personal use, quick setup | `BITBUCKET_AUTH=basic`, `BITBUCKET_EMAIL`, `BITBUCKET_API_TOKEN` |
| **OAuth 2.0** (experimental) | Multi-user, production deployments | `BITBUCKET_AUTH=oauth`, `SERVER_URL`, `OAUTH_ISSUER`, `OAUTH_SCOPES` |

> **Note:** OAuth 2.0 support is experimental and has not been fully tested yet. Use Basic Auth for a reliable setup.

To create a Bitbucket API token for basic auth, go to **Bitbucket Settings > Personal Settings > App passwords** and create a token with `repository:read` and `pullrequest:read` scopes.

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8080` | HTTP server port |
| `BITBUCKET_AUTH` | `oauth` | Auth type: `basic` or `oauth` |
| `BITBUCKET_URL` | `https://api.bitbucket.org/2.0` | Bitbucket API base URL |
| `BITBUCKET_TIMEOUT` | `5` | HTTP request timeout in seconds |
| `BITBUCKET_EMAIL` | — | Username/email for basic auth |
| `BITBUCKET_API_TOKEN` | — | API token for basic auth |
| `SERVER_URL` | — | Base URL of this server (OAuth only) |
| `OAUTH_ISSUER` | `https://bitbucket.org` | OAuth issuer URL |
| `OAUTH_SCOPES` | `repository;pullrequest` | Required OAuth scopes (semicolon-separated) |

## Build from Source

```bash
# Clone and build
git clone https://github.com/branow/mcp-bitbucket.git
cd mcp-bitbucket
docker build -t mcp-bitbucket .

# Run
docker run -d --restart unless-stopped \
  -p 9847:8080 \
  -e BITBUCKET_AUTH=basic \
  -e BITBUCKET_EMAIL=your@email.com \
  -e BITBUCKET_API_TOKEN=your-token \
  --name mcp-bitbucket \
  mcp-bitbucket
```

Or run directly with Go:

```bash
go run ./cmd
```

## Architecture

```
cmd/main.go                          # Entry point
internal/
  server/server.go                   # HTTP server setup
  mcp/handler.go                     # MCP protocol handler
  mcp/tools/                         # 6 MCP tools
  mcp/templates/                     # 6 resource templates
  bitbucket/                         # Bitbucket API client & service
  auth/                              # OAuth 2.0 & Basic Auth
  config/                            # Environment-based configuration
  health/                            # Health check endpoint (/health)
```

Key design choices:
- **Bundled tool responses** — tools like `get_pull_request` and `get_repository` fetch related data in parallel using `errgroup`, returning everything in one call instead of requiring multiple round-trips
- **Mapped responses** — domain models strip irrelevant Bitbucket API metadata, reducing token usage
- **Minimal tool surface** — 6 focused tools with concise schemas keep system prompt size small

## Contributing

Contributions are welcome! Please open an issue first to discuss what you'd like to change.

## License

[MIT](LICENSE) — Copyright (c) 2026 Orest Bodnar
