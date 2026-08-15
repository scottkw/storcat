// Write-through settings module -- the architectural answer to
// 26-RESEARCH.md Pitfall 1: main.tsx must read theme/density/rail-side
// *synchronously* before first paint (initThemeTokens in themeTokens.ts) to
// avoid the launch flash THEME-06 fixed, but Wails bindings are inherently
// asynchronous. Every setter here writes BOTH stores in the same
// synchronous handler, in this order: (1) the localStorage boot cache
// initThemeTokens() reads next launch, (2) the durable Go config via
// wailsAPI. Both happen in the same tick from the same user event, so the
// two stores cannot diverge under normal operation. No timer-based
// batching layer is added around either write -- every change writes
// immediately (26-CONTEXT.md locked decision; SET-05 needs no save step).
//
// Every later setter this phase adds (theme, rail side, catalog directory,
// filename root, the four toggles) copies this exact two-write shape.

import { Density, DENSITY_KEY, RailSide, RAIL_SIDE_KEY, THEME_KEY, safeSetItem } from './themeTokens';
import { wailsAPI } from './services/wailsAPI';
import { Theme } from './themes';

export { DENSITY_KEY };

export function setDensitySetting(density: Density): void {
  safeSetItem(DENSITY_KEY, density);
  void wailsAPI.setDensity(density);
}

// Dispatches the existing 'themeChange' CustomEvent rather than calling
// applyTokens() directly -- App.tsx's listener is the one existing apply
// path (also used by DevStateSwitcher.tsx) and this deliberately does not
// become a second one (26-CONTEXT.md Open Question 1 resolution). Note:
// App.tsx's listener writes THEME_KEY a second time -- that duplicate write
// is harmless and left alone rather than restructuring App.tsx, which plan
// 26-04 owns.
export function setThemeSetting(theme: Theme): void {
  safeSetItem(THEME_KEY, theme.id);
  void wailsAPI.setTheme(theme.id);
  window.dispatchEvent(new CustomEvent('themeChange', { detail: { theme } }));
}

export function setRailSideSetting(side: RailSide): void {
  safeSetItem(RAIL_SIDE_KEY, side);
  void wailsAPI.setRailSide(side);
}

// The one home for the catalog-directory storage-key literal -- CatalogRail
// and CatalogSettingsSection both import it from here rather than each
// declaring their own copy.
export const CATALOG_DIR_KEY = 'storcat-catalog-directory';

export function setCatalogDirectorySetting(dir: string): void {
  safeSetItem(CATALOG_DIR_KEY, dir);
  void wailsAPI.setCatalogDirectory(dir);
}

// The complete AppSettings shape, defined once here rather than widened by
// each later plan -- writeHtml/copyToSecondary/secondaryDirectory/
// watchDirectory/rememberWindow are read by plan 26-05's toggles section;
// rememberWindow maps onto the pre-existing windowPersistenceEnabled config
// field, not a new one.
export interface AppSettings {
  defaultFilenameRoot: string;
  writeHtml: boolean;
  copyToSecondary: boolean;
  secondaryDirectory: string;
  watchDirectory: boolean;
  rememberWindow: boolean;
}

// Matches Go's DefaultConfig() for every field that exists today.
export const DEFAULT_APP_SETTINGS: AppSettings = {
  defaultFilenameRoot: '',
  writeHtml: true,
  copyToSecondary: false,
  secondaryDirectory: '',
  watchDirectory: false,
  rememberWindow: true,
};

// Deliberately config-only -- no localStorage cache. The cache exists
// solely for the values initThemeTokens() and the rail read synchronously
// before an async binding could answer (theme/density/rail side/catalog
// directory); nothing reads the filename root pre-paint, so adding a sixth
// cached key here would re-create the two-store problem this phase exists
// to remove.
export function setDefaultFilenameRootSetting(root: string): void {
  void wailsAPI.setDefaultFilenameRoot(root);
}
