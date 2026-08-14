import { Theme, getThemeById, getDefaultTheme } from './themes';

export type Density = 'Compact' | 'Comfortable';
export type RailSide = 'Left' | 'Right';

export const THEME_KEY = 'storcat-theme-id';
export const DENSITY_KEY = 'storcat-density';
export const RAIL_SIDE_KEY = 'storcat-rail-side';

// The 14 extended theme tokens (THEME-03), enumerated once so the set cannot
// drift between computeTokens/applyTokens and any future consumer.
const TOKEN_NAMES = [
  '--bg', '--p', '--p2', '--ch', '--l', '--l2', '--tx', '--dm', '--fn',
  '--sel', '--hov', '--ac', '--acs', '--onac',
] as const;

/**
 * Relative luminance (WCAG), ported verbatim from
 * design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html:618-621.
 * Uses the WCAG 0.03928 transfer-function threshold -- distinct from the
 * 0.04045 threshold used by the OKLab path below; do not unify the two.
 */
export function lum(hex: string): number {
  const h = hex.replace('#', '');
  const v = [0, 2, 4]
    .map(i => parseInt(h.slice(i, i + 2), 16) / 255)
    .map(x => (x <= 0.03928 ? x / 12.92 : Math.pow((x + 0.055) / 1.055, 2.4)));
  return 0.2126 * v[0] + 0.7152 * v[1] + 0.0722 * v[2];
}

function hexToRgb(hex: string): { r: number; g: number; b: number } {
  const h = hex.replace('#', '');
  return {
    r: parseInt(h.slice(0, 2), 16),
    g: parseInt(h.slice(2, 4), 16),
    b: parseInt(h.slice(4, 6), 16),
  };
}

function srgbToLinear(c: number): number {
  const x = c / 255;
  return x <= 0.04045 ? x / 12.92 : Math.pow((x + 0.055) / 1.055, 2.4);
}

function linearToSrgb(x: number): number {
  const v = x <= 0.0031308 ? 12.92 * x : 1.055 * Math.pow(x, 1 / 2.4) - 0.055;
  return Math.round(Math.min(255, Math.max(0, v * 255)));
}

// OKLab conversion, Björn Ottosson's matrices.
// https://bottosson.github.io/posts/oklab/
function linearRgbToOklab(r: number, g: number, b: number): { L: number; a: number; b: number } {
  const l = 0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b;
  const m = 0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b;
  const s = 0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b;

  const l_ = Math.cbrt(l);
  const m_ = Math.cbrt(m);
  const s_ = Math.cbrt(s);

  return {
    L: 0.2104542553 * l_ + 0.7936177850 * m_ - 0.0040720468 * s_,
    a: 1.9779984951 * l_ - 2.4285922050 * m_ + 0.4505937099 * s_,
    b: 0.0259040371 * l_ + 0.7827717662 * m_ - 0.8086757660 * s_,
  };
}

function oklabToLinearRgb(L: number, a: number, b: number): { r: number; g: number; b: number } {
  const l_ = L + 0.3963377774 * a + 0.2158037573 * b;
  const m_ = L - 0.1055613458 * a - 0.0638541728 * b;
  const s_ = L - 0.0894841775 * a - 1.2914855480 * b;

  const l = l_ * l_ * l_;
  const m = m_ * m_ * m_;
  const s = s_ * s_ * s_;

  return {
    r: 4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
    g: -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
    b: -0.0041960863 * l - 0.7034186147 * m + 1.7076147010 * s,
  };
}

/**
 * Perceptual OKLab blend, TS equivalent of the prototype's
 * `color-mix(in oklab, a pct%, b)`. pct = 100 returns a, pct = 0 returns b.
 * color-mix() itself is banned at runtime (WebKitGTK version risk on Linux),
 * so the blend is precomputed here to a concrete rgb() string.
 */
export function mixHex(a: string, pct: number, b: string): string {
  const ca = hexToRgb(a);
  const cb = hexToRgb(b);

  const laba = linearRgbToOklab(srgbToLinear(ca.r), srgbToLinear(ca.g), srgbToLinear(ca.b));
  const labb = linearRgbToOklab(srgbToLinear(cb.r), srgbToLinear(cb.g), srgbToLinear(cb.b));

  const t = pct / 100;
  const L = laba.L * t + labb.L * (1 - t);
  const A = laba.a * t + labb.a * (1 - t);
  const B = laba.b * t + labb.b * (1 - t);

  const lin = oklabToLinearRgb(L, A, B);
  const r = linearToSrgb(lin.r);
  const g = linearToSrgb(lin.g);
  const bl = linearToSrgb(lin.b);

  return `rgb(${r}, ${g}, ${bl})`;
}

/** Blend-with-transparent case -- source channels unchanged, alpha = pct/100. */
export function mixAlpha(hex: string, pct: number): string {
  const { r, g, b } = hexToRgb(hex);
  return `rgba(${r}, ${g}, ${b}, ${pct / 100})`;
}

const DENSITY_VARS: Record<Density, Record<string, string>> = {
  Compact: { '--rh': '27px', '--rp': '6px 8px', '--mp': '6px', '--hp': '7px 14px', '--fs': '12px' },
  Comfortable: { '--rh': '34px', '--rp': '10px 10px', '--mp': '10px', '--hp': '11px 14px', '--fs': '13px' },
};

/** TS port of the prototype's varsFor() + density branch. */
export function computeTokens(theme: Theme, density: Density): Record<string, string> {
  const t = theme.tokens;

  const values: Record<string, string> = {
    '--bg': t.bg,
    '--p': t.p,
    '--p2': t.p2,
    '--ch': t.ch,
    '--l': t.l,
    '--l2': mixHex(t.l, 55, t.p),
    '--tx': t.tx,
    '--dm': mixHex(t.tx, 66, t.bg),
    '--fn': mixHex(t.tx, 44, t.bg),
    '--sel': mixAlpha(t.ac, 14),
    '--hov': mixHex(t.tx, 8, t.bg),
    '--ac': t.ac,
    '--acs': mixAlpha(t.ac, 16),
    '--onac': lum(t.ac) > 0.45 ? '#0b0e13' : '#ffffff',
  };

  if (import.meta.env.DEV) {
    const got = Object.keys(values).sort().join(',');
    const want = [...TOKEN_NAMES].sort().join(',');
    if (got !== want) {
      throw new Error(`computeTokens token set drifted from TOKEN_NAMES: got [${got}] want [${want}]`);
    }
  }

  return { ...values, ...DENSITY_VARS[density] };
}

/**
 * Single successor to App.tsx's former applyTheme(): writes the 14 extended
 * tokens + 5 density vars, and keeps writing the 16 legacy ThemeColors
 * properties so index.css's antd overrides (CatalogModal) keep working.
 */
export function applyTokens(theme: Theme, density: Density): void {
  const root = document.documentElement;
  root.setAttribute('data-theme', theme.id);

  const computed = computeTokens(theme, density);
  for (const [name, value] of Object.entries(computed)) {
    root.style.setProperty(name, value);
  }

  const legacy = theme.colors;
  root.style.setProperty('--app-bg', legacy.appBg);
  root.style.setProperty('--app-text', legacy.appText);
  root.style.setProperty('--card-bg', legacy.cardBg);
  root.style.setProperty('--border-color', legacy.borderColor);
  root.style.setProperty('--header-bg', legacy.headerBg);
  root.style.setProperty('--header-text', legacy.headerText);
  root.style.setProperty('--sidebar-bg', legacy.sidebarBg);
  root.style.setProperty('--table-stripe', legacy.tableStripe);
  root.style.setProperty('--table-hover', legacy.tableHover);
  root.style.setProperty('--modal-bg', legacy.modalBg);
  root.style.setProperty('--shadow-color', legacy.shadowColor);
  root.style.setProperty('--input-bg', legacy.inputBg);
  root.style.setProperty('--code-bg', legacy.codeBg);
  root.style.setProperty('--icon-filter', legacy.iconFilter);
  root.style.setProperty('--link-color', legacy.linkColor);
  root.style.setProperty('--link-hover', legacy.linkHover);
}

// Exported so other modules persisting their own `storcat-*` keys (e.g. the
// rail's catalog-directory chip in plan 23-04) reuse this try/catch wrapper
// instead of a second copy -- the pattern this file established, not a new one.
export function safeGetItem(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function safeSetItem(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch {
    // Storage unavailable/throwing -- fall through silently, defaults still apply in memory.
  }
}

function safeRemoveItem(key: string): void {
  try {
    localStorage.removeItem(key);
  } catch {
    // no-op
  }
}

/**
 * Reads and validates the three persisted preference keys against strict
 * allowlists. A missing, unknown, or corrupt value falls back to the locked
 * defaults (storcat-light / Comfortable / Left) and rewrites a valid value.
 * Never throws -- a throwing storage backend yields defaults.
 */
export function readPersistedPrefs(): { theme: Theme; density: Density; railSide: RailSide } {
  // Theme, with the pre-existing storcat-theme -> storcat-theme-id migration.
  let themeId = safeGetItem(THEME_KEY);
  if (!themeId) {
    const oldTheme = safeGetItem('storcat-theme');
    if (oldTheme === 'dark') {
      themeId = 'storcat-dark';
    } else {
      themeId = 'storcat-light';
    }
    safeSetItem(THEME_KEY, themeId);
    safeRemoveItem('storcat-theme');
  }

  const resolvedTheme = getThemeById(themeId) ?? getDefaultTheme();
  if (resolvedTheme.id !== themeId) {
    safeSetItem(THEME_KEY, resolvedTheme.id);
  }

  // Density: exact-string allowlist only.
  const storedDensity = safeGetItem(DENSITY_KEY);
  const density: Density =
    storedDensity === 'Compact' || storedDensity === 'Comfortable' ? storedDensity : 'Comfortable';
  if (storedDensity !== density) {
    safeSetItem(DENSITY_KEY, density);
  }

  // Rail side: exact-string allowlist only.
  const storedRailSide = safeGetItem(RAIL_SIDE_KEY);
  const railSide: RailSide = storedRailSide === 'Left' || storedRailSide === 'Right' ? storedRailSide : 'Left';
  if (storedRailSide !== railSide) {
    safeSetItem(RAIL_SIDE_KEY, railSide);
  }

  return { theme: resolvedTheme, density, railSide };
}

/**
 * Reads persisted prefs, applies tokens to :root, and returns the resolved
 * triple. Must run synchronously at module scope in main.tsx, before
 * createRoot's render call -- a post-mount effect fires after first paint
 * and reintroduces the launch flash.
 */
export function initThemeTokens(): { theme: Theme; density: Density; railSide: RailSide } {
  const prefs = readPersistedPrefs();
  applyTokens(prefs.theme, prefs.density);
  document.documentElement.setAttribute('data-rail-side', prefs.railSide);
  return prefs;
}
