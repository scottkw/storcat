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
}
