const { createCatalog, searchCatalogs } = require('./src/catalog-service');
const fs = require('fs').promises;
const path = require('path');

async function testCompatibility() {
  console.log('Testing StorCat Electron compatibility...\n');
  
  try {
    // Test creating a catalog of the examples directory
    console.log('1. Creating catalog of examples directory...');
    const result = await createCatalog({
      title: 'Examples Test',
      directoryPath: './examples',
      outputRoot: 'test-catalog'
    });
    
    console.log(`✓ JSON created: ${result.jsonPath}`);
    console.log(`✓ HTML created: ${result.htmlPath}`);
    console.log(`✓ File count: ${result.fileCount}`);
    console.log(`✓ Total size: ${result.totalSize} bytes\n`);
    
    // Test searching in the examples directory (should find the existing sd01.json)
    console.log('2. Testing search functionality...');
    const searchResults = await searchCatalogs('Brad Williams', './examples');
    
    if (searchResults.length > 0) {
      console.log(`✓ Found ${searchResults.length} search results`);
      searchResults.slice(0, 3).forEach((result, i) => {
        console.log(`  ${i + 1}. ${result.catalog}: ${result.name} (${result.type})`);
      });
    } else {
      console.log('ℹ No search results found (this is normal if no catalogs contain "Brad Williams")');
    }
    
    console.log('\n3. Verifying JSON format compatibility...');
    
    // Read the created catalog and verify structure
    const catalogContent = await fs.readFile(result.jsonPath, 'utf8');
    const catalog = JSON.parse(catalogContent);
    
    // Check required fields
    const hasRequiredFields = catalog.type && catalog.name && catalog.size !== undefined;
    if (hasRequiredFields) {
      console.log('✓ JSON structure matches expected format');
      console.log(`  - Type: ${catalog.type}`);
      console.log(`  - Name: ${catalog.name}`);
      console.log(`  - Size: ${catalog.size}`);
      if (catalog.contents) {
        console.log(`  - Contents: ${catalog.contents.length} items`);
      }
    } else {
      console.log('✗ JSON structure does not match expected format');
    }
    
    console.log('\n4. Cleaning up test files...');
    await fs.unlink(result.jsonPath);
    await fs.unlink(result.htmlPath);
    console.log('✓ Test files cleaned up');
    
    console.log('\n🎉 StorCat Electron compatibility test completed successfully!');
    
  } catch (error) {
    console.error('❌ Test failed:', error.message);
    process.exit(1);
  }
}

testCompatibility();