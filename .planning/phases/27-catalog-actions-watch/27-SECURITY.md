---
phase: 27
slug: catalog-actions-watch
status: secured
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: 2026-08-16
register_authored_at_plan_time: true
---

# Phase 27 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

Register authored at PLAN time across all seven `27-0N-PLAN.md` `<threat_model>` blocks, then verified
against the implementation by `gsd-security-auditor` (ASVS L1 configured). The auditor re-ran
`go build ./...`, `go vet ./...`, `go test ./... -race -count=1` and `npx tsc --noEmit` live during the
audit rather than accepting the carried green-suite claim.

This phase introduced **two** new Go dependencies — `github.com/Bios-Marcel/wastebasket/v2` v2.0.3 (the
ROADMAP's anticipated one) and `github.com/fsnotify/fsnotify` v1.10.1 (a second, accepted as an explicit
user decision during discuss). Both are exact-pinned in `go.mod`.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Renderer JS → Wails binding | Any renderer JS can call any bound `App` method; the renderer is not a trusted caller | Catalog paths, `catalogDir`, new titles |
| App → third-party Trash library | `wastebasket/v2`'s macOS backend hands a path to `osascript` as an interpolated AppleScript string | Containment-validated absolute file paths only |
| Filesystem → watcher → renderer | `fsnotify` events on the catalog directory become a `catalogs:changed` signal | A bare signal — no payload, no path |
| App → disk (catalog writes) | rename / duplicate / delete each mutate user data | Catalog `.json` and `.html` pairs |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-27-01 | Tampering | `wastebasket` macOS AppleScript interpolation | high | mitigate | `trash.go:53-98` — every path `Lstat` → `Abs` → `EvalSymlinks` → regular-file → extension-allowlist → `ContainsPath`, all before `trashSeam` is reached (`:104`). No branch skips a step. Upstream residual recorded, not claimed closed | closed |
| T-27-02 | Tampering / EoP | renderer-facing catalog bindings | high | mitigate | `app.go` — `RenameCatalog` (903-931), `DuplicateCatalog` (941-970), `DeleteCatalog` (993-1026) each fail closed on empty `catalogDir`, then `Abs` → `EvalSymlinks` → `ContainsPath` → extension check before entering `catalog`/`osutil` | closed |
| T-27-03 | Tampering | title escaping across read/write | medium | mitigate | `rename.go:239-252` escapes at both `<title>` and `<h1>`; `service.go:263` unescapes on the HTML fallback read ONLY (the JSON title is never unescaped); titles render as JSX text children — no `dangerouslySetInnerHTML` anywhere | closed |
| T-27-04 | Denial of Service | watcher event/error starvation | medium | mitigate | `watcher.go:140-180` — `Events` and `Errors` drained in one `select`; any error including `ErrEventOverflow` fires the callback, never dropped | closed |
| T-27-05 | Denial of Service | watcher release (WATCH-03) | medium | mitigate | `watcher.go:185-192` `sync.Once`-guarded `Close`; `app.go:637-659` `applyWatchState` closes-first; `app.go:823-830` `shutdown`; `main.go:76` `OnShutdown` registered | closed |
| T-27-06 | Tampering | crash mid-write corrupting a catalog (ACT-09) | high | mitigate | `atomicwrite.go` — `tmp.Sync()` before close, best-effort `syncDir` after `Rename`, both grep/read-verified in production code. Proof: 42 real `SIGKILL` iterations, SHA-256-verified survival. **Coupling caveat recorded below and as ledger entry 14** | closed |
| T-27-07 | Information Disclosure | watcher observing directory contents | low | accept | The watcher performs no read/write itself; it only triggers an idempotent `BrowseCatalogs` re-list | closed |
| T-27-09 | Repudiation | failed Trash reporting (ACT-05) | high | mitigate | `trash.go:104-106` wraps the seam error with `%w`; `DeleteConfirmDialog.tsx:62,128` renders `result.error` verbatim with no substitution; `onDeleted()` fires only on `result.success` | closed |
| T-27-10 | Tampering | JSON envelope preservation on rename | medium | mitigate | `rename.go:130-237` token-walks with `json.RawMessage` — no map/struct round-trip. `TestRenameCatalog_PreservesArrayEnvelope` / `_PreservesNestedContentsBytes` pass | closed |
| T-27-11 | Denial of Service | directory-fsync failure on Windows | medium | mitigate | `atomicwrite.go` — error captured and logged through a `sync.Once`-bounded `log.Printf`, never propagated, destination never removed on that path | closed |
| T-27-12 | Tampering | test-only code reaching production | medium | mitigate | `killtarget/main.go` imports nothing from this project; `grep -c time.Sleep internal/catalog/atomicwrite.go` = 0 | closed |
| T-27-13 | Denial of Service | temp residue after a kill | low | accept | The SIGKILL test reports temp-path residue via `t.Logf` and never asserts it zero; destination-path residue IS asserted zero | closed |
| T-27-14 | Tampering | duplicate colliding with an orphan `.html` | medium | mitigate | `duplicate.go:108-120` — `isCandidateRootFree` checks both `.json` and `.html`; `TestDuplicateCatalog_SkipsRootWithOrphanHTML` passes | closed |
| T-27-15 | Spoofing | menu ARIA advertised before the menu exists | low | mitigate | `DetailsPanel.tsx:113-115` — `aria-haspopup`/`aria-expanded`/`aria-controls` emitted only while the menu is mounted | closed |
| T-27-16 | Denial of Service | duplicate key handlers / listener leak | medium | mitigate | `Menu.tsx` has no Escape handler (Escape is owned solely by `useModalBehavior`); its one `pointerdown` listener is added and removed in the same effect (70-101) | closed |
| T-27-17 | Elevation of Privilege | dialogs acting with a null `catalogDir` | medium | mitigate | `RenameDialog.tsx:52-59` and `DeleteConfirmDialog.tsx:51-54` fail closed; the `app.go` bindings independently fail closed too | closed |
| T-27-18 | Spoofing | renderer supplying the `.html` companion path | medium | mitigate | `DeleteConfirmDialog.tsx:41` computes `htmlPath` for DISPLAY only; `app.go DeleteCatalog:1017-1020` derives the trashed `.html` independently in Go | closed |
| T-27-19 | Tampering | permanent-delete fallback (ACT-05) | high | mitigate | No permanence vocabulary in `DeleteConfirmDialog.tsx` (grep exits 1); no `os.Remove`/`os.RemoveAll` reachable in `trash.go` or `DeleteCatalog` | closed |
| T-27-20 | Denial of Service | double-submit on the delete dialog | low | mitigate | `busy` guard returns early; both footer buttons carry `disabled={busy}` | closed |
| T-27-21 | Denial of Service | callback invoked under the coalescer mutex | medium | mitigate | `watcher.go:63-76` — `fireNow()` releases the mutex before invoking `fn()` | closed |
| T-27-22 | Tampering | Wails runtime leaking into `internal/*` | medium | mitigate | `internal/watch` imports nothing from `wailsapp/wails`; `runtime.EventsEmit` appears only in `app.go` (2 call sites) | closed |
| T-27-23 | Spoofing | watching indicator implying a confirmed watcher | low | accept | `StatusBar.tsx:58` derives the segment from the setting + `catalogDir`, not from a confirmed-running watcher — documented as deliberate in 27-06/27-07 | closed |
| T-27-24 | Denial of Service | event-subscription leak in the rail | low | mitigate | `CatalogRail.tsx:64-70` — the effect returns `EventsOn`'s own unsubscribe function | closed |
| T-27-25 | Information Disclosure | catalog directory path in the status bar | low | accept | Rendered as a JSX text child, never as HTML; already displayed elsewhere in the app | closed |
| T-27-26 | Repudiation | ledger claiming unexercised coverage | medium | mitigate | All six Phase-27 `WINDOWS.md` entries (8–13) describe what was NOT exercised; entry #6 disposed `fixed` citing 27-02's real evidence rather than asserting it | closed |
| T-27-SC | Tampering | supply chain (two new Go dependencies) | high | mitigate | `go.mod` exact-pins `wastebasket/v2 v2.0.3` and `fsnotify v1.10.1`; package-legitimacy audits recorded in the 27-03 and 27-06 threat registers | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above `workflow.security_block_on` (`high`) count toward `threats_open`*

### Unregistered attack surface sweep

None. Only `27-07-SUMMARY.md` carries an explicit `## Threat Flags` section, stating "None — no new network
endpoints, auth paths, or trust-boundary schema changes." The auditor found no new attack surface across the
seven implementation diffs that is not already mapped to a threat ID above.

---

## Auditor notes on the two threats flagged for special scrutiny

**T-27-01 — the containment gate cannot be bypassed.** `TrashPaths` was traced end to end: `Lstat` → `Abs` →
`EvalSymlinks` → regular-file → extension-allowlist → `ContainsPath` all run before `trashSeam` is reached,
with no branch that skips a step. The upstream AppleScript-interpolation weakness itself is correctly NOT
claimed closed — it is an accepted, upstream-owned residual recorded both in `trash.go`'s doc comment and in
`WINDOWS.md` #12.

**T-27-06 — adequate evidence, with a coupling gap recorded.** The two required code-level properties
(`Sync()` before close, `syncDir` after rename) are directly verified in production code, independent of any
test. The SIGKILL proof rests on `killtarget/main.go`, which deliberately imports nothing from this project
and hand-reproduces the same create-temp → write → sync → close → chmod → rename sequence (necessary to
inject a mid-write delay without putting test-only sleep code into production). The auditor compared the two
sequences line by line and confirmed they match in order and syscalls. The residual: **if a future change
alters `WriteFileAtomic`'s internal ordering without a matching `killtarget` update, the SIGKILL test would
keep passing while no longer proving anything about the real function.** Recorded as `WINDOWS.md` entry 14.
This is a test-maintenance coupling gap, not a missing mitigation.

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-27-01 | T-27-07 | The watcher performs no filesystem read or write of its own; it only triggers the same idempotent `BrowseCatalogs` re-list the rail already performs | plan-time disposition | 2026-08-16 |
| AR-27-02 | T-27-13 | Temp-file residue left by a killed process is expected — the killed process never reaches its own cleanup. Destination-path residue is what matters and IS asserted zero | plan-time disposition | 2026-08-16 |
| AR-27-03 | T-27-23 | The watching indicator reflects the user's setting, not a confirmed-running watcher. Showing a "confirmed" state would require a round-trip the indicator does not warrant | plan-time disposition | 2026-08-16 |
| AR-27-04 | T-27-25 | The catalog directory path is already displayed elsewhere in the app and renders as a JSX text child, never as HTML | plan-time disposition | 2026-08-16 |
| AR-27-05 | T-27-01 (residual) | `wastebasket`'s macOS AppleScript interpolation is third-party code this project does not own. The `ContainsPath` gate bounds the worst case but does not close the interpolation. A future feature letting a user type an arbitrary filename that reaches the trash binding would need this revisited | recorded, `WINDOWS.md` #12 | 2026-08-16 |

*Accepted risks do not resurface in future audit runs.*

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-08-16 | 26 | 26 | 0 | gsd-security-auditor (ASVS L1, block_on: high) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
