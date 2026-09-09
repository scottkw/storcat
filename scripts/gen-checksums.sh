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

# Walk the input directory recursively — actions/download-artifact with
# `path: artifacts/` yields artifacts/<job-artifact-name>/<file>, never a flat
# directory — and emit one `<64 lowercase hex><two spaces><basename>` line per
# regular file, sorted by the BASENAME field alone.
#
# The sort key is carried on a tab-delimited scratch line so the sort acts on
# the filename and nothing else. Two failure modes are designed out here:
# sorting the finished hash-first lines sorts by hash rather than by filename,
# and a naive whitespace-field sort breaks on any filename containing a space.
generate() {
  local dir="$1"
  local out="$2"
  local pairs="$3"
  local f base path hash

  : > "$pairs"
  while IFS= read -r -d '' f; do
    printf '%s%s%s\n' "${f##*/}" "$TAB" "$f" >> "$pairs"
  done < <(find "$dir" -type f -print0)

  : > "$out"
  while IFS="$TAB" read -r base path; do
    hash=$(hash_cmd "$path")
    printf '%s  %s\n' "${hash%% *}" "$base" >> "$out"
  done < <(sort -t "$TAB" -k1,1 "$pairs")
}

main() {
  local dir out pairs

  require_input_dir "${1:-}"
  require_hash_tool

  dir="$1"
  out="$PWD/$OUTPUT_NAME"

  SCRATCH_DIR=$(mktemp -d)
  pairs="$SCRATCH_DIR/pairs"

  generate "$dir" "$out" "$pairs"

  # Echo the result into the run log — the cheapest CHECK-01 evidence available
  # in a CI run nobody is watching. distribute.yml establishes this habit.
  echo "Wrote $out"
  cat "$out"
}

main "$@"
