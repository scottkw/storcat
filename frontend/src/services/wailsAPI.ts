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
  SetDensity,
  SetSettingsMigrated,
  SetRailSide,
  SetCatalogDirectory,
  SetDefaultFilenameRoot,
  SetSecondaryDirectory,
  SetWriteHTML,
  SetCopyToSecondary,
  SetWatchDirectory,
  SelectDirectory,
  ReadHtmlFile,
  GetCatalogHtmlPath,
  RenameCatalog,
  DuplicateCatalog,
  DeleteCatalog,
  OpenExternal,
  GetWindowPersistence,
  SetWindowPersistence,
  GetVersion,
  StartScan,
  CancelScan,
  WritePartialCatalog,
  ListVolumes,
  RescanCatalog,
  ResolveRescan,
} from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import type { ResolveMode } from '../types/rescan';

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

  // sourcePath (walked) and outputDir (written) are genuinely distinct --
  // the create slide-over always passes the app's configured catalog
  // directory as outputDir, never the scanned source.
  startScan: async (
    title: string,
    sourcePath: string,
    outputDir: string,
    outputRoot: string,
    opts: { writeHTML: boolean; includeHidden: boolean; copyToDirectory: string; totalBytesHint: number }
  ) => {
    try {
      const result = await StartScan(title, sourcePath, outputDir, outputRoot, opts as main.ScanOptions);
      return { success: true as const, result };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Cancels the in-flight scan, if any -- a no-op server-side when nothing
  // is running, so this call is always safe to fire from the panel's
  // cancel button without first checking scan state.
  cancelScan: async () => {
    try {
      await CancelScan();
      return { success: true as const };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Writes the tree retained from a source-loss scan through the shared
  // write path. Idempotent server-side: a second call returns the first
  // call's cached result without touching the filesystem again.
  writePartialCatalog: async () => {
    try {
      const result = await WritePartialCatalog();
      return { success: true as const, result };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Walks sourcePath directly (no write, ever) and, when oldTreeAvailable,
  // diffs the result against jsonPath's already-on-disk catalog. Reuses the
  // same scanMu one-scan-at-a-time guard as startScan -- progress arrives
  // through the same shared scan:progress event CreateSlideOver's listener
  // already forwards into state.scan, not a second subscription.
  rescanCatalog: async (jsonPath: string, sourcePath: string, oldTreeAvailable: boolean) => {
    try {
      const diff = await RescanCatalog(jsonPath, sourcePath, oldTreeAvailable);
      return { success: true as const, diff };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Writes the tree a just-completed re-scan diffed to disk -- overwrite in
  // place, or keep-both alongside the original via the same "-copy"/
  // "-copy-N" collision loop Duplicate already uses. mode is bridged as a
  // plain string (ResolveMode's own doc comment explains why); "discard"
  // has no Go call at all and is never passed here.
  resolveRescan: async (jsonPath: string, catalogDir: string, mode: ResolveMode) => {
    try {
      const result = await ResolveRescan(jsonPath, catalogDir, mode);
      return { success: true as const, result };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Lists the machine's currently mounted volumes (name, mount path,
  // size, free space, readable flag) for the create flow's volume-card
  // picker. Always resolves to an array, even when nothing is mounted.
  listVolumes: async () => {
    try {
      const result = await ListVolumes();
      return { success: true as const, volumes: result };
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

  setDensity: async (density: string) => {
    try {
      await SetDensity(density);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setSettingsMigrated: async (migrated: boolean) => {
    try {
      await SetSettingsMigrated(migrated);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setRailSide: async (side: string) => {
    try {
      await SetRailSide(side);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setCatalogDirectory: async (dir: string) => {
    try {
      await SetCatalogDirectory(dir);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setDefaultFilenameRoot: async (root: string) => {
    try {
      await SetDefaultFilenameRoot(root);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setSecondaryDirectory: async (dir: string) => {
    try {
      await SetSecondaryDirectory(dir);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setWriteHTML: async (enabled: boolean) => {
    try {
      await SetWriteHTML(enabled);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setCopyToSecondary: async (enabled: boolean) => {
    try {
      await SetCopyToSecondary(enabled);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  setWatchDirectory: async (enabled: boolean) => {
    try {
      await SetWatchDirectory(enabled);
      return { success: true };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // File operations. catalogDir is the frontend's currently configured
  // catalog directory -- threaded through so the Go side can reject any
  // path that does not resolve inside it (FU-23-A: both bindings are
  // callable from any renderer JS, not only these call sites), exactly as
  // revealInFileManager already does below.
  getCatalogHtmlPath: async (catalogPath: string, catalogDir: string) => {
    try {
      const htmlPath = await GetCatalogHtmlPath(catalogPath, catalogDir);
      return { success: true as const, htmlPath };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // catalogDir is the frontend's currently configured catalog directory --
  // threaded through so the Go side can reject any path that does not
  // resolve inside it (T-27-02), exactly as getCatalogHtmlPath already does
  // above.
  renameCatalog: async (jsonPath: string, catalogDir: string, newTitle: string) => {
    try {
      await RenameCatalog(jsonPath, catalogDir, newTitle);
      return { success: true as const };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // catalogDir is the frontend's currently configured catalog directory --
  // threaded through so the Go side can reject any path that does not
  // resolve inside it (T-27-02), exactly as renameCatalog already does
  // above. jsonPath is the resulting new catalog's .json path.
  duplicateCatalog: async (jsonPath: string, catalogDir: string) => {
    try {
      const newPath = await DuplicateCatalog(jsonPath, catalogDir);
      return { success: true as const, jsonPath: newPath };
    } catch (error: any) {
      return wailsError(error);
    }
  },

  // Moves jsonPath -- and, when deleteHtml is true, its .html companion --
  // to the OS Trash. The .html path is always derived on the Go side, never
  // accepted here, so this call can never name an arbitrary second file for
  // deletion.
  deleteCatalog: async (jsonPath: string, catalogDir: string, deleteHtml: boolean) => {
    try {
      await DeleteCatalog(jsonPath, catalogDir, deleteHtml);
      return { success: true as const };
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

  openExternal: async (url: string, catalogDir: string) => {
    try {
      await OpenExternal(url, catalogDir);
      return { success: true as const };
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
