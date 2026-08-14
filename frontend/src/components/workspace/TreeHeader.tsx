import { useAppContext } from '../../contexts/AppContext';
import { formatBytes, formatCount, formatDate } from '../../lib/format';

// Shared chip visual -- the .json chip always renders, the .html chip only
// when hasHtml says the companion file exists (no fabricated/greyed chip).
const chipStyle = {
  fontSize: 11,
  color: 'var(--dm)',
  background: 'var(--ch)',
  borderRadius: 5,
  padding: '2px 7px',
  flex: 'none',
  whiteSpace: 'nowrap',
} as const;

// Renders only in the tree's 'ready' state -- never while a load is in
// flight (no placeholder title/dash metadata) and never for a catalog that
// failed to load (TreePane routes that to UnreadableCatalogPanel instead,
// before this component is ever mounted). The reducer's TREE_LOADED guard
// (catalogId must still match currentCatalogId) is what keeps this from
// ever repainting with a superseded catalog's title or counts.
function TreeHeader() {
  const { state } = useAppContext();

  if (state.tree.status !== 'ready') return null;

  const catalog = state.catalogs.find((c) => c.path === state.currentCatalogId);
  if (!catalog) return null;

  const htmlFilename = catalog.filename.replace(/\.json$/, '.html');

  return (
    <div
      style={{
        padding: '14px 18px 12px',
        display: 'flex',
        flexDirection: 'column',
        gap: 9,
        borderBottom: '1px solid var(--l)',
        minWidth: 0,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
        <span
          style={{
            fontSize: 17,
            fontWeight: 600,
            letterSpacing: '-0.01em',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            minWidth: 0,
            flex: '1 1 auto',
          }}
        >
          {catalog.title}
        </span>
        <span className="mono" style={chipStyle}>
          {catalog.filename}
        </span>
        {/* No HTML chip at all when the companion doesn't exist -- not a
            greyed/disabled one. hasHtml is the only signal consulted; no new
            existence-check capability is added here. */}
        {catalog.hasHtml ? (
          <span className="mono" style={chipStyle}>
            {htmlFilename}
          </span>
        ) : null}
      </div>
      <div
        className="mono"
        style={{ display: 'flex', alignItems: 'center', gap: 14, fontSize: 11.5, color: 'var(--dm)' }}
      >
        {/* fileCount/totalBytes come from the loaded FlatCatalog (state.tree),
            not the rail's cache-backed CatalogMetadata fields -- the loaded
            values are always exact and never absent. */}
        <span>{formatCount(state.tree.fileCount)} files</span>
        <span style={{ color: 'var(--l)' }}>|</span>
        <span>{formatBytes(catalog.size)}</span>
        <span style={{ color: 'var(--l)' }}>|</span>
        <span>{formatBytes(state.tree.totalBytes)} catalogued</span>
        <span style={{ color: 'var(--l)' }}>|</span>
        <span>modified {formatDate(catalog.modified)}</span>
      </div>
    </div>
  );
}

export default TreeHeader;
