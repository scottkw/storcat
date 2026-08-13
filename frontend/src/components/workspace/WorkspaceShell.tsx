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
  const { state } = useAppContext();

  // Single place React learns the width tier -- must match workspace.css's
  // widest breakpoint character-for-character. Plan 22-07 uses isWide to
  // choose the details panel's pane-or-drawer variant; for now this task
  // only renders it when wide, to keep the binding used under noUnusedLocals.
  const isWide = useMediaQuery('(min-width: 1280px)');

  // Density lives in the reducer; re-apply the token layer whenever it
  // changes so the reducer field is not inert. Theme isn't reducer state
  // yet (Phase 26), so it's re-read here for now.
  useEffect(() => {
    applyTokens(readPersistedPrefs().theme, state.density);
  }, [state.density]);

  return (
    <div className="ws-root" data-rail-side={state.railSide}>
      <Toolbar />
      <div className="ws-grid">
        <CatalogRail />
        <TreePane />
        {isWide && <DetailsPanel />}
      </div>
      <StatusBar />
    </div>
  );
}

export default WorkspaceShell;
