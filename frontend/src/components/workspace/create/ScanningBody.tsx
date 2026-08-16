import { formatBytes } from '../../../lib/format';
import { formatEta, scanPercent } from '../../../lib/scanFormat';
import type { ScanState } from '../../../types/scan';

type ScanningScanState = Extract<ScanState, { status: 'counting' } | { status: 'scanning' }>;

export interface ScanningBodyProps {
  scan: ScanningScanState;
  // Optional (28-01) -- when omitted, the "Run in background" button and
  // its helper text simply don't render. Re-scan has nothing to background
  // (28-UI-SPEC.md's Architecture & State: closing RescanDialog mid-scan
  // always cancels), so its call site passes neither. CreateSlideOver's own
  // call site is unchanged -- it still always passes this.
  onRunInBackground?: () => void;
}

// UI-SPEC E5 overflow: the log box retains at most 9 newest-first lines
// inside a fixed max-height, so it never scrolls. AppContext's reducer
// enforces the same cap on the retained state itself (T-25-23); this is a
// second, defensive cap at render time so the two can never drift.
const LOG_CAP = 9;

/**
 * Both sub-states of `scanning` (25-UI-SPEC's Step 2 Scanning Contract),
 * rendered from one component instance and one state object -- no
 * indeterminate loading indicator in either, per the project-wide
 * "progress is always a real number" rule. `pct === null` (the denominator
 * is genuinely unknown) is the sole signal for the counting sub-state;
 * there is no second flag.
 */
function ScanningBody({ scan, onRunInBackground }: ScanningBodyProps) {
  const pct = scan.status === 'scanning' ? scanPercent(scan.bytesSeen, scan.totalBytes) : null;
  const isCounting = pct === null;
  const log = scan.log.slice(0, LOG_CAP);

  return (
    <div className="ws-create-scan-body">
      <div className="ws-create-scan-title-row">
        <span className="ws-create-scan-title">{scan.title}</span>
        <span className={`ws-create-scan-pct mono${isCounting ? ' ws-create-scan-pct-counting' : ''}`}>
          {isCounting ? 'Counting…' : `${pct}%`}
        </span>
      </div>

      {!isCounting && (
        <div className="ws-create-progress">
          <div className="ws-create-progress-fill" style={{ width: `${pct}%` }} />
        </div>
      )}

      <div className="ws-create-scan-counters mono">
        {scan.status === 'scanning'
          ? `${scan.filesSeen} files · ${formatBytes(scan.bytesSeen)} · ${formatEta(
              scan.bytesSeen,
              scan.totalBytes,
              Date.now() - scan.startedAt
            )}`
          : `${scan.filesSeen} files found so far`}
      </div>

      <div className="ws-create-walking">
        <div className="ws-create-walking-label">Walking</div>
        <div className="ws-create-walking-path mono">{scan.currentPath}</div>
      </div>

      <div className="ws-create-log">
        {log.map((path, index) => (
          // Newest-first: index 0 is the most recent path, so a stable key
          // needs the path itself, not just the index (which would collide
          // as older lines shift down the array on every push).
          <div className="ws-create-log-line mono" key={`${index}-${path}`}>
            + {path}
          </div>
        ))}
      </div>

      {onRunInBackground && (
        <div className="ws-create-scan-footer">
          <button type="button" className="ws-create-btn-outline" onClick={onRunInBackground}>
            Run in background
          </button>
          <span className="ws-create-scan-footer-note">Progress also shows in the status bar.</span>
        </div>
      )}
    </div>
  );
}

export default ScanningBody;
