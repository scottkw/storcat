#!/usr/bin/env bash
# gen-checksums.sh — Emits a flat, sorted checksums.txt (sha256, one line per
# artifact) from a directory of release artifacts.
# Addresses CHECK-01, CHECK-02.
#
# Exit code: 0 = checksums.txt written, 1 = a precondition or guard failed.
#
# NOTE: unlike every other script in this directory, gen-checksums.sh does NOT
# self-locate to the repository root. Its target directory arrives as "$1" from
# two callers with different working directories — the `release` job passes
# `artifacts`, and the v3.0.0 backfill passes a scratch download directory — so
# self-locating would silently reinterpret those caller-relative paths. The
# divergence from the sibling scripts in this directory is a decision, not an
# omission.
#
# Usage: gen-checksums.sh <artifact-directory>
#        checksums.txt is written into the CURRENT WORKING DIRECTORY, never
#        into the directory being hashed.

set -euo pipefail

# `sort` is locale-sensitive and this release mixes two naming families
# (StorCat-v3.0.0-… capital/hyphen and storcat_3.0.0_… lowercase/underscore).
# Under a UTF-8 collation the two interleave differently than under C, so the
# ubuntu-22.04 runner and the macOS dev machine would produce different files
# from identical inputs and the "one script, same output for both callers"
# property this file exists to provide would look broken. Pin the collation so
# the output is reproducible across both callers.
LC_ALL=C
export LC_ALL

OUTPUT_NAME="checksums.txt"
TAB=$(printf '\t')

# Scratch directory for the sort key file. Global, not a function local: the
# EXIT trap body is expanded when the shell exits, by which point a local has
# gone out of scope and `set -u` would abort inside the trap.
SCRATCH_DIR=""
cleanup_scratch() { [ -n "$SCRATCH_DIR" ] && rm -rf "$SCRATCH_DIR"; return 0; }
trap cleanup_scratch EXIT

# Select the hasher at run time. The ubuntu-22.04 runner has sha256sum (GNU
# coreutils); macOS always has /usr/bin/shasum (Perl Digest::SHA) and only
# recently gained sha256sum. Their output is byte-identical, so selection costs
# four lines and changes nothing downstream. Finding neither is a hard failure:
# substituting another digest here would be the most damaging silent default in
# the pipeline.
require_hash_tool() {
  if command -v sha256sum >/dev/null 2>&1; then
    hash_cmd() { sha256sum "$@"; }
  elif command -v shasum >/dev/null 2>&1; then
    hash_cmd() { shasum -a 256 "$@"; }
  else
    echo "Error: no sha256 tool found (looked for: sha256sum, shasum)" >&2
    echo "Install GNU coreutils (sha256sum) or Perl's shasum, then re-run." >&2
    exit 1
  fi
}

require_input_dir() {
  local dir="${1:-}"
  if [ -z "$dir" ]; then
    echo "Error: expected a directory of artifacts as the first argument, received none" >&2
    echo "Usage: gen-checksums.sh <artifact-directory>" >&2
    exit 1
  fi
  if [ ! -d "$dir" ]; then
    echo "Error: expected a directory of artifacts, received: $dir" >&2
    echo "Pass the directory that holds the release artifacts." >&2
    exit 1
  fi
}

# No guard below substitutes a default for input it could not read. An
# integrity file that ratifies a short or mangled artifact set is worse than no
# file at all: it converts a visibly short release into an apparently-verified
# one. Every guard names what it searched for and exits 1.

# An empty input directory is an error, not an empty output file.
assert_dir_not_empty() {
  local dir="$1"
  local pairs="$2"
  if [ ! -s "$pairs" ]; then
    echo "Error: no regular file found in the artifact directory: $dir" >&2
    echo "Pass a directory that holds the release artifacts." >&2
    exit 1
  fi
}

# Refuse to write checksums.txt into the tree being hashed. This makes "never
# hash your own output" structural rather than order-dependent, and it keeps the
# count guard's arithmetic honest — with the output outside the tree, files
# found and lines written compare directly with no self-exclusion carve-out.
assert_output_outside_input() {
  local dir="$1"
  local out="$2"
  local dir_abs out_dir
  dir_abs=$(cd "$dir" && pwd -P)
  out_dir=$(cd "$(dirname "$out")" && pwd -P)
  case "$out_dir" in
    "$dir_abs"|"$dir_abs"/*)
      echo "Error: refusing to write $OUTPUT_NAME inside the directory being hashed: $dir_abs" >&2
      echo "Run gen-checksums.sh from a directory outside the artifact tree." >&2
      exit 1
      ;;
  esac
}

# Two files anywhere in the tree that flatten to the same basename are a
# collision, not a merge. The count guard structurally cannot see this: the
# count still matches, and `shasum -c` then checks one file against two
# different hashes and reports FAILED.
assert_no_duplicate_basenames() {
  local pairs="$1"
  local dupes
  dupes=$(cut -f1 "$pairs" | sort | uniq -d | tr '\n' ' ')
  dupes=${dupes% }
  if [ -n "$dupes" ]; then
    echo "Error: two or more artifacts flatten to the same basename: $dupes" >&2
    echo "Rename the colliding artifact so every basename in the release is distinct." >&2
    exit 1
  fi
}

# Both integers are derived from the input tree at run time. Hardcoding the
# expected artifact count would fail the release on the next legitimate platform
# change until someone found and bumped the literal.
assert_count_matches() {
  local found="$1"
  local written="$2"
  if [ "$found" -ne "$written" ]; then
    echo "Error: found $found artifact file(s) but wrote $written checksum line(s)" >&2
    echo "Refusing to publish an integrity file that does not cover every artifact." >&2
    exit 1
  fi
}

# Walk the input directory recursively — actions/download-artifact with
# `path: artifacts/` yields artifacts/<job-artifact-name>/<file>, never a flat
# directory — recording `<basename><TAB><path>` per regular file. The sort key
# is carried on a tab-delimited scratch line so the sort acts on the filename
# and nothing else. Two failure modes are designed out here: sorting the
# finished hash-first lines sorts by hash rather than by filename, and a naive
# whitespace-field sort breaks on any filename containing a space.
collect() {
  local dir="$1"
  local pairs="$2"
  local f

  : > "$pairs"
  while IFS= read -r -d '' f; do
    printf '%s%s%s\n' "${f##*/}" "$TAB" "$f" >> "$pairs"
  done < <(find "$dir" -type f -print0)
}

# Emit one `<64 lowercase hex><two spaces><basename>` line per collected file,
# sorted by the basename field alone under LC_ALL=C. The last byte is a single
# newline and there is no empty final line: a blank line is the one non-hash
# line that makes `shasum -c` print `WARNING: 1 line is improperly formatted`.
generate() {
  local pairs="$1"
  local out="$2"
  local base path hash

  : > "$out"
  while IFS="$TAB" read -r base path; do
    hash=$(hash_cmd "$path")
    printf '%s  %s\n' "${hash%% *}" "$base" >> "$out"
  done < <(sort -t "$TAB" -k1,1 "$pairs")
}

main() {
  local dir out pairs

  case "${1:-}" in
    --self-test)
      self_test && exit 0
      exit 1
      ;;
  esac

  require_input_dir "${1:-}"
  require_hash_tool

  dir="$1"
  out="$PWD/$OUTPUT_NAME"

  SCRATCH_DIR=$(mktemp -d)
  pairs="$SCRATCH_DIR/pairs"

  # Guard order is deliberate: nothing is hashed until the input is known good,
  # nothing is written until the output location is known safe, and a basename
  # collision is caught on the collected list so it costs nothing to detect.
  collect "$dir" "$pairs"
  assert_dir_not_empty "$dir" "$pairs"
  assert_output_outside_input "$dir" "$out"
  assert_no_duplicate_basenames "$pairs"

  generate "$pairs" "$out"

  # Derived at run time from both sides of the transform, after the write.
  assert_count_matches "$(awk 'END{print NR}' "$pairs")" "$(awk 'END{print NR}' "$out")"

  # Echo the result into the run log — the cheapest CHECK-01 evidence available
  # in a CI run nobody is watching. distribute.yml establishes this habit.
  echo "Wrote $out"
  cat "$out"
}

# ---------------------------------------------------------------------------
# Self-test. Everything below this point is test machinery and never runs on
# the production path.
#
# self_test() is deliberately defined BELOW the entire production path — below
# require_hash_tool, require_input_dir, every guard and generate. That layout is
# load-bearing: the "artifact count is derived, never hardcoded" gate scans
# everything above this definition, and the fixtures below deliberately carry
# 9.9.9 version strings and literal guard arguments.
# ---------------------------------------------------------------------------
self_test() {
  local fix work names lastbyte nr hashn named rc rc_noarg rc_file err s1 s2

  SCRATCH_DIR=$(mktemp -d)   # removed by the global EXIT trap
  SELF="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/$(basename "${BASH_SOURCE[0]}")"
  failures=0
  assertions=0

  echo "gen-checksums.sh --self-test"
  echo ""

  # 1. FLATTEN — nested files emit flat basenames.
  fix="$SCRATCH_DIR/flatten"
  mkdir -p "$fix/a" "$fix/b"
  printf 'one\n' > "$fix/a/one.bin"
  printf 'two\n' > "$fix/b/two.bin"
  _st_run "$fix"
  names=$(awk '{print $2}' "$st_out" | tr '\n' ' '); names=${names% }
  check "FLATTEN nested files emit flat basenames" \
    "one.bin two.bin" "$names" str

  # 2. SORT ORDER — both naming families land in one fixed LC_ALL=C order.
  fix="$SCRATCH_DIR/sortorder"
  mkdir -p "$fix/mac" "$fix/linux" "$fix/deb"
  printf 'dmg\n' > "$fix/mac/StorCat-v9.9.9-darwin-universal.dmg"
  printf 'tgz\n' > "$fix/linux/StorCat-v9.9.9-linux-amd64.tar.gz"
  printf 'deb\n' > "$fix/deb/storcat_9.9.9_amd64.deb"
  _st_run "$fix"
  names=$(awk '{print $2}' "$st_out" | tr '\n' ' '); names=${names% }
  check "SORT ORDER uppercase family first, exact sequence" \
    "StorCat-v9.9.9-darwin-universal.dmg StorCat-v9.9.9-linux-amd64.tar.gz storcat_9.9.9_amd64.deb" \
    "$names" str

  # 3. NO TRAILING BLANK LINE — reuses assertion 2's three-file output.
  lastbyte=$(tail -c 1 "$st_out" | od -An -c | tr -d ' ')
  nr=$(awk 'END{print NR}' "$st_out")
  hashn=$(awk '/^[0-9a-f]{64}  [^ ]/{n++} END{print n+0}' "$st_out")
  check "NO TRAILING BLANK LINE last byte is newline, every line is a hash line" \
    "lastbyte=\\n lines=3 hashlines=3" \
    "lastbyte=$lastbyte lines=$nr hashlines=$hashn" str

  # 4. KNOWN HASH — catches a wrong algorithm or flag with no network involved.
  fix="$SCRATCH_DIR/knownhash"
  mkdir -p "$fix"
  : > "$fix/empty.bin"
  _st_run "$fix"
  check "KNOWN HASH zero-byte file hashes to the published empty-string SHA-256" \
    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  empty.bin" \
    "$(cat "$st_out")" str

  # 5. EMPTY DIR — an error, not an empty output file.
  fix="$SCRATCH_DIR/emptydir"
  mkdir -p "$fix"
  _st_run "$fix"
  named=no; case "$st_err" in *"$fix"*) named=yes ;; esac
  check "EMPTY DIR exits 1 and names the directory searched" \
    "rc=1 names_dir=yes" "rc=$st_rc names_dir=$named" str
  echo "        stderr: $st_err"

  # 6. BAD ARGUMENT — missing, and a regular file instead of a directory.
  printf 'not a directory\n' > "$SCRATCH_DIR/regular.file"
  rc_noarg=0
  ( cd "$SCRATCH_DIR" && "$BASH" "$SELF" ) >/dev/null 2>&1 || rc_noarg=$?
  rc_file=0
  ( cd "$SCRATCH_DIR" && "$BASH" "$SELF" "$SCRATCH_DIR/regular.file" ) >/dev/null 2>&1 || rc_file=$?
  check "BAD ARGUMENT missing argument and non-directory argument both exit 1" \
    "noarg=1 file=1" "noarg=$rc_noarg file=$rc_file" str

  # 7. COUNT MISMATCH — the guard fires on unequal integers.
  rc=0
  ( assert_count_matches 9 8 ) >/dev/null 2>&1 || rc=$?
  check "COUNT MISMATCH assert_count_matches 9 8 exits 1" 1 "$rc" eq

  # 8. COUNT MATCH — the floor under assertion 7. If the guard is renamed or its
  #    call shape changes, this row goes red too, so 7 cannot pass for the wrong
  #    reason (a missing function also "exits non-zero").
  rc=0
  ( assert_count_matches 9 9 ) >/dev/null 2>&1 || rc=$?
  check "COUNT MATCH assert_count_matches 9 9 exits 0" 0 "$rc" eq

  # 9. DUPLICATE BASENAME — the case the count guard structurally cannot see.
  fix="$SCRATCH_DIR/dupes"
  mkdir -p "$fix/x" "$fix/y"
  printf 'x\n' > "$fix/x/dup.bin"
  printf 'y\n' > "$fix/y/dup.bin"
  _st_run "$fix"
  named=no; case "$st_err" in *dup.bin*) named=yes ;; esac
  check "DUPLICATE BASENAME exits 1 and names the colliding basename" \
    "rc=1 names_dup=yes" "rc=$st_rc names_dup=$named" str
  echo "        stderr: $st_err"

  # 10. NO HASH TOOL — PATH is stripped for the CHILD ONLY. Stripping it for the
  #     self-test process would lose mktemp/find/rm and collapse every later row
  #     for the wrong reason.
  work=$(mktemp -d "$SCRATCH_DIR/work.XXXXXX")
  rc=0
  err=$( cd "$work" && env PATH=/nonexistent "$BASH" "$SELF" "$SCRATCH_DIR/flatten" 2>&1 1>/dev/null ) || rc=$?
  s1=no; case "$err" in *sha256sum*) s1=yes ;; esac
  s2=no; case "$err" in *shasum*) s2=yes ;; esac
  check "NO HASH TOOL child with PATH stripped exits 1 naming both tools" \
    "rc=1 sha256sum=yes shasum=yes" "rc=$rc sha256sum=$s1 shasum=$s2" str
  echo "        stderr: $err"

  # 11. OUTPUT INSIDE INPUT — refuse to write into the tree being hashed.
  fix="$SCRATCH_DIR/inside"
  mkdir -p "$fix/a"
  printf 'one\n' > "$fix/a/one.bin"
  rc=0
  err=$( cd "$fix" && "$BASH" "$SELF" . 2>&1 1>/dev/null ) || rc=$?
  check "OUTPUT INSIDE INPUT cwd inside the hashed tree exits 1" 1 "$rc" eq
  echo "        stderr: $err"

  # The assertion-run total is the anti-vacuity floor: a harness that silently
  # stopped running assertions would otherwise report zero failures and exit 0.
  echo ""
  echo "$assertions assertion(s) run, $failures failed."
  [ "$failures" -eq 0 ]
}

# Assertion helper. Same shape as verify-envelopes.sh: named locals, a case on
# the comparison mode, two-space-indented PASS/FAIL output, and accumulation
# into `failures`. `str` is the one added mode — the exit-code rows capture the
# child's status into a variable and use `eq`, so no `status` mode is needed.
check() {
  local description="$1"
  local expected="$2"
  local actual="$3"
  local comparison="$4"  # "gte" | "eq" | "zero" | "str"

  assertions=$((assertions + 1))

  local pass=false
  case "$comparison" in
    gte)  if [ "$actual" -ge "$expected" ]; then pass=true; fi ;;
    eq)   if [ "$actual" -eq "$expected" ]; then pass=true; fi ;;
    zero) if [ "$actual" -eq 0 ];           then pass=true; fi ;;
    str)  if [ "$actual" = "$expected" ];   then pass=true; fi ;;
  esac

  if $pass; then
    echo "  PASS  $description (got $actual)"
  else
    echo "  FAIL  $description (got $actual, expected $comparison $expected)"
    failures=$((failures + 1))
  fi
}

# Runs the generator as a child from a scratch cwd OUTSIDE the fixture tree.
# Sets st_rc, st_err (stderr only) and st_out (path to the generated file).
_st_run() {
  st_workdir=$(mktemp -d "$SCRATCH_DIR/work.XXXXXX")
  st_rc=0
  st_err=$( cd "$st_workdir" && "$BASH" "$SELF" "$1" 2>&1 1>/dev/null ) || st_rc=$?
  st_out="$st_workdir/$OUTPUT_NAME"
}

main "$@"
