const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  selectDirectory: () => ipcRenderer.invoke('select-directory'),
  selectSearchDirectory: () => ipcRenderer.invoke('select-search-directory'),
  selectOutputDirectory: () => ipcRenderer.invoke('select-output-directory'),
  createCatalog: (options) => ipcRenderer.invoke('create-catalog', options),
  searchCatalogs: (searchTerm, catalogDirectory) => ipcRenderer.invoke('search-catalogs', searchTerm, catalogDirectory),
  loadCatalog: (filePath) => ipcRenderer.invoke('load-catalog', filePath),
  getCatalogFiles: (directory) => ipcRenderer.invoke('get-catalog-files', directory),
  getCatalogHtmlPath: (catalogPath) => ipcRenderer.invoke('get-catalog-html-path', catalogPath),
  readHtmlFile: (filePath) => ipcRenderer.invoke('read-html-file', filePath),
  setWindowPersistence: (enabled) => ipcRenderer.invoke('set-window-persistence', enabled),
  getWindowPersistence: () => ipcRenderer.invoke('get-window-persistence')
});