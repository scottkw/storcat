---
status: testing
phase: 27-catalog-actions-watch
source: [27-VERIFICATION.md]
started: 2026-08-16
updated: 2026-08-16
---

## Current Test

number: 1
name: Watcher released on a real app quit (WATCH-03)
expected: |
  With watching enabled, quitting the BUILT app (not `wails dev`) and relaunching leaves no dangling
  watch descriptor from the prior session, and the app relaunches cleanly with no crash or hang.
awaiting: user response

## Tests

### 1. Watcher released on a real app quit (WATCH-03)
expected: With watching enabled, quit the built app and relaunch. No leaked watch handle, no crash, no hang. A fresh `lsof -p <pid>` shows no dangling watch descriptor from the prior session.
why_human: 27-07's live matrix (row 27) deliberately did not exercise a real quit — doing so would have killed the shared `wails dev` session the other 27 rows depended on. Substitute evidence (`lsof` fd count 101→74 on the mechanistically identical `Close()` from the toggle-off path) is reasoned, not a direct observation of the `OnShutdown` path.
result: [pending]

### 2. Concurrent DeleteCatalog calls for the same catalog (backstop)
expected: Two concurrent deletes of the same catalog cannot both report success while leaving a file behind.
why_human: Declared `verification: backstop` by the plans themselves — reasoned, not tested. Racing two real Wails binding calls needs a harness the phase did not build.
result: [pending]

### 3. Menu does not overflow the viewport (backstop)
expected: The actions menu stays within the viewport when the `⋯` trigger is near a window edge.
why_human: Declared `verification: backstop`. Needs a resized/edge-positioned window, which the dev-browser pass did not cover.
result: [pending]

### 4. Delete dialog busy-state visual treatment (backstop)
expected: While a delete is in flight, both footer buttons are visibly disabled and a second submit cannot fire.
why_human: Declared `verification: backstop`. The `busy` guard and `disabled={busy}` are code-verified; the visual treatment during the real in-flight window was not observed.
result: [pending]

### 5. `catalogs:changed` arriving during an in-flight rail fetch (backstop)
expected: A watch event landing mid-`browseCatalogs` does not produce a torn or stale listing.
why_human: Declared `verification: backstop`. Requires deterministic control over the event/fetch interleaving.
result: [pending]

### 6. RenameDialog failure sub-state, end to end
expected: A rename that fails (e.g. an unwritable `.html`) keeps the dialog open, shows the real error inline, and allows a retry.
why_human: Verified by direct binding rejection plus code-review parity with `Footer`'s proven pattern, not clicked through end-to-end in the dialog UI. Surfaced by the phase 27 UI review (priority fix 3).
result: [pending]

### 7. Windows / Linux platform behavior (WINDOWS.md 8–11)
expected: fsnotify's Windows rename-release behavior, both new dependencies' Windows and Linux backends, and the parent-directory fsync's Windows discard path all behave as reasoned.
why_human: No Windows or Linux machine was available this session. Tracked as ledger entries 8–11 for the pre-v3.0.0 sweep, not claimed as covered.
result: [pending]

## Summary

total: 7
passed: 0
issues: 0
pending: 7
skipped: 0
blocked: 0

## Gaps

Note: the verifier's original item #6 (rail staleness after Delete/Duplicate with watching off) is **not**
listed here — it was a real defect and was fixed rather than deferred (commit `5c7e460a`), then live-verified
with watching OFF. See the amendment in `27-VERIFICATION.md`.
