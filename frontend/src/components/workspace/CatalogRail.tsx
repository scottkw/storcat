function CatalogRail() {
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
            Catalogs <span style={{ color: 'var(--fn)' }}>0</span>
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
            No catalog directory set
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
        <div style={{ padding: '16px 10px', display: 'flex', flexDirection: 'column', gap: 10 }}>
          <span style={{ fontSize: 12.5, fontWeight: 500 }}>No catalogs here yet</span>
          <span style={{ fontSize: 11.5, lineHeight: 1.5, color: 'var(--dm)' }}>
            This folder has no .json catalogs. Point StorCat somewhere else, or catalog a volume to create the first one.
          </span>
          <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--ac)' }}>
            Catalog a volume →
          </span>
        </div>
      </div>
    </div>
  );
}

export default CatalogRail;
