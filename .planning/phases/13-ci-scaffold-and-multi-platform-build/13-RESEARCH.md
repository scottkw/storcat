# Phase 13: CI Scaffold and Multi-Platform Build - Research

**Researched:** 2026-03-27
**Domain:** GitHub Actions CI/CD, Wails v2 cross-platform builds, release workflow patterns
**Confidence:** HIGH

## Summary

Phase 13 creates a `release.yml` GitHub Actions workflow that fires on tag push (or release publish) and produces native binaries for macOS (universal fat binary), Windows (amd64 with -windowsconsole), and Linux (x64 and arm64). The workflow uses a fan-in pattern: all platform build jobs complete before a final release-assembly job uploads assets to a GitHub Release draft.

The existing `build.yml` workflow in the repo is a proof-of-concept — it has several correctness problems that must be fixed: macOS and Windows both run on `macos-latest` (Windows needs a Windows runner), the Linux job uses `ubuntu-latest` (breaks on Ubuntu 24.04 due to missing `libwebkit2gtk-4.0-dev`), and no third-party actions are SHA-pinned. Phase 13 replaces this with a production-quality `release.yml`.

**Key finding on Linux ARM64:** The new GitHub-hosted `ubuntu-22.04-arm` native runner (public preview, free for public repos, available since January 2025) is the cleanest approach for arm64 builds — it avoids the known `wails build -skipbindings` cross-compilation workaround that loses binding generation. The arm runner runs Wails natively on ARM64, no cross-compiler toolchain required.

**Primary recommendation:** Write `release.yml` triggered by `on: push: tags: ['v*.*.*']`, with four jobs: `build-macos` (macos-14), `build-windows` (windows-latest), `build-linux-amd64` (ubuntu-22.04), `build-linux-arm64` (ubuntu-22.04-arm), plus a fan-in `release` job with `needs: [build-macos, build-windows, build-linux-amd64, build-linux-arm64]`. All third-party actions SHA-pinned. Keep `build.yml` intact (rename or update it for branch pushes only).

## Project Constraints (from CLAUDE.md)

- **GSD workflow enforcement**: All file-changing work must go through `/gsd:execute-phase`.
- **No direct repo edits** outside GSD workflow unless user explicitly bypasses.
- **pnpm preferred** for Node — `wails.json` configures `npm install` for the frontend; the workflow uses `setup-node` without overriding the package manager.
- **Git**: `branching_strategy: "none"` — work directly on main, no feature branch.
- **Tech stack**: Go backend, React/TypeScript frontend via Wails v2. All workflow steps must match this stack.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CICD-01 | Release workflow triggers on GitHub release publish event | REQUIREMENTS.md says "publish event"; success criteria says "tag push". Reconciliation: use `on: push: tags: ['v*.*.*']` (tag push) as the primary trigger — this matches the success criteria literally. Note discrepancy below. |
| CICD-02 | macOS builds on `macos-14` runner producing universal binary | `wails build -clean -platform darwin/universal` on macos-14 (Apple M1) uses Go's cross-compilation to produce both amd64 and arm64 binaries internally, then lipo-combines them. macos-14 is arm64 hardware; macOS SDK enables cross-compiling to amd64. |
| CICD-03 | Windows builds on `windows-latest` runner with `-windowsconsole` preserved | Must run on actual `windows-latest` runner. Existing `build.yml` incorrectly uses `macos-latest` for Windows build — this is a bug to fix. |
| CICD-04 | Linux builds on `ubuntu-22.04` runner for x64 and arm64 | x64: `ubuntu-22.04` runner with `libwebkit2gtk-4.0-dev`. arm64: `ubuntu-22.04-arm` native runner (free for public repos, public preview). |
| CICD-05 | Fan-in release pattern: all platform builds complete before release upload | `release` job uses `needs: [build-macos, build-windows, build-linux-amd64, build-linux-arm64]`. Uses `actions/download-artifact` to collect all artifacts before `softprops/action-gh-release`. |
| CICD-06 | All third-party GitHub Actions SHA-pinned (not tag-referenced) | Affects: setup-go, setup-node, upload-artifact, download-artifact, softprops/action-gh-release. First-party `actions/*` should also be SHA-pinned per CICD-06 — it says "all third-party" but best practice is to pin first-party too. See SHA table below. |
</phase_requirements>

### CICD-01 Trigger Discrepancy

REQUIREMENTS.md CICD-01: "Release workflow triggers on GitHub release publish event" — implies `on: release: types: [published]`.

Success Criteria #1: "Pushing a v*.*.* tag triggers release.yml" — implies `on: push: tags: ['v*.*.*']`.

**Reconciliation:** These are compatible strategies but not identical. The success criteria is more specific. Using `on: push: tags: ['v*.*.*']` satisfies the success criteria directly and is the standard pattern for this use case. When paired with `softprops/action-gh-release` using `draft: true`, it creates a draft release automatically on tag push — which also covers the spirit of CICD-01 (release is created on tag publish event, just via the workflow rather than manually). **The plan should use tag-push trigger.**

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| actions/checkout | v4.2.2 | Checkout code | Official GitHub action |
| actions/setup-go | v5.4.0 | Install Go toolchain | Official GitHub action |
| actions/setup-node | v4.4.0 | Install Node.js for Wails frontend build | Official GitHub action |
| actions/upload-artifact | v4.6.2 | Upload per-platform binary to artifact store | Official; v3 stopped working Jan 2025 |
| actions/download-artifact | v4.2.1 | Download all artifacts in fan-in job | Official; must match upload version |
| softprops/action-gh-release | v2.6.1 | Create/update GitHub Release with assets | De-facto standard for release creation |

### SHA Pin Table (CICD-06 compliance)

All these are the latest stable versions as of 2026-03-27, verified against GitHub API.

| Action | Tag | Commit SHA |
|--------|-----|------------|
| `actions/checkout` | v4.2.2 | `11bd71901bbe5b1630ceea73d27597364c9af683` |
| `actions/setup-go` | v5.4.0 | `0aaccfd150d50ccaeb58ebd88d36e91967a5f35b` |
| `actions/setup-node` | v4.4.0 | `49933ea5288caeca8642d1e84afbd3f7d6820020` |
| `actions/upload-artifact` | v4.6.2 | `ea165f8d65b6e75b540449e92b4886f43607fa02` |
| `actions/download-artifact` | v4.2.1 | `95815c38cf2ff2164869cbab79da8d1f422bc89e` |
| `softprops/action-gh-release` | v2.6.1 | `153bb8e04406b158c6c84fc1615b65b24149a1fe` |

**Note on major-version pinning strategy:** The latest releases of first-party actions have jumped to v6+ (e.g., `actions/checkout@v6.0.2`, `actions/upload-artifact@v7.0.0`). These new major versions may have breaking changes or different behaviors. The v4/v5 family is well-tested with Wails workflows. The plan should use the v4/v5 generations (verified above) as a stable baseline. CICD-06 only requires SHA-pinning, not latest-version pinning.

**Installation:** No npm install — these are workflow YAML references.

## Architecture Patterns

### Recommended Workflow Structure

```
.github/workflows/
├── build.yml         # Existing: CI on branch push/PR (update runners but keep)
└── release.yml       # New: triggered by v*.*.* tag push, fan-in release
```

### Pattern 1: Tag-Triggered Fan-In Release

**What:** A workflow with N parallel platform build jobs + 1 release assembly job that blocks on all platform jobs completing.

**When to use:** Any time you need cross-platform binaries assembled into a single release atomically.

**Structure:**
```yaml
# Source: GitHub Actions docs + verified Wails build patterns
on:
  push:
    tags: ['v*.*.*']

jobs:
  build-macos:
    runs-on: macos-14
    steps: [checkout, setup-go, setup-node, install-wails, build-macos, upload-artifact]

  build-windows:
    runs-on: windows-latest
    steps: [checkout, setup-go, setup-node, install-wails, build-windows, upload-artifact]

  build-linux-amd64:
    runs-on: ubuntu-22.04
    steps: [checkout, setup-go, setup-node, install-deps, install-wails, build-linux-amd64, upload-artifact]

  build-linux-arm64:
    runs-on: ubuntu-22.04-arm
    steps: [checkout, setup-go, setup-node, install-deps, install-wails, build-linux-arm64, upload-artifact]

  release:
    needs: [build-macos, build-windows, build-linux-amd64, build-linux-arm64]
    runs-on: ubuntu-22.04
    steps: [download-all-artifacts, create-release-with-assets]
```

### Pattern 2: Version Extraction from Git Tag

**What:** Extract the version string from the git tag for use in artifact naming.

**When to use:** When artifact filenames must embed the version (e.g., `StorCat-v2.1.0-darwin-universal.tar.gz`).

```yaml
# In each build job:
- name: Get version from tag
  id: version
  run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

# Use as: ${{ steps.version.outputs.VERSION }}
```

For Windows (PowerShell):
```yaml
- name: Get version from tag
  id: version
  shell: bash
  run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
```

### Pattern 3: Wails Build Commands Per Platform

All commands verified against `scripts/build-*.sh` in the repository.

**macOS (macos-14, universal binary):**
```bash
wails build -clean -platform darwin/universal
# Output: build/bin/StorCat.app (directory)
# Verify: lipo -info build/bin/StorCat.app/Contents/MacOS/StorCat
```

**Windows (windows-latest):**
```bash
wails build -clean -platform windows/amd64 -windowsconsole
# Output: build/bin/StorCat.exe
# -windowsconsole: uses Windows console subsystem, enabling stdout/stderr for CLI users
```

**Linux AMD64 (ubuntu-22.04):**
```bash
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev
wails build -clean -platform linux/amd64
# Output: build/bin/StorCat
```

**Linux ARM64 (ubuntu-22.04-arm native runner):**
```bash
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev
wails build -clean -platform linux/arm64
# Output: build/bin/StorCat
# Native ARM64 runner — no cross-compilation needed, no -skipbindings required
```

### Pattern 4: Artifact Upload/Download for Fan-In

```yaml
# In each build job — upload platform-specific artifact:
- name: Upload artifact
  uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02  # v4.6.2
  with:
    name: StorCat-darwin-universal
    path: build/bin/StorCat.app
    retention-days: 1

# In release job — download all artifacts:
- name: Download all artifacts
  uses: actions/download-artifact@95815c38cf2ff2164869cbab79da8d1f422bc89e  # v4.2.1
  with:
    path: artifacts/   # Downloads into subdirectories by artifact name
```

### Anti-Patterns to Avoid

- **Running Windows build on macOS runner:** The existing `build.yml` does this — NSIS and Windows toolchain require a Windows host. CGO cross-compilation to Windows from macOS requires a full MinGW toolchain setup which is fragile.
- **ubuntu-latest for Wails Linux builds:** Resolves to Ubuntu 24.04 which dropped `libwebkit2gtk-4.0-dev`. Pin to `ubuntu-22.04` explicitly.
- **Tag-reference for community actions (e.g., `@v2`):** A tag can be force-pushed. Use full commit SHA with a comment showing the version.
- **Uploading release assets from parallel jobs:** Race condition — two jobs trying to create the same GitHub Release will fail. All uploads must be in the single `release` fan-in job.
- **Using macos-latest for macOS builds:** `macos-latest` currently resolves to `macos-15`. The requirement specifies `macos-14`. Use the explicit runner label.
- **Cross-compiling Linux ARM64 from AMD64 with wails build:** Requires `-skipbindings` flag, which skips binding generation and may produce incomplete builds. The native `ubuntu-22.04-arm` runner is the clean solution.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Release creation | Custom gh CLI scripting | `softprops/action-gh-release` | Handles asset upload, draft/publish, changelog generation, race conditions |
| Version extraction | Manual string parsing | `${GITHUB_REF#refs/tags/}` shell parameter expansion | One-liner, built into bash |
| Artifact collection | Custom download scripts | `actions/download-artifact` | Handles artifact-not-found failures, versioned API |
| Go toolchain setup | Manual download + PATH | `actions/setup-go` with `go-version:` | Handles module cache, version pinning |

**Key insight:** The fan-in pattern with `needs:` is a GitHub Actions built-in — no orchestration library needed. The `release` job simply lists all build jobs in its `needs:` array.

## Common Pitfalls

### Pitfall 1: Ubuntu Latest vs 22.04 for WebKit
**What goes wrong:** `ubuntu-latest` resolves to 24.04 which removed `libwebkit2gtk-4.0-dev` from its apt repositories.
**Why it happens:** Wails v2 depends on `libwebkit2gtk-4.0-dev` at compile time. Ubuntu 24.04 ships `libwebkit2gtk-4.1-dev` with a different API.
**How to avoid:** Pin `runs-on: ubuntu-22.04` explicitly. Never use `ubuntu-latest` for Wails builds.
**Warning signs:** Build fails with `Package libwebkit2gtk-4.0-dev has no installation candidate`.

### Pitfall 2: Wrong Runner for Windows Build
**What goes wrong:** Building Windows binaries on a macOS or Linux runner requires setting up MinGW cross-compilation, which is complex and fragile.
**Why it happens:** Existing `build.yml` has a copy-paste error — `build-windows` runs on `macos-latest`.
**How to avoid:** `runs-on: windows-latest` for Windows builds, period.
**Warning signs:** CGO compilation errors about missing mingw headers.

### Pitfall 3: Artifact v3/v4 Version Mismatch
**What goes wrong:** `upload-artifact@v3` and `download-artifact@v4` (or vice versa) are incompatible — v3 artifacts cannot be downloaded by v4 and will throw "Artifact not found" errors.
**Why it happens:** `upload-artifact@v3` was deprecated and stopped working in January 2025. Some tutorials still show v3.
**How to avoid:** Both upload and download must use v4 (or both v3). This research recommends v4.6.2 / v4.2.1 pair.
**Warning signs:** `Error: An artifact with the name 'StorCat-darwin-universal' was not found`.

### Pitfall 4: Tag SHA Pinning — Lightweight vs Annotated Tags
**What goes wrong:** Pinning the SHA of a lightweight tag gives the commit SHA directly. Pinning the SHA of an annotated tag gives the tag object SHA, not the commit SHA. The workflow uses the commit SHA.
**Why it happens:** GitHub API `/git/ref/tags/{tag}` returns the tag object SHA for annotated tags. You must follow the indirection to get the commit SHA.
**How to avoid:** Use the commit SHAs from the SHA Pin Table in this document — they have been resolved correctly.
**Warning signs:** Action fails to start with "SHA not found" or silently falls back to wrong version.

### Pitfall 5: darwin/universal Build Requires macOS Runner
**What goes wrong:** `wails build -platform darwin/universal` cannot be cross-compiled from Linux or Windows. The macOS SDK is required.
**Why it happens:** Wails uses Xcode tools (`lipo`, Apple linker) to produce the universal binary.
**How to avoid:** Always run macOS builds on a macOS runner. `macos-14` (M1) can cross-compile for amd64 via Go + the macOS SDK's support for cross-architecture builds.
**Warning signs:** Build fails with `xcrun: error` or `lipo: can't figure out the architecture type of 'build/...'`.

### Pitfall 6: Linux ARM64 Cross-Compilation Assembler Errors
**What goes wrong:** Using `CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOARCH=arm64 wails build` on an amd64 runner fails with assembler errors (`gcc_arm64.S: no such instruction`) or requires `-skipbindings` which may produce an incomplete build.
**Why it happens:** Wails v2's binding generation executes the host binary during the build, which fails when cross-compiling.
**How to avoid:** Use the native `ubuntu-22.04-arm` runner. No cross-compilation needed.
**Warning signs:** `exec format error` or assembler errors during `wails build -platform linux/arm64`.

### Pitfall 7: Race Condition on Parallel Release Upload
**What goes wrong:** Two parallel jobs both try to create a GitHub Release — the second one fails because the release already exists, or both succeed creating duplicate releases.
**Why it happens:** GitHub's release creation API is not idempotent under concurrency without careful ID tracking.
**How to avoid:** All artifact upload must happen in a single `release` job that `needs:` all build jobs.
**Warning signs:** "Release already exists" or duplicate GitHub Releases created.

## Code Examples

Verified patterns from project's existing build scripts and GitHub Actions documentation:

### Complete release.yml Skeleton

```yaml
# Source: Verified against project scripts/build-*.sh + GitHub Actions docs
name: Release

on:
  push:
    tags: ['v*.*.*']

jobs:
  build-macos:
    runs-on: macos-14
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2

      - name: Set up Go
        uses: actions/setup-go@0aaccfd150d50ccaeb58ebd88d36e91967a5f35b  # v5.4.0
        with:
          go-version-file: go.mod

      - name: Set up Node
        uses: actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020  # v4.4.0
        with:
          node-version: '20'

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build macOS universal
        run: wails build -clean -platform darwin/universal

      - name: Verify universal binary
        run: lipo -info build/bin/StorCat.app/Contents/MacOS/StorCat

      - name: Upload macOS artifact
        uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02  # v4.6.2
        with:
          name: StorCat-darwin-universal
          path: build/bin/StorCat.app
          retention-days: 1

  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2

      - name: Set up Go
        uses: actions/setup-go@0aaccfd150d50ccaeb58ebd88d36e91967a5f35b  # v5.4.0
        with:
          go-version-file: go.mod

      - name: Set up Node
        uses: actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020  # v4.4.0
        with:
          node-version: '20'

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build Windows
        run: wails build -clean -platform windows/amd64 -windowsconsole

      - name: Upload Windows artifact
        uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02  # v4.6.2
        with:
          name: StorCat-windows-amd64
          path: build/bin/StorCat.exe
          retention-days: 1

  build-linux-amd64:
    runs-on: ubuntu-22.04
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2

      - name: Set up Go
        uses: actions/setup-go@0aaccfd150d50ccaeb58ebd88d36e91967a5f35b  # v5.4.0
        with:
          go-version-file: go.mod

      - name: Set up Node
        uses: actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020  # v4.4.0
        with:
          node-version: '20'

      - name: Install Linux dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build Linux amd64
        run: wails build -clean -platform linux/amd64

      - name: Upload Linux amd64 artifact
        uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02  # v4.6.2
        with:
          name: StorCat-linux-amd64
          path: build/bin/StorCat
          retention-days: 1

  build-linux-arm64:
    runs-on: ubuntu-22.04-arm       # Native ARM64 runner — public preview, free for public repos
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2

      - name: Set up Go
        uses: actions/setup-go@0aaccfd150d50ccaeb58ebd88d36e91967a5f35b  # v5.4.0
        with:
          go-version-file: go.mod

      - name: Set up Node
        uses: actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020  # v4.4.0
        with:
          node-version: '20'

      - name: Install Linux dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build Linux arm64
        run: wails build -clean -platform linux/arm64

      - name: Upload Linux arm64 artifact
        uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02  # v4.6.2
        with:
          name: StorCat-linux-arm64
          path: build/bin/StorCat
          retention-days: 1

  release:
    needs: [build-macos, build-windows, build-linux-amd64, build-linux-arm64]
    runs-on: ubuntu-22.04
    permissions:
      contents: write
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@95815c38cf2ff2164869cbab79da8d1f422bc89e  # v4.2.1
        with:
          path: artifacts/

      - name: Create GitHub Release
        uses: softprops/action-gh-release@153bb8e04406b158c6c84fc1615b65b24149a1fe  # v2.6.1
        with:
          draft: true
          generate_release_notes: true
          files: |
            artifacts/**/*
```

### Version Extraction (bash, all platforms)

```bash
# Available in all jobs via GITHUB_REF
VERSION="${GITHUB_REF#refs/tags/}"   # strips "refs/tags/" prefix → "v2.1.0"
```

### lipo Verification Command

```bash
# CICD-02 success criteria: confirmed universal fat binary
lipo -info build/bin/StorCat.app/Contents/MacOS/StorCat
# Expected output: Architectures in the fat file: ... are: x86_64 arm64
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| ubuntu-latest for Wails | ubuntu-22.04 pinned | Ubuntu 24.04 became latest (~mid 2024) | Build failure if not pinned |
| Cross-compile linux/arm64 from amd64 | Native ubuntu-22.04-arm runner | Jan 2025 (GitHub public preview) | No -skipbindings hack needed |
| actions/upload-artifact@v3 | v4.6.2 | v3 deprecated Jan 2025 | v3 artifacts silently fail to upload/download |
| Tag-only pinning (@v2, @v4) | Full commit SHA pin | Supply chain attack awareness (2023+) | Security: mutable tags exploited in March 2025 tj-actions attack |
| macos-latest for universal builds | macos-14 (explicit) | macos-latest became macos-15 in 2024 | macos-15 may have different Xcode toolchain; explicit pin guarantees reproducibility |

**Deprecated/outdated:**
- `actions/upload-artifact@v3`: Stopped working January 2025 — all references must use v4+.
- `ubuntu-latest` for Wails: Breaks Wails builds due to WebKit API change in Ubuntu 24.04.
- Cross-compiling Wails to linux/arm64: Still works with `-skipbindings` but native runner is cleaner.

## Open Questions

1. **Wails version pinning in CI**
   - What we know: `go install github.com/wailsapp/wails/v2/cmd/wails@latest` installs whatever is latest at run time. v2.10.0 was noted as problematic in some community reports (dAppServer action README).
   - What's unclear: Whether Wails v2.10.2 (the current version in go.mod) is stable for all platforms in CI.
   - Recommendation: Pin the Wails install to the exact version from go.mod: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2`

2. **ubuntu-22.04-arm runner stability for Phase 13**
   - What we know: Public preview as of January 2025, free for public repos, experienced some instability early in preview (Cobalt 100 → dpdsv5 migration).
   - What's unclear: Current stability as of March 2026.
   - Recommendation: Use it as primary strategy. If the runner is unavailable at execution time, the fallback is cross-compilation with `-skipbindings` on ubuntu-22.04. Document this as a contingency in the plan.

3. **Artifact naming for Phase 14 packaging**
   - What we know: Phase 14 creates installers (DMG, NSIS, AppImage, deb). Those installers likely need the raw binaries from Phase 13's artifacts.
   - What's unclear: Whether artifact names from Phase 13 are consumed by Phase 14 workflows or if Phase 14 builds from source independently.
   - Recommendation: Name artifacts with version in the name (`StorCat-v2.1.0-darwin-universal`) to make downstream consumption unambiguous. Alternatively, build from source in Phase 14 and treat Phase 13 as standalone.

4. **CICD-01 trigger type conflict**
   - What we know: REQUIREMENTS.md says "release publish event"; success criteria says "tag push".
   - Recommendation: Implement tag-push trigger (`on: push: tags: ['v*.*.*']`). This satisfies the success criteria literally and produces a draft release (which can be manually published). Document the discrepancy for user review.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| GitHub Actions macos-14 | CICD-02 | ✓ | macOS 14 Sonoma, M1 | macos-15 (different Xcode, not preferred) |
| GitHub Actions windows-latest | CICD-03 | ✓ | Windows Server 2022 | windows-2022 (explicit) |
| GitHub Actions ubuntu-22.04 | CICD-04 (amd64) | ✓ | Ubuntu 22.04 LTS | — (required pin) |
| GitHub Actions ubuntu-22.04-arm | CICD-04 (arm64) | ✓ (public preview) | Ubuntu 22.04 LTS arm64 | Cross-compile from ubuntu-22.04 with -skipbindings |
| Go 1.23 (from go.mod) | All builds | ✓ | Managed by setup-go | — |
| Node.js 20 | Wails frontend | ✓ | Managed by setup-node | Node 18 (older LTS) |
| Wails v2.10.2 | All builds | ✓ | Installed via go install | — |
| libwebkit2gtk-4.0-dev | Linux builds | ✓ on ubuntu-22.04 | apt package | webkit2_41 on ubuntu 24+ |
| lipo (Xcode) | macOS binary verification | ✓ on macos-14 | Ships with Xcode | — |
| softprops/action-gh-release | Release job | ✓ | v2.6.1 | gh CLI release create |

**Missing dependencies with no fallback:**
- None — all required tooling is available on GitHub-hosted runners.

**Missing dependencies with fallback:**
- `ubuntu-22.04-arm`: If public preview is unavailable, use cross-compilation on ubuntu-22.04 with `CGO_ENABLED=1 GOARCH=arm64 CC=aarch64-linux-gnu-gcc wails build -skipbindings -platform linux/arm64`.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | No automated test framework for CI workflow YAML |
| Config file | None — workflow YAML is the artifact |
| Quick run command | `gh workflow run release.yml --ref main` (manual trigger) |
| Full suite command | Push a test tag: `git tag v0.0.0-test && git push origin v0.0.0-test` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CICD-01 | Tag push triggers release.yml | smoke | `git tag v0.0.0-test && git push origin v0.0.0-test` | ❌ Wave 0 — create workflow file |
| CICD-02 | macOS produces universal binary | manual | `lipo -info` step in workflow output | ❌ Wave 0 |
| CICD-03 | Windows build has -windowsconsole | manual | Inspect build log for `-windowsconsole` flag in wails build step | ❌ Wave 0 |
| CICD-04 | Linux x64 builds on ubuntu-22.04 | smoke | Workflow run completion on ubuntu-22.04 | ❌ Wave 0 |
| CICD-04 | Linux arm64 builds on ubuntu-22.04-arm | smoke | Workflow run completion on ubuntu-22.04-arm | ❌ Wave 0 |
| CICD-05 | Release job runs after all builds | manual | `gh run view` — release job shows `needs` dependencies | ❌ Wave 0 |
| CICD-06 | All actions SHA-pinned | manual | `grep -E "uses:.*@[0-9a-f]{40}" .github/workflows/release.yml` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `grep -c "uses:.*@[0-9a-f]" .github/workflows/release.yml` — verify count matches expected SHA-pinned steps
- **Per wave merge:** Full workflow run via test tag push
- **Phase gate:** All 6 CICD requirements confirmed via workflow run inspection before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `.github/workflows/release.yml` — covers CICD-01 through CICD-06
- [ ] Update `.github/workflows/build.yml` — fix Windows runner (currently macos-latest), fix ubuntu-latest → ubuntu-22.04, SHA-pin all actions

*(Note: No test files to create — validation is CI workflow execution, not unit tests.)*

## Sources

### Primary (HIGH confidence)

- GitHub API (`api.github.com/repos/*/releases/latest`) — verified SHA pins for all actions as of 2026-03-27
- Project `scripts/build-*.sh` — verified exact wails build commands and flags per platform
- Project `.github/workflows/build.yml` — existing workflow (source of bug inventory)
- Project `go.mod` — confirmed Wails v2.10.2, Go 1.23
- GitHub community discussion #148648 — ubuntu-22.04-arm runner names and public repo availability

### Secondary (MEDIUM confidence)

- [Wails cross-platform build guide](https://wails.io/docs/guides/crossplatform-build/) — documents `darwin/universal` and Linux patterns (page returned 403 on fetch, but content sourced from multiple corroborating results)
- [softprops/action-gh-release releases](https://github.com/softprops/action-gh-release/releases) — v2.6.1 SHA verified
- [dAppServer/wails-build-action action.yml](https://github.com/dAppServer/wails-build-action) — Ubuntu 22.04 vs 24.04 WebKit dependency differences
- [GitHub ARM64 runner announcement](https://github.blog/changelog/2025-01-16-linux-arm64-hosted-runners-now-available-for-free-in-public-repositories-public-preview/) — confirmed free for public repos

### Tertiary (LOW confidence)

- [wailsapp/wails issue #3559](https://github.com/wailsapp/wails/issues/3559) — `-skipbindings` workaround for arm64 cross-compilation (LOW: issue thread, single reporter confirmed fix)
- [wailsapp/wails issue #2833](https://github.com/wailsapp/wails/issues/2833) — fundamental limitation of Wails v2 cross-compilation for arm64 (supports native runner recommendation)

## Metadata

**Confidence breakdown:**
- Standard stack (SHA pins): HIGH — verified directly against GitHub API
- Architecture (workflow structure): HIGH — verified against project's own build scripts
- Wails build commands: HIGH — verified against project's scripts/build-*.sh
- Linux arm64 native runner: MEDIUM — public preview, confirmed available, minor stability unknowns
- Pitfalls: HIGH — several verified against existing bugs in project's build.yml

**Research date:** 2026-03-27
**Valid until:** 2026-06-27 (90 days — runners and action versions are stable; SHA pins should be reverified if > 30 days)
