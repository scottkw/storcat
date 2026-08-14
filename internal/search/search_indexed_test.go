package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"storcat-wails/pkg/models"
)

// writeSearchIndexedFixture writes a single-catalog JSON file whose root
// directory contains n files, each named so it contains needle as a
// substring -- the boundary counts (50, 51) are derived from n, never
// hardcoded from a guess.
func writeSearchIndexedFixture(t *testing.T, dir, filename, needle string, n int) string {
	t.Helper()

	contents := make([]*models.CatalogItem, n)
	for i := 0; i < n; i++ {
		contents[i] = &models.CatalogItem{
			Type: "file",
			Name: fmt.Sprintf("./%s-%03d.txt", needle, i),
			Size: int64(i),
		}
	}
	root := &models.CatalogItem{
		Type:     "directory",
		Name:     "./",
		Contents: contents,
	}

	data, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return filePath
}

func TestSearchIndexed_ZeroMatches(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	writeSearchIndexedFixture(t, dir, "cat.json", "needle", 5)

	result, err := s.SearchIndexed("zzznomatchzzz", dir)
	if err != nil {
		t.Fatalf("SearchIndexed failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for zero matches")
	}
	if result.Total != 0 {
		t.Errorf("expected Total 0, got %d", result.Total)
	}
	if len(result.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(result.Results))
	}
}

func TestSearchIndexed_ExactlyCapMatches(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	writeSearchIndexedFixture(t, dir, "cat.json", "needle", SearchIndexedCap)

	result, err := s.SearchIndexed("needle", dir)
	if err != nil {
		t.Fatalf("SearchIndexed failed: %v", err)
	}
	if result.Total != SearchIndexedCap {
		t.Errorf("expected Total %d, got %d", SearchIndexedCap, result.Total)
	}
	if len(result.Results) != SearchIndexedCap {
		t.Errorf("expected %d results (no truncation at the boundary), got %d", SearchIndexedCap, len(result.Results))
	}
}

func TestSearchIndexed_OverCapMatches(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	writeSearchIndexedFixture(t, dir, "cat.json", "needle", SearchIndexedCap+1)

	result, err := s.SearchIndexed("needle", dir)
	if err != nil {
		t.Fatalf("SearchIndexed failed: %v", err)
	}
	if result.Total != SearchIndexedCap+1 {
		t.Errorf("expected Total %d, got %d", SearchIndexedCap+1, result.Total)
	}
	if len(result.Results) != SearchIndexedCap {
		t.Errorf("expected %d results (truncated), got %d", SearchIndexedCap, len(result.Results))
	}
}

// TestSearchIndexed_ParityWithSearchCatalogs is the load-bearing assertion:
// SearchIndexed must never re-sort, re-filter, or re-match -- its Results
// must be element-for-element equal to the first min(cap, len) elements of
// SearchCatalogs' own output, in the same order, and Total must equal
// SearchCatalogs' full length.
func TestSearchIndexed_ParityWithSearchCatalogs(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	writeSearchIndexedFixture(t, dir, "cat.json", "needle", SearchIndexedCap+1)

	all, err := s.SearchCatalogs("needle", dir)
	if err != nil {
		t.Fatalf("SearchCatalogs failed: %v", err)
	}
	indexed, err := s.SearchIndexed("needle", dir)
	if err != nil {
		t.Fatalf("SearchIndexed failed: %v", err)
	}

	if indexed.Total != len(all) {
		t.Fatalf("Total = %d, want len(SearchCatalogs(...)) = %d", indexed.Total, len(all))
	}

	wantCap := len(all)
	if wantCap > SearchIndexedCap {
		wantCap = SearchIndexedCap
	}
	if !reflect.DeepEqual(indexed.Results, all[:wantCap]) {
		t.Errorf("Results is not element-for-element equal to SearchCatalogs(...)[:%d]", wantCap)
	}
}

// TestSearchIndexed_CrossCatalogDuplicatePath proves two catalogs sharing an
// identical node path never collapse into one result -- each carries its
// own Catalog/CatalogFilePath.
func TestSearchIndexed_CrossCatalogDuplicatePath(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	writeSearchIndexedFixture(t, dir, "alpha.json", "needle", 1)
	writeSearchIndexedFixture(t, dir, "beta.json", "needle", 1)

	result, err := s.SearchIndexed("needle", dir)
	if err != nil {
		t.Fatalf("SearchIndexed failed: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2 results across two catalogs, got %d", result.Total)
	}
	if result.Results[0].Catalog == result.Results[1].Catalog {
		t.Errorf("expected distinct Catalog values, got %q and %q", result.Results[0].Catalog, result.Results[1].Catalog)
	}
}

func TestSearchIndexed_UnreadableDirectory(t *testing.T) {
	s := NewService()

	_, err := s.SearchIndexed("needle", filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected an error for an unreadable directory, got nil")
	}
}
