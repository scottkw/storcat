export interface ThemeColors {
  // App-level colors
  appBg: string;
  appText: string;
  cardBg: string;
  borderColor: string;
  
  // Header colors
  headerBg: string;
  headerText: string;
  
  // Sidebar colors
  sidebarBg: string;
  
  // Table colors
  tableStripe: string;
  tableHover: string;
  
  // Modal colors
  modalBg: string;
  shadowColor: string;
  
  // Input colors
  inputBg: string;
  codeBg: string;
  
  // Link colors
  linkColor: string;
  linkHover: string;
  
  // Icon filter
  iconFilter: string;
}

export interface Theme {
  id: string;
  name: string;
  type: 'light' | 'dark';
  colors: ThemeColors;
  antdAlgorithm: 'default' | 'dark';
  antdPrimaryColor?: string;
}

export const themes: Theme[] = [
  // StorCat Light (Original)
  {
    id: 'storcat-light',
    name: 'StorCat Light',
    type: 'light',
    antdAlgorithm: 'default',
    antdPrimaryColor: '#5D6569FF',
    colors: {
      appBg: '#f8f9fa',
      appText: '#212529',
      cardBg: '#ffffff',
      borderColor: '#dee2e6',
      headerBg: '#5D6569FF',
      headerText: 'white',
      sidebarBg: '#ffffff',
      tableStripe: '#f8f9fa',
      tableHover: '#e9ecef',
      modalBg: 'rgba(0, 0, 0, 0.5)',
      shadowColor: 'rgba(0, 0, 0, 0.1)',
      inputBg: '#ffffff',
      codeBg: '#f8f9fa',
      linkColor: '#1d4ed8',
      linkHover: '#1e40af',
      iconFilter: 'none',
    },
  },

  // StorCat Dark (Original)
  {
    id: 'storcat-dark',
    name: 'StorCat Dark',
    type: 'dark',
    antdAlgorithm: 'dark',
    antdPrimaryColor: '#5D6569FF',
    colors: {
      appBg: '#0f172a',
      appText: '#f1f5f9',
      cardBg: '#1e293b',
      borderColor: '#334155',
      headerBg: '#5D6569FF',
      headerText: '#f8fafc',
      sidebarBg: '#1e293b',
      tableStripe: '#2d3b4e',
      tableHover: '#3b4d63',
      modalBg: 'rgba(0, 0, 0, 0.8)',
      shadowColor: 'rgba(0, 0, 0, 0.5)',
      inputBg: '#334155',
      codeBg: '#334155',
      linkColor: '#60a5fa',
      linkHover: '#93c5fd',
      iconFilter: 'brightness(1.1) contrast(1.1)',
    },
  },

  // Dracula
  {
    id: 'dracula',
    name: 'Dracula',
    type: 'dark',
    antdAlgorithm: 'dark',
    antdPrimaryColor: '#bd93f9',
    colors: {
      appBg: '#282a36',
      appText: '#f8f8f2',
      cardBg: '#44475a',
      borderColor: '#6272a4',
      headerBg: '#8be9fd',
      headerText: '#282a36',
      sidebarBg: '#44475a',
      tableStripe: '#44475a',
      tableHover: '#6272a4',
      modalBg: 'rgba(40, 42, 54, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.6)',
      inputBg: '#44475a',
      codeBg: '#44475a',
      linkColor: '#8be9fd',
      linkHover: '#50fa7b',
      iconFilter: 'brightness(1.2) contrast(1.1)',
    },
  },

  // Solarized Dark
  {
    id: 'solarized-dark',
    name: 'Solarized Dark',
    type: 'dark',
    antdAlgorithm: 'dark',
    antdPrimaryColor: '#268bd2',
    colors: {
      appBg: '#002b36',
      appText: '#839496',
      cardBg: '#073642',
      borderColor: '#586e75',
      headerBg: '#268bd2',
      headerText: '#fdf6e3',
      sidebarBg: '#073642',
      tableStripe: '#073642',
      tableHover: '#586e75',
      modalBg: 'rgba(0, 43, 54, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.7)',
      inputBg: '#073642',
      codeBg: '#073642',
      linkColor: '#268bd2',
      linkHover: '#2aa198',
      iconFilter: 'brightness(1.3) contrast(1.2)',
    },
  },

  // Solarized Light
  {
    id: 'solarized-light',
    name: 'Solarized Light',
    type: 'light',
    antdAlgorithm: 'default',
    antdPrimaryColor: '#268bd2',
    colors: {
      appBg: '#fdf6e3',
      appText: '#657b83',
      cardBg: '#eee8d5',
      borderColor: '#93a1a1',
      headerBg: '#268bd2',
      headerText: '#fdf6e3',
      sidebarBg: '#eee8d5',
      tableStripe: '#eee8d5',
      tableHover: '#93a1a1',
      modalBg: 'rgba(253, 246, 227, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.2)',
      inputBg: '#eee8d5',
      codeBg: '#eee8d5',
      linkColor: '#268bd2',
      linkHover: '#2aa198',
      iconFilter: 'none',
    },
  },

  // Nord
  {
    id: 'nord',
    name: 'Nord',
    type: 'dark',
    antdAlgorithm: 'dark',
    antdPrimaryColor: '#5e81ac',
    colors: {
      appBg: '#2e3440',
      appText: '#d8dee9',
      cardBg: '#3b4252',
      borderColor: '#4c566a',
      headerBg: '#5e81ac',
      headerText: '#eceff4',
      sidebarBg: '#3b4252',
      tableStripe: '#3b4252',
      tableHover: '#434c5e',
      modalBg: 'rgba(46, 52, 64, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.6)',
      inputBg: '#434c5e',
      codeBg: '#434c5e',
      linkColor: '#88c0d0',
      linkHover: '#8fbcbb',
      iconFilter: 'brightness(1.2) contrast(1.1)',
    },
  },

  // One Dark
  {
    id: 'one-dark',
    name: 'One Dark',
    type: 'dark',
    antdAlgorithm: 'dark',
    antdPrimaryColor: '#61afef',
    colors: {
      appBg: '#282c34',
      appText: '#abb2bf',
      cardBg: '#21252b',
      borderColor: '#3e4451',
      headerBg: '#61afef',
      headerText: '#282c34',
      sidebarBg: '#21252b',
      tableStripe: '#21252b',
      tableHover: '#2c313c',
      modalBg: 'rgba(40, 44, 52, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.6)',
      inputBg: '#2c313c',
      codeBg: '#2c313c',
      linkColor: '#61afef',
      linkHover: '#56b6c2',
      iconFilter: 'brightness(1.2) contrast(1.1)',
    },
  },

  // Monokai
  {
    id: 'monokai',
    name: 'Monokai',
    type: 'dark',
    antdAlgorithm: 'dark',
    antdPrimaryColor: '#fd971f',
    colors: {
      appBg: '#272822',
      appText: '#f8f8f2',
      cardBg: '#3e3d32',
      borderColor: '#49483e',
      headerBg: '#fd971f',
      headerText: '#272822',
      sidebarBg: '#3e3d32',
      tableStripe: '#3e3d32',
      tableHover: '#49483e',
      modalBg: 'rgba(39, 40, 34, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.7)',
      inputBg: '#49483e',
      codeBg: '#49483e',
      linkColor: '#66d9ef',
      linkHover: '#a6e22e',
      iconFilter: 'brightness(1.3) contrast(1.2)',
    },
  },

  // GitHub Light
  {
    id: 'github-light',
    name: 'GitHub Light',
    type: 'light',
    antdAlgorithm: 'default',
    antdPrimaryColor: '#0969da',
    colors: {
      appBg: '#ffffff',
      appText: '#24292f',
      cardBg: '#f6f8fa',
      borderColor: '#d0d7de',
      headerBg: '#0969da',
      headerText: '#ffffff',
      sidebarBg: '#f6f8fa',
      tableStripe: '#f6f8fa',
      tableHover: '#f3f4f6',
      modalBg: 'rgba(255, 255, 255, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.12)',
      inputBg: '#ffffff',
      codeBg: '#f6f8fa',
      linkColor: '#0969da',
      linkHover: '#0550ae',
      iconFilter: 'none',
    },
  },

  // GitHub Dark
  {
    id: 'github-dark',
    name: 'GitHub Dark',
    type: 'dark',
    antdAlgorithm: 'dark',
    antdPrimaryColor: '#238636',
    colors: {
      appBg: '#0d1117',
      appText: '#c9d1d9',
      cardBg: '#161b22',
      borderColor: '#30363d',
      headerBg: '#238636',
      headerText: '#f0f6fc',
      sidebarBg: '#161b22',
      tableStripe: '#161b22',
      tableHover: '#21262d',
      modalBg: 'rgba(13, 17, 23, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.7)',
      inputBg: '#21262d',
      codeBg: '#21262d',
      linkColor: '#58a6ff',
      linkHover: '#79c0ff',
      iconFilter: 'brightness(1.2) contrast(1.1)',
    },
  },

  // Gruvbox Dark
  {
    id: 'gruvbox-dark',
    name: 'Gruvbox Dark',
    type: 'dark',
    antdAlgorithm: 'dark',
    antdPrimaryColor: '#fe8019',
    colors: {
      appBg: '#282828',
      appText: '#ebdbb2',
      cardBg: '#3c3836',
      borderColor: '#504945',
      headerBg: '#fe8019',
      headerText: '#282828',
      sidebarBg: '#3c3836',
      tableStripe: '#3c3836',
      tableHover: '#504945',
      modalBg: 'rgba(40, 40, 40, 0.9)',
      shadowColor: 'rgba(0, 0, 0, 0.7)',
      inputBg: '#504945',
      codeBg: '#504945',
      linkColor: '#83a598',
      linkHover: '#8ec07c',
      iconFilter: 'brightness(1.3) contrast(1.2)',
    },
  },
];

export const getThemeById = (id: string): Theme | undefined => {
  return themes.find(theme => theme.id === id);
};

export const getDefaultTheme = (): Theme => {
  return themes[0]; // StorCat Light
};