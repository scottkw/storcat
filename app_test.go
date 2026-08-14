package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storcat-wails/internal/catalog"
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
