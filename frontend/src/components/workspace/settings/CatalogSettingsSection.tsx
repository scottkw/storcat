import { useAppContext } from '../../../contexts/AppContext';
import { wailsAPI } from '../../../services/wailsAPI';
import { setCatalogDirectorySetting } from '../../../settingsStore';

// The Catalogs section (SET-04): the catalog-directory row today, joined by
// the default-filename-root row in plan 26-03 Task 2.
function CatalogSettingsSection() {
  const { state, dispatch } = useAppContext();

  // Shared by both the chip and the "Change…" link -- same guard shape as
  // CatalogRail.tsx's handleChooseDirectory: a cancelled picker is a no-op,
  // and re-choosing the already-configured directory is a no-op too (the
  // reducer's own SET_CATALOG_DIR guard makes this belt-and-suspenders, not
  // load-bearing on its own).
  async function handleChooseDirectory() {
    const result = await wailsAPI.selectDirectory();
    if (!result.success || !result.path) return;
    if (result.path === state.catalogDir) return;
    setCatalogDirectorySetting(result.path);
    dispatch({ type: 'SET_CATALOG_DIR', payload: result.path });
  }

  return (
    <div className="ws-settings-section">
      <div className="ws-settings-section-label">Catalogs</div>
      <div className="ws-settings-row">
        <div className="ws-settings-row-text">
          <span className="ws-settings-row-label">Catalog directory</span>
        </div>
        <div className="ws-settings-value-col">
          <button
            type="button"
            className="ws-settings-dir-chip mono"
            onClick={handleChooseDirectory}
            aria-label="Choose catalog directory"
          >
            {state.catalogDir ?? 'No catalog directory set'}
          </button>
          <button type="button" className="ws-settings-change-link" onClick={handleChooseDirectory}>
            Change…
          </button>
        </div>
      </div>
    </div>
  );
}

export default CatalogSettingsSection;
