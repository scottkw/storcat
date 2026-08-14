package models

// CatalogItem represents a file or directory in the catalog
type CatalogItem struct {
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	Size     int64          `json:"size"`
	Contents []*CatalogItem `json:"contents"`
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

// CreateCatalogResult holds the output paths and statistics from a CreateCatalog call
type CreateCatalogResult struct {
	JsonPath     string `json:"jsonPath"`
	HtmlPath     string `json:"htmlPath"`
	FileCount    int    `json:"fileCount"`
	TotalSize    int64  `json:"totalSize"`
	CopyJsonPath string `json:"copyJsonPath,omitempty"`
	CopyHtmlPath string `json:"copyHtmlPath,omitempty"`
}
