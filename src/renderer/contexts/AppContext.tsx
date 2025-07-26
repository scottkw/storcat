import React, { createContext, useContext, useReducer, ReactNode } from 'react';

// Types
export interface AppState {
  selectedDirectory: string | null;
  selectedOutputDirectory: string | null;
  selectedSearchDirectory: string | null;
  selectedBrowseDirectory: string | null;
  isCreating: boolean;
  isSearching: boolean;
  isLoading: boolean;
  searchResults: any[];
  sortColumn: string | null;
  sortDirection: 'asc' | 'desc';
  browseCatalogs: any[];
  browseSortColumn: string | null;
  browseSortDirection: 'asc' | 'desc';
  sidebarCollapsed: boolean;
  activeTab: string;
  catalogModalOpen: boolean;
  catalogModalTitle: string;
  catalogModalHtmlPath: string;
}

type AppAction =
  | { type: 'SET_SELECTED_DIRECTORY'; payload: string | null }
  | { type: 'SET_SELECTED_OUTPUT_DIRECTORY'; payload: string | null }
  | { type: 'SET_SELECTED_SEARCH_DIRECTORY'; payload: string | null }
  | { type: 'SET_SELECTED_BROWSE_DIRECTORY'; payload: string | null }
  | { type: 'SET_IS_CREATING'; payload: boolean }
  | { type: 'SET_IS_SEARCHING'; payload: boolean }
  | { type: 'SET_IS_LOADING'; payload: boolean }
  | { type: 'SET_SEARCH_RESULTS'; payload: any[] }
  | { type: 'SET_SORT_COLUMN'; payload: string | null }
  | { type: 'SET_SORT_DIRECTION'; payload: 'asc' | 'desc' }
  | { type: 'SET_BROWSE_CATALOGS'; payload: any[] }
  | { type: 'SET_BROWSE_SORT_COLUMN'; payload: string | null }
  | { type: 'SET_BROWSE_SORT_DIRECTION'; payload: 'asc' | 'desc' }
  | { type: 'SET_SIDEBAR_COLLAPSED'; payload: boolean }
  | { type: 'SET_ACTIVE_TAB'; payload: string }
  | { type: 'OPEN_CATALOG_MODAL'; payload: { title: string; htmlPath: string } }
  | { type: 'CLOSE_CATALOG_MODAL' };

const initialState: AppState = {
  selectedDirectory: null,
  selectedOutputDirectory: null,
  selectedSearchDirectory: null,
  selectedBrowseDirectory: null,
  isCreating: false,
  isSearching: false,
  isLoading: false,
  searchResults: [],
  sortColumn: null,
  sortDirection: 'asc',
  browseCatalogs: [],
  browseSortColumn: null,
  browseSortDirection: 'asc',
  sidebarCollapsed: false,
  activeTab: 'create',
  catalogModalOpen: false,
  catalogModalTitle: '',
  catalogModalHtmlPath: '',
};

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case 'SET_SELECTED_DIRECTORY':
      return { ...state, selectedDirectory: action.payload };
    case 'SET_SELECTED_OUTPUT_DIRECTORY':
      return { ...state, selectedOutputDirectory: action.payload };
    case 'SET_SELECTED_SEARCH_DIRECTORY':
      return { ...state, selectedSearchDirectory: action.payload };
    case 'SET_SELECTED_BROWSE_DIRECTORY':
      return { ...state, selectedBrowseDirectory: action.payload };
    case 'SET_IS_CREATING':
      return { ...state, isCreating: action.payload };
    case 'SET_IS_SEARCHING':
      return { ...state, isSearching: action.payload };
    case 'SET_IS_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_SEARCH_RESULTS':
      return { ...state, searchResults: action.payload };
    case 'SET_SORT_COLUMN':
      return { ...state, sortColumn: action.payload };
    case 'SET_SORT_DIRECTION':
      return { ...state, sortDirection: action.payload };
    case 'SET_BROWSE_CATALOGS':
      return { ...state, browseCatalogs: action.payload };
    case 'SET_BROWSE_SORT_COLUMN':
      return { ...state, browseSortColumn: action.payload };
    case 'SET_BROWSE_SORT_DIRECTION':
      return { ...state, browseSortDirection: action.payload };
    case 'SET_SIDEBAR_COLLAPSED':
      return { ...state, sidebarCollapsed: action.payload };
    case 'SET_ACTIVE_TAB':
      return { ...state, activeTab: action.payload };
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