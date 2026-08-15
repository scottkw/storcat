import { useAppContext } from '../../../contexts/AppContext';
import { wailsAPI } from '../../../services/wailsAPI';
import { setCatalogDirectorySetting, setDefaultFilenameRootSetting } from '../../../settingsStore';

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

  // Every keystroke commits immediately (SET-05, no save step) -- the
  // whitespace strip mirrors the design demo's transform. No validation
  // error, no disabled state, no max length: an empty value is valid and
  // simply falls back to the create form's own source-derived placeholder.
  function handleRootChange(event: React.ChangeEvent<HTMLInputElement>) {
    const stripped = event.target.value.replace(/\s/g, '');
    dispatch({ type: 'SET_SETTINGS', payload: { defaultFilenameRoot: stripped } });
    setDefaultFilenameRootSetting(stripped);
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
      <div className="ws-settings-row">
        <div className="ws-settings-row-text">
          <span className="ws-settings-row-label">Default filename root</span>
          <span className="ws-settings-row-note">Pre-filled for every new catalog</span>
        </div>
        <input
          className="ws-settings-root-input mono"
          placeholder="catalog"
          value={state.settings.defaultFilenameRoot}
          onChange={handleRootChange}
        />
      </div>
    </div>
  );
}

export default CatalogSettingsSection;
