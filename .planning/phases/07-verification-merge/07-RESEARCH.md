# Phase 7: Verification + Merge - Research

**Researched:** 2026-03-25
**Domain:** Git merge hygiene, Wails cross-platform build verification, feature parity smoke testing
**Confidence:** HIGH

## Summary

Phase 7 is the final gate before `feature/go-refactor-2.0.0-clean` lands on `main`. Its two requirements are mechanical: verify the branch merges cleanly without committing build artifacts (REL-02: .gitignore coverage) and confirm the merged result is free of bloat (REL-01: no node_modules, build/bin, or archives in git history after merge).

All six upstream phases are complete and verified. The feature branch is currently 78 commits ahead of `main` (which itself is the merge base — main has not moved since the branch was cut from it in August 2025). A fast-forward merge is therefore possible and produces the simplest history. The branch already has comprehensive .gitignore coverage added during migration.

The only non-trivial work in this phase is: (1) a pre-merge audit confirming nothing tracked in the branch is bloat, (2) a macOS smoke test covering the three tabs, (3) a `wails build` clean-build confirmation, and (4) the merge itself with a tag. Windows and Linux cross-compilation verification is noted as a concern from STATE.md but can be delegated to CI (GitHub Actions workflow already exists).

**Primary recommendation:** Audit the tracked file list, run a full `wails build`, execute the three-tab smoke test on macOS, then perform a squash-or-merge-commit merge to main with a v2.0.0 tag.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| REL-01 | Clean branch merges to main with no bloat (no node_modules, binaries, or archive) | git ls-files audit confirms node_modules and build/bin are NOT tracked; storcat-project.html/json (537K/665K) are tracked but are example catalogs, not build artifacts — planner must decide disposition |
| REL-02 | .gitignore covers node_modules/, build/bin/, .DS_Store, dist/ | .gitignore already contains all four patterns; verified by reading the file directly |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

- **Branch**: Work on `feature/go-refactor-2.0.0-clean`, merge to `main` when complete
- **No Electron dependencies**: All fixes must use Go/Wails patterns, not Electron APIs
- **Backward compatibility**: v1.0 catalog JSON/HTML files must remain readable
- **API surface**: `window.electronAPI` interface must be maintained
- **GSD Workflow**: Use `/gsd:execute-phase` entry point for phase work
- **Code navigation**: Prefer LSP over Grep/Read

## Standard Stack

### Core Tools
| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| wails | v2.10.2 | Desktop app framework + build system | Project runtime |
| go | 1.26.1 (local) / 1.21 (CI) | Language runtime | Project language |
| node | v20.19.3 | Frontend build (Vite) | Required by wails frontend build |
| npm | 11.12.0 | Package manager | Used by wails frontend:install |
| git | system | Version control | Merge and tag operations |
| docker | v29.3.0 | Linux cross-compilation | `scripts/build-linux-docker.sh` |

### Build Commands
| Command | Platform | Output |
|---------|----------|--------|
| `wails build -clean -platform darwin/universal` | macOS (local) | `build/bin/StorCat.app` |
| `wails build -clean -platform windows/amd64` | macOS cross-compile | `build/bin/StorCat.exe` |
| `wails build -clean -platform linux/amd64` | Docker or Linux CI | `build/bin/StorCat` |
| `go test ./...` | Any | Unit test results |

**Version injection (ldflags):**
```bash
wails build -clean -platform darwin/universal -ldflags "-X main.Version=2.0.0"
```

## Architecture Patterns

### Merge Strategy

The feature branch was cut from `main` at commit `9c38abb1` (August 2025), and main has not moved since. The merge base IS main's HEAD. Options:

**Option A: Fast-forward merge (recommended)**
```bash
git checkout main
git merge --ff-only feature/go-refactor-2.0.0-clean
git tag v2.0.0
git push origin main --tags
```
Produces the cleanest history. No merge commit. All 78 commits visible on main.

**Option B: Merge commit**
```bash
git checkout main
git merge --no-ff feature/go-refactor-2.0.0-clean -m "feat: v2.0.0 Go/Wails migration"
git tag v2.0.0
```
Creates a merge commit. Useful as a clear milestone marker in log.

**Option C: Squash merge (not recommended)**
Loses all commit granularity. 78 commits of documented work collapsed to one. Not appropriate here.

**Recommendation:** Fast-forward merge. The history IS the documentation. The MERGE-CHECKLIST.md approach (PR → merge) is also viable and would trigger CI builds.

### Pre-Merge Audit Pattern

```bash
# 1. Verify no bloat is tracked
git ls-files | grep -E "^(node_modules|build/bin|dist/)" | wc -l
# Expected: 0

# 2. Verify .gitignore covers required patterns
grep -E "(node_modules|build/bin|\.DS_Store|dist/)" .gitignore
# Expected: all four present

# 3. Check for large tracked files (potential accidental commits)
git ls-files | xargs -I{} du -sh {} 2>/dev/null | sort -rh | head -20

# 4. Confirm Go tests pass
go test ./...

# 5. Confirm clean build
wails build -clean -platform darwin/universal -ldflags "-X main.Version=2.0.0"
```

### Three-Tab Smoke Test Pattern

The MERGE-CHECKLIST.md already defines the acceptance criteria. The planner should structure this as an explicit task:

| Tab | Operation | Pass Condition |
|-----|-----------|----------------|
| Create | Select a directory, click Create | Catalog .json and .html files appear in output dir; fileCount > 0 displayed |
| Search | Select catalog directory, enter search term, click Search | Results table populates with matches |
| Browse | Select catalog directory | Files list loads with size and date columns populated |

### Post-Merge Tag Pattern
```bash
git tag -a v2.0.0 -m "StorCat v2.0.0 — Go/Wails migration"
git push origin v2.0.0
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-platform Linux build | Native cross-compile | `scripts/build-linux-docker.sh` or GitHub Actions | Linux WebKitGTK deps not available on macOS |
| Windows build from macOS | NSIS installer setup | `wails build -platform windows/amd64` (no installer needed, just .exe) | Wails handles Windows EXE without NSIS |
| Git history cleanup | `git filter-branch` | Not needed — node_modules never entered this branch | The clean branch was created specifically to avoid this |

## Common Pitfalls

### Pitfall 1: storcat-project.html/json are tracked (537K / 665K)
**What goes wrong:** Two large catalog example files (`storcat-project.html`, `storcat-project.json`) are committed to the feature branch. They are example output, not source code.
**Why it happens:** They were copied over during branch creation and never excluded.
**How to avoid:** Decide explicitly: keep them as reference examples (they're in the repo root, not build/bin) OR remove and add to .gitignore. They are NOT build artifacts — they're content files, so they don't violate REL-01 literally. But they add ~1.2MB of catalog data to git.
**Warning signs:** `git ls-files | grep "storcat-project"` returns results

### Pitfall 2: storcat-repo-consolidation.md is untracked
**What goes wrong:** The file `storcat-repo-consolidation.md` appears in `git status` as untracked. If committed accidentally, it adds a development artifact to main.
**How to avoid:** Verify it should NOT be committed. It appears to be a planning document from repo consolidation work. Add to .gitignore or delete before merge.

### Pitfall 3: build/darwin/Info.dev.plist is untracked
**What goes wrong:** `build/darwin/Info.dev.plist` is untracked. The `build/darwin/` directory contains plist files used by Wails for macOS builds. `Info.plist` IS tracked; `Info.dev.plist` is a dev-mode variant.
**How to avoid:** Either add `build/darwin/Info.dev.plist` to .gitignore (it's a generated dev file) or commit it if it's intentional. Leaving it untracked means it clutters `git status` but doesn't enter git history.
**Warning signs:** `build/darwin/Info.dev.plist` in `git status --short`

### Pitfall 4: .planning/codebase/ is untracked
**What goes wrong:** The `.planning/codebase/` directory (7 architecture analysis files) is untracked. These are planning artifacts.
**How to avoid:** The `.planning/` directory IS tracked in git (phase plans and docs are committed). If `codebase/` is untracked, it was intentionally not committed or was accidentally excluded. Decision needed before merge: commit these analysis docs or leave them local-only.

### Pitfall 5: go.mod has a commented-out local replace directive
**What goes wrong:** `go.mod` ends with `// replace github.com/wailsapp/wails/v2 v2.10.2 => /Users/ken/go/pkg/mod`. A commented-out local path replace directive. If uncommented, it breaks builds on other machines.
**Why it matters:** It is currently commented out, so it does NOT break builds. But it is a dev artifact that should be removed before merge.
**Warning signs:** `grep "replace" go.mod` returns the line

### Pitfall 6: GitHub Actions workflow uses Go 1.21, local is 1.26.1
**What goes wrong:** The `build.yml` workflow specifies `go-version: '1.21'` while the local toolchain is 1.26.1 and `go.mod` requires `go 1.23`. This means CI may fail to build.
**Why it happens:** The workflow was written for an earlier Go version and not updated.
**How to avoid:** Update `build.yml` to specify `go-version: '1.23'` (matching go.mod minimum).
**Warning signs:** CI build failure after merge with "go.mod requires go 1.23 but toolchain is 1.21"

### Pitfall 7: Windows cross-compile from macOS
**What goes wrong:** `wails build -platform windows/amd64` from macOS requires the NSIS installer to be installed for creating .exe installers, but bare `.exe` builds work without it.
**How to avoid:** The build script `scripts/build-windows.sh` uses plain `wails build` which produces a bare .exe. This works from macOS. Installer packaging (.msi) would require NSIS and `--nsis` flag. Don't add `--nsis` unless explicitly needed.

## Code Examples

### Verify no bloat is tracked
```bash
# Source: local git inspection
git ls-files | grep -E "^(node_modules|build/bin|dist/)" | wc -l
# Must return 0

git ls-files | grep -c "."
# Currently ~143 files tracked — reasonable for this codebase
```

### Current .gitignore (verified HIGH confidence)
```
# Build outputs
dist/
build/bin/

# Dependencies
node_modules/
frontend/node_modules/

# OS files
.DS_Store
Thumbs.db

# IDE
.vscode/
.idea/

# Go
*.exe~

# Misc
storcat-icons/
```
REL-02 patterns present: `node_modules/`, `frontend/node_modules/`, `build/bin/`, `.DS_Store`, `dist/` — all four required patterns are covered.

### Clean build with version injection
```bash
# Source: scripts/build-macos.sh + phase 06 decisions
wails build -clean -platform darwin/universal -ldflags "-X main.Version=2.0.0"
# Output: build/bin/StorCat.app (~8-11MB)
```

### Fast-forward merge sequence
```bash
# Source: git documentation + project merge strategy
git checkout main
git pull origin main                          # ensure main is current
git merge --ff-only feature/go-refactor-2.0.0-clean
git tag -a v2.0.0 -m "StorCat v2.0.0 — Go/Wails migration"
git push origin main --tags
```

### Go test run
```bash
# Source: local execution (confirmed passing)
go test ./...
# ok  storcat-wails/internal/catalog
# ok  storcat-wails/internal/config
# ok  storcat-wails/internal/search
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Electron + node_modules (415MB) | Go + Wails (8-11MB) | This migration | 97% size reduction |
| Hardcoded APP_VERSION constant | ldflags injection from wails.json | Phase 6 | Version always matches build |
| Array-wrapped catalog JSON | Bare object JSON | Phase 1 | v1.0 backward compat maintained |
| Window persistence stub | Full config-backed persistence | Phase 3 | Size, position, toggle all work |

## Open Questions

1. **storcat-project.html/json disposition**
   - What we know: Two ~600KB catalog files are tracked in the repo root. They are example output, not build artifacts.
   - What's unclear: Should they stay in git (useful as reference examples)? Should they be removed (they're large and clutter the repo)?
   - Recommendation: Remove them from git and add `storcat-project.*` to .gitignore. They are generated output and don't belong in source control. The `examples/` directory already has `sd01.html` and `sd01.json` as proper examples.

2. **storcat-repo-consolidation.md disposition**
   - What we know: Untracked file in repo root. Name suggests it was a planning document for repo consolidation work.
   - What's unclear: Is it meant to be committed or deleted?
   - Recommendation: Delete it before merge. It is a development artifact, not user-facing documentation.

3. **build/darwin/Info.dev.plist disposition**
   - What we know: Generated by `wails dev` for dev-mode macOS builds. Currently untracked.
   - What's unclear: Should it be gitignored to prevent future accidental commits?
   - Recommendation: Add `build/darwin/Info.dev.plist` to .gitignore. The production plist (`Info.plist`) is already tracked; the dev variant is generated noise.

4. **.planning/codebase/ disposition**
   - What we know: 7 analysis documents are untracked. The rest of `.planning/` IS committed.
   - What's unclear: Were these intentionally excluded or accidentally never staged?
   - Recommendation: Commit them. They are planning artifacts documenting the codebase analysis that drove this migration. They belong in git alongside the phase plans.

5. **GitHub Actions workflow Go version**
   - What we know: `build.yml` specifies `go-version: '1.21'`, but `go.mod` requires `go 1.23`.
   - What's unclear: Will Go 1.21 toolchain accept a `go 1.23` module directive? Go 1.21+ allows toolchain selection via `go` directive, but older toolchains reject higher module versions.
   - Recommendation: Update `build.yml` to `go-version: '1.23'` before merge. LOW confidence that 1.21 will succeed; HIGH confidence that 1.23+ will succeed.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| wails CLI | Build + dev | ✓ | v2.10.2 | — |
| Go | All Go compilation | ✓ | 1.26.1 (darwin/arm64) | — |
| Node.js | Frontend build (Vite) | ✓ | v20.19.3 | — |
| npm | frontend:install | ✓ | 11.12.0 | — |
| Docker | Linux cross-compilation | ✓ | v29.3.0 | GitHub Actions linux runner |
| git | Merge + tag | ✓ | system | — |
| Vite (frontend) | Frontend build | ✓ | In node_modules/.bin/vite | — |
| GitHub Actions | CI verification | ✓ (workflow exists) | — | Manual build scripts |

**Missing dependencies with no fallback:** None.

**Missing dependencies with fallback:**
- Linux build toolchain (GTK3/WebKitGTK): Not on macOS — use Docker script or delegate to CI after merge.
- Windows runtime testing: No Windows machine available locally — CI can verify build; runtime testing is best-effort.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package (built-in) |
| Config file | none — standard `go test` |
| Quick run command | `go test ./...` |
| Full suite command | `go test ./... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| REL-01 | No tracked bloat in git | smoke | `git ls-files \| grep -E "^(node_modules\|build/bin\|dist/)" \| wc -l` returns 0 | ✅ (git audit) |
| REL-02 | .gitignore covers required patterns | smoke | `grep -c "node_modules\|build/bin\|\.DS_Store\|dist/" .gitignore` returns 4+ | ✅ (file read) |
| Phase gate | Go tests pass | unit | `go test ./...` | ✅ Wave 0 |
| Phase gate | macOS build succeeds | integration | `wails build -clean -platform darwin/universal` | ✅ (script) |
| Phase gate | Three tabs work | manual/UAT | Launch app, exercise Create/Search/Browse | human |

### Sampling Rate
- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test ./... && git ls-files | grep -E "^(node_modules|build/bin|dist/)" | wc -l`
- **Phase gate:** Full `wails build` clean output + three-tab smoke test before merge commit

### Wave 0 Gaps
None — existing test infrastructure covers all automated phase requirements. The three-tab smoke test is manual-only (requires running Wails app on macOS).

## Sources

### Primary (HIGH confidence)
- Local git inspection — `git ls-files`, `git log`, `git status`, `git diff main..HEAD`
- Local file reads — `.gitignore`, `go.mod`, `wails.json`, `MERGE-CHECKLIST.md`, `REQUIREMENTS.md`, `STATE.md`, `ROADMAP.md`
- Local execution — `go test ./...` (all pass), `wails version` (v2.10.2), tool versions

### Secondary (MEDIUM confidence)
- GitHub Actions `build.yml` — documents CI build matrix and Go version mismatch
- `scripts/build-macos.sh` — documents exact wails build command used in production builds

### Tertiary (LOW confidence)
- Go 1.21 → 1.23 module directive compatibility: not verified by running CI locally; based on Go toolchain version selection docs

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all tool versions verified by running them locally
- Architecture: HIGH — merge strategy verified against actual git state (fast-forward possible)
- Pitfalls: HIGH for items verified by file inspection; LOW for CI compatibility (not run)
- REL-01 compliance: HIGH — `git ls-files` confirms no node_modules or build/bin tracked
- REL-02 compliance: HIGH — .gitignore file read directly, all four patterns present

**Research date:** 2026-03-25
**Valid until:** 2026-04-25 (stable domain; only relevant while branch is open)
