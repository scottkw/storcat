import { useEffect, useRef, useState } from 'react';
import { useAppContext } from '../../../contexts/AppContext';
import { wailsAPI } from '../../../services/wailsAPI';
import { useModalBehavior } from '../../../hooks/useModalBehavior';
import { classifyScanFailure, sourceDisplayNameOf, sourcePathOf, type ScanSource } from '../../../types/scan';
import type { DiffResult, DiffState } from '../../../types/rescan';
import { models } from '../../../../wailsjs/go/models';
import { formatBytes } from '../../../lib/format';
import VolumePicker from '../create/VolumePicker';
import ScanningBody from '../create/ScanningBody';
import ErrorBody from '../create/ErrorBody';
import DiffList from './DiffList';

export interface RescanDialogProps {
  catalog: models.CatalogMetadata;
  catalogDir: string | null;
  oldTreeAvailable: boolean;
  onClose: () => void;
  onError: (message: string | null) => void;
}

const STAT_TILES: Array<{ key: Exclude<DiffState, 'unreadable'> | 'unreadable'; label: string; color: string }> = [
  { key: 'added', label: 'added', color: 'var(--ac)' },
  { key: 'removed', label: 'removed', color: 'var(--danger)' },
  { key: 'changed', label: 'changed', color: '#f0b429' },
  { key: 'unreadable', label: 'unreadable', color: 'var(--danger)' },
  { key: 'unchanged', label: 'unchanged', color: 'var(--dm)' },
];

// "N" in the Variant A header title -- never includes unchanged.
function diffTotal(diff: DiffResult): number {
  return diff.added + diff.removed + diff.changed + diff.unreadable;
}

// The similarity banner's "of {total}" denominator -- all five categories,
// the same distinct-path-count basis the sum invariant itself uses.
function diffGrandTotal(diff: DiffResult): number {
  return diffTotal(diff) + diff.unchanged;
}

// Variant B (oldTreeAvailable: false) has no old tree to diff against, so
// ComputeDiff reports every new-tree path as "added" (catalog.ComputeDiff's
// nil-old-tree behavior) -- these are real added entries, not a fabricated
// count, and this is the only place fileCount/totalBytes for Variant B's
// summary line can come from: RescanCatalog's binding returns only a
// DiffResult, no separate scan-result struct for this path.
function newTreeStatsFrom(diff: DiffResult): { fileCount: number; totalBytes: number } {
  let fileCount = 0;
  let totalBytes = 0;
  for (const entry of diff.entries) {
    if (entry.state === 'added' && entry.type === 'file') {
      fileCount += 1;
      totalBytes += entry.newSize ?? 0;
    }
  }
  return { fileCount, totalBytes };
}

/**
 * The tracer slice of 28-UI-SPEC.md's RescanDialog: pick a volume, watch
 * the shared live scan progress (state.scan, the same slice
 * CreateSlideOver drives), see the five diff counts. The diff row list,
 * similarity warning, error/interrupted step, and the two write
 * resolutions (Overwrite/Keep-both) are plan 28-02/03/04 -- step 3 here
 * carries only the stat grid and a single "Discard scan and close" action,
 * per this plan's explicit tracer scope.
 *
 * Conditionally mounted by its parent (Footer) -- unlike CreateSlideOver/
 * SettingsDialog's always-mounted pattern, this component is only rendered
 * while open, so a plain mount effect (not an isOpen-keyed re-seed) is
 * correct for registering its own state.rescan slice.
 */
function RescanDialog({ catalog, catalogDir, oldTreeAvailable, onClose, onError }: RescanDialogProps) {
  const { state, dispatch } = useAppContext();
  const [selectedSource, setSelectedSource] = useState<ScanSource | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const submittingRef = useRef(false);
  const [resolving, setResolving] = useState(false);
  const [resolveError, setResolveError] = useState<string | null>(null);

  useEffect(() => {
    dispatch({ type: 'RESCAN_OPENED', payload: { catalog, oldTreeAvailable } });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // The single close handler every close trigger (Escape, x, scrim,
  // "Cancel"/"Discard scan and close") routes through -- mirrors
  // CreateSlideOver's handleCloseRequest. All four close paths write
  // nothing and need no confirmation (28-UI-SPEC.md): nothing is written
  // anywhere in this tracer's RescanCatalog binding.
  async function handleCloseRequest() {
    const scanning = state.scan.status === 'counting' || state.scan.status === 'scanning';
    if (scanning) {
      await wailsAPI.cancelScan();
      dispatch({ type: 'SCAN_RESET' });
    }
    dispatch({ type: 'RESCAN_CLOSED' });
    onClose();
  }

  const { containerRef } = useModalBehavior({ isOpen: true, onClose: handleCloseRequest });

  async function handleStart() {
    if (submittingRef.current || !selectedSource) return;
    submittingRef.current = true;
    setSubmitting(true);
    onError(null);

    const sourcePath = sourcePathOf(selectedSource);
    dispatch({ type: 'RESCAN_STARTED', payload: { sourcePath } });
    // Drives the SAME state.scan slice Create uses, through the SAME
    // scan:progress event subscription CreateSlideOver already owns -- no
    // second listener anywhere in this component.
    dispatch({ type: 'SCAN_STARTED', payload: { title: catalog.title } });

    const outcome = await wailsAPI.rescanCatalog(catalog.path, sourcePath, oldTreeAvailable);

    submittingRef.current = false;
    setSubmitting(false);

    if (!outcome.success) {
      const failure = classifyScanFailure(outcome.error);
      if (failure.kind === 'sourceLoss') {
        // The picked volume vanished mid-walk -- transition the shared
        // scan slice into its own error member so this component's
        // 'scanning' step renders the interrupted-scan body instead of the
        // live-progress one. state.rescan.step stays 'scanning' (there is
        // no separate rescan-owned error step): the same shared-slice
        // architecture the happy path already uses.
        dispatch({ type: 'SCAN_FAILED', payload: { message: failure.message, sourcePath } });
        return;
      }
      // Any other rejection (e.g. a cancellation): back to step 1 with the
      // same catalog/oldTreeAvailable, surfaced through the same error slot
      // Footer's other actions already share.
      dispatch({ type: 'SCAN_RESET' });
      dispatch({ type: 'RESCAN_OPENED', payload: { catalog, oldTreeAvailable } });
      onError(failure.message);
      return;
    }

    // The generated models.DiffResult types `state`/`type` as bare `string`
    // (Wails' codegen doesn't preserve Go's named DiffState string-const
    // type across the bridge) -- structurally identical to this app's own
    // DiffResult/DiffEntry, so this is a nominal-typing cast, not a real
    // shape mismatch.
    dispatch({ type: 'RESCAN_DIFFED', payload: { diff: outcome.diff as unknown as DiffResult } });
  }

  // Overwrite/Keep-both -- the two write resolutions. Discard calls no
  // binding at all (handleCloseRequest covers it, nothing written).
  // catalogDir is required for the Go side's containment check, the same
  // guard every other catalog-mutating action in this app already applies
  // (Rename/Duplicate/Delete) -- fail closed here rather than sending an
  // empty directory the backend would just reject anyway.
  async function handleResolve(mode: 'overwrite' | 'keep-both') {
    if (resolving) return;
    if (!catalogDir) {
      setResolveError('No catalog directory configured.');
      return;
    }
    setResolving(true);
    setResolveError(null);

    const outcome = await wailsAPI.resolveRescan(catalog.path, catalogDir, mode);

    setResolving(false);
    if (!outcome.success) {
      const actionLabel = mode === 'overwrite' ? 'Overwrite' : 'Keep both';
      setResolveError(`${actionLabel} failed: ${outcome.error}`);
      return;
    }

    // Success path mirrors Duplicate's own: re-trigger the rail's one
    // authoritative listing so the resolved/new file appears without
    // requiring watching to be on (it defaults to off).
    dispatch({ type: 'REQUEST_RAIL_REFRESH' });
    dispatch({ type: 'RESCAN_CLOSED' });
    onClose();
  }

  const step = state.rescan?.step ?? 'pick-volume';
  const diff = state.rescan?.diff ?? null;
  const rescanSourcePath = state.rescan?.sourcePath ?? null;

  // Wires ⌘↵ to the same handler "Start re-scan" uses, only while step 1 is
  // active -- mirrors CreateSlideOver's own keydown effect.
  useEffect(() => {
    if (step !== 'pick-volume') return;
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
        event.preventDefault();
        handleStart();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [step, selectedSource]);

  // Triggered identically to Create's own source-loss step: the picked
  // volume vanishes mid-walk. Checked ahead of the ordinary step-2 title so
  // it always wins while step is 'scanning' and the shared slice has
  // failed.
  const isErrorStep = step === 'scanning' && state.scan.status === 'error';

  const headerTitle = isErrorStep
    ? 'Scan interrupted'
    : step === 'diff'
      ? oldTreeAvailable
        ? `Re-scan changed ${diff ? diffTotal(diff) : 0} entries`
        // Deliberately doesn't presuppose which resolution is coming --
        // the scan alone fixes nothing, and "Keep both" specifically
        // leaves the original unreadable catalog untouched
        // (28-UI-SPEC.md's Step 3 Variant B).
        : 'Scan complete'
      : step === 'scanning'
        ? `Re-scanning ${selectedSource ? sourceDisplayNameOf(selectedSource) : ''}`
        : `Re-scan ${catalog.title}`;
  const headerStep = isErrorStep
    ? 'failed'
    : step === 'diff'
      ? 'step 3 of 3'
      : step === 'scanning'
        ? 'step 2 of 3'
        : 'step 1 of 3';

  return (
    <div className="ws-rescan-scrim" onClick={handleCloseRequest}>
      <div
        className="ws-rescan-panel"
        ref={containerRef}
        role="dialog"
        aria-modal="true"
        aria-label="Re-scan volume"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="ws-rescan-header">
          <span className="ws-rescan-header-title">{headerTitle}</span>
          <span className="ws-rescan-step-label mono">{headerStep}</span>
          <button type="button" className="ws-rescan-close" aria-label="Close" onClick={handleCloseRequest}>
            ×
          </button>
        </div>

        {step === 'pick-volume' && (
          <div className="ws-rescan-body">
            <VolumePicker selected={selectedSource} onSelect={setSelectedSource} />
          </div>
        )}

        {step === 'scanning' && (state.scan.status === 'counting' || state.scan.status === 'scanning') && (
          <div className="ws-rescan-body">
            <ScanningBody scan={state.scan} />
          </div>
        )}

        {/* No partial-write affordance renders anywhere in this flow -- a
            partial diff (a fully-loaded old tree against a half-walked new
            one) has no well-defined resolution, so writingPartial/
            onWritePartial are both omitted and "Retry scan" takes the
            primary-styled slot in the omitted button's place. */}
        {isErrorStep && state.scan.status === 'error' && (
          <div className="ws-rescan-body">
            <ErrorBody
              scan={state.scan}
              onRetry={handleStart}
              onCloseWithoutWriting={handleCloseRequest}
              explanation={`Nothing about ${catalog.filename} has changed — your existing catalog is exactly as it was. Retry the scan, or close without writing anything.`}
            />
          </div>
        )}

        {step === 'diff' && diff && oldTreeAvailable && (
          <div className="ws-rescan-body ws-rescan-body-diff">
            <div className="ws-rescan-subline mono">
              {catalog.filename} · scanned {rescanSourcePath} just now
            </div>
            {/* Rendered above the stat grid, Variant A only, only when the
                Go side's low-similarity flag is set -- purely informational,
                never disables or gates any footer action (28-UI-SPEC.md's
                Similarity Warning Contract). */}
            {diff.lowSimilarity && (
              <div className="ws-rescan-warn">
                This looks like a different volume than the one this catalog came from — {diff.added + diff.removed}{' '}
                of {diffGrandTotal(diff)} entries differ. Double-check before overwriting; a large, legitimate change
                is still possible.
              </div>
            )}
            <div className="ws-rescan-stats">
              {STAT_TILES.map((tile) => (
                <div key={tile.key} className="ws-rescan-stat">
                  <span className="ws-rescan-stat-value mono" style={{ color: tile.color }}>
                    {diff[tile.key]} files
                  </span>
                  <span className="ws-rescan-stat-label">{tile.label}</span>
                </div>
              ))}
            </div>
            <DiffList entries={diff.entries} />
            <div className="ws-rescan-caption">
              Overwriting replaces {catalog.filename}&rsquo;s current contents with this scan&rsquo;s results — the
              previous version can't be recovered. Choose Keep both instead if you want to preserve it.
            </div>
          </div>
        )}

        {step === 'diff' && diff && !oldTreeAvailable && (
          <div className="ws-rescan-body ws-rescan-body-diff">
            <div className="ws-rescan-subline mono">
              {(() => {
                const stats = newTreeStatsFrom(diff);
                return `${stats.fileCount} files · ${formatBytes(stats.totalBytes)} · scanned ${rescanSourcePath} just now`;
              })()}
            </div>
            <div className="ws-rescan-caption">
              Overwriting rebuilds {catalog.filename} in place from this scan. Choosing Keep both leaves{' '}
              {catalog.filename} exactly as unreadable as it is now, and saves this scan separately.
            </div>
          </div>
        )}

        {step === 'pick-volume' && (
          <div className="ws-rescan-footer">
            <button
              type="button"
              className="ws-rescan-btn-primary"
              disabled={!selectedSource || submitting}
              onClick={handleStart}
            >
              Start re-scan
            </button>
            <span className="mono" style={{ fontSize: 11, color: 'var(--fn)' }}>
              ⌘↵
            </span>
            <button type="button" className="ws-rescan-btn-text" style={{ marginLeft: 'auto' }} onClick={handleCloseRequest}>
              Cancel
            </button>
          </div>
        )}

        {step === 'diff' && resolveError && <div className="ws-rescan-resolve-error">{resolveError}</div>}

        {step === 'diff' && (
          <div className="ws-rescan-footer">
            <div className="ws-rescan-footer-actions">
              {/* Accent-filled, not danger-styled -- the handoff's own
                  choice, unchanged by CONTEXT. The destructiveness is
                  carried by the resolution caption above, not this button's
                  color (28-UI-SPEC.md's Resolution Footer Contract). */}
              <button
                type="button"
                className="ws-rescan-btn-primary"
                disabled={resolving}
                onClick={() => handleResolve('overwrite')}
              >
                Overwrite catalog
              </button>
              {/* No filename in the label -- the actual write target comes
                  from the backend's shared collision loop (nextCopyRoot,
                  reused unmodified from Duplicate) and can land on
                  "-copy-2" or later if "-copy" is already taken. Naming a
                  file here that the write might not use would be a stale
                  promise; the button says only what it always does. */}
              <button
                type="button"
                className="ws-rescan-btn-outline"
                disabled={resolving}
                onClick={() => handleResolve('keep-both')}
              >
                Keep both
              </button>
            </div>
            <button
              type="button"
              className="ws-rescan-btn-text"
              style={{ marginLeft: 'auto' }}
              disabled={resolving}
              onClick={handleCloseRequest}
            >
              Discard scan and close
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default RescanDialog;
