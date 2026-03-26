# Technology Stack

**Analysis Date:** 2026-03-24

## Languages

**Primary:**
- JavaScript (ES2020+) - Electron main process (`src/main.js`, `src/preload.js`, `src/catalog-service.js`)
- TypeScript (^5.2.2) - Renderer/React UI (`src/renderer/**/*.tsx`, `src/renderer/**/*.ts`)

**Secondary:**
- Bash - Build scripts (`build-all.sh`, `build-complete.sh`, `build-fast.sh`, `build-simple.sh`) and legacy CLI tool (`sdcat/sdcat.sh`, `sdcat/sdcat_zsh.sh`)

## Runtime

**Environment:**
- Electron ^35.0.0 (Chromium + Node.js)
- Node.js (required by Electron; no `.nvmrc` or `.node-version` detected)
- V8 flags applied at startup: `--jitless --no-opt` (ARM64 memory workaround in `src/main.js` line 4)

**Package Manager:**
- npm (no `pnpm-lock.yaml` or `yarn.lock` detected)
- Lockfile: `package-lock.json` present

## Frameworks

**Core:**
- Electron ^35.0.0 - Desktop application shell, IPC between main/renderer
- React ^18.2.0 - Renderer UI framework
- Ant Design (antd) ^5.12.8 - UI component library (Layout, ConfigProvider, theme, Tabs, Table, Modal, etc.)

**Testing:**
- Not detected - No test framework, test config, or test files found

**Build/Dev:**
- Vite ^5.0.8 - Bundler for renderer process (`vite.config.ts`)
- @vitejs/plugin-react ^4.2.1 - React Fast Refresh for Vite
- electron-builder ^24.9.1 - Application packaging and distribution
- concurrently ^8.2.2 - Run Vite dev server and Electron simultaneously
- wait-on ^7.2.0 - Wait for Vite dev server before launching Electron

## Key Dependencies

**Critical (4 production dependencies):**
- `react` ^18.2.0 - UI rendering
- `react-dom` ^18.2.0 - DOM binding for React
- `antd` ^5.12.8 - Complete UI component library (tables, modals, forms, tabs, layout)
- `@ant-design/icons` ^5.2.6 - Icon set for Ant Design components

**Dev Dependencies (10):**
- `electron` ^35.0.0 - Runtime (dev dependency because electron-builder bundles it)
- `electron-builder` ^24.9.1 - Cross-platform packaging (DMG, MSI, AppImage, deb, rpm, snap)
- `typescript` ^5.2.2 - Type checking for renderer code
- `vite` ^5.0.8 - Dev server and bundler
- `@vitejs/plugin-react` ^4.2.1 - React plugin for Vite
- `concurrently` ^8.2.2 - Parallel process runner
- `wait-on` ^7.2.0 - HTTP wait utility
- `@types/react` ^18.2.45 - React type definitions
- `@types/react-dom` ^18.2.18 - ReactDOM type definitions
- `@typescript-eslint/eslint-plugin` ^6.14.0 - ESLint rules for TypeScript
- `@typescript-eslint/parser` ^6.14.0 - TypeScript parser for ESLint
- `eslint` ^8.55.0 - Linting
- `eslint-plugin-react-hooks` ^4.6.0 - React hooks linting
- `eslint-plugin-react-refresh` ^0.4.5 - React Refresh linting

## Configuration

**TypeScript:**
- `tsconfig.json` - Renderer code only (`include: ["src/renderer"]`), target ES2020, strict mode, react-jsx, bundler module resolution
- `tsconfig.node.json` - Vite config only (`include: ["vite.config.ts"]`)

**Vite:**
- `vite.config.ts` - Root set to `src/renderer`, builds to `../../dist`, dev server on port 5173

**Electron Builder:**
- Configured inline in `package.json` under `"build"` key
- App ID: `com.kenscott.storcat`
- macOS: Universal DMG, hardened runtime, code signing (identity `S2K7P43927`), notarization enabled
- Windows: Portable + MSI for x64 and arm64
- Linux: AppImage, deb, rpm, snap for x64 and arm64
- macOS entitlements: `build/entitlements.mac.plist` (allow-unsigned-executable-memory, disable-library-validation, allow-jit)

**Build:**
- `build/` - Icons and macOS entitlements
- `build/icons/` - App icons for all platforms

**Linting:**
- ESLint ^8.55.0 with TypeScript and React plugins installed, but no `.eslintrc*` config file detected at project root

## Platform Requirements

**Development:**
- Node.js (version not pinned)
- npm
- macOS recommended (build scripts assume macOS for cross-platform builds)

**Production Targets:**
- macOS: Universal binary (x64 + arm64) as DMG
- Windows: x64 and arm64 as Portable (.exe) and MSI
- Linux: x64 and arm64 as AppImage, deb, rpm, snap

## NPM Scripts

| Command | Purpose |
|---------|---------|
| `npm start` | Launch Electron directly (production mode) |
| `npm run dev` | Start Vite dev server + Electron with hot reload |
| `npm run build` | Build renderer + package with electron-builder |
| `npm run build:mac` | Build macOS package only |
| `npm run build:win` | Build Windows package only |
| `npm run build:linux` | Build Linux package only |
| `npm run build:all` | Build all platforms via electron-builder |
| `npm run build:complete` | Full build via `build-complete.sh` |
| `npm run version:patch-build` | Bump patch version + complete build |

---

*Stack analysis: 2026-03-24*
