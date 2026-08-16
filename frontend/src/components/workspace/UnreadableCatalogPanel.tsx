import { useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import RescanDialog from './rescan/RescanDialog';
import DeleteConfirmDialog from './DeleteConfirmDialog';

// A catalog can arrive broken by either of two independent routes: the rail
// listing's own parse-status check (BrowseCatalogs' detectParseError, which
// formats a JSON syntax error as "byte N: reason" and carries a real byte
// offset), or the tree load itself failing (LoadCatalogFlat -- covers a file
// deleted between listing and click, a permission error, or valid JSON that
// isn't a catalog, none of which carry a byte offset). The listing error is
// preferred when both exist because it is the richer diagnostic.
function pickRawError(catalog: { parseError: string } | undefined, treeMessage: string): string {
  const listingError = catalog?.parseError ?? '';
  return listingError !== '' ? listingError : treeMessage;
}

const BYTE_OFFSET_RE = /^byte (\d+): (.*)$/s;
const READ_FAILURE_RE = /no such file or directory|permission denied|failed to read catalog file/i;

function UnreadableCatalogPanel() {
  const { state, dispatch } = useAppContext();
  const catalog = state.catalogs.find((c) => c.path === state.currentCatalogId);
  const treeMessage = state.tree.status === 'error' ? state.tree.message : '';
  const rawError = pickRawError(catalog, treeMessage);

  // Hooks declared unconditionally, ahead of the early return below, so
  // their call order never varies across renders -- same discipline
  // DetailsPanel's own action-error slot follows.
  const [rescanOpen, setRescanOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [openBusy, setOpenBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const isScanningNow = state.scan.status === 'counting' || state.scan.status === 'scanning';

  if (rawError === '') return null;

  const byteMatch = BYTE_OFFSET_RE.exec(rawError);
  // A byte offset means the string is already the "byte N: reason" shape
  // detectParseError produces -- the reason is everything after that prefix.
  // No offset means the failure never got that far: a missing file, a
  // permission error, or (rarely) a structural check after a successful
  // parse -- "Failed at" is an em dash in all of those cases, per FPA-23-05-A.
  const failedAt = byteMatch ? `byte ${byteMatch[1]}` : '—';
  const reason = byteMatch ? byteMatch[2] : rawError.replace(/^load catalog for flatten: /, '');
  const isReadFailure = !byteMatch && READ_FAILURE_RE.test(rawError);
  const parser = isReadFailure ? 'not reached' : 'v2 object / v1 array';

  const metaRows: [string, string][] = [
    ['File', catalog?.filename ?? '—'],
    ['Failed at', failedAt],
    ['Reason', reason],
    ['Parser', parser],
  ];

  // Reuses DetailsPanel's Footer.handleOpenHtml logic verbatim (same
  // getCatalogHtmlPath + openExternal call pair) -- an unreadable catalog
  // can still have a valid .html companion since only the JSON failed to
  // parse. No second path derivation.
  async function handleOpenHtml() {
    if (!catalog || openBusy) return;
    setOpenBusy(true);
    setActionError(null);
    if (!state.catalogDir) {
      setActionError('No catalog directory configured.');
      setOpenBusy(false);
      return;
    }
    const htmlPathResult = await wailsAPI.getCatalogHtmlPath(catalog.path, state.catalogDir);
    if (htmlPathResult.success) {
      const openResult = await wailsAPI.openExternal(htmlPathResult.htmlPath, state.catalogDir);
      if (!openResult.success) {
        setActionError(openResult.error);
      }
    } else {
      setActionError(htmlPathResult.error);
    }
    setOpenBusy(false);
  }

  return (
    <div
      className="pane-scroll"
      style={{ padding: '24px 22px', display: 'flex', flexDirection: 'column', gap: 18 }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 11 }}>
        <span
          aria-hidden="true"
          style={{
            width: 22,
            height: 22,
            borderRadius: 6,
            background: 'rgba(229,83,75,.16)',
            color: '#e5534b',
            fontSize: 13,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flex: 'none',
          }}
        >
          !
        </span>
        <span style={{ fontSize: 15, fontWeight: 600 }}>This catalog can&rsquo;t be read</span>
      </div>

      <span style={{ fontSize: 12.5, lineHeight: 1.6, color: 'var(--dm)', maxWidth: 520 }}>
        This catalog&rsquo;s JSON couldn&rsquo;t be parsed. The .html copy may still open, and the volume
        can be re-scanned later to rebuild it.
      </span>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 1, maxWidth: 520 }}>
        {metaRows.map(([label, value]) => (
          <div
            key={label}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              gap: 12,
              padding: '7px 0',
              borderBottom: '1px solid var(--l2)',
              fontSize: 11.5,
            }}
          >
            <span style={{ color: 'var(--dm)' }}>{label}</span>
            <span className="mono">{value}</span>
          </div>
        ))}
      </div>

      {/* Verbatim, unedited -- never truncated, paraphrased or summarised.
          No max height, no internal scroll: a single-line parser message is
          the expected shape (23-UI-SPEC E4 overflow). */}
      <div
        className="mono"
        style={{
          padding: 12,
          borderRadius: 9,
          background: 'var(--ch)',
          fontSize: 11.5,
          color: '#e5534b',
          wordBreak: 'break-all',
        }}
      >
        {rawError}
      </div>

      {catalog && (
        <div className="ws-unreadable-actions">
          {/* Button 1 -- "Re-scan volume" (no ellipsis, locked): deliberately
              distinct from the details-panel footer's own longer label
              (which carries a trailing ellipsis) and the catalog-actions
              menu's disambiguated "& diff" variant -- all three are locked
              labels, and the differences are intentional, not drift. Opens
              RescanDialog at the pick-volume step with the old-tree flag
              false: the JSON here doesn't parse, so there is no old tree to
              diff against and step 3 renders the reduced Variant B
              (scan-complete summary, overwrite/keep-both only, no stat
              grid, no diff list). */}
          <button
            type="button"
            className="ws-unreadable-action ws-unreadable-action-primary"
            disabled={isScanningNow}
            aria-disabled={isScanningNow}
            title={isScanningNow ? 'A scan is already running — open it from the status bar.' : undefined}
            onClick={() => setRescanOpen(true)}
          >
            Re-scan volume
          </button>
          {/* Omitted entirely (not greyed) when there's no .html companion --
              the same "no button whose only outcome is inapplicable" rule
              the delete dialog's own conditional .html row follows. */}
          {catalog.hasHtml && (
            <button
              type="button"
              className="ws-unreadable-action"
              disabled={openBusy}
              onClick={handleOpenHtml}
            >
              Open the .html instead
            </button>
          )}
          {/* Opens the existing Phase 27 delete confirmation, unchanged --
              there is no library-membership concept to toggle; membership IS
              the file living in the configured catalog directory, so this is
              Phase 27's delete-to-Trash, reused, not re-implemented. */}
          <button
            type="button"
            className="ws-unreadable-action ws-unreadable-action-danger"
            onClick={() => setDeleteOpen(true)}
          >
            Remove from library
          </button>
        </div>
      )}

      {actionError && (
        <span style={{ fontSize: 11, color: 'var(--danger)', lineHeight: 1.4 }}>{actionError}</span>
      )}

      {catalog && rescanOpen && (
        <RescanDialog
          catalog={catalog}
          catalogDir={state.catalogDir}
          oldTreeAvailable={false}
          onError={setActionError}
          onClose={() => setRescanOpen(false)}
        />
      )}

      {catalog && (
        <DeleteConfirmDialog
          isOpen={deleteOpen}
          onClose={() => setDeleteOpen(false)}
          catalog={catalog}
          catalogDir={state.catalogDir}
          onDeleted={() => {
            // This panel only renders for the current catalog selection, so
            // the deleted catalog is always the current one -- same
            // fall-back-to-placeholder + rail-refresh pair CatalogActions'
            // own onDeleted already uses.
            dispatch({ type: 'CLEAR_CURRENT_CATALOG' });
            dispatch({ type: 'REQUEST_RAIL_REFRESH' });
          }}
        />
      )}
    </div>
  );
}

export default UnreadableCatalogPanel;
