# Phase 28: Re-scan & Diff - Context

**Gathered:** 2026-08-16
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous) — 4 grey areas, 16 questions, all accepted as recommended

<domain>
## Phase Boundary

Users re-scan a catalog's source volume, see what changed as a diff, and reconcile it without risking the
existing catalog. Plus the unreadable-catalog action trio, and the milestone's cross-platform build proof.

In scope: ACT-06 (re-scan + diff of added/removed/changed/unchanged with counts), ACT-08 (always ask which
source volume — never guess), ACT-07 (resolve by overwrite / keep both / discard, overwrite reusing the
Phase 25+27 crash-safe atomic write), STATE-03 (an unreadable catalog can be re-scanned, its `.html` opened,
or removed from the library), COMPAT-06 (the app builds, signs, notarizes and releases on every existing CI
platform target with the full milestone's changes included).

This is the final phase of the v3.0.0 milestone. The ROADMAP flags it as "the handoff's own biggest backend
piece" and required a dedicated research pass before planning, specifically to resolve the volume-relocation
policy and the on-disk unreadable-subtree marker format from Phase 25. Both are resolved below.

</domain>

<decisions>
## Implementation Decisions

### Diff semantics
- **The scanner gains mtime capture, and "changed" means size OR mtime differs.** The scout confirmed
  `CatalogItem` carries only `Size` today — `info.ModTime()` is never read (`internal/catalog/service.go:287-294`).
  Size-only comparison misses the same-size edit, which is precisely the case a re-scan exists to catch. The new
  field is `omitempty`, so catalogs without it stay byte-identical and every existing reader keeps working.
- **Node identity is the existing `Name` field.** It already holds the full relative display path
  (`"./sub/file.txt"`), set once from `filepath.Rel` at traversal time. It is the natural diff key and needs no
  model change.
- **An unreadable subtree is a FOURTH diff state, distinct from `removed`.** This resolves the ROADMAP's
  research flag. A permission error or an unreadable disc region is not the same event as "the user deleted
  this", and collapsing the two would let an overwrite silently propose destroying catalog data that is merely
  unreadable at this moment. Phase 25's flat `Unreadable bool` / `ReadError string` scalars stay as they are —
  the format is confirmed final; what this phase adds is a diff-level state, not an on-disk format change.
- **Directories are diffed, but a directory is `changed` only when its own entry changes** — never because a
  descendant changed. Otherwise every ancestor of one edited file lights up and the counts stop meaning anything.

### Re-scan flow & volume selection
- **Split the existing walk from the write, and export the walk half.** There is currently no way to build a
  tree without writing a catalog file — `CreateCatalogWithContext` always ends in `WriteCatalogFrom`
  (`service.go:133,202`). But it already calls `traverseDirectory` then `WriteCatalogFrom` as two cleanly
  separable steps, so this is a split, not a rewrite. Do NOT write a second traversal implementation; two
  walkers will drift on hidden-file handling, symlinks, and error classification.
- **Nothing about the chosen volume is persisted.** ACT-08 says always ask. The scout confirmed no volume
  name, UUID, or serial is recorded anywhere in the catalog JSON today, so there is genuinely nothing to guess
  from — and persisting one would invite a future "remember this volume" shortcut that silently re-scans the
  wrong disc after a media swap.
- **A wrong-disc pick is shown honestly, never blocked.** It will read as near-total add+remove. Surface a
  plain warning above the diff when similarity is very low, but never refuse to proceed — the app cannot know
  the user is wrong, and a legitimately enormous change must stay possible.
- **Progress and cancellation reuse `scan:progress` and `CancelScan` unchanged.** One scan at a time; the
  existing `scanMu` / `activeScanCancel` guard (`app.go:317-343,404`) already covers a re-scan.

### Diff resolution (ACT-07)
- **Overwrite reuses `WriteCatalogFrom` wholesale.** It is the scan's own write path and already routes
  through the Phase 27-hardened `WriteFileAtomic` (`tmp.Sync()` → `os.Rename` → best-effort `syncDir`).
  ACT-07 asks for exactly that guarantee; re-deriving per-file writes would create a second path that can drift
  from the primitive it is supposed to reuse.
- **"Keep both" produces a new catalog beside the original using Phase 27's `-copy` suffix loop**
  (`-copy`, `-copy-2`, …). The collision handling exists and is tested, and the user already learned that
  naming convention from Duplicate.
- **The `.html` sidecar is rewritten when the original had one, and never created where none existed.** The
  pair must stay consistent, and silently gaining an `.html` the user deliberately opted out of is a surprise.
- **Discard writes nothing and leaves the original byte-identical, with no confirmation prompt.** Nothing was
  committed, so there is nothing to lose.

### STATE-03 actions & COMPAT-06
- **"Remove from library" reuses Phase 27's `DeleteCatalog` delete-to-Trash, including its confirmation.**
  There is no library-membership concept to toggle — membership IS the file living in the configured catalog
  directory (`app.go:993-1026`). Recoverable via the OS Trash, consistent with every other removal in the app.
  Do not invent a hidden/excluded list; that would be a second membership concept.
- **Re-scanning an unreadable catalog is a fresh scan, not a diff.** Its JSON does not parse, so there is no
  old tree to compare against and a diff would be undefined. Pick the volume, scan, then offer overwrite or
  keep-both with no diff view.
- **COMPAT-06 is proven by a real pushed CI run of `build.yml`** across its existing matrix (macOS universal,
  Windows amd64, Linux amd64). **Signing and notarization are NOT claimed by that run** — they live in
  `release.yml` behind a tag, and Windows signing is currently a skip-with-warning branch because CRED-04/CRED-05
  (`ES_USERNAME`/`ES_PASSWORD`/`CREDENTIAL_ID`/`ES_TOTP_SECRET`) are not provisioned (`.planning/STATE.md:240`).
  If no release tag is cut in this phase, the signing/notarization half is recorded as an open ledger item
  rather than claimed. A green `build.yml` does not prove Windows signing and must not be reported as if it does.
- **The open `.planning/WINDOWS.md` entries are audited and disposed in this phase** (11 open at scout time).
  Close what CI now genuinely proves, keep what truly needs Windows/Linux hardware, and leave the milestone with
  an honest ledger. COMPAT-06 sits directly upstream of that backlog.

### Post-research resolutions (2026-08-16, from 28-RESEARCH.md)

- **`traverseDirectory` gains an opt-in `MarkUnreadableOnSkip` option, and the re-scan walk sets it.** This is
  load-bearing, not a refinement. Research found the locked "unreadable is a fourth state distinct from
  removed" decision is currently **unimplementable**: the `Unreadable`/`ReadError` marker can only be set when
  the scan *root* itself becomes unreachable, which aborts the entire walk and routes to the error step -- it
  never reaches the diff. A single unreadable subdirectory with the root still readable is silently DROPPED by
  the existing skip-and-continue branches (`service.go:318-320`, `:392-394`) with no marker at all. Diffed, that
  is indistinguishable from "removed" -- so an overwrite would destroy catalog data that is merely unreadable
  right now, which is the exact outcome the fourth state was locked to prevent. The option makes those existing
  branches MARK the node instead of dropping it, without aborting the walk. **Opt-in, so the create flow's
  behavior is unchanged.**
- **Re-scan calls the new exported walk directly and must never route through `startScan`.** Research flagged a
  real trap: Create's `a.lastPartial` / `a.lastScanReq` state backs its CRT-11 partial-write retention. Re-scan
  must not populate that shared state at all -- its error step offers no partial-write option by design, and
  leaking into it would let a failed re-scan corrupt Create's retained-partial contract.
- **The walk/write split is an EXTRACTION, not a reimplementation.** `CreateCatalogWithContext`'s inline
  walk+classification block (`service.go:134-200`) becomes an exported `Walk`; research traced every branch and
  confirmed zero behavior change to Create. Do not write a parallel walker.
- **mtime is captured as Unix seconds (`int64`), `omitempty`.** `os.FileInfo` is already in scope from the
  `os.Stat` at the top of `traverseDirectory` -- zero extra syscalls, for both file and directory nodes. Unix
  seconds over RFC3339: smaller on the wire, and it matches FAT32's 2-second granularity rather than inviting
  sub-second false positives. **An absent `ModTime` on an old catalog falls back to size-only comparison and is
  never treated as the epoch** -- otherwise every pre-Phase-28 catalog would diff as entirely changed.
- **COMPAT-06 requires a real push.** Research confirmed local `main` is ~100 commits ahead of `origin/main`
  and `build.yml` has never run against any of Phases 22-28. Reading the workflow YAML proves nothing. This is
  consistent with the locked Area-4 decision ("proven by a real pushed CI run of `build.yml`").
- **WINDOWS.md sweep disposition, per research:** only **#7** (RailSide OS quit-and-relaunch) is genuinely
  closable this phase -- a cheap live check whose original blocker (protecting a shared `wails dev` session) no
  longer applies. **#4/#5** get reinforced by native-runner CI compile evidence but NOT closed. The remaining 7
  stay open pending Windows/Linux hardware this project has never had. Do not close an entry that CI does not
  actually prove.

### Claude's Discretion
- The mtime field's name, JSON tag, and time representation (RFC3339 string vs Unix seconds) — but it must be
  `omitempty` and must not disturb byte-parity for catalogs that lack it.
- The exported walk function's name and signature, and where the split lands.
- Diff result data shape crossing the Wails bridge, and whether counts are computed in Go or the frontend.
- The similarity heuristic and threshold behind the wrong-disc warning.
- Diff view presentation: grouping, ordering, virtualization (note Phase 23 already built a virtualized tree
  that may be reusable for large diffs).
- Whether the four diff states get their own color tokens or reuse existing ones.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/catalog/service.go:133,202` — `CreateCatalogWithContext` calls `traverseDirectory` then
  `WriteCatalogFrom` as two separable steps. The split point for a write-free walk.
- `internal/catalog/atomicwrite.go:37` — `WriteFileAtomic(path, data, perm)`, Phase 27-hardened. Unchanged
  signature since Phase 25.
- `internal/catalog/duplicate.go:23` — `DuplicateCatalog`'s `-copy` collision loop, reusable for "keep both".
- `app.go:993-1026` — `DeleteCatalog` (containment-gated delete-to-Trash), reusable for "remove from library".
- `internal/volumes/volumes.go:47` — `volumes.List()`, and `App.ListVolumes` (`app.go:844`).
- `frontend/src/components/workspace/create/VolumePicker.tsx` — a generic controlled `selected`/`onSelect`
  component with no create-flow coupling. Directly reusable for the re-scan volume picker.
- `app.go:188` — `throttledProgress` (≤1 emit/200ms) and the `scan:progress` event.
- Phase 23's virtualized tree — a candidate for rendering a large diff.

### Established Patterns
- `runtime.EventsEmit` is called from `app.go` ONLY; `internal/*` stays usable from the CLI with no Wails
  runtime attached (COMPAT-04).
- Every renderer-facing binding that takes a path fails closed on an empty `catalogDir` and gates through
  `osutil.ContainsPath`.
- Go tests are table-driven `*_test.go` beside source. **No frontend test framework by design** (TEST-01
  deferred) — frontend proof is `npx tsc --noEmit` + `npm run build` + live dev-browser on `:34115`.

### Integration Points
- `pkg/models/catalog.go` → new omitempty mtime field; a diff result type.
- `internal/catalog/` → the exported write-free walk; a new diff package or file.
- `app.go` → re-scan and diff-resolution bindings; reuses existing scan progress/cancel.
- `frontend/src/components/workspace/UnreadableCatalogPanel.tsx:111-113` → the explicit stub comment where
  STATE-03's trio lands.
- `frontend/src/components/workspace/DetailsPanel.tsx` → the re-scan entry point (its actions menu is Phase 27's).
- `.github/workflows/build.yml` → the COMPAT-06 proof run.
- `.planning/WINDOWS.md` → the ledger sweep.

</code_context>

<specifics>
## Specific Ideas

- **No diff logic exists anywhere in the repo.** The scout grepped `internal/`, `pkg/`, and `frontend/src`:
  the only hits are unrelated prose. Added/removed/changed/unchanged reconciliation is genuinely from scratch.
- **The `Unreadable`/`ReadError` marker is set in exactly two places** (`service.go:178-185` for a vanished scan
  root, `service.go:310-316,388-390` for the recursive child case) and always on exactly one origin node, never
  propagated to ancestors. That invariant matters for the diff's fourth state.
- **`wails.json` already reads `productVersion: "3.0.0"`** — bumped in Phase 26.
- `.planning/STATE.md:240` records CRED-04/CRED-05 (Windows OV signing credentials) as an open deferred item
  from the v2.3.0 close. That is pre-existing and not this phase's to fix, but it bounds what COMPAT-06 can honestly claim.
- WINDOWS.md entries #4/#5/#8/#9/#10/#11 are all Windows/Linux runtime-unverified items — runtime behavior
  gaps, not build-breaking, but they are the "sweep before v3.0.0 ships" backlog this phase dispositions.

</specifics>

<deferred>
## Deferred Ideas

- Content-hash-based change detection — rejected for this phase: it re-reads every byte, which is unusable on
  the optical and removable media this app targets. Size+mtime is the proportionate check.
- Persisting volume identity (name/UUID) in the catalog to enable a future "re-scan the same disc" shortcut —
  deliberately rejected as contrary to ACT-08.
- Provisioning the Windows OV signing credentials (CRED-04/CRED-05) — carried from the v2.3.0 close, out of
  scope here.
- Three-way merge or per-entry selective reconciliation — the phase specifies overwrite / keep both / discard,
  not a per-file picker.

</deferred>
