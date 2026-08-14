import {
  CreateCatalog,
  SearchCatalogs,
  SearchIndexed,
  BrowseCatalogs,
  LoadCatalog,
  LoadCatalogFlat,
  RevealInFileManager,
  GetConfig,
  SetTheme,
  SetSidebarPosition,
  SetWindowSize,
  SelectDirectory,
  ReadHtmlFile,
  GetCatalogHtmlPath,
  OpenExternal,
  GetWindowPersistence,
  SetWindowPersistence,
  GetVersion,
} from '../../wailsjs/go/main/App';

// Wails' generated bindings reject a Go error's message as a plain string,
// not an Error instance -- `error.message` is undefined for every rejection
// from this bridge, which silently downgraded every real Go error (a
// missing file, a permission error, a parse failure) to the fallback
// 'Unknown error' string. Every catch block below reads through this so a
// caller (e.g. the unreadable-catalog panel) sees the real cause verbatim.
function extractErrorMessage(error: any): string {
  if (typeof error === 'string') return error;
  return error?.message || 'Unknown error';
}

// The one place a caught Wails rejection turns into the wrapper's
// {success: false, error} shape. Previously every catch block below
// duplicated this three-line object literal, and five of them drifted to
// read the (always-undefined, per extractErrorMessage's comment above)
// `error.message` directly instead. Routing every catch block through this
// single function makes that drift impossible to reintroduce one call site
// at a time -- there is now exactly one line that reads a caught error.
function wailsError(error: any): { success: false; error: string } {
  return { success: false, error: extractErrorMessage(error) };
}

// Wrapper to match Electron API structure
export const wailsAPI = {
  // Catalog operations
  createCatalog: async (title: string, directoryPath: string, outputName: string, copyToDirectory: string) => {
    try {
      const result = await CreateCatalog(title, directoryPath, outputName, copyToDirectory);
      return {
        success: true as const,
        jsonPath: result.jsonPath,
        htmlPath: result.htmlPath,
        fileCount: result.fileCount,
        totalSize: result.totalSize,
        copyJsonPath: result.copyJsonPath,
        copyHtmlPath: result.copyHtmlPath,
      };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  searchCatalogs: async (searchTerm: string, catalogDir: string) => {
    try {
      const results = await SearchCatalogs(searchTerm, catalogDir);
      return { success: true, results };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // GUI-only capped sibling of searchCatalogs, used by the ⌘K command
  // palette: caps the response server-side while carrying the true match
  // count in `indexed.total`.
  searchIndexed: async (searchTerm: string, catalogDir: string) => {
    try {
      const indexed = await SearchIndexed(searchTerm, catalogDir);
      return { success: true as const, indexed };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  browseCatalogs: async (catalogDir: string) => {
    try {
      const catalogs = await BrowseCatalogs(catalogDir);
      return { success: true, catalogs };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  loadCatalog: async (filePath: string) => {
    try {
      const catalog = await LoadCatalog(filePath);
      return { success: true, catalog };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  loadCatalogFlat: async (filePath: string) => {
    try {
      const flat = await LoadCatalogFlat(filePath);
      return { success: true as const, flat };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Directory selection
  selectDirectory: async () => {
    try {
      const path = await SelectDirectory();
      return { success: true as const, path };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  selectSearchDirectory: async () => {
    try {
      const path = await SelectDirectory();
      return { success: true as const, path };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  selectOutputDirectory: async () => {
    try {
      const path = await SelectDirectory();
      return { success: true as const, path };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Config operations
  getConfig: async () => {
    try {
      const config = await GetConfig();
      return { success: true as const, config };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setTheme: async (theme: string) => {
    try {
      await SetTheme(theme);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setSidebarPosition: async (position: string) => {
    try {
      await SetSidebarPosition(position);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setWindowSize: async (width: number, height: number) => {
    try {
      await SetWindowSize(width, height);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // File operations
  getCatalogHtmlPath: async (catalogPath: string) => {
    try {
      const htmlPath = await GetCatalogHtmlPath(catalogPath);
      return { success: true as const, htmlPath };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  readHtmlFile: async (filePath: string) => {
    try {
      const content = await ReadHtmlFile(filePath);
      return { success: true as const, content };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  openExternal: async (url: string) => {
    try {
      await OpenExternal(url);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // catalogDir is the frontend's currently configured catalog directory --
  // threaded through so the Go side can reject any path that does not
  // resolve inside it (WR-02: this binding is callable from any renderer
  // JS, not only this call site).
  revealInFileManager: async (path: string, catalogDir: string) => {
    try {
      await RevealInFileManager(path, catalogDir);
      return { success: true as const };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Catalog files (alias for browseCatalogs)
  getCatalogFiles: async (catalogDir: string) => {
    try {
      const catalogs = await BrowseCatalogs(catalogDir);
      return { success: true, catalogs };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Window persistence
  getWindowPersistence: async () => {
    try {
      const enabled = await GetWindowPersistence();
      return { success: true as const, enabled };
    } catch (error: any) {
      return { success: false as const, enabled: true }; // default to enabled on error
    }
  },

  setWindowPersistence: async (enabled: boolean) => {
    try {
      await SetWindowPersistence(enabled);
      return { success: true };
    } catch (error: any) {
      console.error('Failed to save window persistence setting:', error);
      return wailsError(error);
    }
  },

  getVersion: async () => {
    try {
      const version = await GetVersion();
      return { success: true as const, version };
    } catch (error: any) {
      return { success: false as const, version: 'dev' };
    }
  },
};

// Declare for TypeScript
declare global {
  interface Window {
    electronAPI: typeof wailsAPI;
  }
}

// Make available as window.electronAPI for compatibility
if (typeof window !== 'undefined') {
  window.electronAPI = wailsAPI;
}
