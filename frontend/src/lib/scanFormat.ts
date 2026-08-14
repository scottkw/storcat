/**
 * scanPercent and formatEta are the create flow's single source of truth
 * for percentage/ETA text -- the slide-over's scanning body and any later
 * plan's status-bar segment both import from here, so their numbers can
 * never independently drift (25-UI-SPEC.md's "one consistent rounding rule
 * everywhere" contract).
 */

/**
 * Returns null when totalBytes is zero or negative -- the denominator is
 * unknown, i.e. the counting sub-state -- otherwise a single rounded
 * integer clamped so it can never read 100 while bytes remain outstanding
 * (the project-wide "progress is always a real number" rule: 100% must mean
 * done, never "close enough").
 */
export function scanPercent(bytesSeen: number, totalBytes: number): number | null {
  if (totalBytes <= 0) return null;
  const pct = Math.round((bytesSeen / totalBytes) * 100);
  if (pct >= 100 && bytesSeen < totalBytes) return 99;
  return Math.min(Math.max(pct, 0), 100);
}

/**
 * Returns "finishing…" once bytesSeen has caught up to totalBytes (or the
 * denominator is unknown -- there is nothing to estimate against), otherwise
 * "about {N}s left" computed from the throughput observed so far.
 */
export function formatEta(bytesSeen: number, totalBytes: number, elapsedMs: number): string {
  if (totalBytes <= 0 || bytesSeen >= totalBytes) return 'finishing…';
  if (elapsedMs <= 0 || bytesSeen <= 0) return 'about …s left';
  const bytesPerMs = bytesSeen / elapsedMs;
  const remainingMs = (totalBytes - bytesSeen) / bytesPerMs;
  const remainingSeconds = Math.max(1, Math.round(remainingMs / 1000));
  return `about ${remainingSeconds}s left`;
}
