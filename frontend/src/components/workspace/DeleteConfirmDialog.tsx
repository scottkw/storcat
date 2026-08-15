import { useEffect, useRef, useState } from 'react';
import { wailsAPI } from '../../services/wailsAPI';
import { models } from '../../../wailsjs/go/models';
import DialogShell from './DialogShell';

export interface DeleteConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  catalog: models.CatalogMetadata;
  catalogDir: string | null;
  onDeleted: () => void;
}

// ACT-04/ACT-05's confirm and error sub-states, in one mounted dialog on the
// shared DialogShell -- never a second overlay. This is the phase's only
// genuinely destructive surface: every string below is verbatim from
// 27-UI-SPEC.md's Delete-Confirmation Dialog section, and none of them may
// imply any other way to remove a file exists beyond the OS Trash, on any
// path, ever.
function DeleteConfirmDialog({ isOpen, onClose, catalog, catalogDir, onDeleted }: DeleteConfirmDialogProps) {
  const [deleteHtml, setDeleteHtml] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const keepRef = useRef<HTMLButtonElement | null>(null);

  // Re-seeded on every open, the same isOpen-keyed shape RenameDialog uses --
  // this component is always mounted (DialogShell's contract), so a
  // useState initializer alone would carry a previous catalog's checkbox/
  // error/busy state into the next open.
  useEffect(() => {
    if (!isOpen) return;
    setDeleteHtml(true);
    setBusy(false);
    setError(null);
  }, [isOpen, catalog]);

  // Display-only -- the .html actually trashed is derived independently in
  // Go by App.DeleteCatalog, which never accepts an HTML path from the
  // renderer (T-27-18), so a mismatch here could mislead but could never
  // redirect the deletion.
  const htmlPath = catalog.path.replace(/\.json$/, '.html');
  const primaryLabel = catalog.hasHtml && deleteHtml ? 'Move both to Trash' : 'Move to Trash';

  // Shared by the confirm primary and the error state's retry primary --
  // retry re-invokes with the same path and the same checkbox state and
  // needs no bookkeeping about what already succeeded: TrashPaths (27-03)
  // skips a path that is no longer on disk.
  async function submit() {
    if (busy) return;
    setError(null);
    if (!catalogDir) {
      setError('No catalog directory configured.');
      return;
    }
    setBusy(true);
    const result = await wailsAPI.deleteCatalog(catalog.path, catalogDir, deleteHtml);
    if (result.success) {
      onDeleted();
      onClose();
      return;
    }
    setError(result.error);
    setBusy(false);
  }

  return (
    <DialogShell
      isOpen={isOpen}
      onClose={onClose}
      titleId="ws-delete-title"
      title="Delete catalog"
      initialFocusRef={keepRef}
      footer={
        error ? (
          <>
            <button type="button" className="ws-dialog-btn" ref={keepRef} disabled={busy} onClick={onClose}>
              Keep catalog
            </button>
            <button
              type="button"
              className="ws-dialog-btn ws-dialog-btn-danger"
              disabled={busy}
              style={{ opacity: busy ? 0.7 : 1 }}
              onClick={submit}
            >
              Try moving to Trash again
            </button>
          </>
        ) : (
          <>
            <button type="button" className="ws-dialog-btn" ref={keepRef} disabled={busy} onClick={onClose}>
              Keep catalog
            </button>
            <button
              type="button"
              className="ws-dialog-btn ws-dialog-btn-danger"
              disabled={busy}
              style={{ opacity: busy ? 0.7 : 1 }}
              onClick={submit}
            >
              {primaryLabel}
            </button>
          </>
        )
      }
    >
      <p className="ws-delete-lead">
        This moves the file(s) below to the Trash — you can restore them from the Trash until it's emptied.
      </p>
      <div className="ws-delete-row">
        <span className="ws-delete-path-label">Catalog (.json)</span>
        <span className="ws-delete-path-box mono">{catalog.path}</span>
      </div>
      {catalog.hasHtml && (
        <div className="ws-delete-row">
          <span className="ws-delete-path-label">HTML (.html)</span>
          <span className="ws-delete-path-box mono">{htmlPath}</span>
        </div>
      )}
      {catalog.hasHtml && (
        <label className="ws-delete-check">
          <input type="checkbox" checked={deleteHtml} onChange={(event) => setDeleteHtml(event.target.checked)} />
          Also delete the matching .html
        </label>
      )}
      {error && (
        <div className="ws-dialog-error">
          Couldn't move {catalog.title} to the Trash: {error}.
        </div>
      )}
    </DialogShell>
  );
}

export default DeleteConfirmDialog;
