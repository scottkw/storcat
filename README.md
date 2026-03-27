# StorCat v2.2.1

**Storage Media Cataloging Tool**

StorCat is a cross-platform desktop and CLI application for creating, browsing, and searching directory catalogs. It generates JSON and HTML representations of directory trees. Built with Go and React, StorCat provides a fast, native experience across all major platforms.

Run `storcat` with no arguments for the GUI, or use CLI subcommands (`storcat create`, `storcat search`, etc.) for scripting and terminal workflows.

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

- **Create Catalogs**: Scan any directory and create searchable catalogs
  - Recursive directory scanning with symlink following
  - File metadata capture (name, size, dates, creation time)
  - JSON and HTML output formats
  - v1.0 catalog backward compatibility

- **Advanced Search**: Fast search across all your catalogs
  - Search by filename patterns
  - Filter by multiple catalogs
  - Sort by any column
  - Per-column filtering

- **Browse Catalogs**: View catalog metadata
  - See all available catalogs with file size and dates
  - View catalog statistics
  - Open HTML catalog viewer

- **Modern UI**: React-based interface with Ant Design
  - Responsive table with sticky headers
  - Column resizing, sorting, and filtering
  - 11 themes (StorCat Light/Dark, Dracula, Nord, Solarized, One Dark, Monokai, GitHub, Gruvbox)
  - Collapsible sidebar with configurable positioning

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

StorCat is a single binary for both GUI and CLI. After installing:

```bash
# macOS — symlink the app binary
sudo ln -sf /Applications/StorCat.app/Contents/MacOS/StorCat /usr/local/bin/storcat

# Windows — the installer adds StorCat to PATH
# Linux — the .deb installs to /usr/bin/StorCat
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

1. Click the "Create Catalog" tab
2. Click "Select Directory" and choose the folder to catalog
3. Enter a title and description
4. Click "Create Catalog"
5. The catalog is saved to your configured catalog directory

### Searching Catalogs

1. Click the "Search Catalogs" tab
2. Enter search terms (wildcards supported: `*.mp3`, `photo*`)
3. Select which catalogs to search (or leave empty for all)
4. Results show in a sortable, filterable table
5. Click column headers to sort
6. Use column filters for refined search

### Browsing Catalogs

1. Click the "Browse Catalogs" tab
2. View all available catalogs with metadata (size, dates)
3. Click a catalog to open its HTML view
4. Sort and filter the catalog list

## Configuration

StorCat stores configuration in:
- **macOS**: `~/Library/Application Support/storcat/config.json`
- **Windows**: `%APPDATA%\storcat\config.json`
- **Linux**: `~/.config/storcat/config.json`

Default catalog directory:
- **macOS/Linux**: `~/StorCat/catalogs`
- **Windows**: `%USERPROFILE%\StorCat\catalogs`

### Window State Persistence

StorCat remembers window size and position across restarts. This can be toggled in Settings.

## Architecture

### Technology Stack

**Backend (Go 1.23):**
- **Wails v2**: Desktop app framework with native webview
- **Standard Library**: File I/O, JSON encoding, file walking
- **djherbis/times**: Cross-platform file creation time
- **fatih/color**: Colorized CLI output
- **tablewriter**: Tabular CLI output formatting
- **pkg/browser**: Cross-platform browser launch

**Frontend (React 18 + TypeScript 5):**
- **React 18**: UI framework
- **TypeScript 5**: Type safety
- **Vite 5**: Build tool and dev server
- **Ant Design 5**: UI component library
- **Custom components**: ModernTable with advanced features

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
│   ├── catalog/           # Catalog creation service
│   ├── search/            # Search service
│   └── config/            # Configuration management
├── pkg/
│   └── models/            # Shared data models
├── frontend/              # React frontend
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── contexts/      # AppContext state management
│   │   ├── themes.ts      # 11 theme definitions
│   │   └── App.tsx        # Main application
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

## Migration from v1.x (Electron)

For those upgrading from StorCat v1.x (Electron):

### What Changed
- **Backend**: Complete rewrite from Node.js to Go — all catalog operations are native Go
- **Framework**: Electron replaced with Wails v2 — uses system webview instead of bundled Chromium
- **CLI**: Full command-line interface in the same binary (v2.1.0) — replaces the legacy `sdcat` bash scripts
- **Table**: Sticky headers with per-column filtering, sorting, and resizing work correctly
- **API**: Direct Wails function calls instead of Electron IPC, with `{success,...}` envelope pattern
- **Performance**: 93% smaller, 80% faster startup, 5x faster search

### What Stayed the Same
- Catalog file format (v1.x catalogs work in v2.x — both JSON formats supported)
- Search functionality
- UI/UX design (React + Ant Design)
- 11 themes
- Configuration structure

### Compatibility

StorCat v2.x can read catalogs created by any v1.x version. No migration needed.

## License

Copyright © 2024-2026 Ken Scott

## Links

- **GitHub**: https://github.com/scottkw/storcat
- **Issues**: https://github.com/scottkw/storcat/issues
- **Wails**: https://wails.io
- **Releases**: https://github.com/scottkw/storcat/releases

## Acknowledgments

- Built with [Wails](https://wails.io) - Go + Web for Desktop Apps
- UI components from [Ant Design](https://ant.design)
- Icons from [Ant Design Icons](https://ant.design/components/icon)

---

**StorCat v2.2.1** - Fast, Native, Cross-Platform Storage Media Cataloging (Desktop + CLI)
