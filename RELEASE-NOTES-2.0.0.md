# StorCat v2.0.0 Release Notes

**Release Date:** October 16, 2025

StorCat v2.0.0 represents a complete architectural overhaul, migrating from Electron to Go with the Wails framework. This major release brings dramatic improvements in performance, size, and user experience.

---

## What's New in v2.0.0

### Complete Architectural Rewrite

StorCat has been completely rebuilt from the ground up using Go and the Wails framework, replacing the Electron-based architecture. This change brings massive improvements across every metric:

**Performance Improvements:**
- **93% smaller application size**: 8-11MB vs 150-200MB (Electron)
- **75% less memory usage**: ~50MB vs ~200MB runtime footprint
- **80% faster startup time**: Sub-second startup vs 3-5 seconds
- **5x faster search**: ~100ms vs ~500ms for typical searches
- **Instant UI responsiveness**: Native webview rendering

**Technology Stack:**
- **Backend**: Go 1.21+ with Wails v2.10.2
- **Frontend**: React 18 + TypeScript + Vite
- **UI Library**: Ant Design 5.27+
- **Native Webviews**:
  - macOS: WebKit (native)
  - Windows: WebView2 (native)
  - Linux: WebKitGTK

---

## Major Features

### 1. Fixed Table Component
The persistent table rendering issues from v1.0 are **completely resolved**:
- ✅ Sticky headers that actually work
- ✅ Per-column filtering and sorting
- ✅ Column resizing and reordering
- ✅ Smooth scrolling with large datasets
- ✅ Proper keyboard navigation

### 2. Improved Catalog Creation
- **Concurrent scanning**: Uses Go goroutines for lightning-fast directory traversal
- **Progress tracking**: Real-time feedback during catalog creation
- **Better error handling**: Clear error messages and recovery
- **Secondary location support**: Copy catalogs to backup locations

### 3. Enhanced Search Experience
- **Multi-catalog search**: Search across multiple catalogs simultaneously
- **Wildcard support**: Use `*.mp3`, `photo*`, etc.
- **Column-level filtering**: Filter by any column in the results
- **Instant results**: Fast in-memory search with Go
- **Sortable results**: Click any column header to sort

### 4. Modern UI Updates
- **Cleaner design**: Updated interface with Ant Design 5
- **Better icons**: Comprehensive icon set from Ant Design Icons
- **Responsive layout**: Better handling of different window sizes
- **Dark theme support**: System theme detection and dark mode
- **Accessibility improvements**: Better keyboard navigation and screen reader support

### 5. Cross-Platform Builds
Pre-built binaries for all major platforms:
- **macOS**: Universal binary (Intel + Apple Silicon)
- **Windows**: 64-bit AMD64
- **Linux**: AMD64 and ARM64

---

## What Changed from v1.0

### Architecture
- ❌ **Removed**: Electron + Node.js IPC
- ✅ **Added**: Go + Wails with native webview
- ✅ **Improved**: Direct function calls instead of IPC messaging
- ✅ **Simplified**: No separate backend process

### Dependencies
- ❌ **Removed**: Electron, all Chromium overhead
- ✅ **Added**: Wails runtime (minimal overhead)
- ✅ **Kept**: React, TypeScript, Ant Design (frontend)

### Features
All v1.0 features are preserved and improved:
- ✅ Catalog creation (faster)
- ✅ Multi-catalog search (faster)
- ✅ Browse catalogs (better UI)
- ✅ Configuration management (same format)

---

## Compatibility

### Backward Compatibility
- ✅ **v1.0 catalogs work in v2.0** - No migration needed!
- ✅ Same catalog JSON format
- ✅ Same search functionality
- ✅ Same configuration structure

### Breaking Changes
- ⚠️ **Application bundle size**: Much smaller, but may require fresh install
- ⚠️ **System requirements**: Now requires native webview components (see Installation)

---

## Installation

### System Requirements

**macOS:**
- macOS 10.13 (High Sierra) or later
- WebKit (pre-installed)
- 20MB disk space

**Windows:**
- Windows 10 or later
- WebView2 Runtime (usually pre-installed on Windows 10/11)
- 25MB disk space

**Linux:**
- GTK3
- WebKitGTK 2.0
- 25MB disk space

### Installation Instructions

**macOS:**
1. Download `StorCat.app` from [Releases](https://github.com/scottkw/storcat/releases/tag/2.0.0)
2. Move to Applications folder
3. Right-click and select "Open" (first time only)

**Windows:**
1. Download `StorCat.exe` from [Releases](https://github.com/scottkw/storcat/releases/tag/2.0.0)
2. Run the executable
3. Windows may show SmartScreen warning (click "More info" → "Run anyway")

**Linux:**
1. Download appropriate binary (`StorCat-linux-amd64` or `StorCat-linux-arm64`)
2. Make executable: `chmod +x StorCat-linux-amd64`
3. Install dependencies:
   ```bash
   # Debian/Ubuntu
   sudo apt-get install libgtk-3-0 libwebkit2gtk-4.0-37

   # Fedora
   sudo dnf install gtk3 webkit2gtk3

   # Arch
   sudo pacman -S gtk3 webkit2gtk
   ```
4. Run: `./StorCat-linux-amd64`

---

## Download Links

### Pre-built Binaries

All binaries are available on the [Releases page](https://github.com/scottkw/storcat/releases/tag/2.0.0):

- **macOS Universal** (Intel + Apple Silicon): `StorCat-2.0.0-darwin-universal.app.zip`
- **Windows 64-bit**: `StorCat-2.0.0-windows-amd64.zip`
- **Linux AMD64**: `StorCat-2.0.0-linux-amd64.tar.gz`
- **Linux ARM64**: `StorCat-2.0.0-linux-arm64.tar.gz`

### Checksums

SHA256 checksums are provided in `checksums.txt` on the releases page.

---

## Known Issues

### macOS
- First launch may show "unverified developer" warning - use right-click → Open to bypass

### Windows
- SmartScreen may flag the executable - this is expected for unsigned applications
- WebView2 Runtime required (usually pre-installed on Windows 10/11)

### Linux
- Requires GTK3 and WebKitGTK to be installed
- Some older distributions may need manual dependency installation

---

## Performance Benchmarks

Tested on macOS 15.0 (M3 MacBook Air, 24GB RAM):

| Operation | v1.0 (Electron) | v2.0 (Wails) | Improvement |
|-----------|-----------------|--------------|-------------|
| Cold start | 3.2 seconds | 0.6 seconds | **5.3x faster** |
| Catalog 10,000 files | 4.5 seconds | 1.8 seconds | **2.5x faster** |
| Search across 5 catalogs | 520ms | 95ms | **5.5x faster** |
| Memory at idle | 195MB | 48MB | **75% less** |
| Application size | 152MB | 8.3MB | **93% smaller** |

---

## Migration Guide

### From v1.0 to v2.0

**No migration needed!** v2.0 is fully compatible with v1.0 catalogs.

1. Install StorCat v2.0
2. Your existing catalogs will be automatically detected
3. All searches will work with no changes needed

**Configuration:**
- Configuration files remain in the same locations
- No changes to catalog directory structure
- All settings are preserved

---

## Building from Source

### Prerequisites
- Go 1.21 or later
- Node.js 16+ and npm
- Wails CLI v2.10.2+
- Platform-specific dependencies (see README.md)

### Quick Build
```bash
git clone https://github.com/scottkw/storcat-wails.git
cd storcat-wails
cd frontend && npm install && cd ..
wails build
```

See [README.md](README.md) for detailed build instructions.

---

## Contributors

**Lead Developer:** Ken Scott

Special thanks to:
- The [Wails](https://wails.io) team for an excellent framework
- The [Ant Design](https://ant.design) team for beautiful UI components
- All beta testers and early adopters

---

## What's Next

Looking ahead to v2.1:
- [ ] Cloud sync for catalogs
- [ ] Catalog comparison feature
- [ ] Export to various formats (CSV, Excel, PDF)
- [ ] Advanced filtering with saved searches
- [ ] Catalog merging and splitting
- [ ] Duplicate file detection across catalogs
- [ ] Tag-based organization

---

## Support

- **Documentation**: [GitHub Wiki](https://github.com/scottkw/storcat/wiki)
- **Issues**: [GitHub Issues](https://github.com/scottkw/storcat/issues)
- **Source Code**: [GitHub Repository](https://github.com/scottkw/storcat)

---

## License

Copyright © 2024 Ken Scott

---

**StorCat v2.0.0** - Fast, Native, Cross-Platform Storage Media Cataloging

Built with [Wails](https://wails.io) • [GitHub](https://github.com/scottkw/storcat) • [Download](https://github.com/scottkw/storcat/releases/tag/2.0.0)
