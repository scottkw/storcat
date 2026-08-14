import { models } from '../../../../wailsjs/go/models';
import { formatBytes } from '../../../lib/format';

export interface PaletteResultRowProps {
  result: models.SearchResult;
  query: string;
  active: boolean;
  optionId: string;
  onActivate: () => void;
  onHover: () => void;
}

// One listbox option: shape, highlighted basename, dimmed path, catalog
// chip, size. The highlight is built from JSX text children only -- never
// an HTML string -- because a catalog on disk can legitimately contain a
// filename with angle brackets or an ampersand, and this palette runs
// inside the webview that holds the window.go bindings (T-24-11).
function PaletteResultRow({ result, query, active, optionId, onActivate, onHover }: PaletteResultRowProps) {
  const trimmedQuery = query.trim();
  const lowerBasename = result.basename.toLowerCase();
  const lowerQuery = trimmedQuery.toLowerCase();
  const matchIndex = lowerQuery.length > 0 ? lowerBasename.indexOf(lowerQuery) : -1;

  let name: React.ReactNode = result.basename;
  if (matchIndex >= 0) {
    const before = result.basename.slice(0, matchIndex);
    const match = result.basename.slice(matchIndex, matchIndex + trimmedQuery.length);
    const after = result.basename.slice(matchIndex + trimmedQuery.length);
    name = (
      <>
        {before}
        <span style={{ color: 'var(--ac)', fontWeight: 600 }}>{match}</span>
        {after}
      </>
    );
  }

  return (
    <div
      className="ws-palette-row"
      role="option"
      id={optionId}
      aria-selected={active}
      data-active={active || undefined}
      onClick={onActivate}
      onMouseMove={onHover}
    >
      <span className="ws-palette-shape" aria-hidden="true" data-kind={result.type === 'directory' ? 'directory' : 'file'} />
      <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 2 }}>
        <span className="ws-palette-name mono">{name}</span>
        <span className="ws-palette-path mono">{result.fullPath}</span>
      </div>
      <span className="ws-palette-chip mono">{result.catalog}</span>
      <span className="ws-palette-size mono">{formatBytes(result.size)}</span>
    </div>
  );
}

export default PaletteResultRow;
