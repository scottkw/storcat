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
  | { type: 'TREE_FAILED'; payload: { catalogId: string; message: string } };

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
      return { ...state, catalogDir: action.payload };
    case 'SET_CATALOGS':
      return { ...state, catalogs: action.payload };
    case 'SELECT_CATALOG':
      // Atomic: sets currentCatalogId, starts the tree load, and clears
      // expanded/selected together -- no intermediate state where one is
      // cleared and the other is not (TREE-06).
      return {
        ...state,
        currentCatalogId: action.payload,
        tree: { status: 'loading' },
        expanded: {},
        selected: null,
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
