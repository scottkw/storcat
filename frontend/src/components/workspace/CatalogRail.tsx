import { useCallback, useDeferredValue, useEffect, useMemo, useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { formatBytes, formatCount } from '../../lib/format';
import { safeGetItem, safeSetItem } from '../../themeTokens';

const CATALOG_DIR_STORAGE_KEY = 'storcat-catalog-directory';

function CatalogRail() {
  const { state, dispatch } = useAppContext();

  // The filter string lives only here, never in AppContext, read through
  // useDeferredValue for the matching pass. This isolation -- not a
  // debounce timer -- is the entire mechanism by which the tree survives
  // typing untouched (RAIL-02, locked in 23-CONTEXT.md).
  const [filterInput, setFilterInput] = useState('');
  const deferredFilter = useDeferredValue(filterInput);
  const trimmedFilter = deferredFilter.trim();

  const loadCatalogsForDirectory = useCallback(
    (dir: string) => {
      wailsAPI.browseCatalogs(dir).then((result) => {
        if (result.success) {
          dispatch({ type: 'SET_CATALOGS', payload: result.catalogs ?? [] });
        }
        // A failed listing (missing/unreadable directory) leaves state.catalogs
        // untouched -- combined with the empty initial array this renders the
        // same empty-library state as zero catalogs, with no console error.
      });
    },
    [dispatch]
  );

  // On mount, read the persisted catalog directory (if any) and load its
  // catalogs once. No persisted directory -> do not call the listing binding
  // at all; the empty-library state below covers that case (STATE-01).
  useEffect(() => {
    const persistedDir = safeGetItem(CATALOG_DIR_STORAGE_KEY);
    if (!persistedDir) return;
    dispatch({ type: 'SET_CATALOG_DIR', payload: persistedDir });
    loadCatalogsForDirectory(persistedDir);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleChooseDirectory = async () => {
    const result = await wailsAPI.selectDirectory();
    if (!result.success || !result.path) return;
    safeSetItem(CATALOG_DIR_STORAGE_KEY, result.path);
    dispatch({ type: 'SET_CATALOG_DIR', payload: result.path });
    loadCatalogsForDirectory(result.path);
  };

  // Case-insensitive substring match against title + filename together,
  // preserving the array's existing order -- no re-sort, no ranking.
  const filteredCatalogs = useMemo(() => {
    if (!trimmedFilter) return state.catalogs;
    const needle = trimmedFilter.toLowerCase();
    return state.catalogs.filter(
      (catalog) =>
        catalog.title.toLowerCase().includes(needle) || catalog.filename.toLowerCase().includes(needle)
    );
  }, [state.catalogs, trimmedFilter]);

  const showZeroMatch = trimmedFilter !== '' && filteredCatalogs.length === 0 && state.catalogs.length > 0;

  return (
    <div className="ws-rail">
      <div
        style={{
          padding: '12px 12px 10px',
          display: 'flex',
          flexDirection: 'column',
          gap: 10,
          borderBottom: '1px solid var(--l)',
          flex: 'none',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span
            style={{
              fontSize: 12,
              fontWeight: 600,
              letterSpacing: '0.04em',
              textTransform: 'uppercase',
              color: 'var(--dm)',
            }}
          >
            {/* Always the total, never the filtered subset (RAIL-01/RAIL-02) --
                reads state.catalogs.length directly, independent of trimmedFilter. */}
            Catalogs <span style={{ color: 'var(--fn)' }}>{state.catalogs.length}</span>
          </span>

          {/* Renders, hover-styled, and stays inert -- its target (the create
              slide-over) is Phase 25. Never attach a handler here (RAIL-06). */}
          <button
            type="button"
            className="ws-new-pill"
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 5,
              fontSize: 12,
              fontWeight: 600,
              border: 'none',
              borderRadius: 6,
              padding: '3px 8px',
              cursor: 'pointer',
            }}
          >
            <svg
              width="10"
              height="10"
              viewBox="0 0 12 12"
              stroke="currentColor"
              strokeWidth={1.8}
              aria-hidden="true"
              focusable="false"
            >
              <line x1="6" y1="1.5" x2="6" y2="10.5" />
              <line x1="1.5" y1="6" x2="10.5" y2="6" />
            </svg>
            New
          </button>
        </div>

        <button
          type="button"
          className="ws-dir-chip mono"
          onClick={handleChooseDirectory}
          aria-label="Choose catalog directory"
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 7,
            width: '100%',
            fontSize: 11,
            color: 'var(--dm)',
            background: 'var(--ch)',
            border: '1px solid var(--l)',
            borderRadius: 6,
            padding: '5px 8px',
            cursor: 'pointer',
            textAlign: 'left',
            font: 'inherit',
          }}
        >
          <svg
            width="11"
            height="11"
            viewBox="0 0 14 14"
            fill="none"
            stroke="var(--fn)"
            strokeWidth={1.4}
            aria-hidden="true"
            focusable="false"
          >
            <rect x="1" y="3" width="12" height="9" rx="1.5" />
            <line x1="1" y1="5.5" x2="13" y2="5.5" />
          </svg>
          <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {state.catalogDir ?? 'No catalog directory set'}
          </span>
        </button>

        <div
          style={{
            height: 26,
            borderRadius: 6,
            background: 'var(--bg)',
            border: '1px solid var(--l)',
            display: 'flex',
            alignItems: 'center',
            padding: '0 8px',
          }}
        >
          <input
            aria-label="Filter catalogs"
            placeholder="Filter catalogs…"
            value={filterInput}
            onChange={(event) => setFilterInput(event.target.value)}
            style={{
              width: '100%',
              fontSize: 12,
              color: 'var(--tx)',
              background: 'transparent',
              border: 'none',
              outline: 'none',
            }}
          />
        </div>
      </div>

      <div
        className="pane-scroll"
        style={{ padding: 6, display: 'flex', flexDirection: 'column', gap: 1 }}
      >
        {state.catalogs.length === 0 ? (
          <div style={{ padding: '16px 10px', display: 'flex', flexDirection: 'column', gap: 10 }}>
            <span style={{ fontSize: 12.5, fontWeight: 500 }}>No catalogs here yet</span>
            <span style={{ fontSize: 11.5, lineHeight: 1.5, color: 'var(--dm)' }}>
              This folder has no .json catalogs. Point StorCat somewhere else, or catalog a volume to create the first one.
            </span>
            <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--ac)' }}>
              Catalog a volume →
            </span>
          </div>
        ) : showZeroMatch ? (
          <div style={{ padding: '16px 10px', textAlign: 'center' }}>
            <span style={{ fontSize: 11.5, color: 'var(--dm)' }}>
              No catalogs match &quot;{deferredFilter}&quot;.
            </span>
          </div>
        ) : (
          filteredCatalogs.map((catalog) => {
            const isSelected = state.currentCatalogId === catalog.path;
            const isBroken = catalog.parseError !== '';

            return (
              <div
                key={catalog.path}
                className="ws-rail-row"
                role="button"
                tabIndex={0}
                // data-selected drives the CSS fill; aria-current is what
                // actually tells a screen reader which catalog is open. The
                // visual state without the ARIA one means the selection is
                // invisible to anyone not looking at the pixels.
                aria-current={isSelected || undefined}
                data-selected={isSelected || undefined}
                onClick={() => dispatch({ type: 'SELECT_CATALOG', payload: catalog.path })}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault();
                    dispatch({ type: 'SELECT_CATALOG', payload: catalog.path });
                  }
                }}
              >
                <div className="ws-rail-row-line1">
                  {/* Always in the DOM, transparent when healthy -- a broken
                      catalog never changes row height (RAIL-04). */}
                  <span
                    className="ws-rail-dot"
                    aria-hidden="true"
                    style={{ background: isBroken ? '#e5534b' : 'transparent' }}
                  />
                  <span className="ws-rail-row-title">{catalog.title}</span>
                  <span className="ws-rail-row-size mono">{formatBytes(catalog.size)}</span>
                </div>
                <div className="ws-rail-row-meta mono">
                  {catalog.filename}
                  {typeof catalog.fileCount === 'number'
                    ? ` · ${formatCount(catalog.fileCount)} files`
                    : ''}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

export default CatalogRail;
