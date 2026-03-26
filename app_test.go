package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
