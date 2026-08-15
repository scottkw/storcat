package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"storcat-wails/internal/config"
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

func TestBrowseCatalogs_ParseError_WellFormedV2(t *testing.T) {
	s := NewService()
	dir, _, _ := writeTestCatalog(t)

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	if results[0].ParseError != "" {
		t.Errorf("expected empty ParseError for a well-formed v2 catalog, got %q", results[0].ParseError)
	}
}

func TestBrowseCatalogs_ParseError_WellFormedV1Array(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	content := []byte(`[{"type":"directory","name":"root","size":100,"contents":[]}]`)
	if err := os.WriteFile(filepath.Join(dir, "array.json"), content, 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	if results[0].ParseError != "" {
		t.Errorf("expected empty ParseError for a well-formed v1 array-wrapped catalog, got %q", results[0].ParseError)
	}
}

func TestBrowseCatalogs_ParseError_Truncated(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	content := []byte(`{"type":"directory","name":"./","size":0,"contents":[{"type":"file","name":"a"`)
	if err := os.WriteFile(filepath.Join(dir, "truncated.json"), content, 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	if !strings.Contains(results[0].ParseError, "byte ") {
		t.Errorf("expected ParseError to contain a byte offset for a truncated document, got %q", results[0].ParseError)
	}
}

func TestBrowseCatalogs_ParseError_Malformed(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	content := []byte(`{not valid json at all`)
	if err := os.WriteFile(filepath.Join(dir, "malformed.json"), content, 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	if !strings.Contains(results[0].ParseError, "byte ") {
		t.Errorf("expected ParseError to contain a byte offset for malformed JSON, got %q", results[0].ParseError)
	}
}

func TestBrowseCatalogs_ParseError_UnreadableFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("skipping permission test: running as root ignores mode bits")
	}

	s := NewService()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "noperm.json")
	if err := os.WriteFile(filePath, []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}
	if err := os.Chmod(filePath, 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(filePath, 0644) })

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	if results[0].ParseError == "" {
		t.Fatal("expected a non-empty ParseError for an unreadable file")
	}
	if strings.Contains(results[0].ParseError, "byte ") {
		t.Errorf("expected a read error (not a byte-offset syntax error) for an unreadable file, got %q", results[0].ParseError)
	}
}

func TestBrowseCatalogs_ReturnsFilenameOrder(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	// Written out of alphabetical order so a passing test proves
	// BrowseCatalogs sorts by filename rather than relying on incidental
	// filesystem/creation order.
	names := []string{"zebra.json", "apple.json", "mango.json"}
	for _, n := range names {
		content := []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`)
		if err := os.WriteFile(filepath.Join(dir, n), content, 0644); err != nil {
			t.Fatalf("failed to write %s: %v", n, err)
		}
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 catalog results, got %d", len(results))
	}

	want := []string{"apple.json", "mango.json", "zebra.json"}
	for i, w := range want {
		if results[i].Name != w {
			t.Errorf("result[%d].Name = %q, want %q (filename order)", i, results[i].Name, w)
		}
	}
}

func TestBrowseCatalogs_CountsCache_HitAndMiss(t *testing.T) {
	s := NewService()
	dir, filePath, _ := writeTestCatalog(t)

	// No cache wired at all: fields must be nil, never a fabricated zero.
	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs (no cache) failed: %v", err)
	}
	if results[0].FileCount != nil || results[0].TotalBytes != nil {
		t.Fatalf("expected nil FileCount/TotalBytes with no cache wired, got %v/%v", results[0].FileCount, results[0].TotalBytes)
	}

	cache, err := config.NewCountsCacheAt(filepath.Join(t.TempDir(), "counts-cache.json"))
	if err != nil {
		t.Fatalf("NewCountsCacheAt: %v", err)
	}
	s.SetCountsCache(cache)

	// Cache wired but cold: still a miss for this exact key.
	results, err = s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs (cold cache) failed: %v", err)
	}
	if results[0].FileCount != nil || results[0].TotalBytes != nil {
		t.Fatalf("expected nil FileCount/TotalBytes on cache miss, got %v/%v", results[0].FileCount, results[0].TotalBytes)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	key := config.CountsKey(filePath, info.ModTime(), info.Size())
	if err := cache.Put(key, config.CountEntry{FileCount: 7, TotalBytes: 4096}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	results, err = s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs (warm cache) failed: %v", err)
	}
	if results[0].FileCount == nil || *results[0].FileCount != 7 {
		t.Errorf("expected FileCount=7 on cache hit, got %v", results[0].FileCount)
	}
	if results[0].TotalBytes == nil || *results[0].TotalBytes != 4096 {
		t.Errorf("expected TotalBytes=4096 on cache hit, got %v", results[0].TotalBytes)
	}
}

func TestDetectParseError_AllocsValidLessThanInvalid(t *testing.T) {
	valid := []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`)
	invalid := []byte(`{not valid json at all`)

	validAllocs := testing.AllocsPerRun(100, func() {
		_ = detectParseError(valid)
	})
	invalidAllocs := testing.AllocsPerRun(100, func() {
		_ = detectParseError(invalid)
	})

	if !(validAllocs < invalidAllocs) {
		t.Errorf("expected fewer allocations for a valid document (%v) than an invalid one (%v)", validAllocs, invalidAllocs)
	}
}

// TestLoadCatalog_ToleratesMarkerFields verifies that a catalog JSON
// carrying the phase 25-02 partial-catalog marker keys (unreadable,
// readError) parses without error and produces the same node count as the
// same tree without them -- no reader in this repo rejects unrecognized
// JSON keys.
func TestLoadCatalog_ToleratesMarkerFields(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	content := []byte(`{"type":"directory","name":"./","size":5,"contents":[` +
		`{"type":"file","name":"./a.txt","size":5,"contents":null},` +
		`{"type":"directory","name":"./gone","size":0,"contents":[],"unreadable":true,"readError":"stat ./gone: no such file or directory"}` +
		`]}`)
	filePath := filepath.Join(dir, "partial.json")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	item, err := s.LoadCatalog(filePath)
	if err != nil {
		t.Fatalf("LoadCatalog failed to tolerate marker fields: %v", err)
	}
	if len(item.Contents) != 2 {
		t.Fatalf("expected 2 direct children, got %d", len(item.Contents))
	}
}

// TestBrowseCatalogs_TitlePrecedence pins the three-tier title chain: JSON
// root "title" wins outright over a sibling .html's <title>, which in turn
// wins over the filename-derived fallback used when neither is present.
func TestBrowseCatalogs_TitlePrecedence(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	// (a) JSON title present AND a sibling .html with a different title:
	// the JSON title wins.
	if err := os.WriteFile(filepath.Join(dir, "a.json"),
		[]byte(`{"type":"directory","name":"./","title":"JSON Wins","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("write a.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.html"),
		[]byte(`<html><head><title>HTML Loses</title></head></html>`), 0644); err != nil {
		t.Fatalf("write a.html: %v", err)
	}

	// (b) No JSON title, sibling .html present: the HTML title wins.
	if err := os.WriteFile(filepath.Join(dir, "b.json"),
		[]byte(`{"type":"directory","name":"./","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("write b.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.html"),
		[]byte(`<html><head><title>HTML Wins</title></head></html>`), 0644); err != nil {
		t.Fatalf("write b.html: %v", err)
	}

	// (c) No JSON title, no .html: the filename minus .json wins.
	if err := os.WriteFile(filepath.Join(dir, "c.json"),
		[]byte(`{"type":"directory","name":"./","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("write c.json: %v", err)
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 catalog results, got %d", len(results))
	}

	byName := map[string]string{}
	for _, r := range results {
		byName[r.Name] = r.Title
	}

	if got := byName["a.json"]; got != "JSON Wins" {
		t.Errorf("a.json title = %q, want %q", got, "JSON Wins")
	}
	if got := byName["b.json"]; got != "HTML Wins" {
		t.Errorf("b.json title = %q, want %q", got, "HTML Wins")
	}
	if got := byName["c.json"]; got != "c" {
		t.Errorf("c.json title = %q, want %q", got, "c")
	}
}

// TestBrowseCatalogs_UnescapesHTMLTitle proves the read-side escaping bug
// is fixed: an .html <title> containing HTML entities must round-trip back
// to their literal characters, not display the escaped form.
func TestBrowseCatalogs_UnescapesHTMLTitle(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "photos.json"),
		[]byte(`{"type":"directory","name":"./","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "photos.html"),
		[]byte(`<html><head><title>Tom &amp; Jerry &lt;2024&gt;</title></head></html>`), 0644); err != nil {
		t.Fatalf("write html: %v", err)
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	want := "Tom & Jerry <2024>"
	if results[0].Title != want {
		t.Errorf("Title = %q, want %q", results[0].Title, want)
	}
}

// TestBrowseCatalogs_JSONTitleIsNotUnescaped proves a JSON-sourced title is
// returned verbatim -- the JSON field holds the user's literal string, so
// unescaping it would corrupt a title that genuinely contains that text.
func TestBrowseCatalogs_JSONTitleIsNotUnescaped(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "photos.json"),
		[]byte(`{"type":"directory","name":"./","title":"A &amp; B","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	want := "A &amp; B"
	if results[0].Title != want {
		t.Errorf("Title = %q, want %q (verbatim, not unescaped)", results[0].Title, want)
	}
}

// TestBrowseCatalogs_ArrayWrappedTitle proves a v1 array-wrapped catalog
// whose element 0 carries a root "title" resolves to that title.
func TestBrowseCatalogs_ArrayWrappedTitle(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	content := []byte(`[{"type":"directory","name":"root","title":"Array Title","size":100,"contents":[]},` +
		`{"type":"report","directories":1,"files":0}]`)
	if err := os.WriteFile(filepath.Join(dir, "array.json"), content, 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	if results[0].Title != "Array Title" {
		t.Errorf("Title = %q, want %q", results[0].Title, "Array Title")
	}
}

// TestBrowseCatalogs_UnparseableJSONSkipsTitleProbe proves the title probe
// never masks or replaces the existing parse diagnostic STATE-02 depends
// on -- an invalid catalog still returns a populated ParseError and falls
// back to the HTML-or-filename title.
func TestBrowseCatalogs_UnparseableJSONSkipsTitleProbe(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "broken.json"),
		[]byte(`{not valid json at all`), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.html"),
		[]byte(`<html><head><title>Fallback Title</title></head></html>`), 0644); err != nil {
		t.Fatalf("write html: %v", err)
	}

	results, err := s.BrowseCatalogs(dir)
	if err != nil {
		t.Fatalf("BrowseCatalogs failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 catalog result, got %d", len(results))
	}
	if results[0].ParseError == "" {
		t.Fatal("expected a non-empty ParseError for unparseable JSON")
	}
	if results[0].Title != "Fallback Title" {
		t.Errorf("Title = %q, want %q (HTML fallback, since JSON title probe must be skipped)", results[0].Title, "Fallback Title")
	}
}

// TestLoadCatalogFlat_ToleratesMarkerFields is LoadCatalog's flattening
// sibling: same fixture, same tolerance requirement, plus a node-count
// check against the equivalent marker-free tree.
func TestLoadCatalogFlat_ToleratesMarkerFields(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	withMarkers := []byte(`{"type":"directory","name":"./","size":5,"contents":[` +
		`{"type":"file","name":"./a.txt","size":5,"contents":null},` +
		`{"type":"directory","name":"./gone","size":0,"contents":[],"unreadable":true,"readError":"stat ./gone: no such file or directory"}` +
		`]}`)
	withMarkersPath := filepath.Join(dir, "partial.json")
	if err := os.WriteFile(withMarkersPath, withMarkers, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	withoutMarkers := []byte(`{"type":"directory","name":"./","size":5,"contents":[` +
		`{"type":"file","name":"./a.txt","size":5,"contents":null},` +
		`{"type":"directory","name":"./gone","size":0,"contents":[]}` +
		`]}`)
	withoutMarkersPath := filepath.Join(dir, "clean.json")
	if err := os.WriteFile(withoutMarkersPath, withoutMarkers, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	flatWithMarkers, err := s.LoadCatalogFlat(withMarkersPath)
	if err != nil {
		t.Fatalf("LoadCatalogFlat failed to tolerate marker fields: %v", err)
	}
	flatWithoutMarkers, err := s.LoadCatalogFlat(withoutMarkersPath)
	if err != nil {
		t.Fatalf("LoadCatalogFlat (marker-free) failed: %v", err)
	}

	if len(flatWithMarkers.Nodes) != len(flatWithoutMarkers.Nodes) {
		t.Errorf("node count with markers = %d, without markers = %d, want equal",
			len(flatWithMarkers.Nodes), len(flatWithoutMarkers.Nodes))
	}
}
