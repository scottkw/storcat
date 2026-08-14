import { useEffect } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';

const CATALOG_DIR_STORAGE_KEY = 'storcat-catalog-directory';

function CatalogRail() {
  const { state, dispatch } = useAppContext();

  // Tracer-minimal: on mount, read the persisted catalog directory and load
  // its catalogs once. The full directory-chip wiring (picking/changing the
  // directory interactively) belongs to plan 23-04.
  useEffect(() => {
    const persistedDir = localStorage.getItem(CATALOG_DIR_STORAGE_KEY);
    if (!persistedDir) return;
    dispatch({ type: 'SET_CATALOG_DIR', payload: persistedDir });
    wailsAPI.browseCatalogs(persistedDir).then((result) => {
      if (result.success) {
        dispatch({ type: 'SET_CATALOGS', payload: result.catalogs ?? [] });
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
            Catalogs <span style={{ color: 'var(--fn)' }}>{state.catalogs.length}</span>
          </span>

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

        <div
          className="mono"
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 7,
            fontSize: 11,
            color: 'var(--dm)',
            background: 'var(--ch)',
            border: '1px solid var(--l)',
            borderRadius: 6,
            padding: '5px 8px',
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
        </div>

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
        ) : (
          state.catalogs.map((catalog) => (
            <div
              key={catalog.path}
              className="ws-rail-row"
              role="button"
              tabIndex={0}
              onClick={() => dispatch({ type: 'SELECT_CATALOG', payload: catalog.path })}
              onKeyDown={(event) => {
                if (event.key === 'Enter' || event.key === ' ') {
                  dispatch({ type: 'SELECT_CATALOG', payload: catalog.path });
                }
              }}
              style={{
                cursor: 'pointer',
                fontSize: 12.5,
                borderRadius: 6,
                background: state.currentCatalogId === catalog.path ? 'var(--sel)' : 'transparent',
              }}
            >
              {catalog.title}
            </div>
          ))
        )}
      </div>
    </div>
  );
}

export default CatalogRail;
