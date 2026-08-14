import { useEffect, useState } from 'react';
import Toolbar from './Toolbar';
import CatalogRail from './CatalogRail';
import TreePane from './TreePane';
import DetailsPanel from './DetailsPanel';
import StatusBar from './StatusBar';
import CommandPalette from './CommandPalette';
import CreateSlideOver from './CreateSlideOver';
import { useMediaQuery } from '../../hooks/useMediaQuery';
import { useAppContext } from '../../contexts/AppContext';
import { applyTokens, readPersistedPrefs } from '../../themeTokens';

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

  // Density lives in the reducer; re-apply the token layer whenever it
  // changes so the reducer field is not inert. Theme isn't reducer state
  // yet (Phase 26), so it's re-read here for now.
  useEffect(() => {
    applyTokens(readPersistedPrefs().theme, state.density);
  }, [state.density]);

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
      <StatusBar />
    </div>
  );
}

export default WorkspaceShell;
