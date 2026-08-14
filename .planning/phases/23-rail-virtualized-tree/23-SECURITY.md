---
phase: 23-rail-virtualized-tree
audited: 2026-08-14
asvs_level: 1
block_on: high
verdict: SECURED
threats_total: 26
threats_closed: 26
threats_open: 0
carryover_redecided: 1
follow_ups_open: 2
---

# Phase 23: Rail + Virtualized Tree — Security Audit

**Verdict: SECURED** — 26/26 threats from the six `23-0*-PLAN.md` `<threat_model>` blocks verified against implemented code, plus one carried-over acceptance from Phase 22 re-opened and re-decided. 0 open at or above `block_on: high`.

The auditor independently re-ran `go build ./...`, `go test ./internal/osutil/... ./internal/search/... ./internal/config/... -race`, and `npx tsc --noEmit` at current HEAD, confirming the five code-review fixes (`786b8ddf`, `d5f41f1c`, `23692e32`, `ada22533`) actually hold rather than trusting their disposition notes.

Unlike Phase 22, this phase has genuine security surface: a process-spawning binding, untrusted JSON parsed from disk, and a cache file the app writes and later trusts.

## Highest-severity findings, verified

| Threat | Sev | Disposition | Evidence |
|---|---|---|---|
| **T-23-01** — `RevealInFileManager` spawns a process | **critical** | mitigate | `internal/osutil/reveal.go:131-188` uses `exec.Command(name, args...)` — no shell, no interpreter anywhere in the file. `reveal_test.go` asserts exact argv element count and content for a path carrying space, single and double quotes, `;`, `&&`, backtick, `$()`, pipe and newline, on all three platform builders. |
| **T-23-02** — reveal reachable from any renderer JS | **high** | mitigate (**upgraded** by review finding WR-02, superseding the plan's weaker original text) | `reveal.go:91-108,168-177` — `containsPath` uses `filepath.Rel` on `EvalSymlinks`-resolved absolute paths, **not** `strings.HasPrefix`, so `/catalogs-evil` cannot pass a `/catalogs` test and `..` cannot escape. `TestContainsPath` covers legitimate / sibling-name-prefix / `../`-escape / symlink-escape. `TestRevealInFileManager_Rejects{MissingCatalogDir,PathOutsideCatalogDir}` prove it is wired into the real function and stops **before** `exec.Command`. |
| **T-23-SC** — new npm dependency | **high** | mitigate | `frontend/package-lock.json:695` pins `@tanstack/react-virtual` to `3.14.9` from the official registry tarball, committed alongside the `^3.14.9` range. |
| **T-23-03** — deeply nested JSON → stack exhaustion | **high** | mitigate | `internal/search/flatten.go:14,35` — 512 cap with a `>=` guard (the IN-02 boundary fix confirmed correct). `TestLoadCatalogFlat_DepthCap` exercises it with a depth-600 fixture. |

## Remaining register

All 22 further threats verified closed. Highlights: the superseded-load guard (`AppContext.tsx:99,110` — a slow load for catalog A cannot repaint over a newer selection of B); the counts cache's mutex covering **reads as well as writes** plus temp-file-and-rename atomicity with `os.Remove` on every failure branch and a reset-to-empty on unmarshal failure (`TestCountsCache_ConcurrentPutGet_NoRace`, `TestCountsCache_AtomicWrite_NoLeftoverTempFile`); the `json.Valid` fast path before structural unmarshal; and script-injection closed everywhere by construction — every catalog-derived string (`node.name`, `segment`, `catalog.title`, `catalog.filename`, `rawError`, the filter echo) renders as a plain React text child, with a repo-wide grep for `dangerouslySetInnerHTML|innerHTML|eval(|new Function(` returning zero matches under `frontend/src`.

## Carryover re-decided: T-22-05 (`CatalogModal` `srcDoc` iframe)

Phase 22 accepted this **only** because the component was unreachable, with an explicit instruction to re-open it no later than this phase. It was re-opened and re-verified from scratch, not carried forward:

- `frontend/src/components/CatalogModal.tsx:133` still assigns `srcDoc={htmlContent}` with **no sanitization**, and `htmlContent` comes from a catalog-adjacent `.html` file — attacker-influenceable by this phase's own trust-boundary reasoning.
- **Materially changed since Phase 22:** the `window.electronAPI = wailsAPI` shim (`wailsAPI.ts:252-255`) is now live, so `CatalogModal`'s `getCatalogHtmlPath`/`readHtmlFile` calls route to real Go, not dead stubs. The plumbing beneath the component is proven working.
- Reachability re-checked fresh: `grep -rn "dispatchEvent" frontend/src` yields only `themeChange`. **Zero dispatchers of `openCatalogModal` exist**, so `App.tsx:37`'s listener still cannot fire.
- This phase's own "Open HTML catalog" action deliberately does **not** route through `CatalogModal` — it calls `OpenExternal` → `runtime.BrowserOpenURL`, handing the file to the system browser, outside the app's webview entirely.

**Re-decision: ACCEPT, on freshly verified unreachability — but this must not survive a third carry-forward.** The acceptance is strictly riskier than Phase 22's, because only the missing dispatcher now stands between attacker-influenceable HTML and an unsanitized `srcDoc`. **Phase 26 must resolve it** (sanitize `htmlContent`, or delete `CatalogModal.tsx` if Settings replaces it) as part of that phase's plan — not re-affirm it.

## Open follow-ups (medium, below `block_on: high`)

Recorded rather than silently closed. Both are non-blocking and neither gates this phase.

### FU-23-A — `GetCatalogHtmlPath` / `OpenExternal` lack the WR-02 containment treatment

`app.go:224-238` and `app.go:241-243` are Wails-exposed and callable from **any** renderer JS, but have no containment check, no symlink resolution and no regular-file gate — unlike `RevealInFileManager` after WR-02. T-23-09 is correctly CLOSED as declared, because the panel provably never passes a free-form path (`DetailsPanel.tsx:121` always passes `catalog.path`). But "reachable from any renderer JS, not just this call site" is *precisely* the reasoning WR-02 used to escalate and fix the identical gap in the sibling binding, and that treatment was not extended here.

Worst case is bounded — a browser tab opening an attacker-chosen `.html`, or a file-manager window on a directory — hence medium, not high. **Owner: Phase 26**, alongside the T-22-05 resolution, since both touch the same HTML-opening surface. The fix is mechanical: thread `catalogDir` through both bindings and reuse the existing `containsPath` helper, exactly as WR-02 did.

### FU-23-B — unregistered platform-branch instantiation

`DetailsPanel.tsx:93,102-115` reads `Environment().platform` to choose the reveal button's label — the same pattern Phase 22 accepted as T-22-08 for the macOS toolbar inset, but this instantiation has no `T-23-xx` ID of its own. **No security impact**: the value is string-compared and feeds only button label text, never a style, path or command. Registered here for completeness so a future audit does not rediscover it as unknown.

## Unregistered flags

None. No `## Threat Flags` section appears in any of the six `23-0*-SUMMARY.md` files; independent review found no attack surface beyond the register above and the two follow-ups.
