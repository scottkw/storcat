import React, { useState, useEffect } from 'react';
import { Layout, ConfigProvider, theme } from 'antd';
import { AppProvider } from './contexts/AppContext';
import { getThemeById, getDefaultTheme, Theme } from './themes';
import Header from './components/Header';
import MainContent from './components/MainContent';
import CatalogModal from './components/CatalogModal';

const { Content } = Layout;

function App() {
  const [currentTheme, setCurrentTheme] = useState<Theme>(getDefaultTheme());
  const [catalogModalVisible, setCatalogModalVisible] = useState(false);
  const [catalogModalPath, setCatalogModalPath] = useState<string | null>(null);

  // Load theme preference on startup
  useEffect(() => {
    const savedThemeId = localStorage.getItem('storcat-theme-id');
    let themeToLoad = getDefaultTheme();
    
    if (savedThemeId) {
      const savedTheme = getThemeById(savedThemeId);
      if (savedTheme) {
        themeToLoad = savedTheme;
      }
    } else {
      // Migration: handle old theme setting
      const oldTheme = localStorage.getItem('storcat-theme');
      if (oldTheme === 'dark') {
        themeToLoad = getThemeById('storcat-dark') || getDefaultTheme();
        localStorage.setItem('storcat-theme-id', 'storcat-dark');
      } else {
        localStorage.setItem('storcat-theme-id', 'storcat-light');
      }
      localStorage.removeItem('storcat-theme');
    }
    
    setCurrentTheme(themeToLoad);
    applyTheme(themeToLoad);

    // Listen for theme changes from settings modal
    const handleThemeChange = (event: CustomEvent) => {
      const { theme: newTheme } = event.detail;
      if (newTheme) {
        setCurrentTheme(newTheme);
        applyTheme(newTheme);
      }
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

  const applyTheme = (theme: Theme) => {
    // Apply theme to document for CSS custom properties
    document.documentElement.setAttribute('data-theme', theme.id);
    
    // Set CSS custom properties
    const root = document.documentElement;
    root.style.setProperty('--app-bg', theme.colors.appBg);
    root.style.setProperty('--app-text', theme.colors.appText);
    root.style.setProperty('--card-bg', theme.colors.cardBg);
    root.style.setProperty('--border-color', theme.colors.borderColor);
    root.style.setProperty('--header-bg', theme.colors.headerBg);
    root.style.setProperty('--header-text', theme.colors.headerText);
    root.style.setProperty('--sidebar-bg', theme.colors.sidebarBg);
    root.style.setProperty('--table-stripe', theme.colors.tableStripe);
    root.style.setProperty('--table-hover', theme.colors.tableHover);
    root.style.setProperty('--modal-bg', theme.colors.modalBg);
    root.style.setProperty('--shadow-color', theme.colors.shadowColor);
    root.style.setProperty('--input-bg', theme.colors.inputBg);
    root.style.setProperty('--code-bg', theme.colors.codeBg);
    root.style.setProperty('--icon-filter', theme.colors.iconFilter);
    root.style.setProperty('--link-color', theme.colors.linkColor);
    root.style.setProperty('--link-hover', theme.colors.linkHover);
  };

  const handleCloseCatalogModal = () => {
    setCatalogModalVisible(false);
    setCatalogModalPath(null);
  };

  return (
    <ConfigProvider
      theme={{
        algorithm: currentTheme.antdAlgorithm === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: currentTheme.antdPrimaryColor || '#5D6569FF',
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