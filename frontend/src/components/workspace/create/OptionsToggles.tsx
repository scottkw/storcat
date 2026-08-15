import { wailsAPI } from '../../../services/wailsAPI';
import { SECONDARY_DIR_KEY, setSecondaryDirectorySetting } from '../../../settingsStore';

// Re-exported so CreateSlideOver's existing import keeps compiling --
// settingsStore.ts is now the one home for this storage-key literal.
export const SECONDARY_DIR_STORAGE_KEY = SECONDARY_DIR_KEY;

export interface OptionsToggleValues {
  writeHTML: boolean;
  copyToSecondary: boolean;
  includeHidden: boolean;
}

export interface OptionsTogglesProps {
  values: OptionsToggleValues;
  onValuesChange: (values: OptionsToggleValues) => void;
  // The persisted secondary-copy path, independent of whether the toggle
  // is currently on -- toggling off never clears it (CRT-05 idempotency).
  secondaryDir: string;
  onSecondaryDirChange: (dir: string) => void;
  disabled: boolean;
  // The resolved filename root (CreateForm's effective value), for the
  // write-HTML toggle's dynamic `{root}.html` note.
  effectiveRoot: string;
}

interface ToggleRowProps {
  checked: boolean;
  label: string;
  note?: string;
  noteClassName?: string;
  disabled: boolean;
  onToggle: () => void;
  onNoteClick?: (event: React.MouseEvent) => void;
}

// This project's first toggle-switch control -- plain markup and
// create-prefixed styling, no package added. A row is always definitively
// on or off; there is no unset/indeterminate visual state.
function ToggleRow({ checked, label, note, noteClassName, disabled, onToggle, onNoteClick }: ToggleRowProps) {
  return (
    <div
      className={`ws-create-toggle-row${disabled ? ' ws-create-toggle-row-disabled' : ''}`}
      role="switch"
      aria-checked={checked}
      aria-disabled={disabled || undefined}
      tabIndex={disabled ? -1 : 0}
      onClick={() => {
        if (!disabled) onToggle();
      }}
      onKeyDown={(event) => {
        if (disabled) return;
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          onToggle();
        }
      }}
    >
      <span className={`ws-create-toggle-track${checked ? ' ws-create-toggle-track-on' : ''}`} aria-hidden="true">
        <span className="ws-create-toggle-knob" />
      </span>
      <div className="ws-create-toggle-text">
        <span className="ws-create-toggle-label">{label}</span>
        {note !== undefined && (
          <span
            className={`ws-create-toggle-note mono${noteClassName ? ` ${noteClassName}` : ''}`}
            onClick={onNoteClick}
          >
            {note}
          </span>
        )}
      </div>
    </div>
  );
}

// Three rows: write-HTML (default on), copy-to-secondary-location (default
// off -- diverges from the design demo's on, which existed only to keep the
// demo's mock preview populated; a fresh install has no secondary location
// configured, so defaulting it on would force a folder dialog on the very
// first open), include-hidden-files (default off). Disabled visibly
// (dimmed, aria-disabled, not-allowed cursor) while a scan is running --
// never silently inert.
function OptionsToggles({
  values,
  onValuesChange,
  secondaryDir,
  onSecondaryDirChange,
  disabled,
  effectiveRoot,
}: OptionsTogglesProps) {
  async function handleToggleSecondary() {
    if (values.copyToSecondary) {
      onValuesChange({ ...values, copyToSecondary: false });
      return;
    }
    if (secondaryDir) {
      // A path is already persisted -- reuse it with no dialog.
      onValuesChange({ ...values, copyToSecondary: true });
      return;
    }
    const result = await wailsAPI.selectDirectory();
    if (!result.success || !result.path) return; // cancelled -- a declined choice, not an error
    setSecondaryDirectorySetting(result.path);
    onSecondaryDirChange(result.path);
    onValuesChange({ ...values, copyToSecondary: true });
  }

  async function handleEditSecondaryPath(event: React.MouseEvent) {
    event.stopPropagation();
    if (disabled || !secondaryDir) return;
    const result = await wailsAPI.selectDirectory();
    if (!result.success || !result.path) return;
    setSecondaryDirectorySetting(result.path);
    onSecondaryDirChange(result.path);
  }

  return (
    <div className="ws-create-field">
      <label className="ws-create-label">Options</label>
      <div className="ws-create-toggle-list">
        <ToggleRow
          checked={values.writeHTML}
          label="Also write HTML catalog"
          note={`${effectiveRoot}.html`}
          disabled={disabled}
          onToggle={() => onValuesChange({ ...values, writeHTML: !values.writeHTML })}
        />
        <ToggleRow
          checked={values.copyToSecondary}
          label="Copy both files to secondary location"
          note={secondaryDir || 'Choose a folder when enabled'}
          noteClassName="ws-create-note-path"
          disabled={disabled}
          onToggle={handleToggleSecondary}
          onNoteClick={secondaryDir ? handleEditSecondaryPath : undefined}
        />
        <ToggleRow
          checked={values.includeHidden}
          label="Include hidden files"
          disabled={disabled}
          onToggle={() => onValuesChange({ ...values, includeHidden: !values.includeHidden })}
        />
      </div>
    </div>
  );
}

export default OptionsToggles;
