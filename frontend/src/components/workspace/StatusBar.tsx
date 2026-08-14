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
  //
  // partial tracks whether any catalog's count was absent -- there is no
  // background/eager cache fill (WR-03 deliberately doesn't add one), so a
  // sum over a directory with cold entries is a real number, just not a
  // complete one. Rendering it unqualified would present "0 files indexed"
  // as authoritative when it actually means "not computed yet".
  const { catalogCount, filesIndexed, totalBytes, partial } = useMemo(() => {
    let files = 0;
    let bytes = 0;
    let partial = false;
    for (const catalog of state.catalogs) {
      if (typeof catalog.fileCount === 'number') {
        files += catalog.fileCount;
      } else {
        partial = true;
      }
      if (typeof catalog.totalBytes === 'number') {
        bytes += catalog.totalBytes;
      } else {
        partial = true;
      }
    }
    return { catalogCount: state.catalogs.length, filesIndexed: files, totalBytes: bytes, partial };
  }, [state.catalogs]);

  const qualifier = partial ? '≥' : '';

  return (
    <div className="ws-status mono">
      <span style={{ flexShrink: 0 }}>{catalogCount} catalogs</span>
      <span style={{ flexShrink: 0 }}>
        {qualifier}
        {formatCount(filesIndexed)} files indexed
      </span>
      <span style={{ flexShrink: 0 }}>
        {qualifier}
        {formatGB(totalBytes)}
      </span>
    </div>
  );
}

export default StatusBar;
