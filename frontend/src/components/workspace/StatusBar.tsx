function StatusBar() {
  return (
    <div className="ws-status mono">
      <span style={{ flexShrink: 0 }}>0 catalogs</span>
      <span style={{ flexShrink: 0 }}>0 files indexed</span>
      <span style={{ flexShrink: 0 }}>0.0 GB</span>
    </div>
  );
}

export default StatusBar;
