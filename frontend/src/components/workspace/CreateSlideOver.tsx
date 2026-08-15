import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { useModalBehavior } from '../../hooks/useModalBehavior';
import { formatBytes } from '../../lib/format';
import { slugifyRoot } from '../../lib/scanFormat';
import { safeGetItem } from '../../themeTokens';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import type { CloseReason, ScanProgress, ScanSource } from '../../types/scan';
import { classifyScanFailure, sourceDisplayNameOf, sourcePathOf } from '../../types/scan';
import VolumePicker from './create/VolumePicker';
import CreateForm from './create/CreateForm';
import OptionsToggles, { SECONDARY_DIR_STORAGE_KEY, type OptionsToggleValues } from './create/OptionsToggles';
import ScanningBody from './create/ScanningBody';

export interface CreateSlideOverProps {
  isOpen: boolean;
  onClose: () => void;
}

const EXIT_DURATION_MS = 260;

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

  const [selectedSource, setSelectedSource] = useState<ScanSource | null>(null);
  const [title, setTitle] = useState('');
  const [root, setRoot] = useState('');
  const [options, setOptions] = useState<OptionsToggleValues>({
    writeHTML: true,
    copyToSecondary: false,
    includeHidden: false,
  });
  // Read once at mount, same persisted key OptionsToggles writes to -- both
  // start from the same value, and OptionsToggles reports every change back
  // through onSecondaryDirChange so this copy never drifts from its own.
  const [secondaryDir, setSecondaryDir] = useState(() => safeGetItem(SECONDARY_DIR_STORAGE_KEY) ?? '');
  const [submitting, setSubmitting] = useState(false);
  const submittingRef = useRef(false);

  // useModalBehavior gets the real isOpen, never the closing-inclusive
  // render condition below -- its scroll-unlock and focus-restore must fire
  // the instant a close is requested, not 260ms later when the exit
  // animation finishes (same contract CommandPalette documents).
  // handleCloseRequest is a hoisted function declaration (defined further
  // down in this component), so referencing it here is safe -- by the time
  // Escape can actually fire, the component has already finished at least
  // one full render.
  const { containerRef } = useModalBehavior({
    isOpen,
    onClose: () => {
      handleCloseRequest('cancel-the-scan');
    },
  });

  const scan = state.scan;

  // Always subscribed, never gated on isOpen -- CreateSlideOver is always
  // mounted by WorkspaceShell (see this component's own top comment), so a
  // scan sent to the background via "Run in background" (CRT-08) keeps
  // updating AppContext's lifted scan state -- and therefore the status
  // bar's segment -- while the panel itself is closed. Returning EventsOn's
  // own unsubscribe function is what keeps a StrictMode double-invoke from
  // leaking a second listener.
  useEffect(() => {
    const unsubscribe = EventsOn('scan:progress', (payload: ScanProgress) => {
      dispatch({ type: 'SCAN_PROGRESS', payload });
    });
    return unsubscribe;
  }, [dispatch]);

  // The single close handler every close trigger routes through (CRT-01,
  // CRT-09). During the scanning state a "cancel-the-scan" request cancels
  // the underlying context before the exit runs; "leave-it-running" (Run in
  // background) and every non-scanning-state request close with no
  // cancellation. Never a bespoke handler per trigger.
  async function handleCloseRequest(reason: CloseReason) {
    const currentlyScanning = scan.status === 'counting' || scan.status === 'scanning';
    if (currentlyScanning && reason === 'cancel-the-scan') {
      await wailsAPI.cancelScan();
    }
    onClose();
  }

  async function handleCreate() {
    // The ref guard (not the `submitting` state) is what actually makes a
    // double-click/double-⌘↵ start exactly one scan -- state updates are
    // batched and would not be visible to a second synchronous call before
    // this component re-renders (CRT-06 idempotency/concurrency).
    if (submittingRef.current) return;
    if (scan.status !== 'idle' && scan.status !== 'error') return;
    if (!selectedSource || !state.catalogDir) return;

    submittingRef.current = true;
    setSubmitting(true);

    const sourcePath = sourcePathOf(selectedSource);
    const displayName = sourceDisplayNameOf(selectedSource);
    const resolvedTitle = title.trim() || displayName;
    const resolvedRoot = root.trim() || slugifyRoot(displayName) || 'catalog';
    // Zero means "no known total" -- resolveScanTotal then runs a count-only
    // pre-pass instead (CRT-07). A plain folder has no volume-level probe.
    const totalBytesHint = selectedSource.kind === 'volume' ? selectedSource.volume.totalBytes : 0;

    dispatch({ type: 'SCAN_STARTED', payload: { title: resolvedTitle } });

    const outcome = await wailsAPI.startScan(resolvedTitle, sourcePath, state.catalogDir, resolvedRoot, {
      writeHTML: options.writeHTML,
      includeHidden: options.includeHidden,
      copyToDirectory: options.copyToSecondary ? secondaryDir : '',
      totalBytesHint,
    });

    submittingRef.current = false;
    setSubmitting(false);

    if (!outcome.success) {
      // A cancellation (Escape / close / scrim / window close mid-scan)
      // produces no error UI at all -- the panel is already closing, so the
      // scan simply resets to idle. Only a source loss (the volume vanished
      // mid-walk) transitions to the error member, which plan 25-07 renders.
      const failure = classifyScanFailure(outcome.error);
      if (failure.kind === 'sourceLoss') {
        dispatch({ type: 'SCAN_FAILED', payload: { message: failure.message } });
      } else {
        dispatch({ type: 'SCAN_RESET' });
      }
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
  // running a no-op regardless of which path triggered it. The listener is
  // only registered while the panel is open AND the form step is active
  // (idle/error) -- inlined rather than reading the later `isForm` const,
  // which is declared after this effect in render order.
  useEffect(() => {
    if (!isOpen) return;
    if (scan.status !== 'idle' && scan.status !== 'error') return;
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
        event.preventDefault();
        handleCreate();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, selectedSource, title, root, scan.status, state.catalogDir]);

  // The panel keeps rendering for the full 260ms exit animation after a
  // close request, not just until isOpen flips false.
  if (!(isOpen || closing)) return null;

  const isScanning = scan.status === 'counting' || scan.status === 'scanning';
  const isDone = scan.status === 'done';
  const isForm = !isScanning && !isDone;

  const headerTitle = isDone ? 'Catalog written' : isScanning ? 'Cataloguing volume' : 'New catalog';
  const headerStep = isDone ? 'done' : isScanning ? 'step 2 of 2' : 'step 1 of 2';

  const canCreate = !submitting && (scan.status === 'idle' || scan.status === 'error') && !!selectedSource && !!state.catalogDir;

  // The rail's already-loaded listing for the configured catalog directory
  // -- the cheapest correct source for the WILL WRITE preview's "already
  // exists" qualifier, per this plan's explicit no-new-binding instruction.
  const existingCatalogFilenames = new Set(state.catalogs.map((catalog) => catalog.filename));

  // Shared with CreateForm's own (identically-derived) preview value and
  // with the write-HTML toggle's dynamic `{root}.html` note -- both must
  // read the same resolved root a scan would actually use.
  const sourceDisplayName = selectedSource ? sourceDisplayNameOf(selectedSource) : '';
  const effectiveRoot = root.trim() || (sourceDisplayName && slugifyRoot(sourceDisplayName)) || 'catalog';
  const activeSecondaryDir = options.copyToSecondary ? secondaryDir : '';

  return (
    <div
      className={`ws-create-scrim${isOpen ? '' : ' ws-create-scrim-exit'}`}
      onClick={() => handleCloseRequest('cancel-the-scan')}
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
          <button
            type="button"
            className="ws-create-close"
            onClick={() => handleCloseRequest('cancel-the-scan')}
            aria-label="Close"
          >
            ×
          </button>
        </div>

        {isForm && (
          <div className="ws-create-body">
            {scan.status === 'error' && <div className="ws-create-error-banner">{scan.message}</div>}
            <VolumePicker selected={selectedSource} onSelect={setSelectedSource} />
            <CreateForm
              source={selectedSource}
              title={title}
              onTitleChange={setTitle}
              root={root}
              onRootChange={setRoot}
              catalogDir={state.catalogDir ?? ''}
              writeHTML={options.writeHTML}
              secondaryDir={activeSecondaryDir}
              existingCatalogFilenames={existingCatalogFilenames}
            />
            <OptionsToggles
              values={options}
              onValuesChange={setOptions}
              secondaryDir={secondaryDir}
              onSecondaryDirChange={setSecondaryDir}
              disabled={isScanning}
              effectiveRoot={effectiveRoot}
            />
            {!state.catalogDir && (
              <div className="ws-create-error-banner">
                Choose a catalog directory from the rail before creating a catalog.
              </div>
            )}
          </div>
        )}

        {(scan.status === 'counting' || scan.status === 'scanning') && (
          <ScanningBody scan={scan} onRunInBackground={() => handleCloseRequest('leave-it-running')} />
        )}

        {isDone && scan.status === 'done' && (
          <div className="ws-create-body">
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
          </div>
        )}

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
              onClick={() => handleCloseRequest('cancel-the-scan')}
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
