import React, { useState, useEffect } from 'react';
import { Button, Input, Checkbox, message, Card, Typography, Space, Divider } from 'antd';
import { FolderOutlined, FolderAddOutlined, SettingOutlined, InfoCircleOutlined, GlobalOutlined } from '@ant-design/icons';
import storcatIcon from '../../storcat-icon.svg';
import { useAppContext } from '../../contexts/AppContext';
import packageJson from '../../../../package.json';

const { Title, Text, Paragraph } = Typography;

function CreateCatalogSidebar() {
  const { state, dispatch } = useAppContext();
  const [catalogTitle, setCatalogTitle] = useState('');
  const [outputRoot, setOutputRoot] = useState('catalog');
  const [copyToSecondary, setCopyToSecondary] = useState(false);

  // Load saved preferences on mount
  useEffect(() => {
    const savedCreateDirectory = localStorage.getItem('storcat-last-create-directory');
    const savedOutputDirectory = localStorage.getItem('storcat-last-output-directory');
    
    if (savedCreateDirectory) {
      dispatch({ type: 'SET_SELECTED_DIRECTORY', payload: savedCreateDirectory });
    }
    if (savedOutputDirectory) {
      dispatch({ type: 'SET_SELECTED_OUTPUT_DIRECTORY', payload: savedOutputDirectory });
    }
  }, [dispatch]);

  const selectDirectory = async () => {
    try {
      const directory = await window.electronAPI.selectDirectory();
      if (directory) {
        dispatch({ type: 'SET_SELECTED_DIRECTORY', payload: directory });
        localStorage.setItem('storcat-last-create-directory', directory);
      }
    } catch (error) {
      message.error('Failed to select directory');
    }
  };

  const selectOutputDirectory = async () => {
    try {
      const directory = await window.electronAPI.selectOutputDirectory();
      if (directory) {
        dispatch({ type: 'SET_SELECTED_OUTPUT_DIRECTORY', payload: directory });
        localStorage.setItem('storcat-last-output-directory', directory);
      }
    } catch (error) {
      message.error('Failed to select output directory');
    }
  };

  const createCatalog = async () => {
    if (!catalogTitle.trim()) {
      message.error('Please enter a catalog title');
      return;
    }
    if (!state.selectedDirectory) {
      message.error('Please select a directory to catalog');
      return;
    }
    if (!outputRoot.trim()) {
      message.error('Please enter an output filename root');
      return;
    }

    dispatch({ type: 'SET_IS_CREATING', payload: true });

    try {
      const options = {
        title: catalogTitle.trim(),
        directoryPath: state.selectedDirectory,
        outputRoot: outputRoot.trim(),
        copyToDirectory: copyToSecondary ? state.selectedOutputDirectory : null,
      };

      const result = await window.electronAPI.createCatalog(options);

      if (result.success) {
        message.success(`Catalog created successfully! ${result.fileCount} files cataloged.`);
      } else {
        message.error(`Failed to create catalog: ${result.error}`);
      }
    } catch (error) {
      message.error('Failed to create catalog');
    } finally {
      dispatch({ type: 'SET_IS_CREATING', payload: false });
    }
  };

  const canCreate = catalogTitle.trim() && state.selectedDirectory && outputRoot.trim();

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {/* Catalog Title */}
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
          <label style={{ fontSize: '1rem', fontWeight: 500, color: 'var(--app-text)', margin: 0 }}>
            Catalog Title
          </label>
          <InfoCircleOutlined 
            style={{ fontSize: 16, color: 'var(--sl-color-neutral-500)' }}
            title="This will be displayed in the HTML output"
          />
        </div>
        <Input
          value={catalogTitle}
          onChange={(e) => setCatalogTitle(e.target.value)}
          placeholder="Enter catalog title"
          style={{ width: '100%' }}
        />
      </div>

      {/* Select Directory */}
      <div>
        <Button
          type="primary"
          icon={<FolderOutlined />}
          onClick={selectDirectory}
          style={{ width: '100%', justifyContent: 'center' }}
        >
          Select Directory
        </Button>
        {state.selectedDirectory && (
          <div className="path-display" style={{ marginTop: 8 }}>
            {state.selectedDirectory}
          </div>
        )}
      </div>

      {/* Output Filename Root */}
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
          <label style={{ fontSize: '1rem', fontWeight: 500, color: 'var(--app-text)', margin: 0 }}>
            Output Filename Root
          </label>
          <InfoCircleOutlined 
            style={{ fontSize: 16, color: 'var(--sl-color-neutral-500)' }}
            title="Files will be saved as [root].json and [root].html"
          />
        </div>
        <Input
          value={outputRoot}
          onChange={(e) => setOutputRoot(e.target.value)}
          placeholder="catalog"
          style={{ width: '100%' }}
        />
      </div>

      {/* Copy to Secondary Location */}
      <div>
        <Checkbox
          checked={copyToSecondary}
          onChange={(e) => setCopyToSecondary(e.target.checked)}
          style={{ fontSize: '0.8rem' }}
        >
          Copy files to secondary location
        </Checkbox>
      </div>

      {/* Secondary Location */}
      {copyToSecondary && (
        <div style={{ marginTop: 16 }}>
          <Button
            type="primary"
            icon={<FolderAddOutlined />}
            onClick={selectOutputDirectory}
            style={{ width: '100%', justifyContent: 'center' }}
          >
            Select Output Directory
          </Button>
          {state.selectedOutputDirectory && (
            <div className="path-display" style={{ marginTop: 8 }}>
              {state.selectedOutputDirectory}
            </div>
          )}
        </div>
      )}

      {/* Create Button */}
      <Button
        type="primary"
        icon={<SettingOutlined />}
        onClick={createCatalog}
        loading={state.isCreating}
        disabled={!canCreate}
        style={{ width: '100%', justifyContent: 'center', marginTop: 24 }}
      >
        Create Catalog
      </Button>
    </Space>
  );
}

function CreateCatalogContent() {
  return (
    <div style={{
      height: '100%',
      background: 'var(--app-bg)',
      overflowY: 'auto',
      padding: '24px',
    }}>
      {/* Logo and Version Block - Full Width at Top */}
      <Card 
        style={{
          width: '100%',
          textAlign: 'center',
          background: 'var(--card-bg)',
          border: '1px solid var(--border-color)',
          boxShadow: '0 2px 8px var(--shadow-color)',
          marginBottom: '24px',
        }}
        bodyStyle={{ padding: '32px' }}
      >
        {/* Logo and App Name */}
        <div style={{ marginBottom: '24px' }}>
          <img
            src={storcatIcon}
            alt="StorCat Logo"
            style={{
              width: '80px',
              height: '80px',
              marginBottom: '16px',
              filter: 'var(--icon-filter)',
            }}
          />
          <Title level={1} style={{ 
            margin: '0 0 8px 0', 
            color: 'var(--app-text)',
            fontSize: '2.5rem',
            fontWeight: 700 
          }}>
            StorCat
          </Title>
          <Text style={{ 
            fontSize: '1.1rem', 
            color: 'var(--app-text)',
            opacity: 0.7 
          }}>
            Directory Catalog Manager
          </Text>
        </div>

        <Divider style={{ borderColor: 'var(--border-color)', margin: '24px 0' }} />

        {/* Version and Author Info */}
        <Space direction="vertical" size="small" style={{ width: '100%', marginBottom: '24px' }}>
          <Text strong style={{ color: 'var(--app-text)', fontSize: '1rem' }}>
            Version {packageJson.version}
          </Text>
          <Text style={{ color: 'var(--app-text)', opacity: 0.8 }}>
            Created by Ken Scott
          </Text>
          <div style={{ marginTop: '12px' }}>
            <Space direction="horizontal" size="large" wrap>
              <Text 
                style={{ color: 'var(--app-text)', opacity: 0.7, cursor: 'pointer' }}
                onClick={() => window.electronAPI?.openExternal('https://github.com/scottkw/storcat/')}
              >
                <GlobalOutlined /> GitHub Repository
              </Text>
              <Text 
                style={{ color: 'var(--app-text)', opacity: 0.7, cursor: 'pointer' }}
                onClick={() => window.electronAPI?.openExternal(`https://github.com/scottkw/storcat/releases/tag/${packageJson.version}`)}
              >
                📋 Release Notes
              </Text>
              <Text 
                style={{ color: 'var(--app-text)', opacity: 0.7, cursor: 'pointer' }}
                onClick={() => window.electronAPI?.openExternal('https://github.com/scottkw/storcat/wiki')}
              >
                📖 Wiki & Documentation
              </Text>
            </Space>
          </div>
        </Space>

        <Divider style={{ borderColor: 'var(--border-color)', margin: '24px 0' }} />

        {/* Description and Instructions */}
        <div style={{ textAlign: 'left' }}>
          <Paragraph style={{ 
            color: 'var(--app-text)', 
            fontSize: '1rem',
            lineHeight: '1.6',
            marginBottom: '20px'
          }}>
            StorCat is a modern directory catalog management application that helps you create, 
            search, and browse comprehensive catalogs of your file systems. Generate both JSON 
            and HTML representations of directory structures for easy documentation and sharing.
          </Paragraph>

          <Title level={4} style={{ 
            color: 'var(--app-text)', 
            marginBottom: '12px',
            fontSize: '1.1rem'
          }}>
            Getting Started:
          </Title>

          <Space direction="vertical" size="small" style={{ width: '100%' }}>
            <Text style={{ color: 'var(--app-text)', opacity: 0.9 }}>
              <strong>1. Create Catalog:</strong> Use the sidebar to select a directory and generate a comprehensive catalog
            </Text>
            <Text style={{ color: 'var(--app-text)', opacity: 0.9 }}>
              <strong>2. Search Catalogs:</strong> Switch to the Search tab to find files across multiple catalog files
            </Text>
            <Text style={{ color: 'var(--app-text)', opacity: 0.9 }}>
              <strong>3. Browse Catalogs:</strong> Use the Browse tab to view and manage your catalog collections
            </Text>
          </Space>
        </div>
      </Card>
    </div>
  );
}

const CreateCatalogTab = {
  Sidebar: CreateCatalogSidebar,
  Content: CreateCatalogContent,
};

export default CreateCatalogTab;