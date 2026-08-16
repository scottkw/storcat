---
phase: 27-catalog-actions-watch
reviewed: 2026-08-16T15:10:00Z
depth: standard
files_reviewed: 12
files_reviewed_list:
  - internal/catalog/atomicwrite.go
  - internal/catalog/rename.go
  - internal/catalog/rename_test.go
  - internal/catalog/duplicate.go
  - internal/catalog/duplicate_test.go
  - internal/watch/watcher.go
  - internal/watch/watcher_test.go
  - internal/osutil/trash.go
  - frontend/src/components/workspace/Menu.tsx
  - frontend/src/hooks/useModalBehavior.ts
  - frontend/src/components/workspace/DialogShell.tsx
  - app.go
findings:
  critical: 0
  warning: 1
  info: 0
  total: 1
status: issues_found
---

# Phase 27: Code Review Report

**Reviewed:** 2026-08-16T15:10:00Z
**Depth:** standard
**Files Reviewed:** 12
**Status:** issues_found

## Summary

Re-review iteration 2 of the `--auto` fix loop. All four applied fixes from iteration 1's review
(`27-REVIEW.iter2.md`) were re-verified directly against the current code, not taken on faith:

- **CR-01 (Menu.tsx focus restore):** confirmed correctly scoped — `event.preventDefault()` only runs on the
  "target is outside both `containerRef` and `triggerRef`" close branch, never on a click inside the menu or on
  the trigger itself. Per the scrutiny instruction, I traced through the side effect explicitly: because
  `preventDefault()` on a cancelable `pointerdown` suppresses the browser's compatibility `mousedown`/`click`
  dispatch for that gesture, a click on a *different* interactive control elsewhere in the app while the menu is
  open will close the menu but will **not** activate that control on the same click — a second click is needed.
  This is real, but it is the same "first click away only dismisses" behavior used by native OS context menus and
  every major popover/dropdown implementation (Radix, MUI, react-aria), it was already identified and explicitly
  accepted as a trade-off (not a regression) in the originating review, and no product spec here calls for
  click-through-to-target behavior. Not re-raised as a new finding.
- **WR-01 (watcher Errors dispatch):** `go w.c.fireNow()` is race-safe — `coalescer.stop()`'s `c.stopped` guard,
  read under the same mutex `fireNow()` locks, correctly prevents a late-spawned goroutine from firing after
  `Close()`. Verified with `go test ./internal/watch/... -race -count=3`: clean.
- **WR-02 (fsync log spam):** the `sync.Once` guard is exactly one of the two remedies the originating review
  itself proposed, is correctly placed at package scope, and is now unconditional-log-once rather than
  windows-only, so it still catches a first-time failure on any platform. It does narrow observability to "at
  most once ever, process-wide" rather than "once per distinct failure" (e.g., a second, unrelated persistent
  failure on a different catalog directory later in the same session would not be logged) — a real limitation, but
  the specific remedy applied here is one the prior review named as acceptable, so this is not re-raised as new.
- **WR-03 (`.html` sibling containment bypass):** `resolveContainedSibling` calls `osutil.ContainsPath` directly —
  it is not a second, weaker reimplementation, it *is* the same function `TrashPaths`/`GetCatalogHtmlPath`/
  `RenameCatalog`'s own primary-path check use. Confirmed via the regression tests
  (`TestRenameCatalog_RejectsHTMLSymlinkEscapingCatalogDir`, `TestDuplicateCatalog_RejectsHTMLSymlinkEscapingCatalogDir`)
  and `go build ./... && go vet ./... && go test ./... -race -count=1`, all clean.
- **WR-04 (DialogShell focus loss):** correctly left untouched; the speculated failure mode was live-verified not
  to reproduce, and `useModalBehavior.ts` is shared by every overlay in the app, so touching it speculatively would
  have been the higher-risk move.

While tracing WR-03's new code path, I found one **new, previously-unflagged, and experimentally confirmed**
correctness bug in `rename.go` that this fix pass touched but did not introduce (it predates WR-03; WR-03 just
makes it newly *reachable* via the symlink-escape rejection path it added, in addition to its pre-existing
plain-I/O-error triggers). Detailed below as WR-01 of this review.

## Warnings

### WR-01: `RenameCatalog` mutates the JSON title before validating the `.html` sibling, so a rejected rename leaves the JSON silently renamed anyway

**File:** `internal/catalog/rename.go:27-69`
**Issue:** `RenameCatalog` writes the new title to the JSON file (`WriteFileAtomic(jsonPath, ...)`, line 43) *before*
resolving or containment-checking the derived `.html` sibling (lines 47-57). If the sibling step then fails for any
reason — including the exact symlink-escape rejection `resolveContainedSibling` (WR-03) exists to produce, or a
plain permission error on the `.html` file — `RenameCatalog` returns a non-nil `error`, but the JSON file has
already been renamed. The caller (and, through it, the user) is told the rename failed, while the catalog's title
was in fact changed on disk.

I reproduced this experimentally against the current tree (test removed after confirming; not left in the repo):

```
$ go test ./internal/catalog/ -run TestRenameCatalog_PartialWriteOnHTMLEscapeRejection -v
    got expected error: rename .../photos.json: .../photos.html escapes catalog directory
    json title after rejected rename: "Photos 2024"
    BUG CONFIRMED: json title was mutated to "Photos 2024" despite RenameCatalog returning an error
```

`app.go`'s `RenameCatalog` binding and `wailsAPI.ts`'s `renameCatalog` wrapper both propagate only the error (Wails
discards a Go binding's non-error return value on the JS side when the promise rejects), so there is no
partial-success signal reaching the frontend either — the UI will show a rename-failed toast while the on-disk
title has already changed, and the discrepancy is only visible the next time the catalog is browsed/reloaded. This
is a logic/consistency bug, not the "report exactly what succeeded" pattern `DuplicateCatalog` deliberately
documents for itself (that function creates a *new* file on partial failure and still returns its path; this
function silently mutates the *existing* file with no equivalent signal).

**Fix:** Resolve/read the `.html` sibling before writing the JSON, so a rejected operation leaves nothing mutated:

```go
func RenameCatalog(jsonPath string, newTitle string) error {
	trimmed := strings.TrimSpace(newTitle)
	if trimmed == "" {
		return fmt.Errorf("rename %s: title is empty", jsonPath)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("rename %s: %w", jsonPath, err)
	}
	out, err := setTitleInDocument(data, trimmed)
	if err != nil {
		return fmt.Errorf("rename %s: %w", jsonPath, err)
	}

	// Resolve/validate/read the .html sibling BEFORE writing the JSON, so a
	// rejected sibling (symlink escape, permission error, ...) leaves the
	// JSON completely untouched instead of half-renamed.
	htmlPath := strings.TrimSuffix(jsonPath, ".json") + ".html"
	hasHTML := true
	resolvedHTMLPath, err := resolveContainedSibling(htmlPath, filepath.Dir(jsonPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("rename %s: %w", jsonPath, err)
		}
		hasHTML = false
	}
	var patchedHTML []byte
	if hasHTML {
		htmlData, err := os.ReadFile(resolvedHTMLPath)
		if err != nil {
			return fmt.Errorf("rename %s: %w", jsonPath, err)
		}
		patchedHTML = rewriteHTMLTitle(htmlData, trimmed)
	}

	if err := WriteFileAtomic(jsonPath, out, 0644); err != nil {
		return fmt.Errorf("rename %s: %w", jsonPath, err)
	}
	if hasHTML {
		if err := WriteFileAtomic(resolvedHTMLPath, patchedHTML, 0644); err != nil {
			return fmt.Errorf("rename %s: %w", jsonPath, err)
		}
	}
	return nil
}
```

This closes the containment/read failure class entirely (the dominant one, and the one WR-03 just added a new
trigger for) and narrows the residual gap to a genuine mid-write disk failure between the two `WriteFileAtomic`
calls, which is a much smaller and harder-to-avoid window inherent to any two-file update. Worth adding a
regression test alongside the fix (`TestRenameCatalog_RejectsHTMLSymlinkEscapingCatalogDir` currently only asserts
the outside file is untouched; it does not assert the JSON title is untouched — extending it, or adding a sibling
test, would catch a regression of this exact bug).

---

_Reviewed: 2026-08-16T15:10:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
