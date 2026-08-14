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
