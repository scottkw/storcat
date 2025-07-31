# StorCat v1.2.3 Release Notes

**Release Date:** July 31, 2025  
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

## 📦 **Build Artifacts**

### macOS
- **StorCat-1.2.3.dmg** - Intel x64 build (runs on all Intel Macs and Apple Silicon via Rosetta 2)
- **StorCat-1.2.3-arm64.dmg** - Native Apple Silicon build (M1/M2/M3 optimized)

Both builds are code-signed with Developer ID and ready for distribution.

## 🔄 **Migration Notes**

### From v1.2.2 to v1.2.3
- **No breaking changes** - existing catalogs and settings are fully compatible
- **No user action required** - the update only affects internal engine behavior
- Previous window positions and preferences are preserved

### Apple Silicon Users
- **Recommended:** Use the ARM64 build (`StorCat-1.2.3-arm64.dmg`) for optimal performance
- **Alternative:** Intel build continues to work via Rosetta 2 translation

## 🐛 **Resolved Issues**

| Issue | Description | Resolution |
|-------|-------------|------------|
| Blank Screen | ARM64 builds showed empty window | Fixed with JIT-less V8 mode |
| V8 OOM Errors | Memory allocation failures on Apple Silicon | Resolved with Electron 35 + engine flags |
| Garbled Text | Corrupted content in some builds | Fixed asset loading mechanism |

## 🛠 **Technical Details**

### System Requirements
- **macOS:** 10.12 Sierra or later
- **Architecture:** Intel x64 or Apple Silicon (ARM64)
- **Memory:** 4GB RAM recommended
- **Storage:** 200MB available space

### Performance Improvements
- **Native ARM64 performance** on Apple Silicon Macs
- **Reduced memory footprint** with optimized V8 flags
- **Faster startup times** with improved asset loading

## 📋 **Testing Completed**

✅ **ARM64 native execution** - No crashes or blank screens  
✅ **Intel compatibility** - Maintains backward compatibility  
✅ **File operations** - Catalog creation and search functionality  
✅ **Window management** - Proper display and resizing  
✅ **Asset loading** - HTML, CSS, and JavaScript resources load correctly  

## 🔒 **Security**

- **Code signing maintained** - Both builds signed with valid Developer ID
- **Hardened runtime enabled** - Enhanced security on macOS
- **Notarization compatible** - Ready for Apple notarization process

## 📞 **Support**

For issues or questions regarding this release:
- **GitHub Issues:** [StorCat Issues](https://github.com/your-repo/storcat/issues)
- **Developer:** Ken Scott (ken@eightabyte.com)
- **Website:** https://www.eightabyte.com

---

## 🎯 **Summary**

Version 1.2.3 is a **critical patch release** that resolves the major ARM64 compatibility issue introduced in previous versions. Apple Silicon Mac users can now enjoy native performance without the blank screen bug. This release ensures StorCat works seamlessly across all supported Mac architectures.

**Recommended for all users, especially those with Apple Silicon Macs.**