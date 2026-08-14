import { useMemo } from 'react';
import { models } from '../../wailsjs/go/models';

/**
 * Derives the DFS-order list of visible node indices from the flat node
 * array plus the expanded map -- a single O(n) linear pass over an already
 * pre-ordered array, never a recursive re-walk and never a rebuild of
 * `nodes` itself. This is a useMemo (not an effect) because it is a pure
 * derivation of its two inputs, recomputed only when either changes.
 *
 * A node is visible only when its parent is BOTH visible AND present as a
 * true value in `expanded` (keyed by the parent's path) -- checking only
 * parent-visibility would make every node visible regardless of expansion.
 */
export function useVisibleRows(nodes: models.FlatNode[], expanded: Record<string, boolean>): number[] {
  return useMemo(() => {
    const isVisible: boolean[] = new Array(nodes.length);
    const indices: number[] = [];

    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i];
      const visible =
        node.parentIdx === -1
          ? true
          : isVisible[node.parentIdx] === true && expanded[nodes[node.parentIdx].path] === true;
      isVisible[i] = visible;
      if (visible) indices.push(i);
    }

    return indices;
  }, [nodes, expanded]);
}
