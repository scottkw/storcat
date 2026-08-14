import React, { createContext, useContext, useReducer, ReactNode } from 'react';
import { Density, RailSide, readPersistedPrefs } from '../themeTokens';
import { models } from '../../wailsjs/go/models';

// Types

// A discriminated union on `status` rather than three fields (loading/error/
// data) that could disagree with each other -- there is exactly one shape
// per state, so a "loading but has stale nodes" bug is unrepresentable.
export type TreeState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'ready'; nodes: models.FlatNode[]; fileCount: number; totalBytes: number }
  | { status: 'error'; message: string };

export interface AppState {
  density: Density;
  railSide: RailSide;
  detailOverlay: boolean;
  catalogDir: string | null;
  catalogs: models.CatalogMetadata[];
  currentCatalogId: string | null;
  tree: TreeState;
  expanded: Record<string, boolean>;
  selected: string | null;
  // The path of a node the ⌘K palette asked the tree to reveal, held across
  // the asynchronous gap between the catalog switch and the flat tree
  // arriving. TreePane consumes and clears it once the tree is ready.
  pendingReveal: string | null;
}

type AppAction =
  | { type: 'SET_DENSITY'; payload: Density }
  | { type: 'SET_RAIL_SIDE'; payload: RailSide }
  | { type: 'SET_DETAIL_OVERLAY'; payload: boolean }
  | { type: 'SET_CATALOG_DIR'; payload: string }
  | { type: 'SET_CATALOGS'; payload: models.CatalogMetadata[] }
  | { type: 'SELECT_CATALOG'; payload: string }
  | {
      type: 'TREE_LOADED';
      payload: { catalogId: string; nodes: models.FlatNode[]; fileCount: number; totalBytes: number };
    }
  | { type: 'TREE_FAILED'; payload: { catalogId: string; message: string } }
  | { type: 'TOGGLE_EXPAND'; payload: string }
  | { type: 'SET_EXPANDED'; payload: Record<string, boolean> }
  | { type: 'MERGE_EXPANDED'; payload: string[] }
  | { type: 'SET_SELECTED'; payload: string | null }
  | { type: 'SET_PENDING_REVEAL'; payload: string | null }
  | { type: 'REVEAL_HIT'; payload: { catalogId: string; path: string } };

// Seeded once at module scope so a relaunch restores the user's persisted
// density/rail-side choices rather than hardcoded defaults. readPersistedPrefs
// already applies the allowlist and locked fallbacks -- no second validation here.
const persistedPrefs = readPersistedPrefs();

const initialState: AppState = {
  density: persistedPrefs.density,
  railSide: persistedPrefs.railSide,
  detailOverlay: false,
  catalogDir: null,
  catalogs: [],
  currentCatalogId: null,
  tree: { status: 'idle' },
  expanded: {},
  selected: null,
  pendingReveal: null,
};

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case 'SET_DENSITY':
      return { ...state, density: action.payload };
    case 'SET_RAIL_SIDE':
      return { ...state, railSide: action.payload };
    case 'SET_DETAIL_OVERLAY':
      return { ...state, detailOverlay: action.payload };
    case 'SET_CATALOG_DIR':
      // Changing directories clears the current catalog, returns the tree to
      // idle, and empties expansion/selection in the same case -- without
      // this a directory change can leave the tree showing a catalog that is
      // no longer in the (not-yet-reloaded) list (RAIL-05).
      // pendingReveal is cleared here too -- a directory change invalidates
      // any reveal that was waiting on a catalog load, the same discard
      // discipline SELECT_CATALOG applies below.
      return {
        ...state,
        catalogDir: action.payload,
        currentCatalogId: null,
        tree: { status: 'idle' },
        expanded: {},
        selected: null,
        pendingReveal: null,
      };
    case 'SET_CATALOGS':
      return { ...state, catalogs: action.payload };
    case 'SELECT_CATALOG':
      // Atomic: sets currentCatalogId, starts the tree load, and clears
      // expanded/selected together -- no intermediate state where one is
      // cleared and the other is not (TREE-06). pendingReveal is cleared in
      // the same update -- this is the stale-discard mechanism for the ⌘K
      // reveal path: a rail-driven catalog switch cancels any reveal still
      // waiting on a load, mirroring the discipline TREE_LOADED already
      // applies to a load superseded by a newer selection.
      return {
        ...state,
        currentCatalogId: action.payload,
        tree: { status: 'loading' },
        expanded: {},
        selected: null,
        pendingReveal: null,
      };
    case 'TREE_LOADED':
      // A load that resolves after the user has already selected a
      // different catalog is discarded -- only applied when the id it was
      // issued for still matches current state (RAIL-03, TREE-06).
      if (action.payload.catalogId !== state.currentCatalogId) return state;
      return {
        ...state,
        tree: {
          status: 'ready',
          nodes: action.payload.nodes,
          fileCount: action.payload.fileCount,
          totalBytes: action.payload.totalBytes,
        },
      };
    case 'TREE_FAILED':
      if (action.payload.catalogId !== state.currentCatalogId) return state;
      return { ...state, tree: { status: 'error', message: action.payload.message } };
    case 'TOGGLE_EXPAND': {
      // A synchronous flip keyed by path -- a burst of clicks on the same
      // caret applies in order with no lost update, and this never touches
      // the node array (TREE-02, TREE-03 concurrency).
      const path = action.payload;
      if (state.expanded[path]) {
        const { [path]: _dropped, ...rest } = state.expanded;
        return { ...state, expanded: rest };
      }
      return { ...state, expanded: { ...state.expanded, [path]: true } };
    }
    case 'SET_EXPANDED':
      // Replaces the whole map in one state update -- expand-all and
      // collapse-to-root both use this, never a per-node dispatch loop.
      return { ...state, expanded: action.payload };
    case 'MERGE_EXPANDED': {
      // Structurally distinct from SET_EXPANDED: this case can only ever
      // add entries, never drop the ones already there. WR-01 -- the reveal
      // path used to rely on a caller (lib/reveal.ts#mergeExpanded) spreading
      // `state.expanded` before dispatching SET_EXPANDED; a future caller
      // that forgot that step would silently collapse every open branch.
      // Folding the merge into the reducer itself removes that possibility
      // at the type level -- there is no way to dispatch MERGE_EXPANDED with
      // "replace" semantics.
      let changed = false;
      const next = { ...state.expanded };
      for (const path of action.payload) {
        if (next[path] !== true) {
          next[path] = true;
          changed = true;
        }
      }
      // Returning the same state object when nothing changed lets React's
      // reducer bail-out skip the re-render -- this is what makes a repeat
      // reveal of an already-visible node a no-op, the same idempotence the
      // old mergeExpanded() reference check provided at the call site.
      return changed ? { ...state, expanded: next } : state;
    }
    case 'SET_SELECTED':
      return { ...state, selected: action.payload };
    case 'SET_PENDING_REVEAL':
      return { ...state, pendingReveal: action.payload };
    case 'REVEAL_HIT': {
      // WR-02 -- folds CommandPalette's two load-bearing, order-dependent
      // dispatches (SELECT_CATALOG then SET_PENDING_REVEAL) into one atomic
      // update. SELECT_CATALOG's clearing of pendingReveal was what made the
      // dispatch order matter; a single action has no ordering to get wrong.
      // Only resets tree/expanded/selected when the reveal actually targets
      // a different catalog -- a same-catalog reveal (the hit is already the
      // open catalog) must not blow away in-progress expansion/selection.
      const switching = action.payload.catalogId !== state.currentCatalogId;
      return {
        ...state,
        currentCatalogId: action.payload.catalogId,
        tree: switching ? { status: 'loading' } : state.tree,
        expanded: switching ? {} : state.expanded,
        selected: switching ? null : state.selected,
        pendingReveal: action.payload.path,
      };
    }
    default:
      return state;
  }
}

interface AppContextType {
  state: AppState;
  dispatch: React.Dispatch<AppAction>;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(appReducer, initialState);

  return (
    <AppContext.Provider value={{ state, dispatch }}>
      {children}
    </AppContext.Provider>
  );
}

export function useAppContext() {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useAppContext must be used within an AppProvider');
  }
  return context;
}
