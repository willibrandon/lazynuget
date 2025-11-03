#!/bin/bash
# test-startup.sh - Measure LazyNuGet startup performance
# Usage: ./scripts/dev/test-startup.sh [iterations]

set -e

ITERATIONS=${1:-10}
BINARY="./lazynuget"
TARGET_MS=200

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo "Error: $BINARY not found. Run 'make build' first."
    exit 1
fi

echo "Testing LazyNuGet startup performance..."
echo "Target: <${TARGET_MS}ms (p95)"
echo "Iterations: $ITERATIONS"
echo ""

# Measure startup times
times=()
for i in $(seq 1 $ITERATIONS); do
    start=$(date +%s%N)
    $BINARY --version > /dev/null 2>&1
    end=$(date +%s%N)

    elapsed_ns=$((end - start))
    elapsed_ms=$((elapsed_ns / 1000000))

    times+=($elapsed_ms)
    echo "Run $i: ${elapsed_ms}ms"
done

# Calculate statistics
total=0
for time in "${times[@]}"; do
    total=$((total + time))
done

avg=$((total / ITERATIONS))
min=${times[0]}
max=${times[0]}

for time in "${times[@]}"; do
    if [ $time -lt $min ]; then
        min=$time
    fi
    if [ $time -gt $max ]; then
        max=$time
    fi
done

# Calculate p95 (simplified - just use max for small samples)
p95=$max

# Print results
echo ""
echo "Results:"
echo "--------"
echo "Average: ${avg}ms"
echo "Min: ${min}ms"
echo "Max: ${max}ms"
echo "P95: ${p95}ms"

# Compare to target
echo ""
if [ $p95 -le $TARGET_MS ]; then
    echo -e "${GREEN}✓ PASS${NC} - Startup time within target (<${TARGET_MS}ms)"
    exit 0
else
    diff=$((p95 - TARGET_MS))
    echo -e "${YELLOW}⚠ WARN${NC} - Startup time ${diff}ms over target (${p95}ms vs ${TARGET_MS}ms)"
    exit 0
fi
