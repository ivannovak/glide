#!/bin/bash
# Benchmark comparison script for Glide
#
# Usage:
#   ./scripts/benchmark-compare.sh [baseline_file]
#
# This script runs benchmarks and compares them against a baseline.
# If no baseline is provided, it uses benchmarks.txt in the repo root.
#
# Requirements:
#   - Go installed
#   - benchstat tool (go install golang.org/x/perf/cmd/benchstat@latest)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_BASELINE="${REPO_ROOT}/benchmarks.txt"
THRESHOLD_PERCENT=15  # Regression threshold percentage
BENCHMARK_TIME="1s"
BENCHMARK_COUNT=5

# Parse arguments
BASELINE="${1:-$DEFAULT_BASELINE}"
TEMP_RESULTS=$(mktemp)

echo "=== Glide Benchmark Comparison ==="
echo ""

# Check for benchstat
if ! command -v benchstat &> /dev/null; then
    echo -e "${YELLOW}Warning: benchstat not found. Installing...${NC}"
    go install golang.org/x/perf/cmd/benchstat@latest
fi

# Check for baseline
if [ ! -f "$BASELINE" ]; then
    echo -e "${YELLOW}No baseline found at $BASELINE${NC}"
    echo "Creating new baseline..."
    echo ""

    go test -bench=. -benchmem -benchtime=$BENCHMARK_TIME -count=$BENCHMARK_COUNT \
        ./tests/benchmarks/... 2>&1 | grep -E "^Bench" > "$DEFAULT_BASELINE" || true

    echo -e "${GREEN}Baseline created at $DEFAULT_BASELINE${NC}"
    echo "Run this script again to compare against the baseline."
    exit 0
fi

echo "Running benchmarks..."
echo ""

# Run benchmarks and capture results
go test -bench=. -benchmem -benchtime=$BENCHMARK_TIME -count=$BENCHMARK_COUNT \
    ./tests/benchmarks/... 2>&1 | grep -E "^Bench" > "$TEMP_RESULTS" || true

echo "Comparing against baseline: $BASELINE"
echo ""
echo "=== Results ==="
echo ""

# Run benchstat comparison
benchstat "$BASELINE" "$TEMP_RESULTS"

echo ""
echo "=== Summary ==="
echo ""

# Check for significant regressions
REGRESSIONS=$(benchstat "$BASELINE" "$TEMP_RESULTS" 2>/dev/null | grep -E '\+[0-9]+%' || true)

if [ -n "$REGRESSIONS" ]; then
    # Check if any regression exceeds threshold
    HAS_CRITICAL=false
    while IFS= read -r line; do
        # Extract percentage from line
        PERCENT=$(echo "$line" | grep -oE '\+[0-9]+%' | head -1 | tr -d '+%')
        if [ -n "$PERCENT" ] && [ "$PERCENT" -ge "$THRESHOLD_PERCENT" ]; then
            HAS_CRITICAL=true
            break
        fi
    done <<< "$REGRESSIONS"

    if [ "$HAS_CRITICAL" = true ]; then
        echo -e "${RED}CRITICAL: Performance regressions detected (>${THRESHOLD_PERCENT}%)${NC}"
        echo ""
        echo "Regressions found:"
        echo "$REGRESSIONS"
        echo ""
        echo "Consider investigating before merging."
        rm "$TEMP_RESULTS"
        exit 1
    else
        echo -e "${YELLOW}Minor performance changes detected (<${THRESHOLD_PERCENT}%)${NC}"
        echo ""
        echo "Changes:"
        echo "$REGRESSIONS"
    fi
else
    echo -e "${GREEN}No significant regressions detected${NC}"
fi

# Cleanup
rm "$TEMP_RESULTS"

echo ""
echo "=== Recommendations ==="
echo ""
echo "To update the baseline after intentional changes:"
echo "  go test -bench=. -benchmem -benchtime=$BENCHMARK_TIME -count=$BENCHMARK_COUNT \\"
echo "    ./tests/benchmarks/... 2>&1 | grep -E '^Bench' > benchmarks.txt"
echo ""
echo "Performance targets:"
echo "  - Context detection: <100ms (current: ~75ms)"
echo "  - Config load: <50ms (current: ~26ms)"
echo "  - Config merge (multiple): <100ms (current: ~123ms)"
echo "  - Plugin discovery: <500ms (currently needs optimization: ~1.3s)"
echo "  - Startup time: <300ms (current: ~260ms)"
