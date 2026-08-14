import { useMemo } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { formatCount, formatGB } from '../../lib/format';

function StatusBar() {
  const { state } = useAppContext();

  // Derived from the rail array in one memo, recomputed from the current
  // state.catalogs reference on every render -- never a running total, so a
  // listing refresh mid-render cannot produce a torn number (SHELL-06).
  // Both sums skip catalogs whose counts are absent (a cold cache entry)
  // rather than treating them as zero, and accumulate raw integer bytes --
  // formatGB does the single division at the end.
  const { catalogCount, filesIndexed, totalBytes } = useMemo(() => {
    let files = 0;
    let bytes = 0;
    for (const catalog of state.catalogs) {
      if (typeof catalog.fileCount === 'number') files += catalog.fileCount;
      if (typeof catalog.totalBytes === 'number') bytes += catalog.totalBytes;
    }
    return { catalogCount: state.catalogs.length, filesIndexed: files, totalBytes: bytes };
  }, [state.catalogs]);

  return (
    <div className="ws-status mono">
      <span style={{ flexShrink: 0 }}>{catalogCount} catalogs</span>
      <span style={{ flexShrink: 0 }}>{formatCount(filesIndexed)} files indexed</span>
      <span style={{ flexShrink: 0 }}>{formatGB(totalBytes)}</span>
    </div>
  );
}

export default StatusBar;
