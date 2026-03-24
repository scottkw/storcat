package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"storcat-wails/pkg/models"
)

// Service handles catalog searching
type Service struct{}

// NewService creates a new search service
func NewService() *Service {
	return &Service{}
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

		// Try to get title from HTML file first
		htmlPath := strings.TrimSuffix(filePath, ".json") + ".html"
		title := ""

		// Check if HTML file exists and try to extract title
		if htmlData, err := os.ReadFile(htmlPath); err == nil {
			htmlContent := string(htmlData)
			// Extract title from <title> tag
			if startIdx := strings.Index(htmlContent, "<title>"); startIdx != -1 {
				startIdx += 7 // len("<title>")
				if endIdx := strings.Index(htmlContent[startIdx:], "</title>"); endIdx != -1 {
					title = htmlContent[startIdx : startIdx+endIdx]
				}
			}
		}

		// If no title from HTML, fall back to filename without extension
		if title == "" {
			title = strings.TrimSuffix(entry.Name(), ".json")
		}

		// Get file info for modification time
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if HTML file exists (reuse htmlPath from above)
		_, htmlErr := os.Stat(htmlPath)
		hasHtml := htmlErr == nil

		catalogs = append(catalogs, &models.CatalogMetadata{
			Title:    title,
			Name:     entry.Name(),
			Filename: entry.Name(),
			Created:  info.ModTime().Format("2006-01-02 15:04:05"),
			Modified: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			FilePath: filePath,
			HasHtml:  hasHtml,
		})
	}

	return catalogs, nil
}
