#!/bin/bash
set -euo pipefail

# ============================================================================
# MCP Bitbucket Benchmark
#
# Compares different Bitbucket MCP servers on a PR review task.
# Measures: token usage, cost, latency, and number of turns.
#
# Prerequisites:
#   - claude CLI installed
#   - For branow: Docker container running on localhost:9847
#   - For stdio servers: npx/uvx available
#   - Environment variables set (see below)
#
# Required env vars:
#   BITBUCKET_EMAIL       - Bitbucket username/email
#   BITBUCKET_API_TOKEN   - Bitbucket app password
#   BENCHMARK_WORKSPACE   - Bitbucket workspace slug
#   BENCHMARK_REPO        - Repository slug
#   BENCHMARK_PR_ID       - Pull request ID to review
#
# Usage:
#   ./benchmark/run.sh                    # run all configs
#   ./benchmark/run.sh branow             # run single config
#   ./benchmark/run.sh branow pdogra1299  # run specific configs
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/configs"
RESULTS_DIR="$SCRIPT_DIR/results"

# Validate required env vars
for var in BITBUCKET_EMAIL BITBUCKET_API_TOKEN BENCHMARK_WORKSPACE BENCHMARK_REPO BENCHMARK_PR_ID; do
  if [ -z "${!var:-}" ]; then
    echo "Error: $var is not set"
    exit 1
  fi
done

WORKSPACE="$BENCHMARK_WORKSPACE"
REPO="$BENCHMARK_REPO"
PR_ID="$BENCHMARK_PR_ID"

MODEL="${BENCHMARK_MODEL:-opus}"

PROMPT="Review pull request #${PR_ID} in workspace '${WORKSPACE}', repository '${REPO}' using bitbucket mcp."

mkdir -p "$RESULTS_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SUMMARY_FILE="$RESULTS_DIR/summary_${TIMESTAMP}.md"

# Determine which configs to run
if [ $# -gt 0 ]; then
  CONFIGS=("$@")
else
  CONFIGS=()
  for f in "$CONFIG_DIR"/*.json; do
    CONFIGS+=("$(basename "$f" .json)")
  done
fi

echo "# Benchmark Results — $(date)" > "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "PR: ${WORKSPACE}/${REPO}#${PR_ID}" >> "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "| Server | Duration (s) | API Duration (s) | Turns | Input Tokens | Output Tokens | Cache Read | Cost (USD) |" >> "$SUMMARY_FILE"
echo "|--------|-------------|-----------------|-------|-------------|--------------|------------|------------|" >> "$SUMMARY_FILE"

for name in "${CONFIGS[@]}"; do
  config="$CONFIG_DIR/${name}.json"

  if [ ! -f "$config" ]; then
    echo "Warning: config $config not found, skipping"
    continue
  fi

  echo "=========================================="
  echo "Running benchmark: $name"
  echo "=========================================="

  result_file="$RESULTS_DIR/${name}_${TIMESTAMP}.json"

  # Run claude with this MCP config
  if claude -p "$PROMPT" \
    --output-format json \
    --model "$MODEL" \
    --mcp-config "$config" \
    --strict-mcp-config \
    --dangerously-skip-permissions \
    --max-turns 20 \
    --max-budget-usd 5.00 \
    --no-session-persistence \
    > "$result_file" 2>"$RESULTS_DIR/${name}_${TIMESTAMP}.log"; then

    # Extract metrics
    duration_ms=$(jq -r '.duration_ms // 0' "$result_file")
    duration_api_ms=$(jq -r '.duration_api_ms // 0' "$result_file")
    num_turns=$(jq -r '.num_turns // 0' "$result_file")
    input_tokens=$(jq -r '.usage.input_tokens // 0' "$result_file")
    output_tokens=$(jq -r '.usage.output_tokens // 0' "$result_file")
    cache_read=$(jq -r '.usage.cache_read_input_tokens // 0' "$result_file")
    cost=$(jq -r '.total_cost_usd // 0' "$result_file")

    duration_s=$(echo "scale=1; $duration_ms / 1000" | bc)
    duration_api_s=$(echo "scale=1; $duration_api_ms / 1000" | bc)

    echo "  Duration:     ${duration_s}s"
    echo "  API Duration: ${duration_api_s}s"
    echo "  Turns:        $num_turns"
    echo "  Input tokens: $input_tokens"
    echo "  Output tokens: $output_tokens"
    echo "  Cache read:   $cache_read"
    echo "  Cost:         \$${cost}"

    echo "| $name | $duration_s | $duration_api_s | $num_turns | $input_tokens | $output_tokens | $cache_read | \$$cost |" >> "$SUMMARY_FILE"

    # Save the review text separately
    jq -r '.result // "no result"' "$result_file" > "$RESULTS_DIR/${name}_${TIMESTAMP}_review.md"
  else
    echo "  FAILED — see $RESULTS_DIR/${name}_${TIMESTAMP}.log"
    echo "| $name | FAILED | — | — | — | — | — | — |" >> "$SUMMARY_FILE"
  fi

  echo ""
done

echo ""
echo "=========================================="
echo "Summary saved to: $SUMMARY_FILE"
echo "=========================================="
echo ""
cat "$SUMMARY_FILE"
