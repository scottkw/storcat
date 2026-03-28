---
phase: 19-homebrew-cli-path
plan: "01"
subsystem: packaging/homebrew
tags: [homebrew, cli, path, cask, packaging]
dependency_graph:
  requires: [phase-17-macos-signing-notarization]
  provides: [homebrew-cli-path]
  affects: [packaging/homebrew/storcat.rb.template, packaging/homebrew/update-tap.sh]
tech_stack:
  added: []
  patterns: [homebrew-binary-stanza]
key_files:
  created: []
  modified:
    - packaging/homebrew/storcat.rb.template
    - packaging/homebrew/update-tap.sh
decisions:
  - "Binary stanza points to StorCat.app/Contents/MacOS/StorCat (the unified CLI+GUI binary)"
  - "No template tokens in binary stanza — sed substitution in CI passes it through untouched"
  - "Unquoted EOF heredoc in update-tap.sh correctly passes Ruby #{appdir} literals while expanding bash variables"
metrics:
  duration: 28s
  completed: "2026-03-28"
  tasks_completed: 2
  files_modified: 2
requirements:
  - PKG-01
  - PKG-03
---

# Phase 19 Plan 01: Homebrew CLI PATH Summary

## One-liner

Added Homebrew `binary` stanza to cask template and update-tap.sh so `brew install --cask storcat` creates `storcat` symlink in brew prefix bin.

## What Was Done

Added the `binary` stanza to two files so that Homebrew cask installation creates a `storcat` CLI symlink in `$(brew --prefix)/bin`:

1. **`packaging/homebrew/storcat.rb.template`** — Added `binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"` after the `app "StorCat.app"` line. This is the CI-rendered template used by `distribute.yml`.

2. **`packaging/homebrew/update-tap.sh`** — Added the matching `binary` stanza inside the heredoc block. This keeps the manual fallback tap update script in sync with the CI template.

## Tasks

### Task 1: Add binary stanza to storcat.rb.template
- **Status:** Complete
- **Commit:** 79b45e36
- **Files:** `packaging/homebrew/storcat.rb.template`
- **Verification:** `grep` returns 1 match; `sed` substitution preserves stanza unchanged

### Task 2: Add binary stanza to update-tap.sh heredoc
- **Status:** Complete
- **Commit:** 5a01beb3
- **Files:** `packaging/homebrew/update-tap.sh`
- **Verification:** `grep` returns 1 match; `bash -n` syntax check passes

## Verification Results

1. Template binary stanza: `grep 'binary.*StorCat.app.*target.*storcat' packaging/homebrew/storcat.rb.template` — 1 match
2. Script binary stanza: `grep 'binary.*StorCat.app.*target.*storcat' packaging/homebrew/update-tap.sh` — 1 match
3. Template sed passthrough: binary stanza preserved unchanged after `{{VERSION}}`/`{{SHA256}}` substitution
4. Script syntax: `bash -n packaging/homebrew/update-tap.sh` exits 0
5. Only 2 files modified (template-only phase, no CI workflow changes needed)

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

- `packaging/homebrew/storcat.rb.template` exists and contains binary stanza
- `packaging/homebrew/update-tap.sh` exists and contains binary stanza
- Commits 79b45e36 and 5a01beb3 present in git log
