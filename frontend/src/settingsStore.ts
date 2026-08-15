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

import { Density, DENSITY_KEY, RailSide, RAIL_SIDE_KEY, THEME_KEY, safeGetItem, safeSetItem } from './themeTokens';
import { wailsAPI } from './services/wailsAPI';
import { Theme, getThemeById } from './themes';

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

// The one home for the secondary-directory storage-key literal --
// OptionsToggles and CreateSlideOver both import it from here.
export const SECONDARY_DIR_KEY = 'storcat-secondary-directory';

export function setSecondaryDirectorySetting(dir: string): void {
  safeSetItem(SECONDARY_DIR_KEY, dir);
  void wailsAPI.setSecondaryDirectory(dir);
}

// Config-only, same reasoning as setDefaultFilenameRootSetting above --
// nothing reads these three before first paint, so no boot cache is added
// for them.
export function setWriteHtmlSetting(enabled: boolean): void {
  void wailsAPI.setWriteHTML(enabled);
}

export function setCopyToSecondarySetting(enabled: boolean): void {
  void wailsAPI.setCopyToSecondary(enabled);
}

export function setWatchDirectorySetting(enabled: boolean): void {
  void wailsAPI.setWatchDirectory(enabled);
}

export interface HydratedSettings {
  settings: AppSettings;
  catalogDirectory: string;
}

// Deduped behind this module-level in-flight promise so React 18
// StrictMode's development double-invoke of the calling effect (and any
// other concurrent caller) can never run the migration below twice --
// every caller in the same tick (or while the first call is still in
// flight) shares the one getConfig() round trip and the one migration
// pass.
let hydratePromise: Promise<HydratedSettings | null> | null = null;

export function hydrateSettings(): Promise<HydratedSettings | null> {
  if (!hydratePromise) hydratePromise = doHydrate();
  return hydratePromise;
}

async function doHydrate(): Promise<HydratedSettings | null> {
  const initial = await wailsAPI.getConfig();
  // A config the app cannot read is not a reason to crash the shell -- the
  // localStorage boot cache already painted the right theme, and the app
  // stays usable, un-hydrated, for this launch.
  if (!initial.success) return null;
  let cfg = initial.config;

  // Migration gate: the persisted marker is the ONLY signal -- never a
  // comparison against a config field looking like a zero value (a real
  // user setting could legitimately be "" or the same as a default).
  if (!cfg.settingsMigrated) {
    // Each of the five cached keys is validated through the exact same
    // allowlist readPersistedPrefs() (themeTokens.ts) uses for the boot
    // read, before being written to a durable config field -- an invalid
    // or missing value is skipped, leaving the Go default in place
    // (T-26-09).
    const storedThemeId = safeGetItem(THEME_KEY);
    if (storedThemeId && getThemeById(storedThemeId)) {
      await wailsAPI.setTheme(storedThemeId);
    }
    const storedDensity = safeGetItem(DENSITY_KEY);
    if (storedDensity === 'Compact' || storedDensity === 'Comfortable') {
      await wailsAPI.setDensity(storedDensity);
    }
    const storedRailSide = safeGetItem(RAIL_SIDE_KEY);
    if (storedRailSide === 'Left' || storedRailSide === 'Right') {
      await wailsAPI.setRailSide(storedRailSide);
    }
    const storedCatalogDir = safeGetItem(CATALOG_DIR_KEY);
    if (storedCatalogDir) {
      await wailsAPI.setCatalogDirectory(storedCatalogDir);
    }
    const storedSecondaryDir = safeGetItem(SECONDARY_DIR_KEY);
    if (storedSecondaryDir) {
      await wailsAPI.setSecondaryDirectory(storedSecondaryDir);
    }
    // Never deletes a localStorage key -- they remain the synchronous boot
    // cache, and keeping them is also what makes a mis-migration
    // recoverable (clearing the marker re-runs the migration against
    // untouched source values).
    await wailsAPI.setSettingsMigrated(true);
    const reread = await wailsAPI.getConfig();
    if (reread.success) cfg = reread.config;
  }

  // Write back the other direction: config -> cache, only when the
  // config's own value passes that key's allowlist (T-26-10) -- so a
  // config edited outside the app (or by a future CLI) converges into the
  // boot cache on the next launch, and a stale/legacy value can never
  // poison the pre-paint read.
  if (getThemeById(cfg.theme)) safeSetItem(THEME_KEY, cfg.theme);
  if (cfg.density === 'Compact' || cfg.density === 'Comfortable') safeSetItem(DENSITY_KEY, cfg.density);
  if (cfg.railSide === 'Left' || cfg.railSide === 'Right') safeSetItem(RAIL_SIDE_KEY, cfg.railSide);
  if (cfg.catalogDirectory) safeSetItem(CATALOG_DIR_KEY, cfg.catalogDirectory);
  if (cfg.secondaryDirectory) safeSetItem(SECONDARY_DIR_KEY, cfg.secondaryDirectory);

  const settings: AppSettings = {
    defaultFilenameRoot: cfg.defaultFilenameRoot,
    writeHtml: cfg.writeHtml,
    copyToSecondary: cfg.copyToSecondary,
    secondaryDirectory: cfg.secondaryDirectory,
    watchDirectory: cfg.watchDirectory,
    // rememberWindow maps onto the pre-existing windowPersistenceEnabled
    // field -- there is no separate "remember window" config field.
    rememberWindow: cfg.windowPersistenceEnabled,
  };

  return { settings, catalogDirectory: cfg.catalogDirectory };
}
