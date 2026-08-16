---
phase: 28-re-scan-diff
reviewed: 2026-08-16T23:00:42Z
depth: standard
files_reviewed: 23
files_reviewed_list:
  - app.go
  - app_test.go
  - frontend/src/components/workspace/DetailsPanel.tsx
  - frontend/src/components/workspace/UnreadableCatalogPanel.tsx
  - frontend/src/components/workspace/create/ErrorBody.tsx
  - frontend/src/components/workspace/create/ScanningBody.tsx
  - frontend/src/components/workspace/rescan/DiffList.tsx
  - frontend/src/components/workspace/rescan/RescanDialog.tsx
  - frontend/src/contexts/AppContext.tsx
  - frontend/src/services/wailsAPI.ts
  - frontend/src/types/rescan.ts
  - frontend/src/workspace.css
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/wailsjs/go/models.ts
  - internal/catalog/diff.go
  - internal/catalog/diff_test.go
  - internal/catalog/options.go
  - internal/catalog/resolve.go
  - internal/catalog/resolve_test.go
  - internal/catalog/service.go
  - internal/catalog/service_test.go
  - internal/catalog/walk.go
  - pkg/models/catalog.go
findings:
  critical: 1
  warning: 3
  info: 2
  total: 6
status: issues_found
---

# Phase 28: Code Review Report

**Reviewed:** 2026-08-16T23:00:42Z
**Depth:** standard
**Files Reviewed:** 23
**Status:** issues_found

## Summary

Reviewed the full re-scan/diff surface: the diff engine (`ComputeDiff`), the write-resolution path
(`WriteRescanResult`/`ResolveRescan`), the extracted `Walk` primitive, the `App` bindings
(`RescanCatalog`/`ResolveRescan`/`WritePartialCatalog` interaction), and the React dialog/panel layer
(`RescanDialog`, `DiffList`, `DetailsPanel`, `UnreadableCatalogPanel`).

The five highest-risk items called out in the phase context check out cleanly: `WriteRescanResult`
genuinely reuses `WriteCatalogFrom`'s atomic-write primitive and `nextCopyRoot` rather than opening a
second write/naming path; `walk.go`'s extraction is behavior-preserving and `MarkUnreadableOnSkip`'s zero
value leaves every Create call site untouched; `RescanCatalog` never touches
`lastPartial`/`lastPartialResult`/`lastScanReq` on any path (asserted by
`TestRescan_DoesNotRetainPartialForWritePartialCatalog`); and the DOM-rendering surfaces (`DiffList.tsx`,
`RescanDialog.tsx`) use React text children only, no `dangerouslySetInnerHTML`/`innerHTML`.

However, direct execution of `ComputeDiff` surfaced a real, provable correctness bug in the diff engine's
categorization order (CR-01) — a brand-new-but-currently-unreadable path is silently reported as `added`
instead of `unreadable`, hiding exactly the signal the phase exists to surface. A second issue in
`ResolveRescan` (WR-01) clears the retained re-scan tree before the write it enables is attempted, so any
write failure — even a non-destructive collision-probe failure with zero bytes written — strands the user
with no way to retry without a full re-scan, unlike the deliberately more careful `WritePartialCatalog`
pattern in the same file. Two smaller items round out the findings.

## Critical Issues

### CR-01: A brand-new unreadable path is miscategorized as `added`, hiding its ReadError

**File:** `internal/catalog/diff.go:74-84`
**Issue:**

`ComputeDiff`'s per-new-entry switch checks `!existed` (i.e. "not in the old tree" → `added`) *before* it
checks `newItem.Unreadable`:

```go
for path, newItem := range newFlat {
    oldItem, existed := oldFlat[path]
    switch {
    case !existed:
        entries = append(entries, models.DiffEntry{
            Path: path, State: models.DiffAdded, Type: newItem.Type, NewSize: newItem.Size,
        })
    case newItem.Unreadable:
        entries = append(entries, models.DiffEntry{
            Path: path, State: models.DiffUnreadable, Type: newItem.Type, ReadError: newItem.ReadError,
        })
    ...
```

Any path that is genuinely new (not present in the old catalog) *and* was marked `Unreadable` by this
walk (`MarkUnreadableOnSkip`) falls into the `!existed` branch and is reported `added` with
`NewSize: newItem.Size` — which is always `0` for an unreadable node — instead of `unreadable` with its
`ReadError`. The diff's own doc comment states an unreadable node "is categorized `unreadable`... an
unreadable node has no meaningful size or type comparison to make," but the switch order contradicts
that for the one case where the path is also new.

Reproduced directly against `ComputeDiff` (old tree empty, new tree has one directory
`Unreadable: true, ReadError: "permission denied"`):

```
Added=1 Unreadable=0 Entries=[{Path:./newlocked State:added Type:directory OldSize:0 NewSize:0 ReadError:}]
```

Consequences: the `DiffList`/`RescanDialog` UI shows this row under **Added** with a **0B** size and no
read-error text (`rightColumnFor` only surfaces `readError` for the `unreadable` state), the stat tile
counts (`Added++` instead of `Unreadable++`) are wrong, and `LowSimilarity`'s Added/Removed ratio is
computed against the wrong bucket. A user deciding whether to Overwrite/Keep-both is given no signal that
this "added, empty" entry is actually a path the walk could not read — directly undermining the phase's
stated data-integrity purpose (T-28-05).

No existing test in `diff_test.go` covers "unreadable AND not in the old tree" — every unreadable fixture
there (`TestDiff_UnreadableIsNotRemoved`, `TestDiff_UnreadableCarriesReadError`,
`TestComputeDiff_EndToEndWithRealUnreadableSubdirectory`) uses a path that already existed in the old
tree, so this branch ordering was never exercised.

**Fix:** Check `newItem.Unreadable` first, unconditionally, before the existence check:

```go
for path, newItem := range newFlat {
    oldItem, existed := oldFlat[path]
    switch {
    case newItem.Unreadable:
        entries = append(entries, models.DiffEntry{
            Path: path, State: models.DiffUnreadable, Type: newItem.Type, ReadError: newItem.ReadError,
        })
    case !existed:
        entries = append(entries, models.DiffEntry{
            Path: path, State: models.DiffAdded, Type: newItem.Type, NewSize: newItem.Size,
        })
    case oldItem.Type != newItem.Type:
        ...
```

Add a regression test (e.g. `TestDiff_NewUnreadableEntryIsUnreadableNotAdded`) with an old tree that has
no entry at the path and a new tree entry marked `Unreadable` at that same path, asserting
`result.Unreadable == 1` and `result.Added == 0`.

## Warnings

### WR-01: `ResolveRescan` clears the retained tree before attempting the write, losing retry on any write failure

**File:** `app.go:592-616`
**Issue:**

```go
a.scanMu.Lock()
if a.lastRescanTree == nil || a.lastRescanJSONPath != resolved {
    a.scanMu.Unlock()
    return nil, fmt.Errorf("resolve re-scan %s: no re-scanned tree held for this catalog -- re-scan again", jsonPath)
}
tree := a.lastRescanTree
a.lastRescanTree = nil
a.lastRescanJSONPath = ""
a.scanMu.Unlock()
...
result, err := a.catalogService.WriteRescanResult(tree, title, resolved, resolveMode, catalog.Options{})
if err != nil {
    return nil, fmt.Errorf("resolve re-scan %s: %w", jsonPath, err)
}
```

`a.lastRescanTree`/`a.lastRescanJSONPath` are cleared unconditionally *before* `WriteRescanResult` is
even called. If that call fails for any reason — including a `nextCopyRoot` collision-probe `os.Stat`
failure inside `resolve.go` that never reaches a single write, i.e. **zero bytes touched on disk** — the
in-memory tree is already gone. The frontend (`RescanDialog.handleResolve`) surfaces the real error via
`resolveError` but leaves the diff step's UI fully rendered with working Overwrite/Keep-both buttons
(`resolving` only guards against a second in-flight call, not against the tree having been consumed). A
retry click at that point fails with the unrelated, confusing `"no re-scanned tree held for this catalog
-- re-scan again"`, forcing the user to redo the entire (potentially slow) re-scan walk for a failure that
may not have written anything at all.

This is the opposite discipline from `WritePartialCatalog` in the same file, which deliberately reads the
retained state under `scanMu`, releases the lock, performs the (slow) write, and only clears/caches the
retained state **after** the write succeeds (see its own doc comment: "Idempotent... a second click a true
no-op rather than a duplicate write"). `ResolveRescan` has no equivalent pattern.

**Fix:** Move the clear to after a successful write, mirroring `WritePartialCatalog`'s check-decide-write
pattern — read `tree`/`resolved` under `scanMu`, release, call `WriteRescanResult`, and only clear
`lastRescanTree`/`lastRescanJSONPath` once `err == nil`:

```go
a.scanMu.Lock()
if a.lastRescanTree == nil || a.lastRescanJSONPath != resolved {
    a.scanMu.Unlock()
    return nil, fmt.Errorf("resolve re-scan %s: no re-scanned tree held for this catalog -- re-scan again", jsonPath)
}
tree := a.lastRescanTree
a.scanMu.Unlock()

...
result, err := a.catalogService.WriteRescanResult(tree, title, resolved, resolveMode, catalog.Options{})
if err != nil {
    return nil, fmt.Errorf("resolve re-scan %s: %w", jsonPath, err)
}

a.scanMu.Lock()
if a.lastRescanJSONPath == resolved {
    a.lastRescanTree = nil
    a.lastRescanJSONPath = ""
}
a.scanMu.Unlock()
return result, nil
```
(Guard the final clear against a superseded `lastRescanJSONPath`, the same `retainedGen`-style discipline
`WritePartialCatalog` uses, so a concurrent new `RescanCatalog` that started after this write isn't
clobbered.)

### WR-02: `ResolveRescan` trusts a frontend-supplied `catalogDir`, unlike its sibling `RescanCatalog`

**File:** `app.go:426-436` vs `app.go:564,569-587`
**Issue:** `RescanCatalog` derives `catalogDir` itself from `a.configManager.Get().CatalogDirectory` — the
renderer cannot influence which directory its containment check runs against. `ResolveRescan`, by
contrast, accepts `catalogDir` as a caller-supplied parameter (matching the pattern used by
`RenameCatalog`/`DuplicateCatalog`/`DeleteCatalog` elsewhere in this file) and validates `jsonPath`'s
containment against *that* value. A compromised renderer could pass an arbitrary `catalogDir` to make the
containment check trivially pass for a `jsonPath` outside the real configured directory.

In practice this is not currently exploitable: the write can only proceed if `resolved == a.lastRescanJSONPath`
exactly, and that value was set by a prior `RescanCatalog` call whose containment check ran against the
*real*, config-manager-derived `catalogDir` — so the actual write target is still constrained to the true
catalog directory regardless of what `catalogDir` this call receives. But the inconsistency between the two
sibling bindings in the same review scope is worth flagging: `ResolveRescan`'s own containment check is
decorative given the secondary exact-path guard, and a future edit that removes or loosens the
`lastRescanJSONPath` check (not unreasonable — WR-01's fix above touches this exact code) would silently
reopen a real containment gap.

**Fix:** For defense in depth, derive `catalogDir` inside `ResolveRescan` from `a.configManager` the same
way `RescanCatalog` does, rather than accepting it as a parameter — removing the reliance on the secondary
guard to keep this binding safe.

### WR-03: Dead CSS rule `.ws-rescan-error`

**File:** `frontend/src/workspace.css:2137`
**Issue:** `.ws-rescan-error` is defined but never referenced by any component — `RescanDialog.tsx` uses
`ws-rescan-resolve-error` (a different, similarly-named class defined separately at line 2229) for its
resolve-failure banner. `.ws-rescan-error` appears to be a leftover from an earlier naming iteration.
**Fix:** Remove the unused `.ws-rescan-error` rule.

## Info

### IN-01: `RescanCatalog`'s denominator always pays the full `MeasureTree` pre-pass cost

**File:** `app.go:502-506`
**Issue:** Unlike `StartScan` (which accepts `TotalBytesHint` from the frontend's already-known volume
size), `RescanCatalog`'s binding signature has no hint parameter, so every re-scan runs
`resolveScanTotal` with `hint == 0`, always paying a full `MeasureTree` count-only pass before the real
walk. This is out of scope as a performance finding per the review's stated v1 scope, but is worth noting
as a known gap versus Create's hint-aware path, since re-scans are re-run against sources the frontend
often already has a size for (the same `VolumePicker` sources Create uses).
**Fix:** Not required for this review; consider threading a size hint through `RescanCatalog` in a future
phase if re-scan performance on large volumes becomes a concern.

### IN-02: `ScanOptions.TotalBytesHint` zero-value ambiguity is documented but easy to violate silently

**File:** `app.go:116-123`
**Issue:** The doc comment explicitly calls out that `0` is overloaded ("the wire signal for 'denominator
unknown,' never a real hint of an empty volume"), which is a reasonable, deliberate design already in
place — not a phase 28 defect. Noted only because `RescanCatalog` (this phase) reuses the same
`resolveScanTotal` seam with a hardcoded `0`, inheriting the same ambiguity by construction; no action
needed beyond awareness for future call sites.

---

_Reviewed: 2026-08-16T23:00:42Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
