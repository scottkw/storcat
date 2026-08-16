import { useEffect, useRef, useState } from 'react';
import { useAppContext } from '../../../contexts/AppContext';
import { wailsAPI } from '../../../services/wailsAPI';
import { useModalBehavior } from '../../../hooks/useModalBehavior';
import { classifyScanFailure, sourceDisplayNameOf, sourcePathOf, type ScanSource } from '../../../types/scan';
import type { DiffResult, DiffState } from '../../../types/rescan';
import { models } from '../../../../wailsjs/go/models';
import VolumePicker from '../create/VolumePicker';
import ScanningBody from '../create/ScanningBody';

export interface RescanDialogProps {
  catalog: models.CatalogMetadata;
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

function diffTotal(diff: DiffResult): number {
  return diff.added + diff.removed + diff.changed + diff.unreadable;
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
function RescanDialog({ catalog, oldTreeAvailable, onClose, onError }: RescanDialogProps) {
  const { state, dispatch } = useAppContext();
  const [selectedSource, setSelectedSource] = useState<ScanSource | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const submittingRef = useRef(false);

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
      dispatch({ type: 'SCAN_RESET' });
      // Back to step 1 with the same catalog/oldTreeAvailable -- lets the
      // user retry without losing the dialog. The full Retry/Close error
      // step (28-UI-SPEC.md's Error Step) is plan 28-02+; this tracer
      // surfaces the failure through the same error slot Footer's other
      // actions already share.
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

  const headerTitle =
    step === 'diff'
      ? `Re-scan changed ${diff ? diffTotal(diff) : 0} entries`
      : step === 'scanning'
        ? `Re-scanning ${selectedSource ? sourceDisplayNameOf(selectedSource) : ''}`
        : `Re-scan ${catalog.title}`;
  const headerStep = step === 'diff' ? 'step 3 of 3' : step === 'scanning' ? 'step 2 of 3' : 'step 1 of 3';

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

        {step === 'diff' && diff && (
          <div className="ws-rescan-body">
            <div className="ws-rescan-subline mono">
              {catalog.filename} · scanned {rescanSourcePath} just now
            </div>
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

        {step === 'diff' && (
          <div className="ws-rescan-footer">
            <button
              type="button"
              className="ws-rescan-btn-text"
              style={{ marginLeft: 'auto' }}
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
