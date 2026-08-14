package search

import "storcat-wails/pkg/models"

// SearchIndexedCap is the maximum number of rows SearchIndexed returns to
// the GUI. It bounds the marshaled payload for the ⌘K palette's live
// search -- Phase 23 measured 5.83MB of JSON for a single 42,551-node
// catalog, and an uncapped cross-catalog search response is that multiplied
// by the number of catalogs in the directory.
const SearchIndexedCap = 50

// SearchIndexed is the GUI-only capped sibling of SearchCatalogs, used by
// the ⌘K command palette. It wraps the unmodified SearchCatalogs walk --
// the CLI calls SearchCatalogs directly at cli/search.go:61-62 and is
// untouched by this method's existence -- and caps only the response, not
// the walk cost. Results are truncated to SearchIndexedCap only when the
// true match count exceeds it; ordering and the matcher are inherited
// verbatim from SearchCatalogs, so this method and the CLI can never
// diverge on what counts as a match or in what order matches appear.
func (s *Service) SearchIndexed(searchTerm, catalogDirectory string) (*models.SearchIndexResult, error) {
	all, err := s.SearchCatalogs(searchTerm, catalogDirectory)
	if err != nil {
		return nil, err
	}

	total := len(all)
	results := all
	if total > SearchIndexedCap {
		results = all[:SearchIndexedCap]
	}

	return &models.SearchIndexResult{Results: results, Total: total}, nil
}
