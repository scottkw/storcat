import { useAppContext } from '../../contexts/AppContext';
import { useModalBehavior } from '../../hooks/useModalBehavior';
import { setDensitySetting } from '../../settingsStore';
import SegmentedControl from './settings/SegmentedControl';

export interface SettingsDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

// Always mounted by WorkspaceShell and returns null when closed -- it must
// not be conditionally mounted, because the shared useModalBehavior hook
// below observes the isOpen: true -> false transition to release scroll
// lock and restore focus (same load-bearing contract CommandPalette.tsx
// documents).
function SettingsDialog({ isOpen, onClose }: SettingsDialogProps) {
  const { state, dispatch } = useAppContext();

  // Focus trap, Escape-to-close, scroll lock, and focus restore all arrive
  // through the shared hook -- this dialog implements none of the four
  // itself (PLT-07 standing constraint). No initialFocusRef: the hook falls
  // back to the panel's first focusable element.
  const { containerRef } = useModalBehavior({ isOpen, onClose });

  if (!isOpen) return null;

  return (
    <div className="ws-settings-scrim" onClick={onClose}>
      <div
        ref={containerRef}
        className="ws-settings-panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ws-settings-title"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="ws-settings-header">
          <span id="ws-settings-title" className="ws-settings-title">
            Settings
          </span>
          <span className="ws-settings-hint mono">⌘,</span>
          <button type="button" className="ws-settings-close-x" aria-label="Close" onClick={onClose}>
            ×
          </button>
        </div>

        <div className="ws-settings-body">
          <div className="ws-settings-section">
            <div className="ws-settings-section-label">Layout</div>
            <div className="ws-settings-row">
              <div className="ws-settings-row-text">
                <span className="ws-settings-row-label">Row density</span>
                <span className="ws-settings-row-note">How tight the tree and lists pack</span>
              </div>
              <SegmentedControl
                options={[
                  { value: 'Compact', label: 'Compact' },
                  { value: 'Comfortable', label: 'Comfortable' },
                ]}
                value={state.density}
                ariaLabel="Row density"
                onChange={(next) => {
                  dispatch({ type: 'SET_DENSITY', payload: next });
                  setDensitySetting(next);
                }}
              />
            </div>
          </div>
        </div>

        <div className="ws-settings-footer">
          <span className="ws-settings-status mono" />
          <button type="button" className="ws-settings-close-btn" onClick={onClose}>
            Close settings
          </button>
        </div>
      </div>
    </div>
  );
}

export default SettingsDialog;
