package models

// CatalogItem represents a file or directory in the catalog.
//
// Title is set ONLY on a catalog's root node, and only once a user has
// renamed that catalog (ACT-02, Phase 27) -- it is absent (omitempty) for
// every catalog nobody has ever renamed, so COMPAT-02's byte-for-byte
// v2.3.0 shape is preserved unchanged for the overwhelming majority of
// catalogs. Before this field existed a catalog's title lived only in its
// generated .html <title>; Title is now the authoritative source and the
// HTML title is a fallback for catalogs written before this field existed
// (internal/search/service.go's BrowseCatalogs). Its struct position
// (immediately after Name) is unobservable while it is absent -- Go's
// encoding/json marshals struct fields in declaration order, and an absent
// omitempty field contributes no bytes either way.
//
// Unreadable and ReadError appear ONLY on a partial catalog written after a
// scan-root loss (CRT-10/CRT-11) -- they mark the single directory node
// where the loss was first detected. Both are absent-when-zero, so a
// complete scan (the overwhelming majority of catalogs) omits them entirely
// and its JSON is byte-for-byte the v2.3.0 shape (COMPAT-02). This is the
// one scoped, deliberate divergence from that byte-identical guarantee, and
// it exists only in the partial-catalog path. No reader in this repository
// rejects unrecognized JSON keys, so these two keys are silently ignored by
// every catalog reader that doesn't yet know about them.
type CatalogItem struct {
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	Title    string         `json:"title,omitempty"`
	Size     int64          `json:"size"`
	Contents []*CatalogItem `json:"contents"`
	// Unreadable is true only on the directory node where a scan-root loss
	// was first detected during the walk.
	Unreadable bool `json:"unreadable,omitempty"`
	// ReadError carries the reason the walk stopped at this node. Non-empty
	// exactly when Unreadable is true.
	ReadError string `json:"readError,omitempty"`
	// ModTime is the entry's modification time, captured at scan time as
	// Unix seconds -- not RFC3339 (25-RESEARCH.md A1: smaller on the wire,
	// and FAT32's own 2-second mtime granularity makes sub-second precision
	// meaningless here). omitempty so a catalog written before this field
	// existed stays byte-for-byte its original v2.3.0 shape (COMPAT-02):
	// absent, never zero-as-the-epoch. internal/catalog/diff.go treats
	// ModTime == 0 on the OLD side as "unknown, predates this field" and
	// falls back to a size-only comparison -- it must never be read as the
	// Unix epoch by any comparison.
	ModTime int64 `json:"modTime,omitempty"`
}

// DiffState is one of the five categories a re-scan's diff assigns to every
// path encountered across the old and new trees combined (28-CONTEXT.md).
// Directories are only ever added/removed/unchanged -- there is
// deliberately no directory-changed state (see internal/catalog/diff.go).
type DiffState string

const (
	DiffAdded      DiffState = "added"
	DiffRemoved    DiffState = "removed"
	DiffChanged    DiffState = "changed"
	DiffUnreadable DiffState = "unreadable"
	DiffUnchanged  DiffState = "unchanged"
)

// DiffEntry is one row of a re-scan's diff -- present for every added,
// removed, changed or unreadable path (never for unchanged, which is
// count-only in DiffResult). OldSize/NewSize are omitempty and zero when
// not applicable to State (an added entry has no OldSize, a removed entry
// has no NewSize); ReadError is set only when State is DiffUnreadable.
type DiffEntry struct {
	Path      string    `json:"path"`
	State     DiffState `json:"state"`
	Type      string    `json:"type"`
	OldSize   int64     `json:"oldSize,omitempty"`
	NewSize   int64     `json:"newSize,omitempty"`
	ReadError string    `json:"readError,omitempty"`
}

// DiffResult is a re-scan's full comparison output. The five counts must
// sum to the total number of distinct paths encountered across the old
// tree union the new tree (28-UI-SPEC.md's stated invariant); Entries holds
// one row per added/removed/changed/unreadable path, in no particular
// order -- ordering for display is the frontend's concern.
type DiffResult struct {
	Added      int         `json:"added"`
	Removed    int         `json:"removed"`
	Changed    int         `json:"changed"`
	Unreadable int         `json:"unreadable"`
	Unchanged  int         `json:"unchanged"`
	Entries    []DiffEntry `json:"entries"`
}

// SearchResult represents a search result from catalog files
type SearchResult struct {
	Catalog         string `json:"catalog"`
	CatalogFilePath string `json:"catalogFilePath"`
	Basename        string `json:"basename"`
	FullPath        string `json:"fullPath"`
	FullName        string `json:"fullName"`
	Type            string `json:"type"`
	Size            int64  `json:"size"`
}

// SearchIndexResult is the GUI-only capped transport for the ⌘K palette's
// live search. Total is the true match count across every catalog in the
// directory; Results is the first SearchIndexedCap of them, in the same
// order SearchCatalogs produced -- never re-sorted, never re-matched. The
// CLI's `storcat search` keeps using SearchCatalogs directly and is
// unaffected by this struct's existence.
type SearchIndexResult struct {
	Results []*SearchResult `json:"results"`
	Total   int             `json:"total"`
}

// CatalogMetadata represents metadata about a catalog file
type CatalogMetadata struct {
	Title    string `json:"title"`
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
	FilePath string `json:"path"`
	HasHtml  bool   `json:"hasHtml"`
	// FileCount and TotalBytes are pointers so the frontend can tell "not
	// computed yet" (a sidecar-cache miss) from "genuinely zero" -- nil
	// rather than a confident 0 on every rail row until the cache warms.
	FileCount  *int   `json:"fileCount"`
	TotalBytes *int64 `json:"totalBytes"`
	// ParseError is empty for a catalog that reads and parses cleanly, and
	// otherwise carries a byte offset plus the parser's own reason (or the
	// raw read error, for a file that can't be opened at all).
	ParseError string `json:"parseError"`
}

// FlatNode is one node of a flattened catalog tree, render-ready for the
// virtualized tree pane. Name is the basename (filepath.Base of the source
// CatalogItem.Name); Path is that same source field verbatim -- the two are
// kept separate because CatalogItem.Name holds a full relative display path,
// not a basename.
type FlatNode struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
	Depth       int    `json:"depth"`
	ParentIdx   int    `json:"parentIdx"`
	HasChildren bool   `json:"hasChildren"`
}

// FlatCatalog is the full flattened node array for one catalog, returned in
// a single call so the frontend never round-trips to Go per expand/collapse.
type FlatCatalog struct {
	Nodes      []FlatNode `json:"nodes"`
	FileCount  int        `json:"fileCount"`
	TotalBytes int64      `json:"totalBytes"`
}

// CreateCatalogResult holds the output paths and statistics from a CreateCatalog call.
//
// The *Size fields are each output file's own real on-disk byte count (the
// exact length of the bytes written, or copied, to that path) -- distinct
// from TotalSize, which is the scanned tree's total content size, not any
// one output file's size. Added for the done state's written-files list
// (25-07, CRT-12: "every file actually written, with its size"), which
// would otherwise have no honest per-row size to render. JsonSize has no
// omitempty, matching JsonPath's own always-present convention; the other
// three are omitempty, matching their path counterparts, since they are
// only set when that output was actually produced.
type CreateCatalogResult struct {
	JsonPath     string `json:"jsonPath"`
	JsonSize     int64  `json:"jsonSize"`
	HtmlPath     string `json:"htmlPath"`
	HtmlSize     int64  `json:"htmlSize,omitempty"`
	FileCount    int    `json:"fileCount"`
	TotalSize    int64  `json:"totalSize"`
	CopyJsonPath string `json:"copyJsonPath,omitempty"`
	CopyJsonSize int64  `json:"copyJsonSize,omitempty"`
	CopyHtmlPath string `json:"copyHtmlPath,omitempty"`
	CopyHtmlSize int64  `json:"copyHtmlSize,omitempty"`
}
