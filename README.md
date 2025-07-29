# StorCat - Directory Catalog Manager

<div align="center">
  <img src="https://raw.githubusercontent.com/scottkw/storcat/main/build/icons/storcat-logo.png" alt="StorCat Logo" width="128" height="128">
  
  **A modern Electron application for creating and searching directory catalogs**
  
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Electron](https://img.shields.io/badge/Electron-28.0.0-blue.svg)](https://electronjs.org/)
  [![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg)](https://github.com/electron/electron)
</div>

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Screenshots](#screenshots)
- [System Requirements](#system-requirements)
- [Installation](#installation)
- [Build Environment Setup](#build-environment-setup)
- [Building the Application](#building-the-application)
- [Usage Guide](#usage-guide)
  - [Getting Started](#getting-started)
  - [Configuring Application Settings](#configuring-application-settings)
  - [Creating Directory Catalogs](#creating-directory-catalogs)
  - [Searching Catalogs](#searching-catalogs)
  - [Browsing Catalog Collections](#browsing-catalog-collections)
  - [Interface Features](#interface-features)
- [File Compatibility](#file-compatibility)
- [Development](#development)
- [Project Structure](#project-structure)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Overview

StorCat is a modern Electron-based directory catalog management application that provides a clean, intuitive GUI interface for creating, searching, and browsing directory catalogs. It maintains 100% compatibility with catalog files created by the original `sdcat.sh` bash script while offering enhanced functionality through a modern React-based interface.

The application generates both JSON and HTML representations of directory structures, making it easy to document, search, and share information about file systems across different platforms.

### 🆕 Recent Major Updates
- **Complete React Refactor**: Modernized from HTMX to React 18 with TypeScript
- **Enhanced UI Components**: Professional table components with filtering, sorting, and pagination
- **Comprehensive Theme System**: 11 beautiful themes including Dracula, Nord, Solarized, GitHub, and more
- **Flexible Interface Layout**: Configurable sidebar positioning (left/right) with smart icon placement
- **Better User Experience**: Modal catalog viewer, branded title bar, and redesigned settings panel

## ✨ Features

### Core Functionality
- **📁 Create Catalogs**: Generate comprehensive JSON and HTML catalogs of any directory with detailed file size information
- **🔍 Search Catalogs**: Fast, full-text search across multiple catalog files with advanced filtering
- **📂 Browse Catalogs**: Interactive catalog browser with modern table interface
- **📊 Professional Tables**: Modern table component with sticky headers, column resizing, filtering, sorting, and pagination
- **🖼️ Modal Catalog Viewer**: Full-screen modal for viewing HTML catalogs with automatic theme adaptation

### User Interface
- **🎨 Modern React UI**: Clean, responsive interface built with React and Ant Design components
- **🏠 Branded Title Bar**: StorCat logo in the title bar with theme-aware styling
- **⚙️ Enhanced Settings**: Redesigned settings panel with comprehensive customization options
- **🌈 Advanced Theme System**: 11 beautiful themes including StorCat Light/Dark, Dracula, Nord, Solarized Light/Dark, One Dark, Monokai, GitHub Light/Dark, and Gruvbox Dark
- **📐 Flexible Layout**: Configurable sidebar positioning (left or right) with intelligent icon placement
- **📱 Responsive Design**: Intelligent sidebar management with persistent icon bar when collapsed
- **⚡ Real-time Updates**: Live search results and progress indicators
- **💾 Persistent Preferences**: All settings automatically save and restore between sessions
- **🎯 Dynamic Theming**: Real-time theme switching with CSS variables for instant updates

### Technical Features
- **⚛️ Modern React Architecture**: Built with React 18, TypeScript, and modern development practices
- **🔄 Cross-Platform**: Works seamlessly on Windows, macOS, and Linux
- **🔒 Secure**: Uses Electron's context isolation and secure preload scripts
- **⚙️ No Dependencies**: No external system dependencies or command-line tools required
- **📋 100% Compatible**: Works seamlessly with existing bash script catalog files
- **🎨 Component-Based**: Modular React components with Ant Design integration
- **⚡ Vite-Powered**: Fast development and build process with Vite bundler

## 📸 Screenshots

### Light Mode
<div align="center">
  <img src="https://raw.githubusercontent.com/scottkw/storcat/main/screenshots/light-mode-1.png" alt="StorCat Light Mode - Create Catalog" width="800">
  <p><em>Create Catalog Tab - Welcome Screen</em></p>
</div>

<div align="center">
  <img src="https://raw.githubusercontent.com/scottkw/storcat/main/screenshots/light-mode-2.png" alt="StorCat Light Mode - Search Results" width="800">
  <p><em>Search Catalogs Tab - Results Table</em></p>
</div>

<div align="center">
  <img src="https://raw.githubusercontent.com/scottkw/storcat/main/screenshots/light-mode-3.png" alt="StorCat Light Mode - Browse Catalogs" width="800">
  <p><em>Browse Catalogs Tab - Catalog Collection</em></p>
</div>

### Dark Mode
<div align="center">
  <img src="https://raw.githubusercontent.com/scottkw/storcat/main/screenshots/dark-mode-1.png" alt="StorCat Dark Mode - Create Catalog" width="800">
  <p><em>Create Catalog Tab - Dark Theme</em></p>
</div>

<div align="center">
  <img src="https://raw.githubusercontent.com/scottkw/storcat/main/screenshots/dark-mode-2.png" alt="StorCat Dark Mode - Search Results" width="800">
  <p><em>Search Catalogs Tab - Dark Theme</em></p>
</div>

<div align="center">
  <img src="https://raw.githubusercontent.com/scottkw/storcat/main/screenshots/dark-mode-3.png" alt="StorCat Dark Mode - Browse Catalogs" width="800">
  <p><em>Browse Catalogs Tab - Dark Theme</em></p>
</div>

## 🖥️ System Requirements

### Minimum Requirements
- **Operating System**: Windows 10+, macOS 10.14+, or Linux (Ubuntu 18.04+ or equivalent)
- **RAM**: 4GB minimum, 8GB recommended
- **Disk Space**: 200MB for application installation
- **Display**: 1024x768 minimum resolution

### Software Requirements
- **Node.js**: Version 16.0.0 or higher (LTS recommended)
- **npm**: Version 7.0.0 or higher (comes with Node.js)

## 📦 Installation

### 🍺 Homebrew (macOS - Recommended)

The easiest way to install StorCat on macOS:

```bash
# Add the tap and install
brew tap scottkw/storcat
brew install storcat

# Or install directly in one command
brew install scottkw/storcat/storcat
```

**Benefits:**
- ✅ **Automatic architecture detection** (Intel vs Apple Silicon)
- ✅ **Easy updates** with `brew upgrade storcat`
- ✅ **Clean uninstall** with `brew uninstall --zap storcat`
- ✅ **No security warnings** after first launch approval

### 📦 winget (Windows - Recommended)

The easiest way to install StorCat on Windows:

```powershell
# Add the source and install
winget source add storcat https://github.com/scottkw/winget-storcat
winget install scottkw.StorCat

# Or install directly
winget install --source storcat scottkw.StorCat
```

**Benefits:**
- ✅ **Built into Windows 10/11** - no additional setup required
- ✅ **Portable installation** - no admin rights needed  
- ✅ **Easy updates** with `winget upgrade scottkw.StorCat`
- ✅ **Clean uninstall** with `winget uninstall scottkw.StorCat`
- ✅ **Command line access** - type `storcat` in any terminal

### Quick Install (Pre-built Binaries)

1. **Visit the [Releases page](https://github.com/scottkw/storcat/releases)** on GitHub
2. **Download the appropriate package** for your platform and architecture:
   - **Windows**: `.exe` (portable) or `.msi` (installer) for x64/ARM64
   - **macOS**: `.dmg` installer for Intel (x64) or Apple Silicon (ARM64)
   - **Linux**: `.AppImage` (portable), `.deb` (Debian/Ubuntu), `.rpm` (RedHat/CentOS), or `.snap` (Ubuntu)
3. **Install or run** the downloaded package using your platform's standard method
4. **Launch StorCat** and start creating catalogs!

### Install from Source

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-username/storcat.git
   cd storcat
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Run the application**:
   ```bash
   npm start
   ```

## 🛠️ Build Environment Setup

### macOS Setup

1. **Install Xcode Command Line Tools**:
   ```bash
   xcode-select --install
   ```

2. **Install Node.js** (using Homebrew - recommended):
   ```bash
   # Install Homebrew if not already installed
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   
   # Install Node.js LTS
   brew install node
   ```

3. **Alternative Node.js Installation**:
   - Download from [nodejs.org](https://nodejs.org/)
   - Or use [nvm](https://github.com/nvm-sh/nvm):
     ```bash
     curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
     nvm install --lts
     nvm use --lts
     ```

### Windows Setup

1. **Install Node.js**:
   - Download the Windows installer from [nodejs.org](https://nodejs.org/)
   - Choose the LTS version
   - Run the installer and follow the setup wizard

2. **Install Visual Studio Build Tools** (required for native modules):
   ```powershell
   # Using Chocolatey (recommended)
   choco install visualstudio2019buildtools
   
   # Or download from Microsoft:
   # https://visualstudio.microsoft.com/visual-cpp-build-tools/
   ```

3. **Alternative Node.js Installation** (using Chocolatey):
   ```powershell
   # Install Chocolatey first
   Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
   
   # Install Node.js
   choco install nodejs
   ```

### Linux Setup

#### Ubuntu/Debian:
```bash
# Update package index
sudo apt update

# Install Node.js LTS
curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install build essentials (required for native modules)
sudo apt-get install -y build-essential
```

#### CentOS/RHEL/Fedora:
```bash
# Install Node.js LTS
curl -fsSL https://rpm.nodesource.com/setup_lts.x | sudo bash -
sudo dnf install -y nodejs

# Install development tools
sudo dnf groupinstall -y "Development Tools"
```

#### Arch Linux:
```bash
# Install Node.js and npm
sudo pacman -S nodejs npm

# Install base development packages
sudo pacman -S base-devel
```

## 🔨 Building the Application

### Development Build

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Start development server**:
   ```bash
   npm run dev
   # or
   npm run dev-tools  # Opens with DevTools
   ```

3. **Build renderer for testing**:
   ```bash
   npm run build:renderer  # Build React frontend only
   ```

### Production Builds

#### Build for Current Platform
```bash
# Create distributable for your current platform
npm run build
```

#### Build for Specific Platforms
```bash
# macOS
npm run build:mac

# Windows
npm run build:win

# Linux
npm run build:linux

# All platforms (requires appropriate build environment)
npm run build:all
```

#### Version Management

StorCat includes automated version management that updates both the package.json and the Create Catalog welcome screen:

**🔥 Recommended: Version + Build (One Command)**
```bash
# For bug fixes (1.0.0 → 1.0.1)
npm run version:patch-build

# For new features (1.0.0 → 1.1.0)
npm run version:minor-build

# For breaking changes (1.0.0 → 2.0.0)
npm run version:major-build
```

**Version-Only Updates:**
```bash
# Update version without building
npm run version:patch    # Bug fixes
npm run version:minor    # New features  
npm run version:major    # Breaking changes
```

#### Multi-Platform Build Scripts

StorCat includes several build scripts optimized for different scenarios:

**🔥 Recommended: Smart OS-Aware Build**
```bash
# Automatically builds optimal packages for your OS
npm run build:complete

# On macOS: Builds 8 packages (portable Windows + macOS + Linux AppImage/DEB)
# On Windows: Builds 10 packages (+ MSI installers)
# On Linux: Builds 12 packages (+ RPM and Snap packages)
./build-complete.sh
```

**Alternative Build Options:**
```bash
# Parallel build (faster, may have compatibility issues)
npm run build:all-platforms

# Sequential build (more compatible)  
npm run build:fast

# Core packages only (6 packages)
npm run build:simple
```

**Supported Build Targets:**
- **Windows**: x64 and ARM64 portable executables (MSI installers available on Windows hosts)
- **macOS**: ARM64 (Apple Silicon) and x64 (Intel) DMG installers
- **Linux**: AppImage, DEB packages (all hosts), RPM and Snap packages (Linux hosts only)

**Generated Files:**
```
dist/
├── StorCat 1.0.0.exe                    # Windows x64/ARM64 portable (both architectures)
├── StorCat Setup 1.0.0.msi              # Windows x64 MSI installer (Windows hosts only)
├── StorCat Setup 1.0.0 (arm64).msi      # Windows ARM64 MSI installer (Windows hosts only)
├── StorCat-1.0.0.dmg                    # macOS x64 (Intel) installer
├── StorCat-1.0.0-arm64.dmg              # macOS ARM64 (Apple Silicon) installer
├── StorCat-1.0.0.AppImage               # Linux x64 portable
├── StorCat-1.0.0-arm64.AppImage         # Linux ARM64 portable
├── storcat-electron_1.0.0_amd64.deb     # Debian/Ubuntu x64 package
├── storcat-electron_1.0.0_arm64.deb     # Debian/Ubuntu ARM64 package
├── storcat-electron-1.0.0.x86_64.rpm    # RedHat/CentOS x64 package (Linux hosts only)
├── storcat-electron-1.0.0.aarch64.rpm   # RedHat/CentOS ARM64 package (Linux hosts only)
├── storcat-electron_1.0.0_amd64.snap    # Ubuntu Snap x64 package (Linux hosts only)
└── storcat-electron_1.0.0_arm64.snap    # Ubuntu Snap ARM64 package (Linux hosts only)
```

The build script automatically:
- Cleans previous builds
- **Displays current version** from package.json
- Builds the React renderer with Vite
- Creates distribution packages for all compatible platforms
- **Detects host OS** and builds appropriate packages
- **On macOS**: Builds 8 packages (skips MSI/RPM/Snap due to tooling limitations)
- **On Windows**: Builds 10 packages (+ MSI installers)
- **On Linux**: Builds 12 packages (+ RPM and Snap packages)
- Provides detailed progress output with **version information** and comprehensive summary

#### Package Without Distribution
```bash
# Create unpacked build for testing
npm run pack

# Platform-specific unpacked builds
npm run pack:mac
npm run pack:win
npm run pack:linux
```

### Build Output

Built applications will be available in the `dist/` directory:
- **Windows**: Portable `.exe` executables (all hosts), `.msi` installers (Windows hosts only)
- **macOS**: `.dmg` installers for both x64 (Intel) and ARM64 (Apple Silicon)
- **Linux**: `.AppImage` and `.deb` packages (all hosts), `.rpm` and `.snap` packages (Linux hosts only)

### Build Configuration

The build process is configured in `package.json` under the `build` section:

```json
{
  "build": {
    "appId": "com.kenscott.storcat",
    "productName": "StorCat",
    "directories": {
      "output": "dist"
    },
    "files": [
      "src/**/*",
      "assets/**/*",
      "node_modules/**/*"
    ]
  }
}
```

## 📖 Usage Guide

### Getting Started

1. **Launch StorCat** - Open the application after installation
2. **Configure Settings** - Click the gear icon in the header to access application settings
3. **Choose a Function** - Use the tab navigation to select your desired operation:
   - **Create Catalog**: Generate new directory catalogs
   - **Search Catalogs**: Find files across existing catalogs
   - **Browse Catalogs**: View and navigate catalog collections

### Configuring Application Settings

1. **Access Settings**:
   - Click the **gear icon** in the top-right corner of the header
   - The settings modal will open with organized configuration options

2. **Theme Selection**:
   - **Theme Dropdown**: Choose from 11 beautiful themes using the theme selector
   - **Available Themes**:
     - **StorCat Light/Dark**: Original application themes
     - **Dracula**: Popular dark theme with purple accents
     - **Solarized Light/Dark**: Low-contrast, easy on the eyes
     - **Nord**: Arctic-inspired cool color palette
     - **One Dark**: Popular VS Code theme
     - **Monokai**: Classic dark theme with orange accents
     - **GitHub Light/Dark**: Clean, professional themes
     - **Gruvbox Dark**: Retro groove color scheme
   - **Real-time Preview**: Themes apply instantly for immediate feedback

3. **Sidebar Position**:
   - **Position Toggle**: Choose between left or right sidebar placement
   - **Smart Icon Placement**: Toggle and settings buttons automatically position based on sidebar location
   - **Instant Updates**: Layout changes apply immediately without restart

4. **Settings Persistence**:
   - **Automatic Saving**: All settings save immediately when changed
   - **Session Restoration**: Settings persist between application launches
   - **Visual Feedback**: Active settings are clearly highlighted

5. **Closing Settings**:
   - Click the **X** button in the modal header
   - Click outside the modal to close
   - Press **Escape** key

### Creating Directory Catalogs

1. **Navigate to Create Tab**:
   - Click on the "Create Catalog" tab in the navigation bar

2. **Configure Catalog Settings**:
   - **Catalog Title**: Enter a descriptive title (appears in HTML output)
   - **Select Directory**: Click "Select Directory" to choose the folder to catalog
   - **Output Filename**: Specify the base name for generated files (e.g., "my-catalog" creates "my-catalog.json" and "my-catalog.html")

3. **Optional Secondary Copy**:
   - Check "Copy files to secondary location" to save additional copies
   - Select the destination directory for copies

4. **Generate Catalog**:
   - Click "Create Catalog" to start the process
   - View progress and results in the modal dialog

**Output Files**:
- `[filename].json`: Machine-readable catalog data
- `[filename].html`: Human-readable tree visualization

### Searching Catalogs

1. **Navigate to Search Tab**:
   - Click on the "Search Catalogs" tab

2. **Configure Search**:
   - **Search Term**: Enter the file or directory name to find
   - **Catalog Directory**: Select folder containing `.json` catalog files

3. **Execute Search**:
   - Click "Search Catalogs" to find matching items
   - Results display in a sortable, resizable table

4. **Working with Results**:
   - **Sort**: Click column headers to sort by File/Directory, Path, Type, Size, or Catalog
   - **Filter**: Use column filter inputs to narrow down results
   - **Resize**: Drag column borders to adjust widths
   - **View Catalog**: Click catalog names to open HTML representations in a full-screen modal
   - **Pagination**: Navigate large result sets with built-in pagination controls
   - **Manage Layout**: Use the sidebar toggle to maximize table viewing space

### Browsing Catalog Collections

1. **Navigate to Browse Tab**:
   - Click on the "Browse Catalogs" tab

2. **Load Catalogs**:
   - **Select Directory**: Choose folder containing catalog files
   - **Load Catalogs**: Click to display available catalogs

3. **Explore Catalogs**:
   - View catalog metadata (title, filename, creation date) in a modern table interface
   - Click catalog titles to open interactive HTML views in a full-screen modal
   - Sort by title, filename, or date created using table headers
   - Filter catalogs using the built-in filter controls
   - Navigate large collections with pagination

### Interface Features

#### Version Management
- **Dynamic Version Display**: Version number automatically updates from package.json throughout the application
- **Create Catalog Welcome Screen**: Displays current version dynamically without manual updates
- **Automated Version Bumping**: Single commands to increment version numbers and build distributions
- **Build Process Integration**: Version information included in build output and success messages

#### Settings Management
- **Enhanced Settings Panel**: Redesigned settings modal accessible via the gear icon in the bottom-left sidebar
- **Theme Toggle**: Modern switch control for changing between light and dark modes
- **Window Persistence**: Option to save and restore window size and position between sessions
- **Immediate Feedback**: All settings apply instantly with visual confirmation
- **Persistent Storage**: All preferences automatically save and restore between sessions

#### Advanced Table Features
- **Resizable Columns**: All tables support column resizing with visual separator indicators
- **Visual Column Separators**: Subtle border lines show exactly where to click and drag for resizing
- **Interactive Hover Effects**: Column separators highlight on hover for clear user feedback
- **Sticky Headers**: Table headers remain visible during scrolling
- **Advanced Filtering**: Per-column filter inputs with real-time search

#### Advanced Sidebar Management
- **Smart Toggle**: Collapsible sidebar with persistent icon bar when minimized
- **Intelligent Layout**: Toggle button positioned in top-left, settings in bottom-left
- **Space Optimization**: Maximizes viewing area while keeping essential controls accessible
- **Visual Hierarchy**: Clear separation between navigation and settings areas

#### Advanced Theme System
- **11 Professional Themes**: Choose from StorCat Light/Dark, Dracula, Nord, Solarized, One Dark, Monokai, GitHub, and Gruvbox themes
- **Real-time Switching**: Themes apply instantly using CSS variables and dynamic styling
- **Complete Coverage**: All UI components including tables, modals, and HTML catalog content adapt to selected theme
- **Smart Content Adaptation**: Catalog HTML automatically adapts to current theme with proper text contrast
- **Persistent Selection**: Theme preference saves automatically and restores between sessions
- **Migration Support**: Seamlessly migrates from old light/dark system to new theme selection

#### Flexible Interface Layout
- **Configurable Sidebar Position**: Choose between left or right sidebar placement via settings
- **Smart Icon Positioning**: Toggle and settings buttons automatically position based on sidebar location
- **Dynamic Layout Updates**: Position changes apply instantly without requiring application restart
- **Consistent Spacing**: Content padding and borders adjust intelligently based on sidebar position

#### Keyboard Shortcuts
- **Escape**: Close modal dialogs (settings or catalog views)
- **Tab Navigation**: Use arrow keys or Tab to navigate between controls
- **Settings Access**: Click gear icon or use mouse to access settings

## 🔄 File Compatibility

StorCat maintains 100% compatibility with catalog files from the original `sdcat.sh` bash script:

### JSON Format Compatibility
- **Structure**: Identical `type`, `name`, `size`, and `contents` fields
- **Data Types**: Maintains exact data formats and relationships
- **Encoding**: UTF-8 compatible with special characters and international filenames

### HTML Format Compatibility
- **Styling**: Same tree structure and CSS classes as original
- **Navigation**: Identical expand/collapse behavior
- **Metadata**: Preserves file counts and size summaries

### Interoperability
- **Mixed Usage**: Use catalogs created by either tool interchangeably
- **Search Results**: Same format and structure as bash script output
- **Migration**: Seamlessly transition between command-line and GUI workflows

### Original Bash Script
The original `sdcat.sh` script is available in the `examples/` directory along with sample output files for reference and testing.

## 🧑‍💻 Development

### Development Setup

1. **Clone and Install**:
   ```bash
   git clone https://github.com/your-username/storcat.git
   cd storcat
   npm install
   ```

2. **Start Development Server**:
   ```bash
   npm run dev
   ```

3. **Development with DevTools**:
   ```bash
   npm run dev-tools
   ```

### Available Scripts

| Script | Description |
|--------|-------------|
| `npm start` | Run the application in production mode |
| `npm run dev` | Start development server with hot reload |
| `npm run dev-tools` | Development mode with DevTools open |
| `npm run build:renderer` | Build React frontend only (for development) |
| `npm run build` | Build for current platform |
| `npm run build:mac` | Build macOS version |
| `npm run build:win` | Build Windows version |
| `npm run build:linux` | Build Linux version |
| `npm run build:all` | Build for all platforms |
| `npm run build:all-platforms` | **Multi-platform build script (parallel)** |
| `npm run build:fast` | **Multi-platform build script (sequential)** |
| `npm run build:simple` | **Core 6 packages (AppImage + Windows + macOS)** |
| `npm run build:complete` | **✅ RECOMMENDED: Smart OS-aware build (8-12 packages)** |
| `npm run version:patch` | **Bump patch version (1.0.0 → 1.0.1)** |
| `npm run version:minor` | **Bump minor version (1.0.0 → 1.1.0)** |
| `npm run version:major` | **Bump major version (1.0.0 → 2.0.0)** |
| `npm run version:patch-build` | **🚀 Bump patch version + build all packages** |
| `npm run version:minor-build` | **🚀 Bump minor version + build all packages** |
| `npm run version:major-build` | **🚀 Bump major version + build all packages** |
| `npm run pack` | Create unpacked build |

### Code Organization

The application follows a modern React-based architecture:

- **Main Process** (`src/main.js`): Electron main process and IPC handlers
- **React Renderer** (`src/renderer/`): Modern React-based frontend with TypeScript
- **Component Library** (`src/renderer/components/`): Reusable React components
- **Context Management** (`src/renderer/contexts/`): Global state management
- **Catalog Service** (`src/catalog-service.js`): Core catalog creation and search functionality
- **Preload Script** (`src/preload.js`): Secure bridge between main and renderer processes

### Key Technologies

- **Electron 28.0.0**: Cross-platform desktop framework
- **React 18.2.0**: Modern component-based UI framework
- **TypeScript 5.2.2**: Type-safe development and enhanced IDE support
- **Ant Design 5.12.8**: Professional React component library
- **Vite 5.0.8**: Fast build tool and development server
- **Node.js File System APIs**: Directory traversal and file operations

## 📁 Project Structure

```
storcat/
├── src/                          # Source code
│   ├── main.js                   # Electron main process
│   ├── catalog-service.js        # Catalog creation and search
│   ├── preload.js               # Secure IPC bridge
│   └── renderer/                # React frontend
│       ├── App.tsx              # Main React application
│       ├── main.tsx             # React entry point
│       ├── index.css            # Global styles and CSS variables
│       ├── components/          # React components
│       │   ├── Header.tsx       # Title bar with logo
│       │   ├── MainContent.tsx  # Main layout and sidebar
│       │   ├── CatalogModal.tsx # Full-screen catalog viewer
│       │   ├── ModernTable.tsx  # Enhanced table component
│       │   └── tabs/            # Tab-specific components
│       │       ├── CreateCatalogTab.tsx
│       │       ├── SearchCatalogsTab.tsx
│       │       └── BrowseCatalogsTab.tsx
│       ├── contexts/            # React context providers
│       │   └── AppContext.tsx   # Global state management
│       └── types/               # TypeScript definitions
│           └── electron.d.ts    # Electron API types
├── examples/                     # Reference files
│   ├── sdcat.sh                 # Original bash script
│   ├── sd01.json               # Sample JSON catalog
│   └── sd01.html               # Sample HTML catalog
├── storcat-icons/               # Icon resources
│   ├── storcat-logo.png        # Main logo
│   ├── icon.icns               # macOS icon
│   └── [various icon sizes]    # Platform-specific icons
├── build/                      # Build configuration
├── build-all.sh               # Multi-platform build script (parallel)
├── build-fast.sh              # Multi-platform build script (sequential)  
├── build-simple.sh            # Core packages build script
├── build-complete.sh          # ✅ Smart OS-aware build script (recommended)
├── dist/                       # Built distributions
├── node_modules/              # Dependencies
├── package.json              # Project configuration
├── package-lock.json         # Dependency lock file
└── README.md                # This file
```

### Key Files

| File | Purpose |
|------|---------|
| `src/main.js` | Electron main process, window management, IPC |
| `src/renderer/App.tsx` | Main React application with routing and global state |
| `src/renderer/components/` | Reusable React components and UI elements |
| `src/renderer/contexts/AppContext.tsx` | Global state management with React Context |
| `src/catalog-service.js` | Directory traversal, catalog generation, search |
| `src/preload.js` | Secure communication bridge |
| `package.json` | Dependencies, scripts, build configuration |

## 🐛 Troubleshooting

### Common Issues

#### Build Failures

**Issue**: `npm install` fails with native module errors
```bash
# Solution: Install platform build tools
# macOS:
xcode-select --install

# Windows:
npm install --global windows-build-tools

# Linux:
sudo apt-get install build-essential
```

**Issue**: Electron build fails with permission errors
```bash
# Solution: Clear npm cache and reinstall
npm cache clean --force
rm -rf node_modules package-lock.json
npm install
```

#### Runtime Issues

**Issue**: Application won't start or shows blank window
- **Check**: Node.js version (must be 16.0.0+)
- **Solution**: Update Node.js or use nvm to manage versions

**Issue**: Catalog creation fails with permission errors
- **Check**: Directory permissions and file system access
- **Solution**: Run as administrator/sudo if necessary, or choose accessible directories

**Issue**: Search returns no results
- **Check**: Catalog directory contains `.json` files
- **Verify**: Catalog files are valid JSON format
- **Solution**: Recreate catalogs if corrupted

#### Performance Issues

**Issue**: Large directories cause application to freeze
- **Limit**: Very large directories (>100,000 files) may require patience
- **Solution**: Consider cataloging subdirectories separately

**Issue**: High memory usage during catalog creation
- **Expected**: Memory usage scales with directory size
- **Solution**: Close other applications during large catalog operations

### Getting Help

1. **Check Issues**: Search existing [GitHub Issues](https://github.com/your-username/storcat/issues)
2. **Create Issue**: Report bugs with system information and steps to reproduce
3. **Discussions**: Use [GitHub Discussions](https://github.com/your-username/storcat/discussions) for questions

### Debug Mode

Enable debug logging:
```bash
# Set environment variable before starting
DEBUG=storcat:* npm run dev
```

## 🤝 Contributing

We welcome contributions to StorCat! Here's how to get involved:

### Development Process

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Contribution Guidelines

- **Code Style**: Follow existing patterns and conventions
- **Testing**: Test on multiple platforms when possible
- **Documentation**: Update README and comments for new features
- **Security**: Follow Electron security best practices

### Areas for Contribution

- 🌐 **Internationalization**: Add support for multiple languages
- ⚡ **Performance**: Optimize large directory handling
- 🎨 **UI/UX**: Improve interface design and usability
- 📱 **Features**: Add new catalog management capabilities
- 🧪 **Testing**: Expand test coverage and automation

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2024 Ken Scott

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

<div align="center">
  <p><strong>StorCat - Making Directory Management Simple</strong></p>
  <p>Built with ❤️ using Electron and modern web technologies</p>
</div>