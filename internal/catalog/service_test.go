package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storcat-wails/pkg/models"
)

// TestWriteJSONFile_BareObject verifies that writeJSONFile produces a bare object
// (starts with '{') not an array-wrapped object (starts with '[').
// Addresses DATA-01.
func TestWriteJSONFile_BareObject(t *testing.T) {
	s := NewService()

	tmpDir, err := os.MkdirTemp("", "storcat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	catalog := &models.CatalogItem{
		Type:     "directory",
		Name:     "./",
		Size:     1024,
		Contents: []*models.CatalogItem{},
	}

	jsonPath := filepath.Join(tmpDir, "test.json")
	if err := s.writeJSONFile(catalog, jsonPath); err != nil {
		t.Fatalf("writeJSONFile failed: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	content := strings.TrimSpace(string(data))
	if !strings.HasPrefix(content, "{") {
		t.Errorf("JSON output should start with '{' (bare object), got: %q", content[:min(len(content), 10)])
	}
	if strings.HasPrefix(content, "[") {
		t.Errorf("JSON output must not start with '[' (array-wrapped), got: %q", content[:min(len(content), 10)])
	}
}

// TestEmptyDirContents verifies that a directory with only hidden files (all skipped)
// serializes with "contents":[] not null or absent.
// Addresses DATA-02.
func TestEmptyDirContents(t *testing.T) {
	s := NewService()

	tmpDir, err := os.MkdirTemp("", "storcat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create only hidden files so all entries are skipped
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644); err != nil {
		t.Fatalf("failed to create hidden file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".DS_Store"), []byte("ds_store"), 0644); err != nil {
		t.Fatalf("failed to create .DS_Store: %v", err)
	}

	item, err := s.traverseDirectory(tmpDir, tmpDir, nil)
	if err != nil {
		t.Fatalf("traverseDirectory failed: %v", err)
	}

	jsonBytes, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, `"contents":[]`) {
		t.Errorf("empty directory should have \"contents\":[], got JSON: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"contents":null`) {
		t.Errorf("empty directory must not have \"contents\":null, got JSON: %s", jsonStr)
	}
}

// TestSymlinkTraversal verifies that symlinks to files are followed and their
// target size is counted, not silently skipped.
// Addresses CATL-03.
func TestSymlinkTraversal(t *testing.T) {
	s := NewService()

	tmpDir, err := os.MkdirTemp("", "storcat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real file
	targetFile := filepath.Join(tmpDir, "target.txt")
	content := []byte("hello symlink world")
	if err := os.WriteFile(targetFile, content, 0644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	// Create a symlink pointing to the real file
	linkFile := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(targetFile, linkFile); err != nil {
		t.Skipf("symlinks not supported on this system: %v", err)
	}

	item, err := s.traverseDirectory(tmpDir, tmpDir, nil)
	if err != nil {
		t.Fatalf("traverseDirectory failed: %v", err)
	}

	// Find the symlink in the results
	var foundSymlink bool
	for _, child := range item.Contents {
		if filepath.Base(child.Name) == "link.txt" {
			foundSymlink = true
			if child.Size != int64(len(content)) {
				t.Errorf("symlink target size should be %d, got %d", len(content), child.Size)
			}
		}
	}

	if !foundSymlink {
		t.Errorf("symlink 'link.txt' not found in catalog contents; contents: %+v", item.Contents)
	}
}

// TestHTMLRootNode verifies that the root item (Name="./") renders with a
// connector and size bracket, not just "./<br>".
// Addresses CATL-04.
func TestHTMLRootNode(t *testing.T) {
	s := NewService()

	root := &models.CatalogItem{
		Type:     "directory",
		Name:     "./",
		Size:     2048,
		Contents: []*models.CatalogItem{},
	}

	output := s.generateTreeStructure(root, true, "")

	if !strings.Contains(output, "└── ") {
		t.Errorf("root node HTML should contain '└── ' connector, got: %q", output)
	}
	if !strings.Contains(output, "[") {
		t.Errorf("root node HTML should contain '[' size bracket, got: %q", output)
	}
	if strings.HasPrefix(output, "./<br>") {
		t.Errorf("root node HTML must not start with './<br>', got: %q", output)
	}
}

// TestCreateCatalogResult verifies that CreateCatalog returns a non-nil
// *CreateCatalogResult with populated JsonPath, HtmlPath, FileCount, TotalSize.
// Addresses CATL-02.
func TestCreateCatalogResult(t *testing.T) {
	s := NewService()

	tmpDir, err := os.MkdirTemp("", "storcat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file so FileCount > 0
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create readme.txt: %v", err)
	}

	result, err := s.CreateCatalog("Test Catalog", tmpDir, "test-output", "", nil)
	if err != nil {
		t.Fatalf("CreateCatalog failed: %v", err)
	}

	if result == nil {
		t.Fatal("CreateCatalog returned nil result")
	}
	if result.JsonPath == "" {
		t.Error("result.JsonPath should not be empty")
	}
	if result.HtmlPath == "" {
		t.Error("result.HtmlPath should not be empty")
	}
	if result.FileCount <= 0 {
		t.Errorf("result.FileCount should be > 0, got %d", result.FileCount)
	}
	if result.TotalSize <= 0 {
		t.Errorf("result.TotalSize should be > 0, got %d", result.TotalSize)
	}
}

// min returns the smaller of two ints (helper for test error messages).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
