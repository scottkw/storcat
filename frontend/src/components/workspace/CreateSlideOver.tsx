import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { useModalBehavior } from '../../hooks/useModalBehavior';
import { slugifyRoot } from '../../lib/scanFormat';
import { safeGetItem } from '../../themeTokens';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import type { CloseReason, ScanProgress, ScanResultFile, ScanSource } from '../../types/scan';
import { classifyScanFailure, sourceDisplayNameOf, sourcePathOf } from '../../types/scan';
import VolumePicker from './create/VolumePicker';
import CreateForm from './create/CreateForm';
import OptionsToggles, { SECONDARY_DIR_STORAGE_KEY, type OptionsToggleValues } from './create/OptionsToggles';
import ScanningBody from './create/ScanningBody';
import ErrorBody from './create/ErrorBody';
import DoneBody from './create/DoneBody';

export interface CreateSlideOverProps {
  isOpen: boolean;
  onClose: () => void;
}

const EXIT_DURATION_MS = 260;

// Shared by the form step's initial useState and "Catalog another volume"'s
// reset (DoneBody, plan 25-07) -- one literal, not two, so the two can never
// drift apart.
const DEFAULT_OPTIONS: OptionsToggleValues = {
  writeHTML: true,
  copyToSecondary: false,
  includeHidden: false,
};

// Shared by handleCreate's success branch and handleWritePartial: one row
// per file CreateCatalogResult actually reports as written, gated on each
// path field's own presence rather than the toggle values -- htmlPath is
// always a field (possibly ""), copyJsonPath/copyHtmlPath are omitempty, so
// checking the string itself is the same "was this actually written" test
// either way. Structurally typed (not a named import) so both StartScan's
// and WritePartialCatalog's identically-shaped results satisfy it.
function filesFromResult(result: {
  jsonPath: string;
  jsonSize: number;
  htmlPath: string;
  htmlSize?: number;
  copyJsonPath?: string;
  copyJsonSize?: number;
  copyHtmlPath?: string;
  copyHtmlSize?: number;
}): ScanResultFile[] {
  const files: ScanResultFile[] = [{ path: result.jsonPath, size: result.jsonSize }];
  if (result.htmlPath) files.push({ path: result.htmlPath, size: result.htmlSize });
  if (result.copyJsonPath) files.push({ path: result.copyJsonPath, size: result.copyJsonSize });
  if (result.copyHtmlPath) files.push({ path: result.copyHtmlPath, size: result.copyHtmlSize });
  return files;
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

  const [selectedSource, setSelectedSource] = useState<ScanSource | null>(null);
  const [title, setTitle] = useState('');
  // Seeded from the Settings-owned default, not the empty string -- when
  // the setting is empty this is a behavioural no-op and CreateForm's own
  // source-derived placeholder still applies (SET-04, CRT-04 adjacency).
  const [root, setRoot] = useState(() => state.settings.defaultFilenameRoot);
  const [options, setOptions] = useState<OptionsToggleValues>(DEFAULT_OPTIONS);
  // Read once at mount, same persisted key OptionsToggles writes to -- both
  // start from the same value, and OptionsToggles reports every change back
  // through onSecondaryDirChange so this copy never drifts from its own.
  const [secondaryDir, setSecondaryDir] = useState(() => safeGetItem(SECONDARY_DIR_STORAGE_KEY) ?? '');
  const [submitting, setSubmitting] = useState(false);
  const submittingRef = useRef(false);
  // Guards "Write partial catalog" the same way submittingRef guards
  // Create: the ref is what actually stops a second click's async body from
  // proceeding (state updates aren't visible synchronously); the `writing`
  // state is what visibly disables the button so a second click can't even
  // be issued in the first place (CRT-11 idempotency, T-25-13).
  const [writingPartial, setWritingPartial] = useState(false);
  const writingPartialRef = useRef(false);

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

  // Consumes the folder-picker intent set by the tree pane's secondary
  // entry point (TreePane.tsx): when the panel opens with the flag set,
  // invoke the folder dialog once and clear the flag immediately -- before
  // the dialog even resolves, so a re-render (or a later close/reopen)
  // never re-triggers it. Lands the user directly on the picker that entry
  // point's label promises, rather than the generic volume-card list.
  useEffect(() => {
    if (!isOpen) return;
    if (!state.createFolderPickerIntent) return;
    dispatch({ type: 'SET_CREATE_FOLDER_PICKER_INTENT', payload: false });
    wailsAPI.selectDirectory().then((result) => {
      if (result.success && result.path) {
        setSelectedSource({ kind: 'folder', path: result.path });
      }
    });
  }, [isOpen, state.createFolderPickerIntent, dispatch]);

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
    // Defense in depth alongside ErrorBody's disabled={writingPartial} on
    // "Retry scan" (WR-02): even if that disabled prop were somehow
    // bypassed, this guard stops a retry from clearing the retained partial
    // tree while "Write partial catalog" is still writing it (CR-02).
    if (writingPartialRef.current) return;
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
    // Captured locally rather than re-read from state.scan.startedAt after
    // the fact -- state updates are async, and by the time this function's
    // await resolves, scan.startedAt could belong to a different render's
    // closure (or, after a retry, a stale one). This local value is what
    // the completed scan's real duration is measured against.
    const startedAt = Date.now();

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
        dispatch({ type: 'SCAN_FAILED', payload: { message: failure.message, sourcePath } });
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
        files: filesFromResult(result),
        fileCount: result.fileCount,
        totalSize: result.totalSize,
        durationMs: Date.now() - startedAt,
        partial: false,
      },
    });
  }

  // The error state's primary action (CRT-11): writes the tree retained
  // from the source loss through the shared write path, exactly once. The
  // synchronous ref guard is what actually stops a second call's async body
  // from proceeding before this component re-renders; setWritingPartial is
  // what visibly disables the button (see the CreateSlideOver-level state
  // declaration above for why both exist).
  async function handleWritePartial() {
    if (writingPartialRef.current) return;
    if (scan.status !== 'error') return;

    writingPartialRef.current = true;
    setWritingPartial(true);

    const failedScan = scan;
    const outcome = await wailsAPI.writePartialCatalog();

    writingPartialRef.current = false;
    setWritingPartial(false);

    if (!outcome.success) {
      // The write failed after all (e.g. nothing was actually retained --
      // a stale panel reopened after a new scan already started elsewhere).
      // The error state is left exactly as it was; nothing here fabricates
      // a done state on a failure the binding itself reported.
      console.error('writePartialCatalog failed:', outcome.error);
      return;
    }

    const result = outcome.result;
    dispatch({
      type: 'SCAN_DONE',
      payload: {
        title: failedScan.title,
        jsonPath: result.jsonPath,
        files: filesFromResult(result),
        fileCount: result.fileCount,
        totalSize: result.totalSize,
        // A partial write's duration is meaningless (the walk that
        // produced this tree never finished) -- stopPercent is what
        // DoneBody's partial flavour renders in its place.
        durationMs: 0,
        partial: true,
        stopPercent: failedScan.stopPercent,
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

  // The done state's secondary action -- does NOT close the panel (unlike
  // every CRT-01 close path): resets the form fields to their defaults and
  // returns to the form step in the same still-open panel. No re-entry
  // animation is needed since the panel itself never left; VolumePicker's
  // own on-mount effect re-enumerates volumes fresh the instant it remounts
  // (SCAN_RESET flips scan.status to 'idle', so isForm renders it again).
  function handleCatalogAnother() {
    setSelectedSource(null);
    setTitle('');
    setRoot(state.settings.defaultFilenameRoot);
    setOptions(DEFAULT_OPTIONS);
    dispatch({ type: 'SCAN_RESET' });
  }

  // handleCreate is a fresh closure every render (it reads options,
  // secondaryDir, etc. from render-scope state) -- kept in a ref, updated on
  // every render, so the keydown listener below always calls the *current*
  // render's handleCreate no matter which deps the listener-registration
  // effect actually re-runs on. Without this, toggling an option (which
  // isn't in that effect's dep array) then pressing ⌘↵ would start a scan
  // using the *previous* render's option values -- diverging from the WILL
  // WRITE preview the user is looking at (WR-01).
  const handleCreateRef = useRef(handleCreate);
  handleCreateRef.current = handleCreate;

  // Wires ⌘↵ to the same handler the Create button uses (CRT-06) --
  // handleCreate's own guards make a second activation while a scan is
  // running a no-op regardless of which path triggered it. The listener is
  // only registered while the panel is open AND the form step is active
  // (idle/error) -- inlined rather than reading the later `isForm` const,
  // which is declared after this effect in render order. Always calls
  // handleCreateRef.current() (never handleCreate directly) so the deps
  // below only need to control *when the listener is (de)registered*, not
  // which closure it captures.
  useEffect(() => {
    if (!isOpen) return;
    if (scan.status !== 'idle' && scan.status !== 'error') return;
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
        event.preventDefault();
        handleCreateRef.current();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [isOpen, scan.status]);

  // The panel keeps rendering for the full 260ms exit animation after a
  // close request, not just until isOpen flips false.
  if (!(isOpen || closing)) return null;

  const isScanning = scan.status === 'counting' || scan.status === 'scanning';
  const isDone = scan.status === 'done';
  const isError = scan.status === 'error';
  const isForm = !isScanning && !isDone && !isError;

  const headerTitle = isDone ? 'Catalog written' : isError ? 'Scan interrupted' : isScanning ? 'Cataloguing volume' : 'New catalog';
  const headerStep = isDone ? 'done' : isError ? 'failed' : isScanning ? 'step 2 of 2' : 'step 1 of 2';

  // The error state's own Retry action reuses handleCreate directly (with
  // scan.status === 'error') rather than a bespoke restart path -- this
  // button only ever renders while isForm, i.e. scan.status === 'idle'.
  const canCreate = !submitting && scan.status === 'idle' && !!selectedSource && !!state.catalogDir;

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

        {isError && scan.status === 'error' && (
          <ErrorBody
            scan={scan}
            writingPartial={writingPartial}
            onWritePartial={handleWritePartial}
            onRetry={handleCreate}
            onCloseWithoutWriting={() => handleCloseRequest('cancel-the-scan')}
          />
        )}

        {isDone && scan.status === 'done' && (
          <DoneBody scan={scan} onOpenInWorkspace={handleOpenInWorkspace} onCatalogAnother={handleCatalogAnother} />
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

      </div>
    </div>
  );
}

export default CreateSlideOver;
