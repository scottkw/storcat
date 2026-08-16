import type { ScanState } from '../../../types/scan';

type ErrorScanState = Extract<ScanState, { status: 'error' }>;

// Create's own original explanation, now the default for the `explanation`
// prop below -- keeping this as the default (rather than requiring every
// call site to repeat it) is what lets CreateSlideOver's call site stay
// unedited (28-03).
const DEFAULT_EXPLANATION =
  "Nothing was written yet. A partial catalog is still useful for a failing card — it records what could be read, and marks the gap.";

export interface ErrorBodyProps {
  scan: ErrorScanState;
  // Optional, TOGETHER (28-03) -- when both are omitted, the partial-write
  // button doesn't render and "Retry scan" is promoted into the
  // primary-styled slot in its place, so exactly one strong call to action
  // exists either way. Re-scan's error step has nothing well-defined to
  // write a partial version of (a partial diff against a half-walked new
  // tree has no resolution -- 28-UI-SPEC.md), so its call site omits both.
  // CreateSlideOver's own call site is unchanged -- it still always passes
  // both.
  writingPartial?: boolean;
  onWritePartial?: () => void;
  onRetry: () => void;
  onCloseWithoutWriting: () => void;
  // Optional (28-03), defaulting to Create's own literal above so
  // CreateSlideOver's call site needs no edit. Re-scan's call site overrides
  // this with copy about its own EXISTING catalog, not a not-yet-written one.
  explanation?: string;
}

/**
 * CRT-10/CRT-11's error state (25-UI-SPEC.md's Error State Contract, E6).
 * E6 is the one element the design contract's own consideration probe
 * returned `unclassified` for -- its three rows were authored directly from
 * CRT-10/CRT-11 rather than proposed by the engine, so this component's
 * state coverage is an explicit, re-derived assumption, not a
 * machine-checked one (25-07-PLAN.md's Flagged Assumptions).
 *
 * The read-error log renders one aggregate count line, not one line per
 * failure. This plan is frontend-only (no new Go binding): the wire payload
 * (ScanProgress, and the StartScan rejection's message) only ever carries a
 * read-error *count*, never individual paths/reasons -- internal/catalog's
 * per-entry ReadErrorEntry list (path + reason) exists only on the Go side
 * and never crosses the bridge. Inventing per-path text to match the
 * design mockup's illustrative "read error: {path} -- input/output error"
 * line would be fabrication (CLAUDE.md's "Silent Fallbacks" rule extends to
 * inventing detail, not just suppressing errors); an honest count is what
 * the data actually supports. When the count is zero (a total, instant
 * disconnect with no prior read errors) this line is omitted entirely,
 * satisfying E6's zero-one-many resolution with the same template.
 */
function ErrorBody({
  scan,
  writingPartial = false,
  onWritePartial,
  onRetry,
  onCloseWithoutWriting,
  explanation = DEFAULT_EXPLANATION,
}: ErrorBodyProps) {
  // The locked headline template ("Stopped at {pct}% -- the volume went
  // away") assumes a percentage is always known. A source loss can also
  // happen during the counting sub-state, before any total -- and therefore
  // any percentage -- has ever been resolved. E6 is flagged unresolved for
  // exactly this kind of gap; this is a reasoned, honest extension (never a
  // fabricated 0%), the same discretion 25-UI-SPEC.md itself already
  // exercised for the counting sub-state's status-bar copy.
  const headline =
    scan.stopPercent !== null
      ? `Stopped at ${scan.stopPercent}% — the volume went away`
      : 'Stopped — the volume went away';

  return (
    <div className="ws-create-state-body">
      <div className="ws-create-error-row">
        <span className="ws-create-badge ws-create-badge-error" aria-hidden="true">
          !
        </span>
        <div className="ws-create-error-col">
          <div className="ws-create-error-headline">{headline}</div>
          <div className="ws-create-error-subline mono">
            {scan.sourcePath} · {scan.filesSeen} files walked
          </div>
        </div>
      </div>

      <div className="ws-create-errlog">
        {scan.readErrors > 0 && (
          <div className="ws-create-errlog-line mono">
            {scan.readErrors} read {scan.readErrors === 1 ? 'error' : 'errors'} recorded before the stop.
          </div>
        )}
        <div className="ws-create-errlog-summary mono">
          volume {scan.sourcePath} disappeared — card removed or failing
        </div>
      </div>

      <p className="ws-create-explain">{explanation}</p>

      <div className="ws-create-actions">
        {/* Primary: calls wailsAPI.writePartialCatalog (via onWritePartial,
            owned by CreateSlideOver alongside its own submitting-ref guard).
            Disabling from the first click onward -- not just the parent's
            synchronous ref guard -- is what stops a second call from ever
            being issued; the binding itself also caches its first result,
            so two clicks land on one catalog even if this disable somehow
            raced (T-25-13, defense in depth). Omitted entirely when the
            caller passes no onWritePartial (28-03) -- "Retry scan" below
            takes this slot's primary styling instead, so exactly one strong
            CTA exists either way. */}
        {onWritePartial && (
          <button
            type="button"
            className="ws-create-btn ws-create-btn-primary"
            disabled={writingPartial}
            onClick={onWritePartial}
          >
            Write partial catalog
          </button>
        )}
        {/* Restarts the scan on the same already-selected source. StartScan
            runs synchronously on the Go side and its deferred cleanup clears
            the active-scan handle before the rejected promise ever reaches
            this component, so by the time this button is clickable the prior
            context has already released -- no separate wait is needed beyond
            the submitting-ref guard CreateSlideOver's handleCreate already
            applies to every start (T-25-24). Disabled while writingPartial is
            true (CR-02/WR-02): retrying while a partial write is still in
            flight would clear the retained tree that write is reading from,
            racing the write for the state it records on completion.
            Outlined secondary when a partial-write action exists alongside
            it; promoted to the primary-styled slot (28-03) when it doesn't,
            per the doc comment on the button above. */}
        <button
          type="button"
          className={
            onWritePartial ? 'ws-create-btn-outline ws-create-btn-outline-34' : 'ws-create-btn ws-create-btn-primary'
          }
          disabled={writingPartial}
          onClick={onRetry}
        >
          Retry scan
        </button>
        {/* Text tertiary: names the consequence, not the gesture -- a user
            here is never left guessing whether closing quietly kept a
            partial result. Runs the same close handler every other close
            path uses; writes nothing on its own. Disabled while
            writingPartial is true (CR-02/WR-02) so an in-flight write always
            finishes and updates state before the user can navigate away. */}
        <button
          type="button"
          className="ws-create-btn ws-create-btn-text"
          disabled={writingPartial}
          onClick={onCloseWithoutWriting}
        >
          Close without writing
        </button>
      </div>
    </div>
  );
}

export default ErrorBody;
