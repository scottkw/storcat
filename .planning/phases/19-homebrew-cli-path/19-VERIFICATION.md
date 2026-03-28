---
phase: 19-homebrew-cli-path
verified: 2026-03-28T00:00:00Z
status: human_needed
score: 3/3 must-haves verified
re_verification: false
human_verification:
  - test: "Install cask and verify storcat on PATH in a new terminal"
    expected: "which storcat returns /opt/homebrew/bin/storcat (or /usr/local/bin/storcat on Intel); storcat version prints the version string; ls -la $(which storcat) shows a symlink to .../StorCat.app/Contents/MacOS/StorCat"
    why_human: "Requires a real Homebrew cask installation on macOS. Cannot simulate cask install programmatically without side effects. PKG-03 is explicitly marked manual-only in VALIDATION.md because it depends on a live brew install --cask storcat."
---

# Phase 19: Homebrew CLI PATH Verification Report

**Phase Goal:** Users who run `brew install --cask storcat` get a `storcat` command immediately available in any new terminal session
**Verified:** 2026-03-28
**Status:** human_needed — all automated checks passed; one manual smoke test required for PKG-03
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | storcat.rb.template contains a binary stanza that symlinks StorCat binary into brew prefix bin | VERIFIED | Line 17: `binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"` — present after `app "StorCat.app"` (line 16) and before `zap` (line 19) |
| 2 | update-tap.sh heredoc includes matching binary stanza so manual runs stay in sync | VERIFIED | Line 97 inside heredoc (lines 80-106): identical stanza after `app "StorCat.app"` (line 96) and before `zap` (line 99) |
| 3 | sed template rendering passes the binary stanza through untouched (no template tokens in it) | VERIFIED | `sed -e 's/{{VERSION}}/2.3.0/g' -e 's/{{SHA256}}/abc123/g' storcat.rb.template \| grep binary` returns the stanza unchanged — binary stanza contains no `{{...}}` tokens |

**Score:** 3/3 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `packaging/homebrew/storcat.rb.template` | Homebrew cask formula with binary stanza for CLI PATH | VERIFIED | File exists, 26 lines, contains exact string `binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"` at line 17 — confirmed by commit 79b45e36 |
| `packaging/homebrew/update-tap.sh` | Manual tap update script with matching binary stanza | VERIFIED | File exists, 161 lines, contains exact string at line 97 inside heredoc — confirmed by commit 5a01beb3; `bash -n` syntax check passes |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `packaging/homebrew/storcat.rb.template` | `StorCat.app/Contents/MacOS/StorCat` | Homebrew binary stanza symlink | WIRED | Pattern `binary.*StorCat.app/Contents/MacOS/StorCat.*target.*storcat` found at line 17 — stanza ordering verified (after `app`, before `zap`) |
| `.github/workflows/distribute.yml` | `packaging/homebrew/storcat.rb.template` | sed substitution of VERSION and SHA256 only | WIRED | Lines 38-41: `sed -e "s/{{VERSION}}/.../" -e "s/{{SHA256}}/.../" packaging/homebrew/storcat.rb.template > tap/Casks/storcat.rb` — only VERSION and SHA256 tokens are substituted; binary stanza has no template tokens and passes through unchanged |

### Data-Flow Trace (Level 4)

Not applicable. This phase modifies static configuration template files (Homebrew cask DSL), not code that renders dynamic data. There are no React components, API routes, or data-fetching patterns to trace.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Binary stanza present in template | `grep -c 'binary.*StorCat.app/Contents/MacOS/StorCat.*target.*storcat' packaging/homebrew/storcat.rb.template` | 1 | PASS |
| Binary stanza present in update-tap.sh | `grep -c 'binary.*StorCat.app/Contents/MacOS/StorCat.*target.*storcat' packaging/homebrew/update-tap.sh` | 1 | PASS |
| sed passthrough preserves binary stanza | `sed -e 's/{{VERSION}}/2.3.0/g' -e 's/{{SHA256}}/abc123/g' template \| grep binary` | stanza unchanged | PASS |
| update-tap.sh bash syntax check | `bash -n packaging/homebrew/update-tap.sh` | exit 0 | PASS |
| Stanza ordering (app then binary then zap) | `grep -n 'app\|binary\|zap' storcat.rb.template` | lines 16, 17, 19 in correct order | PASS |
| Stanza ordering in update-tap.sh heredoc | `grep -n 'app\|binary\|zap' update-tap.sh` | lines 96, 97, 99 in correct order | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|---------|
| PKG-01 | 19-01-PLAN.md | Homebrew cask `binary` stanza puts `storcat` on PATH after `brew install --cask storcat` | SATISFIED | `binary` stanza with `target: "storcat"` present in `storcat.rb.template` (line 17); CI workflow renders template via sed and writes to `tap/Casks/storcat.rb`; stanza is syntactically correct per Homebrew Cask Cookbook |
| PKG-03 | 19-01-PLAN.md | `storcat version` works from any new terminal after Homebrew install (macOS) | NEEDS HUMAN | Static analysis confirms: (1) binary stanza points to correct path `StorCat.app/Contents/MacOS/StorCat`, (2) the unified binary handles `version` subcommand per CLI dispatch in `main.go` (confirmed in RESEARCH.md). End-to-end behavior requires real cask install — cannot verify programmatically |

**Orphaned requirements check:** REQUIREMENTS.md maps PKG-01 and PKG-03 to Phase 19. PKG-02 and PKG-04 are mapped to Phase 20. No orphaned requirements for this phase.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | — | — | No anti-patterns found in either modified file |

### Human Verification Required

#### 1. End-to-end Homebrew cask install smoke test

**Test:**
1. `brew install --cask storcat` (or reinstall if already installed)
2. Open a **new** terminal session
3. `which storcat`
4. `storcat version`
5. `ls -la $(which storcat)`

**Expected:**
- `which storcat` returns `/opt/homebrew/bin/storcat` (Apple Silicon) or `/usr/local/bin/storcat` (Intel)
- `storcat version` prints the version string (e.g., `storcat version 2.x.x`)
- `ls -la $(which storcat)` shows a symlink pointing to `.../StorCat.app/Contents/MacOS/StorCat`

**Why human:** Requires a real Homebrew cask installation on macOS — cannot simulate cask install and symlink creation programmatically without actual side effects. PKG-03 behavioral correctness (binary runs correctly via symlink, Gatekeeper allows execution) depends on the live macOS security model and the signed/notarized app bundle from Phase 17.

### Gaps Summary

No gaps. All three automated must-have truths are verified. Both artifacts are substantive (correct content, correct ordering, correct syntax). Both key links are wired (binary stanza chains from template through CI sed rendering to the cask; update-tap.sh heredoc matches). No anti-patterns found. One item routes to human verification (PKG-03 end-to-end smoke test) because it requires a live cask install.

---

_Verified: 2026-03-28_
_Verifier: Claude (gsd-verifier)_
