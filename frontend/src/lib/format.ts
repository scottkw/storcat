/**
 * The phase's only number and date formatters. Plans 23-04, 23-05 and 23-06
 * import from here and must not need to edit it -- this surface is fixed by
 * plan 23-03.
 */

/**
 * Ports Go's `internal/catalog/service.go` `formatBytes` value for value:
 * zero renders as "0B", below 1024 renders as the integer plus "B", and
 * above that it divides by 1024 through K/M/G/T, one decimal place, with a
 * trailing ".0" stripped. Used by the rail row, tree row size column and
 * the catalog header metadata line.
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0B';

  const unit = 1024;
  const sizes = ['B', 'K', 'M', 'G', 'T'];

  if (bytes < unit) return `${bytes}B`;

  let div = unit;
  let exp = 0;
  for (let n = Math.floor(bytes / unit); n >= unit && exp < sizes.length - 1; n = Math.floor(n / unit)) {
    div *= unit;
    exp++;
  }

  const value = bytes / div;
  let formatted = value.toFixed(1);
  if (formatted.endsWith('.0')) formatted = formatted.slice(0, -2);

  return `${formatted}${sizes[exp + 1]}`;
}

/**
 * Renders a file count. Below one million it uses locale grouping (matching
 * the prototype's own `fmt` helper); at one million and above it abbreviates
 * to one decimal place followed by "M" so a status-bar segment growing into
 * the millions shortens rather than pushes another segment off a
 * fixed-height strip (23-UI-SPEC E5 overflow).
 */
export function formatCount(n: number): string {
  if (n < 1_000_000) return n.toLocaleString('en-US');
  return `${(n / 1_000_000).toFixed(1)}M`;
}

/**
 * Renders a byte total as the status bar's third segment. Divides once by
 * 1024^3 to get gigabytes, then follows the prototype's `gb` helper: at
 * 1024 and above it divides again and suffixes " TB", otherwise it suffixes
 * " GB" -- always one decimal place, threshold inclusive. Divides once from
 * the byte total; never sums values already rounded for display.
 */
export function formatGB(bytes: number): string {
  const gigabytes = bytes / 1024 ** 3;
  if (gigabytes >= 1024) {
    return `${(gigabytes / 1024).toFixed(1)} TB`;
  }
  return `${gigabytes.toFixed(1)} GB`;
}

/**
 * Renders a catalog's modification timestamp (an RFC3339 string the Go
 * layer already produces) as a short local date. An unparseable input
 * returns the original string unchanged rather than the word browsers
 * produce for an invalid date.
 */
export function formatDate(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  return date.toLocaleDateString();
}
