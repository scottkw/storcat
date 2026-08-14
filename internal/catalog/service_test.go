package catalog

import (
	"context"
	"encoding/json"
	"errors"
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

	item, err := s.traverseDirectory(context.Background(), tmpDir, tmpDir, &walkState{})
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

	item, err := s.traverseDirectory(context.Background(), tmpDir, tmpDir, &walkState{})
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

// TestCreateCatalog_WrapperWritesHTML is the regression guard for the
// zero-value option-struct pitfall: the CLI wrapper path must always
// produce a non-empty HtmlPath, and that file must actually exist on disk.
func TestCreateCatalog_WrapperWritesHTML(t *testing.T) {
	s := NewService()
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	result, err := s.CreateCatalog("Test", tmpDir, "out", "", nil)
	if err != nil {
		t.Fatalf("CreateCatalog failed: %v", err)
	}
	if result.HtmlPath == "" {
		t.Fatal("expected a non-empty HtmlPath from the CLI wrapper")
	}
	if _, err := os.Stat(result.HtmlPath); err != nil {
		t.Errorf("expected HTML file to exist at %s: %v", result.HtmlPath, err)
	}
}

// TestCreateCatalog_WrapperWritesIntoScannedDirectory verifies the wrapper
// writes <root>.json and <root>.html inside the scanned directory, exactly
// as v2.3.0 did.
func TestCreateCatalog_WrapperWritesIntoScannedDirectory(t *testing.T) {
	s := NewService()
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	result, err := s.CreateCatalog("Test", tmpDir, "out", "", nil)
	if err != nil {
		t.Fatalf("CreateCatalog failed: %v", err)
	}

	wantJSON := filepath.Join(tmpDir, "out.json")
	wantHTML := filepath.Join(tmpDir, "out.html")
	if result.JsonPath != wantJSON {
		t.Errorf("JsonPath = %q, want %q", result.JsonPath, wantJSON)
	}
	if result.HtmlPath != wantHTML {
		t.Errorf("HtmlPath = %q, want %q", result.HtmlPath, wantHTML)
	}
}

// TestCreateCatalogWithContext_OutputDirDistinctFromSource verifies that,
// given a source temp dir and a separate output temp dir, the JSON and HTML
// land in the output dir and the source dir gains no new files.
func TestCreateCatalogWithContext_OutputDirDistinctFromSource(t *testing.T) {
	s := NewService()
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	result, err := s.CreateCatalogWithContext(
		context.Background(), "Test", sourceDir, outputDir, "out", "", Options{WriteHTML: true}, nil,
	)
	if err != nil {
		t.Fatalf("CreateCatalogWithContext failed: %v", err)
	}

	if filepath.Dir(result.JsonPath) != outputDir {
		t.Errorf("JSON written to %q, want inside %q", result.JsonPath, outputDir)
	}
	if filepath.Dir(result.HtmlPath) != outputDir {
		t.Errorf("HTML written to %q, want inside %q", result.HtmlPath, outputDir)
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		t.Fatalf("read sourceDir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("sourceDir gained new files, want only the original fixture: %+v", entries)
	}
}

// TestCreateCatalogWithContext_CancelWritesNothing verifies that a context
// cancelled before the call returns an error satisfying
// errors.Is(err, context.Canceled), and the output directory contains zero
// entries afterwards.
func TestCreateCatalogWithContext_CancelWritesNothing(t *testing.T) {
	s := NewService()
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.CreateCatalogWithContext(
		ctx, "Test", sourceDir, outputDir, "out", "", Options{WriteHTML: true}, nil,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected errors.Is(err, context.Canceled), got %v", err)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("read outputDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected outputDir to remain empty after cancellation, got %+v", entries)
	}
}

// TestCreateCatalogWithContext_ProgressCounters verifies the progress
// callback receives monotonically non-decreasing FilesSeen and BytesSeen,
// and a final FilesSeen equal to the result's FileCount.
func TestCreateCatalogWithContext_ProgressCounters(t *testing.T) {
	s := NewService()
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte("hello"), 0644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}

	var updates []ProgressUpdate
	onProgress := func(u ProgressUpdate) {
		updates = append(updates, u)
	}

	result, err := s.CreateCatalogWithContext(
		context.Background(), "Test", sourceDir, outputDir, "out", "", Options{WriteHTML: true}, onProgress,
	)
	if err != nil {
		t.Fatalf("CreateCatalogWithContext failed: %v", err)
	}

	if len(updates) == 0 {
		t.Fatal("expected at least one progress update")
	}

	lastFiles, lastBytes := 0, int64(0)
	for _, u := range updates {
		if u.FilesSeen < lastFiles {
			t.Errorf("FilesSeen went backwards: %d then %d", lastFiles, u.FilesSeen)
		}
		if u.BytesSeen < lastBytes {
			t.Errorf("BytesSeen went backwards: %d then %d", lastBytes, u.BytesSeen)
		}
		lastFiles, lastBytes = u.FilesSeen, u.BytesSeen
	}

	final := updates[len(updates)-1]
	if final.FilesSeen != result.FileCount {
		t.Errorf("final FilesSeen = %d, want result.FileCount = %d", final.FilesSeen, result.FileCount)
	}
}

// TestCreateCatalog_JSONShapeUnchanged verifies that marshalling a
// hand-built tree through the write path yields the exact byte sequence the
// pre-change writer produced for the same tree -- key order, no
// indentation, no new keys (COMPAT-02).
func TestCreateCatalog_JSONShapeUnchanged(t *testing.T) {
	s := NewService()
	tree := &models.CatalogItem{
		Type: "directory",
		Name: "./",
		Size: 5,
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./a.txt", Size: 5},
		},
	}

	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test.json")
	if err := s.writeJSONFile(tree, jsonPath); err != nil {
		t.Fatalf("writeJSONFile failed: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	want := `{"type":"directory","name":"./","size":5,"contents":[{"type":"file","name":"./a.txt","size":5,"contents":null}]}`
	if string(data) != want {
		t.Errorf("JSON shape changed:\ngot:  %s\nwant: %s", string(data), want)
	}
}

// TestCreateCatalogWithContext_IncludeHidden verifies that with the
// include-hidden option off a dotfile is absent from contents; with it on
// the dotfile appears. Default (wrapper) behavior is off.
func TestCreateCatalogWithContext_IncludeHidden(t *testing.T) {
	s := NewService()

	buildDir := func(t *testing.T) string {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("h"), 0644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("v"), 0644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		return dir
	}

	t.Run("off by default", func(t *testing.T) {
		sourceDir := buildDir(t)
		outputDir := t.TempDir()
		result, err := s.CreateCatalogWithContext(
			context.Background(), "Test", sourceDir, outputDir, "out", "", Options{WriteHTML: true}, nil,
		)
		if err != nil {
			t.Fatalf("CreateCatalogWithContext failed: %v", err)
		}
		if result.FileCount != 1 {
			t.Errorf("FileCount = %d, want 1 (hidden file excluded)", result.FileCount)
		}
	})

	t.Run("on when requested", func(t *testing.T) {
		sourceDir := buildDir(t)
		outputDir := t.TempDir()
		result, err := s.CreateCatalogWithContext(
			context.Background(), "Test", sourceDir, outputDir, "out", "", Options{WriteHTML: true, IncludeHidden: true}, nil,
		)
		if err != nil {
			t.Fatalf("CreateCatalogWithContext failed: %v", err)
		}
		if result.FileCount != 2 {
			t.Errorf("FileCount = %d, want 2 (hidden file included)", result.FileCount)
		}
	})

	t.Run("wrapper default is off", func(t *testing.T) {
		sourceDir := buildDir(t)
		result, err := s.CreateCatalog("Test", sourceDir, "out", "", nil)
		if err != nil {
			t.Fatalf("CreateCatalog failed: %v", err)
		}
		if result.FileCount != 1 {
			t.Errorf("FileCount = %d, want 1 (wrapper default excludes hidden files)", result.FileCount)
		}
	})
}
