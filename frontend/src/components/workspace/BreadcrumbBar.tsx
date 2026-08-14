import { useAppContext } from '../../contexts/AppContext';
import { models } from '../../../wailsjs/go/models';

/**
 * Walks the flat array's parentIdx chain upward from the selected node and
 * reverses it, so the result reads root-to-leaf -- the shape the breadcrumb
 * renders left to right. Returns an empty chain when nothing is selected or
 * the selected path can no longer be found (e.g. a stale selection after a
 * catalog switch that has not yet cleared it).
 */
function buildAncestorChain(nodes: models.FlatNode[], selectedPath: string | null): models.FlatNode[] {
  if (!selectedPath) return [];
  const startIdx = nodes.findIndex((node) => node.path === selectedPath);
  if (startIdx === -1) return [];

  const chain: models.FlatNode[] = [];
  let idx = startIdx;
  while (idx !== -1) {
    chain.push(nodes[idx]);
    idx = nodes[idx].parentIdx;
  }
  return chain.reverse();
}

function BreadcrumbBar() {
  const { state, dispatch } = useAppContext();

  const catalog = state.catalogs.find((c) => c.path === state.currentCatalogId);
  const catalogSegment = catalog ? catalog.filename.replace(/\.json$/, '') : '';
  const nodes = state.tree.status === 'ready' ? state.tree.nodes : [];
  const ancestors = buildAncestorChain(nodes, state.selected);
  const segments = [catalogSegment, ...ancestors.map((node) => node.name)];
  const lastIndex = segments.length - 1;
  const loaded = state.tree.status === 'ready';

  // One O(n) map construction, one dispatch that replaces the whole
  // expansion map -- never a dispatch per node, which is the shape that
  // freezes at 40,000 nodes.
  const handleExpandAll = () => {
    if (state.tree.status !== 'ready') return;
    const next: Record<string, boolean> = {};
    for (const node of state.tree.nodes) {
      if (node.hasChildren) next[node.path] = true;
    }
    dispatch({ type: 'SET_EXPANDED', payload: next });
  };

  // An empty map returns the tree to top-level entries only -- matches the
  // prototype's collapseAll, which re-opens only the synthetic root.
  const handleCollapse = () => {
    dispatch({ type: 'SET_EXPANDED', payload: {} });
  };

  return (
    <div
      style={{
        height: 34,
        flex: 'none',
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '0 18px',
        background: 'var(--p2)',
        borderBottom: '1px solid var(--l)',
      }}
    >
      <span
        className="mono"
        style={{
          flex: '1 1 auto',
          minWidth: 0,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
          fontSize: 11.5,
        }}
      >
        {segments.map((segment, index) => (
          <span key={index}>
            {index > 0 && <span style={{ color: 'var(--l)' }}> / </span>}
            <span className={index === lastIndex ? 'ws-crumb-current' : 'ws-crumb'}>{segment}</span>
          </span>
        ))}
      </span>
      {loaded && (
        <button
          type="button"
          onClick={handleExpandAll}
          style={{
            flexShrink: 0,
            fontSize: 11.5,
            color: 'var(--ac)',
            cursor: 'pointer',
            background: 'transparent',
            border: 'none',
            padding: 0,
            font: 'inherit',
          }}
        >
          Expand all
        </button>
      )}
      {loaded && (
        <button
          type="button"
          onClick={handleCollapse}
          className="ws-crumb-collapse"
          style={{
            flexShrink: 0,
            fontSize: 11.5,
            cursor: 'pointer',
            background: 'transparent',
            border: 'none',
            padding: 0,
            font: 'inherit',
          }}
        >
          Collapse
        </button>
      )}
    </div>
  );
}

export default BreadcrumbBar;
