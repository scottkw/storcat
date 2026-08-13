export interface DetailsPanelProps {
  variant?: 'pane' | 'drawer';
}

function DetailsPanel({ variant = 'pane' }: DetailsPanelProps) {
  return (
    <div
      className={`ws-details ws-details--${variant}`}
      style={{ padding: 14, gap: 16 }}
    >
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

export default DetailsPanel;
