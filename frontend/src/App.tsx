import { useState, useEffect } from 'react';
import { ConfigProvider, theme as antdTheme } from 'antd';
import { AppProvider } from './contexts/AppContext';
import { Theme } from './themes';
import { readPersistedPrefs, applyTokens, THEME_KEY } from './themeTokens';
import WorkspaceShell from './components/workspace/WorkspaceShell';
import DevStateSwitcher from './components/dev/DevStateSwitcher';
import './services/wailsAPI'; // Initialize Wails API wrapper

function App() {
  const [currentTheme, setCurrentTheme] = useState<Theme>(() => readPersistedPrefs().theme);

  useEffect(() => {
    // Listen for theme changes -- both the future Settings surface (Phase 26)
    // and the shipped Wave-0 dev affordance (Ctrl+Alt+T) dispatch this event,
    // so there is one path that applies tokens, updates state, and persists.
    const handleThemeChange = (event: CustomEvent) => {
      const { theme: newTheme } = event.detail;
      if (newTheme) {
        setCurrentTheme(newTheme);
        localStorage.setItem(THEME_KEY, newTheme.id);
        applyTokens(newTheme, readPersistedPrefs().density);
      }
    };

    window.addEventListener('themeChange', handleThemeChange as EventListener);

    return () => {
      window.removeEventListener('themeChange', handleThemeChange as EventListener);
    };
  }, []);

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
        <WorkspaceShell themeName={currentTheme.name} />
        {import.meta.env.DEV && <DevStateSwitcher currentTheme={currentTheme} />}
      </AppProvider>
    </ConfigProvider>
  );
}

export default App;
