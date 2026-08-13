import Toolbar from './Toolbar';
import CatalogRail from './CatalogRail';
import TreePane from './TreePane';
import DetailsPanel from './DetailsPanel';
import StatusBar from './StatusBar';

function WorkspaceShell() {
  return (
    <div className="ws-root">
      <Toolbar />
      <div className="ws-grid">
        <CatalogRail />
        <TreePane />
        <DetailsPanel />
      </div>
      <StatusBar />
    </div>
  );
}

export default WorkspaceShell;
