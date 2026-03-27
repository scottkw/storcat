# Milestones

## v2.2.0 Repo Consolidation & CI/CD (Shipped: 2026-03-27)

**Phases completed:** 4 phases, 7 plans, 11 tasks

**Key accomplishments:**

- Consolidated WinGet manifests and Homebrew cask template into main repo under `packaging/`
- Archived `winget-storcat` satellite repo; marked `homebrew-storcat` as auto-managed
- Tag-triggered `release.yml` with 4 parallel platform builds (macOS universal, Windows, Linux x64+arm64) and fan-in draft release
- Platform packaging: macOS DMG (create-dmg), Windows NSIS installer, Linux AppImage + .deb
- `distribute.yml` auto-updates Homebrew cask and submits WinGet PR on release publish

---

## v2.1.0 CLI Commands (Shipped: 2026-03-26)

**Phases completed:** 4 phases, 7 plans, 9 tasks

**Key accomplishments:**

- stdlib flag.FlagSet CLI dispatch package with Run() entry point, version command, and 5 stub handlers — zero external dependencies, full --help/exit-code/stdout-stderr contract
- storcat list command with table/JSON output using tablewriter v1.1.4, shared printJSON/formatBytes helpers in cli/output.go
- storcat search and storcat create commands with table/JSON output, flag wiring, and comprehensive tests
- storcat show command with colorized tree rendering (fatih/color), --depth N truncation, and --json flag
- storcat open command with cross-platform browser launch (pkg/browser) and HTML path derivation
- Tech debt cleanup — closed all v2.1.0 audit gaps: NO_COLOR test, stale import, help stream consistency, orphaned export

---

## v2.0.0 Go/Wails Migration (Shipped: 2026-03-26)

**Phases completed:** 7 phases, 11 plans, 15 tasks

**Key accomplishments:**

- Go data models and catalog service match Electron format exactly — JSON bare object, empty dir `[]`, HTML tree with `└──` connectors, v1 backward compatibility
- Search and browse metadata with LoadCatalog dual-format parsing, browse Size column with human-readable formatting, RFC3339 dates
- Full window state persistence — size + position save/restore via Go config lifecycle hooks, settings toggle wired end-to-end
- All 17 wailsAPI wrappers return `{success,...}` envelopes matching Electron's contract — all 5 consumer components updated
- macOS header draggable via `--wails-draggable` CSS; version injected at build time via ldflags + GetVersion bound method
- Three-tab smoke test approved on macOS, feature branch merged to main — no bloat, proper .gitignore, CI aligned

---
