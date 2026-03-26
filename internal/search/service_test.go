package search

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTestCatalog creates a temp directory with a minimal valid JSON catalog file
// and returns the dir path, file path, and file size.
func writeTestCatalog(t *testing.T) (dir string, filePath string, fileSize int64) {
	t.Helper()

	dir = t.TempDir()
	content := []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`)
	filePath = filepath.Join(dir, "test-catalog.json")

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat test catalog: %v", err)
	}
	fileSize = info.Size()

	return dir, filePath, fileSize
}

func TestBrowseCatalogsSize(t *testing.T) {
	s := NewService()
	dir, _, expectedSize := writeTestCatalog(t)

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}

	if results[0].Size != expectedSize {
		t.Errorf("expected Size=%d, got Size=%d", expectedSize, results[0].Size)
	}
}

func TestBrowseCatalogsModified(t *testing.T) {
	s := NewService()
	dir, _, _ := writeTestCatalog(t)

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}

	modified := results[0].Modified
	if _, err := time.Parse(time.RFC3339, modified); err != nil {
		t.Errorf("Modified %q is not RFC3339-parseable: %v", modified, err)
	}
}

func TestLoadCatalog(t *testing.T) {
	s := NewService()
	_, filePath, _ := writeTestCatalog(t)

	item, err := s.LoadCatalog(filePath)
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil CatalogItem, got nil")
	}
	if item.Name != "./" {
		t.Errorf("expected Name='./', got %q", item.Name)
	}
	if item.Type != "directory" {
		t.Errorf("expected Type='directory', got %q", item.Type)
	}
}

func TestLoadCatalogArrayFormat(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	content := []byte(`[{"type":"directory","name":"root","size":100,"contents":[]}]`)
	filePath := filepath.Join(dir, "array-catalog.json")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}

	item, err := s.LoadCatalog(filePath)
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}
	if item == nil {
		t.Fatal("expected non-nil CatalogItem, got nil")
	}
	if item.Name != "root" {
		t.Errorf("expected Name='root', got %q", item.Name)
	}
	if item.Type != "directory" {
		t.Errorf("expected Type='directory', got %q", item.Type)
	}
}

func TestLoadCatalogNotFound(t *testing.T) {
	s := NewService()
	_, err := s.LoadCatalog("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadCatalogInvalidJSON(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(filePath, []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := s.LoadCatalog(filePath)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestBrowseCatalogsCreated(t *testing.T) {
	s := NewService()
	dir, _, _ := writeTestCatalog(t)

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}

	created := results[0].Created

	// Must be RFC3339-parseable
	if _, err := time.Parse(time.RFC3339, created); err != nil {
		t.Errorf("Created %q is not RFC3339-parseable: %v", created, err)
	}

	// Must NOT be in the old "2006-01-02 15:04:05" format (no T separator)
	// RFC3339 strings contain a 'T' between date and time
	found := false
	for _, ch := range created {
		if ch == 'T' {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Created %q does not contain 'T' separator — looks like old non-RFC3339 format", created)
	}
}
