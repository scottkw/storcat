import { useEffect, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { useVisibleRows } from '../../hooks/useVisibleRows';
import { models } from '../../../wailsjs/go/models';

const EMPTY_NODES: models.FlatNode[] = [];

function TreePane() {
  const { state, dispatch } = useAppContext();
  const scrollRef = useRef<HTMLDivElement>(null);

  // One effect keyed on currentCatalogId: it loads that catalog's flat tree
  // in a single call and dispatches TREE_LOADED/TREE_FAILED carrying the id
  // it was issued for, so a superseded load (the user picked another
  // catalog before this one resolved) loses to the newer selection instead
  // of repainting the tree with the wrong catalog's rows.
  useEffect(() => {
    const catalogId = state.currentCatalogId;
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

  if (!state.currentCatalogId) {
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

  // Tracer-minimal row: name only. Carets, shapes, sizes, selection
  // styling, expansion and the catalog header/breadcrumb are plan 23-03's
  // work -- this task proves the load-to-render path end to end.
  return (
    <div className="ws-tree">
      <div className="pane-scroll" ref={scrollRef}>
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
          {virtualizer.getVirtualItems().map((virtualRow) => {
            const nodeIdx = visibleIndices[virtualRow.index];
            const node = nodes[nodeIdx];
            return (
              <div
                key={nodeIdx}
                className="ws-tree-row"
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualRow.start}px)`,
                  display: 'flex',
                  alignItems: 'center',
                  paddingLeft: 18 + node.depth * 16,
                  paddingRight: 18,
                }}
              >
                {node.name}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

export default TreePane;
