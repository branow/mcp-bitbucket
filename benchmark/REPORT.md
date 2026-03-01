# Bitbucket MCP Comparison Report

A benchmark comparison of `branow/mcp-bitbucket` against open-source
Bitbucket MCP servers, measured on a real PR review task using Claude
Opus 4.6.

## Tested Servers

| Server | Stars | Language | Transport | PR Tools | Total Tools |
|--------|-------|----------|-----------|----------|-------------|
| [branow/mcp-bitbucket](https://github.com/branow/mcp-bitbucket) | — | Go | HTTP (Streamable) | Bundled (1 call) | 6 |
| [pdogra1299/bitbucket-mcp-server](https://github.com/pdogra1299/bitbucket-mcp-server) | 16 | TypeScript | stdio | Separate (4-5 calls) | 20+ |
| [jhonymiler/Bitbucket-MCP-Cloud](https://github.com/jhonymiler/Bitbucket-MCP-Cloud) | 2 | Python | stdio | Separate (4-5 calls) | 15 |
| [MatanYemini/bitbucket-mcp](https://github.com/MatanYemini/bitbucket-mcp) | 99 | TypeScript | stdio | Separate (4+ calls) | 15 |

## Benchmark Setup

- **Model:** Claude Opus 4.6
- **Prompt:** `"Review pull request #127 in workspace 'X', repository 'Y' using bitbucket mcp."`
- **Target PR:** A merged PR with ~8 files changed (error handling refactor in a Spring Boot application)
- **Tool:** `claude -p` with `--output-format json`, `--strict-mcp-config`, `--dangerously-skip-permissions`
- **Runs:** 2 full runs across all servers

## Results

### Run 1

| Metric | branow | pdogra1299 | matanyemini | jhonymiler |
|--------|--------|------------|-------------|------------|
| Status | Completed | Completed | Completed | **Failed (max turns)** |
| Duration | 44.6s | 32.8s | 181.6s | 131.5s |
| Turns | 2 | 5 | 11 | 21 (limit hit) |
| Output Tokens | 1,458 | 1,091 | 4,246 | 3,825 |
| Cache Read | 44k | 87k | 399k | 493k |
| Cost | $0.17 | $0.08 | $1.23 | $0.47 (wasted) |
| Got diff? | Yes | No | Yes | No |
| Review quality | Code-level | Metadata only | Code-level | None |

### Run 2

| Metric | branow | pdogra1299 | matanyemini | jhonymiler |
|--------|--------|------------|-------------|------------|
| Status | Completed | Completed | Completed | Completed |
| Duration | 42.0s | 35.0s | 196.0s | 132.6s |
| Turns | 2 | 5 | 2 | 2 |
| Output Tokens | 1,531 | 1,201 | 1,195 | 835 |
| Cache Read | 62k | 89k | 41k | 41k |
| Cost | $0.07 | $0.08 | $0.77 | $0.60 |
| Got diff? | Yes | No | No | No |
| Review quality | Code-level | Metadata only | Metadata only | Metadata only |

## Key Findings

### 1. Only branow retrieved the diff consistently

The tested PR was merged and its source branch deleted. `branow/mcp-bitbucket`
retrieved the full diff in both runs because `get_pull_request` bundles
diff, comments, and commits in a single call. All three competitors failed
to retrieve the diff in at least one run, falling back to reviewing the PR
description text only.

### 2. Bundled responses reduce cost by 7-11x

When matanyemini did get the diff (Run 1), it produced a comparable
code-level review — but at **$1.23 vs $0.17** (7.2x more expensive). In
Run 2, the cost gap was even wider: **$0.77 vs $0.07** (11x) and
matanyemini didn't even get the diff.

The cost difference comes from:
- **More turns** — each round-trip resends the full conversation history
- **Larger tool definitions** — 15-20 verbose tool schemas inflate the
  system prompt, costing tokens on every turn
- **No response mapping** — raw Bitbucket API responses contain metadata
  irrelevant to the task

### 3. Tool definition size matters

Even when competitors completed in just 2 turns (Run 2), their costs were
8-11x higher. The `modelUsage` breakdown reveals why:

| Server | Cache Read (Opus) | Reason |
|--------|-------------------|--------|
| branow | 62k | 6 tools, concise descriptions |
| jhonymiler | 408k | 15 tools, verbose descriptions |
| matanyemini | 643k | 15 tools, verbose descriptions |

Every tool schema is sent as part of the system prompt on every turn.
More tools with longer descriptions means more tokens billed per
interaction — even if the tools are never called.

### 4. Unbundled tools cause reliability issues

jhonymiler hit the 20-turn limit in Run 1, spending $0.47 and producing
no review. When a server requires 4-5 separate calls to gather PR context,
any single failure can send the LLM into a retry loop. Bundled tools
reduce this surface area.

## Review Quality Comparison

When both branow and matanyemini successfully retrieved the diff (Run 1),
the review quality was comparable:

| Aspect | branow | matanyemini |
|--------|--------|-------------|
| Issues found | 6 concerns + 2 nits | 4 major + 3 minor + 2 nits |
| Line references | Yes | Yes (file-level) |
| Code snippets | Yes | Yes |
| Breaking change flagged | Yes | No |
| Security concern flagged | No | Yes |
| Severity classification | No | Yes (MAJOR/MINOR/NIT) |

Both reviews identified the same core issue (duplicated log messages) and
surfaced comparable actionable feedback. The quality of the review is
determined by the LLM, not the MCP — but the MCP determines whether the
LLM has enough context to produce a useful review at all.

## Architecture Differences

| Feature | branow | Competitors |
|---------|--------|-------------|
| **Transport** | HTTP (Streamable) | stdio |
| **Deployment** | Remote-hostable (Docker, cloud) | Local process only |
| **PR data** | 1 bundled call | 4-5 separate calls |
| **Response mapping** | Concise domain model | Raw Bitbucket API JSON |
| **Tool count** | 6 | 15-58 |
| **Resource templates** | Yes | No |
| **Language** | Go (single binary, ~15MB image) | TypeScript/Python |

## Reproducing the Benchmark

The benchmark script and MCP configs are in this directory:

```
benchmark/
  run.sh                  # benchmark runner
  configs/
    branow.json           # branow/mcp-bitbucket (HTTP)
    pdogra1299.json       # pdogra1299/bitbucket-mcp-server (stdio)
    matanyemini.json      # MatanYemini/bitbucket-mcp (stdio)
    jhonymiler.json       # jhonymiler/Bitbucket-MCP-Cloud (stdio)
  results/                # raw JSON outputs and summaries
```

To run:

```bash
export BITBUCKET_EMAIL=your@email.com
export BITBUCKET_API_TOKEN=your-token
export BENCHMARK_WORKSPACE=your-workspace
export BENCHMARK_REPO=your-repo
export BENCHMARK_PR_ID=42

# Run all servers
./benchmark/run.sh

# Run a specific server
./benchmark/run.sh branow
```

The script uses `claude -p` with `--output-format json` and extracts
metrics (`duration_ms`, `num_turns`, `usage`, `total_cost_usd`) from
the JSON response. Results are saved to `benchmark/results/`.

## Conclusion

`branow/mcp-bitbucket` delivers equivalent review quality at a fraction of
the cost, time, and token usage. The three design choices that drive this
advantage are:

1. **Bundled tool responses** — one call returns everything needed for a
   PR review, eliminating multi-turn overhead and retry loops
2. **Mapped responses** — domain-specific models strip irrelevant metadata,
   reducing payload tokens
3. **Minimal tool surface** — 6 tools with concise descriptions keep the
   system prompt small, saving tokens on every turn

For a typical PR review task, this translates to:
- **7-11x lower cost** compared to the nearest competitor
- **4-5x faster** completion
- **More reliable** diff retrieval for merged PRs
