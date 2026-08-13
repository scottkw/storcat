import React, { createContext, useContext, useReducer, ReactNode } from 'react';
import { Density, RailSide, readPersistedPrefs } from '../themeTokens';

// Types
export interface AppState {
  density: Density;
  railSide: RailSide;
  detailOverlay: boolean;
  catalogModalOpen: boolean;
  catalogModalTitle: string;
  catalogModalHtmlPath: string;
}

type AppAction =
  | { type: 'SET_DENSITY'; payload: Density }
  | { type: 'SET_RAIL_SIDE'; payload: RailSide }
  | { type: 'SET_DETAIL_OVERLAY'; payload: boolean }
  | { type: 'OPEN_CATALOG_MODAL'; payload: { title: string; htmlPath: string } }
  | { type: 'CLOSE_CATALOG_MODAL' };

// Seeded once at module scope so a relaunch restores the user's persisted
// density/rail-side choices rather than hardcoded defaults. readPersistedPrefs
// already applies the allowlist and locked fallbacks -- no second validation here.
const persistedPrefs = readPersistedPrefs();

const initialState: AppState = {
  density: persistedPrefs.density,
  railSide: persistedPrefs.railSide,
  detailOverlay: false,
  catalogModalOpen: false,
  catalogModalTitle: '',
  catalogModalHtmlPath: '',
};

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case 'SET_DENSITY':
      return { ...state, density: action.payload };
    case 'SET_RAIL_SIDE':
      return { ...state, railSide: action.payload };
    case 'SET_DETAIL_OVERLAY':
      return { ...state, detailOverlay: action.payload };
    case 'OPEN_CATALOG_MODAL':
      return {
        ...state,
        catalogModalOpen: true,
        catalogModalTitle: action.payload.title,
        catalogModalHtmlPath: action.payload.htmlPath
      };
    case 'CLOSE_CATALOG_MODAL':
      return {
        ...state,
        catalogModalOpen: false,
        catalogModalTitle: '',
        catalogModalHtmlPath: ''
      };
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
