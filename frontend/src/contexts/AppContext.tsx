import React, { createContext, useContext, useReducer, ReactNode } from 'react';
import { Density, RailSide, readPersistedPrefs } from '../themeTokens';
import { models } from '../../wailsjs/go/models';
import { ScanProgress, ScanResultFile, ScanState } from '../types/scan';
import { scanPercent } from '../lib/scanFormat';
import { AppSettings, DEFAULT_APP_SETTINGS } from '../settingsStore';

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
  // Whether the create slide-over is open. Lifted to AppContext (not local
  // CreateSlideOver state) so a later plan's background-handoff/status-bar
  // segment can re-open it into its current live state (25-UI-SPEC.md).
  createOpen: boolean;
  // The create flow's scan state machine, lifted for the same reason --
  // survives the slide-over's own mount/unmount across a close/reopen.
  scan: ScanState;
  // Additive (25-07) rather than a change to SET_CREATE_OPEN's payload, so
  // no call site written in an earlier plan needs editing. Set by the tree
  // pane's "Choose catalog folder…" entry point; CreateSlideOver reads it
  // once on open, invokes the folder dialog, and clears it -- landing that
  // one entry point directly on the picker its label promises rather than
  // the generic volume-card list.
  createFolderPickerIntent: boolean;
  // The config-backed settings slice not already covered by density/
  // railSide/catalogDir above. Initialised to DEFAULT_APP_SETTINGS and
  // field-merged by SETTINGS_HYDRATED once WorkspaceShell's mount effect
  // resolves the real config (settingsStore.hydrateSettings) -- see
  // touchedSettings below for which fields that merge is allowed to touch.
  settings: AppSettings;
  // WR-B: keys explicitly written via SET_SETTINGS (a real user action),
  // tracked so SETTINGS_HYDRATED can skip them rather than inferring
  // "untouched" from value-equals-default. The equals-default heuristic
  // alone misfires when a user deliberately re-sets a field back to its
  // default value during the hydration race window -- that write looked
  // identical to "never touched" and a stale in-flight hydration result
  // could silently revert it. Never cleared -- SETTINGS_HYDRATED fires
  // once per launch, so nothing depends on this set shrinking again.
  touchedSettings: Set<keyof AppSettings>;
  // Bumped by a locally-initiated, already-succeeded Delete/Duplicate
  // (27-RAIL-FIX) so CatalogRail's existing catalogDir effect re-runs
  // loadCatalogsForDirectory -- the same single refresh path
  // catalogs:changed already re-triggers. Not a second way to compute the
  // rail's contents: this only re-invokes the one authoritative
  // browseCatalogs listing, on a state.catalogDir-independent trigger, so
  // the rail reflects the user's own action even with watching off (the
  // shipped default).
  railRefreshToken: number;
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
  // Added by 27-05 -- no existing action clears currentCatalogId without
  // also selecting/loading something else, and DeleteConfirmDialog's
  // onDeleted needs exactly that: clear currentCatalogId and selected so
  // DetailsPanel (and TreePane's own currentCatalogId guard) both fall
  // back to their existing "nothing selected" placeholders, no new empty
  // state needed.
  | { type: 'CLEAR_CURRENT_CATALOG' }
  // 27-RAIL-FIX: fired after a local Delete/Duplicate already succeeded, to
  // re-trigger the rail's one authoritative listing (see railRefreshToken).
  | { type: 'REQUEST_RAIL_REFRESH' }
  | { type: 'SET_PENDING_REVEAL'; payload: string | null }
  | { type: 'REVEAL_HIT'; payload: { catalogId: string; path: string } }
  | { type: 'SET_CREATE_OPEN'; payload: boolean }
  | { type: 'SET_CREATE_FOLDER_PICKER_INTENT'; payload: boolean }
  | { type: 'SCAN_STARTED'; payload: { title: string } }
  | { type: 'SCAN_PROGRESS'; payload: ScanProgress }
  // sourcePath: the failed scan's source, needed for ErrorBody's sub-line
  // (plan 25-07) -- CreateSlideOver knows it locally (the selected source),
  // but AppContext's own scan state has never tracked it, so it travels on
  // this action's payload rather than being re-derived in the reducer.
  | { type: 'SCAN_FAILED'; payload: { message: string; sourcePath: string } }
  | {
      type: 'SCAN_DONE';
      payload: {
        title: string;
        jsonPath: string;
        files: ScanResultFile[];
        fileCount: number;
        totalSize: number;
        durationMs: number;
        partial: boolean;
        stopPercent?: number | null;
      };
    }
  | { type: 'SCAN_RESET' }
  // Replaces the whole settings slice -- fired once, at hydration.
  | { type: 'SETTINGS_HYDRATED'; payload: AppSettings }
  // Merges a partial update -- every Catalogs/Toggles row dispatches this
  // alongside its own settingsStore setter call.
  | { type: 'SET_SETTINGS'; payload: Partial<AppSettings> };

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
  createOpen: false,
  scan: { status: 'idle' },
  createFolderPickerIntent: false,
  settings: DEFAULT_APP_SETTINGS,
  touchedSettings: new Set(),
  railRefreshToken: 0,
};

// Extracts the in-progress/terminal title from whichever ScanState variant
// is active, for SCAN_FAILED (which only carries a message, not a title) to
// preserve the title the scan was already running under. 'idle' has no
// title of its own.
function scanTitleOf(scan: ScanState): string {
  return 'title' in scan ? scan.title : '';
}

// The scanning body's log box retains at most this many newest-first lines
// (25-UI-SPEC E5 overflow) -- enforced here, at the point a line is
// appended to retained state, rather than only at render, so the state
// itself never grows unbounded over a long-running scan (T-25-23).
const SCAN_LOG_CAP = 9;

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case 'SET_DENSITY':
      return { ...state, density: action.payload };
    case 'SET_RAIL_SIDE':
      return { ...state, railSide: action.payload };
    case 'SET_DETAIL_OVERLAY':
      return { ...state, detailOverlay: action.payload };
    case 'SET_CATALOG_DIR':
      // Re-selecting the already-configured directory is a true no-op --
      // the same state object is returned (React's reducer bail-out skips
      // the re-render), so choosing the same directory from Settings or the
      // rail can never blow away currentCatalogId/tree/expanded/selected
      // below (SET-04 adjacency), from any call site, mirroring
      // SET_CREATE_OPEN's identical bail-out above.
      if (state.catalogDir === action.payload) return state;
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
    case 'CLEAR_CURRENT_CATALOG':
      if (state.currentCatalogId === null) return state;
      return { ...state, currentCatalogId: null, selected: null };
    case 'REQUEST_RAIL_REFRESH':
      return { ...state, railRefreshToken: state.railRefreshToken + 1 };
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
    case 'SET_CREATE_OPEN':
      // Returning the same state object when the value is unchanged is what
      // makes opening an already-open slide-over a true no-op (CRT-01
      // idempotency) -- React's reducer bail-out skips the re-render.
      if (state.createOpen === action.payload) return state;
      return { ...state, createOpen: action.payload };
    case 'SET_CREATE_FOLDER_PICKER_INTENT':
      if (state.createFolderPickerIntent === action.payload) return state;
      return { ...state, createFolderPickerIntent: action.payload };
    case 'SCAN_STARTED':
      return {
        ...state,
        scan: {
          status: 'counting',
          title: action.payload.title,
          filesSeen: 0,
          startedAt: Date.now(),
          currentPath: '',
          log: [],
          readErrors: 0,
        },
      };
    case 'SCAN_PROGRESS': {
      // A late event after the scan has already reached a terminal state
      // (or was reset to idle) must never resurrect it -- CRT-07
      // concurrency.
      if (state.scan.status !== 'counting' && state.scan.status !== 'scanning') return state;

      const prev = state.scan;
      const incoming = action.payload;
      // Clamped to the max of the previous and incoming value so an
      // out-of-order event can never make a counter go backwards (CRT-07
      // concurrency).
      const filesSeen = Math.max(prev.filesSeen, incoming.filesSeen);
      const prevBytesSeen = prev.status === 'scanning' ? prev.bytesSeen : 0;
      const bytesSeen = Math.max(prevBytesSeen, incoming.bytesSeen);
      // Both variants track readErrors now (added 25-07) -- unlike
      // bytesSeen, which is genuinely meaningless before a total is known,
      // a read-error count is real and worth keeping even during the
      // counting sub-state, since a scan that fails before ever resolving a
      // total still needs an accurate count for the error state.
      const readErrors = Math.max(prev.readErrors, incoming.readErrors);
      // The WALKING line always shows the newest known path; an event
      // carrying no path (never expected, but harmless if it happened)
      // leaves it unchanged rather than blanking it. A new path is pushed
      // onto the log only when it actually differs from the current top
      // line, so a repeated event can't duplicate the same entry.
      const currentPath = incoming.path || prev.currentPath;
      const log =
        incoming.path && incoming.path !== prev.log[0]
          ? [incoming.path, ...prev.log].slice(0, SCAN_LOG_CAP)
          : prev.log;

      if (incoming.totalBytes > 0) {
        return {
          ...state,
          scan: {
            status: 'scanning',
            title: prev.title,
            filesSeen,
            bytesSeen,
            totalBytes: incoming.totalBytes,
            readErrors,
            startedAt: prev.startedAt,
            currentPath,
            log,
          },
        };
      }

      return {
        ...state,
        scan: {
          status: 'counting',
          title: prev.title,
          filesSeen,
          startedAt: prev.startedAt,
          currentPath,
          log,
          readErrors,
        },
      };
    }
    case 'SCAN_FAILED': {
      // Snapshotted from whatever the scan state held the instant it
      // failed -- 'idle'/'done' have neither field, and TypeScript's
      // narrowing needs the explicit status checks below (not just `in`)
      // since only 'counting'/'scanning' carry filesSeen/readErrors here.
      const prev = state.scan;
      const filesSeen = prev.status === 'counting' || prev.status === 'scanning' ? prev.filesSeen : 0;
      const readErrors = prev.status === 'counting' || prev.status === 'scanning' ? prev.readErrors : 0;
      const stopPercent = prev.status === 'scanning' ? scanPercent(prev.bytesSeen, prev.totalBytes) : null;
      return {
        ...state,
        scan: {
          status: 'error',
          title: scanTitleOf(state.scan),
          message: action.payload.message,
          sourcePath: action.payload.sourcePath,
          filesSeen,
          stopPercent,
          readErrors,
        },
      };
    }
    case 'SCAN_DONE':
      return {
        ...state,
        scan: {
          status: 'done',
          title: action.payload.title,
          jsonPath: action.payload.jsonPath,
          files: action.payload.files,
          fileCount: action.payload.fileCount,
          totalSize: action.payload.totalSize,
          durationMs: action.payload.durationMs,
          partial: action.payload.partial,
          stopPercent: action.payload.stopPercent,
        },
      };
    case 'SCAN_RESET':
      return { ...state, scan: { status: 'idle' } };
    case 'SETTINGS_HYDRATED': {
      // Field-aware merge, not a wholesale replace: hydrateSettings()'s
      // getConfig() round trip is in flight from mount, so a user can
      // dispatch SET_SETTINGS (toggling a row in the Settings dialog,
      // already persisted via its own wailsAPI call) before this resolves.
      // Only fold in a hydrated field if SET_SETTINGS has never touched it
      // (WR-B) -- an equals-default check alone misfires when the user's
      // own write happens to land back on the default value.
      const merged = { ...state.settings };
      (Object.keys(action.payload) as (keyof AppSettings)[]).forEach((key) => {
        if (!state.touchedSettings.has(key)) {
          merged[key] = action.payload[key] as never;
        }
      });
      return { ...state, settings: merged };
    }
    case 'SET_SETTINGS': {
      // Returns the identical state object when every key in the payload
      // already holds the incoming value AND is already recorded touched --
      // the same bail-out convention SET_CREATE_OPEN/SET_CATALOG_DIR use
      // above, so a genuine repeat dispatch triggers no re-render. A value
      // that happens to be unchanged but not yet touched (the user's first
      // write landing back on the current/default value, WR-B) still needs
      // to record the touch so SETTINGS_HYDRATED can't later clobber it.
      const keys = Object.keys(action.payload) as (keyof AppSettings)[];
      const alreadyTouched = keys.every((key) => state.touchedSettings.has(key));
      const unchanged = keys.every((key) => state.settings[key] === action.payload[key]);
      if (unchanged && alreadyTouched) return state;
      const touchedSettings = alreadyTouched
        ? state.touchedSettings
        : new Set([...state.touchedSettings, ...keys]);
      return { ...state, settings: { ...state.settings, ...action.payload }, touchedSettings };
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
