import Toolbar from './Toolbar';
import CatalogRail from './CatalogRail';
import TreePane from './TreePane';
import DetailsPanel from './DetailsPanel';
import StatusBar from './StatusBar';
import { useMediaQuery } from '../../hooks/useMediaQuery';

function WorkspaceShell() {
  // Single place React learns the width tier -- must match workspace.css's
  // widest breakpoint character-for-character. Plan 22-07 uses isWide to
  // choose the details panel's pane-or-drawer variant; for now this task
  // only renders it when wide, to keep the binding used under noUnusedLocals.
  const isWide = useMediaQuery('(min-width: 1280px)');

  return (
    <div className="ws-root">
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
