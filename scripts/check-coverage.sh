#!/bin/bash
# Coverage gate enforcement script
# Used in CI to enforce coverage standards

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Coverage thresholds
MIN_TOTAL_COVERAGE=25      # Absolute minimum (fail build)
TARGET_TOTAL_COVERAGE=80   # Gold standard target
MIN_CRITICAL_COVERAGE=80   # Minimum for critical packages (after Phase 2)

# Critical packages that must meet higher standards
CRITICAL_PACKAGES=(
    "pkg/plugin/sdk/v1"
    "pkg/plugin/sdk"
    "internal/cli"
    "pkg/prompt"
)

# High priority packages
HIGH_PRIORITY_PACKAGES=(
    "internal/config"
    "pkg/errors"
    "pkg/output"
)

echo -e "${GREEN}=== Coverage Gate Enforcement ===${NC}"
echo ""

# Check if coverage file exists
if [ ! -f "coverage.out" ]; then
    echo -e "${RED}✗ coverage.out not found${NC}"
    echo "  Run 'go test -coverprofile=coverage.out ./...' first"
    exit 1
fi

# Get total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}' | sed 's/%//')
echo -e "${YELLOW}Total Coverage: ${TOTAL_COVERAGE}%${NC}"

# Check minimum coverage
if (( $(echo "$TOTAL_COVERAGE < $MIN_TOTAL_COVERAGE" | bc -l) )); then
    echo -e "${RED}✗ FAIL: Coverage ${TOTAL_COVERAGE}% is below minimum ${MIN_TOTAL_COVERAGE}%${NC}"
    exit 1
fi

# Check target coverage (warning, not error)
if (( $(echo "$TOTAL_COVERAGE < $TARGET_TOTAL_COVERAGE" | bc -l) )); then
    echo -e "${YELLOW}⚠️  Coverage ${TOTAL_COVERAGE}% is below target ${TARGET_TOTAL_COVERAGE}%${NC}"
    echo -e "${YELLOW}   Continue improving test coverage to meet gold standard.${NC}"
    TARGET_MET=false
else
    echo -e "${GREEN}✓ Coverage meets gold standard target!${NC}"
    TARGET_MET=true
fi

echo ""
echo -e "${YELLOW}=== Per-Package Coverage ===${NC}"

# Function to get package coverage
get_package_coverage() {
    local package=$1
    local coverage=$(go tool cover -func=coverage.out | grep "^github.com/ivannovak/glide/v2/${package}/" | \
        awk '{sum += $3; count++} END {if (count > 0) printf "%.1f", sum/count; else print "0"}' | sed 's/%//')
    echo "$coverage"
}

# Check critical packages
echo -e "${YELLOW}Critical Packages (Target: ${MIN_CRITICAL_COVERAGE}%):${NC}"
CRITICAL_FAILURES=0

for pkg in "${CRITICAL_PACKAGES[@]}"; do
    coverage=$(get_package_coverage "$pkg")

    if [ "$coverage" == "0" ]; then
        echo -e "  ${pkg}: ${RED}No coverage data${NC}"
        continue
    fi

    if (( $(echo "$coverage < $MIN_CRITICAL_COVERAGE" | bc -l) )); then
        echo -e "  ${pkg}: ${RED}${coverage}%${NC} (below target ${MIN_CRITICAL_COVERAGE}%)"
        CRITICAL_FAILURES=$((CRITICAL_FAILURES + 1))
    else
        echo -e "  ${pkg}: ${GREEN}${coverage}%${NC} ✓"
    fi
done

# Check high priority packages
echo ""
echo -e "${YELLOW}High Priority Packages (Target: ${TARGET_TOTAL_COVERAGE}%):${NC}"
HIGH_PRIORITY_FAILURES=0

for pkg in "${HIGH_PRIORITY_PACKAGES[@]}"; do
    coverage=$(get_package_coverage "$pkg")

    if [ "$coverage" == "0" ]; then
        echo -e "  ${pkg}: ${RED}No coverage data${NC}"
        continue
    fi

    if (( $(echo "$coverage < $TARGET_TOTAL_COVERAGE" | bc -l) )); then
        echo -e "  ${pkg}: ${YELLOW}${coverage}%${NC} (below target ${TARGET_TOTAL_COVERAGE}%)"
        HIGH_PRIORITY_FAILURES=$((HIGH_PRIORITY_FAILURES + 1))
    else
        echo -e "  ${pkg}: ${GREEN}${coverage}%${NC} ✓"
    fi
done

echo ""
echo -e "${YELLOW}=== Coverage Gate Summary ===${NC}"

# Summary
PASSED=true

if [ "$CRITICAL_FAILURES" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  ${CRITICAL_FAILURES} critical package(s) below target${NC}"
    echo -e "${YELLOW}   This is expected during Phase 2 development${NC}"
fi

if [ "$HIGH_PRIORITY_FAILURES" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  ${HIGH_PRIORITY_FAILURES} high-priority package(s) below target${NC}"
fi

if [ "$TARGET_MET" = true ] && [ "$CRITICAL_FAILURES" -eq 0 ] && [ "$HIGH_PRIORITY_FAILURES" -eq 0 ]; then
    echo -e "${GREEN}✓ All coverage gates passed!${NC}"
else
    echo -e "${YELLOW}⚠️  Coverage gates: PASS (minimum met, working toward target)${NC}"
fi

# Always pass during Phase 2 development (as long as minimum is met)
# After Phase 2, we'll enforce the 80% target
exit 0
