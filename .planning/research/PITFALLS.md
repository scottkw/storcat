# Pitfalls Research

**Domain:** CI/CD release pipeline, cross-platform packaging, and repo consolidation for a Go/Wails desktop app
**Project:** StorCat v2.2.0
**Researched:** 2026-03-27
**Confidence:** HIGH (verified against official Wails docs, GitHub Actions docs, Homebrew docs, and WinGet sources)

---

## Critical Pitfalls

---

### Pitfall 1: NSIS Installer Requires a Native Windows Runner — No Cross-Compilation

**What goes wrong:**
`wails build -nsis` silently fails or errors with `ERROR: cannot build nsis installer - no windows targets` when run on macOS or Linux runners, even when `-platform windows/amd64` is specified. You get an `.exe` without an installer, or nothing at all. The failure mode is inconsistent — sometimes it just skips the NSIS step silently.

**Why it happens:**
The Wails NSIS flag requires NSIS to be installed locally AND requires the build to be executing on a Windows host. Cross-compilation of the NSIS installer from Linux/macOS is not supported. The `-nsis` flag is effectively a "run NSIS packager" instruction, not a "build NSIS installer format" instruction.

**How to avoid:**
Always run Windows installer builds (`-nsis`) on `windows-latest` GitHub Actions runner. Use a matrix strategy:
```yaml
strategy:
  matrix:
    include:
      - os: macos-latest
        platform: darwin/universal
      - os: windows-latest
        platform: windows/amd64
        nsis: true
      - os: ubuntu-latest
        platform: linux/amd64
```
On the Windows job, add `-nsis` to the wails build command. On other jobs, omit it.

**Warning signs:**
- Single-runner workflow (e.g., ubuntu-only) building "all platforms"
- NSIS step produces no `.exe_installer` file, only the raw `.exe`
- Workflow "succeeds" but release assets are missing the MSI/NSIS file

**Phase to address:** Phase 1 (CI workflow scaffolding). Lock down runner-per-platform matrix before implementing packaging steps.

---

### Pitfall 2: macOS Universal Binary Cannot Be Built on Non-macOS Runners

**What goes wrong:**
Running `wails build -platform darwin/universal` on a Linux or Windows runner produces a build failure because CGO for Darwin requires Apple's Xcode toolchain. There is no supported cross-compilation path to macOS. Wails will skip the darwin target with a warning: `WARNING Crosscompiling to Mac not currently supported`.

**Why it happens:**
Wails depends on Cocoa/WebKit frameworks that require the macOS SDK. CGO binds to system libraries at build time. GitHub's hosted Linux/Windows runners do not include the macOS SDK.

**How to avoid:**
Always build macOS artifacts on `macos-latest`. The `macos-latest` runner is currently arm64 (Apple Silicon). Building `darwin/universal` on an arm64 runner requires the Intel cross-compilation toolchain — verify the runner image includes it. If universal builds fail on `macos-latest`, pin to `macos-13` (last Intel runner) or use `macos-latest-xlarge` (M1, paid).

Also note: `macos-14` and newer are arm64-only. `macos-13` is the last x86_64 runner. For universal binaries (combining arm64+x86), the runner must be able to compile for both architectures.

**Warning signs:**
- macOS build step on ubuntu/windows runner
- `darwin/universal` flag used without a macOS runner
- Workflow matrix sends all platforms to a single Linux runner

**Phase to address:** Phase 1 (CI workflow scaffolding). Verify runner selection in the matrix before any build work.

---

### Pitfall 3: Linux WebKit Version Mismatch on Ubuntu 24.04+ Runners

**What goes wrong:**
On Ubuntu 24.04 (which `ubuntu-latest` now resolves to as of 2024-2025), `libwebkit2gtk-4.0-dev` is removed. The standard Wails build command fails with `libwebkit not found`. AppImage bundles may also have runtime failures where `WebKitNetworkProcess` cannot be found.

**Why it happens:**
Ubuntu 24.04 ships `libwebkit2gtk-4.1` (GTK4 variant), dropping the GTK3 `4.0` version. Wails v2 defaults to `4.0`. If the wails-build-action or your manual setup does not account for this, the build fails.

**How to avoid:**
Option A — Use `ubuntu-22.04` runner explicitly (still supported, has 4.0):
```yaml
- os: ubuntu-22.04
  platform: linux/amd64
```

Option B — On Ubuntu 24.04, install `libwebkit2gtk-4.1-dev` and add the build tag:
```bash
sudo apt-get install libwebkit2gtk-4.1-dev
wails build -tags webkit2_41
```

Option C — Use the community `wails-build-action` which handles the tag detection automatically on Ubuntu 24.04.

**Warning signs:**
- CI failure mentioning `libwebkit` not found
- `ubuntu-latest` pinned without checking current version (it changed in 2024)
- AppImage builds succeed on CI but fail at runtime with `WebKitNetworkProcess: No such file or directory`

**Phase to address:** Phase 1 (CI workflow scaffolding). Lock Ubuntu version or add webkit2_41 handling before any Linux build work.

---

### Pitfall 4: Parallel Jobs Race-Condition on Release Asset Upload

**What goes wrong:**
When macOS, Windows, and Linux jobs all run in parallel and each tries to create the GitHub release and upload its assets independently, you get "Release already exists" errors on 2 of the 3 jobs. The result is a partial release with some assets missing, no clear error surfaced, and potential draft releases left behind.

**Why it happens:**
GitHub's release creation API is not idempotent across concurrent requests for the same tag. When three runners hit `gh release create v2.2.0` simultaneously, two fail. `softprops/action-gh-release` has retry logic but can still produce duplicate draft releases when the repository has many existing releases.

**How to avoid:**
Use a fan-in pattern:
1. Each platform job runs its build and uploads artifacts using `actions/upload-artifact`
2. A final `release` job depends on all platform jobs (`needs: [build-macos, build-windows, build-linux]`)
3. The release job downloads all artifacts and creates the release once

```yaml
release:
  needs: [build-macos, build-windows, build-linux]
  runs-on: ubuntu-latest
  steps:
    - uses: actions/download-artifact@v4
    - uses: softprops/action-gh-release@v2
      with:
        files: |
          artifacts/**/*
```

**Warning signs:**
- Three separate jobs each using `softprops/action-gh-release` independently
- Release has assets from only 1-2 platforms despite all jobs succeeding
- "Release already exists" or "tag already has a release" in job logs

**Phase to address:** Phase 1 (CI workflow scaffolding). The fan-in pattern must be the baseline design; it cannot be bolted on later.

---

### Pitfall 5: Homebrew SHA256 Computed Before GitHub CDN Propagates the Release Asset

**What goes wrong:**
The Homebrew auto-update step runs immediately after `gh release create` and computes the SHA256 of the downloaded asset. If the GitHub CDN has not fully propagated the new release (which takes 30-120 seconds), `curl` silently returns the previous release's asset or a 404, producing a wrong SHA256. The Homebrew formula is committed with an incorrect checksum, breaking `brew install storcat` for all users.

**Why it happens:**
GitHub's release asset delivery uses a CDN. Immediately after a release is published, some CDN edge nodes may not yet have the new file. The `curl` download succeeds (200 OK) but returns stale content. This is particularly common when both the release and the Homebrew update happen in the same workflow job without a delay.

**How to avoid:**
Add an explicit wait between release publication and checksum computation:
```yaml
- name: Wait for release CDN propagation
  run: sleep 60

- name: Update Homebrew formula
  # Now compute sha256 from the release URL
```

Alternatively, compute the SHA256 locally from the built artifact before uploading to GitHub, then pass it to the Homebrew update step as a workflow variable — no CDN download required.

**Warning signs:**
- Homebrew update step runs in the same job immediately after `gh release create`
- `brew install storcat` fails with SHA256 mismatch shortly after release
- Checksum mismatch reports come from users but your CI shows green

**Phase to address:** Phase 3 (Homebrew automation). Build SHA pre-computation into the release job; never re-download from GitHub to get the hash.

---

### Pitfall 6: WinGet Manifest PRs Require a Classic PAT — Fine-Grained Tokens Break

**What goes wrong:**
WinGet automation (Komac, winget-releaser) requires a Personal Access Token with `public_repo` scope on the `microsoft/winget-pkgs` repository. Fine-grained PATs (the newer, more secure option) cannot create pull requests to external repositories. The action silently fails or produces a cryptic "403 Forbidden" when a fine-grained PAT is used.

**Why it happens:**
The `winget-pkgs` repository uses a GitHub App and PR workflow that requires classic PAT `public_repo` scope. Fine-grained PATs scoped to specific repos do not have cross-repo PR creation rights by design.

**How to avoid:**
Create a classic PAT with `public_repo` scope. Store as `WINGET_TOKEN` repository secret. Use `winget-releaser` or Komac action with this token:
```yaml
- uses: vedantmgoyal9/winget-releaser@latest
  with:
    identifier: KenScott.StorCat
    token: ${{ secrets.WINGET_TOKEN }}
```

Also: the package must already exist in `winget-pkgs` with at least one version before automation can update it. The first submission must be manual.

**Warning signs:**
- Using `${{ secrets.GITHUB_TOKEN }}` for WinGet PRs (this never works)
- Fine-grained PAT scoped only to `storcat` repo used for WinGet step
- `403 Forbidden` or "could not create pull request" in WinGet automation logs

**Phase to address:** Phase 4 (WinGet manifest automation). Set up the classic PAT as a prerequisite before implementing the automation step.

---

### Pitfall 7: Homebrew Tap Update Requires PAT With `contents: write` on the Tap Repo — GITHUB_TOKEN Is Insufficient

**What goes wrong:**
The main repo's `GITHUB_TOKEN` cannot push commits or create PRs to `homebrew-storcat` (a different repository). Workflows that use `GITHUB_TOKEN` for the cross-repo push will fail with "403: Resource not accessible by integration". This is a silent failure in some action versions.

**Why it happens:**
`GITHUB_TOKEN` is scoped to the current repository by design. Cross-repository operations require an explicit PAT or a GitHub App token.

**How to avoid:**
Create a PAT with `contents: write` permission on the `homebrew-storcat` repo. Store as `HOMEBREW_TAP_TOKEN`. Use `peter-evans/create-pull-request` or direct `git push` with this token:
```yaml
- name: Push Homebrew formula update
  env:
    HOMEBREW_REPO_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
  run: |
    git clone https://x-access-token:${HOMEBREW_REPO_TOKEN}@github.com/YOUR_ORG/homebrew-storcat
    # update formula
    cd homebrew-storcat
    git commit -am "storcat v${{ github.ref_name }}"
    git push
```

Note: `repository_dispatch` events for cross-repo workflow triggers also require a PAT; they do not work with `GITHUB_TOKEN` from the triggering repo.

**Warning signs:**
- Using `${{ secrets.GITHUB_TOKEN }}` for cross-repo git push
- Workflow succeeds but homebrew-storcat repo has no new commit
- Error: "Permission to homebrew-storcat.git denied to github-actions[bot]"

**Phase to address:** Phase 3 (Homebrew automation). PAT setup is a prerequisite; document which secrets are needed and where.

---

### Pitfall 8: GitHub Actions Third-Party Action Tag Pinning — Supply Chain Attack Surface

**What goes wrong:**
Using version tags (e.g., `uses: softprops/action-gh-release@v2`) instead of commit SHA pins means a compromised action maintainer account can push malicious code under the existing tag. In March 2025, the `tj-actions/changed-files` action was compromised this way, affecting the Wails project's CI and 23,000+ other repos — secrets were dumped to workflow logs.

**Why it happens:**
Git tags are mutable. Any tag can be force-pushed to point to different code. CI users who pin to a tag believe they have a fixed version but actually get whatever the tag points to at runtime.

**How to avoid:**
Pin all third-party actions to their commit SHA with the version as a comment:
```yaml
- uses: softprops/action-gh-release@c95fe1489396fe8a9eb5b7bd2a584a3dc95ae8b3  # v2.2.1
- uses: actions/upload-artifact@v4  # First-party GitHub actions can use tags
```

First-party GitHub actions (`actions/*`) are lower risk and can use tags. Third-party actions should be SHA-pinned. Use `dependabot` to keep SHA pins updated.

**Warning signs:**
- Workflow uses `@v1`, `@v2`, `@latest` tags for third-party actions
- No `dependabot.yml` configured for GitHub Actions updates
- No SHA pinning for community actions (wails-build-action, winget-releaser, etc.)

**Phase to address:** Phase 1 (CI workflow scaffolding). Pin all SHAs from the start; retrofitting is error-prone.

---

### Pitfall 9: Release Trigger — `on: push: tags` vs `on: release: published` Behavioral Differences

**What goes wrong:**
Using `on: push: tags: ['v*']` as the release trigger runs the workflow the moment a tag is pushed, before any release notes are written. Artifact uploads may race against the auto-created draft release. Using `on: release: published` requires that a GitHub release be manually created (or created by a prior job) before the pipeline runs — which means the tag must be pushed first, then a release created separately. Teams conflate these and get either premature builds or never-triggering workflows.

**Why it happens:**
The distinction between pushing a tag and publishing a release is non-obvious. `push: tags` fires immediately on `git push --tags`. `release: published` fires only when a GitHub Release object is explicitly published (not just created as draft).

**How to avoid:**
For automated release pipelines, use `push: tags` as the canonical trigger. The workflow creates the release itself:
```yaml
on:
  push:
    tags:
      - 'v[0-9]+.[0-9]+.[0-9]+'
```
The build job creates the GitHub release using `softprops/action-gh-release` or `gh release create`. This is the pattern where CI owns the release lifecycle end-to-end.

**Warning signs:**
- `on: release: published` trigger with no prior workflow that creates the release
- Manual process required to "publish" the release after tagging
- Release workflow never fires on `git push --tags`

**Phase to address:** Phase 1 (CI workflow scaffolding). Choose one trigger strategy and document it as the canonical release process.

---

### Pitfall 10: Wails Frontend Build Step Must Run Before `wails build` in CI

**What goes wrong:**
On a fresh CI runner, `wails build` fails with:
```
pattern all:frontend/dist: directory prefix frontend/dist does not exist
```
The `//go:embed all:frontend/dist` directive requires the compiled frontend to exist before the Go build runs. CI runners don't have a cached `frontend/dist`.

**Why it happens:**
`wails build` does run the frontend build internally, but only when using the standard Wails CLI. If the wails-build-action or a custom step calls `wails build` without Node.js/npm available on the PATH, or if Node.js version is wrong, the embedded frontend build silently fails and the dist directory is missing.

**How to avoid:**
Ensure the CI runner has the correct Node.js version before `wails build`:
```yaml
- uses: actions/setup-node@v4
  with:
    node-version: '20'
    cache: 'npm'
    cache-dependency-path: frontend/package-lock.json

- name: Install frontend dependencies
  run: cd frontend && npm ci

- name: Build with Wails
  run: wails build -clean -trimpath
```

**Warning signs:**
- `wails build` step fails with `frontend/dist does not exist`
- No `actions/setup-node` step before the build step
- Workflow uses `npm install` instead of `npm ci` (slower, non-deterministic)

**Phase to address:** Phase 1 (CI workflow scaffolding). Frontend dependency installation must be part of every platform job.

---

## Moderate Pitfalls

---

### Pitfall 11: Draft Release Leaves Artifacts Unreachable During Homebrew Formula Computation

**What goes wrong:**
If the release is initially created as a draft (`draft: true`) while artifacts are being uploaded in parallel, the Homebrew/WinGet update step may run before the release is published. Download URLs for draft releases return 404 for unauthenticated requests. The Homebrew formula update fails silently or with a network error.

**How to avoid:**
Either (a) use the fan-in pattern where the release is published in a single job after all artifacts are ready, or (b) create the release as published immediately and upload artifacts to it. Never trigger downstream automation off a draft release.

**Warning signs:**
- `draft: true` in release creation step
- Homebrew/WinGet steps in the same workflow without `needs:` dependency on the release job
- Formula SHA computation step returns 404 errors

**Phase to address:** Phase 3 (Homebrew automation) and Phase 4 (WinGet automation). Ensure release is published before either runs.

---

### Pitfall 12: WinGet Package Identifier Must Already Exist Before Automation Works

**What goes wrong:**
`winget-releaser` and Komac both require at least one existing version in `microsoft/winget-pkgs` to use as a template for the new version's manifests. Running the automation on the very first release (v2.2.0 as the first winget submission) will fail because there is no base manifest to update.

**How to avoid:**
The initial WinGet submission must be done manually by submitting a PR to `microsoft/winget-pkgs` with the three manifest files (version, installer, locale). Only after that PR is merged can automation update future versions. The identifier `Publisher.AppName` (e.g., `KenScott.StorCat`) must match exactly across manual and automated submissions.

**Warning signs:**
- WinGet automation configured before any manual submission
- Identifier in automation config doesn't match the manually submitted identifier
- "Package not found" error in Komac/winget-releaser logs

**Phase to address:** Phase 4 (WinGet automation). Manual submission is a prerequisite gate.

---

### Pitfall 13: Go Embed Version Injection vs ldflags — CI Build Behavior Differs

**What goes wrong:**
StorCat uses `//go:embed wails.json` for version injection. If `wails.json` is updated manually before the CI build but the tag does not match the version in `wails.json`, the binary reports the wrong version. This is less of a problem with ldflags (which can be set at build time) but `//go:embed` reads a file — if that file is wrong, the version is wrong.

**How to avoid:**
Add a CI step that validates `wails.json` version matches the pushed git tag before building:
```yaml
- name: Verify version matches tag
  run: |
    TAG_VERSION="${GITHUB_REF_NAME#v}"
    WAILS_VERSION=$(jq -r '.info.productVersion' wails.json)
    if [ "$TAG_VERSION" != "$WAILS_VERSION" ]; then
      echo "Version mismatch: tag=$TAG_VERSION, wails.json=$WAILS_VERSION"
      exit 1
    fi
```

**Warning signs:**
- No version validation step in release workflow
- `storcat version` output doesn't match the GitHub release tag
- `wails.json` updated by hand before tagging

**Phase to address:** Phase 1 or 2. Add version validation as an early gate in the workflow.

---

### Pitfall 14: macOS Runner Version Changes Break Universal Binary Builds

**What goes wrong:**
`macos-latest` silently changes to a new OS version (e.g., from macOS 14 arm64 to macOS 26 arm64) when GitHub updates runner images. If the universal binary build depends on specific Xcode version, SDK, or Go CGO flags, the build may break silently or produce a non-universal binary without warning.

**How to avoid:**
Pin to a specific macOS runner version (e.g., `macos-14`) instead of `macos-latest`. Review [actions/runner-images](https://github.com/actions/runner-images) changelog before upgrading. Use `lipo -info` in CI to verify the produced binary is truly universal:
```yaml
- name: Verify universal binary
  run: lipo -info build/bin/storcat.app/Contents/MacOS/storcat
  # Should output: Architectures in the fat file: ... are: x86_64 arm64
```

**Warning signs:**
- `macos-latest` used without pinning
- No `lipo -info` verification step
- Users report "bad CPU type" on Intel Macs after a release

**Phase to address:** Phase 1 (CI workflow scaffolding). Pin runner version and add lipo verification.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Single workflow job for all platforms | Simple YAML | macOS/Windows/Linux build constraints prevent this from working; all three need native runners | Never |
| Using `@latest` or `@v1` tags for community actions | Automatic updates | Supply chain attack surface; March 2025 incident showed real risk | Never for community actions; OK for `actions/*` first-party |
| Computing SHA256 from GitHub CDN download in Homebrew step | No local artifact storage needed | CDN propagation delay causes stale hash; breaks `brew install` for users | Never; always compute SHA locally before upload |
| Hardcoded WinGet identifier | Simple config | If identifier changes, all automation breaks silently | Only if identifier is stable from day one and documented |
| Manual release creation before tagging | Familiar workflow | Breaks `on: push: tags` trigger; requires two-step human process | Never in an automated pipeline |
| Skipping version-tag validation | Faster release | Binary ships with wrong version; mismatch confuses users and Homebrew | Never |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Wails + GitHub Actions | Single Linux runner for all platforms | Matrix with `macos-latest`, `windows-latest`, `ubuntu-22.04` (or ubuntu-latest + webkit2_41 tag) |
| NSIS installer | Cross-compiling from macOS/Linux | Must run on `windows-latest`; no cross-platform NSIS support |
| Homebrew tap update | Using `GITHUB_TOKEN` for cross-repo push | Classic PAT with `contents: write` on tap repo stored as separate secret |
| WinGet automation | Fine-grained PAT or `GITHUB_TOKEN` | Classic PAT with `public_repo` scope; first manual submission required |
| Release creation | Multiple jobs each calling `gh release create` | Fan-in pattern: one final job creates the release after all platform builds succeed |
| Homebrew SHA256 | Re-downloading artifact from GitHub CDN | Compute SHA256 locally during build, pass as workflow output variable |
| `macos-latest` runner | Assuming runner arch is stable | Pin to specific version (e.g., `macos-14`); verify universal binary with `lipo -info` |
| `ubuntu-latest` runner | Assuming webkit2gtk-4.0 is present | Check current ubuntu-latest version; add webkit2_41 tag or pin to ubuntu-22.04 |

---

## "Looks Done But Isn't" Checklist

- [ ] **Multi-platform matrix:** Workflow uses three separate runners (macOS, Windows, Linux) — not one runner building all platforms.
- [ ] **NSIS on Windows runner only:** `-nsis` flag only appears in the Windows job; not in macOS or Linux jobs.
- [ ] **Universal binary verified:** `lipo -info` confirms `x86_64 arm64` in the macOS artifact.
- [ ] **Fan-in release job:** Release creation happens in a single job that `needs:` all platform build jobs.
- [ ] **SHA256 computed locally:** Homebrew formula SHA is computed before upload, not re-downloaded from CDN.
- [ ] **Homebrew PAT configured:** `HOMEBREW_TAP_TOKEN` secret exists with `contents: write` on tap repo.
- [ ] **WinGet PAT configured:** Classic PAT with `public_repo` scope stored as `WINGET_TOKEN`; first manual submission completed.
- [ ] **Version tag validated:** Workflow checks that the git tag version matches `wails.json` version before building.
- [ ] **Action SHAs pinned:** All third-party actions use commit SHA pins, not mutable tags.
- [ ] **Linux webkit handled:** Linux build specifies `ubuntu-22.04` or adds `-tags webkit2_41` for Ubuntu 24.04+.
- [ ] **Node.js setup step present:** Every platform job has `actions/setup-node` before `wails build`.
- [ ] **`on: push: tags` trigger:** Release workflow fires on tag push, not `on: release: published` requiring manual publish.
- [ ] **AppImage runtime test:** Linux AppImage launches on target distro; WebKitNetworkProcess found at runtime.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Wrong SHA256 in Homebrew formula | MEDIUM | Push corrected formula to tap repo; users who already ran `brew install` unaffected; new installs fix automatically |
| NSIS installer missing from release | LOW | Re-run build workflow on `windows-latest` with `-nsis` flag; upload asset to existing release manually |
| Release assets incomplete (race condition) | LOW | Delete draft releases; re-run workflow with fan-in pattern |
| WinGet PR fails (token issue) | LOW | Create/rotate classic PAT; re-run automation job |
| Universal binary is arm64-only | MEDIUM | Re-build on runner that can produce fat binaries; re-upload macOS asset; user impact: Intel Mac users get "bad CPU type" error |
| Homebrew tap not updated after release | LOW | Run Homebrew update job manually via `workflow_dispatch`; push formula fix |
| Linux AppImage missing WebKit at runtime | HIGH | Must rebuild with `linuxdeploy` bundling webkit libs; not a trivial fix; consider distributing `.deb` as primary Linux artifact instead |
| Compromised third-party action (supply chain) | HIGH | Rotate all secrets immediately; audit workflow logs; re-pin all action SHAs |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| NSIS requires Windows runner | Phase 1: CI scaffold | Windows job has `-nsis`; macOS/Linux jobs do not |
| macOS cross-compile impossible | Phase 1: CI scaffold | Matrix has `macos-latest` runner for Darwin targets |
| Linux webkit 4.0 missing on Ubuntu 24.04 | Phase 1: CI scaffold | Linux build succeeds; runner version pinned |
| Parallel release upload race condition | Phase 1: CI scaffold | Fan-in job pattern with `needs:` verified |
| Homebrew SHA256 CDN delay | Phase 3: Homebrew automation | SHA computed locally; no CDN download in formula update step |
| WinGet classic PAT required | Phase 4: WinGet automation | Classic PAT configured; manual submission confirmed as prerequisite |
| Homebrew cross-repo GITHUB_TOKEN limitation | Phase 3: Homebrew automation | `HOMEBREW_TAP_TOKEN` secret configured; push succeeds |
| Third-party action supply chain | Phase 1: CI scaffold | All community actions use commit SHA pins |
| `on: push tags` vs `on: release published` | Phase 1: CI scaffold | Single trigger strategy documented; workflow fires on `git push --tags` |
| Frontend not built before wails build | Phase 1: CI scaffold | `setup-node` + `npm ci` precede `wails build` in every job |
| Version tag / wails.json mismatch | Phase 2: Build pipeline | Validation step exits non-zero on mismatch before build begins |
| macOS runner version drift | Phase 1 + ongoing | Runner version pinned; `lipo -info` step in CI output |
| WinGet package not yet in registry | Phase 4 pre-work | Manual submission PR merged to `winget-pkgs` before automation is enabled |

---

## Sources

- Wails cross-platform build guide: https://wails.io/docs/guides/crossplatform-build/
- Wails NSIS installer guide: https://wails.io/docs/guides/windows-installer/
- Wails Ubuntu 24.04 webkit issue #3581: https://github.com/wailsapp/wails/issues/3581
- Wails NSIS cross-compile error #1714: https://github.com/wailsapp/wails/issues/1714
- Wails security incident response March 2025: https://wails.io/blog/security-incident-response-march-2025/
- tj-actions/changed-files supply chain attack (CVE-2025-30066): https://github.com/advisories/ghsa-mrrh-fwg8-r2c3
- GitHub Actions SHA pinning guide (StepSecurity): https://www.stepsecurity.io/blog/pinning-github-actions-for-enhanced-security-a-complete-guide
- softprops/action-gh-release race condition #602: https://github.com/softprops/action-gh-release/issues/602
- Homebrew auto-update with GitHub Actions (BuiltFast): https://builtfast.dev/blog/automating-homebrew-tap-updates-with-github-actions/
- Homebrew SHA256 mismatch discussion: https://github.com/orgs/Homebrew/discussions/1312
- WinGet Releaser action (fine-grained PAT limitation): https://github.com/vedantmgoyal9/winget-releaser
- Komac (WinGet manifest creator): https://github.com/russellbanks/Komac
- GitHub Actions cross-repo workflow PAT requirements: https://medium.com/hostspaceng/triggering-workflows-in-another-repository-with-github-actions-4f581f8e0ceb
- macOS 26 runner (arm64) now GA: https://github.blog/changelog/2026-02-26-macos-26-is-now-generally-available-for-github-hosted-runners/
- AppImage WebKitNetworkProcess runtime issue #4313: https://github.com/wailsapp/wails/issues/4313

---
*Pitfalls research for: CI/CD release pipeline, cross-platform packaging, and repo consolidation for Go/Wails desktop app*
*Researched: 2026-03-27*
