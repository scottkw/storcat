package catalog

// Options controls optional behavior of CreateCatalogWithContext. The zero
// value is deliberately NOT the CLI's behavior: CreateCatalog (the CLI's
// path, via its thin wrapper) has unconditionally written both JSON and
// HTML since v1 -- so every construction site of Options must set WriteHTML
// explicitly. A bare Options{} silently drops HTML output from a create
// run, with no compile-time signal that anything changed.
type Options struct {
	// WriteHTML controls whether CreateCatalogWithContext writes the .html
	// sibling alongside the .json file. Must be explicitly true wherever
	// v2.3.0 CLI behavior is being preserved.
	WriteHTML bool
	// IncludeHidden controls whether dot-prefixed entries are walked. False
	// (the zero value) preserves today's default: dotfiles are skipped.
	IncludeHidden bool
	// HaltOnSourceLoss controls whether a read failure that also takes the
	// scan root itself unreachable is treated as terminal (stop descending,
	// return a *SourceUnavailableError carrying the partial tree) rather
	// than today's skip-and-continue. The zero value (false) is what
	// preserves v2.3.0's CLI behavior: the CLI wrapper must NEVER set this
	// true. The GUI binding sets it true so a vanished volume reaches the
	// error/partial-catalog UI (CRT-10/CRT-11) instead of silently
	// producing a catalog missing everything after the loss.
	HaltOnSourceLoss bool
	// MarkUnreadableOnSkip controls whether a skip-and-continue single-
	// entry/subtree failure (scan root still reachable) also records the
	// Unreadable/ReadError marker on that node instead of silently
	// dropping it. It does NOT abort the walk -- with this set, a skipped
	// node is marked and the walk continues past it exactly as it does
	// today. The zero value (false) preserves today's behavior for every
	// Create call site (CLI and GUI): neither sets this true. Re-scan's
	// binding (app.go) is the one caller that sets it true, so a diff can
	// distinguish "genuinely removed" from "merely unreadable right now"
	// (28-RESEARCH.md's load-bearing gap: without this, a completed scan
	// can never populate the fourth diff state).
	MarkUnreadableOnSkip bool
}
