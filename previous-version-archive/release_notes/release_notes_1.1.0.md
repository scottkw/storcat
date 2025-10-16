# StorCat v1.1.0 Release Notes

**Release Date:** July 26, 2025  
**Version:** 1.1.0  
**Build:** Comprehensive multi-platform distribution

---

## 🎉 What's New in StorCat 1.1.0

### ✨ Major Enhancements

#### 🔧 **Comprehensive Build System**
- **Multi-platform builds** supporting 8-14 distribution packages
- **Smart OS detection** automatically builds optimal packages for your system
- **Enhanced build scripts** with version integration and progress tracking
- **Complete architecture support** for x64 and ARM64 across all platforms

#### 📊 **Dynamic Version Management**
- **Automatic version display** throughout the application from package.json
- **One-command version bumping** with integrated builds (`npm run version:minor-build`)
- **Build process integration** showing version information in all build outputs
- **Dynamic Create Catalog welcome screen** always shows current version

#### 🎯 **Enhanced Table Experience**
- **Visual column separators** in table headers with resize indicators
- **Interactive hover effects** for clear user feedback on resizable columns
- **Improved column resizing** with better visual cues
- **Sticky headers** remain visible during table scrolling

### 🏗️ **Build & Distribution Improvements**

#### 📦 **Distribution Packages**
- **Windows**: Portable executables + MSI installers (x64/ARM64)
- **macOS**: DMG installers for Intel and Apple Silicon
- **Linux**: AppImage, DEB, RPM, and Snap packages (x64/ARM64)
- **Smart package selection** based on build host capabilities

#### 🔄 **Automated Workflows**
- **Version + Build commands** for streamlined releases
- **OS-aware building** skips incompatible packages automatically
- **Enhanced error handling** with clear messaging about platform limitations
- **Progress tracking** with version information throughout build process

### 🎨 **User Interface Refinements**

#### 📋 **Table Enhancements**
- **Column resize indicators** show exactly where to click and drag
- **Visual feedback** on hover for better user experience
- **Consistent table styling** across all tabs
- **Improved accessibility** with clear visual cues

#### 💾 **Data Persistence Improvements**
- **Auto-reload functionality** on app startup for Search and Browse tabs
- **Restored table data** from previous sessions
- **Enhanced state management** for better user experience
- **Seamless session restoration** without manual reloading

---

## 🔧 Technical Improvements

### **Build System Architecture**
- **Electron-builder optimization** with individual platform builds
- **MSI build compatibility** handled via OS detection
- **RPM/Snap builds** restricted to Linux hosts for reliability
- **Parallel and sequential** build options for different scenarios

### **Version Management System**
- **Dynamic imports** from package.json for version display
- **Automated version bumping** with npm scripts
- **Build integration** showing version in all outputs
- **Future-proof versioning** for release management

### **User Experience**
- **Visual consistency** across all interface elements
- **Enhanced feedback systems** for interactive elements
- **Improved accessibility** with better visual indicators
- **Professional polish** in table interactions

---

## 📋 Available Distribution Packages

### **Windows** (2-4 packages)
- `StorCat 1.1.0.exe` - Portable executable (x64/ARM64)
- `StorCat Setup 1.1.0.msi` - MSI installer x64 *(Windows hosts only)*
- `StorCat Setup 1.1.0 (arm64).msi` - MSI installer ARM64 *(Windows hosts only)*

### **macOS** (2 packages)
- `StorCat-1.1.0.dmg` - Intel (x64) installer
- `StorCat-1.1.0-arm64.dmg` - Apple Silicon (ARM64) installer

### **Linux** (4-8 packages)
- `StorCat-1.1.0.AppImage` - x64 portable executable
- `StorCat-1.1.0-arm64.AppImage` - ARM64 portable executable
- `storcat-electron_1.1.0_amd64.deb` - Debian/Ubuntu x64 package
- `storcat-electron_1.1.0_arm64.deb` - Debian/Ubuntu ARM64 package
- `storcat-electron-1.1.0.x86_64.rpm` - RedHat/CentOS x64 *(Linux hosts only)*
- `storcat-electron-1.1.0.aarch64.rpm` - RedHat/CentOS ARM64 *(Linux hosts only)*
- `storcat-electron_1.1.0_amd64.snap` - Ubuntu Snap x64 *(Linux hosts only)*
- `storcat-electron_1.1.0_arm64.snap` - Ubuntu Snap ARM64 *(Linux hosts only)*

---

## 🚀 Getting Started

### **Installation**
1. Download the appropriate package for your platform and architecture
2. Install using your platform's standard method
3. Launch StorCat and enjoy the enhanced experience

### **For Developers**
```bash
# Clone and install
git clone https://github.com/your-username/storcat.git
cd storcat
npm install

# Development
npm run dev

# Build all compatible packages
npm run build:complete

# Version management
npm run version:minor-build  # Bump version and build
```

---

## 🔍 What's Next

### **Future Enhancements**
- **Auto-update system** utilizing the .blockmap files
- **Enhanced catalog visualization** features
- **Performance optimizations** for large directories
- **Additional export formats** and sharing options

### **Platform Expansion**
- **Complete Linux distribution** support with native RPM/Snap builds
- **Windows Store** distribution package
- **Enhanced macOS integration** with code signing

---

## 🐛 Bug Fixes & Improvements

- **Fixed table column resizing** visual indicators
- **Improved build reliability** across different host systems
- **Enhanced error handling** in build processes
- **Better session state management** for persistent user experience
- **Resolved MSI build issues** on non-Windows systems

---

## 📞 Support & Feedback

- **GitHub Issues**: Report bugs and request features
- **Documentation**: Updated README with all new features
- **Website**: [www.eightabyte.com](https://www.eightabyte.com)
- **Contact**: ken@eightabyte.com

---

**Thank you for using StorCat!** This release represents a significant step forward in making directory catalog management more accessible and professional across all platforms.

*Built with ❤️ using Electron and modern web technologies*