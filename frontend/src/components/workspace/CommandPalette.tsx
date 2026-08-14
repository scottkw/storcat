import { useEffect, useRef, useState } from 'react';
import { useAppContext } from '../../contexts/AppContext';
import { wailsAPI } from '../../services/wailsAPI';
import { formatCount } from '../../lib/format';
import { models } from '../../../wailsjs/go/models';
import { useModalBehavior } from '../../hooks/useModalBehavior';
import PaletteResultList, { PALETTE_PAGE_STEP_FALLBACK } from './palette/PaletteResultList';

export interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}

const PALETTE_DEBOUNCE_MS = 200;
const PALETTE_MIN_QUERY = 2;
const PALETTE_LISTBOX_ID = 'ws-palette-listbox';

// Always mounted by WorkspaceShell and returns null when closed -- it must
// not be conditionally mounted, because the shared useModalBehavior hook
// below observes the isOpen: true -> false transition to release scroll
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
  const [activeIndex, setActiveIndex] = useState(-1);

  // Wails bindings take no abort signal, so cancelling an in-flight request
  // is impossible -- gating the *handling* of the response is the only
  // mechanism that works. Every dispatch increments this ref and captures
  // the value in its closure; a response is applied only if the captured
  // value still matches the ref's current value when it lands.
  const requestIdRef = useRef(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listScrollRef = useRef<HTMLDivElement>(null);

  // Focus trap, Escape-to-close, scroll lock, and focus restore all arrive
  // through the shared hook -- the palette implements none of the four
  // itself. The panel (not the scrim) is the trap boundary, since the
  // scrim is the click-to-close surface and focus must never land there.
  const { containerRef } = useModalBehavior({ isOpen, onClose, initialFocusRef: inputRef });

  // No persisted "last query" across open/close cycles (deferred idea, not
  // this phase) -- every open starts from a clean slate.
  useEffect(() => {
    if (!isOpen) return;
    setQuery('');
    setResults([]);
    setTotal(0);
    setSettled(false);
  }, [isOpen]);

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

  // A fresh result set always starts on its first row; an empty one has no
  // active row to activate. Keyed on the results array reference, which
  // only changes when a new response (or a reset) lands.
  useEffect(() => {
    setActiveIndex(results.length > 0 ? 0 : -1);
  }, [results]);

  // The activation seam: in this plan, Enter/click just closes the palette
  // (PLT-04's complete scope). Plan 24-05 extends this same handler with
  // the catalog switch and reveal request PLT-05 specifies -- it adds to
  // this seam rather than replacing it.
  function handleActivate(_result: models.SearchResult) {
    onClose();
  }

  // Reads the live viewport instead of hardcoding a step: clientHeight of
  // the scroll region divided by the rendered height of one option row,
  // floored and clamped to at least one row. Falls back when the ref or an
  // option isn't measurable yet (e.g. first render).
  function computePageStep(): number {
    const container = listScrollRef.current;
    if (!container) return PALETTE_PAGE_STEP_FALLBACK;
    const firstOption = container.querySelector<HTMLElement>('[role="option"]');
    if (!firstOption || firstOption.offsetHeight === 0) return PALETTE_PAGE_STEP_FALLBACK;
    return Math.max(Math.floor(container.clientHeight / firstOption.offsetHeight), 1);
  }

  // Handles exactly the seven navigation/activation keys. Escape is
  // deliberately left alone -- the shared useModalBehavior hook already
  // closes on Escape from anywhere inside the panel, and intercepting it
  // here would produce two close paths for one press.
  function handleKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    const lastIndex = results.length - 1;
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        setActiveIndex((i) => Math.min(i + 1, lastIndex));
        break;
      case 'ArrowUp':
        event.preventDefault();
        setActiveIndex((i) => Math.max(i - 1, 0));
        break;
      case 'Home':
        event.preventDefault();
        setActiveIndex(0);
        break;
      case 'End':
        event.preventDefault();
        setActiveIndex(lastIndex);
        break;
      case 'PageDown':
        event.preventDefault();
        setActiveIndex((i) => Math.min(i + computePageStep(), lastIndex));
        break;
      case 'PageUp':
        event.preventDefault();
        setActiveIndex((i) => Math.max(i - computePageStep(), 0));
        break;
      case 'Enter': {
        event.preventDefault();
        const active = results[activeIndex];
        if (active) handleActivate(active);
        break;
      }
      default:
        return;
    }
  }

  if (!isOpen) return null;

  const trimmedQuery = query.trim();
  const isHint = trimmedQuery.length < PALETTE_MIN_QUERY || !state.catalogDir;
  const isSearching = !isHint && !settled;
  const isResults = !isHint && !isSearching && total > 0;

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

  const activeOptionId = isResults && activeIndex >= 0 ? `ws-palette-option-${activeIndex}` : undefined;

  return (
    <div className="ws-palette-scrim" onClick={onClose}>
      <div
        className="ws-palette-panel"
        ref={containerRef}
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
            ref={inputRef}
            role="combobox"
            aria-expanded={isResults}
            aria-controls={PALETTE_LISTBOX_ID}
            aria-activedescendant={activeOptionId}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            aria-label="Search every catalog"
          />
          <span className="mono" style={{ fontSize: 11, color: 'var(--fn)', flex: 'none' }}>
            {readout}
          </span>
        </div>
        {isHint ? (
          <div className="ws-palette-state">Type to search…</div>
        ) : isSearching ? (
          <div className="ws-palette-state">Searching…</div>
        ) : isResults ? (
          <PaletteResultList
            id={PALETTE_LISTBOX_ID}
            results={results}
            total={total}
            query={query}
            activeIndex={activeIndex}
            onActiveIndexChange={setActiveIndex}
            onActivate={handleActivate}
            scrollRef={listScrollRef}
          />
        ) : (
          <div className="ws-palette-state">No file in any catalog matches that.</div>
        )}
        <div className="ws-palette-footer mono">
          <span>↵ reveal in catalog</span>
          <span>esc close</span>
          <span style={{ marginLeft: 'auto' }}>searches names and paths</span>
        </div>
      </div>
    </div>
  );
}

export default CommandPalette;
