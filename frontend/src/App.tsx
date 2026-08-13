import { useState, useEffect } from 'react';
import { ConfigProvider, theme as antdTheme } from 'antd';
import { AppProvider } from './contexts/AppContext';
import { Theme } from './themes';
import { readPersistedPrefs, applyTokens } from './themeTokens';
import CatalogModal from './components/CatalogModal';
import WorkspaceShell from './components/workspace/WorkspaceShell';
import DevStateSwitcher from './components/dev/DevStateSwitcher';
import './services/wailsAPI'; // Initialize Wails API wrapper

function App() {
  const [currentTheme, setCurrentTheme] = useState<Theme>(() => readPersistedPrefs().theme);
  const [catalogModalVisible, setCatalogModalVisible] = useState(false);
  const [catalogModalPath, setCatalogModalPath] = useState<string | null>(null);

  useEffect(() => {
    // Listen for theme changes from settings modal
    const handleThemeChange = (event: CustomEvent) => {
      const { theme: newTheme } = event.detail;
      if (newTheme) {
        setCurrentTheme(newTheme);
        applyTokens(newTheme, readPersistedPrefs().density);
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

  const handleCloseCatalogModal = () => {
    setCatalogModalVisible(false);
    setCatalogModalPath(null);
  };

  return (
    <ConfigProvider
      theme={{
        algorithm: currentTheme.antdAlgorithm === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        token: {
          colorPrimary: currentTheme.antdPrimaryColor || '#5D6569FF',
          fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        },
      }}
    >
      <AppProvider>
        <WorkspaceShell />
        <CatalogModal
          visible={catalogModalVisible}
          catalogPath={catalogModalPath}
          onClose={handleCloseCatalogModal}
        />
        {import.meta.env.DEV && <DevStateSwitcher />}
      </AppProvider>
    </ConfigProvider>
  );
}

export default App;
