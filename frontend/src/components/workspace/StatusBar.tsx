import { useMemo } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { formatCount, formatGB } from '../../lib/format';
import { scanPercent } from '../../lib/scanFormat';

function StatusBar() {
  const { state, dispatch } = useAppContext();

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

  // The status bar's first-ever right-aligned segment (25-UI-SPEC.md's
  // Background Handoff Contract) -- present only while a scan is actually
  // running, in either sub-state. With no scan running the status bar
  // renders exactly as it did before this phase: no fourth segment, not an
  // empty reserved slot. AppContext's scan state is kept live even while
  // the create panel is closed (CreateSlideOver's scan:progress
  // subscription is never gated on the panel's open state), so this
  // segment's percentage always agrees with the panel's own.
  const scan = state.scan;
  const showScanSegment = scan.status === 'counting' || scan.status === 'scanning';
  const scanPct = scan.status === 'scanning' ? scanPercent(scan.bytesSeen, scan.totalBytes) : null;

  // WATCH-01: both inputs are already resolved synchronously in AppContext
  // by the time this renders, so there is no loading branch. Omitted
  // entirely (no placeholder, no dash) when either input is false/unset --
  // the same rule the scan segment already follows.
  const showWatchSegment = state.settings.watchDirectory && !!state.catalogDir;

  return (
    <div className="ws-status mono">
      {/* Grouped so the three original segments stay left-together under
          the container's new space-between distribution -- a single group
          plus the conditional scan segment is what makes the scan segment
          the only one that ever moves to the right edge. */}
      <div className="ws-status-left">
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
      <div className="ws-status-right">
        {/* Watching sits closer to centre (ambient, persistent state); scan
            keeps the outermost, most-attention-grabbing position it already
            occupies (transient, urgent state). Deliberately no live-region
            announcement and no hover tooltip on this segment -- the scan
            segment has neither, and this app has never used the
            hover-tooltip affordance anywhere. */}
        {showWatchSegment && (
          <span className="ws-status-watching">
            <span aria-hidden="true">●</span>
            <span>watching</span>
            <span className="ws-status-watching-dir">{state.catalogDir}</span>
          </span>
        )}
        {showScanSegment && (
          <button
            type="button"
            className="ws-status-scan"
            onClick={() => dispatch({ type: 'SET_CREATE_OPEN', payload: true })}
          >
            <span aria-hidden="true">●</span>
            <span className="ws-status-scan-name">{scan.title}</span>
            <span style={{ flexShrink: 0 }}>· {scanPct !== null ? `${scanPct}%` : 'counting…'}</span>
          </button>
        )}
      </div>
    </div>
  );
}

export default StatusBar;
