---
phase: 26
slug: settings
status: secured
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: 2026-08-15
register_authored_at_plan_time: true
---

# Phase 26 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

Register authored at PLAN time across all five `26-0N-PLAN.md` `<threat_model>` blocks, then verified
against the implementation by `gsd-security-auditor` (ASVS L1 configured; several entries verified at
L2/L3 depth — full data-flow tracing — because they carry milestone-level obligations). The auditor
independently re-ran `go build ./...`, `go test ./... -race -count=1`, and `npx tsc --noEmit` rather
than accepting the carried green-suite claim.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Renderer JS → Wails binding | Any renderer JS can call any bound `App` method; the renderer is not a trusted caller | Filesystem paths (`catalogPath`, `catalogDir`, `rawURL`), settings scalars |
| `localStorage` → Go config | Boot cache is user/attacker-writable via devtools; migration reads it once | Theme id, density, railSide, catalog directory, secondary directory |
| Go config → disk | `storcatConfigDir()/config.json` is the single persisted settings surface | All settings fields |
| App → OS shell | `BrowserOpenURL` / file-manager reveal hand a path to the host OS | Resolved, containment-checked `file://` URL only |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-26-01 | Tampering | `themeTokens.ts` density read | low | mitigate | Exact-string allowlist; invalid → `Comfortable` (`themeTokens.ts:239-244`) | closed |
| T-26-02 | Tampering | `config.Manager` concurrent setters | medium | mitigate | `sync.RWMutex` + copy-returning `Get()`; `-race` test green (`config.go:79,166-194,246-323`) | closed |
| T-26-03 | Denial of Service | ⌘, during foreground scan | medium | mitigate | `openSettings()` early-returns on `counting`/`scanning` (`WorkspaceShell.tsx:122-130`) | closed |
| T-26-04 | Information Disclosure | Settings persistence | low | mitigate | Writes confined to `storcatConfigDir()/config.json`; no `console.*`, `fetch`, or `XMLHttpRequest` in new settings files | closed |
| T-26-05 | Elevation of Privilege | Dialog z-order / scrim | low | accept | `--z-dialog:300` > `--z-overlay:200`; scrim `position:absolute; inset:0` (`workspace.css:30-31,1375`) | closed |
| T-26-06 | Tampering | theme / railSide bindings | low | mitigate | `getThemeById` fallback + exact `'Left'`/`'Right'` allowlist | closed |
| T-26-07 | Tampering | Theme card swatches | low | mitigate | Swatches read only from the compile-time `themes` array, `aria-hidden` (`ThemeGrid.tsx:26-36`) | closed |
| T-26-08 | Spoofing | Theme apply path | medium | mitigate | Exactly 3 `applyTokens(` call sites; `setThemeSetting` dispatches the existing `themeChange` event — no 4th path | closed |
| T-26-09 | Tampering | localStorage→config migration | medium | mitigate | All 5 migrated values pass `readPersistedPrefs()`'s allowlist before any binding call (`settingsStore.ts:151-176`) | closed |
| T-26-10 | Tampering | Migration write-back | medium | mitigate | Identical allowlist applied before `safeSetItem` (`settingsStore.ts:191-195`) | closed |
| T-26-11 | Elevation of Privilege | Catalog-directory boundary | high | mitigate | Both file bindings independently call `osutil.ContainsPath`/`ResolveContainedFileURL` (`app.go:759-806`) | closed |
| T-26-12 | Denial of Service | Repeated `hydrateSettings()` | low | mitigate | Module-level in-flight promise dedup (`settingsStore.ts:133-138`) | closed |
| T-26-13 | Repudiation | localStorage key deletion | medium | mitigate | No `removeItem` anywhere in `settingsStore.ts`; each storage-key literal single-homed | closed |
| **T-26-14** | Elevation of Privilege | `App.OpenExternal` (**FU-23-A**) | **high** | mitigate | `ResolveContainedFileURL`; canonical resolved URL passed to `BrowserOpenURL`, never `rawURL`. 13-subtest table incl. sibling-prefix, symlink escape, non-`file` schemes | closed |
| **T-26-15** | Elevation of Privilege | `App.GetCatalogHtmlPath` (**FU-23-A**) | **high** | mitigate | Empty-`catalogDir` rejection + `osutil.ContainsPath` (`app.go:759-789`); sibling-prefix and outside-dir tests pass | closed |
| **T-26-16** | Tampering | `CatalogModal` `srcDoc` (**T-22-05**) | **high** | mitigate | Component deleted; `openCatalogModal` has zero references; no `dangerouslySetInnerHTML`/`srcDoc` anywhere in `frontend/src` | closed |
| T-26-17 | Tampering | TOCTOU between check and open | low | accept | Documented and accepted, matching `RevealInFileManager`'s existing accepted risk (`openexternal.go:23-27`) | closed |
| T-26-18 | Information Disclosure | Open-external failure surface | low | mitigate | Failure surfaced via `setError`, same path as html-path failure (`DetailsPanel.tsx:130-137`) | closed |
| T-26-19 | Tampering | Four toggle bindings | low | mitigate | All accept booleans only — no path or privilege crosses | closed |
| T-26-20 | Elevation of Privilege | Scan output destinations | high | mitigate | `StartScan`'s output/secondary-copy destinations still gated by `osutil.ContainsPath`, unchanged by this phase (`app.go:266-296`) | closed |
| T-26-21 | Spoofing | Watch-toggle copy implying live watching | medium | mitigate | Note reads "applies once file watching ships"; no "watching" string in `StatusBar.tsx` | closed |
| T-26-22 | Denial of Service | `domReady` geometry restore | low | accept | Unchanged; still skips a `(0,0)` restore (`app.go:684-686`) | closed |
| T-26-23 | Tampering | Window-geometry write race | low | mitigate | Single `mu sync.RWMutex` guards every setter and `beforeClose`'s geometry write | closed |
| T-26-SC | Tampering | Supply chain (new dependency) | high | accept | `git diff --stat` across all 26-01…26-05 commits vs. `frontend/package.json`/`go.mod`/`go.sum`: zero changes — no package installed at any point in the phase | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above `workflow.security_block_on` (`high`) count toward `threats_open`*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

### Unregistered attack surface sweep

None found. No `26-0N-SUMMARY.md` carries a `## Threat Flags` section (verified by grep across all five).
The auditor additionally swept the phase's touched files for `eval`, `new Function`, `postMessage`,
`innerHTML`, and new `exec.Command` / `os/exec` call sites — none introduced.

---

## Milestone-carried obligations discharged

Both had been carried unresolved through Phases 22, 23, 24 and 25, and STATE.md recorded that they
must not be re-accepted a third time. Both are now closed with evidence.

- **T-22-05 → T-26-16.** Deletion is real (`git rm`; file absent on disk). Reachability re-confirmed
  independently at audit time: the only `window.dispatchEvent` anywhere in `frontend/src` is
  `themeChange` (2 call sites). `App.tsx` retains `ConfigProvider`/antd and the `themeChange` listener
  untouched — the deletion did not collaterally break antd's live usage.
- **FU-23-A → T-26-14 / T-26-15.** Both bindings take `catalogDir`, fail closed on empty, and reuse the
  single exported `osutil.ContainsPath` — no second containment implementation
  (`grep -nE 'strings\.HasPrefix\(.*catalogDir|filepath\.Rel' internal/osutil/openexternal.go` → no matches).
  `SearchIndexed`'s exclusion from the sweep is documented in `.planning/STATE.md` and `26-CONTEXT.md`
  with its Phase-25 rationale — correctly out of scope, not a silent gap.

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-26-01 | T-26-05 | Dialog z-order/scrim layering is enforced by the milestone's locked z-index scale rather than a runtime guard; a single-window desktop app has no cross-origin overlay attacker | plan-time disposition | 2026-08-15 |
| AR-26-02 | T-26-17 | TOCTOU window between the containment check and the open call. Matches `RevealInFileManager`'s existing accepted risk from Phase 23 — closing it would require an OS-level atomic open-by-handle the Wails runtime does not expose | plan-time disposition | 2026-08-15 |
| AR-26-03 | T-26-22 | `domReady` skipping a `(0,0)` geometry restore is pre-existing, unchanged behavior; worst case is a window opening at the default position | plan-time disposition | 2026-08-15 |
| AR-26-04 | T-26-SC | Supply-chain risk accepted as a standing phase disposition; verified moot — the phase installed no package (`go.mod`, `go.sum`, `frontend/package.json` all unchanged across every commit) | plan-time disposition | 2026-08-15 |

*Accepted risks do not resurface in future audit runs.*

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-08-15 | 24 | 24 | 0 | gsd-security-auditor (ASVS L1, block_on: high) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
