import { models } from '../../wailsjs/go/models';

/**
 * Pure helpers behind the ⌘K palette's "reveal in tree" path (PLT-05). No
 * React, no DOM, and nothing that mutates reducer state lives in this file
 * -- every function here takes plain data in and returns plain data out, so
 * the ancestor-walk logic is readable and reviewable on its own. Consumed
 * exclusively by TreePane.tsx's reveal effects. The merge-into-`expanded`
 * step used to live here (`mergeExpanded`) but was folded into the
 * reducer's MERGE_EXPANDED case (AppContext.tsx, WR-01) so the merge-vs-
 * replace semantics are enforced structurally rather than by caller
 * discipline.
 */

/**
 * Linear scan for the node whose `path` matches exactly. Returns -1 when no
 * node matches -- the caller (TreePane) treats that as "the catalog changed
 * on disk since the search ran" and discards the pending reveal rather than
 * retrying.
 */
export function findNodeIndexByPath(nodes: models.FlatNode[], path: string): number {
  for (let i = 0; i < nodes.length; i++) {
    if (nodes[i].path === path) return i;
  }
  return -1;
}

/**
 * Walks the `parentIdx` chain from the target's PARENT up to a depth-0 node,
 * returning every ancestor path (excluding the target's own path, so a
 * directory hit is selected but never auto-expanded). `-1` is a sentinel,
 * not an index: LoadCatalogFlat excludes the catalog root from the flat
 * array (internal/search/flatten.go:19-20,63-67), so there is no root
 * object to walk up to and `nodes[-1]` is never dereferenced.
 *
 * Bounded to at most `nodes.length` iterations as a cycle guard -- a
 * malformed or future catalog with a self-referential parentIdx chain must
 * terminate rather than freeze the renderer (T-24-15).
 */
export function ancestorPathsOf(nodes: models.FlatNode[], targetIdx: number): string[] {
  const ancestors: string[] = [];
  let idx = nodes[targetIdx].parentIdx;
  let steps = 0;
  while (idx !== -1 && steps < nodes.length) {
    ancestors.push(nodes[idx].path);
    idx = nodes[idx].parentIdx;
    steps++;
  }
  return ancestors;
}
