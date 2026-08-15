// 11 theme cards in themes.ts's declared order (never re-sorted), each a
// 4-swatch strip + name + light/dark tag. Click applies immediately via the
// caller-supplied onSelect (settingsStore.setThemeSetting) -- no local
// "unsaved" state, no dialog close.

import { themes, Theme } from '../../../themes';

export interface ThemeGridProps {
  activeThemeId: string;
  onSelect: (theme: Theme) => void;
}

function ThemeGrid({ activeThemeId, onSelect }: ThemeGridProps) {
  return (
    <div className="ws-theme-grid">
      {themes.map((theme) => {
        const active = theme.id === activeThemeId;
        return (
          <button
            key={theme.id}
            type="button"
            className={`ws-theme-card${active ? ' ws-theme-card-active' : ''}`}
            aria-pressed={active}
            onClick={() => onSelect(theme)}
          >
            <span className="ws-theme-swatches" aria-hidden="true">
              <span
                className="ws-theme-swatch ws-theme-swatch-first"
                style={{ background: theme.tokens.bg }}
              />
              <span className="ws-theme-swatch" style={{ background: theme.tokens.p2 }} />
              <span className="ws-theme-swatch" style={{ background: theme.tokens.ac }} />
              <span
                className="ws-theme-swatch ws-theme-swatch-last"
                style={{ background: theme.tokens.tx }}
              />
            </span>
            <span className="ws-theme-name">{theme.name}</span>
            <span className="ws-theme-tag mono">{theme.type}</span>
          </button>
        );
      })}
    </div>
  );
}

export default ThemeGrid;
