---
phase: 7
slug: verification-merge
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package (built-in) |
| **Config file** | none — standard `go test` |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./... -v` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./... && git ls-files | grep -E "^(node_modules|build/bin|dist/)" | wc -l`
- **Before `/gsd:verify-work`:** Full suite must be green + `wails build` clean output
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | REL-01 | smoke | `git ls-files \| grep -E "^(node_modules\|build/bin\|dist/)" \| wc -l` returns 0 | ✅ | ⬜ pending |
| 07-01-02 | 01 | 1 | REL-02 | smoke | `grep -c "node_modules\|build/bin\|\.DS_Store\|dist/" .gitignore` returns 4+ | ✅ | ⬜ pending |
| 07-01-03 | 01 | 1 | Phase gate | unit | `go test ./...` | ✅ | ⬜ pending |
| 07-01-04 | 01 | 1 | Phase gate | integration | `wails build -clean -platform darwin/universal` | ✅ | ⬜ pending |
| 07-01-05 | 01 | 1 | Phase gate | manual/UAT | Launch app, exercise Create/Search/Browse | human | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Three tabs (Create, Search, Browse) complete primary operations | Phase gate | Requires live Wails app on macOS with GUI interaction | 1. Launch app via `wails dev` 2. Create a catalog from a test directory 3. Search for files in created catalog 4. Browse catalog structure in Browse tab |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
