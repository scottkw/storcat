# Requirements: StorCat v2.0.0

**Defined:** 2026-03-24
**Core Value:** Full feature parity with Electron v1.2.3 — no regressions for users upgrading to Go/Wails.

## v1 Requirements

Requirements for the v2.0.0 release. Each maps to roadmap phases.

### Data Models

- [x] **DATA-01**: Catalog JSON output uses bare object format `{...}`, matching Electron (not array-wrapped `[{...}]`)
- [x] **DATA-02**: Empty directory `contents` field serializes as `[]`, never `null` or omitted
- [x] **DATA-03**: Browse catalog metadata includes `size` field (file size in bytes)
- [x] **DATA-04**: Browse catalog `modified` field is a Date-compatible value, not an opaque string
- [x] **DATA-05**: Browse catalog `created` field uses actual creation time where available, not mtime

### Catalog Operations

- [x] **CATL-01**: `LoadCatalog` Go method exists and returns parsed catalog data for a given file path
- [x] **CATL-02**: `CreateCatalog` returns result metadata: `jsonPath`, `htmlPath`, `fileCount`, `totalSize`, `copyJsonPath`, `copyHtmlPath`
- [x] **CATL-03**: Directory traversal follows symlinks (matching Electron's `fs.stat` behavior)
- [x] **CATL-04**: HTML catalog root node renders with `└──` connector and size bracket, matching Electron format

### API Surface

- [x] **API-01**: `GetCatalogHtmlPath` verifies file existence and returns via wailsAPI wrapper as `{success, htmlPath}`
- [x] **API-02**: `ReadHtmlFile` returns via wailsAPI wrapper as `{success, content}`
- [x] **API-03**: All wailsAPI wrapper methods consistently return `{success, ...}` envelopes matching Electron's contract

### Window Management

- [x] **WIN-01**: Window size (width, height) persists across app restarts via Go config
- [x] **WIN-02**: Window position (x, y) persists across app restarts (size-only on macOS if coordinate drift unresolved)
- [x] **WIN-03**: Settings toggle enables/disables window state persistence (not a stub)
- [x] **WIN-04**: Window state restores via `OnDomReady` hook (not `OnStartup`)

### Platform Integration

- [x] **PLAT-01**: macOS header is draggable as window titlebar using `--wails-draggable: drag`
- [x] **PLAT-02**: App version is derived at build time (not hardcoded constant), displayed correctly in UI

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
| DATA-01 | Phase 1 | Complete |
| DATA-02 | Phase 1 | Complete |
| DATA-03 | Phase 1 | Complete |
| DATA-04 | Phase 1 | Complete |
| DATA-05 | Phase 1 | Complete |
| CATL-01 | Phase 2 | Complete |
| CATL-02 | Phase 1 | Complete |
| CATL-03 | Phase 1 | Complete |
| CATL-04 | Phase 1 | Complete |
| API-01 | Phase 4 | Complete |
| API-02 | Phase 4 | Complete |
| API-03 | Phase 4 | Complete |
| WIN-01 | Phase 3 | Complete |
| WIN-02 | Phase 3 | Complete |
| WIN-03 | Phase 3 | Complete |
| WIN-04 | Phase 4 | Complete |
| PLAT-01 | Phase 6 | Complete |
| PLAT-02 | Phase 6 | Complete |
| REL-01 | Phase 7 | Pending |
| REL-02 | Phase 7 | Pending |

**Coverage:**
- v1 requirements: 20 total
- Mapped to phases: 20
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-24*
*Last updated: 2026-03-24 after roadmap creation*
