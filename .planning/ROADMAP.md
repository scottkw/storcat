# Roadmap: StorCat

## Milestones

- ✅ **v2.0.0 Go/Wails Migration** — Phases 1-7 (shipped 2026-03-26) — [Archive](milestones/v2.0.0-ROADMAP.md)
- 🔄 **v2.1.0 CLI Commands** — Phases 8-10 (in progress)

## Phases

<details>
<summary>✅ v2.0.0 Go/Wails Migration (Phases 1-7) — SHIPPED 2026-03-26</summary>

- [x] Phase 1: Data Models + Catalog Service (2/2 plans) — completed 2026-03-24
- [x] Phase 2: Search Service + Browse Metadata (1/1 plans) — completed 2026-03-25
- [x] Phase 3: Config Manager (2/2 plans) — completed 2026-03-25
- [x] Phase 4: App Layer + Lifecycle (2/2 plans) — completed 2026-03-25
- [x] Phase 5: Frontend Shim (1/1 plans) — completed 2026-03-26
- [x] Phase 6: Platform Integration (1/1 plans) — completed 2026-03-26
- [x] Phase 7: Verification + Merge (2/2 plans) — completed 2026-03-26

</details>

**v2.1.0 CLI Commands:**

- [ ] **Phase 8: CLI Foundation and Platform Compatibility** — Dispatch skeleton, platform fixes, version command, install script
- [ ] **Phase 9: Core Subcommands — Create, List, Search** — Three primary scripting commands with table and JSON output
- [ ] **Phase 10: Show, Open, and Output Polish** — Tree visualization, cross-platform browser open, color and depth flags

## Phase Details

### Phase 8: CLI Foundation and Platform Compatibility
**Goal**: Users can invoke `storcat <cmd>` from a terminal on macOS, Windows, and Linux — with correct dispatch, output routing, help text, and exit codes — and the GUI launch path remains unchanged
**Depends on**: Nothing (first v2.1.0 phase; v2.0.0 codebase is the foundation)
**Requirements**: CLIP-01, CLIP-02, CLIP-03, CLIP-04, CLIP-05, CLIP-06, CLCM-06, CLPC-01, CLPC-02, CLPC-03, CLPC-04
**Success Criteria** (what must be TRUE):
  1. Running `storcat` with no arguments launches the GUI exactly as before — no regression
  2. Running `storcat version` prints the version string to stdout and exits 0
  3. Running `storcat --help` or `storcat <cmd> --help` prints usage text and exits 0
  4. Running `storcat` from a Windows terminal produces visible output (not silently discarded by GUI subsystem)
  5. Running `storcat` from macOS Finder on first launch does not crash due to `-psn_*` argument injection, and `wails dev` hot-reload still opens the GUI
**Plans:** 2 plans
Plans:
- [ ] 08-01-PLAN.md — CLI package: dispatch, version command, stub commands, tests
- [ ] 08-02-PLAN.md — main.go integration, psn filtering, install script, Windows build

### Phase 9: Core Subcommands — Create, List, Search
**Goal**: Users can create catalogs, list cataloged directories, and search catalog contents entirely from the command line, with human-readable table output by default and machine-readable JSON via a flag
**Depends on**: Phase 8
**Requirements**: CLCM-01, CLCM-02, CLCM-03, CLOF-01, CLOF-03
**Success Criteria** (what must be TRUE):
  1. User runs `storcat create <dir> --output <dir>` and a JSON catalog and HTML file are created on disk with a success confirmation printed to stdout
  2. User runs `storcat list <dir>` and sees a table of catalogs with name, file count, size, and modified date
  3. User runs `storcat search <term> <dir>` and sees a table of matching files with path and catalog source
  4. User passes `--json` to `list`, `search`, or `create` and gets machine-readable JSON output to stdout, suitable for piping to `jq`
  5. Any command error (missing arg, unreadable directory) prints a message to stderr and exits non-zero
**Plans**: TBD

### Phase 10: Show, Open, and Output Polish
**Goal**: Users can inspect a catalog's tree structure in the terminal and open its HTML report in a browser, and all CLI commands respect color preferences and output depth controls
**Depends on**: Phase 9
**Requirements**: CLCM-04, CLCM-05, CLOF-02, CLOF-04, CLOF-05, CLOF-06, CLPC-05
**Success Criteria** (what must be TRUE):
  1. User runs `storcat show <catalog.json>` and sees the full directory tree printed to stdout as plain text
  2. User passes `--depth N` to `show` and the tree is truncated at depth N, with deeper levels omitted
  3. On a TTY, `show` output renders directories in bold/blue color; passing `--no-color` or setting `NO_COLOR` env var suppresses all color on any command
  4. User runs `storcat open <catalog.json>` and the catalog's HTML file opens in the system default browser on macOS, Linux, and Windows
  5. User passes `--json` to `show` and receives the raw catalog JSON to stdout
**Plans**: TBD

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Data Models + Catalog Service | v2.0.0 | 2/2 | Complete | 2026-03-24 |
| 2. Search Service + Browse Metadata | v2.0.0 | 1/1 | Complete | 2026-03-25 |
| 3. Config Manager | v2.0.0 | 2/2 | Complete | 2026-03-25 |
| 4. App Layer + Lifecycle | v2.0.0 | 2/2 | Complete | 2026-03-25 |
| 5. Frontend Shim | v2.0.0 | 1/1 | Complete | 2026-03-26 |
| 6. Platform Integration | v2.0.0 | 1/1 | Complete | 2026-03-26 |
| 7. Verification + Merge | v2.0.0 | 2/2 | Complete | 2026-03-26 |
| 8. CLI Foundation and Platform Compatibility | v2.1.0 | 0/2 | Not started | - |
| 9. Core Subcommands — Create, List, Search | v2.1.0 | 0/? | Not started | - |
| 10. Show, Open, and Output Polish | v2.1.0 | 0/? | Not started | - |
