import { useEffect } from 'react';
import Toolbar from './Toolbar';
import CatalogRail from './CatalogRail';
import TreePane from './TreePane';
import DetailsPanel from './DetailsPanel';
import StatusBar from './StatusBar';
import { useMediaQuery } from '../../hooks/useMediaQuery';
import { useAppContext } from '../../contexts/AppContext';
import { applyTokens, readPersistedPrefs } from '../../themeTokens';

function WorkspaceShell() {
  const { state, dispatch } = useAppContext();

  // Single place React learns the width tier -- must match workspace.css's
  // widest breakpoint character-for-character. Drives both the details
  // panel's pane-or-drawer variant and the toolbar's Details chip visibility.
  const isWide = useMediaQuery('(min-width: 1280px)');

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

  return (
    <div className="ws-root" data-rail-side={state.railSide}>
      <Toolbar
        showDetailsChip={!isWide}
        detailsOpen={state.detailOverlay}
        onToggleDetails={() =>
          dispatch({ type: 'SET_DETAIL_OVERLAY', payload: !state.detailOverlay })
        }
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
      <StatusBar />
    </div>
  );
}

export default WorkspaceShell;
