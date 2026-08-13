import { useEffect } from 'react';
import { themes, Theme } from '../../themes';
import { DENSITY_KEY, RAIL_SIDE_KEY, Density, RailSide } from '../../themeTokens';
import { useAppContext } from '../../contexts/AppContext';

export interface DevStateSwitcherProps {
  currentTheme: Theme;
}

// This component exists only because Phase 22 ships no theme picker, no
// density toggle, and no rail-side control -- it is the only way to
// exercise all 11 themes and both density values before a real theme
// picker/toggle exist. Delete this file once that surface lands.
//
// Density and rail side are reducer state (AppContext); this component
// dispatches into it rather than owning a parallel copy, so WorkspaceShell's
// density-reapply effect and the rail-side CSS attribute stay the source of
// truth. Theme isn't reducer state yet (Phase 26), so it's lifted from
// App.tsx as a prop and cycled via the same 'themeChange' CustomEvent the
// future Settings surface will use, rather than owning a second copy.
function DevStateSwitcher({ currentTheme }: DevStateSwitcherProps) {
  const { state, dispatch } = useAppContext();

  useEffect(() => {
    const handleKeydown = (event: KeyboardEvent) => {
      if (!event.ctrlKey || !event.altKey) return;

      if (event.key === 't' || event.key === 'T') {
        const currentIndex = themes.findIndex(t => t.id === currentTheme.id);
        const nextTheme = themes[(currentIndex + 1) % themes.length];
        // App.tsx's listener applies tokens, updates state, and persists to
        // localStorage -- one path, no parallel write here.
        window.dispatchEvent(new CustomEvent('themeChange', { detail: { theme: nextTheme } }));
      } else if (event.key === 'd' || event.key === 'D') {
        const nextDensity: Density = state.density === 'Compact' ? 'Comfortable' : 'Compact';
        dispatch({ type: 'SET_DENSITY', payload: nextDensity });
        localStorage.setItem(DENSITY_KEY, nextDensity);
      } else if (event.key === 'r' || event.key === 'R') {
        const nextRailSide: RailSide = state.railSide === 'Left' ? 'Right' : 'Left';
        dispatch({ type: 'SET_RAIL_SIDE', payload: nextRailSide });
        localStorage.setItem(RAIL_SIDE_KEY, nextRailSide);
      }
    };

    window.addEventListener('keydown', handleKeydown);
    return () => {
      window.removeEventListener('keydown', handleKeydown);
    };
  }, [currentTheme, state.density, state.railSide, dispatch]);

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
      {currentTheme.name} · {state.density} · {state.railSide}
    </div>
  );
}

export default DevStateSwitcher;
