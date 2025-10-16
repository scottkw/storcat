const { app, BrowserWindow, ipcMain, dialog, shell } = require('electron');

// V8 flags to work around ARM64 memory allocation issues
app.commandLine.appendSwitch('--js-flags', '--jitless --no-opt');
const path = require('path');
const fs = require('fs').promises;
const { createCatalog, searchCatalogs } = require('./catalog-service');

let mainWindow;

// Window state management
const DEFAULT_WINDOW_STATE = {
  width: 1200,
  height: 800,
  x: undefined,
  y: undefined
};

function getWindowStatePath() {
  return path.join(app.getPath('userData'), 'window-state.json');
}

async function saveWindowState() {
  if (!mainWindow || !windowPersistenceEnabled) return;
  
  try {
    const bounds = mainWindow.getBounds();
    const windowState = {
      width: bounds.width,
      height: bounds.height,
      x: bounds.x,
      y: bounds.y
    };
    
    await fs.writeFile(getWindowStatePath(), JSON.stringify(windowState, null, 2));
  } catch (error) {
    console.warn('Failed to save window state:', error.message);
  }
}

async function loadWindowState() {
  if (!windowPersistenceEnabled) {
    return DEFAULT_WINDOW_STATE;
  }
  
  try {
    const data = await fs.readFile(getWindowStatePath(), 'utf8');
    return JSON.parse(data);
  } catch (error) {
    return DEFAULT_WINDOW_STATE;
  }
}

function ensureWindowVisible(state) {
  // Ensure the window is visible on at least one screen
  const { screen } = require('electron');
  const displays = screen.getAllDisplays();
  
  if (state.x !== undefined && state.y !== undefined) {
    const windowVisible = displays.some(display => {
      const area = display.workArea;
      return state.x >= area.x && state.x < area.x + area.width &&
             state.y >= area.y && state.y < area.y + area.height;
    });
    
    if (!windowVisible) {
      // Reset position if window would be off-screen
      state.x = undefined;
      state.y = undefined;
    }
  }
  
  return state;
}

async function createWindow() {
  // Load preferences first
  await loadPreferences();
  
  // Load saved window state
  let windowState = await loadWindowState();
  windowState = ensureWindowVisible(windowState);
  
  const windowOptions = {
    width: windowState.width,
    height: windowState.height,
    minWidth: 800,
    minHeight: 600,
    icon: path.join(__dirname, 'assets/icons/icon.png'),
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    },
    titleBarStyle: 'hiddenInset',
    show: false
  };
  
  // Only set position if we have valid coordinates
  if (windowState.x !== undefined && windowState.y !== undefined) {
    windowOptions.x = windowState.x;
    windowOptions.y = windowState.y;
  }
  
  console.log('Creating main window with options:', windowOptions);
  mainWindow = new BrowserWindow(windowOptions);
  console.log('Main window created successfully');

  // Load React app - use dev server in development, built files in production
  const isDev = process.env.NODE_ENV === 'development';
  let indexPath;
  
  if (isDev) {
    console.log('Dev mode - Loading from Vite dev server');
    mainWindow.loadURL('http://localhost:5173');
  } else {
    // In production, try multiple possible locations
    const possiblePaths = [
      path.join(__dirname, 'index.html'),
      path.join(__dirname, '../dist/index.html'), 
      path.join(__dirname, 'dist/index.html'),
      path.join(process.resourcesPath, 'app.asar.unpacked/dist/index.html'),
      path.join(process.resourcesPath, 'dist/index.html')
    ];
    
    let foundPath = null;
    for (const testPath of possiblePaths) {
      console.log('Testing path:', testPath);
      if (require('fs').existsSync(testPath)) {
        foundPath = testPath;
        console.log('Found index.html at:', foundPath);
        break;
      }
    }
    
    if (foundPath) {
      // Use loadURL with file protocol to ensure proper asset loading
      const fileUrl = `file://${foundPath}`;
      console.log('Loading URL:', fileUrl);
      mainWindow.loadURL(fileUrl);
    } else {
      console.error('Could not find index.html in any expected location');
      console.log('__dirname:', __dirname);
      console.log('process.resourcesPath:', process.resourcesPath);
    }
  }
  
  mainWindow.webContents.on('did-fail-load', (event, errorCode, errorDescription) => {
    console.error('Failed to load:', errorCode, errorDescription);
  });
  
  mainWindow.webContents.on('did-finish-load', () => {
    console.log('Page finished loading');
  });

  mainWindow.once('ready-to-show', () => {
    console.log('Window ready-to-show event fired');
    mainWindow.show();
    console.log('Window show() called');
  });

  // Force show window after 3 seconds if ready-to-show doesn't fire
  setTimeout(() => {
    if (mainWindow && !mainWindow.isVisible()) {
      console.log('Force showing window after timeout');
      mainWindow.show();
    }
  }, 3000);
  
  // Save window state when it changes
  let saveTimeout;
  const debouncedSave = () => {
    clearTimeout(saveTimeout);
    saveTimeout = setTimeout(saveWindowState, 500); // Debounce saves
  };
  
  mainWindow.on('resize', debouncedSave);
  mainWindow.on('move', debouncedSave);
  
  // Save window state when closing
  mainWindow.on('close', () => {
    saveWindowState();
  });

  // Disable dev tools for now
  // mainWindow.webContents.openDevTools();
  
  // Only open dev tools if explicitly requested
  // if (process.argv.includes('--devtools')) {
  //   mainWindow.webContents.openDevTools();
  // }
}

app.whenReady().then(() => createWindow());

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});

// IPC handlers
ipcMain.handle('select-directory', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openDirectory'],
    title: 'Select Directory to Catalog'
  });
  
  if (!result.canceled && result.filePaths.length > 0) {
    return result.filePaths[0];
  }
  return null;
});

ipcMain.handle('select-search-directory', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openDirectory'],
    title: 'Select Directory Containing Catalog Files'
  });
  
  if (!result.canceled && result.filePaths.length > 0) {
    return result.filePaths[0];
  }
  return null;
});

ipcMain.handle('select-output-directory', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openDirectory'],
    title: 'Select Output Directory for Catalog Copy'
  });
  
  if (!result.canceled && result.filePaths.length > 0) {
    return result.filePaths[0];
  }
  return null;
});

ipcMain.handle('create-catalog', async (event, options) => {
  try {
    const result = await createCatalog(options);
    return { success: true, ...result };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('search-catalogs', async (event, searchTerm, catalogDirectory) => {
  try {
    const results = await searchCatalogs(searchTerm, catalogDirectory);
    return { success: true, results };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('load-catalog', async (event, filePath) => {
  try {
    const content = await fs.readFile(filePath, 'utf8');
    const catalog = JSON.parse(content);
    return { success: true, catalog };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('get-catalog-files', async (event, directory) => {
  try {
    const files = await fs.readdir(directory);
    const catalogFiles = files
      .filter(file => file.endsWith('.json'))
      .map(file => path.join(directory, file));
    
    const catalogInfo = await Promise.all(
      catalogFiles.map(async (filePath) => {
        try {
          const stats = await fs.stat(filePath);
          const content = await fs.readFile(filePath, 'utf8');
          const catalog = JSON.parse(content);
          
          // Check if HTML file exists and extract title
          const htmlPath = filePath.replace('.json', '.html');
          let title = path.basename(filePath, '.json'); // Default to filename
          let hasHtml = false;
          
          try {
            await fs.access(htmlPath);
            hasHtml = true;
            
            // Extract title from HTML file
            const htmlContent = await fs.readFile(htmlPath, 'utf8');
            const titleMatch = htmlContent.match(/<title>(.*?)<\/title>/i);
            if (titleMatch && titleMatch[1]) {
              title = titleMatch[1].trim();
            }
          } catch (error) {
            // HTML file doesn't exist or can't be read, use default title
          }
          
          return {
            path: filePath,
            name: path.basename(filePath, '.json'),
            title: title,
            size: stats.size,
            modified: stats.mtime,
            hasHtml: hasHtml
          };
        } catch (error) {
          return null;
        }
      })
    );
    
    return { success: true, catalogs: catalogInfo.filter(Boolean) };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('get-catalog-html-path', async (event, catalogPath) => {
  try {
    const htmlPath = catalogPath.replace('.json', '.html');
    
    // Check if HTML file exists
    await fs.access(htmlPath);
    
    return { success: true, htmlPath };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('read-html-file', async (event, filePath) => {
  try {
    const content = await fs.readFile(filePath, 'utf8');
    return { success: true, content };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

// Window persistence setting management
let windowPersistenceEnabled = true; // Default to enabled

function getPreferencesPath() {
  return path.join(app.getPath('userData'), 'preferences.json');
}

async function loadPreferences() {
  try {
    const data = await fs.readFile(getPreferencesPath(), 'utf8');
    const prefs = JSON.parse(data);
    windowPersistenceEnabled = prefs.windowPersistence !== false; // Default to true
  } catch (error) {
    windowPersistenceEnabled = true; // Default to enabled
  }
}

async function savePreferences() {
  try {
    const prefs = {
      windowPersistence: windowPersistenceEnabled
    };
    await fs.writeFile(getPreferencesPath(), JSON.stringify(prefs, null, 2));
  } catch (error) {
    console.warn('Failed to save preferences:', error.message);
  }
}

ipcMain.handle('set-window-persistence', async (event, enabled) => {
  windowPersistenceEnabled = enabled;
  
  // Save the preference
  await savePreferences();
  
  // If disabled, optionally clear any saved window state
  if (!enabled) {
    try {
      const windowStatePath = getWindowStatePath();
      await fs.unlink(windowStatePath);
    } catch (error) {
      // File might not exist, which is fine
    }
  }
  
  return { success: true };
});

ipcMain.handle('get-window-persistence', async () => {
  return { success: true, enabled: windowPersistenceEnabled };
});

// Open external URLs
ipcMain.handle('open-external', async (event, url) => {
  try {
    await shell.openExternal(url);
    return { success: true };
  } catch (error) {
    console.error('Failed to open external URL:', error);
    return { success: false, error: error.message };
  }
});