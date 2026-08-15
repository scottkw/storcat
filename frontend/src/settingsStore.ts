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

import { Density, DENSITY_KEY, safeSetItem } from './themeTokens';
import { wailsAPI } from './services/wailsAPI';

export { DENSITY_KEY };

export function setDensitySetting(density: Density): void {
  safeSetItem(DENSITY_KEY, density);
  void wailsAPI.setDensity(density);
}
