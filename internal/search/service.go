package search

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/djherbis/times"
	"storcat-wails/internal/config"
	"storcat-wails/pkg/models"
)

// Service handles catalog searching
type Service struct {
	countsCache *config.CountsCache
}

// NewService creates a new search service
func NewService() *Service {
	return &Service{}
}

// SetCountsCache wires a sidecar counts cache into the service so
// BrowseCatalogs can populate FileCount/TotalBytes on a cache hit. Every
// access is nil-safe: cli/search.go and cli/show.go construct the service
// via NewService() with no cache and must keep compiling and behaving
// identically.
func (s *Service) SetCountsCache(cache *config.CountsCache) {
	s.countsCache = cache
}

// detectParseError returns an empty string when data is valid JSON. json.Valid
// scans without allocating the target structure, so a healthy catalog pays
// only that scan. Only on invalid data does it attempt a real unmarshal --
// mirroring LoadCatalog's own array-then-object attempt order -- to extract
// a *json.SyntaxError's byte offset and reason.
func detectParseError(data []byte) string {
	if json.Valid(data) {
		return ""
	}

	// json.Valid already failed, so any Unmarshal below hits the same
	// syntax error at the same byte offset; the array-then-object order is
	// kept anyway to mirror LoadCatalog's own attempt order.
	var arr []*models.CatalogItem
	err := json.Unmarshal(data, &arr)
	if err == nil {
		var obj models.CatalogItem
		err = json.Unmarshal(data, &obj)
	}
	if err == nil {
		return "" // unreachable: json.Valid already reported this data invalid
	}
	if syn, ok := err.(*json.SyntaxError); ok {
		return fmt.Sprintf("byte %d: %s", syn.Offset, syn.Error())
	}
	return err.Error()
}

// extractJSONTitle probes data for a root "title" field, trying the
// bare-object shape (v2) first and, on failure, the array-wrapped shape
// (v1 bash script/`tree -J`, element 0 is the tree root). Any decode
// failure -- including a shape that matches neither -- returns the empty
// string; Go's decoder discards unmatched keys without allocating for
// them, so this costs roughly one more lexical pass over bytes
// BrowseCatalogs has already read and already scanned once via
// detectParseError's json.Valid check.
func extractJSONTitle(data []byte) string {
	type titleProbe struct {
		Title string `json:"title"`
	}

	var obj titleProbe
	if err := json.Unmarshal(data, &obj); err == nil {
		return obj.Title
	}

	var arr []titleProbe
	if err := json.Unmarshal(data, &arr); err == nil && len(arr) > 0 {
		return arr[0].Title
	}

	return ""
}

// SearchCatalogs searches all JSON catalogs in the specified directory for the search term
func (s *Service) SearchCatalogs(searchTerm, catalogDirectory string) ([]*models.SearchResult, error) {
	var results []*models.SearchResult

	entries, err := os.ReadDir(catalogDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog directory: %w", err)
	}

	searchTermLower := strings.ToLower(searchTerm)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(catalogDirectory, entry.Name())
		catalogName := strings.TrimSuffix(entry.Name(), ".json")

		matches, err := s.searchInCatalogFile(filePath, catalogName, searchTermLower)
		if err != nil {
			// Skip files that can't be read or parsed
			continue
		}

		results = append(results, matches...)
	}

	return results, nil
}

// searchInCatalogFile searches a single catalog JSON file
func (s *Service) searchInCatalogFile(filePath, catalogName, searchTermLower string) ([]*models.SearchResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Try parsing as array first (bash script format)
	var catalogArray []*models.CatalogItem
	if err := json.Unmarshal(data, &catalogArray); err == nil && len(catalogArray) > 0 {
		return s.searchInCatalog(catalogArray[0], catalogName, filePath, searchTermLower), nil
	}

	// Try parsing as single object (our format)
	var catalogObj models.CatalogItem
	if err := json.Unmarshal(data, &catalogObj); err != nil {
		return nil, err
	}

	return s.searchInCatalog(&catalogObj, catalogName, filePath, searchTermLower), nil
}

// searchInCatalog recursively searches a catalog structure
func (s *Service) searchInCatalog(item *models.CatalogItem, catalogName, catalogFilePath, searchTermLower string) []*models.SearchResult {
	var results []*models.SearchResult

	// Check if this item matches
	if item.Name != "" && strings.Contains(strings.ToLower(item.Name), searchTermLower) {
		basename := filepath.Base(item.Name)
		fullPath := filepath.Dir(item.Name)
		if fullPath == "." {
			fullPath = ""
		}

		results = append(results, &models.SearchResult{
			Catalog:         catalogName,
			CatalogFilePath: catalogFilePath,
			Basename:        basename,
			FullPath:        fullPath,
			FullName:        item.Name,
			Type:            item.Type,
			Size:            item.Size,
		})
	}

	// Search in children
	if item.Contents != nil {
		for _, child := range item.Contents {
			results = append(results, s.searchInCatalog(child, catalogName, catalogFilePath, searchTermLower)...)
		}
	}

	return results
}

// LoadCatalog reads and parses a catalog JSON file, returning the root CatalogItem.
// Supports both bare-object format (v2.0.0) and array-wrapped format (v1.0 bash script).
func (s *Service) LoadCatalog(filePath string) (*models.CatalogItem, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog file: %w", err)
	}

	// Try array format first (v1.0 bash script compatibility)
	var catalogArray []*models.CatalogItem
	if err := json.Unmarshal(data, &catalogArray); err == nil && len(catalogArray) > 0 {
		return catalogArray[0], nil
	}

	// Try bare object format (v2.0.0)
	var catalogObj models.CatalogItem
	if err := json.Unmarshal(data, &catalogObj); err != nil {
		return nil, fmt.Errorf("failed to parse catalog JSON: %w", err)
	}

	return &catalogObj, nil
}

// BrowseCatalogs loads metadata for all catalogs in a directory
func (s *Service) BrowseCatalogs(catalogDirectory string) ([]*models.CatalogMetadata, error) {
	var catalogs []*models.CatalogMetadata

	entries, err := os.ReadDir(catalogDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(catalogDirectory, entry.Name())
		htmlPath := strings.TrimSuffix(filePath, ".json") + ".html"

		// Get file info for modification time
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Get creation time (birth time on macOS/Windows, fallback to mtime)
		var createdTime time.Time
		t, err := times.Stat(filePath)
		if err == nil && t.HasBirthTime() {
			createdTime = t.BirthTime()
		} else {
			createdTime = info.ModTime()
		}

		// Check if HTML file exists
		_, htmlErr := os.Stat(htmlPath)
		hasHtml := htmlErr == nil

		// Parse status plus the title probe below share this one read of
		// filePath. json.Valid's fast path (inside detectParseError) keeps
		// the common (valid) case to one read plus one linear scan.
		data, readErr := os.ReadFile(filePath)
		var parseErr string
		if readErr != nil {
			parseErr = readErr.Error()
		} else {
			parseErr = detectParseError(data)
		}

		// Title resolution, in order: (1) the JSON root's own "title" field
		// -- checked only when the document parsed cleanly, and reused from
		// the bytes already read above, no second read of filePath; (2) the
		// sibling .html's <title>, unescaped -- the write side already
		// escapes via html.EscapeString, so a title containing "&" must be
		// unescaped back to its literal form here, not left as "&amp;";
		// (3) the filename minus ".json".
		title := ""
		if parseErr == "" {
			title = extractJSONTitle(data)
		}
		if title == "" {
			if htmlData, err := os.ReadFile(htmlPath); err == nil {
				htmlContent := string(htmlData)
				if startIdx := strings.Index(htmlContent, "<title>"); startIdx != -1 {
					startIdx += 7 // len("<title>")
					if endIdx := strings.Index(htmlContent[startIdx:], "</title>"); endIdx != -1 {
						title = html.UnescapeString(htmlContent[startIdx : startIdx+endIdx])
					}
				}
			}
		}
		if title == "" {
			title = strings.TrimSuffix(entry.Name(), ".json")
		}

		// Counts: cache-backed, never computed inline here. A miss leaves
		// both pointers nil rather than blocking the rail or fabricating a
		// zero.
		var fileCount *int
		var totalBytes *int64
		if s.countsCache != nil {
			key := config.CountsKey(filePath, info.ModTime(), info.Size())
			if entry, ok := s.countsCache.Get(key); ok {
				fc := entry.FileCount
				tb := entry.TotalBytes
				fileCount = &fc
				totalBytes = &tb
			}
		}

		catalogs = append(catalogs, &models.CatalogMetadata{
			Title:      title,
			Name:       entry.Name(),
			Filename:   entry.Name(),
			Size:       info.Size(),
			Created:    createdTime.Format(time.RFC3339),
			Modified:   info.ModTime().Format(time.RFC3339),
			FilePath:   filePath,
			HasHtml:    hasHtml,
			FileCount:  fileCount,
			TotalBytes: totalBytes,
			ParseError: parseErr,
		})
	}

	return catalogs, nil
}
