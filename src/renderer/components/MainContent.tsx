import React, { useState, useEffect } from 'react';
import { Layout, Tabs, Button, Modal, Select, Typography, Switch, Radio } from 'antd';
import { MenuOutlined, MenuFoldOutlined, SettingOutlined } from '@ant-design/icons';
import { useAppContext } from '../contexts/AppContext';
import { themes, getThemeById } from '../themes';
import CreateCatalogTab from './tabs/CreateCatalogTab';
import SearchCatalogsTab from './tabs/SearchCatalogsTab';
import BrowseCatalogsTab from './tabs/BrowseCatalogsTab';
import WelcomeContent from './WelcomeContent';

const { Sider, Content } = Layout;
const { Title, Text } = Typography;

function MainContent() {
  const { state, dispatch } = useAppContext();
  const [collapsed, setCollapsed] = useState(false);
  const [settingsVisible, setSettingsVisible] = useState(false);
  const [selectedThemeId, setSelectedThemeId] = useState('storcat-light');
  const [windowPersistence, setWindowPersistence] = useState(true);
  const [sidebarPosition, setSidebarPosition] = useState<'left' | 'right'>('left');

  const handleTabChange = (activeKey: string) => {
    dispatch({ type: 'SET_ACTIVE_TAB', payload: activeKey });
  };

  const toggleSidebar = () => {
    setCollapsed(!collapsed);
    dispatch({ type: 'SET_SIDEBAR_COLLAPSED', payload: !collapsed });
  };

  // Load settings on startup
  useEffect(() => {
    // Load theme preference
    const savedThemeId = localStorage.getItem('storcat-theme-id');
    if (savedThemeId && getThemeById(savedThemeId)) {
      setSelectedThemeId(savedThemeId);
    } else {
      // Migration: convert old theme setting to new theme ID
      const oldTheme = localStorage.getItem('storcat-theme');
      if (oldTheme === 'dark') {
        setSelectedThemeId('storcat-dark');
        localStorage.setItem('storcat-theme-id', 'storcat-dark');
      } else {
        setSelectedThemeId('storcat-light');
        localStorage.setItem('storcat-theme-id', 'storcat-light');
      }
      // Remove old theme setting
      localStorage.removeItem('storcat-theme');
    }

    // Load sidebar position preference
    const savedSidebarPosition = localStorage.getItem('storcat-sidebar-position') as 'left' | 'right';
    if (savedSidebarPosition && (savedSidebarPosition === 'left' || savedSidebarPosition === 'right')) {
      setSidebarPosition(savedSidebarPosition);
      dispatch({ type: 'SET_SIDEBAR_POSITION', payload: savedSidebarPosition });
    }

    // Load window persistence setting
    const loadWindowPersistence = async () => {
      try {
        const result = await window.electronAPI.getWindowPersistence();
        if (result.success) {
          setWindowPersistence(result.enabled);
        }
      } catch (error) {
        console.warn('Failed to load window persistence setting:', error);
      }
    };
    loadWindowPersistence();
  }, []);

  const handleThemeChange = (themeId: string) => {
    const theme = getThemeById(themeId);
    if (!theme) return;

    setSelectedThemeId(themeId);
    localStorage.setItem('storcat-theme-id', themeId);
    
    // Set data-theme attribute for CSS custom properties
    if (theme.type === 'dark') {
      document.documentElement.setAttribute('data-theme', themeId);
    } else {
      document.documentElement.setAttribute('data-theme', themeId);
    }

    // Dispatch custom event to notify App.tsx of theme change
    window.dispatchEvent(new CustomEvent('themeChange', { 
      detail: { 
        themeId,
        theme,
        isDark: theme.type === 'dark' 
      } 
    }));
  };

  const handleWindowPersistenceChange = async (enabled: boolean) => {
    setWindowPersistence(enabled);
    
    try {
      await window.electronAPI.setWindowPersistence(enabled);
    } catch (error) {
      console.error('Failed to save window persistence setting:', error);
    }
  };

  const handleSidebarPositionChange = (position: 'left' | 'right') => {
    setSidebarPosition(position);
    localStorage.setItem('storcat-sidebar-position', position);
    dispatch({ type: 'SET_SIDEBAR_POSITION', payload: position });
  };

  const tabs = [
    {
      key: 'create',
      label: 'Create Catalog',
      children: <CreateCatalogTab />,
    },
    {
      key: 'search',
      label: 'Search Catalogs',
      children: <SearchCatalogsTab />,
    },
    {
      key: 'browse',
      label: 'Browse Catalogs', 
      children: <BrowseCatalogsTab />,
    },
  ];

  const renderIconBar = () => (
    collapsed && (
      <div style={{
        width: '48px',
        background: 'var(--sidebar-bg)',
        borderRight: state.sidebarPosition === 'left' ? '1px solid var(--border-color)' : 'none',
        borderLeft: state.sidebarPosition === 'right' ? '1px solid var(--border-color)' : 'none',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        zIndex: 999,
      }}>
        {/* Top: Toggle button */}
        <div style={{ padding: '12px' }}>
          <Button
            type="text"
            shape="circle"
            size="small"
            icon={<MenuOutlined />}
            onClick={toggleSidebar}
            style={{
              width: 24,
              height: 24,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '12px',
            }}
            title="Open Sidebar"
          />
        </div>
        
        {/* Bottom: Settings button */}
        <div style={{ padding: '12px' }}>
          <Button
            type="text"
            shape="circle"
            size="small"
            icon={<SettingOutlined />}
            onClick={() => setSettingsVisible(true)}
            style={{
              width: 24,
              height: 24,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '12px',
            }}
            title="Settings"
          />
        </div>
      </div>
    )
  );

  const renderSidebar = () => (
    <Sider
      width={300}
      collapsed={collapsed}
      collapsedWidth={0}
      trigger={null}
      style={{
        background: 'var(--sidebar-bg)',
        borderRight: state.sidebarPosition === 'left' ? '1px solid var(--border-color)' : 'none',
        borderLeft: state.sidebarPosition === 'right' ? '1px solid var(--border-color)' : 'none',
        transition: 'all 0.3s ease',
        overflow: 'hidden',
      }}
    >
      {!collapsed && (
        <>
          {/* Sidebar Header with Toggle and Settings */}
          <div style={{
            padding: '12px',
            borderBottom: '1px solid var(--border-color)',
            display: 'flex',
            justifyContent: state.sidebarPosition === 'left' ? 'flex-start' : 'flex-end',
            alignItems: 'flex-start',
            minHeight: '48px',
          }}>
            {/* Toggle button positioned based on sidebar location */}
            <Button
              type="text"
              shape="circle"
              size="small"
              icon={<MenuFoldOutlined />}
              onClick={toggleSidebar}
              style={{
                width: 24,
                height: 24,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '12px',
              }}
              title="Close Sidebar"
            />
          </div>

          {/* Sidebar Content */}
          <div style={{ 
            padding: '24px 24px 24px 24px',
            height: 'calc(100% - 48px - 48px)', // Account for header and footer
            overflow: 'auto',
            display: 'flex',
            flexDirection: 'column'
          }}>
            <div style={{ flex: 1 }}>
              {state.activeTab === 'create' && <CreateCatalogTab.Sidebar />}
              {state.activeTab === 'search' && <SearchCatalogsTab.Sidebar />}
              {state.activeTab === 'browse' && <BrowseCatalogsTab.Sidebar />}
            </div>
          </div>

          {/* Sidebar Footer with Settings */}
          <div style={{
            padding: '12px',
            borderTop: '1px solid var(--border-color)',
            display: 'flex',
            justifyContent: state.sidebarPosition === 'left' ? 'flex-start' : 'flex-end',
            minHeight: '48px',
          }}>
            {/* Settings button positioned based on sidebar location */}
            <Button
              type="text"
              shape="circle"
              size="small"
              icon={<SettingOutlined />}
              onClick={() => setSettingsVisible(true)}
              style={{
                width: 24,
                height: 24,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '12px',
              }}
              title="Settings"
            />
          </div>
        </>
      )}
    </Sider>
  );

  const renderMainLayout = () => (
    <Layout>
      {/* Tab Navigation */}
      <div
        style={{
          height: 'var(--tab-nav-height)',
          background: 'var(--sidebar-bg)',
          borderBottom: '1px solid var(--border-color)',
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
        }}
      >
        <Tabs
          activeKey={state.activeTab}
          onChange={handleTabChange}
          style={{ width: '100%' }}
          items={tabs.map(tab => ({
            key: tab.key,
            label: tab.label,
          }))}
        />
      </div>

      {/* Main Content Area */}
      <Content
        style={{
          padding: '24px',
          paddingLeft: collapsed && state.sidebarPosition === 'left' ? '50px' : '24px',
          paddingRight: collapsed && state.sidebarPosition === 'right' ? '50px' : '24px',
          background: 'var(--app-bg)',
          overflow: 'auto',
          transition: 'padding 0.3s ease',
        }}
      >
        {state.activeTab === 'create' && !state.selectedDirectory && <WelcomeContent />}
        {state.activeTab === 'create' && <CreateCatalogTab.Content />}
        {state.activeTab === 'search' && <SearchCatalogsTab.Content />}
        {state.activeTab === 'browse' && <BrowseCatalogsTab.Content />}
      </Content>
    </Layout>
  );

  return (
    <Layout style={{ height: 'calc(100vh - var(--header-height))' }}>
      {state.sidebarPosition === 'left' && renderIconBar()}
      {state.sidebarPosition === 'left' && renderSidebar()}
      {renderMainLayout()}
      {state.sidebarPosition === 'right' && renderSidebar()}
      {state.sidebarPosition === 'right' && renderIconBar()}

      {/* Settings Modal */}
      <Modal
        title="Settings"
        open={settingsVisible}
        onCancel={() => setSettingsVisible(false)}
        footer={null}
        width={600}
        styles={{
          body: { maxHeight: '60vh', overflowY: 'auto' },
        }}
      >
        <div style={{ padding: '8px 0' }}>
          {/* Appearance Section */}
          <div style={{ marginBottom: 32 }}>
            <Title level={5} style={{ margin: '0 0 16px 0', paddingBottom: 8, borderBottom: '1px solid var(--border-color)' }}>
              Appearance
            </Title>
            
            <div style={{
              display: 'flex',
              alignItems: 'flex-start',
              justifyContent: 'space-between',
              gap: 16,
              padding: '16px 0',
              borderBottom: '1px solid var(--border-color)'
            }}>
              <div style={{ flex: 1 }}>
                <Text strong style={{ display: 'block', marginBottom: 4 }}>
                  Theme
                </Text>
                <Text type="secondary" style={{ fontSize: '0.875rem', lineHeight: 1.4 }}>
                  Choose your preferred application theme
                </Text>
              </div>
              <div style={{ flexShrink: 0, display: 'flex', alignItems: 'center' }}>
                <Select
                  value={selectedThemeId}
                  onChange={handleThemeChange}
                  style={{ minWidth: 200 }}
                  placeholder="Select theme"
                >
                  {themes.map(theme => (
                    <Select.Option key={theme.id} value={theme.id}>
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <span>{theme.name}</span>
                        <span style={{ 
                          fontSize: '0.75rem', 
                          opacity: 0.6, 
                          marginLeft: 8,
                          textTransform: 'capitalize' 
                        }}>
                          {theme.type}
                        </span>
                      </div>
                    </Select.Option>
                  ))}
                </Select>
              </div>
            </div>

            <div style={{
              display: 'flex',
              alignItems: 'flex-start',
              justifyContent: 'space-between',
              gap: 16,
              padding: '16px 0',
              borderBottom: '1px solid var(--border-color)'
            }}>
              <div style={{ flex: 1 }}>
                <Text strong style={{ display: 'block', marginBottom: 4 }}>
                  Sidebar Position
                </Text>
                <Text type="secondary" style={{ fontSize: '0.875rem', lineHeight: 1.4 }}>
                  Choose whether the sidebar appears on the left or right side
                </Text>
              </div>
              <div style={{ flexShrink: 0, display: 'flex', alignItems: 'center' }}>
                <Radio.Group
                  value={sidebarPosition}
                  onChange={(e) => handleSidebarPositionChange(e.target.value)}
                  size="small"
                >
                  <Radio.Button value="left">Left</Radio.Button>
                  <Radio.Button value="right">Right</Radio.Button>
                </Radio.Group>
              </div>
            </div>
          </div>

          {/* General Section */}
          <div style={{ marginBottom: 0 }}>
            <Title level={5} style={{ margin: '0 0 16px 0', paddingBottom: 8, borderBottom: '1px solid var(--border-color)' }}>
              General
            </Title>
            
            <div style={{
              display: 'flex',
              alignItems: 'flex-start',
              justifyContent: 'space-between',
              gap: 16,
              padding: '16px 0 0 0',
            }}>
              <div style={{ flex: 1 }}>
                <Text strong style={{ display: 'block', marginBottom: 4 }}>
                  Remember Window Size & Position
                </Text>
                <Text type="secondary" style={{ fontSize: '0.875rem', lineHeight: 1.4 }}>
                  Automatically restore window size and position when the app starts
                </Text>
              </div>
              <div style={{ flexShrink: 0, display: 'flex', alignItems: 'center' }}>
                <Switch
                  checked={windowPersistence}
                  onChange={handleWindowPersistenceChange}
                  size="default"
                />
              </div>
            </div>
          </div>
        </div>
      </Modal>
    </Layout>
  );
}

export default MainContent;