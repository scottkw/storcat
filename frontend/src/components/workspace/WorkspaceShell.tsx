import { useCallback, useEffect, useState } from 'react';
import Toolbar from './Toolbar';
import CatalogRail from './CatalogRail';
import TreePane from './TreePane';
import DetailsPanel from './DetailsPanel';
import StatusBar from './StatusBar';
import CommandPalette from './CommandPalette';
import CreateSlideOver from './CreateSlideOver';
import SettingsDialog from './SettingsDialog';
import { useMediaQuery } from '../../hooks/useMediaQuery';
import { useAppContext } from '../../contexts/AppContext';
import { applyTokens, readPersistedPrefs } from '../../themeTokens';
import { hydrateSettings } from '../../settingsStore';

export interface WorkspaceShellProps {
  themeName: string;
}

function WorkspaceShell({ themeName }: WorkspaceShellProps) {
  const { state, dispatch } = useAppContext();

  // Single place React learns the width tier -- must match workspace.css's
  // widest breakpoint character-for-character. Drives both the details
  // panel's pane-or-drawer variant and the toolbar's Details chip visibility.
  const isWide = useMediaQuery('(min-width: 1280px)');

  const [paletteOpen, setPaletteOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);

  // Density lives in the reducer; re-apply the token layer whenever it
  // changes so the reducer field is not inert. Theme isn't reducer state
  // yet (Phase 26), so it's re-read here for now.
  useEffect(() => {
    applyTokens(readPersistedPrefs().theme, state.density);
  }, [state.density]);

  // Runs the marker-gated localStorage-to-config migration and hydrates
  // the config-only settings slice (defaultFilenameRoot, etc.) plus the
  // catalog directory into AppContext. Empty deps -- fires once per mount;
  // hydrateSettings() itself is deduped behind a module-level in-flight
  // promise, so React 18 StrictMode's development double-invoke of this
  // effect still issues exactly one getConfig() round trip. The `cancelled`
  // flag matches Toolbar.tsx/DetailsPanel.tsx's own deferred-promise
  // pattern -- a result resolving after unmount is dropped, not dispatched.
  useEffect(() => {
    let cancelled = false;
    hydrateSettings().then((result) => {
      if (cancelled || !result) return;
      dispatch({ type: 'SETTINGS_HYDRATED', payload: result.settings });
      // The reducer's SET_CATALOG_DIR guard (AppContext.tsx) makes this a
      // no-op when CatalogRail's own mount effect already set the same
      // value first -- no ordering coupling exists between the two effects.
      if (result.catalogDirectory) {
        dispatch({ type: 'SET_CATALOG_DIR', payload: result.catalogDirectory });
      }
    });
    return () => {
      cancelled = true;
    };
  }, [dispatch]);

  // One close path for both Escape and backdrop click.
  const closeDrawer = () => dispatch({ type: 'SET_DETAIL_OVERLAY', payload: false });

  // Escape closes the drawer; listener only lives while the drawer is open.
  useEffect(() => {
    if (!state.detailOverlay) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') closeDrawer();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [state.detailOverlay]);

  // Crossing back above 1280px with the drawer open must not leave stale
  // overlay state behind -- the inline pane renders regardless of detailOverlay.
  useEffect(() => {
    if (isWide) dispatch({ type: 'SET_DETAIL_OVERLAY', payload: false });
  }, [isWide]);

  // Global ⌘K/Ctrl+K listener -- must still fire while the rail filter input
  // has focus (⌘K is a global override, no input-focus exclusion), and must
  // preventDefault unconditionally so no webview/browser default engages on
  // any platform. The functional state update makes a second press while
  // already open a no-op rather than a re-open that would discard the query
  // in progress. Opening the palette also closes the create slide-over --
  // both overlays share the --z-overlay layer and are mutually exclusive by
  // construction (25-UI-SPEC.md).
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        setPaletteOpen((open) => {
          if (!open) dispatch({ type: 'SET_CREATE_OPEN', payload: false });
          return open ? open : true;
        });
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [dispatch]);

  // Opening the create slide-over (from the rail's + New pill or any other
  // entry point) closes the palette, the other direction of the same
  // mutual-exclusion contract above.
  useEffect(() => {
    if (state.createOpen) setPaletteOpen(false);
  }, [state.createOpen]);

  // The single named function all three SET-01 entry points (this file's
  // ⌘,/Ctrl+, listener below, Toolbar's gear onClick, Toolbar's theme-chip
  // onClick) route through, so there is exactly one open path. Toolbar
  // entry points 2 and 3 are already unreachable by mouse while the palette
  // or create slide-over is open -- those scrims are position: absolute;
  // inset: 0 and cover the whole toolbar -- so only the keyboard path below
  // needs the explicit no-op-during-scan rule (26-UI-SPEC.md).
  //
  // Calling this while Settings is already open is a harmless no-op: every
  // state update below (setPaletteOpen(false), SET_CREATE_OPEN false,
  // setSettingsOpen(true)) is already a no-op against the current state, so
  // React bails out with no re-render.
  const openSettings = useCallback(() => {
    // A backgrounded scan is unaffected -- it keeps running in Go regardless
    // of UI mount state. Only a *foreground* scan (the slide-over actively
    // showing progress) blocks Settings from opening.
    if (state.scan.status === 'counting' || state.scan.status === 'scanning') return;
    setPaletteOpen(false);
    dispatch({ type: 'SET_CREATE_OPEN', payload: false });
    setSettingsOpen(true);
  }, [state.scan.status, dispatch]);

  // Global ⌘,/Ctrl+, listener -- mirrors the ⌘K listener above: preventDefault
  // unconditionally, fires regardless of which element has focus.
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === ',') {
        event.preventDefault();
        openSettings();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [openSettings]);

  // Reverse direction of the mutual-exclusion contract above: opening the
  // palette or the create slide-over closes Settings, so exactly one
  // overlay ever owns the screen.
  useEffect(() => {
    if (paletteOpen || state.createOpen) setSettingsOpen(false);
  }, [paletteOpen, state.createOpen]);

  return (
    <div className="ws-root" data-rail-side={state.railSide}>
      <Toolbar
        themeName={themeName}
        showDetailsChip={!isWide}
        detailsOpen={state.detailOverlay}
        onToggleDetails={() =>
          dispatch({ type: 'SET_DETAIL_OVERLAY', payload: !state.detailOverlay })
        }
        onOpenSearch={() => setPaletteOpen(true)}
        onOpenSettings={openSettings}
      />
      <div className="ws-grid">
        <CatalogRail />
        <TreePane />
        {isWide && <DetailsPanel variant="pane" />}
      </div>
      {!isWide && state.detailOverlay && (
        <>
          <div className="ws-backdrop" onClick={closeDrawer} />
          <DetailsPanel variant="drawer" />
        </>
      )}
      <CommandPalette isOpen={paletteOpen} onClose={() => setPaletteOpen(false)} />
      <CreateSlideOver isOpen={state.createOpen} onClose={() => dispatch({ type: 'SET_CREATE_OPEN', payload: false })} />
      <SettingsDialog isOpen={settingsOpen} onClose={() => setSettingsOpen(false)} />
      <StatusBar />
    </div>
  );
}

export default WorkspaceShell;
