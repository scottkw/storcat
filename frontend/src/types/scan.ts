import { volumes } from '../../wailsjs/go/models';

/**
 * The two ways a user can choose what to catalog (CRT-02/CRT-03): a
 * detected volume (from ListVolumes, carrying its own known size) or an
 * arbitrary folder chosen through the native picker (no volume-level probe
 * exists for a plain folder). Lifted here rather than kept as separate
 * fields so VolumePicker, CreateForm and CreateSlideOver all derive the
 * scan's source path and title/root placeholders from one shape.
 */
export type ScanSource = { kind: 'volume'; volume: volumes.Volume } | { kind: 'folder'; path: string };

/**
 * The final path element, tolerant of both `/` and `\` separators -- a
 * volume's mountPath or a chosen folder's path can reach the renderer with
 * either, depending on platform.
 */
export function basename(path: string): string {
  const trimmed = path.replace(/[\\/]+$/, '');
  const idx = Math.max(trimmed.lastIndexOf('/'), trimmed.lastIndexOf('\\'));
  return idx >= 0 ? trimmed.slice(idx + 1) : trimmed;
}

/** The path StartScan should walk for a given source. */
export function sourcePathOf(source: ScanSource): string {
  return source.kind === 'volume' ? source.volume.mountPath : source.path;
}

/** The source's display name -- a volume's own name, or a folder's basename. */
export function sourceDisplayNameOf(source: ScanSource): string {
  return source.kind === 'volume' ? source.volume.name : basename(source.path);
}

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
      // currentPath/log are the scanning body's WALKING line and its
      // capped newest-first log (25-UI-SPEC E5) -- both real, live values
      // from the running walk, present in both sub-states. readErrors is
      // tracked here too (added 25-07) so a source-loss failure that hits
      // during the counting sub-state -- before a total is ever known --
      // still carries a real count into the error state, instead of the
      // reducer silently computing then discarding it every progress tick.
      currentPath: string;
      log: string[];
      readErrors: number;
    }
  | {
      status: 'scanning';
      title: string;
      filesSeen: number;
      bytesSeen: number;
      totalBytes: number;
      readErrors: number;
      startedAt: number;
      currentPath: string;
      log: string[];
    }
  | {
      status: 'error';
      title: string;
      message: string;
      // Everything ErrorBody (plan 25-07) renders about where the scan
      // stopped, snapshotted from the live scan state at the instant
      // SCAN_FAILED fires -- the Go rejection itself carries only a message
      // string (classifyScanFailure's SOURCE_LOSS_MARKER substring check),
      // never structured stop-point data.
      sourcePath: string;
      filesSeen: number;
      // null when the failure happened during the counting sub-state (no
      // total was ever known yet) -- never a fabricated percentage.
      stopPercent: number | null;
      readErrors: number;
    }
  | {
      status: 'done';
      title: string;
      jsonPath: string;
      files: ScanResultFile[];
      fileCount: number;
      totalSize: number;
      durationMs: number;
      partial: boolean;
      // The percentage the scan had reached when it stopped, carried
      // through only for the partial flavour's doneLine ("stopped at
      // {pct}%" swaps in for the meaningless-on-a-partial duration). null
      // when the stop happened before any total was known; unused and left
      // undefined for a complete (non-partial) scan.
      stopPercent?: number | null;
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
 * The two intents a close gesture can carry during the scanning state
 * (25-UI-SPEC.md's Cancellation Contract, CRT-09): Escape / the header ×
 * / a scrim click mean "stop the walk," while "Run in background" means
 * "leave it running." Modeled as an explicit argument on the panel's one
 * close handler rather than inferred from which element was clicked, so a
 * future close trigger can't silently default to the wrong intent.
 */
export type CloseReason = 'cancel-the-scan' | 'leave-it-running';

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
