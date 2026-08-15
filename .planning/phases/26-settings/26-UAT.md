---
status: testing
phase: 26-settings
source: [26-VERIFICATION.md]
started: 2026-08-15
updated: 2026-08-15
---

## Current Test

number: 1
name: Create-slide-over foreground-scan guard on ⌘, (T-26-03 DoS mitigation)
expected: |
  While the create slide-over is actively `counting`/`scanning` a foreground scan, pressing ⌘,
  must be a complete no-op — Settings does not open, and the scan continues undisturbed.
awaiting: user response

## Tests

### 1. Create-slide-over foreground-scan guard on ⌘, (T-26-03 DoS mitigation)
expected: While the create slide-over is actively `counting`/`scanning` a foreground scan, pressing ⌘, must be a complete no-op — Settings does not open, and the scan continues undisturbed.
why_human: Code is present and correctly wired (`WorkspaceShell.tsx`'s `openSettings` early-returns on `state.scan.status`), but no dev-browser session in this phase ever kept a real scan in `counting`/`scanning` long enough to trigger it live — the only reachable test volume resolved in under a second. Verified by code review only. Needs a large enough volume (or an artificially slow fixture).
result: [pending]

### 2. RailSide relaunch persistence with no visible flash of the other side
expected: With `railSide` set to a non-default value, a real OS-level quit-and-relaunch paints the workspace with the rail already on the persisted side — no frame where it briefly shows the other side.
why_human: 26-02 substituted a Go unit test plus a direct on-disk `config.json` read for a full OS quit/relaunch, to avoid disrupting the shared `wails dev` process later plans depended on. 26-05's restart proof confirmed the *value* persists but did not test the *no-flash* claim for rail side specifically. `.planning/WINDOWS.md` entry #7 remains open for this gap.
result: [pending]

### 3. Copy-to-secondary toggle: first-enable opens the native OS folder picker
expected: In Settings, with no secondary directory ever configured, enabling "Copy catalogs to a secondary location" opens the real native folder-selection dialog; cancelling leaves the toggle off with no error; choosing a folder persists it and turns the toggle on.
why_human: Verified only by static code-path tracing — the project's standing no-host-OS-GUI-automation rule (recorded in STATE.md after a prior-phase incident) forbids driving a native dialog via CDP/keystrokes. Every other branch (reuse-with-no-dialog, disable-preserves-path) was proven live.
result: [pending]

### 4. Planner-flagged assumptions: `SET-01 / unclassified` edge-probe row and the UI-SPEC UI-Considerations arithmetic mismatch
expected: Confirm the planner's own flagged assumptions are acceptable — (1) SET-01's entry-point/overlay-coexistence coverage in 26-01's acceptance criteria and 26-UI-SPEC's E1 `partial` row substitute adequately for a probe-derived predicate that could not be classified; (2) the one unaccounted backstop row in 26-UI-SPEC.md's header count (24 explicit + 2 backstop found vs. the header's claimed 3) was correctly left unminted rather than fabricated.
why_human: Explicitly surfaced no-silent-drop planner assumptions, directed at this step by name. Neither is a code defect — both are documentation/process judgment calls.
result: [pending]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps
