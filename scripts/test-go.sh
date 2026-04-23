#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
PROFILE_FILE="$ROOT_DIR/backend-coverage.out"
MIN_COVERAGE="${MIN_GO_COVERAGE:-30}"

echo "Running Go tests..."
cd "$BACKEND_DIR"

# Exclude vendor and generated protobuf modules.
ALL_PKGS=$(go list ./... | grep -v '/vendor/' | grep -v '/internal/gen/')

# 1) Always run all package tests (compile+test sweep)
go test $ALL_PKGS

# 2) Coverage gate only for critical business packages with active assertions.
# Plan-only stub tests (t.Skip) are intentionally excluded from gate metrics.
CRITICAL_PKGS=(
  "./auth-service/service"
  "./diary-service/service"
  "./internal/pkg/config"
  "./internal/pkg/middleware"
  "./search-service/service"
)

go test -covermode=atomic -coverprofile="$PROFILE_FILE" "${CRITICAL_PKGS[@]}"

TOTAL_COVERAGE=$(go tool cover -func="$PROFILE_FILE" | awk '/^total:/{print substr($3, 1, length($3)-1)}')
echo "Total Go coverage: ${TOTAL_COVERAGE}%"

awk -v cov="$TOTAL_COVERAGE" -v min="$MIN_COVERAGE" 'BEGIN {
  if (cov+0 < min+0) {
    printf("Coverage gate failed: %.2f%% < %.2f%%\n", cov, min)
    exit 1
  }
}'

echo "Go tests passed."
