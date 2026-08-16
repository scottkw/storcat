import { formatBytes } from '../../../lib/format';
import type { DiffEntry, DiffGroupKey } from '../../../types/rescan';

export interface DiffListProps {
  entries: DiffEntry[];
}

// Fixed display order (28-UI-SPEC.md's Diff List Contract) -- matches
// 28-CONTEXT.md's own prose order and ACT-06's literal requirement wording,
// so it needed no invention. 'unchanged' is never a group here (see
// DiffGroupKey's own doc comment in types/rescan.ts).
const GROUP_ORDER: DiffGroupKey[] = ['added', 'removed', 'changed', 'unreadable'];

const GROUP_META: Record<DiffGroupKey, { label: string; glyph: string; color: string }> = {
  added: { label: 'ADDED', glyph: '+', color: 'var(--ac)' },
  // U+2212 MINUS SIGN, matching the handoff -- not a hyphen.
  removed: { label: 'REMOVED', glyph: '−', color: 'var(--danger)' },
  changed: { label: 'CHANGED', glyph: '~', color: '#f0b429' },
  // Reuses --danger (the same token as removed), differentiated by glyph
  // and grouping, not hue -- 28-UI-SPEC.md's Color section.
  unreadable: { label: 'UNREADABLE', glyph: '!', color: 'var(--danger)' },
};

// The right column's content per group -- added/removed show one size,
// changed shows "old -> new", unreadable shows the short read-error reason
// in place of a size (none is knowable for a node that failed to read).
function rightColumnFor(entry: DiffEntry): string {
  switch (entry.state) {
    case 'added':
      return formatBytes(entry.newSize ?? 0);
    case 'removed':
      return formatBytes(entry.oldSize ?? 0);
    case 'changed':
      return `${formatBytes(entry.oldSize ?? 0)} → ${formatBytes(entry.newSize ?? 0)}`;
    case 'unreadable':
      return entry.readError ?? '';
    default:
      return '';
  }
}

/**
 * The grouped, natively-scrolling diff row list (28-UI-SPEC.md's Diff List
 * Contract) -- a plain `<div>` at `max-height: 200px; overflow-y: auto`,
 * deliberately not built on the tree pane's virtualizer or its
 * visible-rows-resolution hook: a diff has no hierarchy (every entry's Path
 * is already a full relative display path, no expand/collapse), and its
 * scale is bounded by what differs, not by catalog size. Native-scroll
 * performance against a pathological tens-of-thousands-of-rows diff is a
 * reasoned boundary, not a proven one (the UI-SPEC's own stated backstop).
 *
 * Four fixed-order groups; a group with zero entries renders no header and
 * no rows at all -- no reserved empty slot, same rule this app's other
 * conditional sections already follow.
 */
function DiffList({ entries }: DiffListProps) {
  const grouped: Record<DiffGroupKey, DiffEntry[]> = { added: [], removed: [], changed: [], unreadable: [] };
  for (const entry of entries) {
    if (entry.state === 'unchanged') continue;
    grouped[entry.state].push(entry);
  }
  for (const key of GROUP_ORDER) {
    grouped[key].sort((a, b) => a.path.localeCompare(b.path));
  }

  return (
    <div className="ws-rescan-difflist">
      {GROUP_ORDER.map((key) => {
        const rows = grouped[key];
        if (rows.length === 0) return null;
        const meta = GROUP_META[key];
        return (
          <div key={key} className="ws-rescan-diffgroup">
            <div className="ws-rescan-diffgroup-header">
              {meta.label} · {rows.length}
            </div>
            {rows.map((entry) => (
              <div key={`${key}:${entry.path}`} className="ws-rescan-diffrow">
                <span className="ws-rescan-diffrow-mark mono" style={{ color: meta.color }} aria-hidden="true">
                  {meta.glyph}
                </span>
                <span className="ws-rescan-diffrow-path mono">{entry.path}</span>
                <span className="ws-rescan-diffrow-meta mono">{rightColumnFor(entry)}</span>
              </div>
            ))}
          </div>
        );
      })}
    </div>
  );
}

export default DiffList;
