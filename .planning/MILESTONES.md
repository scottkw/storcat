# Milestones

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
