import React, { useState, useEffect } from 'react';
import { Layout, Tabs, Button, Modal, Switch, Typography, Space } from 'antd';
import { MenuOutlined, MenuFoldOutlined, SettingOutlined } from '@ant-design/icons';
import { useAppContext } from '../contexts/AppContext';
import CreateCatalogTab from './tabs/CreateCatalogTab';
import SearchCatalogsTab from './tabs/SearchCatalogsTab';
import BrowseCatalogsTab from './tabs/BrowseCatalogsTab';
import WelcomeContent from './WelcomeContent';

const { Sider, Content } = Layout;
const { TabPane } = Tabs;
const { Title, Text } = Typography;

function MainContent() {
  const { state, dispatch } = useAppContext();
  const [collapsed, setCollapsed] = useState(false);
  const [settingsVisible, setSettingsVisible] = useState(false);
  const [isDarkMode, setIsDarkMode] = useState(false);
  const [windowPersistence, setWindowPersistence] = useState(true);

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
    const savedTheme = localStorage.getItem('storcat-theme');
    const isDark = savedTheme === 'dark';
    setIsDarkMode(isDark);

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

  const handleThemeToggle = (dark: boolean) => {
    setIsDarkMode(dark);
    localStorage.setItem('storcat-theme', dark ? 'dark' : 'light');
    
    if (dark) {
      document.documentElement.setAttribute('data-theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }

    // Dispatch custom event to notify App.tsx of theme change
    window.dispatchEvent(new CustomEvent('themeChange', { detail: { isDark: dark } }));
  };

  const handleWindowPersistenceChange = async (enabled: boolean) => {
    setWindowPersistence(enabled);
    
    try {
      await window.electronAPI.setWindowPersistence(enabled);
    } catch (error) {
      console.error('Failed to save window persistence setting:', error);
    }
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

  return (
    <Layout style={{ height: 'calc(100vh - var(--header-height))' }}>
      {/* Icon Bar - Always visible when collapsed */}
      {collapsed && (
        <div style={{
          width: '48px',
          background: 'var(--sidebar-bg)',
          borderRight: '1px solid var(--border-color)',
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
      )}

      <Sider
        width={300}
        collapsed={collapsed}
        collapsedWidth={0}
        trigger={null}
        style={{
          background: 'var(--sidebar-bg)',
          borderRight: '1px solid var(--border-color)',
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
              justifyContent: 'space-between',
              alignItems: 'flex-start',
              minHeight: '48px',
            }}>
              {/* Top left: Toggle button */}
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
              justifyContent: 'flex-start',
              minHeight: '48px',
            }}>
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
            padding: collapsed ? '24px 24px 24px 50px' : '24px',
            paddingLeft: collapsed ? '50px' : '24px',
            background: 'var(--app-bg)',
            overflow: 'auto',
            transition: 'padding-left 0.3s ease',
          }}
        >
          {state.activeTab === 'create' && !state.selectedDirectory && <WelcomeContent />}
          {state.activeTab === 'create' && <CreateCatalogTab.Content />}
          {state.activeTab === 'search' && <SearchCatalogsTab.Content />}
          {state.activeTab === 'browse' && <BrowseCatalogsTab.Content />}
        </Content>
      </Layout>

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
                  Switch between light and dark mode
                </Text>
              </div>
              <div style={{ flexShrink: 0, display: 'flex', alignItems: 'center' }}>
                <Space>
                  <Text style={{ fontSize: '0.875rem', fontWeight: 500, minWidth: 40, textAlign: 'center' }}>
                    Light
                  </Text>
                  <Switch
                    checked={isDarkMode}
                    onChange={handleThemeToggle}
                    size="default"
                  />
                  <Text style={{ fontSize: '0.875rem', fontWeight: 500, minWidth: 40, textAlign: 'center' }}>
                    Dark
                  </Text>
                </Space>
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