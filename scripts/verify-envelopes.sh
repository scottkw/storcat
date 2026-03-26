#!/usr/bin/env bash
# verify-envelopes.sh — Verifies that all wailsAPI wrappers in wailsAPI.ts
# return {success, ...} envelopes and no wrapper uses raw returns.
# Addresses API-03.
#
# Exit code: 0 = all checks pass, 1 = one or more checks failed.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET="$REPO_ROOT/frontend/src/services/wailsAPI.ts"

failures=0

check() {
  local description="$1"
  local expected="$2"
  local actual="$3"
  local comparison="$4"  # "gte" | "eq" | "zero"

  local pass=false
  case "$comparison" in
    gte)  [ "$actual" -ge "$expected" ] && pass=true ;;
    eq)   [ "$actual" -eq "$expected" ] && pass=true ;;
    zero) [ "$actual" -eq 0 ]           && pass=true ;;
  esac

  if $pass; then
    echo "  PASS  $description (got $actual)"
  else
    echo "  FAIL  $description (got $actual, expected $comparison $expected)"
    failures=$((failures + 1))
  fi
}

echo "Verifying envelope compliance in: $TARGET"
echo ""

# Count wrappers: each async arrow function in the wailsAPI object = one wrapper.
# We count by "async (" lines inside the object to detect the 17 wrappers plus
# their success returns.
success_true_count=$(grep -c "return { success: true" "$TARGET" || true)
success_false_count=$(grep -c "return { success: false" "$TARGET" || true)
total_success_returns=$((success_true_count + success_false_count))

# Each of the 17 wrappers must have at least one success: true and one success: false.
# Minimum: 17 success:true + 17 success:false = 34 total, but plan acceptance criteria
# says >= 24 (some wrappers share branches). Use >= 24 as stated in the plan.
check "success:true return count >= 17 (one per wrapper)" 17 "$success_true_count" "gte"
check "success:false return count >= 17 (one per wrapper)" 17 "$success_false_count" "gte"
check "total {success:...} returns >= 34" 34 "$total_success_returns" "gte"

# Forbidden patterns — none of these should appear in any wrapper.
raw_result_count=$(grep -c "return result;" "$TARGET" || true)
check "no raw 'return result;' passthrough" 0 "$raw_result_count" "zero"

raw_null_count=$(grep -c "return null;" "$TARGET" || true)
check "no 'return null;' raw returns" 0 "$raw_null_count" "zero"

raw_await_count=$(grep -c "return await Get" "$TARGET" || true)
check "no raw 'return await Get...' passthrough" 0 "$raw_await_count" "zero"

# Spot-check: specific field names must appear (one per key wrapper).
htmlpath_count=$(grep -c "return { success: true.*htmlPath" "$TARGET" || true)
check "getCatalogHtmlPath returns htmlPath field" 1 "$htmlpath_count" "gte"

content_count=$(grep -c "return { success: true.*content" "$TARGET" || true)
check "readHtmlFile returns content field" 1 "$content_count" "gte"

path_count=$(grep -c "return { success: true.*path" "$TARGET" || true)
check "selectDirectory wrappers return path field" 1 "$path_count" "gte"

config_count=$(grep -c "return { success: true.*config" "$TARGET" || true)
check "getConfig returns config field" 1 "$config_count" "gte"

enabled_count=$(grep -c "return { success: true.*enabled" "$TARGET" || true)
check "getWindowPersistence returns enabled field" 1 "$enabled_count" "gte"

echo ""
if [ "$failures" -eq 0 ]; then
  echo "All envelope checks passed."
  exit 0
else
  echo "$failures check(s) failed."
  exit 1
fi
