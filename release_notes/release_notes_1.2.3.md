# StorCat v1.2.3 Release Notes

**Release Date:** August 2, 2025  
**Build Type:** Patch Release  

## 🚀 Major Fixes

### ✅ **Apple Silicon ARM64 Support Fixed**
- **Fixed critical issue where ARM64 builds displayed blank screens on Apple Silicon Macs**
- **Resolved V8 JavaScript engine memory allocation errors on M1/M2/M3 processors**
- ARM64 version now runs natively without crashes or display issues

### 🔧 **Technical Improvements**

#### JavaScript Engine Optimization
- **Implemented JIT-less mode** to prevent V8 memory allocation failures on ARM64
- Added `--jitless --no-opt` flags to ensure compatibility with Apple Silicon architecture
- Maintained full application functionality while resolving stability issues

#### Electron Framework Update
- **Updated Electron from v28.0.0 to v35.7.2**
- Latest version includes native ARM64 optimizations and security patches
- Better memory management and performance on Apple Silicon

#### Application Stability
- Fixed blank window issues that affected ARM64 builds in previous versions
- Improved file loading mechanism for better asset resolution
- Enhanced error handling and debugging output

#### Build System Improvements
- **Fixed circular file inclusion issue** that caused excessive build sizes
- **Resolved ASAR file size limit errors** during Linux ARM64 builds
- **Optimized build configuration** to prevent dist directory recursion
- **Enhanced multi-platform build process** with proper file exclusions

## 📦 **Build Artifacts**

### macOS
- **StorCat-1.2.3-universal.dmg** - Universal binary (Intel x64 + Apple Silicon ARM64)
- Native performance on both Intel Macs and Apple Silicon (M1/M2/M3)
- Code-signed with Developer ID and ready for notarization

### Windows
- **StorCat 1.2.3.exe** - Portable executable (x64 and ARM64 versions)
- No installation required, runs directly

### Linux
- **StorCat-1.2.3.AppImage** - x64 AppImage for universal Linux compatibility
- **StorCat-1.2.3-arm64.AppImage** - ARM64 AppImage for ARM-based Linux systems
- **storcat-electron_1.2.3_amd64.deb** - Debian package for x64 systems
- **storcat-electron_1.2.3_arm64.deb** - Debian package for ARM64 systems

All builds are code-signed and ready for distribution across platforms.

## 🔄 **Migration Notes**

### From v1.2.2 to v1.2.3
- **No breaking changes** - existing catalogs and settings are fully compatible
- **No user action required** - the update only affects internal engine behavior
- Previous window positions and preferences are preserved

### All Mac Users
- **Universal Binary:** Single DMG (`StorCat-1.2.3-universal.dmg`) works on all Mac architectures
- **Automatic Detection:** Runs natively on both Intel and Apple Silicon without user intervention
- **No Rosetta Required:** Native ARM64 performance on Apple Silicon Macs

## 🐛 **Resolved Issues**

| Issue | Description | Resolution |
|-------|-------------|------------|
| Blank Screen | ARM64 builds showed empty window | Fixed with JIT-less V8 mode |
| V8 OOM Errors | Memory allocation failures on Apple Silicon | Resolved with Electron 35 + engine flags |
| Garbled Text | Corrupted content in some builds | Fixed asset loading mechanism |
| Build Size Bloat | Linux ARM64 builds exceeded 4.2GB limit | Fixed circular file inclusion in build config |
| ASAR Errors | Build process failed with "file size too large" | Optimized file patterns to exclude build artifacts |

## 🛠 **Technical Details**

### System Requirements
- **macOS:** 10.12 Sierra or later (Universal binary supports Intel x64 and Apple Silicon ARM64)
- **Windows:** Windows 10 or later (x64 and ARM64 support)
- **Linux:** Most modern distributions (AppImage and DEB packages available for x64 and ARM64)
- **Memory:** 4GB RAM recommended
- **Storage:** 200MB available space

### Performance Improvements
- **Native ARM64 performance** on Apple Silicon Macs and ARM64 Windows/Linux
- **Reduced memory footprint** with optimized V8 flags
- **Faster startup times** with improved asset loading
- **Smaller build sizes** with optimized file inclusion patterns
- **Cross-platform consistency** with universal binary approach

## 📋 **Testing Completed**

✅ **ARM64 native execution** - No crashes or blank screens on Apple Silicon  
✅ **Intel compatibility** - Maintains backward compatibility on Intel Macs  
✅ **Universal binary** - Single macOS build works on both architectures  
✅ **Multi-platform builds** - Windows, Linux, macOS all build successfully  
✅ **File operations** - Catalog creation and search functionality  
✅ **Window management** - Proper display and resizing across platforms  
✅ **Asset loading** - HTML, CSS, and JavaScript resources load correctly  
✅ **Build optimization** - No circular references or size limit errors  

## 🔒 **Security**

- **Code signing maintained** - All builds signed with valid Developer ID
- **Hardened runtime enabled** - Enhanced security on macOS
- **Notarization ready** - macOS builds ready for Apple notarization process
- **Multi-platform security** - Consistent security model across all platforms

## 📞 **Support**

For issues or questions regarding this release:
- **GitHub Issues:** [StorCat Issues](https://github.com/your-repo/storcat/issues)
- **Developer:** Ken Scott (ken@eightabyte.com)
- **Website:** https://www.eightabyte.com

---

## 🎯 **Summary**

Version 1.2.3 is a **critical patch release** that resolves the major ARM64 compatibility issue and build system problems from previous versions. This release provides:

- **Universal macOS binary** that works natively on both Intel and Apple Silicon
- **Complete multi-platform support** with Windows and Linux builds
- **Resolved build system issues** that prevented proper distribution
- **Optimized file sizes** and build processes

Apple Silicon Mac users can now enjoy native performance without the blank screen bug, while all platforms benefit from the improved build system and smaller package sizes.

**Recommended for all users across all platforms.**