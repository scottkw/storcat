import { models } from '../../wailsjs/go/models';

/**
 * Mirrors pkg/models/catalog.go's DiffState/DiffEntry/DiffResult
 * field-for-field, the same "mirror, don't reuse the generated class"
 * convention ScanProgress already established for a Wails-bridged shape --
 * kept as plain data types here, independent of the generated
 * models.DiffResult class's constructor/convertValues machinery, which this
 * app's other hand-authored state types (ScanState, RescanState) don't need.
 */
export type DiffState = 'added' | 'removed' | 'changed' | 'unreadable' | 'unchanged';

export interface DiffEntry {
  path: string;
  state: DiffState;
  type: string;
  oldSize?: number;
  newSize?: number;
  readError?: string;
}

/**
 * The five category counts must sum to the total number of distinct paths
 * across the old tree union the new tree (28-UI-SPEC.md's stated
 * invariant) -- entries carries one row per added/removed/changed/
 * unreadable path, never for unchanged (count-only).
 */
export interface DiffResult {
  added: number;
  removed: number;
  changed: number;
  unreadable: number;
  unchanged: number;
  entries: DiffEntry[];
  // oldEntryCount/lowSimilarity mirror pkg/models/catalog.go's DiffResult
  // fields added in 28-02 -- oldEntryCount gates the similarity floor,
  // lowSimilarity is the wrong-disc signal the diff step's warning banner
  // reads (28-UI-SPEC.md's Similarity Warning Contract). Both are a signal
  // only: nothing on the Go side blocks or refuses on lowSimilarity.
  oldEntryCount: number;
  lowSimilarity: boolean;
}

/**
 * The four diff-list group keys, in the fixed display order the Diff List
 * Contract locks (28-UI-SPEC.md): Added, Removed, Changed, Unreadable.
 * 'unchanged' is deliberately excluded -- unchanged entries never appear in
 * DiffResult.entries at all (count-only in the stat tile), so there is no
 * group for them to render.
 */
export type DiffGroupKey = Exclude<DiffState, 'unchanged'>;

/**
 * The three steps this tracer's RescanDialog drives: pick a source volume,
 * watch the shared live scan progress (state.scan, the same slice Create
 * drives), then see the computed diff. The error/interrupted step and the
 * write-resolution footer (Overwrite/Keep-both) are plan 28-02/03/04.
 */
export type RescanStep = 'pick-volume' | 'scanning' | 'diff';

/**
 * state.rescan is a NEW, separate AppContext reducer slice -- it does not
 * extend or replace ScanState (28-UI-SPEC.md's Architecture & State). Live
 * progress keeps driving state.scan via the existing SCAN_STARTED/
 * SCAN_PROGRESS actions (shared with Create, subscribed through the one
 * scan:progress event listener CreateSlideOver already owns); only the
 * terminal diff outcome lives here. null means no re-scan dialog is
 * currently open (it is conditionally mounted by its parent, not
 * always-mounted like CreateSlideOver/SettingsDialog).
 */
export interface RescanState {
  step: RescanStep;
  catalog: models.CatalogMetadata;
  sourcePath: string | null;
  oldTreeAvailable: boolean;
  diff: DiffResult | null;
}
