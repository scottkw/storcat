import React, { useState, useEffect } from 'react';
import { Layout, ConfigProvider, theme } from 'antd';
import { AppProvider } from './contexts/AppContext';
import Header from './components/Header';
import MainContent from './components/MainContent';
import CatalogModal from './components/CatalogModal';

const { Content } = Layout;

function App() {
  const [isDarkMode, setIsDarkMode] = useState(false);
  const [catalogModalVisible, setCatalogModalVisible] = useState(false);
  const [catalogModalPath, setCatalogModalPath] = useState<string | null>(null);

  // Load theme preference on startup
  useEffect(() => {
    const savedTheme = localStorage.getItem('storcat-theme');
    const isDark = savedTheme === 'dark';
    setIsDarkMode(isDark);
    
    // Apply theme to document
    if (isDark) {
      document.documentElement.setAttribute('data-theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }

    // Listen for theme changes from settings modal
    const handleThemeChange = (event: CustomEvent) => {
      setIsDarkMode(event.detail.isDark);
    };

    window.addEventListener('themeChange', handleThemeChange as EventListener);
    
    // Listen for catalog modal events
    const handleOpenCatalog = (event: CustomEvent) => {
      setCatalogModalPath(event.detail.catalogPath);
      setCatalogModalVisible(true);
    };

    window.addEventListener('openCatalogModal', handleOpenCatalog as EventListener);
    
    return () => {
      window.removeEventListener('themeChange', handleThemeChange as EventListener);
      window.removeEventListener('openCatalogModal', handleOpenCatalog as EventListener);
    };
  }, []);

  const toggleTheme = (dark: boolean) => {
    setIsDarkMode(dark);
    localStorage.setItem('storcat-theme', dark ? 'dark' : 'light');
    
    if (dark) {
      document.documentElement.setAttribute('data-theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }
  };

  const handleCloseCatalogModal = () => {
    setCatalogModalVisible(false);
    setCatalogModalPath(null);
  };

  return (
    <ConfigProvider
      theme={{
        algorithm: isDarkMode ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: '#5D6569FF',
          fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        },
      }}
    >
      <AppProvider>
        <Layout style={{ height: '100vh' }}>
          <Header />
          <Content style={{ overflow: 'hidden' }}>
            <MainContent />
          </Content>
        </Layout>
        <CatalogModal
          visible={catalogModalVisible}
          catalogPath={catalogModalPath}
          onClose={handleCloseCatalogModal}
        />
      </AppProvider>
    </ConfigProvider>
  );
}

export default App;