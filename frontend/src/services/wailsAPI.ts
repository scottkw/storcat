import {
  CreateCatalog,
  SearchCatalogs,
  BrowseCatalogs,
  LoadCatalog,
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
} from '../../wailsjs/go/main/App';

// Wrapper to match Electron API structure
export const wailsAPI = {
  // Catalog operations
  createCatalog: async (title: string, directoryPath: string, outputName: string, copyToDirectory: string) => {
    try {
      await CreateCatalog(title, directoryPath, outputName, copyToDirectory);
      return { success: true };
    } catch (error: any) {
      return { success: false, error: error.message || 'Unknown error' };
    }
  },

  searchCatalogs: async (searchTerm: string, catalogDir: string) => {
    try {
      const results = await SearchCatalogs(searchTerm, catalogDir);
      return { success: true, results };
    } catch (error: any) {
      return { success: false, error: error.message || 'Unknown error' };
    }
  },

  browseCatalogs: async (catalogDir: string) => {
    try {
      const catalogs = await BrowseCatalogs(catalogDir);
      return { success: true, catalogs };
    } catch (error: any) {
      return { success: false, error: error.message || 'Unknown error' };
    }
  },

  loadCatalog: async (filePath: string) => {
    try {
      const catalog = await LoadCatalog(filePath);
      return { success: true, catalog };
    } catch (error: any) {
      return { success: false, error: error.message || 'Unknown error' };
    }
  },

  // Directory selection
  selectDirectory: async () => {
    try {
      const path = await SelectDirectory();
      return { success: true as const, path };
    } catch (error: any) {
      return { success: false as const, error: error.message || 'Unknown error' };
    }
  },

  selectSearchDirectory: async () => {
    try {
      const path = await SelectDirectory();
      return { success: true as const, path };
    } catch (error: any) {
      return { success: false as const, error: error.message || 'Unknown error' };
    }
  },

  selectOutputDirectory: async () => {
    try {
      const path = await SelectDirectory();
      return { success: true as const, path };
    } catch (error: any) {
      return { success: false as const, error: error.message || 'Unknown error' };
    }
  },

  // Config operations
  getConfig: async () => {
    try {
      const config = await GetConfig();
      return { success: true as const, config };
    } catch (error: any) {
      return { success: false as const, error: error.message || 'Unknown error' };
    }
  },

  setTheme: async (theme: string) => {
    try {
      await SetTheme(theme);
      return { success: true };
    } catch (error: any) {
      return { success: false, error: error.message };
    }
  },

  setSidebarPosition: async (position: string) => {
    try {
      await SetSidebarPosition(position);
      return { success: true };
    } catch (error: any) {
      return { success: false, error: error.message };
    }
  },

  setWindowSize: async (width: number, height: number) => {
    try {
      await SetWindowSize(width, height);
      return { success: true };
    } catch (error: any) {
      return { success: false, error: error.message };
    }
  },

  // File operations
  getCatalogHtmlPath: async (catalogPath: string) => {
    try {
      const htmlPath = await GetCatalogHtmlPath(catalogPath);
      return { success: true as const, htmlPath };
    } catch (error: any) {
      return { success: false as const, error: error.message || 'Unknown error' };
    }
  },

  readHtmlFile: async (filePath: string) => {
    try {
      const content = await ReadHtmlFile(filePath);
      return { success: true as const, content };
    } catch (error: any) {
      return { success: false as const, error: error.message || 'Unknown error' };
    }
  },

  openExternal: async (url: string) => {
    try {
      await OpenExternal(url);
      return { success: true };
    } catch (error: any) {
      return { success: false, error: error.message };
    }
  },

  // Catalog files (alias for browseCatalogs)
  getCatalogFiles: async (catalogDir: string) => {
    try {
      const catalogs = await BrowseCatalogs(catalogDir);
      return { success: true, catalogs };
    } catch (error: any) {
      return { success: false, error: error.message || 'Unknown error' };
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
      return { success: false, error: error.message };
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
