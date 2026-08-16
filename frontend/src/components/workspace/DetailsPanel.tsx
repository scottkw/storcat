import type { CSSProperties } from 'react';
import { useLayoutEffect, useRef, useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { formatBytes, formatCount, formatDate } from '../../lib/format';
import { wailsAPI } from '../../services/wailsAPI';
import { Environment } from '../../../wailsjs/runtime/runtime';
import { models } from '../../../wailsjs/go/models';
import Menu, { MenuItemSpec } from './Menu';
import RenameDialog from './RenameDialog';
import DeleteConfirmDialog from './DeleteConfirmDialog';
import RescanDialog from './rescan/RescanDialog';

export interface DetailsPanelProps {
  variant?: 'pane' | 'drawer';
}

type MetaRow = [label: string, value: string];

// Shared by both populated states -- label left (--dm), value right (mono,
// ellipsized via the existing ws-meta-value contract so a long value can
// never widen the panel).
function MetaRows({ rows }: { rows: MetaRow[] }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
      {rows.map(([label, value]) => (
        <div
          key={label}
          className="ws-meta-row"
          style={{ display: 'flex', justifyContent: 'space-between', gap: 12, borderBottom: '1px solid var(--l2)' }}
        >
          <span style={{ fontSize: 11.5, color: 'var(--dm)' }}>{label}</span>
          <span className="ws-meta-value mono" style={{ fontSize: 11.5, textAlign: 'right' }}>
            {value}
          </span>
        </div>
      ))}
    </div>
  );
}

// Renders the ⋯ trigger, the anchored menu it opens (ACT-01), the rename
// dialog (ACT-02), the immediate duplicate action (ACT-03), and the delete
// confirmation dialog (ACT-04/ACT-05) -- every item does its real work.
function CatalogActions({
  catalog,
  catalogDir,
  onError,
}: {
  catalog: models.CatalogMetadata;
  catalogDir: string | null;
  onError: (message: string | null) => void;
}) {
  const { state, dispatch } = useAppContext();
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  // No busy state and no disabled item while this runs -- a fast local file
  // copy, this app has no spinners, and 27-UI-SPEC.md's E1 table marks the
  // menu's loading consideration not applicable.
  async function duplicateCatalogAction() {
    // catalogDir is required for the Go side's containment check -- fail
    // closed here rather than sending an empty directory the backend would
    // just reject anyway, the same guard Footer's two actions already use.
    if (!catalogDir) {
      onError('No catalog directory configured.');
      return;
    }
    const result = await wailsAPI.duplicateCatalog(catalog.path, catalogDir);
    if (!result.success) {
      onError(result.error);
      return;
    }
    onError(null);
    // 27-RAIL-FIX: this duplicate already succeeded on disk -- re-trigger
    // the rail's one authoritative listing so the new row appears without
    // requiring watching to be on (it defaults to off).
    dispatch({ type: 'REQUEST_RAIL_REFRESH' });
  }

  const items: MenuItemSpec[] = [
    {
      id: 'rename',
      label: 'Rename catalog…',
      onSelect: () => {
        setMenuOpen(false);
        setRenameOpen(true);
      },
    },
    {
      id: 'duplicate',
      label: 'Duplicate catalog',
      onSelect: () => {
        setMenuOpen(false);
        duplicateCatalogAction();
      },
    },
    {
      id: 'delete',
      label: 'Delete catalog…',
      danger: true,
      dividerBefore: true,
      onSelect: () => {
        setMenuOpen(false);
        setDeleteOpen(true);
      },
    },
  ];

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        className="ws-details-overflow"
        aria-label="Catalog actions"
        aria-haspopup="menu"
        aria-expanded={menuOpen}
        aria-controls={menuOpen ? 'ws-catalog-actions-menu' : undefined}
        style={{
          marginLeft: 'auto',
          flex: 'none',
          width: 22,
          height: 22,
          borderRadius: 6,
          border: '1px solid var(--l)',
          background: 'transparent',
          color: 'var(--dm)',
          fontSize: 13,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          cursor: 'pointer',
        }}
        onClick={() => setMenuOpen((open) => !open)}
      >
        <span aria-hidden="true">⋯</span>
      </button>
      {menuOpen && <Menu
          id="ws-catalog-actions-menu"
          ariaLabel="Catalog actions"
          isOpen
          triggerRef={triggerRef}
          onClose={() => setMenuOpen(false)}
          items={items}
        />}
      <RenameDialog isOpen={renameOpen}
        onClose={() => setRenameOpen(false)}
        catalog={catalog}
        catalogDir={catalogDir}
        onRenamed={(newTitle) => {
          // Optimistic -- no wait for a rail re-list, consistent with this
          // app's no-spinners rule. The rail row and the details panel both
          // show the new title in the same frame the dialog closes.
          dispatch({
            type: 'SET_CATALOGS',
            payload: state.catalogs.map((c) => (c.path === catalog.path ? { ...c, title: newTitle } : c)),
          });
        }}
      />
      <DeleteConfirmDialog isOpen={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        catalog={catalog}
        catalogDir={catalogDir}
        onDeleted={() => {
          // If the deleted catalog was the current selection, fall back to
          // the details panel's existing "nothing selected" placeholder --
          // no new empty state.
          if (catalog.path === state.currentCatalogId) {
            dispatch({ type: 'CLEAR_CURRENT_CATALOG' });
          }
          // 27-RAIL-FIX: this delete already succeeded on disk -- re-trigger
          // the rail's one authoritative listing (same browseCatalogs call
          // catalogs:changed re-triggers) so the row disappears without
          // requiring watching to be on (it defaults to off). Reuses the
          // single refresh path -- not a second way to compute the rail's
          // contents.
          dispatch({ type: 'REQUEST_RAIL_REFRESH' });
        }}
      />
    </>
  );
}

function revealButtonLabel(platform: string | null): string {
  if (platform === 'darwin') return 'Reveal JSON in Finder';
  if (platform === 'windows') return 'Reveal JSON in Explorer';
  return 'Reveal JSON in file manager';
}

const buttonBase: CSSProperties = {
  height: 30,
  borderRadius: 7,
  fontSize: 12.5,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  cursor: 'pointer',
};

// Three actions as of Phase 28 (TREE-08 + ACT-06/ACT-08) -- "Re-scan
// volume…" (28-01) fills the stub this component left for it since 27-04.
// All three operate on the catalog's own file path regardless of whether a
// node is also selected -- none ever receives a free-form or user-typed
// path.
//
// error/onError are hoisted from DetailsPanel (27-04) rather than owned
// locally -- 27-UI-SPEC.md puts a duplicate/delete failure in this same
// error slot, so CatalogActions' menu-item placeholders and this footer's
// actions all share one error surface, including RescanDialog's own
// failures.
function Footer({
  catalog,
  catalogDir,
  error,
  onError,
}: {
  catalog: models.CatalogMetadata;
  catalogDir: string | null;
  error: string | null;
  onError: (message: string | null) => void;
}) {
  const { state } = useAppContext();
  const [platform, setPlatform] = useState<string | null>(null);
  const [openBusy, setOpenBusy] = useState(false);
  const [revealBusy, setRevealBusy] = useState(false);
  // RescanDialog is conditionally mounted by this component (28-UI-SPEC.md's
  // mount pattern), not always-mounted like CreateSlideOver/SettingsDialog
  // -- rendered fresh on every open, no isOpen prop of its own.
  const [rescanOpen, setRescanOpen] = useState(false);
  const isScanningNow = state.scan.status === 'counting' || state.scan.status === 'scanning';

  // Same deferred-call-with-catch pattern Toolbar.tsx already established:
  // Environment() throws synchronously outside a real Wails webview, so
  // deferring through a resolved promise turns that into a catchable
  // rejection instead of one that would take down this component.
  useLayoutEffect(() => {
    let cancelled = false;
    Promise.resolve()
      .then(() => Environment())
      .then((env) => {
        if (!cancelled) setPlatform(env.platform);
      })
      .catch(() => {
        // Platform query unavailable -- the generic wording below covers it.
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleOpenHtml() {
    if (openBusy) return;
    setOpenBusy(true);
    onError(null);
    // catalogDir is required for the Go side's containment check (FU-23-A) --
    // fail closed here rather than sending an empty directory the Go side
    // would just reject anyway. Same guard handleReveal already has.
    if (!catalogDir) {
      onError('No catalog directory configured.');
      setOpenBusy(false);
      return;
    }
    const htmlPathResult = await wailsAPI.getCatalogHtmlPath(catalog.path, catalogDir);
    if (htmlPathResult.success) {
      const openResult = await wailsAPI.openExternal(htmlPathResult.htmlPath, catalogDir);
      if (!openResult.success) {
        onError(openResult.error);
      }
    } else {
      onError(htmlPathResult.error);
    }
    setOpenBusy(false);
  }

  async function handleReveal() {
    if (revealBusy) return;
    setRevealBusy(true);
    onError(null);
    // catalogDir is required for the Go side's containment check (WR-02) --
    // fail closed here rather than sending an empty directory the backend
    // would just reject anyway.
    if (!catalogDir) {
      onError('No catalog directory configured.');
      setRevealBusy(false);
      return;
    }
    const result = await wailsAPI.revealInFileManager(catalog.path, catalogDir);
    if (!result.success) {
      onError(result.error);
    }
    setRevealBusy(false);
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8, flex: 'none' }}>
      {/* Omitted entirely (not greyed) when the catalog has no HTML
          companion -- a button whose only possible outcome is an error is
          worse than no button, mirroring the header chip's own rule. */}
      {catalog.hasHtml && (
        <button
          type="button"
          onClick={handleOpenHtml}
          disabled={openBusy}
          style={{
            ...buttonBase,
            background: 'var(--ac)',
            color: 'var(--onac)',
            fontWeight: 600,
            border: 'none',
            opacity: openBusy ? 0.7 : 1,
          }}
        >
          Open HTML catalog
        </button>
      )}
      <button
        type="button"
        className="ws-details-reveal"
        onClick={handleReveal}
        disabled={revealBusy}
        style={{
          ...buttonBase,
          background: 'transparent',
          border: '1px solid var(--l)',
          color: 'var(--tx)',
          opacity: revealBusy ? 0.7 : 1,
        }}
      >
        {revealButtonLabel(platform)}
      </button>
      <button
        type="button"
        onClick={() => setRescanOpen(true)}
        disabled={isScanningNow}
        aria-disabled={isScanningNow}
        title={isScanningNow ? 'A scan is already running — open it from the status bar.' : undefined}
        style={{
          ...buttonBase,
          background: 'transparent',
          border: '1px solid var(--l)',
          color: 'var(--dm)',
          opacity: isScanningNow ? 0.7 : 1,
        }}
      >
        Re-scan volume…
      </button>
      {error && (
        <span style={{ fontSize: 11, color: 'var(--danger)', lineHeight: 1.4 }}>{error}</span>
      )}
      {rescanOpen && (
        <RescanDialog
          catalog={catalog}
          oldTreeAvailable
          onError={onError}
          onClose={() => setRescanOpen(false)}
        />
      )}
    </div>
  );
}

function DetailsPanel({ variant = 'pane' }: DetailsPanelProps) {
  const { state } = useAppContext();
  // Hoisted so both CatalogActions' menu placeholders and Footer's two
  // actions share one error slot (27-UI-SPEC.md) -- declared before the
  // early returns below, alongside the pre-existing useAppContext() call,
  // so no hook-order rule is broken.
  const [actionError, setActionError] = useState<string | null>(null);

  const catalog = state.catalogs.find((c) => c.path === state.currentCatalogId);
  const selectedNode =
    state.selected !== null && state.tree.status === 'ready'
      ? state.tree.nodes.find((n) => n.path === state.selected)
      : undefined;

  // No catalog selected -- Phase 22's placeholder, unchanged and still
  // reachable (e.g. right after the directory empties).
  if (!catalog) {
    return (
      <div className={`ws-details ws-details--${variant}`} style={{ padding: 14, gap: 16 }}>
        <span
          style={{
            fontSize: 12,
            fontWeight: 600,
            letterSpacing: '0.04em',
            textTransform: 'uppercase',
            color: 'var(--dm)',
            flex: 'none',
          }}
        >
          Details
        </span>
        <div
          className="pane-scroll"
          style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          <span style={{ fontSize: 12.5, color: 'var(--dm)', textAlign: 'center', lineHeight: 1.5 }}>
            Nothing selected. Pick a catalog in the rail, or catalog a volume to get started.
          </span>
        </div>
      </div>
    );
  }

  // Node-level view: a file or folder is selected in the tree. Only
  // reachable while the tree is ready (TreePane only ever dispatches
  // SET_SELECTED from a rendered row), so selectedNode is always found here.
  if (selectedNode) {
    const isDirectory = selectedNode.type === 'directory';
    const metaRows: MetaRow[] = [
      ['Type', isDirectory ? 'Folder' : 'File'],
      ['Size', formatBytes(selectedNode.size)],
      ['Catalog', catalog.title],
      ['Depth', String(selectedNode.depth)],
      ['Indexed', formatDate(catalog.modified)],
    ];

    return (
      <div
        className={`ws-details ws-details--${variant}`}
        style={{ padding: 14, gap: 16, display: 'flex', flexDirection: 'column' }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span
            style={{ fontSize: 12, fontWeight: 600, letterSpacing: '0.04em', textTransform: 'uppercase', color: 'var(--dm)' }}
          >
            {isDirectory ? 'Selected folder' : 'Selected file'}
          </span>
          <CatalogActions catalog={catalog} catalogDir={state.catalogDir} onError={setActionError} />
        </div>
        <div className="pane-scroll" style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div className="mono" style={{ fontSize: 12.5, lineHeight: 1.4, wordBreak: 'break-all' }}>
              {selectedNode.name}
            </div>
            <div
              className="mono"
              style={{
                fontSize: 11,
                color: 'var(--dm)',
                lineHeight: 1.5,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {selectedNode.path}
            </div>
          </div>
          <MetaRows rows={metaRows} />
        </div>
        <Footer catalog={catalog} catalogDir={state.catalogDir} error={actionError} onError={setActionError} />
      </div>
    );
  }

  // Catalog-level view: a catalog is selected, no node chosen -- the
  // default after selecting a rail row, and also what renders for a
  // catalog whose JSON failed to parse (the fields below all come from the
  // rail listing, independent of whether LoadCatalogFlat ever ran).
  //
  // Files/Catalogued prefer the loaded FlatCatalog's exact counts (never
  // absent once the tree for THIS catalog has finished loading) over the
  // rail's cache-backed, possibly-cold/possibly-null CatalogMetadata
  // fields -- same precedent TreeHeader.tsx already established.
  const loadedTree =
    state.tree.status === 'ready' && state.currentCatalogId === catalog.path ? state.tree : null;
  const fileCount = loadedTree ? loadedTree.fileCount : (catalog.fileCount ?? null);
  const totalBytes = loadedTree ? loadedTree.totalBytes : (catalog.totalBytes ?? null);
  const metaRows: MetaRow[] = [
    ['Files', fileCount != null ? formatCount(fileCount) : '—'],
    ['Catalogued', totalBytes != null ? formatBytes(totalBytes) : '—'],
    ['JSON', formatBytes(catalog.size)],
    ['Modified', formatDate(catalog.modified)],
  ];

  return (
    <div
      className={`ws-details ws-details--${variant}`}
      style={{ padding: 14, gap: 16, display: 'flex', flexDirection: 'column' }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <span style={{ fontSize: 12, fontWeight: 600, letterSpacing: '0.04em', textTransform: 'uppercase', color: 'var(--dm)' }}>
          Catalog
        </span>
        <CatalogActions catalog={catalog} catalogDir={state.catalogDir} onError={setActionError} />
      </div>
      <div className="pane-scroll" style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div className="mono" style={{ fontSize: 12.5, lineHeight: 1.4, wordBreak: 'break-all' }}>
            {catalog.title}
          </div>
          <div
            className="mono"
            style={{
              fontSize: 11,
              color: 'var(--dm)',
              lineHeight: 1.5,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {catalog.path}
          </div>
        </div>
        {/* Deliberately no fifth "HTML" row -- the tree header's .json/.html
            chips already communicate that fact; adding it here would need a
            new existence-check capability outside this phase's locked three
            backend surfaces. */}
        <MetaRows rows={metaRows} />
      </div>
      <Footer catalog={catalog} catalogDir={state.catalogDir} error={actionError} onError={setActionError} />
    </div>
  );
}

export default DetailsPanel;
