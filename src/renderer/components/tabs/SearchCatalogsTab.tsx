import React, { useState, useEffect } from 'react';
import { Button, Input, message, Space, Typography } from 'antd';
import { FolderOutlined, SearchOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { useAppContext } from '../../contexts/AppContext';
import ModernTable from '../ModernTable';

const { Title } = Typography;

function SearchCatalogsSidebar() {
  const { state, dispatch } = useAppContext();
  const [searchTerm, setSearchTerm] = useState('');

  // Load saved preferences on mount and auto-search if both directory and term exist
  useEffect(() => {
    const savedCatalogDirectory = localStorage.getItem('storcat-last-catalog-directory');
    const savedSearchTerm = localStorage.getItem('storcat-last-search-term');
    
    if (savedCatalogDirectory) {
      dispatch({ type: 'SET_SELECTED_SEARCH_DIRECTORY', payload: savedCatalogDirectory });
    }
    if (savedSearchTerm) {
      setSearchTerm(savedSearchTerm);
    }

    // Auto-search if both directory and search term are available
    if (savedCatalogDirectory && savedSearchTerm) {
      performSearch(savedSearchTerm, savedCatalogDirectory);
    }
  }, [dispatch]);

  const performSearch = async (searchTermValue: string, directoryPath: string) => {
    dispatch({ type: 'SET_IS_SEARCHING', payload: true });

    try {
      const result = await window.electronAPI.searchCatalogs(searchTermValue.trim(), directoryPath);

      if (result.success) {
        dispatch({ type: 'SET_SEARCH_RESULTS', payload: result.results || [] });
        localStorage.setItem('storcat-last-search-results', JSON.stringify(result.results || []));
      } else {
        message.error(`Search failed: ${result.error}`);
      }
    } catch (error) {
      message.error('Failed to search catalogs');
    } finally {
      dispatch({ type: 'SET_IS_SEARCHING', payload: false });
    }
  };

  const selectSearchDirectory = async () => {
    try {
      const directory = await window.electronAPI.selectSearchDirectory();
      if (directory) {
        dispatch({ type: 'SET_SELECTED_SEARCH_DIRECTORY', payload: directory });
        localStorage.setItem('storcat-last-catalog-directory', directory);
      }
    } catch (error) {
      message.error('Failed to select directory');
    }
  };

  const searchCatalogs = async () => {
    if (!searchTerm.trim()) {
      message.error('Please enter a search term');
      return;
    }
    if (!state.selectedSearchDirectory) {
      message.error('Please select a catalog directory');
      return;
    }

    // Save search term
    localStorage.setItem('storcat-last-search-term', searchTerm.trim());
    
    // Perform the search
    await performSearch(searchTerm, state.selectedSearchDirectory);
    
    // Show success message
    message.success(`Found ${state.searchResults?.length || 0} matches`);
  };

  const canSearch = searchTerm.trim() && state.selectedSearchDirectory;

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {/* Search Term */}
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
          <label style={{ fontSize: '1rem', fontWeight: 500, color: 'var(--app-text)', margin: 0 }}>
            Search Term
          </label>
          <InfoCircleOutlined 
            style={{ fontSize: 16, color: 'var(--sl-color-neutral-500)' }}
            title="Search for files and folders in catalog files"
          />
        </div>
        <Input
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Enter search term"
          style={{ width: '100%' }}
          onPressEnter={canSearch ? searchCatalogs : undefined}
        />
      </div>

      {/* Select Catalog Directory */}
      <div>
        <Button
          type="primary"
          icon={<FolderOutlined />}
          onClick={selectSearchDirectory}
          style={{ width: '100%', justifyContent: 'center' }}
        >
          Select Catalog Directory
        </Button>
        {state.selectedSearchDirectory && (
          <div className="path-display" style={{ marginTop: 8 }}>
            {state.selectedSearchDirectory}
          </div>
        )}
      </div>

      {/* Search Button */}
      <Button
        type="primary"
        icon={<SearchOutlined />}
        onClick={searchCatalogs}
        loading={state.isSearching}
        disabled={!canSearch}
        style={{ width: '100%', justifyContent: 'center' }}
      >
        Search Catalogs
      </Button>
    </Space>
  );
}

function SearchCatalogsContent() {
  const { state } = useAppContext();

  const columns = [
    {
      key: 'basename',
      title: 'Name',
      width: 200,
      minWidth: 120,
      render: (value: string) => (
        <span style={{ fontWeight: 500 }}>{value}</span>
      ),
    },
    {
      key: 'fullPath',
      title: 'Path',
      width: 300,
      minWidth: 200,
      render: (value: string) => (
        <span style={{ fontSize: '0.85rem', color: 'var(--sl-color-neutral-600)' }}>
          {value || ''}
        </span>
      ),
    },
    {
      key: 'type',
      title: 'Type',
      width: 80,
      minWidth: 60,
    },
    {
      key: 'size',
      title: 'Size',
      width: 100,
      minWidth: 80,
      render: (value: number) => formatBytes(value),
    },
    {
      key: 'catalog',
      title: 'Catalog',
      width: 150,
      minWidth: 120,
      render: (value: string, record: any) => (
        <a
          href="#"
          className="catalog-link"
          onClick={(e) => {
            e.preventDefault();
            openCatalogHtml(record.catalogFilePath);
          }}
        >
          {value}
        </a>
      ),
    },
  ];

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0B';
    const k = 1024;
    const sizes = ['B', 'K', 'M', 'G', 'T'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 10) / 10 + sizes[i];
  };

  const openCatalogHtml = async (catalogPath: string) => {
    try {
      // Dispatch custom event to open the catalog modal
      window.dispatchEvent(new CustomEvent('openCatalogModal', {
        detail: { catalogPath }
      }));
    } catch (error) {
      message.error('Failed to open catalog HTML');
    }
  };

  // Test data to verify AG Grid is working
  const testData = [
    { basename: 'test-file.txt', fullPath: '/test/path/test-file.txt', type: 'file', size: 1024, catalog: 'Test Catalog', catalogFilePath: '/test/catalog.json', fullName: 'test-file.txt' },
    { basename: 'another-file.pdf', fullPath: '/test/path/another-file.pdf', type: 'file', size: 2048, catalog: 'Test Catalog 2', catalogFilePath: '/test/catalog2.json', fullName: 'another-file.pdf' },
  ];

  const actualData = state.searchResults.length > 0 ? state.searchResults : testData;

  return (
    <div style={{ 
      height: 'calc(100vh - var(--header-height) - var(--tab-nav-height) - 48px)', 
      display: 'flex', 
      flexDirection: 'column',
      margin: '-24px -24px 0 -24px',
      padding: '8px 24px 0 24px'
    }}>
      <div style={{ padding: '8px 0' }}>
        <Title level={4} style={{ margin: 0, color: 'var(--app-text)', fontSize: '1.1rem' }}>
          Search Results {state.searchResults.length > 0 ? `(${state.searchResults.length})` : ''}
        </Title>
      </div>
      
      <div style={{ flex: 1, minHeight: 0 }}>
        <ModernTable
          columns={columns}
          data={actualData}
          loading={state.isSearching}
          pageSize={50}
          height="100%"
          emptyText="No search results. Enter a search term and click 'Search Catalogs' to get started."
        />
      </div>
    </div>
  );
}

const SearchCatalogsTab = {
  Sidebar: SearchCatalogsSidebar,
  Content: SearchCatalogsContent,
};

export default SearchCatalogsTab;