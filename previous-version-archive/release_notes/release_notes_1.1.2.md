# StorCat v1.1.2 Release Notes

**Release Date:** July 27, 2025  
**Version:** 1.1.2  
**Build:** Bug fix release for macOS Apple Silicon

---

## 🐛 Bug Fixes

### **🍎 Fixed macOS Apple Silicon "Damaged App" Issue**

**Problem Resolved:**
- **ARM64 DMG builds** were showing "StorCat is damaged and can't be opened" on Apple Silicon Macs
- **Issue affected** both direct downloads and Homebrew installations
- **Root cause**: Improper electron-builder configuration for unsigned ARM64 builds

**Technical Fix:**
- **Enhanced macOS build configuration** with proper signing settings for unsigned apps
- **Added explicit target architecture** specification in electron-builder
- **Disabled conflicting security features** that caused bundle validation failures
- **Improved DMG generation** with correct signing parameters

**Result:**
- **✅ ARM64 builds now work properly** on Apple Silicon Macs (M1/M2/M3)
- **✅ No more "damaged app" errors** for Apple Silicon users
- **✅ Homebrew installations** work seamlessly on all Mac architectures
- **✅ Maintains compatibility** with Intel Macs via Rosetta

---

## 📋 Available Distribution Packages

### **Windows** (2 packages)
- `StorCat.1.1.2.exe` - Portable executable (x64)

### **macOS** (2 packages)
- `StorCat-1.1.2.dmg` - Intel (x64) installer
- `StorCat-1.1.2-arm64.dmg` - Apple Silicon (ARM64) installer **[FIXED]**

### **Linux** (4-8 packages)
- `StorCat-1.1.2.AppImage` - x64 portable executable
- `StorCat-1.1.2-arm64.AppImage` - ARM64 portable executable
- `storcat-electron_1.1.2_amd64.deb` - Debian/Ubuntu x64 package
- `storcat-electron_1.1.2_arm64.deb` - Debian/Ubuntu ARM64 package
- `storcat-electron-1.1.2.x86_64.rpm` - RedHat/CentOS x64 *(Linux hosts only)*
- `storcat-electron-1.1.2.aarch64.rpm` - RedHat/CentOS ARM64 *(Linux hosts only)*
- `storcat-electron_1.1.2_amd64.snap` - Ubuntu Snap x64 *(Linux hosts only)*
- `storcat-electron_1.1.2_arm64.snap` - Ubuntu Snap ARM64 *(Linux hosts only)*

---

## 🚀 Getting Started

### **Installation**
1. Download the appropriate package for your platform and architecture
2. Install using your platform's standard method
3. **macOS Apple Silicon users**: The ARM64 build now works without issues!

### **Package Managers**
```bash
# macOS (Homebrew) - Now fully supports Apple Silicon
brew tap scottkw/storcat
brew upgrade storcat

# Windows (winget)
winget source add storcat https://github.com/scottkw/winget-storcat
winget upgrade scottkw.StorCat
```

### **For Developers**
```bash
# Clone and install
git clone https://github.com/scottkw/storcat.git
cd storcat
npm install

# Development
npm run dev

# Build all compatible packages
npm run build:complete
```

---

## 🔧 Technical Details

### **Build Configuration Changes**
```json
{
  "mac": {
    "identity": null,
    "hardenedRuntime": false,
    "gatekeeperAssess": false,
    "target": [
      {
        "target": "dmg",
        "arch": ["x64", "arm64"]
      }
    ]
  },
  "dmg": {
    "sign": false
  }
}
```

### **What Changed:**
- **Explicit architecture targeting** for both Intel and Apple Silicon
- **Disabled code signing conflicts** that caused bundle corruption
- **Fixed app bundle identifier** (was showing as "Electron" instead of proper bundle ID)
- **Improved DMG creation** process for unsigned applications

---

## 📞 Support & Feedback

- **GitHub Issues**: [Report bugs and request features](https://github.com/scottkw/storcat/issues)
- **Documentation**: Updated README with all features
- **Repository**: [github.com/scottkw/storcat](https://github.com/scottkw/storcat)
- **Contact**: ken@eightabyte.com

---

**Thank you for using StorCat!** This release ensures that Apple Silicon Mac users can enjoy StorCat without any installation issues.

*Built with ❤️ using Electron and modern web technologies*