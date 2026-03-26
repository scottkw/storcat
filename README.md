# StorCat v2.0.0

**Storage Media Cataloging Tool**

StorCat is a powerful desktop application for cataloging storage media contents (CDs, DVDs, USB drives, external hard drives, etc.) and searching through them later. Built with Go and React, StorCat provides a fast, native experience across all major platforms.

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

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/scottkw/storcat/releases) page:

- **macOS**: `StorCat.app` (Universal: Intel + Apple Silicon)
- **Windows**: `StorCat.exe` (64-bit)
- **Linux**: `StorCat-linux-amd64` or `StorCat-linux-arm64`

### macOS Installation

1. Download `StorCat.app`
2. Move to Applications folder
3. Right-click and select "Open" (first time only, to bypass Gatekeeper)

### Windows Installation

1. Download `StorCat.exe`
2. Run the executable
3. Windows may show SmartScreen warning (click "More info" → "Run anyway")
4. Ensure WebView2 Runtime is installed (usually pre-installed on Windows 10/11)

### Linux Installation

1. Download `StorCat-linux-amd64` (or `arm64` for ARM systems)
2. Make executable: `chmod +x StorCat-linux-amd64`
3. Run: `./StorCat-linux-amd64`
4. Requires GTK3 and WebKitGTK:
   ```bash
   # Debian/Ubuntu
   sudo apt-get install libgtk-3-0 libwebkit2gtk-4.0-37

   # Fedora
   sudo dnf install gtk3 webkit2gtk3

   # Arch
   sudo pacman -S gtk3 webkit2gtk
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
├── main.go                # Application entry point
├── version.go             # Build-time version injection
├── app_test.go            # Go tests
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
├── scripts/               # Build scripts
└── wails.json             # Wails configuration
```

### How It Works

1. **Go Backend**: Handles all file I/O, catalog creation, search, and configuration
2. **React Frontend**: Provides the UI via `wailsAPI.ts` compatibility shim
3. **Wails Bridge**: Auto-generates TypeScript bindings for Go methods
4. **Native Webview**: Renders the React app using the system's webview
5. **IPC Contract**: All API methods return `{success, ...}` envelopes for consistent error handling

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
- **Table**: Sticky headers with per-column filtering, sorting, and resizing work correctly
- **API**: Direct Wails function calls instead of Electron IPC, with `{success,...}` envelope pattern
- **Performance**: 93% smaller, 80% faster startup, 5x faster search

### What Stayed the Same
- Catalog file format (v1.x catalogs work in v2.0.0 — both JSON formats supported)
- Search functionality
- UI/UX design (React + Ant Design)
- 11 themes
- Configuration structure

### Compatibility

StorCat v2.0.0 can read catalogs created by any v1.x version. No migration needed.

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

**StorCat v2.0.0** - Fast, Native, Cross-Platform Storage Media Cataloging
