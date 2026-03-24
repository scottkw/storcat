# Requirements: StorCat v2.0.0

**Defined:** 2026-03-24
**Core Value:** Full feature parity with Electron v1.2.3 — no regressions for users upgrading to Go/Wails.

## v1 Requirements

Requirements for the v2.0.0 release. Each maps to roadmap phases.

### Data Models

- [ ] **DATA-01**: Catalog JSON output uses bare object format `{...}`, matching Electron (not array-wrapped `[{...}]`)
- [ ] **DATA-02**: Empty directory `contents` field serializes as `[]`, never `null` or omitted
- [ ] **DATA-03**: Browse catalog metadata includes `size` field (file size in bytes)
- [ ] **DATA-04**: Browse catalog `modified` field is a Date-compatible value, not an opaque string
- [ ] **DATA-05**: Browse catalog `created` field uses actual creation time where available, not mtime

### Catalog Operations

- [ ] **CATL-01**: `LoadCatalog` Go method exists and returns parsed catalog data for a given file path
- [ ] **CATL-02**: `CreateCatalog` returns result metadata: `jsonPath`, `htmlPath`, `fileCount`, `totalSize`, `copyJsonPath`, `copyHtmlPath`
- [ ] **CATL-03**: Directory traversal follows symlinks (matching Electron's `fs.stat` behavior)
- [ ] **CATL-04**: HTML catalog root node renders with `└──` connector and size bracket, matching Electron format

### API Surface

- [ ] **API-01**: `GetCatalogHtmlPath` verifies file existence and returns via wailsAPI wrapper as `{success, htmlPath}`
- [ ] **API-02**: `ReadHtmlFile` returns via wailsAPI wrapper as `{success, content}`
- [ ] **API-03**: All wailsAPI wrapper methods consistently return `{success, ...}` envelopes matching Electron's contract

### Window Management

- [ ] **WIN-01**: Window size (width, height) persists across app restarts via Go config
- [ ] **WIN-02**: Window position (x, y) persists across app restarts (size-only on macOS if coordinate drift unresolved)
- [ ] **WIN-03**: Settings toggle enables/disables window state persistence (not a stub)
- [ ] **WIN-04**: Window state restores via `OnDomReady` hook (not `OnStartup`)

### Platform Integration

- [ ] **PLAT-01**: macOS header is draggable as window titlebar using `--wails-draggable: drag`
- [ ] **PLAT-02**: App version is derived at build time (not hardcoded constant), displayed correctly in UI

### Merge & Release

- [ ] **REL-01**: Clean branch merges to main with no bloat (no node_modules, binaries, or archive)
- [ ] **REL-02**: `.gitignore` covers `node_modules/`, `build/bin/`, `.DS_Store`, `dist/`

## v2 Requirements

Deferred to future milestones. Tracked but not in current roadmap.

### Testing

- **TEST-01**: Unit tests for catalog-service pure functions (Go)
- **TEST-02**: Unit tests for AppContext reducer
- **TEST-03**: Integration tests for IPC round-trips

### Enhancements

- **ENH-01**: macOS code signing and notarization
- **ENH-02**: Tailwind CSS migration (if desired)
- **ENH-03**: Progress indicators for long-running catalog operations
- **ENH-04**: Cancel mechanism for catalog creation

## Out of Scope

| Feature | Reason |
|---------|--------|
| New features beyond Electron parity | Future milestones — this is strictly a migration release |
| Automated test suite | Important but separate milestone (TEST-01 through TEST-03 deferred) |
| Performance benchmarking | Go is already faster; no formal benchmarks needed for release |
| Electron fallback/dual-build | Clean break — Go/Wails replaces Electron entirely |
| Wails v3 migration | v3 is in alpha with breaking changes; premature |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| DATA-01 | — | Pending |
| DATA-02 | — | Pending |
| DATA-03 | — | Pending |
| DATA-04 | — | Pending |
| DATA-05 | — | Pending |
| CATL-01 | — | Pending |
| CATL-02 | — | Pending |
| CATL-03 | — | Pending |
| CATL-04 | — | Pending |
| API-01 | — | Pending |
| API-02 | — | Pending |
| API-03 | — | Pending |
| WIN-01 | — | Pending |
| WIN-02 | — | Pending |
| WIN-03 | — | Pending |
| WIN-04 | — | Pending |
| PLAT-01 | — | Pending |
| PLAT-02 | — | Pending |
| REL-01 | — | Pending |
| REL-02 | — | Pending |

**Coverage:**
- v1 requirements: 20 total
- Mapped to phases: 0
- Unmapped: 20 ⚠️

---
*Requirements defined: 2026-03-24*
*Last updated: 2026-03-24 after initial definition*
