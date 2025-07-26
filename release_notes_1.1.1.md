# StorCat v1.1.1 Release Notes

**Release Date:** July 26, 2025  
**Version:** 1.1.1  
**Build:** Minor update with GitHub URL correction

---

## 🎉 What's New in StorCat 1.1.1

### ✨ Updates

#### 🔗 **GitHub URL Update**
- **Updated project URL** on Create Catalog welcome screen to point to the official GitHub repository
- **Clickable link** now directs users to `https://github.com/scottkw/storcat/` for project access
- **Consistent branding** with proper GitHub repository reference throughout the application

---

## 🔧 Technical Improvements

### **URL Management**
- **Updated project homepage** reference in Create Catalog tab
- **External link functionality** properly configured for GitHub repository access
- **User experience enhancement** with direct access to project source and documentation

---

## 📋 Available Distribution Packages

### **Windows** (2-4 packages)
- `StorCat 1.1.1.exe` - Portable executable (x64/ARM64)
- `StorCat Setup 1.1.1.msi` - MSI installer x64 *(Windows hosts only)*
- `StorCat Setup 1.1.1 (arm64).msi` - MSI installer ARM64 *(Windows hosts only)*

### **macOS** (2 packages)
- `StorCat-1.1.1.dmg` - Intel (x64) installer
- `StorCat-1.1.1-arm64.dmg` - Apple Silicon (ARM64) installer

### **Linux** (4-8 packages)
- `StorCat-1.1.1.AppImage` - x64 portable executable
- `StorCat-1.1.1-arm64.AppImage` - ARM64 portable executable
- `storcat-electron_1.1.1_amd64.deb` - Debian/Ubuntu x64 package
- `storcat-electron_1.1.1_arm64.deb` - Debian/Ubuntu ARM64 package
- `storcat-electron-1.1.1.x86_64.rpm` - RedHat/CentOS x64 *(Linux hosts only)*
- `storcat-electron-1.1.1.aarch64.rpm` - RedHat/CentOS ARM64 *(Linux hosts only)*
- `storcat-electron_1.1.1_amd64.snap` - Ubuntu Snap x64 *(Linux hosts only)*
- `storcat-electron_1.1.1_arm64.snap` - Ubuntu Snap ARM64 *(Linux hosts only)*

---

## 🚀 Getting Started

### **Installation**
1. Download the appropriate package for your platform and architecture
2. Install using your platform's standard method
3. Launch StorCat and enjoy the enhanced experience

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

# Version management
npm run version:patch-build  # Bump version and build
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

- **Updated GitHub repository URL** in Create Catalog welcome screen
- **Improved project discoverability** with correct repository link
- **Enhanced user access** to project documentation and source code

---

## 📞 Support & Feedback

- **GitHub Issues**: [Report bugs and request features](https://github.com/scottkw/storcat/issues)
- **Documentation**: Updated README with all features
- **Repository**: [github.com/scottkw/storcat](https://github.com/scottkw/storcat)
- **Contact**: ken@eightabyte.com

---

**Thank you for using StorCat!** This minor update ensures users can easily access the project repository and stay connected with development updates.

*Built with ❤️ using Electron and modern web technologies*