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

/**
 * Lowercases, collapses every run of non-alphanumeric characters to a
 * single hyphen, and trims leading/trailing hyphens -- the filename-root
 * placeholder's derivation from a source's display name. Pure; an empty or
 * all-symbol input returns an empty string, and the caller (CreateForm)
 * applies the final "catalog" fallback, matching the title field's own
 * never-block-Create rule of leaving fallback application to the caller.
 */
export function slugifyRoot(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

export interface WillWritePathsArgs {
  catalogDir: string;
  root: string;
  writeHtml: boolean;
  secondaryDir?: string;
}

/**
 * The WILL WRITE preview's exact output paths, in the fixed, deterministic
 * order the contract requires: the JSON in the catalog directory first,
 * then the HTML there when enabled, then the secondary-location copies (in
 * the same JSON-then-HTML order) when a secondary directory is set. Callers
 * render this array as given and never sort or reverse it.
 */
export function willWritePaths({ catalogDir, root, writeHtml, secondaryDir }: WillWritePathsArgs): string[] {
  const paths = [joinDisplayPath(catalogDir, `${root}.json`)];
  if (writeHtml) paths.push(joinDisplayPath(catalogDir, `${root}.html`));
  if (secondaryDir) {
    paths.push(joinDisplayPath(secondaryDir, `${root}.json`));
    if (writeHtml) paths.push(joinDisplayPath(secondaryDir, `${root}.html`));
  }
  return paths;
}

/** Joins a directory and filename for display, tolerant of a trailing
 * separator on the directory so the preview never shows a doubled slash. */
function joinDisplayPath(dir: string, filename: string): string {
  return `${dir.replace(/[\\/]+$/, '')}/${filename}`;
}
