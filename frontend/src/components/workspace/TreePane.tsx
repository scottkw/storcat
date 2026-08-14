import { useEffect, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { useVisibleRows } from '../../hooks/useVisibleRows';
import { formatBytes } from '../../lib/format';
import BreadcrumbBar from './BreadcrumbBar';
import { models } from '../../../wailsjs/go/models';

const EMPTY_NODES: models.FlatNode[] = [];

function TreePane() {
  const { state, dispatch } = useAppContext();
  const scrollRef = useRef<HTMLDivElement>(null);

  const nodes = state.tree.status === 'ready' ? state.tree.nodes : EMPTY_NODES;
  const visibleIndices = useVisibleRows(nodes, state.expanded);

  // Row height comes from reducer state (the density the user has already
  // chosen), never a DOM measurement -- this is what lets the virtualizer's
  // estimateSize change on the same render a density toggle fires.
  const rowHeight = state.density === 'Compact' ? 27 : 34;

  const virtualizer = useVirtualizer({
    count: visibleIndices.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => rowHeight,
    overscan: 10,
  });

  // One effect keyed on currentCatalogId: resets scroll to the top (TREE-06
  // -- this component is never unmounted on a catalog switch, so relying on
  // a remount would not fire) and loads that catalog's flat tree in a
  // single call, dispatching TREE_LOADED/TREE_FAILED carrying the id it was
  // issued for so a load superseded by a newer selection is discarded.
  useEffect(() => {
    const catalogId = state.currentCatalogId;
    if (scrollRef.current) scrollRef.current.scrollTop = 0;
    virtualizer.scrollToOffset(0);
    if (!catalogId) return;
    wailsAPI.loadCatalogFlat(catalogId).then((result) => {
      if (result.success) {
        dispatch({
          type: 'TREE_LOADED',
          payload: {
            catalogId,
            nodes: result.flat.nodes,
            fileCount: result.flat.fileCount,
            totalBytes: result.flat.totalBytes,
          },
        });
      } else {
        dispatch({ type: 'TREE_FAILED', payload: { catalogId, message: result.error } });
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state.currentCatalogId]);

  // Directory click toggles expansion AND selects, in that order; a file
  // click selects only -- the two are never unified into one handler that
  // always toggles (TREE-02).
  const handleRowClick = (node: models.FlatNode) => {
    if (node.type === 'directory') {
      dispatch({ type: 'TOGGLE_EXPAND', payload: node.path });
    }
    dispatch({ type: 'SET_SELECTED', payload: node.path });
  };

  const selectedCatalog = state.catalogs.find((catalog) => catalog.path === state.currentCatalogId);

  // Four mutually exclusive states, never more than one rendered at once:
  // empty library (no directory configured, or the directory holds no
  // catalogs, or nothing has been selected yet), a quiet loading line while
  // LoadCatalogFlat is in flight, a distinct empty-catalog message when the
  // loaded array has zero nodes, and rows.
  if (!state.catalogDir || state.catalogs.length === 0 || !state.currentCatalogId) {
    return (
      <div className="ws-tree">
        <div className="pane-scroll" style={{ display: 'flex', flexDirection: 'column' }}>
          <div
            style={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 16,
              padding: 40,
            }}
          >
            <span
              aria-hidden="true"
              style={{
                width: 46,
                height: 46,
                borderRadius: 9,
                border: '1px dashed var(--l)',
                display: 'block',
              }}
            />
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: 7,
                maxWidth: 420,
              }}
            >
              <span style={{ fontSize: 16, fontWeight: 600 }}>Nothing catalogued yet</span>
              <span
                style={{
                  fontSize: 12.5,
                  lineHeight: 1.6,
                  color: 'var(--dm)',
                  textAlign: 'center',
                }}
              >
                Insert a card or plug in a drive and StorCat will offer it. Every catalog is a plain .json plus a browsable .html — nothing is stored in a database you can lose.
              </span>
            </div>
            <div style={{ display: 'flex', gap: 10 }}>
              <button
                type="button"
                style={{
                  height: 32,
                  padding: '0 15px',
                  borderRadius: 8,
                  background: 'var(--ac)',
                  color: 'var(--onac)',
                  border: 'none',
                  fontSize: 12.5,
                  fontWeight: 600,
                  cursor: 'pointer',
                }}
              >
                Catalog a volume
              </button>
              <button
                type="button"
                style={{
                  height: 32,
                  padding: '0 15px',
                  borderRadius: 8,
                  background: 'transparent',
                  color: 'var(--tx)',
                  border: '1px solid var(--l)',
                  fontSize: 12.5,
                  cursor: 'pointer',
                }}
              >
                Choose catalog folder…
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (state.tree.status === 'loading') {
    return (
      <div className="ws-tree">
        <div
          className="pane-scroll"
          style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          <span className="mono" style={{ fontSize: 11.5, color: 'var(--dm)' }}>
            Reading {selectedCatalog?.filename ?? 'catalog'}…
          </span>
        </div>
      </div>
    );
  }

  if (state.tree.status === 'ready' && nodes.length === 0) {
    return (
      <div className="ws-tree">
        <div
          className="pane-scroll"
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 7,
            padding: 40,
          }}
        >
          <span style={{ fontSize: 12.5, fontWeight: 500 }}>This catalog is empty</span>
          <span style={{ fontSize: 11.5, lineHeight: 1.5, color: 'var(--dm)', textAlign: 'center', maxWidth: 320 }}>
            This catalog was created with no files or folders inside it.
          </span>
        </div>
      </div>
    );
  }

  return (
    <div className="ws-tree">
      <BreadcrumbBar />
      <div className="pane-scroll" ref={scrollRef} role="tree">
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
          {virtualizer.getVirtualItems().map((virtualRow) => {
            const nodeIdx = visibleIndices[virtualRow.index];
            const node = nodes[nodeIdx];
            const isDirectory = node.type === 'directory';
            const isSelected = state.selected === node.path;
            const isExpanded = state.expanded[node.path] === true;
            const showCaret = isDirectory && node.hasChildren;
            const caret = showCaret ? (isExpanded ? '▾' : '▸') : '';

            return (
              <div
                key={node.path}
                className="ws-tree-row"
                role="treeitem"
                aria-level={node.depth + 1}
                aria-selected={isSelected}
                {...(showCaret ? { 'aria-expanded': isExpanded } : {})}
                data-selected={isSelected || undefined}
                onClick={() => handleRowClick(node)}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualRow.start}px)`,
                  paddingLeft: 18 + node.depth * 16,
                }}
              >
                <span className="ws-tree-caret mono" aria-hidden="true">
                  {caret}
                </span>
                <span className="ws-tree-shape" aria-hidden="true" data-kind={isDirectory ? 'directory' : 'file'} />
                <span className="ws-tree-name mono">{node.name}</span>
                <span className="ws-tree-size mono">{formatBytes(node.size)}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

export default TreePane;
