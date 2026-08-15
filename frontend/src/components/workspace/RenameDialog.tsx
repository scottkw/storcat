import { useEffect, useRef, useState } from 'react';
import { wailsAPI } from '../../services/wailsAPI';
import { models } from '../../../wailsjs/go/models';
import DialogShell from './DialogShell';

export interface RenameDialogProps {
  isOpen: boolean;
  onClose: () => void;
  catalog: models.CatalogMetadata;
  catalogDir: string | null;
  onRenamed: (newTitle: string) => void;
}

// ACT-02's rename surface, built on the shared DialogShell.
function RenameDialog({ isOpen, onClose, catalog, catalogDir, onRenamed }: RenameDialogProps) {
  const [value, setValue] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);

  // Seeded in an isOpen-keyed effect, not a useState initializer -- this
  // component is always mounted (DialogShell's contract), so an initializer
  // would run once at app start against a catalog that has not been chosen
  // yet and would go stale across every subsequent open. This is exactly
  // the bug Phase 26 hit and fixed (26-03). Focus + select happen in the
  // same effect so typing immediately replaces the pre-filled title.
  //
  // The DOM value is set imperatively before .select() is called, ahead of
  // the setValue() re-render: this is a controlled input, and calling
  // .select() against the input's still-stale (pre-render) DOM value
  // selects nothing, which the browser then collapses to a trailing caret
  // the instant React's own re-render lands the real value. Pre-syncing the
  // DOM value first means React's reconciliation finds node.value already
  // equal to the next value and skips touching it, so the selection this
  // effect makes is never clobbered by React's own commit.
  useEffect(() => {
    if (!isOpen) return;
    setValue(catalog.title);
    setError(null);
    setBusy(false);
    const input = inputRef.current;
    if (input) {
      input.value = catalog.title;
      input.focus();
      input.select();
    }
  }, [isOpen, catalog]);

  async function submit() {
    if (busy) return;
    setError(null);
    // catalogDir is required for the Go side's containment check (T-27-02)
    // -- fail closed here rather than sending an empty directory the
    // backend would just reject anyway, the same guard DetailsPanel's
    // Footer already applies to its two actions.
    if (!catalogDir) {
      setError('No catalog directory configured.');
      return;
    }
    setBusy(true);
    const trimmed = value.trim();
    const result = await wailsAPI.renameCatalog(catalog.path, catalogDir, trimmed);
    if (result.success) {
      onRenamed(trimmed);
      onClose();
      return;
    }
    setError(`Couldn't rename this catalog: ${result.error}.`);
    setBusy(false);
  }

  return (
    <DialogShell
      isOpen={isOpen}
      onClose={onClose}
      titleId="ws-rename-title"
      title="Rename catalog"
      initialFocusRef={inputRef}
      footer={
        <>
          <button type="button" className="ws-dialog-btn" onClick={onClose}>
            Keep original title
          </button>
          <button
            type="button"
            className="ws-dialog-btn ws-dialog-btn-primary"
            disabled={!value.trim() || busy}
            onClick={submit}
          >
            Rename catalog
          </button>
        </>
      }
    >
      <div className="ws-dialog-label">Title</div>
      <input
        ref={inputRef}
        className="ws-rename-input"
        value={value}
        onChange={(event) => setValue(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter') {
            event.preventDefault();
            submit();
          }
        }}
      />
      {error && <div className="ws-dialog-error">{error}</div>}
    </DialogShell>
  );
}

export default RenameDialog;
