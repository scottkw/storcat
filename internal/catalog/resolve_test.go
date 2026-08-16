package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storcat-wails/pkg/models"
)

func newTestTree(name string) *models.CatalogItem {
	return &models.CatalogItem{
		Type: "directory",
		Name: "./",
		Size: 1,
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./" + name, Size: 1},
		},
	}
}

// TestWriteRescanResult_OverwriteReplacesInPlace proves overwrite lands the
// new tree back at the ORIGINAL path -- no "-copy" file is produced.
func TestWriteRescanResult_OverwriteReplacesInPlace(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(jsonPath, []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("seed original: %v", err)
	}

	svc := NewService()
	result, err := svc.WriteRescanResult(newTestTree("new-file.txt"), "photos", jsonPath, ResolveOverwrite, Options{})
	if err != nil {
		t.Fatalf("WriteRescanResult: %v", err)
	}
	if result.JsonPath != jsonPath {
		t.Errorf("JsonPath = %q, want %q", result.JsonPath, jsonPath)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read overwritten file: %v", err)
	}
	if !strings.Contains(string(data), "new-file.txt") {
		t.Errorf("overwritten file does not contain the new tree's content: %s", data)
	}

	if _, err := os.Stat(filepath.Join(dir, "photos-copy.json")); !os.IsNotExist(err) {
		t.Errorf("expected no photos-copy.json to be created by overwrite, stat err = %v", err)
	}
}

// TestWriteRescanResult_KeepBothUsesCopySuffix proves keep-both runs the
// SAME shared collision loop Duplicate uses -- with the first candidate
// pre-created, the write lands on the SECOND candidate, not a naive
// single-suffix append.
func TestWriteRescanResult_KeepBothUsesCopySuffix(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(jsonPath, []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("seed original: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "photos-copy.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("seed first candidate: %v", err)
	}

	svc := NewService()
	result, err := svc.WriteRescanResult(newTestTree("new-file.txt"), "photos", jsonPath, ResolveKeepBoth, Options{})
	if err != nil {
		t.Fatalf("WriteRescanResult: %v", err)
	}
	want := filepath.Join(dir, "photos-copy-2.json")
	if result.JsonPath != want {
		t.Errorf("JsonPath = %q, want %q", result.JsonPath, want)
	}

	// The original must be untouched -- keep-both never writes over it.
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if strings.Contains(string(data), "new-file.txt") {
		t.Errorf("original was overwritten by a keep-both write: %s", data)
	}
}

// TestWriteRescanResult_RewritesHtmlWhenPresent proves the .html sibling is
// rewritten when the original catalog had one.
func TestWriteRescanResult_RewritesHtmlWhenPresent(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "photos.json")
	htmlPath := filepath.Join(dir, "photos.html")
	if err := os.WriteFile(jsonPath, []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("seed original json: %v", err)
	}
	if err := os.WriteFile(htmlPath, []byte(`<html>old</html>`), 0644); err != nil {
		t.Fatalf("seed original html: %v", err)
	}

	svc := NewService()
	result, err := svc.WriteRescanResult(newTestTree("new-file.txt"), "photos", jsonPath, ResolveOverwrite, Options{})
	if err != nil {
		t.Fatalf("WriteRescanResult: %v", err)
	}
	if result.HtmlPath != htmlPath {
		t.Errorf("HtmlPath = %q, want %q", result.HtmlPath, htmlPath)
	}
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read rewritten html: %v", err)
	}
	if strings.Contains(string(data), "old") {
		t.Errorf("html was not rewritten: %s", data)
	}
}

// TestWriteRescanResult_DoesNotCreateHtmlWhenAbsent proves an .html is never
// created where none existed on the original.
func TestWriteRescanResult_DoesNotCreateHtmlWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(jsonPath, []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`), 0644); err != nil {
		t.Fatalf("seed original: %v", err)
	}

	svc := NewService()
	result, err := svc.WriteRescanResult(newTestTree("new-file.txt"), "photos", jsonPath, ResolveOverwrite, Options{})
	if err != nil {
		t.Fatalf("WriteRescanResult: %v", err)
	}
	if result.HtmlPath != "" {
		t.Errorf("HtmlPath = %q, want empty (no .html should be produced)", result.HtmlPath)
	}
	if _, err := os.Stat(filepath.Join(dir, "photos.html")); !os.IsNotExist(err) {
		t.Errorf("expected no photos.html to be created, stat err = %v", err)
	}
}
