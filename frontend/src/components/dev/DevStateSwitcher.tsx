import { useEffect, useState } from 'react';
import { themes } from '../../themes';
import { applyTokens, readPersistedPrefs, THEME_KEY, DENSITY_KEY, Density } from '../../themeTokens';

// This component exists only because Phase 22 ships no theme picker, no
// density toggle, and no rail-side control -- it is the only way to
// exercise all 11 themes and both density values before a real theme
// picker/toggle exist. Delete this file once that surface lands.
function DevStateSwitcher() {
  const [prefs, setPrefs] = useState(() => readPersistedPrefs());

  useEffect(() => {
    const handleKeydown = (event: KeyboardEvent) => {
      if (!event.ctrlKey || !event.altKey) return;

      if (event.key === 't' || event.key === 'T') {
        setPrefs(prev => {
          const currentIndex = themes.findIndex(t => t.id === prev.theme.id);
          const nextTheme = themes[(currentIndex + 1) % themes.length];
          applyTokens(nextTheme, prev.density);
          localStorage.setItem(THEME_KEY, nextTheme.id);
          return { ...prev, theme: nextTheme };
        });
      } else if (event.key === 'd' || event.key === 'D') {
        setPrefs(prev => {
          const nextDensity: Density = prev.density === 'Compact' ? 'Comfortable' : 'Compact';
          applyTokens(prev.theme, nextDensity);
          localStorage.setItem(DENSITY_KEY, nextDensity);
          return { ...prev, density: nextDensity };
        });
      }
    };

    window.addEventListener('keydown', handleKeydown);
    return () => {
      window.removeEventListener('keydown', handleKeydown);
    };
  }, []);

  return (
    <div
      id="storcat-dev-switcher"
      style={{
        position: 'fixed',
        bottom: '8px',
        right: '8px',
        zIndex: 'var(--z-dialog)',
        pointerEvents: 'none',
        fontFamily: "'IBM Plex Mono', ui-monospace, SFMono-Regular, Menlo, monospace",
        fontSize: '11px',
        color: 'var(--fn)',
        background: 'var(--ch)',
        border: '1px solid var(--l)',
        borderRadius: '6px',
        padding: '4px 8px',
      }}
    >
      {prefs.theme.name} · {prefs.density} · {prefs.railSide}
    </div>
  );
}

export default DevStateSwitcher;
