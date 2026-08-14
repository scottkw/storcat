import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { useModalBehavior } from '../../hooks/useModalBehavior';
import { formatBytes } from '../../lib/format';
import { scanPercent } from '../../lib/scanFormat';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import type { ScanProgress } from '../../types/scan';

export interface CreateSlideOverProps {
  isOpen: boolean;
  onClose: () => void;
}

const EXIT_DURATION_MS = 260;

function basename(path: string): string {
  const trimmed = path.replace(/[\\/]+$/, '');
  const idx = Math.max(trimmed.lastIndexOf('/'), trimmed.lastIndexOf('\\'));
  return idx >= 0 ? trimmed.slice(idx + 1) : trimmed;
}

function slugify(name: string): string {
  const slug = name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
  return slug || 'catalog';
}

// Always mounted by WorkspaceShell (same pattern as CommandPalette) and must
// not be conditionally mounted -- the shared useModalBehavior hook below
// observes the isOpen: true -> false transition to release scroll lock and
// restore focus, and this component's animated exit depends on the same
// contract. Unlike the palette, this component owns a local `closing` flag
// so the 260ms exit is visible before it stops rendering.
function CreateSlideOver({ isOpen, onClose }: CreateSlideOverProps) {
  const { state, dispatch } = useAppContext();

  const [closing, setClosing] = useState(false);
  const wasOpenRef = useRef(isOpen);
  const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // A single layout effect, keyed on isOpen, owns the whole closing
  // lifecycle. useLayoutEffect (not useEffect) is what keeps this
  // flicker-free: it runs synchronously after the DOM commits but before
  // the browser paints, so the true->false render (which would otherwise
  // return null for one frame, per the render condition below) is
  // superseded by this effect's setClosing(true) before that empty frame
  // is ever painted.
  useLayoutEffect(() => {
    if (isOpen) {
      wasOpenRef.current = true;
      if (closeTimerRef.current) {
        clearTimeout(closeTimerRef.current);
        closeTimerRef.current = null;
      }
      setClosing(false);
      return;
    }
    if (!wasOpenRef.current) return; // never opened yet -- nothing to animate out
    wasOpenRef.current = false;
    setClosing(true);
    closeTimerRef.current = setTimeout(() => {
      setClosing(false);
      closeTimerRef.current = null;
    }, EXIT_DURATION_MS);
    return () => {
      if (closeTimerRef.current) {
        clearTimeout(closeTimerRef.current);
        closeTimerRef.current = null;
      }
    };
  }, [isOpen]);

  const [sourcePath, setSourcePath] = useState('');
  const [title, setTitle] = useState('');
  const [root, setRoot] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const submittingRef = useRef(false);

  // useModalBehavior gets the real isOpen, never the closing-inclusive
  // render condition below -- its scroll-unlock and focus-restore must fire
  // the instant a close is requested, not 260ms later when the exit
  // animation finishes (same contract CommandPalette documents).
  const { containerRef } = useModalBehavior({ isOpen, onClose });

  const scan = state.scan;

  // Progress subscription is keyed on the open state, not on scan status --
  // it tears down when the panel closes (no "run in background" live
  // updates this plan; that is CRT-08, out of scope here) and is
  // re-established on reopen. Returning EventsOn's own unsubscribe function
  // is what keeps a StrictMode double-invoke from leaking a second listener.
  useEffect(() => {
    if (!isOpen) return;
    const unsubscribe = EventsOn('scan:progress', (payload: ScanProgress) => {
      dispatch({ type: 'SCAN_PROGRESS', payload });
    });
    return unsubscribe;
  }, [isOpen, dispatch]);

  async function handleChooseFolder() {
    const result = await wailsAPI.selectDirectory();
    if (!result.success || !result.path) return;
    setSourcePath(result.path);
  }

  async function handleCreate() {
    // The ref guard (not the `submitting` state) is what actually makes a
    // double-click/double-⌘↵ start exactly one scan -- state updates are
    // batched and would not be visible to a second synchronous call before
    // this component re-renders (CRT-06 idempotency/concurrency).
    if (submittingRef.current) return;
    if (scan.status !== 'idle' && scan.status !== 'error') return;
    if (!sourcePath || !state.catalogDir) return;

    submittingRef.current = true;
    setSubmitting(true);

    const resolvedTitle = title.trim() || basename(sourcePath);
    const resolvedRoot = root.trim() || slugify(basename(sourcePath));

    dispatch({ type: 'SCAN_STARTED', payload: { title: resolvedTitle } });

    const outcome = await wailsAPI.startScan(resolvedTitle, sourcePath, state.catalogDir, resolvedRoot, {
      writeHTML: true,
      includeHidden: false,
      copyToDirectory: '',
    });

    submittingRef.current = false;
    setSubmitting(false);

    if (!outcome.success) {
      dispatch({ type: 'SCAN_FAILED', payload: { message: outcome.error } });
      return;
    }

    const result = outcome.result;
    dispatch({
      type: 'SCAN_DONE',
      payload: {
        title: resolvedTitle,
        jsonPath: result.jsonPath,
        // Go's CreateCatalogResult reports the catalog's total scanned
        // content size, not each output file's own on-disk byte count --
        // no `size` field is set here rather than fabricating one.
        files: result.htmlPath
          ? [{ path: result.jsonPath }, { path: result.htmlPath }]
          : [{ path: result.jsonPath }],
        fileCount: result.fileCount,
        totalSize: result.totalSize,
        durationMs: 0,
        partial: false,
      },
    });
  }

  async function handleOpenInWorkspace() {
    if (scan.status !== 'done') return;
    if (state.catalogDir) {
      const result = await wailsAPI.browseCatalogs(state.catalogDir);
      if (result.success) {
        dispatch({ type: 'SET_CATALOGS', payload: result.catalogs ?? [] });
      }
      dispatch({ type: 'SELECT_CATALOG', payload: scan.jsonPath });
    }
    dispatch({ type: 'SCAN_RESET' });
    onClose();
  }

  // Wires ⌘↵ to the same handler the Create button uses (CRT-06) --
  // handleCreate's own guards make a second activation while a scan is
  // running a no-op regardless of which path triggered it.
  useEffect(() => {
    if (!isOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
        event.preventDefault();
        handleCreate();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, sourcePath, title, root, scan.status, state.catalogDir]);

  // The panel keeps rendering for the full 260ms exit animation after a
  // close request, not just until isOpen flips false.
  if (!(isOpen || closing)) return null;

  const isScanning = scan.status === 'counting' || scan.status === 'scanning';
  const isDone = scan.status === 'done';
  const isForm = !isScanning && !isDone;

  const headerTitle = isDone ? 'Catalog written' : isScanning ? 'Cataloguing volume' : 'New catalog';
  const headerStep = isDone ? 'done' : isScanning ? 'step 2 of 2' : 'step 1 of 2';

  const canCreate = !submitting && (scan.status === 'idle' || scan.status === 'error') && !!sourcePath && !!state.catalogDir;

  const pct = scan.status === 'scanning' ? scanPercent(scan.bytesSeen, scan.totalBytes) : null;

  return (
    <div
      className={`ws-create-scrim${isOpen ? '' : ' ws-create-scrim-exit'}`}
      onClick={onClose}
    >
      <div
        className={`ws-create-panel${isOpen ? '' : ' ws-create-panel-exit'}`}
        ref={containerRef}
        role="dialog"
        aria-modal="true"
        aria-label="Create catalog"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="ws-create-header">
          <span className="ws-create-header-title">{headerTitle}</span>
          <span className="ws-create-header-step mono">{headerStep}</span>
          <button type="button" className="ws-create-close" onClick={onClose} aria-label="Close">
            ×
          </button>
        </div>

        <div className="ws-create-body">
          {isForm && (
            <>
              {scan.status === 'error' && <div className="ws-create-error-banner">{scan.message}</div>}
              <div className="ws-create-field">
                <label className="ws-create-label">Source folder</label>
                <button type="button" className="ws-create-choose-folder" onClick={handleChooseFolder}>
                  {sourcePath || 'Choose a folder…'}
                </button>
              </div>
              <div className="ws-create-field">
                <label className="ws-create-label">Catalog title</label>
                <input
                  className="ws-create-input"
                  value={title}
                  onChange={(event) => setTitle(event.target.value)}
                  placeholder={sourcePath ? basename(sourcePath) : 'Untitled catalog'}
                />
              </div>
              <div className="ws-create-field">
                <label className="ws-create-label">Filename root</label>
                <div className="ws-create-input-suffix-row">
                  <input
                    className="ws-create-input mono"
                    value={root}
                    onChange={(event) => setRoot(event.target.value)}
                    placeholder={sourcePath ? slugify(basename(sourcePath)) : 'catalog'}
                  />
                  <span className="ws-create-suffix mono">.json / .html</span>
                </div>
              </div>
              {!state.catalogDir && (
                <div className="ws-create-error-banner">
                  Choose a catalog directory from the rail before creating a catalog.
                </div>
              )}
            </>
          )}

          {isScanning && (
            <div className="ws-create-scan-body">
              <div className="ws-create-scan-title-row">
                <span className="ws-create-scan-title">{scan.title}</span>
                <span className="ws-create-scan-pct mono">{pct !== null ? `${pct}%` : 'Counting…'}</span>
              </div>
              {pct !== null && (
                <div className="ws-create-progress">
                  <div className="ws-create-progress-fill" style={{ width: `${pct}%` }} />
                </div>
              )}
              <div className="ws-create-scan-counters mono">
                {scan.status === 'scanning'
                  ? `${scan.filesSeen} files · ${formatBytes(scan.bytesSeen)}`
                  : `${scan.filesSeen} files found so far`}
              </div>
            </div>
          )}

          {isDone && scan.status === 'done' && (
            <div className="ws-create-done-body">
              <div className="ws-create-done-title">{scan.title} catalogued</div>
              <div className="ws-create-done-line mono">
                {scan.fileCount} files · {formatBytes(scan.totalSize)}
              </div>
              <div className="ws-create-file-list">
                {scan.files.map((file) => (
                  <div className="ws-create-file-row" key={file.path}>
                    <span className="ws-create-file-shape" aria-hidden="true" />
                    <span className="ws-create-file-path mono">{file.path}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {isForm && (
          <div className="ws-create-footer">
            <button
              type="button"
              className="ws-create-btn ws-create-btn-primary"
              disabled={!canCreate}
              onClick={handleCreate}
            >
              Create catalog
            </button>
            <span className="ws-create-hint mono">⌘↵</span>
            <button
              type="button"
              className="ws-create-btn ws-create-btn-text"
              style={{ marginLeft: 'auto' }}
              onClick={onClose}
            >
              Discard and close
            </button>
          </div>
        )}

        {isDone && (
          <div className="ws-create-footer">
            <button type="button" className="ws-create-btn ws-create-btn-primary" onClick={handleOpenInWorkspace}>
              Open in workspace
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default CreateSlideOver;
