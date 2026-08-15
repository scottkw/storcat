import { useAppContext } from '../../../contexts/AppContext';
import { wailsAPI } from '../../../services/wailsAPI';
import {
  setCatalogDirectorySetting,
  setDefaultFilenameRootSetting,
  setWriteHtmlSetting,
  setCopyToSecondarySetting,
  setSecondaryDirectorySetting,
  setWatchDirectorySetting,
} from '../../../settingsStore';
import { ToggleRow } from '../create/OptionsToggles';

// The Catalogs section (SET-04): the catalog-directory and
// default-filename-root rows (plan 26-03), joined by the four locked-order
// toggles (plan 26-05) -- all rendered through the one shared ToggleRow.
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

  // Reproduces OptionsToggles.handleToggleSecondary exactly -- both surfaces
  // read and write the same stored secondary-directory value (SET-04).
  // Turning the toggle off never clears the stored path (the same
  // idempotency rule CRT-05 established for the create form's own toggle).
  async function handleToggleCopyToSecondary() {
    if (state.settings.copyToSecondary) {
      dispatch({ type: 'SET_SETTINGS', payload: { copyToSecondary: false } });
      setCopyToSecondarySetting(false);
      return;
    }
    if (state.settings.secondaryDirectory) {
      // A path is already persisted -- reuse it with no dialog.
      dispatch({ type: 'SET_SETTINGS', payload: { copyToSecondary: true } });
      setCopyToSecondarySetting(true);
      return;
    }
    const result = await wailsAPI.selectDirectory();
    if (!result.success || !result.path) return; // cancelled -- a declined choice, not an error
    setSecondaryDirectorySetting(result.path);
    dispatch({ type: 'SET_SETTINGS', payload: { secondaryDirectory: result.path, copyToSecondary: true } });
    setCopyToSecondarySetting(true);
  }

  // Mirrors handleEditSecondaryPath -- re-pick the folder when its note
  // (already showing a stored path) is clicked.
  async function handleEditSecondaryPath(event: React.MouseEvent) {
    event.stopPropagation();
    if (!state.settings.secondaryDirectory) return;
    const result = await wailsAPI.selectDirectory();
    if (!result.success || !result.path) return;
    setSecondaryDirectorySetting(result.path);
    dispatch({ type: 'SET_SETTINGS', payload: { secondaryDirectory: result.path } });
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
      <div className="ws-settings-toggle-list">
        <ToggleRow
          checked={state.settings.writeHtml}
          label="Write HTML alongside JSON"
          note="every catalog gets a matching .html"
          noteMono={false}
          disabled={false}
          onToggle={() => {
            const next = !state.settings.writeHtml;
            dispatch({ type: 'SET_SETTINGS', payload: { writeHtml: next } });
            setWriteHtmlSetting(next);
          }}
        />
        <ToggleRow
          checked={state.settings.copyToSecondary}
          label="Copy catalogs to a secondary location"
          note={state.settings.secondaryDirectory || 'Choose a folder when enabled'}
          noteClassName="ws-create-note-path"
          noteMono={!!state.settings.secondaryDirectory}
          disabled={false}
          onToggle={handleToggleCopyToSecondary}
          onNoteClick={state.settings.secondaryDirectory ? handleEditSecondaryPath : undefined}
        />
        <ToggleRow
          checked={state.settings.watchDirectory}
          label="Watch catalog directory for changes"
          note="refresh the rail automatically"
          noteMono={false}
          disabled={false}
          onToggle={() => {
            const next = !state.settings.watchDirectory;
            dispatch({ type: 'SET_SETTINGS', payload: { watchDirectory: next } });
            setWatchDirectorySetting(next);
          }}
        />
      </div>
    </div>
  );
}

export default CatalogSettingsSection;
