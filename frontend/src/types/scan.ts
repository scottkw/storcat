/**
 * ScanProgress mirrors app.go's ScanProgress struct field-for-field (json
 * tags path/filesSeen/bytesSeen/readErrors/totalBytes) -- it is the event
 * payload shape for the "scan:progress" Wails event. The two are edited
 * together: a field added on one side with no counterpart on the other
 * silently breaks the live progress readout.
 */
export interface ScanProgress {
  path: string;
  filesSeen: number;
  bytesSeen: number;
  readErrors: number;
  totalBytes: number;
}

/**
 * One row of the done state's written-files list. size is optional: Go's
 * CreateCatalogResult reports the catalog's total scanned content size, not
 * each output file's own on-disk byte count, so a row with an unknown size
 * renders with no size text rather than a fabricated number (the project's
 * "no fabricated values" rule extends to this field).
 */
export interface ScanResultFile {
  path: string;
  size?: number;
}

/**
 * The full five-member state machine the create slide-over drives, per
 * 25-UI-SPEC.md's header table -- declared in full even though this plan
 * only produces idle/counting/done at runtime: App.StartScan's totalBytes
 * denominator is a seam a later plan fills in (sourceTotalBytes always
 * returns 0 this plan), so `scanning`'s percentage-known branch is reachable
 * in the type but never actually hit yet. `error` exists so the reducer's
 * SCAN_FAILED action has a target; this plan's form body doubles as the
 * recovery surface for it (see 25-01-SUMMARY.md) rather than building the
 * full CRT-10 error body.
 */
export type ScanState =
  | { status: 'idle' }
  | {
      status: 'counting';
      title: string;
      filesSeen: number;
      startedAt: number;
    }
  | {
      status: 'scanning';
      title: string;
      filesSeen: number;
      bytesSeen: number;
      totalBytes: number;
      readErrors: number;
      startedAt: number;
    }
  | { status: 'error'; title: string; message: string }
  | {
      status: 'done';
      title: string;
      jsonPath: string;
      files: ScanResultFile[];
      fileCount: number;
      totalSize: number;
      durationMs: number;
      partial: boolean;
    };

// SOURCE_LOSS_MARKER is the substring contract between the two sides of the
// bridge: app.go's catalog.SourceUnavailableError.Error() always reads
// "source unavailable: <path>" (internal/catalog/errors.go). A cancellation
// error carries no such substring (it propagates context.Canceled, wrapped
// or bare). If a future change ever edits the Go error text, this is the
// one place on the frontend that needs updating to match.
const SOURCE_LOSS_MARKER = 'source unavailable';

/**
 * ScanFailure discriminates the only two ways StartScan's promise can
 * reject once a scan is actually running: the user (or a window close)
 * cancelled it, or the scan root itself became unreachable mid-walk
 * (CRT-10). A cancellation must produce no error UI at all -- the panel
 * simply closes; only a source loss routes to the error body (plan 25-07)
 * offering a partial-catalog write (plan 25-07 also wires cancelScan into
 * the scanning body, plan 25-06).
 */
export type ScanFailure =
  | { kind: 'cancelled'; message: string }
  | { kind: 'sourceLoss'; message: string };

/**
 * classifyScanFailure inspects a StartScan rejection's message (already
 * unwrapped by wailsAPI's extractErrorMessage) and returns which of the two
 * ScanFailure kinds it represents. Pure -- no I/O, safe to call from a
 * reducer.
 */
export function classifyScanFailure(message: string): ScanFailure {
  if (message.includes(SOURCE_LOSS_MARKER)) {
    return { kind: 'sourceLoss', message };
  }
  return { kind: 'cancelled', message };
}
