package catalog

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"storcat-wails/pkg/models"
)

// ProgressCallback is called during directory traversal with the current path
type ProgressCallback func(path string)

// Service handles catalog creation and management
type Service struct{}

// NewService creates a new catalog service
func NewService() *Service {
	return &Service{}
}

// CreateCatalog generates JSON and HTML catalogs for the specified directory
func (s *Service) CreateCatalog(title, directoryPath, outputRoot string, copyToDirectory string, onProgress ProgressCallback) (*models.CreateCatalogResult, error) {
	// Traverse directory and build catalog
	catalog, err := s.traverseDirectory(directoryPath, directoryPath, onProgress)
	if err != nil {
		return nil, fmt.Errorf("failed to traverse directory: %w", err)
	}

	// Create JSON file (bare object format)
	jsonPath := filepath.Join(directoryPath, outputRoot+".json")
	if err := s.writeJSONFile(catalog, jsonPath); err != nil {
		return nil, fmt.Errorf("failed to write JSON: %w", err)
	}

	// Create HTML file
	htmlPath := filepath.Join(directoryPath, outputRoot+".html")
	if err := s.writeHTMLFile(catalog, title, htmlPath); err != nil {
		return nil, fmt.Errorf("failed to write HTML: %w", err)
	}

	result := &models.CreateCatalogResult{
		JsonPath:  jsonPath,
		HtmlPath:  htmlPath,
		FileCount: s.countFiles(catalog),
		TotalSize: catalog.Size,
	}

	// Copy to secondary directory if specified
	if copyToDirectory != "" {
		copyJSONPath := filepath.Join(copyToDirectory, outputRoot+".json")
		copyHTMLPath := filepath.Join(copyToDirectory, outputRoot+".html")

		if err := s.copyFile(jsonPath, copyJSONPath); err != nil {
			return nil, fmt.Errorf("failed to copy JSON: %w", err)
		}
		if err := s.copyFile(htmlPath, copyHTMLPath); err != nil {
			return nil, fmt.Errorf("failed to copy HTML: %w", err)
		}

		result.CopyJsonPath = copyJSONPath
		result.CopyHtmlPath = copyHTMLPath
	}

	return result, nil
}

// traverseDirectory recursively builds catalog structure
func (s *Service) traverseDirectory(dirPath, basePath string, onProgress ProgressCallback) (*models.CatalogItem, error) {
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	// Calculate relative path
	relPath, err := filepath.Rel(filepath.Dir(basePath), dirPath)
	if err != nil {
		return nil, err
	}

	// Convert to Unix-style path with ./ prefix
	displayPath := "./" + filepath.ToSlash(relPath)
	if relPath == filepath.Base(basePath) {
		displayPath = "./"
	}

	// Report progress
	if onProgress != nil {
		onProgress(displayPath)
	}

	// Handle files
	if info.Mode().IsRegular() {
		return &models.CatalogItem{
			Type: "file",
			Name: displayPath,
			Size: info.Size(),
		}, nil
	}

	// Handle directories
	if info.IsDir() {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			// Return empty directory if we can't read it
			return &models.CatalogItem{
				Type:     "directory",
				Name:     displayPath,
				Size:     0,
				Contents: []*models.CatalogItem{},
			}, nil
		}

		// Sort entries: directories first, then files, alphabetically
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].IsDir() && !entries[j].IsDir() {
				return true
			}
			if !entries[i].IsDir() && entries[j].IsDir() {
				return false
			}
			return entries[i].Name() < entries[j].Name()
		})

		var contents []*models.CatalogItem
		var totalSize int64

		for _, entry := range entries {
			// Skip hidden files (starting with .)
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			childPath := filepath.Join(dirPath, entry.Name())
			childItem, err := s.traverseDirectory(childPath, basePath, onProgress)
			if err != nil {
				// Skip items we can't access
				continue
			}

			contents = append(contents, childItem)
			totalSize += childItem.Size
		}

		if contents == nil {
			contents = []*models.CatalogItem{}
		}

		return &models.CatalogItem{
			Type:     "directory",
			Name:     displayPath,
			Size:     totalSize,
			Contents: contents,
		}, nil
	}

	return nil, fmt.Errorf("unsupported file type: %s", dirPath)
}

// writeJSONFile writes catalog in bare object format (no indentation)
func (s *Service) writeJSONFile(catalog *models.CatalogItem, path string) error {
	jsonBytes, err := json.Marshal(catalog)
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonBytes, 0644)
}

// writeHTMLFile generates the HTML catalog with exact tree formatting
func (s *Service) writeHTMLFile(catalog *models.CatalogItem, title, path string) error {
	treeStructure := s.generateTreeStructure(catalog, true, "")

	fileCount := s.countFiles(catalog)
	dirCount := s.countDirectories(catalog)
	totalSize := s.formatBytes(catalog.Size)

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
 <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
 <meta name="Author" content="Made by 'tree'">
 <meta name="GENERATOR" content="$Version: $ tree v1.7.0 (c) 1996 - 2014 by Steve Baker, Thomas Moore, Francesc Rocher, Florian Sesser, Kyosuke Tokoro $">
 <title>%s</title>
 <style type="text/css">
  <!--
  BODY { font-family : ariel, monospace, sans-serif; }
  P { font-weight: normal; font-family : ariel, monospace, sans-serif; color: black; background-color: transparent;}
  B { font-weight: normal; color: black; background-color: transparent;}
  A:visited { font-weight : normal; text-decoration : none; background-color : transparent; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:link    { font-weight : normal; text-decoration : none; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:hover   { color : #000000; font-weight : normal; text-decoration : underline; background-color : yellow; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:active  { color : #000000; font-weight: normal; background-color : transparent; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  .VERSION { font-size: small; font-family : arial, sans-serif; }
  .NORM  { color: black;  background-color: transparent;}
  .FIFO  { color: purple; background-color: transparent;}
  .CHAR  { color: yellow; background-color: transparent;}
  .DIR   { color: blue;   background-color: transparent;}
  .BLOCK { color: yellow; background-color: transparent;}
  .LINK  { color: aqua;   background-color: transparent;}
  .SOCK  { color: fuchsia;background-color: transparent;}
  .EXEC  { color: green;  background-color: transparent;}
  -->
 </style>
</head>
<body>
	<h1>%s</h1><p>
	%s
	<br><br>
	</p>
	<p>

 %s used in %d directories, %d files
	<br><br>
	</p>
	<hr>
	<p class="VERSION">
		 tree v1.7.0 © 1996 - 2014 by Steve Baker and Thomas Moore <br>
		 HTML output hacked and copyleft © 1998 by Francesc Rocher <br>
		 JSON output hacked and copyleft © 2014 by Florian Sesser <br>
		 Charsets / OS/2 support © 2001 by Kyosuke Tokoro
	</p>
</body>
</html>`, html.EscapeString(title), html.EscapeString(title), treeStructure, totalSize, dirCount, fileCount)

	return os.WriteFile(path, []byte(htmlContent), 0644)
}

// generateTreeStructure creates the tree visual representation
func (s *Service) generateTreeStructure(item *models.CatalogItem, isLast bool, prefix string) string {
	var result strings.Builder

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	sizeDisplay := s.formatBytesForDisplay(item.Size)
	itemName := filepath.Base(item.Name)

	result.WriteString(fmt.Sprintf("%s%s%s&nbsp;&nbsp;%s<br>\n",
		prefix, connector, sizeDisplay, html.EscapeString(itemName)))

	// Process children for directories
	if item.Type == "directory" && item.Contents != nil && len(item.Contents) > 0 {
		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		for i, child := range item.Contents {
			childIsLast := i == len(item.Contents)-1
			result.WriteString(s.generateTreeStructure(child, childIsLast, newPrefix))
		}
	}

	return result.String()
}

// formatBytesForDisplay formats bytes with bracket and padding (e.g., "[271M]", "[   0]")
func (s *Service) formatBytesForDisplay(bytes int64) string {
	if bytes == 0 {
		return "[   0]"
	}
	formatted := s.formatBytes(bytes)
	return fmt.Sprintf("[%4s]", formatted)
}

// FormatBytes converts bytes to human-readable format (e.g., "271M", "3.4M").
// Exported for use by CLI output formatting.
func FormatBytes(bytes int64) string {
	s := &Service{}
	return s.formatBytes(bytes)
}

// formatBytes converts bytes to human-readable format (e.g., "271M", "3.4M")
func (s *Service) formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	const unit = 1024
	sizes := []string{"B", "K", "M", "G", "T"}

	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit && exp < len(sizes)-1; n /= unit {
		div *= unit
		exp++
	}

	value := float64(bytes) / float64(div)

	// Format to 1 decimal place, but strip .0
	formatted := fmt.Sprintf("%.1f", value)
	if strings.HasSuffix(formatted, ".0") {
		formatted = strings.TrimSuffix(formatted, ".0")
	}

	return formatted + sizes[exp+1]
}

// countFiles counts total files in catalog
func (s *Service) countFiles(catalog *models.CatalogItem) int {
	if catalog.Type == "file" {
		return 1
	}

	count := 0
	if catalog.Type == "directory" && catalog.Contents != nil {
		for _, child := range catalog.Contents {
			count += s.countFiles(child)
		}
	}

	return count
}

// countDirectories counts total directories in catalog
func (s *Service) countDirectories(catalog *models.CatalogItem) int {
	if catalog.Type == "file" {
		return 0
	}

	count := 1 // Count this directory
	if catalog.Type == "directory" && catalog.Contents != nil {
		for _, child := range catalog.Contents {
			count += s.countDirectories(child)
		}
	}

	return count
}

// copyFile copies a file from src to dst
func (s *Service) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
