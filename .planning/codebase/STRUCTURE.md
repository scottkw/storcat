# Directory & File Structure

## Directory Tree

```
storcat/
├── build/                      # Electron builder resources
│   ├── entitlements.mac.*      # macOS code signing entitlements
│   └── icons/                  # App icons (icns, ico, iconset, favicon, PNG sizes)
├── examples/                   # Sample catalog files
│   ├── sd01.json               # Example catalog JSON
│   ├── sd01.html               # Example catalog HTML view
│   └── sdcat.sh                # Example shell script
├── release_notes/              # Versioned release notes (1.1.0 → 1.2.3)
├── screenshots/                # App screenshots (light-mode, dark-mode × 3 each)
├── sdcat/                      # Shell-based catalog tool (standalone CLI companion)
│   ├── sdcat.sh                # Bash version
│   ├── sdcat_zsh.sh            # Zsh version
│   └── README.md               # CLI tool docs
├── src/                        # Application source code
│   ├── main.js                 # Electron main process (window mgmt, IPC handlers)
│   ├── preload.js              # Electron preload (context bridge)
│   ├── catalog-service.js      # Core catalog logic (create, search)
│   ├── app.js                  # LEGACY: vanilla JS UI (~1,300 lines, unused by React)
│   ├── index.html              # LEGACY: vanilla HTML UI (unused by React)
│   ├── favicon.ico             # Legacy favicon
│   ├── assets/icons/           # Runtime app icon
│   └── renderer/               # React/TypeScript frontend
│       ├── main.tsx            # React entry point
│       ├── index.html          # Vite HTML entry
│       ├── index.css           # Global styles
│       ├── App.tsx             # Root component (theme provider, layout)
│       ├── themes.ts           # 11 theme definitions (light/dark variants)
│       ├── contexts/
│       │   └── AppContext.tsx   # Global state (useReducer-based)
│       ├── components/
│       │   ├── Header.tsx      # App header with nav
│       │   ├── MainContent.tsx # Tab container
│       │   ├── ModernTable.tsx # Reusable data table
│       │   ├── CatalogModal.tsx# Catalog HTML viewer modal
│       │   ├── WelcomeContent.tsx # Welcome/landing page
│       │   └── tabs/
│       │       ├── BrowseCatalogsTab.tsx  # Browse catalog files
│       │       ├── SearchCatalogsTab.tsx  # Search across catalogs
│       │       └── CreateCatalogTab.tsx   # Create new catalogs
│       ├── types/
│       │   └── electron.d.ts   # TypeScript declarations for IPC bridge
│       └── public/             # Static assets for Vite
│           ├── favicon.ico
│           └── storcat-icon.png
├── wiki_pages/                 # Documentation/wiki content
│   ├── home.md
│   └── homebrew-upgrade-issue.md
├── package.json                # Project config (v1.2.3, Electron + React + Ant Design)
├── tsconfig.json               # TypeScript config
├── tsconfig.node.json          # TypeScript config for Node/Vite
├── vite.config.ts              # Vite bundler config
├── build-*.sh                  # Build scripts (all, fast, simple, complete)
├── demo.js                     # Manual demo script for catalog-service
├── test-compatibility.js       # Manual compatibility test script
├── storcat-project.json        # Project metadata (large, ~680KB)
├── storcat-project.html        # Project HTML view (~550KB)
├── storcat-repo-consolidation.md # Repo consolidation notes
├── LICENSE                     # MIT
└── README.md                   # Project documentation
```

## Entry Points

| Entry Point | Purpose |
|-------------|---------|
| `src/main.js` | Electron main process — app lifecycle, window creation, IPC handlers |
| `src/preload.js` | Context bridge — exposes IPC methods to renderer |
| `src/renderer/main.tsx` | React app bootstrap |
| `src/renderer/index.html` | Vite HTML entry (development) |

## Build Outputs

- `dist/` — Vite build output (renderer assets)
- `dist/` — electron-builder output (platform installers: DMG, MSI, AppImage, deb, rpm, snap)

## Configuration Files

| File | Purpose |
|------|---------|
| `package.json` | Dependencies, scripts, electron-builder config |
| `vite.config.ts` | Vite bundler settings |
| `tsconfig.json` | TypeScript compiler options |
| `tsconfig.node.json` | TypeScript for Vite/Node context |
| `build/entitlements.mac.plist` | macOS sandbox entitlements |

## Notable Observations

- **Legacy code still present**: `src/app.js` (1,300+ lines) and `src/index.html` are the original vanilla JS UI, now replaced by the React renderer but not yet removed
- **Mixed JS/TS**: Main process files are plain JavaScript; renderer is TypeScript
- **Large metadata files**: `storcat-project.json` (680KB) and `storcat-project.html` (550KB) are checked into the repo
