const fs = require('fs').promises;
const path = require('path');

function formatBytes(bytes) {
  if (bytes === 0) return '0B';
  const k = 1024;
  const sizes = ['B', 'K', 'M', 'G', 'T'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 10) / 10 + sizes[i];
}

function formatBytesForDisplay(bytes) {
  if (bytes === 0) return '[   0]';
  const formatted = formatBytes(bytes);
  return `[${formatted.padStart(4)}]`;
}

async function getDirectorySize(dirPath) {
  let totalSize = 0;
  
  try {
    const entries = await fs.readdir(dirPath, { withFileTypes: true });
    
    for (const entry of entries) {
      const fullPath = path.join(dirPath, entry.name);
      
      if (entry.isFile()) {
        try {
          const stats = await fs.stat(fullPath);
          totalSize += stats.size;
        } catch (error) {
          // Skip files we can't access
        }
      } else if (entry.isDirectory()) {
        try {
          totalSize += await getDirectorySize(fullPath);
        } catch (error) {
          // Skip directories we can't access
        }
      }
    }
  } catch (error) {
    // Return 0 if we can't read the directory
  }
  
  return totalSize;
}

async function traverseDirectory(dirPath, basePath = dirPath, onProgress = null) {
  const stats = await fs.stat(dirPath);
  const relativePath = path.relative(path.dirname(basePath), dirPath);
  const displayPath = relativePath === '' ? './' : `./${relativePath}`;
  
  // Report progress if callback provided
  if (onProgress) {
    onProgress(displayPath);
  }
  
  if (stats.isFile()) {
    return {
      type: 'file',
      name: displayPath,
      size: stats.size
    };
  }
  
  if (stats.isDirectory()) {
    const contents = [];
    let totalSize = 0;
    
    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });
      
      // Sort entries: directories first, then files, both alphabetically
      entries.sort((a, b) => {
        if (a.isDirectory() && !b.isDirectory()) return -1;
        if (!a.isDirectory() && b.isDirectory()) return 1;
        return a.name.localeCompare(b.name);
      });
      
      for (const entry of entries) {
        try {
          // Skip hidden files and directories (those starting with '.')
          if (entry.name.startsWith('.')) {
            continue;
          }
          
          const fullPath = path.join(dirPath, entry.name);
          const childItem = await traverseDirectory(fullPath, basePath, onProgress);
          contents.push(childItem);
          totalSize += childItem.size;
        } catch (error) {
          // Skip items we can't access (permissions, broken symlinks, etc.)
          console.warn(`Skipping ${fullPath}: ${error.message}`);
        }
      }
    } catch (error) {
      // If we can't read the directory, return it as empty
      console.warn(`Cannot read directory ${dirPath}: ${error.message}`);
    }
    
    return {
      type: 'directory',
      name: displayPath,
      size: totalSize,
      contents: contents
    };
  }
}

function generateTreeHTML(catalog, title) {
  const escapeHtml = (text) => {
    const div = {
      '"': '&quot;',
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;'
    };
    return text.replace(/[\"&<>]/g, (a) => div[a]);
  };

  const generateTreeStructure = (item, isLast = true, prefix = '') => {
    let html = '';
    const connector = isLast ? '└── ' : '├── ';
    const sizeDisplay = formatBytesForDisplay(item.size);
    const itemName = path.basename(item.name);
    
    if (item.type === 'directory') {
      html += `${prefix}${connector}${sizeDisplay}&nbsp;&nbsp;${escapeHtml(itemName)}<br>\n`;
      
      if (item.contents && item.contents.length > 0) {
        const newPrefix = prefix + (isLast ? '    ' : '│   ');
        item.contents.forEach((child, index) => {
          const childIsLast = index === item.contents.length - 1;
          html += generateTreeStructure(child, childIsLast, newPrefix);
        });
      }
    } else {
      html += `${prefix}${connector}${sizeDisplay}&nbsp;&nbsp;${escapeHtml(itemName)}<br>\n`;
    }
    
    return html;
  };

  const treeStructure = generateTreeStructure(catalog);

  return `<!DOCTYPE html>
<html>
<head>
 <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
 <meta name="Author" content="Made by 'tree'">
 <meta name="GENERATOR" content="$Version: $ tree v1.7.0 (c) 1996 - 2014 by Steve Baker, Thomas Moore, Francesc Rocher, Florian Sesser, Kyosuke Tokoro $">
 <title>${escapeHtml(title)}</title>
 <style type="text/css">
  <!-- 
  BODY { font-family : ariel, monospace, sans-serif; }
  P { font-weight: normal; font-family : ariel, monospace, sans-serif; color: black; background-color: transparent;}
  B { font-weight: normal; color: black; background-color: transparent;}
  A:visited { font-weight : normal; text-decoration : none; background-color : transparent; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:link    { font-weight : normal; text-decoration : none; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:hover   { color : #000000; font-weight : normal; text-decoration : underline; background-color : yellow; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:active  { color : #000000; font-weight: normal; background-color : transparent; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  .VERSION { font-size: small; font-family : arial, sans-serif; }
  .NORM  { color: black;  background-color: transparent;}
  .FIFO  { color: purple; background-color: transparent;}
  .CHAR  { color: yellow; background-color: transparent;}
  .DIR   { color: blue;   background-color: transparent;}
  .BLOCK { color: yellow; background-color: transparent;}
  .LINK  { color: aqua;   background-color: transparent;}
  .SOCK  { color: fuchsia;background-color: transparent;}
  .EXEC  { color: green;  background-color: transparent;}
  -->
 </style>
</head>
<body>
	<h1>${escapeHtml(title)}</h1><p>
	${treeStructure}
	<br><br>
	</p>
	<p>

 ${formatBytes(catalog.size)} used in ${countDirectories(catalog)} directories, ${countFiles(catalog)} files
	<br><br>
	</p>
	<hr>
	<p class="VERSION">
		 tree v1.7.0 © 1996 - 2014 by Steve Baker and Thomas Moore <br>
		 HTML output hacked and copyleft © 1998 by Francesc Rocher <br>
		 JSON output hacked and copyleft © 2014 by Florian Sesser <br>
		 Charsets / OS/2 support © 2001 by Kyosuke Tokoro
	</p>
</body>
</html>`;
}

async function createCatalog({ title, directoryPath, outputRoot, copyToDirectory = null }) {
  try {
    // Generate the catalog data
    const catalog = await traverseDirectory(directoryPath);
    
    // Create JSON file
    const jsonContent = JSON.stringify(catalog, null, 0);
    const jsonFilePath = path.join(directoryPath, `${outputRoot}.json`);
    await fs.writeFile(jsonFilePath, jsonContent, 'utf8');
    
    // Create HTML file
    const htmlContent = generateTreeHTML(catalog, title);
    const htmlFilePath = path.join(directoryPath, `${outputRoot}.html`);
    await fs.writeFile(htmlFilePath, htmlContent, 'utf8');
    
    const result = {
      jsonPath: jsonFilePath,
      htmlPath: htmlFilePath,
      fileCount: countFiles(catalog),
      totalSize: catalog.size
    };
    
    // Copy to secondary directory if requested
    if (copyToDirectory) {
      const copyJsonPath = path.join(copyToDirectory, `${outputRoot}.json`);
      const copyHtmlPath = path.join(copyToDirectory, `${outputRoot}.html`);
      
      await fs.copyFile(jsonFilePath, copyJsonPath);
      await fs.copyFile(htmlFilePath, copyHtmlPath);
      
      result.copyJsonPath = copyJsonPath;
      result.copyHtmlPath = copyHtmlPath;
    }
    
    return result;
  } catch (error) {
    throw new Error(`Failed to create catalog: ${error.message}`);
  }
}

function countFiles(catalog) {
  if (catalog.type === 'file') {
    return 1;
  }
  
  if (catalog.type === 'directory' && catalog.contents) {
    return catalog.contents.reduce((count, item) => count + countFiles(item), 0);
  }
  
  return 0;
}

function countDirectories(catalog) {
  if (catalog.type === 'file') {
    return 0;
  }
  
  if (catalog.type === 'directory') {
    let count = 1; // Count this directory
    if (catalog.contents) {
      count += catalog.contents.reduce((acc, item) => acc + countDirectories(item), 0);
    }
    return count;
  }
  
  return 0;
}

async function searchCatalogs(searchTerm, catalogDirectory) {
  const results = [];
  
  try {
    const files = await fs.readdir(catalogDirectory);
    const jsonFiles = files.filter(file => file.endsWith('.json'));
    
    for (const file of jsonFiles) {
      const filePath = path.join(catalogDirectory, file);
      const catalogName = path.basename(file, '.json');
      
      try {
        const content = await fs.readFile(filePath, 'utf8');
        const catalog = JSON.parse(content);
        
        // Handle different JSON formats:
        // 1. Array format from bash script (our target): [{"contents": [...], "name": "...", "type": "..."}]
        // 2. Object format from our implementation: {"contents": [...], "name": "...", "type": "..."}
        let catalogRoot;
        if (Array.isArray(catalog)) {
          // Bash script format - use first element
          catalogRoot = catalog[0];
        } else {
          // Our format - use directly
          catalogRoot = catalog;
        }
        
        const matches = searchInCatalog(catalogRoot, searchTerm.toLowerCase(), catalogName, filePath);
        results.push(...matches);
      } catch (error) {
        // Skip files that can't be read or parsed
        console.warn(`Failed to process ${filePath}:`, error.message);
      }
    }
  } catch (error) {
    throw new Error(`Failed to search catalogs: ${error.message}`);
  }
  
  return results;
}

function searchInCatalog(item, searchTerm, catalogName, catalogFilePath) {
  const results = [];
  
  // Handle root-level catalog structure from bash script
  if (item.contents && Array.isArray(item.contents) && !item.name) {
    // This is the root object with contents array
    for (const child of item.contents) {
      results.push(...searchInCatalog(child, searchTerm, catalogName, catalogFilePath));
    }
    return results;
  }
  
  if (item.name && item.name.toLowerCase().includes(searchTerm)) {
    // Extract basename and path from the full name
    const fullName = item.name;
    const basename = path.basename(fullName);
    const itemPath = path.dirname(fullName);
    
    results.push({
      catalog: catalogName,
      catalogFilePath: catalogFilePath,
      basename: basename,
      fullPath: itemPath === '.' ? '' : itemPath,
      fullName: fullName,
      type: item.type,
      size: item.size
    });
  }
  
  if (item.contents && Array.isArray(item.contents)) {
    for (const child of item.contents) {
      results.push(...searchInCatalog(child, searchTerm, catalogName, catalogFilePath));
    }
  }
  
  return results;
}

module.exports = {
  createCatalog,
  searchCatalogs,
  formatBytes
};