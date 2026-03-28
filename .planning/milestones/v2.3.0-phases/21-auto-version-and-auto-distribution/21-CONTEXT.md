# Phase 21: Auto Version and Auto Distribution - Context

**Gathered:** 2026-03-28
**Status:** Ready for planning

<domain>
## Phase Boundary

Automate version bumping, release creation, and distribution so that merging a release PR is the only manual step. No more hand-editing wails.json, no manual tag pushes, no manual publish clicks. The pipeline chain: conventional commits → release-please PR → merge → tag → build → upload → publish → distribute to Homebrew + WinGet.

</domain>

<decisions>
## Implementation Decisions

### Version Bump Strategy
- **D-01:** Use conventional commits (feat:, fix:, breaking:) to auto-determine semver bump
- **D-02:** Use Google's release-please as the versioning/release tool
- **D-03:** release-please auto-generates and maintains CHANGELOG.md from conventional commits

### Release Trigger Flow
- **D-04:** release-please runs on pushes to main branch only (no release branches)
- **D-05:** Merging the release-please PR creates the git tag and GitHub release, triggering the existing release.yml build pipeline
- **D-06:** The splash screen version (CreateCatalogTab.tsx) already reads dynamically from Go backend's GetVersion() which embeds wails.json — no extra work needed, just ensure release-please bumps wails.json

### Version File Sync
- **D-07:** wails.json `info.productVersion` is the single source of truth for version
- **D-08:** release-please updates wails.json only. frontend/package.json version is cosmetic and can be left as-is or pinned

### Draft vs Auto-Publish
- **D-09:** release-please creates the GitHub release (not release.yml). release.yml triggers on tag push, builds all platform artifacts, and uploads them to the existing release
- **D-10:** After artifact upload, release.yml marks the release as published (no longer draft). This triggers distribute.yml automatically
- **D-11:** distribute.yml fires on release published — updates Homebrew tap and WinGet manifests. Full chain is automated: merge PR → tag → build → upload → publish → distribute

### Claude's Discretion
- release-please configuration details (release-type, extra-files, versioning-strategy)
- How release.yml is refactored to upload to existing release instead of creating one
- Whether to use release-please GitHub Action or the standalone release-please CLI
- CHANGELOG.md formatting and sections

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### CI/CD Workflows
- `.github/workflows/release.yml` — Current build + release pipeline (must be modified to upload to existing release instead of creating one)
- `.github/workflows/distribute.yml` — Current distribution pipeline (Homebrew tap + WinGet manifests, triggered on release published)

### Version Source
- `wails.json` — Single source of truth for version (`info.productVersion`, currently 2.2.1)
- `version.go` — Embeds wails.json via `//go:embed`, exposes `Version` variable to Go app
- `frontend/src/services/wailsAPI.ts` — Frontend calls `GetVersion()` for splash screen display
- `frontend/src/components/tabs/CreateCatalogTab.tsx` — Splash screen renders version from backend

### Packaging Templates
- `packaging/homebrew/storcat.rb.template` — Homebrew cask template (uses {{VERSION}} and {{SHA256}})
- `packaging/winget/manifests/` — WinGet manifest templates

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `version.go` — Already embeds wails.json and parses productVersion. release-please just needs to update the JSON field.
- `release.yml` — Mature pipeline with macOS signing/notarization, Windows Authenticode signing, Linux packaging. Needs refactoring from "create release" to "upload to release".
- `distribute.yml` — Already handles Homebrew and WinGet distribution on release publish event. No changes needed if publish event is preserved.

### Established Patterns
- Tag-triggered CI: release.yml fires on `v*.*.*` tags — release-please creates these tags
- Draft-then-publish flow: release.yml creates draft, distribute.yml fires on publish — need to modify so release-please creates the release and release.yml just uploads

### Integration Points
- release-please PR → merge → tag push → release.yml → artifact upload → publish release → distribute.yml
- release-please needs to update `wails.json` `info.productVersion` field (JSON path: `.info.productVersion`)

</code_context>

<specifics>
## Specific Ideas

- User specifically wants the splash screen version to auto-update with each release (already handled by wails.json embed, but important to verify in testing)
- Pipeline must be fully automated after PR merge — no manual publish or distribute steps

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 21-auto-version-and-auto-distribution*
*Context gathered: 2026-03-28*
