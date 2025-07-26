const { createCatalog, searchCatalogs } = require('./src/catalog-service');
const fs = require('fs').promises;

async function demo() {
  console.log('🎯 StorCat Electron Demo\n');
  
  try {
    // Create a catalog of the entire project
    console.log('1. Creating catalog of project directory...');
    const result = await createCatalog({
      title: 'StorCat Electron Project',
      directoryPath: '.',
      outputRoot: 'storcat-project',
      copyToDirectory: null
    });
    
    console.log(`✅ Created catalog with ${result.fileCount} files (${Math.round(result.totalSize/1024)}KB total)`);
    console.log(`   JSON: ${result.jsonPath}`);
    console.log(`   HTML: ${result.htmlPath}\n`);
    
    // Search for JavaScript files
    console.log('2. Searching for JavaScript files...');
    const jsResults = await searchCatalogs('.js', '.');
    console.log(`✅ Found ${jsResults.length} JavaScript files:`);
    jsResults.forEach(r => console.log(`   - ${r.name}`));
    
    console.log('\n3. Searching for JSON files...');
    const jsonResults = await searchCatalogs('.json', '.');
    console.log(`✅ Found ${jsonResults.length} JSON files:`);
    jsonResults.forEach(r => console.log(`   - ${r.name}`));
    
    console.log('\n4. Searching for "catalog" in filenames...');
    const catalogResults = await searchCatalogs('catalog', '.');
    console.log(`✅ Found ${catalogResults.length} files containing "catalog":`);
    catalogResults.forEach(r => console.log(`   - ${r.name} (${r.type})`));
    
    console.log('\n5. File structure overview:');
    const catalogContent = await fs.readFile(result.jsonPath, 'utf8');
    const catalog = JSON.parse(catalogContent);
    
    function countByType(item) {
      if (item.type === 'file') return { files: 1, dirs: 0 };
      let files = 0, dirs = 1;
      if (item.contents) {
        item.contents.forEach(child => {
          const counts = countByType(child);
          files += counts.files;
          dirs += counts.dirs;
        });
      }
      return { files, dirs };
    }
    
    const counts = countByType(catalog);
    console.log(`   📁 Directories: ${counts.dirs}`);
    console.log(`   📄 Files: ${counts.files}`);
    console.log(`   📊 Total Size: ${Math.round(catalog.size/1024)}KB`);
    
    console.log('\n🎉 Demo completed! You can now:');
    console.log('   - Open the HTML file in a browser to see the tree view');
    console.log('   - Use the Electron app with "npm run dev" for the GUI');
    console.log('   - The catalog files are compatible with the original bash script');
    
  } catch (error) {
    console.error('❌ Demo failed:', error.message);
  }
}

demo();