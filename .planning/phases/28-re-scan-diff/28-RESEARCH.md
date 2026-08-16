# Phase 28: Re-scan & Diff - Research

**Researched:** 2026-08-16
**Domain:** Go/Wails backend (tree diffing, mtime capture, atomic write reuse), React/TS frontend (already fully specified by 28-UI-SPEC.md), GitHub Actions CI verification
**Confidence:** HIGH (all findings verified by reading the actual source this session — `internal/catalog/service.go`, `errors.go`, `duplicate.go`, `pkg/models/catalog.go`, `app.go`, `.github/workflows/*.yml`, `frontend/src/**` — no external library research was needed; this phase adds zero new dependencies)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Diff semantics**
- The scanner gains mtime capture, and "changed" means size OR mtime differs. `CatalogItem` carries only `Size` today — `info.ModTime()` is never read (`internal/catalog/service.go:287-294`). Size-only comparison misses the same-size edit. The new field is `omitempty`, so catalogs without it stay byte-identical.
- Node identity is the existing `Name` field — full relative display path (`"./sub/file.txt"`), set once from `filepath.Rel` at traversal time. No model change needed for identity.
- An unreadable subtree is a FOURTH diff state, distinct from `removed`. A permission error or unreadable disc region is not the same event as "the user deleted this." Phase 25's flat `Unreadable bool` / `ReadError string` scalars stay as they are — the on-disk format is confirmed final; this phase adds a diff-level state, not an on-disk format change.
- Directories are diffed, but a directory is `changed` only when its own entry changes — never because a descendant changed.

**Re-scan flow & volume selection**
- Split the existing walk from the write, and export the walk half. `CreateCatalogWithContext` always ends in `WriteCatalogFrom` (`service.go:133,202`) but already calls `traverseDirectory` then `WriteCatalogFrom` as two cleanly separable steps — this is a split, not a rewrite. Do NOT write a second traversal implementation.
- Nothing about the chosen volume is persisted. ACT-08 says always ask. No volume name/UUID/serial is recorded anywhere in the catalog JSON today.
- A wrong-disc pick is shown honestly, never blocked. Near-total add+remove shows a plain warning above the diff when similarity is very low, but never refuses to proceed.
- Progress and cancellation reuse `scan:progress` and `CancelScan` unchanged. One scan at a time; existing `scanMu`/`activeScanCancel` guard (`app.go:317-343,404`) already covers a re-scan.

**Diff resolution (ACT-07)**
- Overwrite reuses `WriteCatalogFrom` wholesale — the scan's own write path, already routed through the Phase 27-hardened `WriteFileAtomic`.
- "Keep both" produces a new catalog beside the original using Phase 27's `-copy` suffix loop (`-copy`, `-copy-2`, …). Tested, existing.
- The `.html` sidecar is rewritten when the original had one, and never created where none existed.
- Discard writes nothing and leaves the original byte-identical, with no confirmation prompt.

**STATE-03 actions & COMPAT-06**
- "Remove from library" reuses Phase 27's `DeleteCatalog` delete-to-Trash, including its confirmation. No library-membership concept to invent.
- Re-scanning an unreadable catalog is a fresh scan, not a diff. Its JSON does not parse, so there is no old tree to compare against. Pick the volume, scan, offer overwrite or keep-both with no diff view.
- COMPAT-06 is proven by a real pushed CI run of `build.yml` across its existing matrix (macOS universal, Windows amd64, Linux amd64). Signing and notarization are NOT claimed by that run — they live in `release.yml` behind a tag, and Windows signing is currently a skip-with-warning branch (CRED-04/CRED-05 not provisioned). If no release tag is cut in this phase, the signing/notarization half is recorded as an open ledger item rather than claimed.
- The open `.planning/WINDOWS.md` entries are audited and disposed in this phase (11 open at scout time). Close what CI now genuinely proves, keep what truly needs Windows/Linux hardware.

### Claude's Discretion
- The mtime field's name, JSON tag, and time representation (RFC3339 string vs Unix seconds) — must be `omitempty` and must not disturb byte-parity for catalogs that lack it.
- The exported walk function's name and signature, and where the split lands.
- Diff result data shape crossing the Wails bridge, and whether counts are computed in Go or the frontend.
- The similarity heuristic and threshold behind the wrong-disc warning.
- Diff view presentation: grouping, ordering, virtualization.
- Whether the four diff states get their own color tokens or reuse existing ones.

### Deferred Ideas (OUT OF SCOPE)
- Content-hash-based change detection — re-reads every byte, unusable on optical/removable media.
- Persisting volume identity (name/UUID) in the catalog for a "re-scan the same disc" shortcut — contrary to ACT-08.
- Provisioning Windows OV signing credentials (CRED-04/CRED-05) — carried from v2.3.0 close, out of scope here.
- Three-way merge or per-entry selective reconciliation — the phase specifies overwrite / keep both / discard only.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ACT-06 | User can re-scan a catalog's source volume and see a diff of added, removed, changed, and unchanged entries with counts | Architecture Patterns §1–3 (walk/write split, mtime capture, diff algorithm) below give the exact Go seam and comparison rules |
| ACT-07 | User can resolve a diff by overwriting the catalog, keeping both, or discarding | Architecture Patterns §1 confirms `WriteCatalogFrom` and `nextCopyRoot`/`DuplicateCatalog`'s loop are directly reusable with no new write primitive |
| ACT-08 | Re-scan always asks the user to select the source volume rather than guessing which media the catalog came from | Confirmed no volume identity is persisted anywhere in `pkg/models/catalog.go` (verified this session) or `internal/config` — `VolumePicker.tsx` is a stateless controlled component, directly reusable per entry point |
| STATE-03 | User can re-scan, open the `.html` instead, or remove an unreadable catalog from the library | `UnreadableCatalogPanel.tsx:111-113` stub confirmed; `DeleteCatalog` (`app.go:993-1026`) and `handleOpenHtml` (`DetailsPanel.tsx:246-268`) are both directly reusable, no new Go binding needed for two of the three actions |
| COMPAT-06 | The app builds, signs, notarizes, and releases on all existing CI platform targets | Architecture Patterns §4 (CI Pipeline) documents exactly what `build.yml`/`release.yml`/`release-please.yml` prove and do not prove, and the concrete gap (local `main` is unpushed) that must be closed to claim this requirement honestly |
</phase_requirements>

## Summary

This phase has no new libraries to research — every piece is either stdlib (`os.FileInfo.ModTime()`, map-based set comparison) or an existing, already-tested primitive in this codebase (`WriteFileAtomic`, the `-copy` suffix loop, `DeleteCatalog`, `VolumePicker`, `ScanningBody`/`ErrorBody`). The real research work was tracing the exact code seams the four locked decisions depend on, and one of them surfaced a finding that changes scope: **the Unreadable/ReadError marker, as `traverseDirectory` is written today, is only ever set when the scan is aborted entirely** (a "skip and continue" single-bad-directory failure — the common case, and the only one a completed re-scan could show in its diff — is currently silently dropped with no marker at all). Read literally, CONTEXT's fourth diff state cannot be populated by anything short of a full scan abort unless this is changed. That gap, its cause, and a concrete non-breaking fix are detailed in Architecture Pattern 3 below — this is the single most important finding in this document and the planner must account for it as a real code change, not just new Go/TS glue.

Three other backend facts matter for planning: (1) the walk/write split is safe and mechanical if the split introduces a small new `Options` field rather than a parallel traversal — the risk is not in the split itself but in accidentally routing a re-scan's failure through Create's `lastPartial`/`lastScanReq` retention, which would silently make "write partial catalog" reachable for a flow that must never offer it; (2) `info.ModTime()` is already in scope at both branches of `traverseDirectory` with zero extra syscalls, and Unix-seconds is the safer wire representation given FAT32's 2-second mtime granularity; (3) `build.yml` has never actually run against this milestone's commits — local `main` is ~100 commits ahead of `origin/main` by deliberate design (`STATE.md`) — so COMPAT-06 cannot be claimed from static review of the YAML; it needs a real push.

**Primary recommendation:** Add one `Options` field (e.g. `MarkUnreadableOnSkip bool`) so re-scan's walk marks a skipped subtree `Unreadable` instead of silently dropping it, refactor `CreateCatalogWithContext`'s inline walk block into an exported `Walk` method that both Create and re-scan call (zero behavior change to Create since `MarkUnreadableOnSkip` defaults false), keep re-scan's own partial-failure state out of `a.lastPartial`/`a.lastScanReq`, capture `ModTime` as Unix seconds on both file and directory nodes at zero extra cost, diff via two path-keyed `map[string]*models.CatalogItem` built by a new flatten pass (trivial at 40k-node scale), and treat COMPAT-06 as requiring an actual `git push` to trigger a real Actions run, not a code-review exercise.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Write-free volume walk | API/Backend (`internal/catalog`) | — | Pure filesystem traversal; must stay usable from CLI-adjacent code paths with no Wails runtime (COMPAT-04 precedent) |
| mtime capture | API/Backend (`internal/catalog`) | — | `os.FileInfo` is already in hand inside `traverseDirectory`; no other tier touches raw filesystem stats |
| Diff computation (added/removed/changed/unreadable/unchanged) | API/Backend (new `internal/catalog` file or new package) | — | Operates on two in-memory `*models.CatalogItem` trees; no I/O of its own once both trees are loaded — kept pure and unit-testable, same pattern as the rest of `internal/catalog` |
| Old-tree loading (parse existing catalog JSON) | API/Backend (`internal/search.Service.LoadCatalog`) | — | Already exists, already handles v1/v2 format duality; the diff orchestrator (app.go) calls it, the diff algorithm itself never touches disk |
| Volume enumeration / selection UI | Frontend Server (Wails webview / React) | API/Backend (`internal/volumes`) | `VolumePicker.tsx` is already a stateless controlled component reusable verbatim; `internal/volumes.List()` is the existing backend source |
| Scan progress & cancellation | API/Backend (`app.go`'s `scanMu`/`activeScanCancel`) | Frontend Server (shared `state.scan` reducer slice) | Explicitly locked to reuse unchanged — no new tier ownership |
| Write resolution (overwrite / keep-both / discard) | API/Backend (`internal/catalog.WriteCatalogFrom`, `nextCopyRoot`) | — | Both already exist and are already crash-safe (`WriteFileAtomic`); no new write primitive needed |
| Remove-from-library | API/Backend (`app.go.DeleteCatalog`) | — | Reuses Phase 27's trash primitive verbatim; no new membership concept |
| Diff presentation, similarity warning, stat tiles | Frontend Server (React) | — | Fully specified by `28-UI-SPEC.md`; no backend involvement beyond the diff payload shape |
| CI build/sign/notarize/release proof | CDN/Static (GitHub Actions, out-of-repo infra) | — | Exists already in `.github/workflows/`; this phase's job is to actually exercise it against this milestone's commits, not to author new pipeline code |

## Standard Stack

No new dependencies, any ecosystem. `go.mod` verified this session (`/Users/ken/dev/storcat/go.mod`): `github.com/djherbis/times v1.6.0` is already present but is used exclusively for the *catalog file's own* creation-time display in `internal/search/service.go:224` (`BrowseCatalogs`'s rail metadata) — it is unrelated to per-entry mtime capture inside the scanned tree and does not need to be touched or extended for this phase. Per-entry mtime is captured via the stdlib `os.FileInfo.ModTime()` already returned by the `os.Stat` call `traverseDirectory` makes today (`service.go:268`).

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (`os`, `path/filepath`, `context`) | go1.23.4 (per `go.mod`'s `go 1.23.4` directive; local toolchain verified this session: `go version go1.26.6 darwin/arm64`, satisfies the directive) | Filesystem walk, mtime, path identity | Already the entire foundation of `internal/catalog`; no reason to introduce anything else for a stat-compare-diff problem at 40k-node scale |

### Supporting
None — no supporting library is needed. The diff is a map-based set comparison over data already in memory once both trees are loaded.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled path-keyed map diff | A general-purpose Go diff/patience-diff library | Rejected: this is not a line-based text diff, it's a keyed set comparison — a text-diff library would be solving the wrong problem and adding a dependency for no benefit |
| `os.FileInfo.ModTime()` | `github.com/djherbis/times` (already a dependency) | `times` exists specifically to recover *birth time* (creation time) on platforms where stdlib can't (macOS/Windows). Re-scan only needs *modification* time, which stdlib's `ModTime()` already returns cross-platform with no extra dependency surface — using `times` here would be reaching for a birth-time tool to solve an mtime problem |

**Installation:** None required — zero new packages, any ecosystem.

## Package Legitimacy Audit

**Not applicable — this phase adds no new external dependencies, frontend or backend.** Confirmed by `28-UI-SPEC.md`'s own Registry Safety section ("this phase adds none, frontend or backend, beyond a possible new mtime-capture field on the existing scanner") and independently verified by reading `go.mod` and `frontend/package.json` this session: every capability this phase needs (walk splitting, mtime, diffing, write reuse, trash reuse) is satisfiable with code already present or trivial additions to code already present.

## Architecture Patterns

### System Architecture Diagram

```
                     ┌─────────────────────────────────────────────┐
                     │  RescanDialog (frontend, fully specified     │
                     │  by 28-UI-SPEC.md — not re-derived here)     │
                     └───────────────┬───────────────────────────────┘
                                     │ Step 1: pick volume (VolumePicker, reused)
                                     ▼
                     wailsAPI.startRescan(jsonPath, sourcePath, opts)
                                     │
                                     ▼
        ┌────────────────────────────────────────────────────────────┐
        │ app.go — new binding (name at planner's discretion,         │
        │ e.g. RescanCatalog), mirrors startScan's shape:              │
        │  1. scanMu-guarded one-scan-at-a-time gate (reused)          │
        │  2. a.catalogService.Walk(ctx, sourcePath, opts, onProgress) │◄── NEW exported method
        │     — walk only, no write. On success: *models.CatalogItem  │     (internal/catalog)
        │     On terminal failure: same SourceUnavailableError shape   │
        │     Does NOT touch a.lastPartial/a.lastScanReq (Create-only) │
        │  3. a.searchService.LoadCatalog(existingJSONPath) — old tree │◄── EXISTING (internal/search)
        │     (skipped entirely when oldTreeAvailable == false,        │
        │      i.e. STATE-03's unreadable-catalog path)                │
        │  4. diff.Compute(oldTree, newTree) — pure, in-memory          │◄── NEW (internal/catalog
        │     returns {added, removed, changed, unreadable, unchanged}  │     or new internal/diff)
        └───────────────────────────┬────────────────────────────────┘
                                     │ diff result (or reduced summary, Variant B)
                                     ▼
                     RescanDialog Step 3 (stat tiles, diff list, similarity warning)
                                     │
                    ┌────────────────┼────────────────────┐
                    ▼                ▼                     ▼
              Overwrite         Keep both              Discard
        WriteCatalogFrom(   nextCopyRoot +          (no Go call —
        newTree, ...)       WriteCatalogFrom(       dialog closes,
        — EXISTING          newTree, ...)           nothing written)
        (service.go:211)    — EXISTING primitives
                             (duplicate.go's loop,
                             reused for the target
                             filename only)
```

### Recommended Project Structure
```
internal/catalog/
├── service.go       # CreateCatalogWithContext refactored to call the new Walk method internally
├── walk.go           # NEW — exported Walk(ctx, sourcePath, opts, onProgress) (*models.CatalogItem, *PartialScan, error)
├── diff.go            # NEW — pure diff.Compute(old, new *models.CatalogItem) (*DiffResult, error); no I/O
├── options.go (or inline in service.go) # Options gains MarkUnreadableOnSkip bool
├── errors.go          # unchanged
└── duplicate.go       # unchanged, reused by keep-both's target-name resolution

pkg/models/
└── catalog.go         # CatalogItem gains ModTime int64 `json:"modTime,omitempty"`; new DiffResult/DiffEntry types

app.go                 # new binding(s): RescanCatalog (walk + diff), ResolveRescan (overwrite/keep-both/discard)
```

### Pattern 1: The Walk/Write Split — Exact Seam

**What:** `CreateCatalogWithContext` (`service.go:127-203`) currently interleaves the walk, its error classification (lines 140-192), and the final `WriteCatalogFrom` call (line 202) in one function body. The classification logic — distinguishing a cancelled context, a `SourceUnavailableError` with populated `Partial`, and a genuine traversal failure — is itself substantial and must not be duplicated for re-scan.

**The split, concretely:** extract lines 134-200 (walk + classification, everything before the final `WriteCatalogFrom` call) into a new exported method on `*Service`:

```go
// Source: internal/catalog/service.go:133-203 (existing code being restructured, read this session)
// Walk builds sourcePath's tree without writing anything. It is the exact
// walk+classification logic CreateCatalogWithContext has always run
// inline; CreateCatalogWithContext now calls this method instead of
// duplicating it, so Create's own behavior is provably unchanged.
func (s *Service) Walk(ctx context.Context, sourcePath string, opts Options, onProgress ProgressCallback) (*models.CatalogItem, error) {
	st := &walkState{scanRoot: sourcePath, opts: opts, onProgress: onProgress}
	tree, err := s.traverseDirectory(ctx, sourcePath, sourcePath, st)
	if err != nil {
		var srcErr *SourceUnavailableError
		if errors.As(err, &srcErr) {
			srcErr.Partial = &PartialScan{Tree: tree, FilesSeen: st.filesSeen, BytesSeen: st.bytesSeen, ReadErrors: st.readErrorEntries}
			return nil, srcErr
		}
		if ctx.Err() == nil && st.classify() {
			return nil, &SourceUnavailableError{
				SourcePath: sourcePath,
				Partial: &PartialScan{
					Tree:       &models.CatalogItem{Type: "directory", Name: "./", Size: 0, Contents: []*models.CatalogItem{}, Unreadable: true, ReadError: err.Error()},
					FilesSeen:  st.filesSeen,
					BytesSeen:  st.bytesSeen,
					ReadErrors: st.readErrorEntries,
				},
			}
		}
		return nil, fmt.Errorf("failed to traverse directory: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return tree, nil
}

// CreateCatalogWithContext now reduces to:
func (s *Service) CreateCatalogWithContext(ctx context.Context, title, sourcePath, outputDir, outputRoot, copyToDirectory string, opts Options, onProgress ProgressCallback) (*models.CreateCatalogResult, error) {
	tree, err := s.Walk(ctx, sourcePath, opts, onProgress)
	if err != nil {
		return nil, err // srcErr's Partial is already populated by Walk, identical to today's inline behavior
	}
	return s.WriteCatalogFrom(tree, title, outputDir, outputRoot, copyToDirectory, opts)
}
```

This is behavior-preserving for Create by construction — every branch's return value and side effect (counter mutation on `st`, error wrapping) is copied verbatim, not reimplemented. A test asserting `TestCreateCatalog_JSONShapeUnchanged` (already exists, `service_test.go:373`) and the source-loss tests (`service_test.go:562,623`) are the regression backstop; they must keep passing unmodified after this refactor — if any of them needs an edit to pass, the split introduced a real behavior change and must be reconsidered.

**When to use:** Re-scan calls `Walk` directly and never calls `WriteCatalogFrom` until the user picks a resolution.

**What `WriteCatalogFrom` needs:** Nothing new. Its signature (`tree *models.CatalogItem, title, outputDir, outputRoot, copyToDirectory string, opts Options`) already accepts a pre-built tree — re-scan's Overwrite path calls it with the newly-walked tree and the *original* catalog's own `outputDir`/`outputRoot` (derived from the existing catalog's JSON path, same as `DuplicateCatalog` already derives `dir`/`root` from a `jsonPath` at `duplicate.go:33-34`). Keep-both calls it with `nextCopyRoot`'s resolved candidate as `outputRoot` instead.

**Risk flagged — the one thing that makes this split riskier than it looks:** `a.startScan` (`app.go:317-335`) unconditionally clears `a.lastPartial`/`a.lastPartialResult`/`a.lastScanReq` at the start of every new scan and, on a `SourceUnavailableError`, repopulates them (`app.go:372-385`) so `WritePartialCatalog` (`app.go:461-501`) can later write that retained tree. This mechanism exists **only** to support Create's CRT-11 "write partial catalog" button. `28-UI-SPEC.md`'s Architecture & State section explicitly states re-scan's error step has **no** partial-write option — "Retry scan" and "Close without writing" only. If re-scan's new binding is implemented by copy-pasting `startScan`'s preamble (the easy, tempting path, since CONTEXT explicitly says to reuse `scanMu`/`activeScanCancel`), it will also copy the `a.lastPartial = srcErr.Partial; a.lastScanReq = &lastScanRequest{...}` assignment on failure — silently making a *re-scan's* partial tree writable through the pre-existing `WritePartialCatalog` binding, which the re-scan UI never calls but which remains reachable (e.g. by a future regression, a stray keyboard shortcut, or dev-tools). **The fix is to make the re-scan binding call `a.catalogService.Walk` directly (never `a.startScan`/`a.catalogService.CreateCatalogWithContext`) and to deliberately not touch `a.lastPartial`/`a.lastScanReq` on failure** — since re-scan offers no partial-write action, no retention is needed at all; a failed re-scan simply returns the error, the dialog shows "Retry" (which re-walks from scratch) or "Close" (which discards). This is a one-line omission (don't run that assignment block) but it is exactly the kind of self-inflicted regression a "just reuse `startScan`'s shape" refactor produces if not called out explicitly.

**Error/cancellation semantics preserved:** `traverseDirectory` checks `ctx.Err()` at the top of every recursive call (`service.go:264-266`), so `CancelScan` (already reused unchanged, `app.go:404-406`) cancels a re-scan's walk exactly as it does Create's — and because the tree is built entirely in memory before any write (the same invariant the doc comment at `service.go:130-132` calls load-bearing), "cancelling a re-scan writes nothing and shows no diff" holds for free, matching the UI-SPEC's cancellation contract.

### Pattern 2: mtime Capture

**What:** `traverseDirectory`'s `os.Stat(dirPath)` call (`service.go:268`) already returns an `os.FileInfo` with `.ModTime()` available — for both the file branch (`service.go:286-294`) and the directory branch (`service.go:298-411`), the `info` variable is in scope before either branch is entered. Capturing mtime costs **zero additional syscalls**.

```go
// Source: internal/catalog/service.go:286-294 (existing code, read this session — the exact
// insertion point; info.ModTime() is already available, nothing new to fetch)
if info.Mode().IsRegular() {
	st.filesSeen++
	st.bytesSeen += info.Size()
	st.report(displayPath)
	return &models.CatalogItem{
		Type:    "file",
		Name:    displayPath,
		Size:    info.Size(),
		ModTime: info.ModTime().Unix(), // NEW — zero extra syscalls, info already in scope
	}, nil
}
```

The same applies to the directory branch's `info` (also already stat'd at the top of the same function call) — capturing a directory's own mtime is equally free and is what makes CONTEXT's "a directory is `changed` only when its own entry changes" rule concretely checkable (a directory's `ModTime` reflects its own immediate-children add/remove event on most filesystems, distinct from any content change deep inside it).

**Time representation — recommend Unix seconds (`int64`), not RFC3339 string:**
- **Byte size:** an `int64` Unix-seconds value serializes as up to 10 ASCII digits (`"modTime":1755350020`); an RFC3339 string is ~20-22 bytes (`"modTime":"2026-08-16T14:53:40Z"`) plus quoting. At 40k+ nodes this is a real, if modest, JSON-size difference — Unix seconds wins on the byte-parity-adjacent goal CONTEXT already cares about (keeping catalogs lean).
- **Cross-platform stability — the actual reason to prefer whole-second precision, verified against known filesystem behavior:** FAT32 (the format most SD cards and USB sticks ship with, and this app's primary target medium) stores modification time with **2-second granularity** — a well-documented FAT32 limitation, not something this app can work around. Optical media (ISO9660/UDF) frequently reports a single burn-time timestamp for every file on the disc, sometimes with only 1-second or coarser resolution. Network mounts (SMB/NFS) can introduce clock-skew-driven or sub-second-precision differences depending on protocol version. Truncating to whole seconds (what `.Unix()` does) matches the coarsest common denominator across these targets and avoids **false-positive "changed" diagnoses** caused by sub-second jitter that reflects nothing about the file's actual content — e.g. two reads of the identical unchanged file returning slightly different sub-second components depending on which layer (kernel cache vs. fresh stat) served the read.
- **Accepted residual, worth documenting rather than silently absorbing:** the flip side of FAT32's 2-second bucket is a **false-negative** risk — a genuine same-size edit that completes within the same 2-second rounding window as the file's prior write can be invisible to size+mtime comparison. This is the same tradeoff class CONTEXT's Deferred Ideas section already accepted when it rejected content-hashing ("hashing 40k files on removable media is prohibitively slow") — not a new gap this phase introduces, but worth naming explicitly as a known limitation of the locked size+mtime approach rather than leaving it implicit.
- **`omitempty` behavior:** `ModTime int64 \`json:"modTime,omitempty"\`` — a catalog written before this field existed simply omits the key (Go's `omitempty` treats zero-value `int64(0)` as absent), which the diff algorithm must handle explicitly: **an old-tree entry with `ModTime == 0` is not "modified at the Unix epoch," it is "mtime unknown" — the comparison must degrade to size-only for that specific entry**, never treat a just-added field as itself evidence of change.

### Pattern 3: The Diff Algorithm — And a Load-Bearing Gap It Surfaces

**What:** Confirmed via grep this session — no diff code exists anywhere in `internal/`, `pkg/`, or `frontend/src`. This is genuinely new.

**The shape — confirmed correct, standard approach:** flatten both trees into `map[string]*models.CatalogItem` keyed by `Name` (the existing full relative path, e.g. `"./sub/file.txt"` — already the diff key CONTEXT identifies, no model change needed for identity), then a straightforward set comparison:

```go
// Sketch — internal/catalog/diff.go (new file). Pattern only, not verbatim source (no
// equivalent function exists yet to quote from).
func flatten(root *models.CatalogItem, out map[string]*models.CatalogItem) {
	for _, child := range root.Contents {
		out[child.Name] = child
		if child.Type == "directory" {
			flatten(child, out)
		}
	}
}
// oldFlat, newFlat := map[string]*models.CatalogItem{}, map[string]*models.CatalogItem{}
// flatten(oldTree, oldFlat); flatten(newTree, newFlat)
// for path in union(keys(oldFlat), keys(newFlat)): categorize into added/removed/changed/unreadable/unchanged
```

This mirrors `internal/search/flatten.go`'s existing `LoadCatalogFlat` walk (root excluded, direct children recursed) closely enough that the diff's own flatten should follow the same "exclude the root itself" convention for consistency — comparing `"./"` against `"./"` between two scans is a meaningless always-unchanged no-op entry, same reasoning `LoadCatalogFlat` already applies. `LoadCatalogFlat` itself is not reusable as-is: its `FlatNode` output type carries `Depth`/`ParentIdx`/`HasChildren` for tree-pane rendering and deliberately drops `ModTime`/`Unreadable`/`ReadError` — the diff needs a different, comparison-oriented shape, not `FlatNode`.

**Memory profile at 40k+ nodes:** trivial. Two maps of ~40,000 pointer-valued entries each (a `*models.CatalogItem` pointer, 8 bytes, plus a string key averaging maybe 40-80 bytes for a realistic path) is on the order of a few MB total, including Go map overhead. No streaming, chunking, or special-cased large-catalog handling is warranted — this is well within what a straightforward in-memory map handles without measurable latency.

**Where to put the diff package:** keep it dependency-free — `internal/catalog` (or a new `internal/diff` package importing only `pkg/models`) operating purely on two already-loaded `*models.CatalogItem` trees, no I/O of its own. Loading the *old* tree from disk is `internal/search.Service.LoadCatalog` (`internal/search/service.go:178`, already exists, already handles the v1/v2 array/object format duality) — confirmed no import cycle exists between `internal/catalog` and `internal/search` (`internal/catalog` imports only `internal/osutil` and `pkg/models`; `internal/search` imports only `internal/config` and `pkg/models`) — but the cleanest design keeps the diff algorithm itself importing neither package, and lets `app.go` (which already holds both `a.catalogService` and `a.searchService` instances) orchestrate: load old tree → walk new tree → call `diff.Compute(oldTree, newTree)`.

**The load-bearing gap — read carefully, this changes what "the split" needs to expose:**

CONTEXT frames the unreadable diff state as interacting with "Phase 25's marker [that] sits on exactly ONE origin node per partial scan, never propagated to ancestors." Tracing `traverseDirectory`'s actual branches this session (`service.go:298-411`) shows the marker is set in exactly two places, and **both require `st.classify()` to return `true`, which only happens when the SCAN ROOT ITSELF has also become unreachable** (`walkState.classify()`, `service.go:95-103`, re-probes `st.scanRoot` via a cheap `os.Stat`, not the failing subtree). When the root is still reachable — the overwhelmingly common case, e.g. one permission-denied subdirectory on an otherwise-healthy re-scan — the code takes the **other** branch: `recordReadError` then `continue`/`return node, nil` with **no marker set at all** (`service.go:318-320`, comment: *"Skip items we can't access -- unchanged byte-for-byte behavior when the root is still reachable"*). Critically, whenever `classify()` *does* return true, the resulting `SourceUnavailableError` **propagates all the way up and aborts the entire walk** — it does not let the walk continue past that one subtree and finish scanning the rest of the volume.

The practical consequence: **as `traverseDirectory` is written today, a completed (non-aborted) re-scan can never contain an `Unreadable`-marked node.** Every single-item/single-subtree read failure that doesn't also take down the scan root is silently dropped with zero marker — indistinguishable, once diffed, from "the user deleted this." The only way the marker is ever set is on the terminal node of a scan that gets aborted outright — which routes to the Error Step ("Scan interrupted"), never reaches the Diff & Resolve step at all. Read literally, CONTEXT's fourth diff state has no data path that can ever populate it.

**What this means for the diff's "unreadable" interaction with an origin-node-only, non-propagated marker** (the objective's specific question): for a subtree the new scan genuinely cannot see into (root reachable, one bad directory), the diff **cannot currently distinguish that from "removed"** — because nothing marks it. Fixing this requires a real, scoped code change to the walk, not just new diff/model code:

**Recommended fix — one new `Options` field, gated so Create's behavior is provably unchanged:**
```go
// Options (existing struct, service.go — exact current fields not re-quoted here since this
// is a proposed addition, not a verified read of an existing field)
type Options struct {
	WriteHTML          bool
	IncludeHidden       bool
	HaltOnSourceLoss     bool
	MarkUnreadableOnSkip bool // NEW — when true, a skip-and-continue single-entry/subtree
	                          // failure (root still reachable) also sets Unreadable/ReadError
	                          // on that node instead of silently dropping it. Defaults false,
	                          // so Create's existing behavior (CLI: false via CreateCatalog's
	                          // wrapper; GUI Create: also false, only re-scan sets it true) is
	                          // byte-for-byte unchanged.
}
```
`traverseDirectory`'s two `// Skip items we can't access` branches (`service.go:318-320` and `service.go:392-394`) would each gain an `if st.opts.MarkUnreadableOnSkip { node.Unreadable = true; node.ReadError = err.Error() }` before the existing skip/continue, and the walk would **not** abort (no `SourceUnavailableError` returned in this branch) — the scan continues past the bad subtree exactly as it does today, it just now leaves a marker behind instead of silent absence. This is on-disk-format-compatible (the `Unreadable`/`ReadError` fields already exist, `omitempty`, per `pkg/models/catalog.go:32-37`, confirmed this session) and matches CONTEXT's explicit statement that the on-disk format is unchanged — only *when* those existing fields get set changes, and only for callers that opt in via the new flag. **Flagging this clearly for the planner: this is a real behavior change to `traverseDirectory`'s skip-and-continue path, not glue code — it must appear in the plan's task list as its own reviewable change, with its own test (e.g. `TestTraverseDirectory_MarksSkippedNodeWhenFlagSet` alongside the existing `TestTraverseDirectory_SingleEntryErrorSkipsAndContinues`, which must keep passing unmodified to prove the default-false path is untouched).**

**Directory-vs-file type change edge case (not addressed by CONTEXT, flagged as an open question):** if a path is a `file` in the old tree and a `directory` in the new tree (or vice versa) — same `Name` key, different `Type` — CONTEXT's decision tree doesn't cover this explicitly. The diff-row schema (mark + path + size, or `old→new` size) doesn't have a natural rendering for "type changed." Recommended default: treat as `removed` (old) + `added` (new) — two separate rows — since it's effectively a different entity that happens to share a path, not a comparable edit. This is a genuinely rare case (SD cards being re-purposed with a file where a folder used to sit) but should be an explicit, tested branch rather than an accidental fallthrough.

### Anti-Patterns to Avoid
- **A second, parallel traversal function for re-scan:** CONTEXT explicitly forbids this ("two walkers will drift on hidden-file handling, symlinks, and error classification") — the `Walk`-extraction approach above is the only sanctioned path.
- **Routing re-scan's failures through `a.lastPartial`/`a.lastScanReq`:** see Pattern 1's risk callout — this is the single easiest way to accidentally regress CRT-11's semantics for a flow that must never offer partial-write.
- **Comparing `ModTime == 0` as a real timestamp:** an old catalog predating this field has `ModTime` absent/zero; the diff must treat that as "unknown," not "epoch," or every pre-existing catalog's first re-scan will spuriously flag every file as `changed`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Crash-safe catalog write | A second atomic-write path for re-scan's Overwrite | `WriteFileAtomic` via `WriteCatalogFrom` (`atomicwrite.go`, already hardened through Phase 27's SIGKILL-tested `tmp.Sync()` → `os.Rename` → best-effort `syncDir` sequence) | CONTEXT locks this reuse explicitly; a second write path is a second thing that can drift from the crash-safety guarantee ACT-09 already proved |
| Collision-safe "keep both" filename | A new collision-check/suffix algorithm | `nextCopyRoot`/`isCandidateRootFree` (`duplicate.go:90-120`, already tested, capped at 1000 candidates) | Already handles both `.json` and `.html` collision checking; a second implementation risks silently diverging on the exact suffix format (`-copy`, `-copy-2`, …) users already learned from Duplicate |
| Catalog removal | A new "hide from library" flag or excluded-list concept | `DeleteCatalog` (`app.go:993-1026`) → OS Trash via `osutil.TrashPaths` | CONTEXT explicitly rejects inventing a membership concept — "membership IS the file living in the configured catalog directory" |
| Old catalog JSON loading | A second JSON parser for the diff's "old tree" side | `internal/search.Service.LoadCatalog` (`service.go:178`) | Already handles v1 (array) and v2 (object) format duality; a second parser is a second place those two formats can silently diverge |
| Path-level tree comparison | A general-purpose tree/diff library | A ~30-line path-keyed map set-comparison (Pattern 3 above) | The problem is a keyed set comparison, not a structural tree-edit-distance problem — no existing library models this app's specific four/five-state semantics anyway |

**Key insight:** every write, every collision-safe filename, and every deletion this phase needs already exists and is already tested from Phases 25-27. The only genuinely new code is the walk/write split, mtime capture, and the diff algorithm itself — everything else is reuse, and the plan should treat any deviation from that reuse list as a red flag to double-check against CONTEXT.

## Common Pitfalls

### Pitfall 1: The Unreadable Marker Never Fires On a Completed Scan
**What goes wrong:** The planner builds the diff's fourth `unreadable` category, wires up the stat tile and diff-row rendering exactly as `28-UI-SPEC.md` specifies, ships it — and it's silently unreachable, because nothing in the current walk ever sets the marker on anything but an aborted scan's terminal node.
**Why it happens:** `classify()`'s scan-root re-probe conflates "this one subtree failed" with "the whole volume is gone" — only the latter sets the marker, and it also aborts the walk.
**How to avoid:** Implement the `MarkUnreadableOnSkip` `Options` field (Pattern 3) as an explicit, tested task in the plan — not an assumed side effect of "add mtime and write the diff."
**Warning signs:** A test that intentionally makes one subdirectory unreadable (root still reachable) during a re-scan and expects an `unreadable` diff entry — if that test can't be written without first touching `traverseDirectory`'s skip branches, the gap is confirmed still open.

### Pitfall 2: Re-scan's Failure Silently Enabling "Write Partial Catalog"
**What goes wrong:** A re-scan that hits a genuine scan-root loss populates `a.lastPartial`/`a.lastScanReq` (because the binding reused `startScan`'s preamble), and — even though the re-scan UI never renders a partial-write button — the pre-existing `WritePartialCatalog` binding is still reachable and would write re-scan's incomplete tree over the wrong output path if ever invoked.
**Why it happens:** `startScan`'s partial-retention logic is unconditional on any `SourceUnavailableError`, with no flow-kind discriminant.
**How to avoid:** Re-scan's binding must call `a.catalogService.Walk` directly (not `a.startScan`), and must not run the `a.lastPartial = ...` assignment at all on failure — see Pattern 1's risk callout for the exact code path.
**Warning signs:** A test that starts a re-scan, forces a source loss, then calls `WritePartialCatalog` and asserts it returns `"no partial scan retained to write"` (the existing error message at `app.go:473`) — if it instead succeeds and writes something, the leak is real.

### Pitfall 3: Treating an Absent `ModTime` as Epoch
**What goes wrong:** Every pre-existing catalog (the overwhelming majority — this field is brand new) has `ModTime` absent (`omitempty`, zero value). If the diff naively compares `old.ModTime != new.ModTime` without checking for "old is zero," every unchanged file in every existing catalog's first re-scan gets flagged `changed`.
**Why it happens:** `omitempty` on a numeric field is indistinguishable from a genuine zero unless the comparison logic explicitly branches on it.
**How to avoid:** The diff's per-entry comparison must be: `changed := old.Size != new.Size || (old.ModTime != 0 && old.ModTime != new.ModTime)` — size-only fallback when the old entry predates this field.
**Warning signs:** Re-scanning a catalog written before this phase and seeing every file marked `changed` with identical sizes.

### Pitfall 4: Claiming COMPAT-06 From Reading the YAML
**What goes wrong:** The plan marks COMPAT-06 complete because `build.yml`'s three jobs look correct and the milestone's new dependencies (`wastebasket/v2`, `fsnotify`, go directive 1.23.4) were already confirmed to cross-build cleanly in earlier phases' WINDOWS.md entries.
**Why it happens:** Static review of a workflow file is much cheaper than actually pushing and watching it run, and everything about it *looks* fine.
**How to avoid:** `STATE.md` (read this session) documents that local `main` is deliberately kept far ahead of `origin/main` (`workflow.use_worktrees: false`, ~100 commits ahead) for the entire milestone — meaning `build.yml` has genuinely never run against phases 22-27's actual code, let alone phase 28's. COMPAT-06 requires an actual push (see Architecture Pattern 4 below) with a real green Actions run as evidence, not a code-reading exercise.
**Warning signs:** No GitHub Actions run URL/timestamp cited as evidence in the phase's completion summary.

## Code Examples

### Diff comparison logic (per-entry), incorporating the omitempty-mtime and unreadable-marker fixes above
```go
// Pattern only — internal/catalog/diff.go (new file, no existing source to quote from)
func categorize(old, new *models.CatalogItem) DiffState {
	switch {
	case old == nil:
		return StateAdded
	case new == nil:
		return StateRemoved
	case new.Unreadable:
		return StateUnreadable // requires MarkUnreadableOnSkip on the re-scan's Walk call (Pattern 3)
	case old.Type != new.Type:
		return StateRemoved // paired with a synthetic StateAdded row for the new type — see
		                       // Pattern 3's "type change edge case" for the two-row rationale
	case old.Size != new.Size:
		return StateChanged
	case old.ModTime != 0 && old.ModTime != new.ModTime:
		return StateChanged // old.ModTime == 0 means "unknown, pre-dates this field" -- size-only
	default:
		return StateUnchanged
	}
}
```

### Existing crash-safe write reuse (verbatim signature, confirmed this session)
```go
// Source: internal/catalog/service.go:211 (existing, unmodified — Overwrite and Keep-both both
// call this directly with the newly-walked tree)
func (s *Service) WriteCatalogFrom(tree *models.CatalogItem, title, outputDir, outputRoot, copyToDirectory string, opts Options) (*models.CreateCatalogResult, error)
```

## State of the Art

Not applicable in the usual "library X replaced library Y" sense — this phase's domain (a size+mtime diff over a filesystem-derived tree, atomic-write reuse) has no external ecosystem to be current or stale against. The one relevant "state of the art" fact: this app's own architecture already evolved through this exact class of problem three times (Phase 25's crash-safe write, Phase 27's trash/duplicate/rename primitives) and consistently chose reuse-through-extraction over new implementations — this phase's approach (extract `Walk`, add one `Options` flag, reuse everything else) is consistent with that established trajectory, not a departure from it.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Unix-seconds `int64` is the right wire representation for `ModTime` (vs. RFC3339 string) | Architecture Pattern 2 | Low — this is explicitly Claude's Discretion per CONTEXT; if the planner or a later reviewer prefers RFC3339 for human-readability in raw JSON, that's a valid alternative choice within the same discretion grant, not a correctness issue |
| A2 | A directory's own `ModTime` (its own immediate add/remove event) is what CONTEXT means by "a directory is changed only when its own entry changes" | Architecture Pattern 2 | Medium — CONTEXT doesn't spell out what property of a directory is compared; if the intended meaning is instead "never mark a directory changed at all, only files," the diff's directory-changed category would need to be removed rather than driven by mtime. Flagged for discuss-phase or planner confirmation before implementation |
| A3 | Treating an old-vs-new `Type` mismatch (file↔directory at the same path) as a removed+added pair (two rows) rather than a single `changed` row | Architecture Pattern 3 | Low — genuinely rare edge case (re-purposed media); the two-row treatment is a reasonable default but not explicitly specified anywhere in CONTEXT or the UI-SPEC, which assume type stays constant per path |
| A4 | `MarkUnreadableOnSkip` is the right mechanism (a new `Options` field gating existing skip-and-continue branches) rather than some other design for making the walk's skip path leave a marker | Architecture Pattern 3 | Medium-High — this is the single largest new-code item this research surfaced beyond what CONTEXT explicitly scoped; the *fact* that the current marker can't fire on a completed scan is verified by reading the code, but the *fix's shape* is this document's own proposal, not a locked decision. The planner should treat this as needing explicit sign-off (a plan-time decision point), not silent adoption |

## Open Questions

1. **Does a `MarkUnreadableOnSkip`-flagged re-scan also need to distinguish "read error" (a `readErrors` counter tick, already reported live via `scan:progress`) from the *new* per-node marker for the diff — or is one always accompanied by the other?**
   - What we know: `recordReadError` (`service.go:80-86`) already increments `st.readErrors` and appends a bounded `ReadErrorEntry` on every skip-and-continue failure today, regardless of the new flag.
   - What's unclear: whether the diff-row's "unreadable" reason text (UI-SPEC: "the short read-error reason (e.g. `permission denied`)") should read from the per-node `ReadError` field (new, only set when the flag fires) or needs its own plumbing.
   - Recommendation: use the per-node `ReadError` field directly — it's already the exact string (`err.Error()`) the flag would set, no new plumbing needed. Confirm during planning that the flag is set unconditionally whenever `recordReadError` fires under `MarkUnreadableOnSkip`, not conditionally.

2. **Does the similarity-warning heuristic need the OLD tree's total entry count, or the diff's combined total?**
   - What we know: `28-UI-SPEC.md`'s Similarity Warning Contract defines `total` as "the same denominator the invariant above uses (all five categories summed)" — i.e. the diff's own combined total, not a separate old-tree-only count.
   - What's unclear: nothing substantively — this is already resolved by the UI-SPEC, restated here only because Priority 3's brief specifically asked about the similarity heuristic's interaction with scale.
   - Recommendation: no further research needed; implement per the UI-SPEC's stated formula.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Building/testing this phase's Go changes | ✓ | go1.26.6 darwin/arm64 (satisfies `go.mod`'s `go 1.23.4` directive) | — |
| GitHub Actions (macOS runner) | COMPAT-06's `build.yml`/`release.yml` macOS legs | ✓ (via GitHub-hosted `macos-latest`/`macos-14`) | — | — |
| GitHub Actions (Windows runner) | COMPAT-06's Windows leg | ✓ (via GitHub-hosted `windows-latest`/`windows-2022`) | — | Native Windows compile via CI is itself the fallback for "no local Windows machine" — this project has never had one, per every prior phase's WINDOWS.md entries |
| GitHub Actions (Linux runner) | COMPAT-06's Linux leg | ✓ (via GitHub-hosted `ubuntu-22.04`) | — | Same as Windows — CI is the only place this project's Linux legs have ever been exercised |
| A local Windows or Linux machine/VM | Runtime verification of WINDOWS.md's hardware-dependent entries (#1, #2, #4, #5, #8, #9, #10) | ✗ | — | None — every prior phase (23, 24, 25, 27) recorded the identical "no Windows/Linux machine available" gap; this phase cannot close those entries any more than its predecessors could. See the WINDOWS.md Ledger Sweep section below |
| Push access to `origin` (the real GitHub remote, not just local git) | Actually triggering a real `build.yml` run for COMPAT-06's evidence | Unconfirmed — depends on the executing session's git remote credentials, not a phase-28-specific tool | — | If push access is unavailable in the execution environment, COMPAT-06 cannot be honestly claimed this phase and must be recorded as blocked, per Pitfall 4 above |

**Missing dependencies with no fallback:**
- A local Windows/Linux machine — unchanged from every prior phase; the plan should not attempt to newly resolve this, only correctly acknowledge which WINDOWS.md entries remain open because of it (see below).

**Missing dependencies with fallback:**
- Windows/Linux compile verification — GitHub's hosted native runners in `build.yml`/`release.yml` already substitute for local hardware at the compile level (not the runtime level).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing`, table-driven, files beside source (`internal/catalog/*_test.go` — 6 existing files confirmed this session: `service_test.go`, `duplicate_test.go`, `atomicwrite_test.go`, `atomicwrite_sigkill_test.go`, `measure_test.go`, `rename_test.go`). No frontend test framework by project-wide design (TEST-01 deferred); frontend proof is `tsc --noEmit`/`vite build` + live dev-browser, unchanged from every prior phase |
| Config file | none — plain `go test ./...` |
| Quick run command | `go test ./internal/catalog/... ./pkg/models/...` |
| Full suite command | `go test ./...` (repo-wide, confirmed this session no test config file exists to scope it further) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ACT-06 | Diff correctly categorizes added/removed/changed/unchanged for a synthetic old/new tree pair | unit | `go test ./internal/catalog/... -run TestDiff -v` | ❌ Wave 0 — `diff_test.go` does not exist yet |
| ACT-06 | A skip-and-continue subtree failure under `MarkUnreadableOnSkip` produces an `unreadable` diff entry, not a `removed` one | unit | `go test ./internal/catalog/... -run TestTraverseDirectory_MarksSkippedNodeWhenFlagSet -v` | ❌ Wave 0 |
| ACT-06 | `TestTraverseDirectory_SingleEntryErrorSkipsAndContinues` (existing, `service_test.go:462`) still passes unmodified after the flag is added | regression | `go test ./internal/catalog/... -run TestTraverseDirectory_SingleEntryErrorSkipsAndContinues -v` | ✓ already exists |
| ACT-06 | mtime-based `changed` detection for a same-size edit | unit | `go test ./internal/catalog/... -run TestDiff_SameSizeMtimeChange -v` | ❌ Wave 0 |
| ACT-06 | An old entry with `ModTime == 0` (pre-existing catalog) falls back to size-only comparison, never flags spurious `changed` | unit | `go test ./internal/catalog/... -run TestDiff_MissingOldModTimeFallsBackToSizeOnly -v` | ❌ Wave 0 |
| ACT-07 | Overwrite reuses `WriteCatalogFrom` and produces the crash-safe write CRT-safety already covers | regression | `go test ./internal/catalog/... -run TestWriteCatalogFrom -v` | ✓ existing coverage via `service_test.go`'s write tests, extend if a re-scan-specific call-site test is warranted |
| ACT-07 | Keep-both resolves via `nextCopyRoot`'s existing collision loop | regression | `go test ./internal/catalog/... -run TestDuplicateCatalog -v` (loop itself already covered; a new test should assert the *re-scan caller* invokes the same function, not a copy) | ❌ Wave 0 for the new call-site test |
| STATE-03 | Re-scan's own failure never populates `a.lastPartial`/`a.lastScanReq` | unit | `go test . -run TestRescan_DoesNotRetainPartialForWritePartialCatalog -v` (new, at the `app` package level) | ❌ Wave 0 |
| COMPAT-06 | `build.yml` compiles cleanly on real macOS/Windows/Linux CI runners against this milestone's full diff | manual (CI-observed) | A real `git push` + observed green run — see Architecture Pattern 4, no local command substitutes for this | ❌ Wave 0 — requires the push itself, not a test file |

### Sampling Rate
- **Per task commit:** `go test ./internal/catalog/... ./pkg/models/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd-verify-work`, plus a real observed green `build.yml` Actions run for COMPAT-06 (no test command substitutes for this — see Pitfall 4)

### Wave 0 Gaps
- [ ] `internal/catalog/diff_test.go` — covers ACT-06's five-state categorization, the mtime-fallback edge case, and the type-change edge case (Assumption A3)
- [ ] A new test in `internal/catalog/service_test.go` (or a new file) covering `MarkUnreadableOnSkip`'s behavior, alongside the existing skip-and-continue regression test
- [ ] A new test at the `app` package level (`app_test.go`, already exists per Phase 25's `TestStartScan_RetainsPartialOnSourceLoss` pattern) proving re-scan's failure path does NOT populate `a.lastPartial`/`a.lastScanReq` (Pitfall 2)
- [ ] Framework install: none — Go's stdlib `testing` is already fully wired

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Single-user desktop app, no auth surface anywhere in this codebase |
| V3 Session Management | no | Not applicable — no sessions |
| V4 Access Control | yes | Every path this phase's new bindings accept must be containment-checked exactly as every prior phase's bindings are — `osutil.ContainsPath` (already used by `StartScan`, `DeleteCatalog`, `DuplicateCatalog`, `RenameCatalog`) is the standard control, reused, not re-implemented |
| V5 Input Validation | yes | The re-scan binding accepts a user-picked `sourcePath` (from `VolumePicker`, either a detected volume's `mountPath` or an arbitrary chosen folder) and the existing catalog's own `jsonPath` — both need the same `filepath.Abs`/`filepath.EvalSymlinks`/`ContainsPath` treatment `startScan` already applies to its `outputDir`/`copyToDirectory` params (`app.go:262-315`) |
| V6 Cryptography | no | No cryptographic operation in this phase |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via a crafted `outputRoot`/target filename escaping the catalog directory | Tampering | `osutil.ContainsPath`, already the pattern every write-capable binding in this app uses (`StartScan`, `DuplicateCatalog`, `RenameCatalog`, `DeleteCatalog`) — the Keep-both write target (derived from `nextCopyRoot`, itself derived from the *existing* catalog's own already-validated path, never free user text) inherits this by construction, not by a new check |
| Overwrite of an unrelated file if the "original catalog path" used for ACT-07's Overwrite resolution is not re-validated at write time | Tampering | The write target for Overwrite must be re-derived from the catalog's own on-disk path (already inside the configured catalog directory, by definition of how the catalog got into the rail) rather than trusting a renderer-supplied path uncritically — same discipline `DeleteCatalog`'s `filepath.Abs` → `filepath.EvalSymlinks` → `ContainsPath` sequence already establishes (`app.go:998-1012`) |
| A re-scan's source volume path used to probe/read arbitrary filesystem locations outside any legitimate volume | Information Disclosure | `VolumePicker`'s two source kinds (`volume` from `internal/volumes.List()`, or `folder` from the native OS picker `wailsAPI.selectDirectory()`) are both already the exact same trusted-provenance sources Create's own volume selection uses today — no new trust boundary is introduced by reusing the same component verbatim |

## Architecture Pattern 4 — COMPAT-06 and the CI Pipeline

**What a real pushed `build.yml` run actually proves:**
- `build.yml` (confirmed read this session, `.github/workflows/build.yml`) has exactly three jobs: `build-macos` (`macos-latest`, `wails build -platform darwin/universal`), `build-windows` (`windows-latest`, `wails build -platform windows/amd64 -windowsconsole`), `build-linux` (`ubuntu-22.04`, `wails build -platform linux/amd64`, after installing `libgtk-3-dev libwebkit2gtk-4.0-dev`). This exactly matches CONTEXT's stated matrix ("macOS universal, Windows amd64, Linux amd64") — verified, not assumed.
- Each job runs on a **native, GitHub-hosted runner for that OS** (not a cross-compile from a single host) — `build-windows` genuinely compiles on Windows, `build-linux` genuinely compiles on Linux. This is a stronger compile-level proof than a cross-compiled `GOOS=windows go build` from a macOS dev machine would be, because it exercises the real native toolchain, real cgo linkage for `webview2`/`webkit2gtk`, and real platform-specific build-tagged files (`volumes_windows.go`, `volumes_linux.go`, the Windows/Linux branches inside `wastebasket` and `fsnotify`).
- Frontend TypeScript type errors **are** caught: `wails.json`'s `frontend:build` runs `npm run build`, and `frontend/package.json`'s `build` script (confirmed this session) is `"tsc && vite build"` — `tsc` runs (and fails the whole job on a nonzero exit) *before* `vite build`, so a TS type error anywhere in the frontend rewrite genuinely fails `build.yml`, not just a local `tsc --noEmit` check the CI happens to skip.
- **What it does NOT prove:** no signing, no notarization, no execution of the built binary. `build.yml` uploads the built `.app`/`.exe`/binary as an artifact and stops — nothing in it launches the app, clicks a button, or exercises any runtime code path (this is why it cannot close any of the WINDOWS.md runtime-behavior entries, only reinforce their compile-level confidence).

**Signing/notarization (`release.yml`) — confirmed, matches CONTEXT exactly:**
- `release.yml` triggers only on a pushed tag matching `v*.*.*` or manual `workflow_dispatch`. Its macOS job imports an Apple certificate, codesigns, notarizes via `xcrun notarytool`, and staples — a real, non-optional pipeline (no skip branch). Its Windows job (confirmed this session, `release.yml:159-196`) has a genuine skip-with-warning branch: `steps.signing.outputs.available` is set `false` when the `ES_USERNAME` secret is absent, and both `esigner-codesign` steps are gated `if: steps.signing.outputs.available == 'true'` — meaning an unsigned Windows build would still complete the job and produce artifacts, just without the "Sign portable exe"/"Sign NSIS installer" steps ever running. `STATE.md`'s Deferred Items table (read this session) confirms CRED-04/CRED-05 (`ES_USERNAME`/`ES_PASSWORD`/`CREDENTIAL_ID`/`ES_TOTP_SECRET`) remain unprovisioned from the v2.3.0 close — so **today, a real tagged release run would produce a signed/notarized macOS artifact and an *unsigned* Windows artifact**, exactly matching CONTEXT's framing that the signing/notarization half must be recorded as an open ledger item unless a tag is actually cut.

**`release-please.yml` — what a push to `main` alone does and does not trigger:**
- Triggers on every push to `main`, running `googleapis/release-please-action`. This action's standard behavior is to open or update a "release PR" (a PR that bumps `CHANGELOG.md`/version metadata based on Conventional Commits since the last release) — it does **not** itself create a git tag or a GitHub Release. A tag (and therefore `release.yml`'s trigger) is only created when that release PR is later **merged** (or a tag is cut manually via `workflow_dispatch`). So: pushing this phase's commits to `main` will, by itself, only update/open a release-please PR — it will not fire `release.yml` and will not itself prove signing/notarization. This is consistent with, and reinforces, CONTEXT's own framing.

**The concrete gap that must be closed to claim COMPAT-06 honestly:** `STATE.md` (read this session, "Worktree isolation: intentionally OFF" note) documents that local `main` runs roughly 100 commits ahead of `origin/main` **by deliberate design** for this entire milestone (Phases 22-27's real commits have never been pushed). This means `build.yml` — despite existing, being correctly configured, and matching CONTEXT's stated matrix — **has not actually run against any of this milestone's code**. A `pull_request: branches: [main]` trigger also exists in `build.yml`, so a real green run could be obtained either by (a) pushing a branch and opening a PR against `main` without merging (gets a real Actions run, defers the "should local main go to origin" question), or (b) a direct push of local `main` to `origin/main` once the milestone is otherwise ready to conclude. **This is a git operation with real consequences (a large multi-phase push, or opening a PR) that the plan must call out explicitly as its own task, with the user's awareness that it is a genuine remote push — not a step to perform silently as part of "verify COMPAT-06."** Given this project's git safety norms (no force-push, no `--no-verify`, confirm before large/irreversible operations), this should be a clearly-labeled, confirmable task in the plan, not folded silently into a verification checklist.

**Risk to the matrix legs from this milestone's actual changes, assessed against what's already known:**
- go directive `1.23.4` (bumped Phase 27, required by `wastebasket/v2`'s own `go.mod`) — `actions/setup-go@v5.4.0` with `go-version-file: go.mod` resolves this correctly; low risk, this pattern is already exercised by every prior `setup-go` invocation in these same workflow files.
- `wastebasket/v2`, `fsnotify` — both already confirmed to cross-build cleanly for Windows/Linux targets per WINDOWS.md entries #4, #5, #8, #9, #10 (each explicitly states "compile... verified" as an already-known fact from earlier phases' sessions) — low risk of a new compile failure; the *runtime* behavior these entries actually track is untouched by a `build.yml` run either way.
- Frontend rewrite (custom CSS, no more Ant Design tab bar, all of Phases 22-27's UI work) — `tsc && vite build` catches type errors as established above; the main residual risk is a dependency resolution failure (`npm install` under `frontend:install`) if `package-lock.json`/`package.json` drifted, which is a generic CI risk unrelated to this phase specifically.
- `wails.json`'s `productVersion: "3.0.0"` (bumped Phase 26, confirmed present this session) — purely metadata, embedded at build time via Wails' own tooling, no build-breaking risk.

## Architecture Pattern 5 — The WINDOWS.md Ledger Sweep

Full ledger read this session (`.planning/WINDOWS.md`, 14 total entries: 11 open, 3 fixed, 0 waived at scout time). Disposition guidance per entry, distinguishing what a green `build.yml` run genuinely closes from what still needs real hardware:

| # | Phase | Description (abbreviated) | What would close it | Closable by CI push? | Closable this phase? |
|---|-------|---------------------------|---------------------|----------------------|----------------------|
| 1 | 23 | Windows `explorer /select,` argv shape, runtime-unverified | A real Windows machine clicking "Reveal JSON in Explorer" and observing the correct file gets selected | No — `build.yml` never executes the binary | No — stays open, no Windows hardware available this session either |
| 2 | 24 | Ctrl+K non-macOS, runtime-unverified | A real Windows/Linux webview receiving a real Ctrl+K keypress | No | No — stays open |
| 4 | 25 | Windows disk-space/drive-letter enumeration, runtime-unverified | A real Windows machine with real drives, checking byte totals and drive-letter set are correct | Partially — a green `build-windows` job on the real `windows-latest` native runner *strengthens* the "compiles correctly on genuine Windows" claim beyond the ledger's current "cross-build verified" framing (which implies a cross-compile, not a native build), but the entry's core concern (runtime correctness) is untouched | No — the compile-level claim can be reinforced with the CI evidence and the entry's wording updated to reflect a native-runner compile rather than a cross-compile, but it should not be marked `fixed` |
| 5 | 25 | Linux volume-enumeration heuristic, runtime-unverified | Same as #4, for Linux `/mnt`/`/media` enumeration against `/proc/mounts` | Same as #4 — native `ubuntu-22.04` compile reinforces, doesn't close | No |
| 7 | 26 | RailSide persistence full OS quit-and-relaunch, not performed live (deferred only to avoid disrupting a *shared* wails dev session other Phase 26 plans depended on) | A single real app quit (`window.runtime.Quit()`) + relaunch, confirming `config.json`'s `RailSide` value survives with no flash | No — this is a live-app runtime check, CI doesn't execute the binary | **Yes, genuinely closable this phase** — the original reason it was deferred (a shared, actively-used `wails dev` session other plans depended on) no longer applies now that Phase 26 is long complete; RailSide persistence is OS-agnostic config I/O (same code path on every platform), so a real quit+relaunch during this phase's own live dev-browser verification pass (which it will already be running for its own UI checks) closes this cheaply. Recommended as an explicit bonus disposition task |
| 8 | 27 | fsnotify Windows rename-release divergence (documented upstream behavior, not a bug) | A real Windows machine renaming the watched directory while StorCat watches it, observing what StorCat actually does (silent staleness vs. visible error vs. crash) | No | No — but the plan could re-read `internal/watch/watcher.go`'s error-handling path to *reason* about the likely resulting behavior without hardware, and consider whether this entry should be reframed as an accepted/waived upstream limitation (like #12) rather than left open indefinitely awaiting hardware that may never arrive — a disposition judgment call, not something this research can resolve unilaterally |
| 9 | 27 | fsnotify Windows/Linux backends, runtime-unverified | Real hardware exercising `ReadDirectoryChangesW`/`inotify` | No | No — stays open |
| 10 | 27 | wastebasket Windows/Linux backends, runtime-unverified | Real hardware exercising a real Trash move | No | No — stays open. **Note:** this phase's STATE-03 "Remove from library" button adds a new reachable call site into this exact same unverified code path (it reuses `DeleteCatalog` verbatim) — this doesn't change the entry's verification status, but the sweep should note the new call site exists |
| 11 | 27 | `WriteFileAtomic` parent-directory fsync unsupported on Windows, once-per-process logged | Real Windows machine confirming the once-per-run log fires correctly and no data-loss results | No | No — stays open. **Note:** this phase's ACT-07 Overwrite path is a new caller of `WriteFileAtomic` (via `WriteCatalogFrom`) — same "new call site, same open status" note as #10 |
| 12 | 27 | wastebasket macOS AppleScript interpolation, accepted upstream-owned residual risk | Not really closable by "testing" — it's a documented, accepted, non-StorCat-owned limitation bounded by the existing containment gate | N/A | No — stays open as an accepted risk, unaffected by this phase (STATE-03's remove-from-library reuses the same already-gated `DeleteCatalog` path, no new attack surface) |
| 14 | 27 | SIGKILL harness coupled to `WriteFileAtomic` by convention not by call, maintenance-only residual | A future reviewer re-diffing the harness against `atomicwrite.go` if its write sequence is ever reordered | N/A | No — unaffected; this phase does not modify `atomicwrite.go`'s internal sequence (ACT-07 reuses it "wholesale," per CONTEXT) |

**Already fixed, no action needed:** #3 (CRT-13, fixed Phase 25), #6 (SIGKILL crash-safety, fixed Phase 27 via #27-02's real subprocess-kill harness), #13 (Menu focus-restore, fixed Phase 27-REVIEW-FIX).

**Net disposition guidance for the plan:** of 11 open entries, this phase can genuinely close **one** (#7, via a cheap live quit+relaunch during its own verification pass), can *reinforce but not close* two more (#4, #5, via the real native-runner CI evidence this phase's own COMPAT-06 push produces), should consider a *waive-vs-leave-open* judgment call on one (#8, a documented upstream limitation rather than an unverified StorCat behavior), and must leave the remaining seven genuinely open pending real Windows/Linux hardware this project has never had access to in any phase — consistent with, not a regression from, every prior phase's experience. The plan should not attempt to close #1, #2, #9, #10, #11 this phase; doing so without real hardware would be recording an unverified claim as verified, exactly the failure mode CLAUDE.md's Evidence Standards principle warns against.

## Sources

### Primary (HIGH confidence — read directly this session)
- `internal/catalog/service.go` (full file) — `CreateCatalogWithContext`, `traverseDirectory`, `WriteCatalogFrom`, `walkState`, `classify()`
- `internal/catalog/errors.go` (full file) — `SourceUnavailableError`, `PartialScan`
- `internal/catalog/duplicate.go` (full file) — `DuplicateCatalog`, `nextCopyRoot`, `isCandidateRootFree`
- `pkg/models/catalog.go` (full file) — `CatalogItem`, existing `omitempty` field precedent (`Title`, `Unreadable`, `ReadError`)
- `app.go` (relevant sections: 180-441, 980-1027) — `throttledProgress`, `resolveScanTotal`, `StartScan`/`startScan`, `CancelScan`/`cancelActiveScan`, `WritePartialCatalog`, `DeleteCatalog`
- `internal/search/service.go` (relevant sections) — `LoadCatalog`, `BrowseCatalogs`, `djherbis/times` usage confirmed unrelated to per-entry mtime
- `internal/search/flatten.go` (full file) — `LoadCatalogFlat`, the existing flatten-pattern precedent
- `go.mod` (full file) — confirmed no new dependencies needed, `djherbis/times v1.6.0` already present for an unrelated purpose
- `.github/workflows/build.yml`, `release.yml`, `release-please.yml`, `distribute.yml` (full files) — the exact CI matrix, signing/notarization gates, release-please behavior
- `frontend/package.json` — confirmed `"build": "tsc && vite build"` (type-checking IS enforced in CI)
- `wails.json` — confirmed `productVersion: "3.0.0"`
- `frontend/src/components/workspace/UnreadableCatalogPanel.tsx` (full file) — the STATE-03 stub location
- `frontend/src/components/workspace/DetailsPanel.tsx` (relevant sections) — the Footer stub, `CatalogActions` menu insertion point
- `frontend/src/components/workspace/create/VolumePicker.tsx` (full file) — confirmed stateless, reusable verbatim
- `frontend/src/components/workspace/create/ScanningBody.tsx`, `ErrorBody.tsx` (relevant sections) — confirmed exact prop signatures for the additive-optional-props extension
- `frontend/src/contexts/AppContext.tsx` (grep + targeted reads) — confirmed `ScanState`/`scan` reducer slice shape
- `frontend/src/types/scan.ts` (full file) — `ScanState`, `ScanFailure`, `ScanProgress`, `classifyScanFailure`
- `.planning/phases/28-re-scan-diff/28-CONTEXT.md`, `28-UI-SPEC.md` — the phase's own locked decisions and approved design contract
- `.planning/REQUIREMENTS.md`, `.planning/STATE.md`, `.planning/WINDOWS.md` — full reads
- `design_handoff_storcat_ui/README.md`, `design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html` (targeted sections: §5, §6, "Mocked functionality" table, the literal `isRescan` dialog markup and `menuItems`/`diff` array) — the original design intent 28-UI-SPEC.md resolved against
- `.planning/config.json` — confirmed `nyquist_validation: true`, no `security_enforcement` override (treated as enabled)
- Local `go version` (`go1.26.6 darwin/arm64`) and `go vet ./internal/catalog/...` (clean) — confirmed this session

### Secondary (MEDIUM confidence)
None — every claim in this document traces to a file read this session; no web search or external documentation lookup was needed for a phase with zero new dependencies.

### Tertiary (LOW confidence)
None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies, entirely verified against `go.mod`/`frontend/package.json` read this session
- Architecture (walk/write split, diff algorithm): HIGH for the mechanical split itself (verbatim source read and traced); MEDIUM-HIGH for the `MarkUnreadableOnSkip` fix's exact shape (the *gap* is HIGH-confidence verified-by-reading, the *proposed fix* is this document's own reasoned proposal, not a locked decision — see Assumption A4)
- CI/COMPAT-06 findings: HIGH — every workflow file read in full this session, cross-referenced against `STATE.md`'s explicit unpushed-`main` note
- WINDOWS.md sweep: HIGH — full ledger read this session, disposition guidance reasoned directly from each entry's own stated evidence gap
- Pitfalls: HIGH — each pitfall traces to a specific, quoted line range in code read this session, not general domain knowledge

**Research date:** 2026-08-16
**Valid until:** Stable — this is a closed-world, zero-external-dependency phase; the only thing that could invalidate this research before execution is a change to `internal/catalog/service.go`, `app.go`, or the `.github/workflows/*.yml` files between now and planning. Recommend re-verifying line numbers (not conclusions) if more than a few days elapse before this phase is planned.
