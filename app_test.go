package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storcat-wails/internal/catalog"
	"storcat-wails/pkg/models"
)

// TestGetCatalogHtmlPath_ReturnsHtmlPathWhenFileExists verifies that
// GetCatalogHtmlPath returns the .html path when the corresponding .html
// file exists alongside a .json catalog. Addresses API-01.
func TestGetCatalogHtmlPath_ReturnsHtmlPathWhenFileExists(t *testing.T) {
	dir := t.TempDir()

	// Create a .json file and its sibling .html file
	jsonPath := filepath.Join(dir, "mycatalog.json")
	htmlPath := filepath.Join(dir, "mycatalog.html")

	if err := os.WriteFile(jsonPath, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create temp json file: %v", err)
	}
	if err := os.WriteFile(htmlPath, []byte(`<html></html>`), 0644); err != nil {
		t.Fatalf("failed to create temp html file: %v", err)
	}

	app := &App{}
	got, err := app.GetCatalogHtmlPath(jsonPath)
	if err != nil {
		t.Fatalf("GetCatalogHtmlPath returned unexpected error: %v", err)
	}
	if got != htmlPath {
		t.Errorf("GetCatalogHtmlPath = %q, want %q", got, htmlPath)
	}
}

// TestGetCatalogHtmlPath_ReturnsErrorWhenHtmlFileMissing verifies that
// GetCatalogHtmlPath returns an error containing "HTML file not found"
// when the .html counterpart does not exist on disk. Addresses API-01.
func TestGetCatalogHtmlPath_ReturnsErrorWhenHtmlFileMissing(t *testing.T) {
	dir := t.TempDir()

	// Create the .json file only — no .html counterpart
	jsonPath := filepath.Join(dir, "mycatalog.json")
	if err := os.WriteFile(jsonPath, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to create temp json file: %v", err)
	}

	app := &App{}
	got, err := app.GetCatalogHtmlPath(jsonPath)
	if err == nil {
		t.Fatalf("GetCatalogHtmlPath expected an error for missing .html file, got path %q", got)
	}
	if !strings.Contains(err.Error(), "HTML file not found") {
		t.Errorf("error message = %q, want it to contain %q", err.Error(), "HTML file not found")
	}
}

// TestGetCatalogHtmlPath_NonJsonInputAppendsHtmlExtension verifies that
// GetCatalogHtmlPath handles a non-.json input path by appending ".html"
// rather than replacing an extension. Addresses API-01.
func TestGetCatalogHtmlPath_NonJsonInputAppendsHtmlExtension(t *testing.T) {
	dir := t.TempDir()

	// Input has no .json extension — method should append .html
	basePath := filepath.Join(dir, "mycatalog")
	htmlPath := basePath + ".html"

	if err := os.WriteFile(htmlPath, []byte(`<html></html>`), 0644); err != nil {
		t.Fatalf("failed to create temp html file: %v", err)
	}

	app := &App{}
	got, err := app.GetCatalogHtmlPath(basePath)
	if err != nil {
		t.Fatalf("GetCatalogHtmlPath returned unexpected error: %v", err)
	}
	if got != htmlPath {
		t.Errorf("GetCatalogHtmlPath = %q, want %q", got, htmlPath)
	}
}

// TestReadHtmlFile_ReturnsContentForValidFile verifies that ReadHtmlFile
// returns the full content of an existing HTML file as a string. Addresses API-02.
func TestReadHtmlFile_ReturnsContentForValidFile(t *testing.T) {
	dir := t.TempDir()
	htmlPath := filepath.Join(dir, "catalog.html")
	expectedContent := "<html><body>Hello StorCat</body></html>"

	if err := os.WriteFile(htmlPath, []byte(expectedContent), 0644); err != nil {
		t.Fatalf("failed to create temp html file: %v", err)
	}

	app := &App{}
	got, err := app.ReadHtmlFile(htmlPath)
	if err != nil {
		t.Fatalf("ReadHtmlFile returned unexpected error: %v", err)
	}
	if got != expectedContent {
		t.Errorf("ReadHtmlFile content = %q, want %q", got, expectedContent)
	}
}

// TestReadHtmlFile_ReturnsErrorForNonexistentFile verifies that ReadHtmlFile
// returns an error (not empty string, not panic) when the file does not exist.
// Addresses API-02.
func TestReadHtmlFile_ReturnsErrorForNonexistentFile(t *testing.T) {
	dir := t.TempDir()
	missingPath := filepath.Join(dir, "does-not-exist.html")

	app := &App{}
	got, err := app.ReadHtmlFile(missingPath)
	if err == nil {
		t.Fatalf("ReadHtmlFile expected an error for nonexistent file, got content %q", got)
	}
	if got != "" {
		t.Errorf("ReadHtmlFile returned non-empty string on error: %q", got)
	}
}

// TestGetVersion_ReturnsVersionFromWailsJson verifies that GetVersion() returns
// the productVersion embedded from wails.json, not an empty string or "dev".
// In test context, version.go's //go:embed wails.json runs at package init,
// so Version is populated from the real wails.json. Addresses PLAT-02.
func TestGetVersion_ReturnsVersionFromWailsJson(t *testing.T) {
	app := &App{}
	got := app.GetVersion()

	if got == "" {
		t.Fatal("GetVersion() returned empty string; expected non-empty version string")
	}

	// Version is populated by parsing wails.json at init time.
	// The embedded wails.json has productVersion "2.2.1"; if parsing ever fails
	// the fallback is "dev". Neither outcome is an empty string, but we also
	// verify it matches the package-level Version variable directly to confirm
	// the method delegates correctly.
	if got != Version {
		t.Errorf("GetVersion() = %q, but package-level Version = %q; method must return Version", got, Version)
	}

	// Confirm parsing succeeded: the test-time wails.json should yield a semver
	// string, not the "dev" fallback. If this assertion fails, the embedded
	// wails.json or its productVersion field is missing/malformed.
	if got == "dev" {
		t.Errorf("GetVersion() returned fallback %q; wails.json productVersion may be missing or unparseable", got)
	}
}

// TestStartScan_WritesIntoOutputDir verifies that, with a source temp dir
// and a distinct output temp dir, StartScan returns a result whose
// JsonPath and HtmlPath both live under the output dir and both exist on
// disk.
func TestStartScan_WritesIntoOutputDir(t *testing.T) {
	app := &App{catalogService: catalog.NewService()}
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	// StartScan resolves outputDir's symlinks (e.g. macOS's /var ->
	// /private/var) before writing, so the expected directory here must be
	// resolved the same way -- otherwise this comparison fails for the
	// wrong reason on any machine where t.TempDir() lives under a symlink.
	resolvedOutputDir, err := filepath.EvalSymlinks(outputDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(outputDir): %v", err)
	}

	result, err := app.StartScan("Test", sourceDir, outputDir, "out", ScanOptions{WriteHTML: true})
	if err != nil {
		t.Fatalf("StartScan failed: %v", err)
	}
	if filepath.Dir(result.JsonPath) != resolvedOutputDir {
		t.Errorf("JsonPath dir = %q, want %q", filepath.Dir(result.JsonPath), resolvedOutputDir)
	}
	if filepath.Dir(result.HtmlPath) != resolvedOutputDir {
		t.Errorf("HtmlPath dir = %q, want %q", filepath.Dir(result.HtmlPath), resolvedOutputDir)
	}
	if _, err := os.Stat(result.JsonPath); err != nil {
		t.Errorf("expected JSON file to exist: %v", err)
	}
	if _, err := os.Stat(result.HtmlPath); err != nil {
		t.Errorf("expected HTML file to exist: %v", err)
	}
}

// TestStartScan_RejectsEscapingOutputRoot verifies that an outputRoot
// containing a parent-directory segment is rejected with a non-nil error
// and no file is created anywhere.
func TestStartScan_RejectsEscapingOutputRoot(t *testing.T) {
	app := &App{catalogService: catalog.NewService()}
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, err := app.StartScan("Test", sourceDir, outputDir, "../escape", ScanOptions{WriteHTML: true})
	if err == nil {
		t.Fatal("expected an error for an outputRoot escaping outputDir, got nil")
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("read outputDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected outputDir to remain empty, got %+v", entries)
	}
}

// TestStartScan_RejectsSecondConcurrentScan verifies that, while a scan
// handle is held, a second StartScan call returns a non-nil error and does
// not clear the first scan's cancel handle.
func TestStartScan_RejectsSecondConcurrentScan(t *testing.T) {
	app := &App{catalogService: catalog.NewService()}
	sourceDir := t.TempDir()
	outputDir := t.TempDir()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.scanMu.Lock()
	app.activeScanCancel = cancel
	app.scanMu.Unlock()

	_, err := app.StartScan("Test", sourceDir, outputDir, "out", ScanOptions{WriteHTML: true})
	if err == nil {
		t.Fatal("expected an error when a scan is already running, got nil")
	}

	app.scanMu.Lock()
	stillSet := app.activeScanCancel != nil
	app.scanMu.Unlock()
	if !stillSet {
		t.Error("expected the first scan's cancel handle to remain set after a rejected second call")
	}
}

// TestThrottledProgress_NilRuntimeContextIsSafe verifies that the progress
// closure returned by throttledProgress does not panic when the App's
// runtime context is nil, so the binding is testable headlessly.
func TestThrottledProgress_NilRuntimeContextIsSafe(t *testing.T) {
	app := &App{}
	progress := app.throttledProgress(1024)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("throttledProgress panicked with nil ctx: %v", r)
		}
	}()
	progress(catalog.ProgressUpdate{Path: "./a.txt", FilesSeen: 1, BytesSeen: 5})
}

// TestCancelScan_NoActiveScanIsNoOp verifies that calling CancelScan with
// no scan running does not panic and leaves the handle nil.
func TestCancelScan_NoActiveScanIsNoOp(t *testing.T) {
	app := &App{}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CancelScan panicked with no active scan: %v", r)
		}
	}()
	app.CancelScan()

	app.scanMu.Lock()
	handle := app.activeScanCancel
	app.scanMu.Unlock()
	if handle != nil {
		t.Error("expected activeScanCancel to remain nil after a no-op cancel")
	}
}

// TestCancelScan_CancelsTheActiveContext verifies that, with a real
// cancellable context stored on the App, CancelScan closes that context's
// done channel.
func TestCancelScan_CancelsTheActiveContext(t *testing.T) {
	app := &App{}
	ctx, cancel := context.WithCancel(context.Background())
	app.scanMu.Lock()
	app.activeScanCancel = cancel
	app.scanMu.Unlock()

	app.CancelScan()

	select {
	case <-ctx.Done():
	default:
		t.Fatal("expected the stored context to be cancelled")
	}
}

// TestStartScan_RetainsPartialOnSourceLoss verifies that a StartScan whose
// walk hits a source loss returns a non-nil error, writes nothing into the
// output directory, and leaves a non-nil retained partial tree plus the
// originating request parameters on the App. The mid-walk removal is
// triggered via startScan's testHook parameter -- the same deterministic
// technique internal/catalog/service_test.go uses at the service layer --
// since a.throttledProgress's real progress path requires a live Wails
// runtime context this headless test cannot supply.
func TestStartScan_RetainsPartialOnSourceLoss(t *testing.T) {
	app := &App{catalogService: catalog.NewService()}
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

	// subA sorts before subB (directories-first, alphabetical), so it is
	// fully walked first. The hook fires exactly when subA's file is
	// reported -- at that point, remove the scan root itself out from
	// under the walk, simulating a volume disappearing mid-scan.
	var removed bool
	testHook := func(u catalog.ProgressUpdate) {
		if !removed && filepath.Base(u.Path) == "ok.txt" {
			removed = true
			if err := os.RemoveAll(sourceDir); err != nil {
				t.Fatalf("remove sourceDir: %v", err)
			}
		}
	}

	_, err := app.startScan("Test", sourceDir, outputDir, "out", ScanOptions{WriteHTML: true}, testHook)
	if err == nil {
		t.Fatal("expected a non-nil error for a source loss, got nil")
	}
	var srcErr *catalog.SourceUnavailableError
	if !errors.As(err, &srcErr) {
		t.Fatalf("expected an error matching *catalog.SourceUnavailableError, got %v", err)
	}

	entries, rerr := os.ReadDir(outputDir)
	if rerr != nil {
		t.Fatalf("read outputDir: %v", rerr)
	}
	if len(entries) != 0 {
		t.Errorf("expected outputDir to remain empty after a source loss, got %+v", entries)
	}

	app.scanMu.Lock()
	partial := app.lastPartial
	req := app.lastScanReq
	app.scanMu.Unlock()
	if partial == nil {
		t.Fatal("expected a retained partial tree on the App")
	}
	if req == nil {
		t.Fatal("expected retained request parameters on the App")
	}
}

// TestStartScan_ClearsRetainedPartialOnNewScan verifies that starting a
// fresh scan discards any previously retained partial, so a stale tree can
// never be written under a new scan's name.
func TestStartScan_ClearsRetainedPartialOnNewScan(t *testing.T) {
	app := &App{catalogService: catalog.NewService()}
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	app.scanMu.Lock()
	app.lastPartial = &catalog.PartialScan{Tree: &models.CatalogItem{Type: "directory", Name: "./"}}
	app.lastPartialResult = &models.CreateCatalogResult{JsonPath: "/tmp/stale.json"}
	app.lastScanReq = &lastScanRequest{title: "Stale"}
	app.scanMu.Unlock()

	if _, err := app.StartScan("Test", sourceDir, outputDir, "out", ScanOptions{WriteHTML: true}); err != nil {
		t.Fatalf("StartScan failed: %v", err)
	}

	app.scanMu.Lock()
	partial, result, req := app.lastPartial, app.lastPartialResult, app.lastScanReq
	app.scanMu.Unlock()
	if partial != nil {
		t.Errorf("expected lastPartial cleared, got %+v", partial)
	}
	if result != nil {
		t.Errorf("expected lastPartialResult cleared, got %+v", result)
	}
	if req != nil {
		t.Errorf("expected lastScanReq cleared, got %+v", req)
	}
}

// TestWritePartialCatalog_WritesOnce verifies that the first call produces
// the JSON in the output directory; the second call returns the identical
// cached result, creates no additional file, and leaves the first file's
// modification time unchanged.
func TestWritePartialCatalog_WritesOnce(t *testing.T) {
	app := &App{catalogService: catalog.NewService()}
	outputDir := t.TempDir()
	tree := &models.CatalogItem{
		Type: "directory",
		Name: "./",
		Size: 2,
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./a.txt", Size: 2},
		},
	}
	app.lastPartial = &catalog.PartialScan{Tree: tree}
	app.lastScanReq = &lastScanRequest{
		title:      "Test",
		outputDir:  outputDir,
		outputRoot: "partial",
		opts:       catalog.Options{WriteHTML: true},
	}

	result1, err := app.WritePartialCatalog()
	if err != nil {
		t.Fatalf("first WritePartialCatalog failed: %v", err)
	}
	if result1 == nil {
		t.Fatal("expected a non-nil result from the first call")
	}

	info1, err := os.Stat(result1.JsonPath)
	if err != nil {
		t.Fatalf("stat json after first call: %v", err)
	}

	result2, err := app.WritePartialCatalog()
	if err != nil {
		t.Fatalf("second WritePartialCatalog failed: %v", err)
	}
	if result2 != result1 {
		t.Error("expected the second call to return the identical cached result")
	}

	info2, err := os.Stat(result1.JsonPath)
	if err != nil {
		t.Fatalf("stat json after second call: %v", err)
	}
	if !info1.ModTime().Equal(info2.ModTime()) {
		t.Errorf("expected the JSON file's mtime unchanged after the second call: %v vs %v", info1.ModTime(), info2.ModTime())
	}

	app.scanMu.Lock()
	partial := app.lastPartial
	app.scanMu.Unlock()
	if partial != nil {
		t.Error("expected the retained partial tree cleared after a successful write")
	}
}

// TestWritePartialCatalog_WithoutRetainedScanErrors verifies that, with
// nothing retained and nothing previously written, the call returns a
// non-nil error and creates no file.
func TestWritePartialCatalog_WithoutRetainedScanErrors(t *testing.T) {
	app := &App{catalogService: catalog.NewService()}

	result, err := app.WritePartialCatalog()
	if err == nil {
		t.Fatal("expected a non-nil error when nothing is retained")
	}
	if result != nil {
		t.Errorf("expected a nil result, got %+v", result)
	}
}

// TestWritePartialCatalog_MarkerSurvivesToDisk verifies that the written
// JSON carries the unreadable marker on exactly the node the walk marked.
func TestWritePartialCatalog_MarkerSurvivesToDisk(t *testing.T) {
	app := &App{catalogService: catalog.NewService()}
	outputDir := t.TempDir()
	tree := &models.CatalogItem{
		Type: "directory",
		Name: "./",
		Size: 0,
		Contents: []*models.CatalogItem{
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
	app.lastPartial = &catalog.PartialScan{Tree: tree}
	app.lastScanReq = &lastScanRequest{
		title:      "Test",
		outputDir:  outputDir,
		outputRoot: "partial",
		opts:       catalog.Options{WriteHTML: false},
	}

	result, err := app.WritePartialCatalog()
	if err != nil {
		t.Fatalf("WritePartialCatalog failed: %v", err)
	}

	data, err := os.ReadFile(result.JsonPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode written JSON: %v", err)
	}
	contents := decoded["contents"].([]interface{})
	dirNode := contents[0].(map[string]interface{})

	if dirNode["unreadable"] != true {
		t.Errorf("expected the marked node to carry unreadable:true, got %+v", dirNode)
	}
	if dirNode["readError"] != "stat ./gone: no such file or directory" {
		t.Errorf("expected the marked node to carry its readError, got %+v", dirNode)
	}
}
