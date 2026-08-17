---
status: passed
phase: 26-settings
source: [26-VERIFICATION.md]
started: 2026-08-15
updated: 2026-08-16
---

## Current Test

number: 4
name: Planner-flagged assumptions
awaiting: none — all scenarios resolved

## Tests

### 1. Create-slide-over foreground-scan guard on ⌘, (T-26-03 DoS mitigation)
expected: While the create slide-over is actively `counting`/`scanning` a foreground scan, pressing ⌘, must be a complete no-op — Settings does not open, and the scan continues undisturbed.
why_human: Code is present and correctly wired (`WorkspaceShell.tsx`'s `openSettings` early-returns on `state.scan.status`), but no dev-browser session in this phase ever kept a real scan in `counting`/`scanning` long enough to trigger it live — the only reachable test volume resolved in under a second. Verified by code review only. Needs a large enough volume (or an artificially slow fixture).
result: [pass] Live-tested via dev-browser against `wails dev` :34115. Stubbed `window.go.main.App.StartScan` in-page to return a never-resolving promise (a controlled substitute for a slow real volume, keeping the frontend's `state.scan.status` genuinely in `counting` — this is real React/AppContext state, not a mock of the guard itself), opened the create slide-over, selected a real detected volume, and submitted. Confirmed `.ws-create-scan-body` showed "Counting…" (StartScan was actually invoked). Dispatched a real `KeyboardEvent('keydown', { key: ',', metaKey: true })` at `window` — `.ws-settings-panel` did NOT appear, and the scan body still read "Counting…" immediately after, confirming the guard is a genuine no-op with the scan undisturbed. Cleaned up: closed the panel (routes through `cancelScan`) and restored the real `StartScan` binding.

### 2. RailSide relaunch persistence with no visible flash of the other side
expected: With `railSide` set to a non-default value, a real OS-level quit-and-relaunch paints the workspace with the rail already on the persisted side — no frame where it briefly shows the other side.
why_human: 26-02 substituted a Go unit test plus a direct on-disk `config.json` read for a full OS quit/relaunch, to avoid disrupting the shared `wails dev` process later plans depended on. 26-05's restart proof confirmed the *value* persists but did not test the *no-flash* claim for rail side specifically. `.planning/WINDOWS.md` entry #7 remains open for this gap.
result: [pass] Already verified live during Phase 28-06's WINDOWS.md ledger sweep (entry #7, marked `fixed` 2026-08-16) — not re-run this session per the task's own instruction. That session toggled rail side Right→Left through the real Settings UI, issued a real `window.runtime.Quit()`, confirmed the process fully exited via `ps`/`lsof`, relaunched `wails dev`, and sampled `.ws-root`'s `data-rail-side` attribute at 50ms intervals for 2s from first paint: all 40 samples read `Left`, zero flash of `Right`. `config.json` read `railSide: Left` both before and after the quit. See `.planning/WINDOWS.md` entry #7 for the full evidence trail.

### 3. Copy-to-secondary toggle: first-enable opens the native OS folder picker
expected: In Settings, with no secondary directory ever configured, enabling "Copy catalogs to a secondary location" opens the real native folder-selection dialog; cancelling leaves the toggle off with no error; choosing a folder persists it and turns the toggle on.
why_human: Verified only by static code-path tracing — the project's standing no-host-OS-GUI-automation rule (recorded in STATE.md after a prior-phase incident) forbids driving a native dialog via CDP/keystrokes. Every other branch (reuse-with-no-dialog, disable-preserves-path) was proven live.
result: [pass] Live-tested via dev-browser against `wails dev` :34115, obeying the no-host-OS-GUI-automation rule by stubbing `window.go.main.App.SelectDirectory` in-page (the technique this phase's own instructions authorize for bypassing a native dialog) rather than driving the real OS picker. Cleared `secondaryDirectory`/`copyToSecondary` to the genuine "never configured" state via `SetSecondaryDirectory('')`, reloaded so `AppContext` picked up the clean config, opened Settings via a real `⌘,` keydown, and clicked the real "Copy catalogs to a secondary location" toggle row (`role="switch"`, real `.click()` dispatch). First branch: `SelectDirectory` was invoked exactly once (spy-counted), the chosen path persisted to both `settingsStore` and `config.json` (`GetConfig()` readback), and the toggle turned on (`aria-checked="true"`). Reset to never-configured again and repeated with `SelectDirectory` stubbed to resolve `''` (the real Wails cancel behavior): toggle stayed off (`aria-checked="false"`), note text reverted to "Choose a folder when enabled", no error surfaced, `config.json`'s `secondaryDirectory` stayed empty. Both sub-branches match the expected behavior exactly.

### 4. Planner-flagged assumptions: `SET-01 / unclassified` edge-probe row and the UI-SPEC UI-Considerations arithmetic mismatch
expected: Confirm the planner's own flagged assumptions are acceptable — (1) SET-01's entry-point/overlay-coexistence coverage in 26-01's acceptance criteria and 26-UI-SPEC's E1 `partial` row substitute adequately for a probe-derived predicate that could not be classified; (2) the one unaccounted backstop row in 26-UI-SPEC.md's header count (24 explicit + 2 backstop found vs. the header's claimed 3) was correctly left unminted rather than fabricated.
why_human: Explicitly surfaced no-silent-drop planner assumptions, directed at this step by name. Neither is a code defect — both are documentation/process judgment calls.
result: [pass] Reviewed both directly. (1) The arithmetic mismatch is already resolved in the document itself: `26-UI-SPEC.md` line 237-239 carries a dated "Count correction (2026-08-15, UI review)" note that corrects the header to "Applicable: 26 — resolved 26 (24 explicit, 2 backstop), unresolved 0" and explicitly records that the original "27...3 backstop" header did not sum against the five tables and was corrected rather than left wrong or fabricated to match. Nothing further to decide — it was already caught and fixed. (2) The SET-01/unclassified edge-probe substitute coverage is adequate: `26-UI-SPEC.md`'s E1 `partial` row (line 249) explicitly states "The ⌘, mutual-exclusion rules in the Dialog Shell section fully cover every combination of (palette open / create open in each step / nothing open) — no unresolved combination remains," and this session's own scenario 1 above directly and live-exercised exactly one of those combinations (a foreground scan blocking ⌘,), reinforcing that the substitute coverage reflects real, correct behavior rather than an untested claim. Both flagged assumptions are accepted as adequate; neither indicates a defect.

## Summary

total: 4
passed: 4
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

None. All four scenarios resolved to `pass`, one by citing already-completed live evidence from Phase 28-06 rather than a redundant re-run.
