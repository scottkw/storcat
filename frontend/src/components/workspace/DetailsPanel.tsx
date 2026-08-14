import { useAppContext } from '../../contexts/AppContext';
import { formatBytes, formatCount, formatDate } from '../../lib/format';

export interface DetailsPanelProps {
  variant?: 'pane' | 'drawer';
}

type MetaRow = [label: string, value: string];

// Shared by both populated states -- label left (--dm), value right (mono,
// ellipsized via the existing ws-meta-value contract so a long value can
// never widen the panel).
function MetaRows({ rows }: { rows: MetaRow[] }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
      {rows.map(([label, value]) => (
        <div
          key={label}
          className="ws-meta-row"
          style={{ display: 'flex', justifyContent: 'space-between', gap: 12, borderBottom: '1px solid var(--l2)' }}
        >
          <span style={{ fontSize: 11.5, color: 'var(--dm)' }}>{label}</span>
          <span className="ws-meta-value mono" style={{ fontSize: 11.5, textAlign: 'right' }}>
            {value}
          </span>
        </div>
      ))}
    </div>
  );
}

// Renders structurally in both populated states, matching the handoff's
// header row -- but stays inert this phase: no handler, no menu. The
// actions menu it will open is ACT-01 (Phase 27). Because it is icon-only
// it carries an explicit accessible name now; menu-related ARIA attributes
// are deliberately withheld -- there is no menu yet, and claiming one would
// misinform assistive tech.
function OverflowButton() {
  return (
    <button
      type="button"
      className="ws-details-overflow"
      aria-label="Catalog actions"
      style={{
        marginLeft: 'auto',
        flex: 'none',
        width: 22,
        height: 22,
        borderRadius: 6,
        border: '1px solid var(--l)',
        background: 'transparent',
        color: 'var(--dm)',
        fontSize: 13,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: 'pointer',
      }}
    >
      <span aria-hidden="true">⋯</span>
    </button>
  );
}

function DetailsPanel({ variant = 'pane' }: DetailsPanelProps) {
  const { state } = useAppContext();

  const catalog = state.catalogs.find((c) => c.path === state.currentCatalogId);
  const selectedNode =
    state.selected !== null && state.tree.status === 'ready'
      ? state.tree.nodes.find((n) => n.path === state.selected)
      : undefined;

  // No catalog selected -- Phase 22's placeholder, unchanged and still
  // reachable (e.g. right after the directory empties).
  if (!catalog) {
    return (
      <div className={`ws-details ws-details--${variant}`} style={{ padding: 14, gap: 16 }}>
        <span
          style={{
            fontSize: 12,
            fontWeight: 600,
            letterSpacing: '0.04em',
            textTransform: 'uppercase',
            color: 'var(--dm)',
            flex: 'none',
          }}
        >
          Details
        </span>
        <div
          className="pane-scroll"
          style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          <span style={{ fontSize: 12.5, color: 'var(--dm)', textAlign: 'center', lineHeight: 1.5 }}>
            Nothing selected. Pick a catalog in the rail, or catalog a volume to get started.
          </span>
        </div>
      </div>
    );
  }

  // Node-level view: a file or folder is selected in the tree. Only
  // reachable while the tree is ready (TreePane only ever dispatches
  // SET_SELECTED from a rendered row), so selectedNode is always found here.
  if (selectedNode) {
    const isDirectory = selectedNode.type === 'directory';
    const metaRows: MetaRow[] = [
      ['Type', isDirectory ? 'Folder' : 'File'],
      ['Size', formatBytes(selectedNode.size)],
      ['Catalog', catalog.title],
      ['Depth', String(selectedNode.depth)],
      ['Indexed', formatDate(catalog.modified)],
    ];

    return (
      <div
        className={`ws-details ws-details--${variant}`}
        style={{ padding: 14, gap: 16, display: 'flex', flexDirection: 'column' }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span
            style={{ fontSize: 12, fontWeight: 600, letterSpacing: '0.04em', textTransform: 'uppercase', color: 'var(--dm)' }}
          >
            {isDirectory ? 'Selected folder' : 'Selected file'}
          </span>
          <OverflowButton />
        </div>
        <div className="pane-scroll" style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div className="mono" style={{ fontSize: 12.5, lineHeight: 1.4, wordBreak: 'break-all' }}>
              {selectedNode.name}
            </div>
            <div
              className="mono"
              style={{
                fontSize: 11,
                color: 'var(--dm)',
                lineHeight: 1.5,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {selectedNode.path}
            </div>
          </div>
          <MetaRows rows={metaRows} />
        </div>
      </div>
    );
  }

  // Catalog-level view: a catalog is selected, no node chosen -- the
  // default after selecting a rail row, and also what renders for a
  // catalog whose JSON failed to parse (the fields below all come from the
  // rail listing, independent of whether LoadCatalogFlat ever ran).
  //
  // Files/Catalogued prefer the loaded FlatCatalog's exact counts (never
  // absent once the tree for THIS catalog has finished loading) over the
  // rail's cache-backed, possibly-cold/possibly-null CatalogMetadata
  // fields -- same precedent TreeHeader.tsx already established.
  const loadedTree =
    state.tree.status === 'ready' && state.currentCatalogId === catalog.path ? state.tree : null;
  const fileCount = loadedTree ? loadedTree.fileCount : (catalog.fileCount ?? null);
  const totalBytes = loadedTree ? loadedTree.totalBytes : (catalog.totalBytes ?? null);
  const metaRows: MetaRow[] = [
    ['Files', fileCount != null ? formatCount(fileCount) : '—'],
    ['Catalogued', totalBytes != null ? formatBytes(totalBytes) : '—'],
    ['JSON', formatBytes(catalog.size)],
    ['Modified', formatDate(catalog.modified)],
  ];

  return (
    <div
      className={`ws-details ws-details--${variant}`}
      style={{ padding: 14, gap: 16, display: 'flex', flexDirection: 'column' }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <span style={{ fontSize: 12, fontWeight: 600, letterSpacing: '0.04em', textTransform: 'uppercase', color: 'var(--dm)' }}>
          Catalog
        </span>
        <OverflowButton />
      </div>
      <div className="pane-scroll" style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div className="mono" style={{ fontSize: 12.5, lineHeight: 1.4, wordBreak: 'break-all' }}>
            {catalog.title}
          </div>
          <div
            className="mono"
            style={{
              fontSize: 11,
              color: 'var(--dm)',
              lineHeight: 1.5,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {catalog.path}
          </div>
        </div>
        {/* Deliberately no fifth "HTML" row -- the tree header's .json/.html
            chips already communicate that fact; adding it here would need a
            new existence-check capability outside this phase's locked three
            backend surfaces. */}
        <MetaRows rows={metaRows} />
      </div>
    </div>
  );
}

export default DetailsPanel;
