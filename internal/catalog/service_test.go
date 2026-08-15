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
	if _, err := s.writeJSONFile(catalog, jsonPath); err != nil {
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
	if _, err := s.writeJSONFile(tree, jsonPath); err != nil {
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

// TestTraverseDirectory_SingleEntryErrorSkipsAndContinues verifies that one
// unreadable child among readable siblings is dropped from contents, no
// error is returned, and the read-error counter increments by exactly one --
// today's v2.3.0 skip-and-continue behavior, unchanged.
func TestTraverseDirectory_SingleEntryErrorSkipsAndContinues(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok1.txt"), []byte("a"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ok2.txt"), []byte("b"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	brokenLink := filepath.Join(dir, "broken")
	if err := os.Symlink(filepath.Join(dir, "does-not-exist"), brokenLink); err != nil {
		t.Skipf("symlinks not supported on this system: %v", err)
	}

	st := &walkState{scanRoot: dir}
	item, err := s.traverseDirectory(context.Background(), dir, dir, st)
	if err != nil {
		t.Fatalf("traverseDirectory failed: %v", err)
	}
	if len(item.Contents) != 2 {
		t.Fatalf("expected 2 readable siblings, got %d: %+v", len(item.Contents), item.Contents)
	}
	for _, c := range item.Contents {
		if filepath.Base(c.Name) == "broken" {
			t.Errorf("unreadable child must not appear in contents: %+v", c)
		}
	}
	if st.readErrors != 1 {
		t.Errorf("readErrors = %d, want 1", st.readErrors)
	}
}

// TestTraverseDirectory_TerminalSourceLossStopsWalk verifies that, with
// HaltOnSourceLoss enabled, a read failure that also takes the scan root
// itself unreachable is classified as terminal: the call returns an error
// errors.As-matching *SourceUnavailableError, the carried partial tree
// contains only the nodes walked before the loss, and no directory after
// the failure point is descended into.
func TestTraverseDirectory_TerminalSourceLossStopsWalk(t *testing.T) {
	s := NewService()
	root := t.TempDir()
	subA := filepath.Join(root, "subA")
	subB := filepath.Join(root, "subB")
	subC := filepath.Join(root, "subC")
	for _, d := range []string{subA, subB, subC} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(subA, "ok.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subC, "other.txt"), []byte("z"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// subA sorts before subB and subC (directories-first, alphabetical), so
	// it is fully walked first. The progress callback fires exactly when
	// subA's file is seen -- at that point, remove the scan root itself out
	// from under the walk, simulating a volume disappearing mid-scan.
	// subB is processed next: its own os.Stat now fails because its parent
	// (root) is gone, triggering classification. subC must never be
	// reached.
	var removed bool
	st := &walkState{
		scanRoot: root,
		opts:     Options{HaltOnSourceLoss: true},
		onProgress: func(u ProgressUpdate) {
			if !removed && filepath.Base(u.Path) == "ok.txt" {
				removed = true
				if err := os.RemoveAll(root); err != nil {
					t.Fatalf("remove root: %v", err)
				}
			}
		},
	}

	item, err := s.traverseDirectory(context.Background(), root, root, st)

	var srcErr *SourceUnavailableError
	if !errors.As(err, &srcErr) {
		t.Fatalf("expected an error matching *SourceUnavailableError, got %v", err)
	}
	if item == nil {
		t.Fatal("expected a non-nil partial node")
	}
	if len(item.Contents) != 1 {
		t.Fatalf("expected exactly 1 node walked before the loss (subA), got %d: %+v", len(item.Contents), item.Contents)
	}
	if filepath.Base(item.Contents[0].Name) != "subA" {
		t.Errorf("expected subA in the partial tree, got %+v", item.Contents[0])
	}
	if !st.terminal {
		t.Error("expected walkState.terminal to be set")
	}
}

// TestCreateCatalogWithContext_SourceLossWritesNothing verifies that
// through the public entry point, a terminal source loss leaves the output
// directory empty -- the write path is never reached.
func TestCreateCatalogWithContext_SourceLossWritesNothing(t *testing.T) {
	s := NewService()
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	subA := filepath.Join(sourceDir, "subA")
	subB := filepath.Join(sourceDir, "subB")
	if err := os.MkdirAll(subA, 0755); err != nil {
		t.Fatalf("mkdir subA: %v", err)
	}
	if err := os.MkdirAll(subB, 0755); err != nil {
		t.Fatalf("mkdir subB: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subA, "ok.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var removed bool
	onProgress := func(u ProgressUpdate) {
		if !removed && filepath.Base(u.Path) == "ok.txt" {
			removed = true
			if err := os.RemoveAll(sourceDir); err != nil {
				t.Fatalf("remove sourceDir: %v", err)
			}
		}
	}

	_, err := s.CreateCatalogWithContext(
		context.Background(), "Test", sourceDir, outputDir, "out", "",
		Options{WriteHTML: true, HaltOnSourceLoss: true}, onProgress,
	)

	var srcErr *SourceUnavailableError
	if !errors.As(err, &srcErr) {
		t.Fatalf("expected an error matching *SourceUnavailableError, got %v", err)
	}
	if srcErr.Partial == nil {
		t.Fatal("expected a populated PartialScan on the returned error")
	}
	if srcErr.Partial.FilesSeen != 1 {
		t.Errorf("Partial.FilesSeen = %d, want 1", srcErr.Partial.FilesSeen)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("read outputDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected outputDir to remain empty after a source-loss stop, got %+v", entries)
	}
}

// TestCreateCatalogWithContext_RootVanishesBeforeAnyProgress verifies the
// "instant, total disconnect" case (25-UI-SPEC.md E6's zero-one-many
// resolution): the scan root is already gone before the walk's very first
// os.Stat call ever succeeds -- no child was ever read, so onProgress never
// fires. This is the scan-root's own top-of-function stat failure, which has
// no parent loop to run it through recordReadError/classify() the way a
// child's failure already does; CreateCatalogWithContext must apply that
// classification itself (found live-verifying 25-07's error state -- prior
// to this fix, this case fell through as a generic "failed to traverse
// directory" error instead of *SourceUnavailableError).
func TestCreateCatalogWithContext_RootVanishesBeforeAnyProgress(t *testing.T) {
	s := NewService()
	parent := t.TempDir()
	sourceDir := filepath.Join(parent, "gone-before-scan-starts")
	outputDir := t.TempDir()

	// Never created -- the walk's first os.Stat(sourceDir) fails immediately,
	// with zero prior progress and zero prior read errors.
	_, err := s.CreateCatalogWithContext(
		context.Background(), "Test", sourceDir, outputDir, "out", "",
		Options{WriteHTML: true, HaltOnSourceLoss: true}, nil,
	)

	var srcErr *SourceUnavailableError
	if !errors.As(err, &srcErr) {
		t.Fatalf("expected an error matching *SourceUnavailableError, got %v", err)
	}
	if srcErr.SourcePath != sourceDir {
		t.Errorf("SourcePath = %q, want %q", srcErr.SourcePath, sourceDir)
	}
	if srcErr.Partial == nil {
		t.Fatal("expected a populated PartialScan on the returned error")
	}
	if srcErr.Partial.FilesSeen != 0 {
		t.Errorf("Partial.FilesSeen = %d, want 0", srcErr.Partial.FilesSeen)
	}
	if len(srcErr.Partial.ReadErrors) != 0 {
		t.Errorf("Partial.ReadErrors = %+v, want empty", srcErr.Partial.ReadErrors)
	}
	if srcErr.Partial.Tree == nil {
		t.Fatal("expected a non-nil Tree -- a nil Tree would marshal to JSON null, not a valid catalog")
	}
	if !srcErr.Partial.Tree.Unreadable {
		t.Error("expected the root node to carry the Unreadable marker")
	}
	if srcErr.Partial.Tree.ReadError == "" {
		t.Error("expected a non-empty ReadError reason on the root node")
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("read outputDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected outputDir to remain empty, got %+v", entries)
	}
}

// TestCreateCatalog_WrapperDoesNotHaltOnSourceLoss verifies that the CLI
// wrapper leaves HaltOnSourceLoss false, so a single-entry failure -- even
// one that would classify as terminal if halting were enabled -- reproduces
// v2.3.0's skip-and-continue behavior exactly and returns no error.
func TestCreateCatalog_WrapperDoesNotHaltOnSourceLoss(t *testing.T) {
	s := NewService()
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "ok.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	brokenLink := filepath.Join(sourceDir, "broken")
	if err := os.Symlink(filepath.Join(sourceDir, "does-not-exist"), brokenLink); err != nil {
		t.Skipf("symlinks not supported on this system: %v", err)
	}

	result, err := s.CreateCatalog("Test", sourceDir, "out", "", nil)
	if err != nil {
		t.Fatalf("CreateCatalog (CLI wrapper) failed: %v, want no error (matches v2.3.0 skip-and-continue)", err)
	}
	if result.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1 (unreadable entry silently skipped)", result.FileCount)
	}
}

// TestCreateCatalog_WrapperRootVanishReturnsPlainError verifies COMPAT-03
// for the root-vanish classification added alongside 25-07's error state: a
// nonexistent source path through the CLI wrapper (HaltOnSourceLoss always
// false) returns v2.3.0's exact plain error, never *SourceUnavailableError
// -- st.classify() itself short-circuits on !HaltOnSourceLoss, so this new
// classification is unreachable from the CLI path by construction.
func TestCreateCatalog_WrapperRootVanishReturnsPlainError(t *testing.T) {
	s := NewService()
	parent := t.TempDir()
	sourceDir := filepath.Join(parent, "does-not-exist")

	_, err := s.CreateCatalog("Test", sourceDir, "out", "", nil)
	if err == nil {
		t.Fatal("expected an error for a nonexistent source directory")
	}
	var srcErr *SourceUnavailableError
	if errors.As(err, &srcErr) {
		t.Fatalf("CLI wrapper must never classify as source-loss, got %v", err)
	}
}

// TestWritePartialCatalog_Marker verifies that writing a partial tree
// through the shared write path produces JSON in which exactly the marked
// directory node carries the marker keys, and no other node does.
func TestWritePartialCatalog_Marker(t *testing.T) {
	s := NewService()
	tree := &models.CatalogItem{
		Type: "directory",
		Name: "./",
		Size: 5,
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./a.txt", Size: 5},
			{
				Type:       "directory",
				Name:       "./gone",
				Size:       0,
				Contents:   []*models.CatalogItem{},
				Unreadable: true,
				ReadError:  "stat ./gone: no such file or directory",
			},
		},
	}

	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test.json")
	if _, err := s.writeJSONFile(tree, jsonPath); err != nil {
		t.Fatalf("writeJSONFile failed: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode written JSON: %v", err)
	}
	contents := decoded["contents"].([]interface{})
	fileNode := contents[0].(map[string]interface{})
	dirNode := contents[1].(map[string]interface{})

	if _, ok := fileNode["unreadable"]; ok {
		t.Errorf("clean node must not carry the unreadable key: %+v", fileNode)
	}
	if _, ok := fileNode["readError"]; ok {
		t.Errorf("clean node must not carry the readError key: %+v", fileNode)
	}
	if dirNode["unreadable"] != true {
		t.Errorf("marked node must carry unreadable:true, got %+v", dirNode)
	}
	if dirNode["readError"] != "stat ./gone: no such file or directory" {
		t.Errorf("marked node must carry its readError, got %+v", dirNode)
	}
	if decoded["unreadable"] != nil {
		t.Errorf("root node must not carry the marker keys, got %+v", decoded)
	}
}

// TestCopyFile_CopiesContent verifies copyFile produces a destination with
// byte-identical content and returns the correct copied-byte count.
// Addresses CR-01.
func TestCopyFile_CopiesContent(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	src := filepath.Join(dir, "src.json")
	want := []byte(`{"hello":"world"}`)
	if err := os.WriteFile(src, want, 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	dst := filepath.Join(dir, "dst.json")
	n, err := s.copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}
	if n != int64(len(want)) {
		t.Errorf("copyFile returned %d bytes, want %d", n, len(want))
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("dst content = %q, want %q", got, want)
	}
}

// TestCopyFile_PreservesExistingDestinationOnFailure is the CR-01
// regression test: copyFile must never truncate an existing, good
// destination file before a copy is known to succeed. With the old
// os.Create+io.Copy implementation, os.Create truncated dst immediately --
// so a write failure (here: an unwritable destination directory) would
// leave dst empty, destroying a previously-good secondary copy. Routed
// through WriteFileAtomic, a failed write must leave the original dst
// content completely untouched.
func TestCopyFile_PreservesExistingDestinationOnFailure(t *testing.T) {
	s := NewService()
	dir := t.TempDir()

	src := filepath.Join(dir, "src.json")
	if err := os.WriteFile(src, []byte("new data"), 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	dstDir := filepath.Join(dir, "dst-dir")
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("mkdir dstDir: %v", err)
	}
	dst := filepath.Join(dstDir, "dst.json")
	previousGood := []byte("PREVIOUSLY GOOD DATA")
	if err := os.WriteFile(dst, previousGood, 0644); err != nil {
		t.Fatalf("seed dst: %v", err)
	}

	// Make dstDir read-only so WriteFileAtomic's os.CreateTemp (and thus the
	// whole copy) fails before ever touching the existing dst file.
	if err := os.Chmod(dstDir, 0555); err != nil {
		t.Fatalf("chmod dstDir read-only: %v", err)
	}
	defer os.Chmod(dstDir, 0755) // restore so t.TempDir() cleanup can remove it

	if _, err := s.copyFile(src, dst); err == nil {
		t.Fatal("expected copyFile to fail against a read-only destination directory")
	}

	os.Chmod(dstDir, 0755)
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst after failed copy: %v", err)
	}
	if string(got) != string(previousGood) {
		t.Errorf("dst content changed after failed copy: got %q, want untouched %q", got, previousGood)
	}
}
