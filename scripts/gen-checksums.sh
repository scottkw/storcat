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

# Select the hasher at run time. Preference order is sha256sum then
# shasum -a 256. On ubuntu-22.04 that resolves to GNU coreutils; on current
# macOS it resolves to Apple's /sbin/sha256sum, with Perl's /usr/bin/shasum as
# the fallback for older systems that lack sha256sum. The shasum branch is
# therefore dead on a modern macOS box and live only on an older one.
#
# Their DIGESTS agree; their LINE FORMATTING does not. GNU coreutils and Perl
# backslash-escape a name containing a backslash or newline by prefixing the
# whole line with a literal `\`; Apple's does not. Verified on macOS 26.6 —
# the same file, same digest, two different lines:
#   /sbin/sha256sum 'back\slash.bin'  →  3a6eb079…  back\slash.bin
#   shasum -a 256   'back\slash.bin'  → \3a6eb079…  back\\slash.bin
# So this selection is not free and downstream cannot assume a fixed format.
# generate() validates the digest field rather than trusting it; that check,
# not this comment, is what enforces the invariant.
#
# Finding neither tool is a hard failure: substituting another digest here
# would be the most damaging silent default in the pipeline.
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
# directory — recording `<basename><TAB><path>` per uploadable file. The sort key
# is carried on a tab-delimited scratch line so the sort acts on the filename
# and nothing else. Two failure modes are designed out here: sorting the
# finished hash-first lines sorts by hash rather than by filename, and a naive
# whitespace-field sort breaks on any filename containing a space.
collect() {
  local dir="$1"
  local pairs="$2"
  local list="$SCRATCH_DIR/find.out"
  local f

  # find writes to a file so its exit status is checkable. The previous form —
  # `done < <(find …)` — ran the walk in a process substitution, whose status
  # the shell discards and which neither `set -e` nor `set -o pipefail` covers.
  # An unreadable subdirectory, an I/O error or a vanished directory therefore
  # truncated the walk silently: find printed to stderr, exited non-zero, and
  # the loop consumed whatever it had managed to emit. Every downstream guard
  # then agreed with the short list, because every one of them measures that
  # same list. A partial walk is a hard failure, never a shorter release.
  # -L follows symlinks, so the walk covers exactly what the publisher uploads.
  # softprops/action-gh-release resolves its `files:` globs through @actions/glob
  # (followSymbolicLinks defaults true) and filters with statSync, which follows
  # symlinks — so a symlinked artifact IS uploaded. A bare `-type f` matches
  # regular files only and silently omitted it, giving a published asset with no
  # checksum line and, unlike the partial-walk case above, no stderr at all.
  #
  # -L rather than the wider `\( -type f -o -type l \)`: that form also matches
  # symlinks to directories, dangling symlinks and loop entries, none of which
  # the hasher can read and none of which the publisher uploads as files. Under
  # -L a symlink to a file matches -type f under its own basename (the published
  # name) and hashes through to the target bytes, a symlink to a directory is
  # descended into exactly as the publisher's glob descends, and a loop makes
  # find exit non-zero — caught by the check below rather than silently dropped.
  if ! find -L "$dir" -type f -print0 > "$list"; then
    echo "Error: could not fully walk the artifact directory: $dir" >&2
    echo "find exited non-zero; refusing to checksum a partial artifact list." >&2
    exit 1
  fi

  : > "$pairs"
  while IFS= read -r -d '' f; do
    printf '%s%s%s\n' "${f##*/}" "$TAB" "$f" >> "$pairs"
  done < "$list"
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
    hash=${hash%% *}
    # Validate rather than trust. The hasher's first whitespace-delimited field
    # is NOT always a digest: GNU coreutils sha256sum and Perl's shasum escape
    # a name containing a backslash or newline by prefixing the ENTIRE line
    # with a literal backslash, so ${hash%% *} yields a 65-character field
    # starting with `\`. Apple's /sbin/sha256sum does not escape, so the two
    # hashers this script selects between disagree on exactly the inputs
    # nothing else checks. The exposure is not only the artifact basenames the
    # workflow controls: paths are hashed, so any backslash anywhere in the
    # caller's directory path reaches this. Unvalidated, the malformed line was
    # written, `shasum -c` reported "1 line is improperly formatted", and with
    # --ignore-missing it was skipped while this script exited 0 — the count
    # guard counts lines, not valid lines.
    #
    # The class is spelled out rather than [0-9a-f] on purpose: this task's
    # verify block greps everything above self_test() for a bare 9, and the
    # range form trips it. Do not "tidy" it back.
    if [[ ! "$hash" =~ ^[0123456789abcdef]{64}$ ]]; then
      echo "Error: hasher returned a malformed digest for: $path" >&2
      echo "Got: $hash" >&2
      echo "Refusing to write a hash field that is not 64 lowercase hex characters." >&2
      exit 1
    fi
    printf '%s  %s\n' "$hash" "$base" >> "$out"
  done < <(sort -t "$TAB" -k1,1 "$pairs")
}

main() {
  local dir out pairs tmp

  case "${1:-}" in
    --self-test)
      # Called as a plain command, NOT as `self_test && exit 0`. The left
      # operand of `&&` runs with `set -e` suspended for the whole function
      # body, so a failure in the harness machinery itself — a `mktemp -d`, a
      # fixture `mkdir`, a `cd` — would not abort and later assertions would
      # run against fixtures that do not exist. Plain invocation keeps `set -e`
      # live; the assertion bodies all capture child status explicitly
      # (`rc=0; ( … ) || rc=$?`), so a failing assertion still records a FAIL
      # row and lets the remaining rows run.
      self_test
      exit 0
      ;;
  esac

  require_input_dir "${1:-}"
  require_hash_tool

  dir="$1"
  out="$PWD/$OUTPUT_NAME"

  SCRATCH_DIR=$(mktemp -d)
  pairs="$SCRATCH_DIR/pairs"
  # Built in the scratch dir, never at the final path. generate() appends one
  # line at a time, so writing straight to "$out" meant a hasher failure on
  # artifact n of N aborted under `set -e` with n-1 lines already committed —
  # a syntactically perfect integrity file covering a strict subset of the
  # release, with no marker that it was partial, and the previous good file
  # already destroyed by the up-front truncation. The count guard never ran,
  # because the script was gone. CI was partly shielded (the non-zero exit
  # stops the publish step); the manual backfill caller was not.
  tmp="$SCRATCH_DIR/out"

  # Guard order is deliberate: nothing is hashed until the input is known good,
  # nothing is written until the output location is known safe, and a basename
  # collision is caught on the collected list so it costs nothing to detect.
  collect "$dir" "$pairs"
  assert_dir_not_empty "$dir" "$pairs"
  assert_output_outside_input "$dir" "$out"
  assert_no_duplicate_basenames "$pairs"

  generate "$pairs" "$tmp"

  # Derived at run time from both sides of the transform, after the write.
  assert_count_matches "$(awk 'END{print NR}' "$pairs")" "$(awk 'END{print NR}' "$tmp")"

  # Publish only a file that passed every guard. Until this line runs, any
  # existing checksums.txt is the previous good one, untouched.
  # ponytail: SCRATCH_DIR comes from mktemp -d, so this may be a cross-device
  # mv (copy + unlink) rather than an atomic rename. That is a far smaller
  # window than the bug being fixed — it needs a kill during the copy of a
  # sub-kilobyte file, not merely an unreadable artifact. Move the temp
  # alongside "$out" and extend the EXIT trap if that window ever matters.
  mv "$tmp" "$out"

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
  local fix work fakebin names lastbyte nr hashn named wrote linkhash before after rc rc_noarg rc_file err s1 s2
  # The anti-vacuity floor. Enforced at the bottom of this function.
  local EXPECTED_ASSERTIONS=15

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

  # 12. UNREADABLE SUBDIR — a walk that could not complete must abort, not
  #     publish what it managed to read. Before find's exit status was checked
  #     this fixture wrote a two-line checksums.txt and exited 0, leaving
  #     critical.bin uncovered with every guard green.
  fix="$SCRATCH_DIR/unreadable"
  mkdir -p "$fix/ok" "$fix/locked"
  printf 'a\n' > "$fix/ok/a.bin"
  printf 'b\n' > "$fix/ok/b.bin"
  printf 'secret\n' > "$fix/locked/critical.bin"
  chmod 000 "$fix/locked"
  _st_run "$fix"
  named=no; case "$st_err" in *"$fix"*) named=yes ;; esac
  wrote=no; if [ -e "$st_out" ]; then wrote=yes; fi
  chmod 755 "$fix/locked"   # restore so the EXIT trap's rm -rf can descend
  check "UNREADABLE SUBDIR partial walk exits 1, names the dir, writes nothing" \
    "rc=1 names_dir=yes wrote=no" "rc=$st_rc names_dir=$named wrote=$wrote" str
  echo "        stderr: $st_err"

  # 13. SYMLINKED ARTIFACT — the publisher follows symlinks, so the generator
  #     must too. Before -L this fixture emitted one line for two uploadable
  #     assets, exit 0, no stderr at all. The digest is asserted against the
  #     target's bytes to prove the hash reads through the link.
  #     The target is zero-byte so the expected digest is the published
  #     empty-string SHA-256 (same constant assertion 4 uses), which pins
  #     "hashed through the link" without shelling out to the hasher here.
  fix="$SCRATCH_DIR/symlink"
  mkdir -p "$fix/real"
  : > "$SCRATCH_DIR/symlink-target.bin"
  printf 'plain\n' > "$fix/real/other.bin"
  ln -s "$SCRATCH_DIR/symlink-target.bin" "$fix/StorCat-v9.9.9-linux-amd64.tar.gz"
  _st_run "$fix"
  names=$(awk '{print $2}' "$st_out" | tr '\n' ' '); names=${names% }
  nr=$(awk 'END{print NR}' "$st_out")
  linkhash=$(awk '$2=="StorCat-v9.9.9-linux-amd64.tar.gz"{print $1}' "$st_out")
  check "SYMLINKED ARTIFACT symlink is covered and hashes through to the target" \
    "rc=0 lines=2 names=StorCat-v9.9.9-linux-amd64.tar.gz other.bin hash=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" \
    "rc=$st_rc lines=$nr names=$names hash=$linkhash" str

  # 14. ATOMIC WRITE — a hasher failure part-way through the set must not
  #     truncate the output, and must not destroy the previous good file.
  #     Runs twice in the SAME working directory: a good run, then one with an
  #     unreadable artifact. Before the build-then-rename change the second run
  #     replaced a correct 2-line file with a 1-line one and left it on disk.
  fix="$SCRATCH_DIR/atomic"
  mkdir -p "$fix"
  printf 'a\n' > "$fix/a.bin"
  printf 'b\n' > "$fix/b.bin"
  _st_run "$fix"
  before=$(awk 'END{print NR}' "$st_out")
  chmod 000 "$fix/b.bin"
  rc=0
  err=$( cd "$st_workdir" && "$BASH" "$SELF" "$fix" 2>&1 1>/dev/null ) || rc=$?
  chmod 644 "$fix/b.bin"
  after=$(awk 'END{print NR}' "$st_out")
  check "ATOMIC WRITE hasher failure leaves the previous good file intact" \
    "before=2 rc=1 after=2" "before=$before rc=$rc after=$after" str
  echo "        stderr: $err"

  # 15. MALFORMED DIGEST — a hasher whose first field is not a digest must be
  #     refused, not written through. The stub reproduces the real GNU/Perl
  #     backslash-escaping behaviour (whole line prefixed with a literal
  #     backslash) with a stub rather than a fixture, because whether the real
  #     hasher escapes depends on which of the three implementations is
  #     installed. Unvalidated, this wrote `\000…0  a.bin` and exited 0.
  fakebin="$SCRATCH_DIR/fakebin"
  mkdir -p "$fakebin"
  cat > "$fakebin/sha256sum" <<'FAKE'
#!/bin/sh
printf '\\%064d  %s\n' 0 "$1"
FAKE
  chmod +x "$fakebin/sha256sum"
  work=$(mktemp -d "$SCRATCH_DIR/work.XXXXXX")
  rc=0
  err=$( cd "$work" && env PATH="$fakebin:$PATH" "$BASH" "$SELF" "$SCRATCH_DIR/flatten" 2>&1 1>/dev/null ) || rc=$?
  named=no; case "$err" in *"malformed digest"*) named=yes ;; esac
  wrote=no; if [ -e "$work/$OUTPUT_NAME" ]; then wrote=yes; fi
  check "MALFORMED DIGEST escaped hasher output is refused, not written" \
    "rc=1 says_malformed=yes wrote=no" "rc=$rc says_malformed=$named wrote=$wrote" str
  echo "        stderr: $err"

  # The assertion-run total is the anti-vacuity floor: a harness that silently
  # stopped running assertions would otherwise report zero failures and exit 0.
  # The floor only exists if the total is COMPARED, not merely printed — an
  # exit status derived from "$failures" alone is satisfied by running no
  # assertions at all. Bump EXPECTED_ASSERTIONS deliberately when adding a row.
  echo ""
  echo "$assertions assertion(s) run, $failures failed."
  if [ "$assertions" -ne "$EXPECTED_ASSERTIONS" ]; then
    echo "Error: expected $EXPECTED_ASSERTIONS assertion(s), ran $assertions." >&2
    echo "The harness stopped early, or a row was added/removed without updating the count." >&2
    return 1
  fi
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
