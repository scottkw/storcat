import { useEffect, useRef, useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { formatBytes, formatCount } from '../../lib/format';
import { models } from '../../../wailsjs/go/models';

export interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}

const PALETTE_DEBOUNCE_MS = 200;
const PALETTE_MIN_QUERY = 2;

// Always mounted by WorkspaceShell and returns null when closed -- it must
// not be conditionally mounted, because 24-03's shared useModalBehavior hook
// has to observe the isOpen: true -> false transition to release scroll
// lock and restore focus, and Phase 25's animated exit depends on the same
// contract.
function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
  const { state } = useAppContext();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<models.SearchResult[]>([]);
  const [total, setTotal] = useState(0);
  // Records whether any query has produced a settled result set this
  // session -- the quiet "Searching…" readout below only shows on the very
  // first in-flight query, so later keystrokes keep prior results visible
  // instead of flashing.
  const [settled, setSettled] = useState(false);

  // Wails bindings take no abort signal, so cancelling an in-flight request
  // is impossible -- gating the *handling* of the response is the only
  // mechanism that works. Every dispatch increments this ref and captures
  // the value in its closure; a response is applied only if the captured
  // value still matches the ref's current value when it lands.
  const requestIdRef = useRef(0);

  // No persisted "last query" across open/close cycles (deferred idea, not
  // this phase) -- every open starts from a clean slate.
  useEffect(() => {
    if (!isOpen) return;
    setQuery('');
    setResults([]);
    setTotal(0);
    setSettled(false);
  }, [isOpen]);

  // Escape closes the palette. Inline here for this slice only -- 24-03's
  // shared useModalBehavior hook replaces this with the trap/scroll-lock/
  // focus-restore-aware version.
  useEffect(() => {
    if (!isOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [isOpen, onClose]);

  useEffect(() => {
    if (!isOpen) return;

    const trimmed = query.trim();
    // Below the minimum, or no catalog directory configured yet: hint
    // branch, no call issued.
    if (trimmed.length < PALETTE_MIN_QUERY || !state.catalogDir) {
      setResults([]);
      setTotal(0);
      setSettled(false);
      return;
    }

    const catalogDir = state.catalogDir;
    const requestId = ++requestIdRef.current;
    const timer = setTimeout(() => {
      wailsAPI.searchIndexed(trimmed, catalogDir).then((result) => {
        if (requestId !== requestIdRef.current) return; // stale response, dropped
        if (result.success) {
          setResults(result.indexed.results ?? []);
          setTotal(result.indexed.total);
        } else {
          // No distinct error surface in this palette -- a failed binding
          // call is treated as zero results.
          setResults([]);
          setTotal(0);
        }
        setSettled(true);
      });
    }, PALETTE_DEBOUNCE_MS);

    return () => clearTimeout(timer);
  }, [query, isOpen, state.catalogDir]);

  if (!isOpen) return null;

  const trimmedQuery = query.trim();
  const isHint = trimmedQuery.length < PALETTE_MIN_QUERY || !state.catalogDir;
  const isSearching = !isHint && !settled;

  const filesIndexed = state.catalogs.reduce(
    (sum, catalog) => (typeof catalog.fileCount === 'number' ? sum + catalog.fileCount : sum),
    0
  );
  const catalogCount = state.catalogs.length;
  const placeholder = `Search ${formatCount(filesIndexed)} files across ${catalogCount} catalogs…`;

  const readout = isHint
    ? 'Type to search…'
    : isSearching
      ? 'Searching…'
      : total > 50
        ? `50 of ${total}`
        : `${total} hits`;

  return (
    <div className="ws-palette-scrim" onClick={onClose}>
      <div
        className="ws-palette-panel"
        role="dialog"
        aria-modal="true"
        aria-label="Search every catalog"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="ws-palette-input-row">
          <svg
            width="15"
            height="15"
            viewBox="0 0 16 16"
            fill="none"
            stroke="var(--ac)"
            strokeWidth={1.6}
            aria-hidden="true"
            focusable="false"
          >
            <circle cx="7" cy="7" r="4.5" />
            <line x1="10.5" y1="10.5" x2="14" y2="14" />
          </svg>
          <input
            className="ws-palette-input"
            autoFocus
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={placeholder}
            aria-label="Search every catalog"
          />
          <span className="mono" style={{ fontSize: 11, color: 'var(--fn)', flex: 'none' }}>
            {readout}
          </span>
        </div>
        <div className="ws-palette-list">
          {results.map((result, i) => (
            <div className="ws-palette-row" key={`${result.catalogFilePath}:${result.fullName}:${i}`}>
              <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 2 }}>
                <span
                  className="mono"
                  style={{ fontSize: 12.5, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                >
                  {result.basename}
                </span>
                <span
                  className="mono"
                  style={{
                    fontSize: 11,
                    color: 'var(--dm)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {result.fullPath}
                </span>
              </div>
              <span className="mono" style={{ fontSize: 11, color: 'var(--dm)', flex: 'none' }}>
                {result.catalog}
              </span>
              <span className="mono" style={{ fontSize: 11, color: 'var(--dm)', flex: 'none' }}>
                {formatBytes(result.size)}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export default CommandPalette;
