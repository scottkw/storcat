---
phase: 19
slug: homebrew-cli-path
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-28
---

# Phase 19 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + Homebrew `brew audit` (local) |
| **Config file** | none — no test config file for cask validation |
| **Quick run command** | `grep "binary" packaging/homebrew/storcat.rb.template` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `grep "binary" packaging/homebrew/storcat.rb.template`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 19-01-01 | 01 | 1 | PKG-01 | content inspection | `grep 'binary.*appdir.*StorCat.*target.*storcat' packaging/homebrew/storcat.rb.template` | N/A (template edit) | ✅ green |
| 19-01-02 | 01 | 1 | PKG-01 | content inspection | `grep 'binary.*appdir.*StorCat.*target.*storcat' packaging/homebrew/update-tap.sh` | N/A (script edit) | ✅ green |
| 19-01-03 | 01 | 1 | PKG-03 | sed passthrough | `sed -e 's/{{VERSION}}/2.3.0/g' -e 's/{{SHA256}}/abc123/g' packaging/homebrew/storcat.rb.template \| grep binary` | N/A | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. No new test files needed.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `storcat version` works via symlink after `brew install --cask storcat` | PKG-03 | Requires real Homebrew cask install on macOS | 1. `brew install --cask storcat` 2. Open new terminal 3. `which storcat` 4. `storcat version` 5. Verify symlink: `ls -la $(which storcat)` |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved

## Validation Audit 2026-03-28

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
