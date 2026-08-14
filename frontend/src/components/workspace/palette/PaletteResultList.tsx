import { useEffect } from 'react';
import { models } from '../../../../wailsjs/go/models';
import PaletteResultRow from './PaletteResultRow';

export interface PaletteResultListProps {
  results: models.SearchResult[];
  total: number;
  query: string;
  activeIndex: number;
  onActiveIndexChange: (index: number) => void;
  onActivate: (result: models.SearchResult) => void;
  scrollRef: React.RefObject<HTMLDivElement>;
  // Lets the consumer's combobox input point aria-controls at this listbox.
  id?: string;
}

// Fallback page step for PageUp/PageDown when the viewport can't be
// measured yet (e.g. first render, before layout).
export const PALETTE_PAGE_STEP_FALLBACK = 10;

// Two different catalogs can legitimately contain the same node path, so
// both rows must render -- the row key is the (catalogFilePath, fullName)
// pair, joined by a NUL (an escape in source, never a raw control byte),
// a separator that cannot occur in a filesystem path.
const PALETTE_KEY_SEPARATOR = String.fromCharCode(0);

// Flat, unvirtualized listbox in backend order (PLT-02/PLT-03/PLT-04). 50
// rows never needs virtualization -- results arriving here are already the
// capped set Go computed, never re-sliced client-side.
function PaletteResultList({
  results,
  total,
  query,
  activeIndex,
  onActiveIndexChange,
  onActivate,
  scrollRef,
  id,
}: PaletteResultListProps) {
  // Keeps the active option scrolled into view as the keyboard drives it.
  // 'nearest' is what keeps arrow-key traversal from jerking the list when
  // the active row is already visible.
  useEffect(() => {
    const container = scrollRef.current;
    if (!container) return;
    const option = container.querySelector(`#ws-palette-option-${activeIndex}`);
    option?.scrollIntoView({ block: 'nearest' });
  }, [activeIndex, scrollRef]);

  return (
    <div id={id} className="ws-palette-list" role="listbox" aria-label="Search results" ref={scrollRef}>
      {results.map((result, index) => (
        <PaletteResultRow
          key={result.catalogFilePath + PALETTE_KEY_SEPARATOR + result.fullName}
          result={result}
          query={query}
          active={index === activeIndex}
          optionId={`ws-palette-option-${index}`}
          onActivate={() => onActivate(result)}
          onHover={() => onActiveIndexChange(index)}
        />
      ))}
      {total > results.length && (
        // WR-04: results.length is the single source of truth for the
        // displayed count (Go's SearchIndexedCap, not a re-stated `50`
        // literal) -- if the backend cap ever changes, this stays correct.
        <div className="ws-palette-truncation">{`Showing the first ${results.length} of ${total} hits`}</div>
      )}
    </div>
  );
}

export default PaletteResultList;
