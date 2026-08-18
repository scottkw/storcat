# StorCat v3.0.0

**Storage Media Cataloging Tool**

StorCat is a cross-platform desktop and CLI application for creating, browsing, and searching directory catalogs. It generates JSON and HTML representations of directory trees. Built with Go and React, StorCat provides a fast, native experience across all major platforms.

Run `storcat` with no arguments for the GUI, or use CLI subcommands (`storcat create`, `storcat search`, etc.) for scripting and terminal workflows.

> **Note:** `main` is at v3.0.0. The latest *published* release is v2.3.0, which still ships the
> older three-tab interface — the workspace described below arrives with the v3.0.0 release.
> Build from source (see [Building from Source](#building-from-source)) to run it today.

## What's New in v3.0.0 — The Workspace Redesign

The three-tab interface is gone. v3.0.0 replaces it with a single-view workspace and adds the
backend capabilities that design implies:

- **One view, not three tabs** — toolbar, catalog rail, file tree, details panel, and status bar are
  visible together. No more losing your place by switching tabs.
- **Handles big catalogs** — the tree is virtualized and has been exercised against a 42,550-node
  catalog; row count stays flat regardless of scroll position.
- **⌘K command palette** — find any file across every catalog without leaving the workspace. Reads
  live from disk, so it never serves a stale index after a rename or delete.
- **Live scan progress** — creating a catalog shows real counts and bytes as it walks, can be
  cancelled, backgrounded, and recovers a partial catalog if the source volume disappears.
- **Settings that save as you go** — theme, row density, rail position, and catalog defaults,
  persisted to the Go config rather than browser storage.
- **Catalog management** — rename, duplicate, and delete to Trash (never a permanent delete), plus a
  filesystem watcher that keeps the rail current when catalogs change outside the app.
- **Re-scan & diff** — re-scan a catalog's source volume and see exactly what changed across five
  states (added / removed / changed / unreadable / unchanged), then resolve by overwriting, keeping
  both, or discarding. Nothing is written until you choose.

**On the diff's safety:** an unreadable subtree is a distinct state, never collapsed into "removed".
A directory you simply can't read today would otherwise look identical to one that was deleted — and
overwriting on that assumption would discard data that is still there. Writes reuse the same
crash-safe atomic-write path used everywhere else; there is no second write path.

## Why StorCat v2.0.0? The Migration from Electron to Go/Wails

StorCat v2.0.0 represents a complete architectural overhaul, migrating from Electron to Go with the Wails framework. This decision was driven by several key factors:

### The Problem with Electron

The original Electron-based version (v1.2.3) had significant limitations:
- **Large bundle size**: ~150-200MB+ due to bundling Chromium
- **High memory usage**: Each instance loaded a full browser engine
- **Slow startup**: Cold start times of 3-5 seconds
- **V8/ARM64 issues**: Required `--jitless --no-opt` workarounds on Apple Silicon

### The Solution: Go + Wails

Go with the Wails framework provides the best of both worlds:

**Performance Benefits:**
- **93% smaller**: Apps are 8-11MB vs 150-200MB (Electron)
- **80% faster startup**: Native webview loads instantly
- **Native memory footprint**: Uses system webview instead of bundling Chromium
- **5x faster search**: Go concurrency for file scanning and searching

**Development Benefits:**
- **Preserved React UI**: All existing React components work with minimal changes
- **Type-safe bindings**: Auto-generated TypeScript interfaces for Go functions
- **True table headers**: Modern table with sticky headers, per-column filtering, sorting, and resizing
- **Simpler architecture**: Direct function calls via Wails bindings instead of IPC

**Cross-Platform:**
- **macOS**: Uses WebKit (native) — Universal binary (Intel + Apple Silicon)
- **Windows**: Uses WebView2 (native) — x64 and arm64
- **Linux**: Uses WebKitGTK — x64 and arm64

## Features

### GUI (Desktop Application)

- **Workspace**: One view — toolbar, catalog rail, file tree, details panel, status bar
  - Catalog rail with live filter, file counts, byte totals, and a status dot per catalog
  - Virtualized tree with breadcrumbs, expand-all / collapse-all, exercised at 42,550 nodes
  - Details panel follows your selection; reveal in file manager, open the HTML view
  - Responsive across three width tiers — the details panel becomes a drawer on narrow windows

- **Create Catalogs**: Scan any volume or folder into a searchable catalog
  - Detected volumes offered as selectable cards with live size and readability status
  - Live scan progress with real counts and bytes — no fake percentages
  - Cancel mid-scan (writes nothing), or send it to the background and keep working
  - Partial-catalog recovery when a source volume disappears mid-scan
  - Recursive scanning with symlink following, JSON and HTML output, v1.0 backward compatibility

- **⌘K Command Palette**: Find any file across every catalog
  - Cross-catalog filename search, capped and reported honestly ("first 50 of N")
  - Full keyboard navigation; jumping to a hit switches catalog, expands the path, and selects it
  - Reads live from disk — no cached index to go stale after a rename, delete, or re-scan

- **Catalog Management**: Rename, duplicate, delete
  - Rename writes the title into the JSON and both HTML title sites
  - Duplicate with automatic `-copy` / `-copy-2` collision handling
  - Delete moves to the OS Trash — there is no permanent-delete path anywhere
  - Filesystem watcher keeps the rail current when catalogs change outside the app

- **Re-scan & Diff**: Reconcile a catalog against its source volume
  - Five diff states with counts: added, removed, changed, unreadable, unchanged
  - Always asks which volume to scan — nothing is remembered or pre-selected
  - Warns when the scan looks like a different disc, without blocking you
  - Resolve by overwrite, keep-both, or discard; nothing is written until you choose

- **Settings**: Theme, density, rail position, and catalog defaults, saved as you go
  - 11 themes (StorCat Light/Dark, Dracula, Solarized Dark/Light, Nord, One Dark, Monokai,
    GitHub Light/Dark, Gruvbox Dark)
  - Comfortable / compact row density; rail on the left or right
  - Self-hosted fonts, no network calls

### CLI (Command Line)

The same binary provides full CLI access — no separate install needed:

| Command | Description |
|---------|-------------|
| `storcat create <dir>` | Create a catalog from a directory |
| `storcat search <term>` | Search catalogs for a filename pattern |
| `storcat list [dir]` | List catalogs with metadata |
| `storcat show <catalog>` | Display a catalog's tree structure |
| `storcat open <catalog>` | Open a catalog's HTML in the default browser |
| `storcat version` | Print the version |

**CLI features:**
- `--json` flag for machine-readable output (create, search, list, show)
- `--depth N` flag to limit tree depth (show)
- Colorized tree output with `--no-color` / `NO_COLOR` support
- Cross-platform browser launch (macOS, Windows, Linux)
- Standard exit codes (0 = success, 1 = error, 2 = usage)

## Installation

### macOS (Homebrew)

```bash
brew tap scottkw/storcat
brew install --cask storcat
```

Or download the DMG from the [Releases](https://github.com/scottkw/storcat/releases) page, open it, and drag StorCat to Applications.

### Windows (WinGet)

```powershell
winget install scottkw.StorCat
```

Or download the installer from the [Releases](https://github.com/scottkw/storcat/releases) page and run it.

### Linux

**Debian/Ubuntu (.deb):**
```bash
# Download the .deb for your architecture from the Releases page
sudo dpkg -i storcat_*.deb
# Install WebKitGTK dependency if needed
sudo apt-get install -f
```

**AppImage (x64):**
```bash
# Download the AppImage from the Releases page
chmod +x StorCat-*.AppImage
./StorCat-*.AppImage
```

Requires GTK3 and WebKitGTK:
```bash
# Debian/Ubuntu
sudo apt-get install libgtk-3-0 libwebkit2gtk-4.0-37

# Fedora
sudo dnf install gtk3 webkit2gtk3

# Arch
sudo pacman -S gtk3 webkit2gtk
```

### Download Pre-built Binaries

All installers and raw binaries are available on the [Releases](https://github.com/scottkw/storcat/releases) page:

| Platform | Installer | Raw Binary |
|----------|-----------|------------|
| **macOS** (Universal) | `.dmg` | `StorCat.app` in `.tar.gz` |
| **Windows** (x64) | NSIS `.exe` installer | `StorCat.exe` |
| **Linux** (x64) | `.deb`, `.AppImage` | `StorCat-linux-amd64` |
| **Linux** (arm64) | `.deb` | `StorCat-linux-arm64` |

### CLI Access

StorCat is a single binary for both GUI and CLI. Package manager installs put `storcat` on PATH automatically:

```bash
# macOS (Homebrew) — storcat is on PATH immediately after install
brew tap scottkw/storcat
brew install --cask storcat
storcat version  # works in any new terminal

# Windows (WinGet) — NSIS installer adds storcat to PATH
winget install scottkw.StorCat
storcat version  # works in any new terminal

# Linux — the .deb installs to /usr/bin/StorCat
```

If you installed from a DMG or raw binary instead:
```bash
# macOS — symlink the app binary
sudo ln -sf /Applications/StorCat.app/Contents/MacOS/StorCat /usr/local/bin/storcat
```

## Building from Source

### Prerequisites

**Required:**
- Go 1.23 or later
- Node.js 16+ and npm
- Wails CLI v2

**Platform-Specific:**

**macOS:**
```bash
xcode-select --install
```

**Linux:**
```bash
# Debian/Ubuntu
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel

# Arch
sudo pacman -S gtk3 webkit2gtk
```

**Windows:**
- WebView2 Runtime (pre-installed on Windows 10/11)

### Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Add to PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Quick Build

```bash
# Clone the repository
git clone https://github.com/scottkw/storcat.git
cd storcat

# Install frontend dependencies
cd frontend
npm install
cd ..

# Development mode (hot reload)
wails dev

# Production build (current platform)
wails build

# Build for specific platform
wails build -platform darwin/universal  # macOS Universal
wails build -platform darwin/arm64      # macOS Apple Silicon
wails build -platform darwin/amd64      # macOS Intel
wails build -platform windows/amd64     # Windows 64-bit
```

### Build Scripts

The project includes convenience scripts:

```bash
# Build for current platform
./scripts/build-macos.sh
./scripts/build-windows.sh
./scripts/build-linux.sh

# Build all platforms (where supported)
./scripts/build-all.sh
```

**Docker Build for Linux** (from macOS/Windows):
```bash
# Builds both AMD64 and ARM64 Linux versions
./scripts/build-linux-docker.sh
```

### Build Outputs

Builds are located in `build/bin/`:
- **macOS**: `StorCat.app`
- **Windows**: `StorCat.exe`
- **Linux AMD64**: `StorCat-linux-amd64`
- **Linux ARM64**: `StorCat-linux-arm64`

## Usage

### Creating a Catalog

1. Click **＋ New** in the catalog rail
2. Pick a detected volume, or choose any folder
3. Set the title and filename — a live preview shows exactly what will be written
4. Click **Start scan** (or press `⌘↵`)
5. Watch live progress. You can cancel, or send the scan to the background and keep working

The catalog is saved to your configured catalog directory. If the source disappears mid-scan,
StorCat offers to retry or write what it got as a partial catalog.

### Browsing a Catalog

1. Select a catalog in the rail — the tree loads beside it
2. Expand folders, or use expand-all / collapse-all in the tree header
3. Select any file to see its metadata in the details panel
4. From the details footer: reveal in your file manager, or open the HTML view

### Finding a File

1. Press `⌘K` (`Ctrl+K` on Windows/Linux), or click the toolbar search field
2. Type at least two characters — results stream in across *all* catalogs
3. Arrow keys to move, `↵` to jump: StorCat switches catalog, expands the path, and selects the file

### Managing Catalogs

Select a catalog, then use the `⋯` menu in the details panel:

| Action | What it does |
|--------|--------------|
| Rename catalog… | Rewrites the title in the JSON and both HTML title sites |
| Re-scan volume & diff… | Re-scans the source and shows what changed (see below) |
| Duplicate catalog | Copies to `-copy`, `-copy-2`, … as needed |
| Delete catalog… | Moves to the OS Trash — recoverable, never a permanent delete |

### Re-scanning and Reconciling

1. Choose **Re-scan volume & diff…** from the `⋯` menu, or the details-panel footer
2. Pick the source volume — StorCat always asks, and never pre-selects
3. Watch the scan, then review the diff: added, removed, changed, unreadable, unchanged
4. Resolve:
   - **Overwrite catalog** — replaces it in place. Irreversible; the caption says so
   - **Keep both** — writes a new catalog and leaves the original untouched
   - **Discard scan and close** — writes nothing; the catalog is byte-identical afterwards

If the numbers suggest you inserted the wrong disc, a warning appears — advisory only, it never
disables anything.

### If a Catalog Won't Open

A catalog whose JSON can't be parsed shows a diagnostic with the real error, plus three ways out:
re-scan its source volume, open the `.html` instead, or remove it from the library.

## Configuration

StorCat stores configuration in:
- **macOS**: `~/Library/Application Support/storcat/config.json`
- **Windows**: `%APPDATA%\storcat\config.json`
- **Linux**: `~/.config/storcat/config.json`

Default catalog directory:
- **macOS/Linux**: `~/StorCat/catalogs`
- **Windows**: `%USERPROFILE%\StorCat\catalogs`

As of v3.0.0 all user settings live in this file rather than browser storage. Existing values are
migrated once, non-destructively, on first launch.

### Settings

Open with `⌘,` (`Ctrl+,`), the toolbar gear, or the theme chip. Changes save as you go — there is no
Save button.

| Setting | Options |
|---------|---------|
| Theme | 11 themes, light and dark |
| Row density | Comfortable or compact |
| Rail position | Left or right |
| Catalog directory | Where catalogs are read from and written to |
| Default filename root | Pre-filled when creating a catalog |
| Write HTML alongside JSON | On / off |
| Copy to a secondary location | On / off, with a folder picker |
| Watch catalog directory | On / off — keeps the rail current on external changes |
| Remember window size and position | On / off |

Settings cannot be opened while a scan is running in the foreground; the entry points dim and
explain why.

## Architecture

### Technology Stack

**Backend (Go 1.23):**
- **Wails v2**: Desktop app framework with native webview
- **Standard Library**: File I/O, JSON encoding, file walking
- **djherbis/times**: Cross-platform file creation time
- **fsnotify**: Filesystem watching for the catalog directory
- **wastebasket/v2**: Cross-platform move-to-Trash (never a permanent delete)
- **fatih/color**: Colorized CLI output
- **tablewriter**: Tabular CLI output formatting
- **pkg/browser**: Cross-platform browser launch

**Frontend (React 18 + TypeScript):**
- **React 18**: UI framework
- **TypeScript**: Type safety
- **Vite**: Build tool and dev server
- **@tanstack/react-virtual**: Tree virtualization — flat row count at any catalog size
- **Custom components**: the workspace is hand-built against a CSS custom-property token layer
- **Ant Design**: residual only. As of v3.0.0 no workspace component imports it; it survives as a
  `ConfigProvider` wrapper supplying a light/dark algorithm, and is a removal candidate

### Project Structure

```
storcat/
├── app.go                 # Main Wails app struct and bound methods
├── main.go                # Application entry point (GUI + CLI dispatch)
├── version.go             # Build-time version injection
├── app_test.go            # Go tests
├── main_test.go           # CLI dispatch tests
├── cli/                   # CLI subcommand package
│   ├── cli.go             # Entry point and routing
│   ├── create.go          # storcat create
│   ├── search.go          # storcat search
│   ├── list.go            # storcat list
│   ├── show.go            # storcat show
│   ├── open.go            # storcat open
│   ├── version.go         # storcat version
│   ├── output.go          # Shared output helpers
│   └── *_test.go          # Per-command tests
├── internal/              # Go backend packages
│   ├── catalog/           # Catalog creation, walk, diff, resolve, atomic write
│   ├── search/            # Search service (incl. cross-catalog indexed search)
│   ├── config/            # Configuration management
│   ├── volumes/           # Per-OS volume enumeration (stdlib only)
│   ├── watch/             # Debounced fsnotify catalog-directory watcher
│   ├── osutil/            # Trash, reveal-in-file-manager, open-external
│   └── fixture/           # Test fixture helpers
├── pkg/
│   └── models/            # Shared data models (catalog, diff)
├── frontend/              # React frontend
│   ├── src/
│   │   ├── components/
│   │   │   └── workspace/ # WorkspaceShell, CatalogRail, TreePane, DetailsPanel,
│   │   │       │          # Toolbar, StatusBar, CommandPalette, dialogs
│   │   │       ├── create/   # Create slide-over (volume picker, scanning, error, done)
│   │   │       ├── rescan/   # Re-scan dialog and diff list
│   │   │       ├── palette/  # ⌘K palette internals
│   │   │       └── settings/ # Theme grid, segmented controls, catalog settings
│   │   ├── contexts/      # AppContext state management
│   │   ├── hooks/         # useMediaQuery, useModalBehavior, …
│   │   ├── services/      # wailsAPI.ts binding wrapper
│   │   ├── settingsStore.ts # Write-through settings to the Go config
│   │   ├── themes.ts      # 11 theme definitions
│   │   ├── themeTokens.ts # Theme → CSS custom properties, applied pre-paint
│   │   └── workspace.css  # Workspace layout and component styles
│   ├── wailsjs/           # Auto-generated Wails bindings
│   ├── package.json
│   └── vite.config.ts
├── build/                 # Build assets and outputs
│   ├── appicon.png        # Source icon
│   └── bin/               # Compiled binaries
├── scripts/               # Build and install scripts
└── wails.json             # Wails configuration
```

### How It Works

1. **Unified Binary**: Single binary serves both GUI and CLI modes — `storcat` (no args) launches the GUI, `storcat <command>` runs CLI
2. **Go Backend**: Handles all file I/O, catalog creation, search, and configuration
3. **React Frontend**: Provides the GUI via `wailsAPI.ts` compatibility shim
4. **Wails Bridge**: Auto-generates TypeScript bindings for Go methods
5. **Native Webview**: Renders the React app using the system's webview
6. **IPC Contract**: All API methods return `{success, ...}` envelopes for consistent error handling

## Performance Comparison

| Metric | Electron v1.2.3 | Wails v2.0.0 | Improvement |
|--------|-----------------|--------------|-------------|
| Bundle Size | ~150MB | ~8-11MB | **93% smaller** |
| Memory Usage | ~200MB | ~50MB | **75% less** |
| Startup Time | 3-5 seconds | <1 second | **80% faster** |
| Search Speed | ~500ms | ~100ms | **5x faster** |

## Development

### Live Development

```bash
wails dev
```

This starts:
- Vite dev server (hot reload for frontend)
- Go application (hot reload for backend)
- Browser dev mode at `http://localhost:34115`

### Generate Bindings

After modifying Go methods:
```bash
wails generate module
```

This regenerates TypeScript bindings in `frontend/wailsjs/`.

### Project Commands

```bash
# Check dependencies
wails doctor

# Update dependencies
cd frontend && npm update && cd ..
go get -u ./...
go mod tidy

# Clean build
wails build -clean
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test on your target platform(s)
5. Submit a pull request

### Development Guidelines

- Keep Go code in `internal/` packages
- Use TypeScript for all React components
- Follow existing code style
- Add tests for new features
- Update documentation

## Code Signing

**macOS**: StorCat releases are signed with Developer ID, notarized by Apple, and stapled. Gatekeeper accepts the DMG without prompting.

**Windows**: Authenticode signing pipeline is built and ready — will be activated once signing credentials are provisioned.

## Releases

StorCat uses [release-please](https://github.com/googleapis/release-please) for automated releases. Conventional commits on `main` automatically maintain a release PR. Merging the PR creates a git tag, builds all platforms, publishes the release, and distributes to Homebrew and WinGet — fully automated.

See [docs/release-pipeline.md](docs/release-pipeline.md) for the full pipeline documentation, code signing details, secrets configuration, and manual trigger instructions.

## Migration from v1.x (Electron)

For those upgrading from StorCat v1.x (Electron):

### What Changed
- **Backend**: Complete rewrite from Node.js to Go — all catalog operations are native Go
- **Framework**: Electron replaced with Wails v2 — uses system webview instead of bundled Chromium
- **CLI**: Full command-line interface in the same binary (v2.1.0) — replaces the legacy `sdcat` bash scripts
- **Code Signing**: macOS releases are Developer ID signed and notarized (v2.3.0)
- **Release Automation**: Conventional commits drive version bumps, builds, and distribution (v2.3.0)
- **API**: Direct Wails function calls instead of Electron IPC, with `{success,...}` envelope pattern
- **Performance**: 93% smaller, 80% faster startup, 5x faster search
- **Interface**: the three-tab layout was replaced entirely by the single-view workspace (v3.0.0)

### What Changed in v3.0.0
- **Interface**: three tabs → one workspace (rail, tree, details, toolbar, status bar)
- **Tree**: virtualized, so large catalogs stay responsive
- **Search**: moved into a ⌘K palette that spans every catalog, reading live from disk
- **New capabilities**: catalog rename / duplicate / delete-to-Trash, a filesystem watcher,
  and re-scan & diff
- **Settings**: moved from browser `localStorage` into the Go config, with a one-time
  non-destructive migration of existing values
- **UI library**: Ant Design components replaced by hand-built components over a CSS
  custom-property token layer

### What Stayed the Same
- Catalog file format — v1.x catalogs still load, both JSON formats supported
- The CLI and its output contracts
- 11 themes
- Configuration structure and locations

### Compatibility

StorCat v3.x reads catalogs created by any v1.x or v2.x version. No migration needed, and the CLI
is unchanged — existing scripts keep working.

## License

Copyright © 2024-2026 Ken Scott

## Links

- **GitHub**: https://github.com/scottkw/storcat
- **Issues**: https://github.com/scottkw/storcat/issues
- **Wails**: https://wails.io
- **Releases**: https://github.com/scottkw/storcat/releases

## Acknowledgments

- Built with [Wails](https://wails.io) - Go + Web for Desktop Apps
- Tree virtualization by [TanStack Virtual](https://tanstack.com/virtual)
- Typography: [IBM Plex](https://www.ibm.com/plex/), self-hosted
- [Ant Design](https://ant.design) powered the pre-v3.0.0 interface

---

**StorCat v3.0.0** - Fast, Native, Cross-Platform Storage Media Cataloging (Desktop + CLI)
