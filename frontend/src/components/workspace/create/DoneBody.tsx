import { formatBytes } from '../../../lib/format';
import type { ScanState } from '../../../types/scan';

type DoneScanState = Extract<ScanState, { status: 'done' }>;

export interface DoneBodyProps {
  scan: DoneScanState;
  onOpenInWorkspace: () => void;
  onCatalogAnother: () => void;
}

/**
 * Renders a duration as this app's only duration text -- under a minute as
 * plain seconds, at or above a minute as "{m}m {s}s". Local to this
 * component: nothing else in the app currently formats a duration, so a
 * shared helper would be an unrequested abstraction ahead of a second
 * caller.
 */
function formatDuration(ms: number): string {
  const totalSeconds = Math.max(0, Math.round(ms / 1000));
  if (totalSeconds < 60) return `${totalSeconds}s`;
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}m ${seconds}s`;
}

/**
 * CRT-12's done state (25-UI-SPEC.md's Done State Contract, E7): one
 * layout for both flavours. The badge never changes color between them --
 * something was successfully written either way -- and the partial flavour
 * adds exactly one tag plus one swapped doneLine clause, never a second
 * template.
 */
function DoneBody({ scan, onOpenInWorkspace, onCatalogAnother }: DoneBodyProps) {
  // Terminal-state guard (defense in depth): the reducer's own SCAN_PROGRESS
  // clamp (AppContext.tsx) already refuses to move a 'done' state back into
  // a scanning one; this re-asserts the same invariant at the component
  // boundary per CRT-12, rather than trusting a single layer.
  console.assert(scan.status === 'done', 'DoneBody rendered with a non-terminal scan state');

  const doneLine = scan.partial
    ? `${scan.fileCount} files · ${formatBytes(scan.totalSize)} · ${
        scan.stopPercent !== null && scan.stopPercent !== undefined
          ? `stopped at ${scan.stopPercent}%`
          : 'stopped early'
      }`
    : `${scan.fileCount} files · ${formatBytes(scan.totalSize)} · ${formatDuration(scan.durationMs)}`;

  return (
    <div className="ws-create-state-body">
      <div className="ws-create-done-row">
        {/* Unchanged between flavours -- still a success color, since
            something was successfully written either way. */}
        <span className="ws-create-badge ws-create-badge-done" aria-hidden="true">
          ✓
        </span>
        <div className="ws-create-done-col">
          <div className="ws-create-done-title-row">
            <span className="ws-create-done-title">{scan.title} catalogued</span>
            {/* The entire visual difference between flavours: this tag plus
                the swapped doneLine clause above, reusing the exact
                warning-tag styling already declared for the volume card's
                "read errors" status -- no new color. This distinction is
                the whole point of CRT-11 offering the partial write: a
                partial catalog must never be mistaken for a complete one. */}
            {scan.partial && <span className="ws-create-tag ws-create-tag-partial">partial</span>}
          </div>
          <div className="ws-create-done-line mono">{doneLine}</div>
        </div>
      </div>

      {/* One row per file CreateCatalogResult actually reports as written
          (CreateSlideOver's filesFromResult), keyed on path -- never on
          size, so two files of identical size render as two distinct rows
          in a stable order, never merged or deduplicated. Rendered in the
          order the result provides, never sorted. A source with zero files
          still reaches here with at least the JSON row present, so there
          is no empty variant of this list. */}
      <div className="ws-create-file-list">
        {scan.files.map((file) => (
          <div className="ws-create-file-row" key={file.path}>
            <span className="ws-create-file-shape" aria-hidden="true" />
            <span className="ws-create-file-path mono">{file.path}</span>
            {typeof file.size === 'number' && (
              <span className="ws-create-file-size mono">{formatBytes(file.size)}</span>
            )}
          </div>
        ))}
      </div>

      <div className="ws-create-actions">
        {/* Primary, and the fifth CRT-01 close path: re-fetches the
            configured catalog directory's listing (CreateSlideOver's
            handleOpenInWorkspace) rather than optimistically prepending --
            per 25-RESEARCH.md's own recommendation, since catalog counts
            are small and a re-fetch reuses the existing loader instead of
            adding optimistic-update logic that would need its own
            staleness rules. Selects the new catalog, then runs the same
            close handler every other close path uses. */}
        <button type="button" className="ws-create-btn ws-create-btn-primary" onClick={onOpenInWorkspace}>
          Open in workspace
        </button>
        {/* Outlined secondary -- does NOT close the panel: resets the form
            and returns to the form step in the same still-open panel. */}
        <button type="button" className="ws-create-btn-outline" onClick={onCatalogAnother}>
          Catalog another volume
        </button>
      </div>
    </div>
  );
}

export default DoneBody;
