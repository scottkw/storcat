---
phase: 22-shell-token-layer
audited: 2026-08-13
asvs_level: 1
block_on: high
verdict: SECURED
threats_total: 20
threats_closed: 20
threats_open: 0
accepted_risks: 2
---

# Phase 22: Shell + Token Layer — Security Audit

**Verdict: SECURED** — 20/20 threats from the seven `22-0*-PLAN.md` `<threat_model>` blocks verified against implemented code. 0 open.

Verification was done against the code, not against plan text or the code review's own claims. Where a single grep would have been insufficient the auditor traced through — notably rebuilding `frontend/dist/` fresh against current HEAD rather than trusting the DEV-gate declaration, and following the `themeChange` CustomEvent payload back to its only emitter.

## Threat surface for this phase

This is the frontend shell of a local desktop app. In Phase 22 there is no network I/O, no authentication, no authorization, no secrets, no user-supplied content rendered, no server, no database, no frontend file writes, and no external API. The applicable ASVS L1 surface is narrow and is audited in full below; the remaining OWASP categories are genuinely N/A and are not padded into this report.

## Threat verification

| Threat ID (Plan) | Category | Component | Severity | Disposition | Evidence |
|---|---|---|---|---|---|
| T-22-01 (01) | Tampering | `readPersistedPrefs` allowlist | low | mitigate | `themeTokens.ts:216-251` — theme resolved via `getThemeById` (`themes.ts:453`, strict `.find()` against the fixed 11-entry table, `?? getDefaultTheme()` → `storcat-light`); density/rail-side compared to exact string literals; all `localStorage` access wrapped in try/catch (`safeGetItem`/`safeSetItem`, 186-208). The resolved theme id is used only as a table lookup key, **never interpolated into a style value** — so no CSS-injection path exists. |
| T-22-04 (01) | Injection | `components/workspace/`, `components/dev/` | medium | mitigate | Repo-wide grep of `frontend/src` for `dangerouslySetInnerHTML\|innerHTML\|outerHTML\|eval(\|new Function(` → zero matches. |
| T-22-05 (01) | Tampering | `CatalogModal` `srcDoc` iframe | low | **accept** | `App.tsx` wires `CatalogModal` only via the pre-existing `openCatalogModal` CustomEvent listener; grep confirms **no dispatcher of that event exists anywhere** in `frontend/src`. The component is genuinely unreachable this phase. See Accepted Risks. |
| T-22-SC (02) | Tampering (supply chain) | `@fontsource/ibm-plex-*` vendoring | high | mitigate | `git diff f5e09dc3..HEAD -- frontend/package.json frontend/package-lock.json` is empty. `frontend/src/assets/fonts/` holds exactly 5 IBM Plex `.woff2` + `IBM-Plex-OFL.txt`; `file(1)` confirms all 5 are genuine "Web Open Font Format (Version 2), TrueType" binaries, not disguised payloads. |
| T-22-02 (02) | Info disclosure | typography delivery | low | mitigate | `grep -rn "http://\|https://"` across `style.css`/`index.css`/`workspace.css` → nothing. All six `@font-face` `src` values are relative local paths. |
| T-22-06 (02) | Repudiation (licence) | vendored font files | low | mitigate | `IBM-Plex-OFL.txt` present and distinct from the pre-existing Nunito `OFL.txt`. |
| T-22-07 (03) | DoS (drag region) | toolbar drag region | medium | mitigate | `Toolbar.tsx`: 4 `.no-drag` occurrences on 4 interactive controls; `workspace.css:38-39` defines `.no-drag { --wails-draggable: no-drag; }`. |
| T-22-08 (03) | Spoofing | `Environment()` platform branch | low | **accept** | `Toolbar.tsx:22-32` — `env.platform === 'darwin'` string comparison sets a fixed `'78px'` constant; no untrusted value reaches a style write. Source is Go's own `GOOS` inside the same binary. See Accepted Risks. |
| T-22-04 (03) | Injection | `Toolbar.tsx` | medium | mitigate | Covered by the repo-wide zero-match grep; all copy is literal JSX text. |
| T-22-04 (04) | Injection | `CatalogRail.tsx`, `StatusBar.tsx` | medium | mitigate | Covered by repo-wide grep; the filter `<input>` (`CatalogRail.tsx:101-112`) is genuinely uncontrolled (no `value`/`onChange`). |
| T-22-09 (04) | Info disclosure | rail directory chip | low | mitigate | `CatalogRail.tsx:85-87` renders the literal `No catalog directory set`; no path is read or displayed. |
| T-22-04 (05) | Injection | `TreePane.tsx`, `DetailsPanel.tsx` | medium | mitigate | Covered by repo-wide grep; both files are 100% literal string children. |
| T-22-10 (05) | Info disclosure | details panel placeholder | low | mitigate | `DetailsPanel.tsx:27-29` renders a fixed sentence; no path, name or metadata. |
| T-22-01 (06) | Tampering | `initialState` seeding | low | mitigate | `AppContext.tsx:19-25` seeds only from `readPersistedPrefs()`; no raw `localStorage` read in the reducer or at module scope. |
| T-22-11 (06) | DoS (data loss) | reducer prune | medium | mitigate | `grep -rn "storcat-last-" frontend/src` → zero matches; tab-era keys untouched, as CONTEXT.md requires. |
| T-22-12 (06) | DoS | `useMediaQuery` subscription | low | mitigate | `useMediaQuery.ts:19-22` — `addEventListener('change', …)` with a symmetric `removeEventListener` in cleanup; no `resize` listener. |
| T-22-13 (07) | DoS | drawer close path | medium | mitigate | `WorkspaceShell.tsx:31,34-41` — one `closeDrawer` shared by Escape and backdrop click; the keydown listener is registered only inside the `if (!state.detailOverlay) return;` guard and removed in cleanup. |
| T-22-14 (07) | Elevation (z-index) | overlay stacking scale | high | mitigate | `workspace.css:29-31` declares exactly three named tiers (`--z-details-drawer:100`, `--z-overlay:200`, `--z-dialog:300`). The only component-level usage is `DevStateSwitcher.tsx:58` referencing `var(--z-dialog)`; no raw numeric literal anywhere under `components/`. |
| T-22-07 (07) | DoS (drag region) | Details chip drag region | medium | mitigate | Same 4-count `.no-drag` verification; the Details chip is included. |
| T-22-04 (07) | Injection | drawer/backdrop markup | medium | mitigate | Covered by repo-wide grep; the drawer reuses `DetailsPanel`'s literals and the backdrop `<div>` has no text content. |

## Independently verified beyond the register

- **Dev-affordance leakage.** `rm -rf frontend/dist && npm run build` fresh against current HEAD, then `grep -c "storcat-dev-switcher\|DevStateSwitcher" dist/assets/*.js` → **0**. The `import.meta.env.DEV` gate in `App.tsx:67` genuinely eliminates the module from the production bundle; the guard was not merely trusted to be present in source.
- **Go capability surface.** `git diff f5e09dc3..HEAD -- main.go` shows one change: `Mac: &mac.Options{TitleBar: mac.TitleBarHiddenInset()}`. No `WebviewIsTransparent`, no devtools/debug flags, no `AssetServer` change. Capability surface unchanged.

## Review-fix regression check

The four dispositions in `22-REVIEW-FIX.md` were confirmed intact at current HEAD (not re-litigated): no `catalogModalOpen`/`OPEN_CATALOG_MODAL` remnants in `AppContext.tsx` (WR-01); `Toolbar.tsx` takes `themeName` as a prop and `App.tsx`'s `handleThemeChange` persists `THEME_KEY` (WR-02); `Toolbar.tsx` uses `useLayoutEffect` (IN-02); no empty `.ws-details--pane` rule in `workspace.css` (IN-03).

## Accepted risks log

No `SECURITY.md` existed before this audit, so this is the first record. Both rationales below are corroborated by code, not merely asserted. Future phases should re-open them against these fixed points.

### T-22-05 — `CatalogModal` `srcDoc` iframe renders catalog-derived HTML
**Accepted because:** the component is unreachable in Phase 22. Its only trigger is the `openCatalogModal` CustomEvent, and no dispatcher of that event exists anywhere in `frontend/src` after the tab UI was deleted.
**Re-open when:** `CatalogModal` gets wired back up. Per `22-CONTEXT.md` that is Phase 26, and catalog HTML starts flowing in Phase 23 — so this must be re-audited no later than Phase 23's security pass, before any surface can render catalog-derived HTML.

### T-22-08 — `Environment().platform` trusted for the macOS toolbar inset
**Accepted because:** the value is compared with `===` against the literal `'darwin'` and, on match, writes a fixed `'78px'` constant. No untrusted value reaches a style write. The value originates from Go's own `GOOS` inside the same binary, so spoofing it requires already controlling the binary.
**Re-open when:** any platform-derived value is ever interpolated into a style, path, or command rather than used as a boolean branch.

## Unregistered flags

None. No `## Threat Flags` section appears in any of the seven `22-0*-SUMMARY.md` files — no new attack surface was self-reported by the executors, and independent review found none beyond the register above.
