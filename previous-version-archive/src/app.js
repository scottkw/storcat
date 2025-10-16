// Application state
let appState = {
  selectedDirectory: null,
  selectedOutputDirectory: null,
  selectedSearchDirectory: null,
  selectedBrowseDirectory: null,
  isCreating: false,
  isSearching: false,
  isLoading: false,
  searchResults: [],
  sortColumn: null,
  sortDirection: 'asc',
  browseCatalogs: [],
  browseSortColumn: null,
  browseSortDirection: 'asc',
  sidebarCollapsed: false,
  isDarkMode: false
};

// DOM elements
const elements = {
  // Create tab
  catalogTitle: document.getElementById('catalog-title'),
  selectDirectoryBtn: document.getElementById('select-directory-btn'),
  selectedDirectoryPath: document.getElementById('selected-directory-path'),
  outputRoot: document.getElementById('output-root'),
  copyToSecondary: document.getElementById('copy-to-secondary'),
  secondaryLocationGroup: document.getElementById('secondary-location-group'),
  selectOutputBtn: document.getElementById('select-output-btn'),
  selectedOutputPath: document.getElementById('selected-output-path'),
  createCatalogBtn: document.getElementById('create-catalog-btn'),
  createStatus: document.getElementById('create-status'),
  
  // Search tab
  searchTerm: document.getElementById('search-term'),
  selectSearchDirectoryBtn: document.getElementById('select-search-directory-btn'),
  selectedSearchPath: document.getElementById('selected-search-path'),
  searchCatalogsBtn: document.getElementById('search-catalogs-btn'),
  searchStatus: document.getElementById('search-status'),
  
  // Browse tab
  selectBrowseDirectoryBtn: document.getElementById('select-browse-directory-btn'),
  selectedBrowsePath: document.getElementById('selected-browse-path'),
  loadCatalogsBtn: document.getElementById('load-catalogs-btn'),
  browseStatus: document.getElementById('browse-status'),
  
  // Main content
  mainDisplay: document.getElementById('main-display'),
  
  // Sidebar and tabs
  sidebar: document.getElementById('sidebar'),
  sidebarToggle: document.getElementById('sidebar-toggle'),
  toggleContainer: document.getElementById('toggle-container'),
  mainTabs: document.getElementById('main-tabs'),
  createSidebar: document.getElementById('create-sidebar'),
  searchSidebar: document.getElementById('search-sidebar'),
  browseSidebar: document.getElementById('browse-sidebar'),
  
  // Modal
  catalogModal: document.getElementById('catalog-modal'),
  modalTitle: document.getElementById('modal-title'),
  modalIframe: document.getElementById('modal-iframe'),
  modalClose: document.getElementById('modal-close'),
  
  // Settings modal
  settingsToggle: document.getElementById('settings-toggle'),
  settingsModal: document.getElementById('settings-modal'),
  settingsModalClose: document.getElementById('settings-modal-close'),
  themeSwitch: document.getElementById('theme-switch'),
  windowPersistenceSwitch: document.getElementById('window-persistence-switch')
};

// Utility functions
function showElement(element) {
  element.classList.remove('hidden');
}

function hideElement(element) {
  element.classList.add('hidden');
}

function showStatus(container, message, type = 'success') {
  container.innerHTML = `<div class="status-message status-${type}">${message}</div>`;
  showElement(container);
}

function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 10) / 10 + ' ' + sizes[i];
}

function updateCreateButton() {
  const hasDirectory = appState.selectedDirectory;
  const hasTitle = elements.catalogTitle && elements.catalogTitle.value ? elements.catalogTitle.value.trim() : '';
  const hasRoot = elements.outputRoot && elements.outputRoot.value ? elements.outputRoot.value.trim() : '';
  
  if (elements.createCatalogBtn) {
    elements.createCatalogBtn.disabled = !hasDirectory || !hasTitle || !hasRoot || appState.isCreating;
    elements.createCatalogBtn.loading = appState.isCreating;
  }
}

function updateSearchButton() {
  const hasSearchTerm = elements.searchTerm && elements.searchTerm.value ? elements.searchTerm.value.trim() : '';
  const hasSearchDirectory = appState.selectedSearchDirectory;
  
  if (elements.searchCatalogsBtn) {
    elements.searchCatalogsBtn.disabled = !hasSearchTerm || !hasSearchDirectory || appState.isSearching;
  }
}

function updateLoadButton() {
  const hasBrowseDirectory = appState.selectedBrowseDirectory;
  
  if (elements.loadCatalogsBtn) {
    elements.loadCatalogsBtn.disabled = !hasBrowseDirectory || appState.isLoading;
  }
}

// Event listeners
elements.selectDirectoryBtn.addEventListener('click', async () => {
  const directory = await window.electronAPI.selectDirectory();
  if (directory) {
    appState.selectedDirectory = directory;
    elements.selectedDirectoryPath.textContent = directory;
    showElement(elements.selectedDirectoryPath);
    updateCreateButton();
    localStorage.setItem('storcat-last-create-directory', directory);
  }
});

elements.selectOutputBtn.addEventListener('click', async () => {
  const directory = await window.electronAPI.selectOutputDirectory();
  if (directory) {
    appState.selectedOutputDirectory = directory;
    elements.selectedOutputPath.textContent = directory;
    showElement(elements.selectedOutputPath);
    localStorage.setItem('storcat-last-output-directory', directory);
  }
});

elements.selectSearchDirectoryBtn.addEventListener('click', async () => {
  const directory = await window.electronAPI.selectSearchDirectory();
  if (directory) {
    appState.selectedSearchDirectory = directory;
    appState.selectedBrowseDirectory = directory; // Also update browse directory
    elements.selectedSearchPath.textContent = directory;
    elements.selectedBrowsePath.textContent = directory; // Also update browse path display
    showElement(elements.selectedSearchPath);
    showElement(elements.selectedBrowsePath);
    updateSearchButton();
    updateLoadButton(); // Also update browse button
    localStorage.setItem('storcat-last-catalog-directory', directory);
  }
});

elements.selectBrowseDirectoryBtn.addEventListener('click', async () => {
  const directory = await window.electronAPI.selectSearchDirectory();
  if (directory) {
    appState.selectedBrowseDirectory = directory;
    appState.selectedSearchDirectory = directory; // Also update search directory
    elements.selectedBrowsePath.textContent = directory;
    elements.selectedSearchPath.textContent = directory; // Also update search path display
    showElement(elements.selectedBrowsePath);
    showElement(elements.selectedSearchPath);
    updateLoadButton();
    updateSearchButton(); // Also update search button
    localStorage.setItem('storcat-last-catalog-directory', directory);
  }
});

elements.copyToSecondary.addEventListener('sl-change', (e) => {
  if (e.target.checked) {
    showElement(elements.secondaryLocationGroup);
  } else {
    hideElement(elements.secondaryLocationGroup);
    appState.selectedOutputDirectory = null;
    hideElement(elements.selectedOutputPath);
  }
});

elements.catalogTitle.addEventListener('sl-input', updateCreateButton);
elements.outputRoot.addEventListener('sl-input', updateCreateButton);
elements.searchTerm.addEventListener('sl-input', () => {
  updateSearchButton();
  // Save search term to localStorage
  const searchTerm = elements.searchTerm.value.trim();
  if (searchTerm) {
    localStorage.setItem('storcat-last-search-term', searchTerm);
  } else {
    localStorage.removeItem('storcat-last-search-term');
  }
});

elements.createCatalogBtn.addEventListener('click', async () => {
  if (appState.isCreating) return;
  
  appState.isCreating = true;
  updateCreateButton();
  hideElement(elements.createStatus);
  
  const options = {
    title: elements.catalogTitle.value.trim(),
    directoryPath: appState.selectedDirectory,
    outputRoot: elements.outputRoot.value.trim(),
    copyToDirectory: elements.copyToSecondary.checked ? appState.selectedOutputDirectory : null
  };
  
  try {
    const result = await window.electronAPI.createCatalog(options);
    
    if (result.success) {
      // Show success in modal
      elements.modalTitle.textContent = 'Catalog Created Successfully';
      
      let message = `<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6;">`;
      message += `<p><strong>Catalog created successfully!</strong></p>`;
      message += `<p><strong>JSON:</strong> ${result.jsonPath}</p>`;
      message += `<p><strong>HTML:</strong> ${result.htmlPath}</p>`;
      message += `<p><strong>Files:</strong> ${result.fileCount.toLocaleString()}</p>`;
      message += `<p><strong>Total Size:</strong> ${formatBytes(result.totalSize)}</p>`;
      
      if (result.copyJsonPath) {
        message += `<br><p><strong>Copies created:</strong></p>`;
        message += `<p><strong>JSON:</strong> ${result.copyJsonPath}</p>`;
        message += `<p><strong>HTML:</strong> ${result.copyHtmlPath}</p>`;
      }
      message += `</div>`;
      
      elements.modalIframe.style.display = 'none';
      const modalContent = elements.catalogModal.querySelector('.modal-content');
      modalContent.classList.add('compact'); // Make modal smaller for text content
      
      let existingDiv = modalContent.querySelector('.catalog-result-content');
      if (!existingDiv) {
        existingDiv = document.createElement('div');
        existingDiv.className = 'catalog-result-content';
        existingDiv.style.padding = '20px';
        existingDiv.style.overflow = 'auto';
        existingDiv.style.maxHeight = '400px';
        modalContent.appendChild(existingDiv);
      }
      existingDiv.innerHTML = message;
      existingDiv.style.display = 'block';
      
      showModal();
    } else {
      showStatus(elements.createStatus, `Error: ${result.error}`, 'error');
    }
  } catch (error) {
    showStatus(elements.createStatus, `Error: ${error.message}`, 'error');
  } finally {
    appState.isCreating = false;
    updateCreateButton();
  }
});

elements.searchCatalogsBtn.addEventListener('click', async () => {
  await executeSearch(true);
});

elements.loadCatalogsBtn.addEventListener('click', async () => {
  await loadBrowseCatalogs(true);
});

function displaySearchResults(results) {
  if (results.length === 0) {
    elements.mainDisplay.innerHTML = `
      <div class="section-card">
        <h2 class="section-title">Search Results</h2>
        <p>No results found for your search term.</p>
      </div>
    `;
    return;
  }
  
  // Store results for sorting
  appState.searchResults = results;
  
  // Sort results if a sort column is selected
  const sortedResults = sortResults(results, appState.sortColumn, appState.sortDirection);
  
  let html = `
    <div class="section-card">
      <h2 class="section-title">Search Results (${results.length} found)</h2>
      <div class="results-container">
        <table class="results-table" id="results-table">
          <thead>
            <tr>
              <th class="sortable" data-column="basename" style="width: 25%">
                <span>File/Directory</span>
                <span class="sort-indicator"></span>
                <div class="column-resizer" data-column="0"></div>
              </th>
              <th class="sortable" data-column="fullPath" style="width: 30%">
                <span>Path</span>
                <span class="sort-indicator"></span>
                <div class="column-resizer" data-column="1"></div>
              </th>
              <th class="sortable" data-column="type" style="width: 15%">
                <span>Type</span>
                <span class="sort-indicator"></span>
                <div class="column-resizer" data-column="2"></div>
              </th>
              <th class="sortable" data-column="size" style="width: 15%">
                <span>Size</span>
                <span class="sort-indicator"></span>
                <div class="column-resizer" data-column="3"></div>
              </th>
              <th class="sortable" data-column="catalog" style="width: 15%">
                <span>Catalog</span>
                <span class="sort-indicator"></span>
                <div class="column-resizer" data-column="4"></div>
              </th>
            </tr>
          </thead>
          <tbody>
  `;
  
  sortedResults.forEach(result => {
    const displayPath = result.fullPath || '/';
    html += `
      <tr>
        <td title="${escapeHtml(result.basename)}">
          <div class="file-basename">${escapeHtml(result.basename)}</div>
        </td>
        <td title="${escapeHtml(displayPath)}">
          <div class="file-path">${escapeHtml(displayPath)}</div>
        </td>
        <td>
          <sl-badge variant="${result.type === 'directory' ? 'primary' : 'neutral'}">
            ${result.type}
          </sl-badge>
        </td>
        <td title="${formatBytes(result.size)}">${formatBytes(result.size)}</td>
        <td title="${escapeHtml(result.catalog)}">
          <a href="#" class="catalog-link" onclick="openCatalogHtml('${escapeHtml(result.catalogFilePath)}')">
            ${escapeHtml(result.catalog)}
          </a>
        </td>
      </tr>
    `;
  });
  
  html += `
          </tbody>
        </table>
      </div>
    </div>
  `;
  
  elements.mainDisplay.innerHTML = html;
  
  // Initialize table functionality
  initializeTable();
}

function displayCatalogList(catalogs) {
  if (catalogs.length === 0) {
    elements.mainDisplay.innerHTML = `
      <div class="section-card">
        <h2 class="section-title">Catalog Browser</h2>
        <p>No catalog files found in the selected directory.</p>
      </div>
    `;
    return;
  }
  
  // Store catalogs for sorting
  appState.browseCatalogs = catalogs;
  
  // Sort catalogs if a sort column is selected
  const sortedCatalogs = sortBrowseCatalogs(catalogs, appState.browseSortColumn, appState.browseSortDirection);
  
  let html = `
    <div class="section-card">
      <h2 class="section-title">Catalog Browser (${catalogs.length} catalogs)</h2>
      <div class="results-container">
        <table class="results-table" id="catalog-browse-table">
          <thead>
            <tr>
              <th class="sortable" data-column="title" style="width: 35%">
                <span>Catalog Title</span>
                <span class="sort-indicator"></span>
                <div class="column-resizer" data-column="0"></div>
              </th>
              <th class="sortable" data-column="name" style="width: 35%">
                <span>Catalog File</span>
                <span class="sort-indicator"></span>
                <div class="column-resizer" data-column="1"></div>
              </th>
              <th class="sortable" data-column="modified" style="width: 30%">
                <span>Date Created</span>
                <span class="sort-indicator"></span>
                <div class="column-resizer" data-column="2"></div>
              </th>
            </tr>
          </thead>
          <tbody>
  `;
  
  sortedCatalogs.forEach(catalog => {
    const modifiedDate = new Date(catalog.modified).toLocaleString();
    const catalogTitle = catalog.title || catalog.name.replace('.json', '');
    const fileName = catalog.path.split('/').pop();
    
    html += `
      <tr>
        <td title="${escapeHtml(catalogTitle)}">
          ${catalog.hasHtml ? 
            `<a href="#" class="catalog-link" onclick="openCatalogHtmlFromPath('${escapeHtml(catalog.path)}')">${escapeHtml(catalogTitle)}</a>` :
            `<span style="color: var(--sl-color-neutral-600)">${escapeHtml(catalogTitle)} (No HTML)</span>`
          }
        </td>
        <td title="${escapeHtml(fileName)}">
          <div class="file-path">${escapeHtml(fileName)}</div>
        </td>
        <td title="${modifiedDate}">
          ${modifiedDate}
        </td>
      </tr>
    `;
  });
  
  html += `
          </tbody>
        </table>
      </div>
    </div>
  `;
  
  elements.mainDisplay.innerHTML = html;
  
  // Initialize table functionality after rendering
  initializeTable();
  
  // Update sort indicators if there's an active sort
  if (appState.browseSortColumn) {
    const table = document.getElementById('catalog-browse-table');
    if (table) {
      const headers = table.querySelectorAll('th.sortable');
      headers.forEach(header => {
        if (header.dataset.column === appState.browseSortColumn) {
          header.classList.add(`sort-${appState.browseSortDirection}`);
        }
      });
    }
  }
}

async function loadCatalogContent(filePath) {
  try {
    const result = await window.electronAPI.loadCatalog(filePath);
    
    if (result.success) {
      displayCatalogContent(result.catalog, filePath);
    } else {
      showStatus(elements.browseStatus, `Error loading catalog: ${result.error}`, 'error');
    }
  } catch (error) {
    showStatus(elements.browseStatus, `Error loading catalog: ${error.message}`, 'error');
  }
}

function displayCatalogContent(catalog, filePath) {
  const fileName = filePath.split('/').pop().replace('.json', '');
  
  let html = `
    <div class="section-card">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h2 class="section-title">Catalog: ${escapeHtml(fileName)}</h2>
        <sl-button size="small" onclick="displayCatalogList([])">
          <sl-icon slot="prefix" name="arrow-left"></sl-icon>
          Back to List
        </sl-button>
      </div>
      <div class="results-container" style="max-height: 600px;">
        ${renderCatalogTree(catalog)}
      </div>
    </div>
  `;
  
  elements.mainDisplay.innerHTML = html;
}

function renderCatalogTree(item, level = 0) {
  const indent = '&nbsp;'.repeat(level * 4);
  let html = '';
  
  if (item.type === 'directory') {
    html += `
      <div class="search-result">
        <div class="result-info">
          <div class="result-name">${indent}📁 ${escapeHtml(item.name)}</div>
        </div>
        <div class="result-meta">
          <sl-badge variant="primary">directory</sl-badge>
          <span>${formatBytes(item.size)}</span>
        </div>
      </div>
    `;
    
    if (item.contents) {
      item.contents.forEach(child => {
        html += renderCatalogTree(child, level + 1);
      });
    }
  } else {
    html += `
      <div class="search-result">
        <div class="result-info">
          <div class="result-name">${indent}📄 ${escapeHtml(item.name)}</div>
        </div>
        <div class="result-meta">
          <sl-badge variant="neutral">file</sl-badge>
          <span>${formatBytes(item.size)}</span>
        </div>
      </div>
    `;
  }
  
  return html;
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function modifyHtmlForTheme(htmlContent) {
  // If in dark mode, modify the CSS to use light text on dark background
  if (appState.isDarkMode) {
    // Replace black text colors with white
    htmlContent = htmlContent.replace(/color:\s*black/g, 'color: white');
    htmlContent = htmlContent.replace(/color:\s*#000000/g, 'color: #ffffff');
    
    // Replace transparent backgrounds with dark background
    htmlContent = htmlContent.replace(/background-color:\s*transparent/g, 'background-color: #1e293b');
    htmlContent = htmlContent.replace(/background-color:\s*white/g, 'background-color: #1e293b');
    
    // Update body background
    htmlContent = htmlContent.replace(
      /BODY\s*{\s*([^}]*)\s*}/,
      'BODY { font-family: arial, monospace, sans-serif; background-color: #1e293b; color: white; margin: 16px; }'
    );
    
    // Adjust colors for better contrast in dark mode
    htmlContent = htmlContent.replace(/color:\s*blue/g, 'color: #60a5fa');     // Lighter blue
    htmlContent = htmlContent.replace(/color:\s*#0066cc/g, 'color: #60a5fa');  // Lighter blue
    htmlContent = htmlContent.replace(/color:\s*purple/g, 'color: #c084fc');   // Lighter purple
    htmlContent = htmlContent.replace(/color:\s*green/g, 'color: #4ade80');    // Lighter green
    htmlContent = htmlContent.replace(/color:\s*#006400/g, 'color: #4ade80');  // Lighter green
    htmlContent = htmlContent.replace(/color:\s*#b8860b/g, 'color: #fbbf24');  // Lighter gold
    htmlContent = htmlContent.replace(/color:\s*#008b8b/g, 'color: #22d3ee');  // Lighter cyan
    htmlContent = htmlContent.replace(/color:\s*#8b008b/g, 'color: #e879f9');  // Lighter magenta
    
    // Update hover background
    htmlContent = htmlContent.replace(/background-color:\s*yellow/g, 'background-color: #fbbf24');
  }
  
  return htmlContent;
}

async function openCatalogHtml(catalogPath) {
  try {
    const result = await window.electronAPI.getCatalogHtmlPath(catalogPath);
    if (result.success) {
      const catalogName = catalogPath.split('/').pop().replace('.json', '');
      elements.modalTitle.textContent = `Catalog: ${catalogName}`;
      
      // Read and modify HTML content based on theme
      const htmlResult = await window.electronAPI.readHtmlFile(result.htmlPath);
      if (!htmlResult.success) {
        console.error('Failed to read HTML file:', htmlResult.error);
        return;
      }
      const modifiedHtml = modifyHtmlForTheme(htmlResult.content);
      
      // Show iframe and hide result content
      elements.modalIframe.style.display = 'block';
      const modalContent = elements.catalogModal.querySelector('.modal-content');
      modalContent.classList.remove('compact'); // Use full size for HTML content
      const existingDiv = modalContent.querySelector('.catalog-result-content');
      if (existingDiv) {
        existingDiv.style.display = 'none';
      }
      
      // Create data URL and load it
      const dataUrl = `data:text/html;charset=utf-8,${encodeURIComponent(modifiedHtml)}`;
      elements.modalIframe.src = dataUrl;
      showModal();
    } else {
      console.error('Failed to get catalog HTML path:', result.error);
    }
  } catch (error) {
    console.error('Error opening catalog HTML:', error);
  }
}

async function openCatalogHtmlFromPath(catalogPath) {
  try {
    const result = await window.electronAPI.getCatalogHtmlPath(catalogPath);
    if (result.success) {
      const catalogName = catalogPath.split('/').pop().replace('.json', '');
      elements.modalTitle.textContent = `Catalog: ${catalogName}`;
      
      // Read and modify HTML content based on theme
      const htmlResult = await window.electronAPI.readHtmlFile(result.htmlPath);
      if (!htmlResult.success) {
        console.error('Failed to read HTML file:', htmlResult.error);
        return;
      }
      const modifiedHtml = modifyHtmlForTheme(htmlResult.content);
      
      // Show iframe and hide result content
      elements.modalIframe.style.display = 'block';
      const modalContent = elements.catalogModal.querySelector('.modal-content');
      modalContent.classList.remove('compact'); // Use full size for HTML content
      const existingDiv = modalContent.querySelector('.catalog-result-content');
      if (existingDiv) {
        existingDiv.style.display = 'none';
      }
      
      // Create data URL and load it
      const dataUrl = `data:text/html;charset=utf-8,${encodeURIComponent(modifiedHtml)}`;
      elements.modalIframe.src = dataUrl;
      showModal();
    } else {
      console.error('Failed to get catalog HTML path:', result.error);
    }
  } catch (error) {
    console.error('Error opening catalog HTML:', error);
  }
}

// Modal functionality
function showModal() {
  elements.catalogModal.classList.add('show');
  document.body.style.overflow = 'hidden'; // Prevent background scrolling
}

function hideModal() {
  elements.catalogModal.classList.remove('show');
  document.body.style.overflow = ''; // Restore scrolling
  elements.modalIframe.src = 'about:blank'; // Clear iframe
  
  // Clear catalog result content if it exists
  const modalContent = elements.catalogModal.querySelector('.modal-content');
  modalContent.classList.remove('compact'); // Reset modal size
  const existingDiv = modalContent.querySelector('.catalog-result-content');
  if (existingDiv) {
    existingDiv.style.display = 'none';
  }
  // Show iframe by default for next use
  elements.modalIframe.style.display = 'block';
}

// Sorting functionality
function sortResults(results, column, direction) {
  if (!column) return results;
  
  const sortedResults = [...results].sort((a, b) => {
    let aVal = a[column];
    let bVal = b[column];
    
    // Handle size column specially (numeric sorting)
    if (column === 'size') {
      return direction === 'asc' ? aVal - bVal : bVal - aVal;
    }
    
    // String sorting
    if (typeof aVal === 'string') aVal = aVal.toLowerCase();
    if (typeof bVal === 'string') bVal = bVal.toLowerCase();
    
    if (aVal < bVal) return direction === 'asc' ? -1 : 1;
    if (aVal > bVal) return direction === 'asc' ? 1 : -1;
    return 0;
  });
  
  return sortedResults;
}

function sortBrowseCatalogs(catalogs, column, direction) {
  if (!column) return catalogs;
  
  const sortedCatalogs = [...catalogs].sort((a, b) => {
    let aVal, bVal;
    
    // Handle different column types
    if (column === 'title') {
      aVal = (a.title || a.name).toLowerCase();
      bVal = (b.title || b.name).toLowerCase();
    } else if (column === 'name') {
      aVal = a.name.toLowerCase();
      bVal = b.name.toLowerCase();
    } else if (column === 'modified') {
      aVal = new Date(a.modified);
      bVal = new Date(b.modified);
      return direction === 'asc' ? aVal - bVal : bVal - aVal;
    } else {
      aVal = a[column];
      bVal = b[column];
    }
    
    // String sorting
    if (typeof aVal === 'string' && typeof bVal === 'string') {
      if (aVal < bVal) return direction === 'asc' ? -1 : 1;
      if (aVal > bVal) return direction === 'asc' ? 1 : -1;
      return 0;
    }
    
    return 0;
  });
  
  return sortedCatalogs;
}

// Initialize table functionality
function initializeTable() {
  const searchTable = document.getElementById('results-table');
  const browseTable = document.getElementById('catalog-browse-table');
  
  if (searchTable) {
    initializeSearchTable(searchTable);
  }
  
  if (browseTable) {
    initializeBrowseTable(browseTable);
  }
}

function initializeSearchTable(table) {
  // Add sorting functionality
  const headers = table.querySelectorAll('th.sortable');
  headers.forEach(header => {
    header.addEventListener('click', (e) => {
      if (e.target.classList.contains('column-resizer')) return;
      
      const column = header.dataset.column;
      
      // Toggle sort direction
      if (appState.sortColumn === column) {
        appState.sortDirection = appState.sortDirection === 'asc' ? 'desc' : 'asc';
      } else {
        appState.sortColumn = column;
        appState.sortDirection = 'asc';
      }
      
      // Update header classes
      headers.forEach(h => h.classList.remove('sort-asc', 'sort-desc'));
      header.classList.add(`sort-${appState.sortDirection}`);
      
      // Re-render table with new sort
      displaySearchResults(appState.searchResults);
    });
  });
  
  // Add column resizing functionality
  initializeColumnResizing();
}

function initializeBrowseTable(table) {
  // Add sorting functionality
  const headers = table.querySelectorAll('th.sortable');
  
  headers.forEach(header => {
    header.addEventListener('click', (e) => {
      if (e.target.classList.contains('column-resizer')) return;
      
      const column = header.dataset.column;
      
      // Toggle sort direction
      if (appState.browseSortColumn === column) {
        appState.browseSortDirection = appState.browseSortDirection === 'asc' ? 'desc' : 'asc';
      } else {
        appState.browseSortColumn = column;
        appState.browseSortDirection = 'asc';
      }
      
      // Update header classes
      headers.forEach(h => h.classList.remove('sort-asc', 'sort-desc'));
      header.classList.add(`sort-${appState.browseSortDirection}`);
      
      // Re-render table with new sort
      displayCatalogList(appState.browseCatalogs);
    });
  });
  
  // Add column resizing functionality
  initializeColumnResizing();
}

// Column resizing functionality
function initializeColumnResizing() {
  const resizers = document.querySelectorAll('.column-resizer');
  let isResizing = false;
  let currentResizer = null;
  let startX = 0;
  let startWidth = 0;
  
  resizers.forEach(resizer => {
    resizer.addEventListener('mousedown', (e) => {
      e.preventDefault();
      isResizing = true;
      currentResizer = resizer;
      startX = e.clientX;
      
      const th = resizer.closest('th');
      startWidth = parseInt(window.getComputedStyle(th).width, 10);
      
      resizer.classList.add('resizing');
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    });
  });
  
  document.addEventListener('mousemove', (e) => {
    if (!isResizing || !currentResizer) return;
    
    const diff = e.clientX - startX;
    const newWidth = startWidth + diff;
    
    if (newWidth > 50) { // Minimum column width
      const th = currentResizer.closest('th');
      const table = th.closest('table');
      const columnIndex = Array.from(th.parentNode.children).indexOf(th);
      
      // Update column width
      th.style.width = newWidth + 'px';
      
      // Update corresponding cells in the body
      const rows = table.querySelectorAll('tbody tr');
      rows.forEach(row => {
        const cell = row.children[columnIndex];
        if (cell) {
          cell.style.width = newWidth + 'px';
        }
      });
    }
  });
  
  document.addEventListener('mouseup', () => {
    if (isResizing) {
      isResizing = false;
      if (currentResizer) {
        currentResizer.classList.remove('resizing');
        currentResizer = null;
      }
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    }
  });
}

// Sidebar toggle functionality
function toggleSidebar() {
  appState.sidebarCollapsed = !appState.sidebarCollapsed;
  const sidebar = elements.sidebar;
  const toggle = elements.sidebarToggle;
  const toggleContainer = elements.toggleContainer;
  const mainContent = document.querySelector('.main-content');
  
  if (appState.sidebarCollapsed) {
    sidebar.classList.add('collapsed');
    toggleContainer.classList.add('collapsed');
    mainContent.classList.remove('sidebar-open');
    toggle.innerHTML = '<span>▶</span>';
    toggle.title = 'Show Sidebar';
  } else {
    sidebar.classList.remove('collapsed');
    toggleContainer.classList.remove('collapsed');
    mainContent.classList.add('sidebar-open');
    toggle.innerHTML = '<span>◀</span>';
    toggle.title = 'Hide Sidebar';
  }
}

// Settings modal management
function showSettingsModal() {
  elements.settingsModal.classList.add('show');
  document.body.style.overflow = 'hidden'; // Prevent background scrolling
}

function hideSettingsModal() {
  elements.settingsModal.classList.remove('show');
  document.body.style.overflow = ''; // Restore scrolling
}

// Theme management
function setTheme(isDark) {
  appState.isDarkMode = isDark;
  
  const lightSheet = document.getElementById('shoelace-light');
  const darkSheet = document.getElementById('shoelace-dark');
  const themeSwitch = elements.themeSwitch;
  const lightLabel = document.querySelector('.theme-switch-container .theme-label:first-child');
  const darkLabel = document.querySelector('.theme-switch-container .theme-label:last-child');
  
  if (isDark) {
    // Switch to dark mode
    document.documentElement.setAttribute('data-theme', 'dark');
    lightSheet.disabled = true;
    darkSheet.disabled = false;
    themeSwitch.checked = true;
    
    // Update labels
    lightLabel.classList.remove('active');
    darkLabel.classList.add('active');
  } else {
    // Switch to light mode
    document.documentElement.removeAttribute('data-theme');
    lightSheet.disabled = false;
    darkSheet.disabled = true;
    themeSwitch.checked = false;
    
    // Update labels
    lightLabel.classList.add('active');
    darkLabel.classList.remove('active');
  }
  
  // Save preference
  localStorage.setItem('storcat-theme', isDark ? 'dark' : 'light');
}

function toggleTheme() {
  setTheme(!appState.isDarkMode);
}

function initializeTheme() {
  // Load saved theme preference
  const savedTheme = localStorage.getItem('storcat-theme');
  if (savedTheme === 'dark') {
    setTheme(true);
  } else {
    setTheme(false);
  }
}

// Window persistence management
function setWindowPersistence(enabled) {
  // Save preference
  localStorage.setItem('storcat-window-persistence', enabled ? 'true' : 'false');
  
  // Update switch state
  elements.windowPersistenceSwitch.checked = enabled;
  
  // Send to main process
  window.electronAPI.setWindowPersistence(enabled);
}

async function initializeWindowPersistence() {
  // Get initial setting from main process
  try {
    const result = await window.electronAPI.getWindowPersistence();
    if (result.success) {
      elements.windowPersistenceSwitch.checked = result.enabled;
      // Also save to localStorage for consistency
      localStorage.setItem('storcat-window-persistence', result.enabled ? 'true' : 'false');
    }
  } catch (error) {
    // Fallback to localStorage if main process call fails
    const savedPersistence = localStorage.getItem('storcat-window-persistence');
    const enabled = savedPersistence !== 'false'; // Default to true if not set
    setWindowPersistence(enabled);
  }
  
}

function initializeDirectoryPreferences() {
  // Load saved directory preferences
  const savedCatalogDirectory = localStorage.getItem('storcat-last-catalog-directory');
  const savedCreateDirectory = localStorage.getItem('storcat-last-create-directory');
  const savedOutputDirectory = localStorage.getItem('storcat-last-output-directory');
  const savedSearchTerm = localStorage.getItem('storcat-last-search-term');
  
  // Both search and browse use the same catalog directory concept
  if (savedCatalogDirectory) {
    appState.selectedSearchDirectory = savedCatalogDirectory;
    appState.selectedBrowseDirectory = savedCatalogDirectory;
    
    if (elements.selectedSearchPath) {
      elements.selectedSearchPath.textContent = savedCatalogDirectory;
      showElement(elements.selectedSearchPath);
    }
    
    if (elements.selectedBrowsePath) {
      elements.selectedBrowsePath.textContent = savedCatalogDirectory;
      showElement(elements.selectedBrowsePath);
    }
    
    updateSearchButton();
    updateLoadButton();
    
    // Auto-load browse catalogs in background (don't display)
    setTimeout(() => {
      loadBrowseCatalogs(false);
    }, 200);
  }
  
  // Restore saved search term
  if (savedSearchTerm && elements.searchTerm) {
    // Use setTimeout to ensure sl-input is fully initialized
    setTimeout(() => {
      elements.searchTerm.value = savedSearchTerm;
      // Trigger the input event to ensure proper state
      elements.searchTerm.dispatchEvent(new Event('sl-input', { bubbles: true }));
      updateSearchButton();
      
      // Auto-execute search if we have both search term and directory (background only)
      if (savedCatalogDirectory) {
        setTimeout(() => {
          executeSearch(false);
        }, 100);
      }
    }, 150);
  }
  
  if (savedCreateDirectory) {
    appState.selectedDirectory = savedCreateDirectory;
    
    if (elements.selectedDirectoryPath) {
      elements.selectedDirectoryPath.textContent = savedCreateDirectory;
      showElement(elements.selectedDirectoryPath);
    }
    
    updateCreateButton();
  }
  
  if (savedOutputDirectory) {
    appState.selectedOutputDirectory = savedOutputDirectory;
    
    if (elements.selectedOutputPath) {
      elements.selectedOutputPath.textContent = savedOutputDirectory;
      showElement(elements.selectedOutputPath);
    }
  }
}

// Tab switching functionality
function switchToTab(tabName) {
  // Hide all tab sidebars
  elements.createSidebar.classList.add('hidden');
  elements.searchSidebar.classList.add('hidden');
  elements.browseSidebar.classList.add('hidden');
  
  // Show the selected tab sidebar and update main display content
  switch(tabName) {
    case 'create-tab':
      elements.createSidebar.classList.remove('hidden');
      showCreateWelcome();
      break;
    case 'search-tab':
      elements.searchSidebar.classList.remove('hidden');
      showSearchContent();
      break;
    case 'browse-tab':
      elements.browseSidebar.classList.remove('hidden');
      showBrowseContent();
      break;
  }
}

function showCreateWelcome() {
  elements.mainDisplay.innerHTML = `
    <div class="section-card">
      <h2 class="section-title">Welcome to StorCat</h2>
      <p>Use the tabs above to create new directory catalogs, search existing catalogs, or browse your catalog collection.</p>
      
      <h3 style="margin: 24px 0 16px 0; font-size: 1.1rem; font-weight: 600; color: var(--app-text);">Features</h3>
      <ul style="margin: 0; padding-left: 20px; line-height: 1.6;">
        <li><strong>Create Catalogs:</strong> Generate JSON and HTML catalogs of any directory</li>
        <li><strong>Search Catalogs:</strong> Find files across multiple catalog files</li>
        <li><strong>Browse Catalogs:</strong> View and manage your catalog collection</li>
        <li><strong>Cross-Platform:</strong> Pure JavaScript implementation works on all platforms</li>
        <li><strong>Compatible:</strong> 100% compatible with existing catalog files</li>
      </ul>
    </div>
  `;
}

function showSearchContent() {
  // If we have search results, display them, otherwise show empty state
  if (appState.searchResults && appState.searchResults.length > 0) {
    displaySearchResults(appState.searchResults);
  } else if (appState.selectedSearchDirectory && elements.searchTerm.value.trim()) {
    // Show searching state or empty results
    elements.mainDisplay.innerHTML = `
      <div class="section-card">
        <h2 class="section-title">Search Results</h2>
        <p>No results found for "${elements.searchTerm.value.trim()}" in the selected catalog directory.</p>
      </div>
    `;
  } else {
    // Show instructions
    elements.mainDisplay.innerHTML = `
      <div class="section-card">
        <h2 class="section-title">Search Catalogs</h2>
        <p>Enter a search term and select a directory containing catalog files to search for files and directories.</p>
        
        <h3 style="margin: 24px 0 16px 0; font-size: 1.1rem; font-weight: 600; color: var(--app-text);">How to Search</h3>
        <ol style="margin: 0; padding-left: 20px; line-height: 1.6;">
          <li>Enter your search term in the sidebar</li>
          <li>Select a directory containing .json catalog files</li>
          <li>Click "Search Catalogs" to find matching files and directories</li>
          <li>Results will appear in a sortable, resizable table</li>
        </ol>
      </div>
    `;
  }
}

function showBrowseContent() {
  // If we have browse catalogs, display them, otherwise show empty state
  if (appState.browseCatalogs && appState.browseCatalogs.length > 0) {
    displayCatalogList(appState.browseCatalogs);
  } else if (appState.selectedBrowseDirectory) {
    // Show loading or empty state
    elements.mainDisplay.innerHTML = `
      <div class="section-card">
        <h2 class="section-title">Catalog Browser</h2>
        <p>No catalog files found in the selected directory.</p>
      </div>
    `;
  } else {
    // Show instructions
    elements.mainDisplay.innerHTML = `
      <div class="section-card">
        <h2 class="section-title">Browse Catalogs</h2>
        <p>Select a directory containing catalog files to view and manage your catalog collection.</p>
        
        <h3 style="margin: 24px 0 16px 0; font-size: 1.1rem; font-weight: 600; color: var(--app-text);">How to Browse</h3>
        <ol style="margin: 0; padding-left: 20px; line-height: 1.6;">
          <li>Select a directory containing .json catalog files</li>
          <li>Click "Load Catalogs" to view available catalogs</li>
          <li>Click on any catalog title to view its HTML representation</li>
          <li>Use the sortable table to organize catalogs by title, file, or date</li>
        </ol>
      </div>
    `;
  }
}

// Helper functions for auto-loading
async function loadBrowseCatalogs(displayResults = true) {
  if (!appState.selectedBrowseDirectory || appState.isLoading) return;
  
  appState.isLoading = true;
  if (elements.loadCatalogsBtn) {
    elements.loadCatalogsBtn.loading = true;
  }
  updateLoadButton();
  if (displayResults) {
    hideElement(elements.browseStatus);
  }
  
  try {
    const result = await window.electronAPI.getCatalogFiles(appState.selectedBrowseDirectory);
    
    if (result.success) {
      // Store the catalogs in state
      appState.browseCatalogs = result.catalogs;
      
      // Only display if requested (when user is on browse tab)
      if (displayResults) {
        displayCatalogList(result.catalogs);
      }
    } else {
      if (displayResults) {
        showStatus(elements.browseStatus, `Error: ${result.error}`, 'error');
      }
    }
  } catch (error) {
    if (displayResults) {
      showStatus(elements.browseStatus, `Error: ${error.message}`, 'error');
    }
  } finally {
    appState.isLoading = false;
    if (elements.loadCatalogsBtn) {
      elements.loadCatalogsBtn.loading = false;
    }
    updateLoadButton();
  }
}

async function executeSearch(displayResults = true) {
  if (!elements.searchTerm.value.trim() || !appState.selectedSearchDirectory || appState.isSearching) return;
  
  appState.isSearching = true;
  if (elements.searchCatalogsBtn) {
    elements.searchCatalogsBtn.loading = true;
  }
  updateSearchButton();
  if (displayResults) {
    hideElement(elements.searchStatus);
  }
  
  try {
    const result = await window.electronAPI.searchCatalogs(
      elements.searchTerm.value.trim(),
      appState.selectedSearchDirectory
    );
    
    if (result.success) {
      // Store the search results in state
      appState.searchResults = result.results;
      
      // Only display if requested (when user is on search tab)
      if (displayResults) {
        displaySearchResults(result.results);
      }
    } else {
      if (displayResults) {
        showStatus(elements.searchStatus, `Error: ${result.error}`, 'error');
      }
    }
  } catch (error) {
    if (displayResults) {
      showStatus(elements.searchStatus, `Error: ${error.message}`, 'error');
    }
  } finally {
    appState.isSearching = false;
    if (elements.searchCatalogsBtn) {
      elements.searchCatalogsBtn.loading = false;
    }
    updateSearchButton();
  }
}

// Event listeners
elements.sidebarToggle.addEventListener('click', toggleSidebar);
elements.settingsToggle.addEventListener('click', showSettingsModal);
elements.themeSwitch.addEventListener('sl-change', (e) => {
  // Set theme based on switch position
  setTheme(e.target.checked);
});

elements.windowPersistenceSwitch.addEventListener('change', (e) => {
  // Set window persistence based on switch position
  setWindowPersistence(e.target.checked);
});

// Make theme labels clickable
document.addEventListener('DOMContentLoaded', () => {
  const lightLabel = document.querySelector('.theme-switch-container .theme-label:first-child');
  const darkLabel = document.querySelector('.theme-switch-container .theme-label:last-child');
  
  if (lightLabel) {
    lightLabel.addEventListener('click', () => {
      setTheme(false); // Set to light mode
    });
  }
  
  if (darkLabel) {
    darkLabel.addEventListener('click', () => {
      setTheme(true); // Set to dark mode
    });
  }
});

// Tab switching event listener
elements.mainTabs.addEventListener('sl-tab-show', (event) => {
  switchToTab(event.detail.name);
});

// Modal event listeners
elements.modalClose.addEventListener('click', hideModal);
elements.catalogModal.addEventListener('click', (e) => {
  if (e.target === elements.catalogModal) {
    hideModal();
  }
});

// Settings modal event listeners
elements.settingsModalClose.addEventListener('click', hideSettingsModal);
elements.settingsModal.addEventListener('click', (e) => {
  if (e.target === elements.settingsModal) {
    hideSettingsModal();
  }
});

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    if (elements.settingsModal.classList.contains('show')) {
      hideSettingsModal();
    } else if (elements.catalogModal.classList.contains('show')) {
      hideModal();
    }
  }
});

// Make functions available globally for onclick handlers
window.openCatalogHtml = openCatalogHtml;
window.openCatalogHtmlFromPath = openCatalogHtmlFromPath;

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
  // Initialize theme first
  initializeTheme();
  
  // Initialize window persistence
  initializeWindowPersistence();
  
  // Set initial main content class
  const mainContent = document.querySelector('.main-content');
  mainContent.classList.add('sidebar-open');
  
  // Show the default Create tab content immediately
  showCreateWelcome();
  
  // Use setTimeout to ensure all elements are fully initialized
  setTimeout(() => {
    updateCreateButton();
    updateSearchButton();
    updateLoadButton();
    
    // Initialize directory preferences after a small delay
    // This will load data in the background but won't change the display
    initializeDirectoryPreferences();
  }, 100);
});