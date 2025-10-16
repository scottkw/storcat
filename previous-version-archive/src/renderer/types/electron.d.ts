export interface ElectronAPI {
  selectDirectory: () => Promise<string | null>;
  selectSearchDirectory: () => Promise<string | null>;
  selectOutputDirectory: () => Promise<string | null>;
  createCatalog: (options: {
    title: string;
    directoryPath: string;
    outputRoot: string;
    copyToDirectory?: string | null;
  }) => Promise<{
    success: boolean;
    jsonPath?: string;
    htmlPath?: string;
    copyJsonPath?: string;
    copyHtmlPath?: string;
    fileCount?: number;
    totalSize?: number;
    error?: string;
  }>;
  searchCatalogs: (searchTerm: string, catalogDirectory: string) => Promise<{
    success: boolean;
    results?: any[];
    error?: string;
  }>;
  loadCatalog: (filePath: string) => Promise<{
    success: boolean;
    catalog?: any;
    error?: string;
  }>;
  getCatalogFiles: (directory: string) => Promise<{
    success: boolean;
    catalogs?: Array<{
      path: string;
      name: string;
      title: string;
      size: number;
      modified: Date;
      hasHtml: boolean;
    }>;
    error?: string;
  }>;
  getCatalogHtmlPath: (catalogPath: string) => Promise<{
    success: boolean;
    htmlPath?: string;
    error?: string;
  }>;
  readHtmlFile: (filePath: string) => Promise<{
    success: boolean;
    content?: string;
    error?: string;
  }>;
  setWindowPersistence: (enabled: boolean) => Promise<{ success: boolean }>;
  getWindowPersistence: () => Promise<{ success: boolean; enabled: boolean }>;
}

declare global {
  interface Window {
    electronAPI: ElectronAPI;
  }
}