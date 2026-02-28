# MCP Bitbucket

A Model Context Protocol (MCP) server for Bitbucket integration.

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

### 2. Add to Claude Code

```bash
claude mcp add --transport http bitbucket http://localhost:9847/mcp
```

That's it. The container auto-restarts across reboots.

## Available Tools

| Tool | Description |
|---|---|
| `list_repositories` | List repositories in a workspace |
| `get_repository` | Get repository details (with optional source listing and README) |
| `list_pull_requests` | List pull requests with state filtering |
| `get_pull_request` | Get pull request details (with optional commits, diff, comments) |
| `get_file_content` | Get file content at a specific ref |
| `get_directory_source` | Get directory listing at a specific ref |

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `SERVER_PORT` | HTTP server port | `8080` |
| `BITBUCKET_AUTH` | Auth type: `basic` or `oauth` | `oauth` |
| `BITBUCKET_URL` | Bitbucket API base URL | `https://api.bitbucket.org/2.0` |
| `BITBUCKET_TIMEOUT` | HTTP request timeout (seconds) | `5` |
| `BITBUCKET_EMAIL` | Username/email for basic auth | -- |
| `BITBUCKET_API_TOKEN` | API token for basic auth | -- |
| `SERVER_URL` | Base URL of this server (OAuth) | -- |
| `OAUTH_ISSUER` | OAuth issuer URL | `https://bitbucket.org` |
| `OAUTH_SCOPES` | OAuth scopes, semicolon-separated | `repository;pullrequest` |

## Build from Source

```bash
docker build -t mcp-bitbucket .
docker run -d --restart unless-stopped \
  -p 9847:8080 \
  -e BITBUCKET_AUTH=basic \
  -e BITBUCKET_EMAIL=your@email.com \
  -e BITBUCKET_API_TOKEN=your-token \
  --name mcp-bitbucket \
  mcp-bitbucket
```
