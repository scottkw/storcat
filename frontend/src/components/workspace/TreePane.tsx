function TreePane() {
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

export default TreePane;
