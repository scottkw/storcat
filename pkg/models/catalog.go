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
