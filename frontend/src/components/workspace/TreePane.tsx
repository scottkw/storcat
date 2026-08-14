import { useEffect, useLayoutEffect, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { useVisibleRows } from '../../hooks/useVisibleRows';
import { formatBytes } from '../../lib/format';
import { findNodeIndexByPath, ancestorPathsOf } from '../../lib/reveal';
import BreadcrumbBar from './BreadcrumbBar';
import TreeHeader from './TreeHeader';
import UnreadableCatalogPanel from './UnreadableCatalogPanel';
import { models } from '../../../wailsjs/go/models';

const EMPTY_NODES: models.FlatNode[] = [];

function TreePane() {
  const { state, dispatch } = useAppContext();
  const scrollRef = useRef<HTMLDivElement>(null);
  // Set by the reveal's merge/select effect (A) once a target has been
  // located and expanded; consumed by the reveal's scroll effect (B), which
  // fires only after visibleIndices has recomputed to include the target.
  const revealScrollPathRef = useRef<string | null>(null);

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

  // Scroll reset has to wait for the 'ready' branch to actually mount the
  // scroll element. Resetting alongside the load dispatch above looks right
  // but is a silent no-op: while status is 'loading' the rendered branch
  // carries no ref, so scrollRef.current is null and the virtualizer has no
  // scroll element to move -- it then reapplies its own stale offset when the
  // real element appears. Keying this on the catalog id AND the ready
  // transition is what makes a revisit land at the top (TREE-06).
  useLayoutEffect(() => {
    if (state.tree.status !== 'ready') return;
    if (scrollRef.current) scrollRef.current.scrollTop = 0;
    virtualizer.scrollToOffset(0);
    // A catalog switch also drops any scroll request left over from a
    // reveal that never landed (e.g. the palette revealed a target, then
    // the user picked a different catalog from the rail before effect B's
    // scroll fired). Layout effects run before passive effects in the same
    // commit, so this can never clobber a request effect A sets in that
    // same commit.
    revealScrollPathRef.current = null;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state.currentCatalogId, state.tree.status]);

  // Reveal effect A: merge the target's ancestors into `expanded` and
  // select the target. Keyed on [pendingReveal, tree.status] so it carries
  // the request across the asynchronous gap while the catalog loads --
  // it re-runs once tree.status flips to 'ready' even if pendingReveal
  // itself hasn't changed since the palette set it.
  useEffect(() => {
    if (!state.pendingReveal || state.tree.status !== 'ready') return;

    const targetIdx = findNodeIndexByPath(nodes, state.pendingReveal);
    if (targetIdx === -1) {
      // The catalog changed on disk between the search returning and the
      // row being activated -- discard rather than retry; the palette has
      // already closed and there is nothing left to point at.
      dispatch({ type: 'SET_PENDING_REVEAL', payload: null });
      return;
    }

    const ancestors = ancestorPathsOf(nodes, targetIdx);
    // MERGE_EXPANDED is structurally merge-only (WR-01) -- the reducer
    // itself bails out to the same state object when every ancestor is
    // already expanded, so a repeat reveal of an already-visible node still
    // produces no expansion change and no open/close flicker.
    dispatch({ type: 'MERGE_EXPANDED', payload: ancestors });

    dispatch({ type: 'SET_SELECTED', payload: state.pendingReveal });
    revealScrollPathRef.current = state.pendingReveal;
    dispatch({ type: 'SET_PENDING_REVEAL', payload: null });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state.pendingReveal, state.tree.status]);

  // Reveal effect B: scroll to the target's post-expansion visible
  // position. Deliberately a SEPARATE effect from A, keyed on
  // [visibleIndices] rather than fired inline in A. The `virtualizer`
  // object in any given render is bound to that render's `count`, derived
  // from `visibleIndices`, derived from the `expanded` map A has only just
  // dispatched -- scrolling inside A would target a row list that does not
  // yet contain the target. This effect waits for visibleIndices to
  // recompute after the expansion commits.
  useEffect(() => {
    const path = revealScrollPathRef.current;
    if (!path) return;

    const visibleIdx = visibleIndices.findIndex((nodeIdx) => nodes[nodeIdx].path === path);
    if (visibleIdx === -1) {
      // The expansion dispatched by effect A hasn't committed into
      // visibleIndices yet -- wait for the next recomputation rather than
      // clearing the ref, so the next [visibleIndices] change tries again.
      return;
    }

    revealScrollPathRef.current = null;
    virtualizer.scrollToIndex(visibleIdx, { align: 'center' });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visibleIndices]);

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

  // A catalog can arrive broken by either route: the rail listing already
  // flagged it (parseError, known the instant it's selected -- no need to
  // wait for the load) or the load itself failed after selection (tree
  // status 'error'). Either one routes to the diagnostic panel, which
  // replaces the entire pane -- no header, no breadcrumb, no rows.
  const isUnreadable =
    (selectedCatalog?.parseError ?? '') !== '' || state.tree.status === 'error';

  // Five mutually exclusive states, never more than one rendered at once:
  // empty library (no directory configured, or the directory holds no
  // catalogs, or nothing has been selected yet), the unreadable-catalog
  // panel, a quiet loading line while LoadCatalogFlat is in flight, a
  // distinct empty-catalog message when the loaded array has zero nodes,
  // and rows.
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

  if (isUnreadable) {
    return (
      <div className="ws-tree">
        <UnreadableCatalogPanel />
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
        <TreeHeader />
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
      <TreeHeader />
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
