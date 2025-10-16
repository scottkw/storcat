import React, { useState, useEffect } from 'react';
import { Button, message, Space, Typography } from 'antd';
import { FolderOutlined, UnorderedListOutlined } from '@ant-design/icons';
import { useAppContext } from '../../contexts/AppContext';
import ModernTable from '../ModernTable';

const { Title } = Typography;

function BrowseCatalogsSidebar() {
  const { state, dispatch } = useAppContext();

  // Load saved preferences on mount and auto-load catalogs if directory exists
  useEffect(() => {
    const savedBrowseDirectory = localStorage.getItem('storcat-last-catalog-directory');
    
    if (savedBrowseDirectory) {
      dispatch({ type: 'SET_SELECTED_BROWSE_DIRECTORY', payload: savedBrowseDirectory });
      // Auto-load catalogs from the saved directory
      performLoadCatalogs(savedBrowseDirectory);
    }
  }, [dispatch]);

  const performLoadCatalogs = async (directoryPath: string) => {
    dispatch({ type: 'SET_IS_LOADING', payload: true });

    try {
      const result = await window.electronAPI.getCatalogFiles(directoryPath);

      if (result.success) {
        dispatch({ type: 'SET_BROWSE_CATALOGS', payload: result.catalogs || [] });
        localStorage.setItem('storcat-last-browse-catalogs', JSON.stringify(result.catalogs || []));
      } else {
        message.error(`Failed to load catalogs: ${result.error}`);
      }
    } catch (error) {
      message.error('Failed to load catalogs');
    } finally {
      dispatch({ type: 'SET_IS_LOADING', payload: false });
    }
  };

  const selectBrowseDirectory = async () => {
    try {
      const directory = await window.electronAPI.selectSearchDirectory();
      if (directory) {
        dispatch({ type: 'SET_SELECTED_BROWSE_DIRECTORY', payload: directory });
        localStorage.setItem('storcat-last-catalog-directory', directory);
      }
    } catch (error) {
      message.error('Failed to select directory');
    }
  };

  const loadCatalogs = async () => {
    if (!state.selectedBrowseDirectory) {
      message.error('Please select a catalog directory');
      return;
    }

    await performLoadCatalogs(state.selectedBrowseDirectory);
    message.success(`Loaded ${state.browseCatalogs?.length || 0} catalogs`);
  };

  const canLoad = state.selectedBrowseDirectory;

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {/* Title */}
      <div style={{ marginBottom: '8px' }}>
        <label style={{ fontSize: '1rem', fontWeight: 500, color: 'var(--app-text)', margin: 0 }}>
          Browse Catalogs
        </label>
      </div>

      {/* Select Catalog Directory */}
      <div>
        <Button
          type="primary"
          icon={<FolderOutlined />}
          onClick={selectBrowseDirectory}
          style={{ width: '100%', justifyContent: 'center' }}
        >
          Select Catalog Directory
        </Button>
        {state.selectedBrowseDirectory && (
          <div className="path-display" style={{ marginTop: 8 }}>
            {state.selectedBrowseDirectory}
          </div>
        )}
      </div>

      {/* Load Catalogs Button */}
      <Button
        type="primary"
        icon={<UnorderedListOutlined />}
        onClick={loadCatalogs}
        loading={state.isLoading}
        disabled={!canLoad}
        style={{ width: '100%', justifyContent: 'center' }}
      >
        Load Catalogs
      </Button>
    </Space>
  );
}

function BrowseCatalogsContent() {
  const { state } = useAppContext();

  const columns = [
    {
      key: 'title',
      title: 'Title',
      width: 300,
      minWidth: 200,
      render: (value: string, record: any) => {
        if (record.hasHtml) {
          return (
            <a
              href="#"
              className="catalog-link"
              onClick={(e) => {
                e.preventDefault();
                openCatalogHtml(record.path);
              }}
            >
              {value}
            </a>
          );
        } else {
          return (
            <span style={{ color: 'var(--sl-color-neutral-600)' }}>
              {value} (No HTML)
            </span>
          );
        }
      },
    },
    {
      key: 'name',
      title: 'Catalog File',
      width: 250,
      minWidth: 150,
    },
    {
      key: 'modified',
      title: 'Modified',
      width: 180,
      minWidth: 140,
      render: (value: string) => new Date(value).toLocaleString(),
    },
  ];

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
    { title: 'Test Catalog 1', name: 'test1.json', modified: new Date().toISOString(), path: '/test/1', hasHtml: true },
    { title: 'Test Catalog 2', name: 'test2.json', modified: new Date().toISOString(), path: '/test/2', hasHtml: false },
  ];

  const actualData = state.browseCatalogs.length > 0 ? state.browseCatalogs : testData;

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
          Catalog Files {state.browseCatalogs.length > 0 ? `(${state.browseCatalogs.length})` : ''}
        </Title>
      </div>
      
      <div style={{ flex: 1, minHeight: 0 }}>
        <ModernTable
          columns={columns}
          data={actualData}
          loading={state.isLoading}
          pageSize={50}
          height="100%"
          emptyText="No catalog files loaded. Click 'Load Catalogs' to get started."
        />
      </div>
    </div>
  );
}

const BrowseCatalogsTab = {
  Sidebar: BrowseCatalogsSidebar,
  Content: BrowseCatalogsContent,
};

export default BrowseCatalogsTab;