# Release Pipeline

StorCat uses a fully automated release pipeline. The only manual step is merging a pull request.

## How It Works

```
Conventional commits on main
        │
        ▼
release-please maintains a Release PR (version bump + CHANGELOG)
        │
        ▼ (you merge the PR)
release-please creates git tag + DRAFT GitHub release
        │
        ▼ (tag push triggers release.yml)
4 parallel build jobs: macOS, Windows, Linux x64, Linux arm64
        │
        ▼
Artifacts uploaded to draft release, then published (draft → false)
        │
        ▼ (published event triggers distribute.yml)
Homebrew cask updated + WinGet manifests committed + WinGet PR submitted
```

## Workflows

### `release-please.yml`

**Trigger:** Push to `main`

Runs [release-please](https://github.com/googleapis/release-please) to analyze conventional commits since the last release. If there are releasable changes, it creates or updates a PR that:

- Bumps the version in `wails.json` (single source of truth)
- Updates `.release-please-manifest.json`
- Generates/updates `CHANGELOG.md`

When you merge this PR, release-please creates a git tag (e.g., `v2.4.0`) and a **draft** GitHub release.

### `release.yml`

**Trigger:** Tag push matching `v*.*.*`, or manual `workflow_dispatch`

Builds StorCat on 4 platform runners in parallel:

| Job | Runner | Outputs |
|-----|--------|---------|
| `build-macos` | `macos-14` | Universal DMG (signed + notarized + stapled), `.tar.gz` |
| `build-windows` | `windows-2022` | NSIS installer, portable `.exe` (signing conditional on secrets) |
| `build-linux-amd64` | `ubuntu-22.04` | AppImage, `.deb`, `.tar.gz` |
| `build-linux-arm64` | `ubuntu-22.04` (QEMU) | `.deb`, `.tar.gz` |

After all builds complete, the `release` job:
1. Downloads all artifacts
2. Uploads them to the GitHub release
3. Publishes the release (`draft: false`)

Publishing fires the `release: published` event that triggers distribution.

### `distribute.yml`

**Trigger:** Release published, or manual `workflow_dispatch`

Three parallel jobs:

| Job | What it does |
|-----|-------------|
| `update-homebrew` | Downloads DMG, computes SHA256, updates `scottkw/homebrew-storcat` tap |
| `update-winget` | Submits a PR to `microsoft/winget-pkgs` via `winget-releaser` |
| `update-winget-manifests` | Commits versioned WinGet manifests to `packaging/winget/` in this repo |

### `build.yml`

**Trigger:** Push to `main`, pull requests

CI-only build that verifies the code compiles on macOS, Windows, and Linux. Does not produce release artifacts.

## Code Signing

### macOS (Fully Automated)

Every macOS DMG is:
1. **Signed** with Developer ID Application certificate (`codesign --options runtime`)
2. **Notarized** by Apple via `xcrun notarytool`
3. **Stapled** with notarization ticket via `xcrun stapler`
4. **Verified** with `spctl --assess` as a CI gate

Entitlements: `build/darwin/entitlements.plist` (hardened runtime for Wails WebKit).

### Windows (Pipeline Ready, Awaiting Credentials)

The signing pipeline is built and will activate when eSigner credentials are stored:
1. SSL.com eSigner cloud HSM signs both `.exe` and installer
2. `signtool verify /pa /v` confirms valid Authenticode signature
3. Signing occurs before `upload-artifact` so WinGet SHA256 hashes match

The signing steps are skipped automatically when `ES_USERNAME` secret is not configured.

## GitHub Secrets

All signing secrets are stored in the `release` environment with `v*.*.*` tag protection.

### Apple (6 secrets — configured)

| Secret | Purpose |
|--------|---------|
| `APPLE_CERTIFICATE` | Base64-encoded Developer ID .p12 |
| `APPLE_CERTIFICATE_PASSWORD` | .p12 passphrase |
| `APPLE_CERTIFICATE_NAME` | Certificate common name for `codesign --sign` |
| `APPLE_ID` | Apple ID for notarization |
| `APPLE_ID_PASSWORD` | App-specific password for notarization |
| `APPLE_TEAM_ID` | Developer team ID for notarization |

### Windows (4 secrets — not yet configured)

| Secret | Purpose |
|--------|---------|
| `ES_USERNAME` | SSL.com eSigner username |
| `ES_PASSWORD` | SSL.com eSigner password |
| `CREDENTIAL_ID` | eSigner credential ID |
| `ES_TOTP_SECRET` | TOTP seed for eSigner 2FA |

### Distribution (2 secrets)

| Secret | Purpose |
|--------|---------|
| `HOMEBREW_TAP_TOKEN` | PAT with repo access to `scottkw/homebrew-storcat` |
| `WINGET_TOKEN` | PAT for `winget-releaser` to submit PRs to `microsoft/winget-pkgs` |

## Conventional Commits

release-please uses [Conventional Commits](https://www.conventionalcommits.org/) to determine version bumps:

| Prefix | Version Bump | Example |
|--------|-------------|---------|
| `feat:` | Minor (2.3.0 → 2.4.0) | `feat: add export to CSV` |
| `fix:` | Patch (2.3.0 → 2.3.1) | `fix: search results not sorting` |
| `feat!:` or `BREAKING CHANGE:` | Major (2.3.0 → 3.0.0) | `feat!: change catalog format` |
| `chore:`, `docs:`, `ci:` | No release | `docs: update README` |

## Version Source of Truth

`wails.json` field `info.productVersion` is the single source of truth. release-please updates it via `extra-files` jsonpath config. The Go binary reads it at build time via `//go:embed wails.json`.

## Manual Triggers

Both release and distribute workflows support `workflow_dispatch` for manual re-runs:

```bash
# Re-run release builds for a tag
gh workflow run release.yml --ref v2.3.0 -f tag=v2.3.0

# Re-run distribution for a published release
gh workflow run distribute.yml -f tag=v2.3.0
```

## Credential Rotation

See `docs/runbooks/credential-rotation.md` for procedures when certificates expire.

## Repository Settings

The following GitHub repo settings are required:

- **Settings → Actions → General → Workflow permissions**: "Read and write permissions"
- **Settings → Actions → General**: "Allow GitHub Actions to create and approve pull requests" (checked)
- **Settings → Environments → release**: Tag protection pattern `v*.*.*`
