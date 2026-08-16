package catalog

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDuplicateCatalog_FirstCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(src, []byte(`{"title":"photos"}`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	newPath, err := DuplicateCatalog(src)
	if err != nil {
		t.Fatalf("DuplicateCatalog: %v", err)
	}
	want := filepath.Join(dir, "photos-copy.json")
	if newPath != want {
		t.Errorf("newPath = %q, want %q", newPath, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected %s to exist: %v", want, err)
	}
}

func TestDuplicateCatalog_SecondCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(src, []byte(`{"title":"photos"}`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "photos-copy.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("write existing copy: %v", err)
	}

	newPath, err := DuplicateCatalog(src)
	if err != nil {
		t.Fatalf("DuplicateCatalog: %v", err)
	}
	want := filepath.Join(dir, "photos-copy-2.json")
	if newPath != want {
		t.Errorf("newPath = %q, want %q", newPath, want)
	}
}

func TestDuplicateCatalog_ThirdCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(src, []byte(`{"title":"photos"}`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "photos-copy.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("write existing copy 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "photos-copy-2.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("write existing copy 2: %v", err)
	}

	newPath, err := DuplicateCatalog(src)
	if err != nil {
		t.Fatalf("DuplicateCatalog: %v", err)
	}
	want := filepath.Join(dir, "photos-copy-3.json")
	if newPath != want {
		t.Errorf("newPath = %q, want %q", newPath, want)
	}
}

// TestDuplicateCatalog_SkipsRootWithOrphanHTML proves a candidate root is
// only free when NEITHER its .json NOR its .html exists -- an orphaned
// .html (e.g. left behind by a partial delete) must never be clobbered by a
// duplicate that only checked the .json extension.
func TestDuplicateCatalog_SkipsRootWithOrphanHTML(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(src, []byte(`{"title":"photos"}`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	// Only photos-copy.html exists (orphan) -- no photos-copy.json.
	if err := os.WriteFile(filepath.Join(dir, "photos-copy.html"), []byte(`<html></html>`), 0644); err != nil {
		t.Fatalf("write orphan html: %v", err)
	}

	newPath, err := DuplicateCatalog(src)
	if err != nil {
		t.Fatalf("DuplicateCatalog: %v", err)
	}
	want := filepath.Join(dir, "photos-copy-2.json")
	if newPath != want {
		t.Errorf("newPath = %q, want %q (orphan .html at -copy root must not be clobbered)", newPath, want)
	}
	// The orphan HTML must survive untouched.
	data, err := os.ReadFile(filepath.Join(dir, "photos-copy.html"))
	if err != nil {
		t.Fatalf("orphan html should still exist: %v", err)
	}
	if string(data) != "<html></html>" {
		t.Errorf("orphan html was modified: %q", data)
	}
}

func TestDuplicateCatalog_CopiesHTMLWhenPresent(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	htmlSrc := filepath.Join(dir, "photos.html")
	if err := os.WriteFile(src, []byte(`{"title":"photos"}`), 0644); err != nil {
		t.Fatalf("write source json: %v", err)
	}
	htmlContent := []byte("<html><title>photos</title></html>")
	if err := os.WriteFile(htmlSrc, htmlContent, 0644); err != nil {
		t.Fatalf("write source html: %v", err)
	}

	if _, err := DuplicateCatalog(src); err != nil {
		t.Fatalf("DuplicateCatalog: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "photos-copy.html"))
	if err != nil {
		t.Fatalf("expected photos-copy.html to exist: %v", err)
	}
	if !bytes.Equal(got, htmlContent) {
		t.Errorf("photos-copy.html = %q, want byte-identical to source %q", got, htmlContent)
	}
}

func TestDuplicateCatalog_NoHTMLCopiesOnlyJSON(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(src, []byte(`{"title":"photos"}`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	newPath, err := DuplicateCatalog(src)
	if err != nil {
		t.Fatalf("DuplicateCatalog: %v", err)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("expected new json to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "photos-copy.html")); !os.IsNotExist(err) {
		t.Errorf("expected no photos-copy.html to be created, stat err = %v", err)
	}
}

func TestDuplicateCatalog_IsByteIdentical(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	original := []byte(`{"title":"My Photos","contents":[{"type":"file","name":"a.jpg","size":123}]}`)
	if err := os.WriteFile(src, original, 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	newPath, err := DuplicateCatalog(src)
	if err != nil {
		t.Fatalf("DuplicateCatalog: %v", err)
	}
	got, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("read duplicated file: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Errorf("duplicated bytes = %q, want byte-identical to source %q", got, original)
	}
}

func TestDuplicateCatalog_LeavesSourceUntouched(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	htmlSrc := filepath.Join(dir, "photos.html")
	originalJSON := []byte(`{"title":"photos"}`)
	originalHTML := []byte("<html><title>photos</title></html>")
	if err := os.WriteFile(src, originalJSON, 0644); err != nil {
		t.Fatalf("write source json: %v", err)
	}
	if err := os.WriteFile(htmlSrc, originalHTML, 0644); err != nil {
		t.Fatalf("write source html: %v", err)
	}

	if _, err := DuplicateCatalog(src); err != nil {
		t.Fatalf("DuplicateCatalog: %v", err)
	}

	gotJSON, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read source json after duplicate: %v", err)
	}
	if !bytes.Equal(gotJSON, originalJSON) {
		t.Errorf("source json changed: got %q, want %q", gotJSON, originalJSON)
	}
	gotHTML, err := os.ReadFile(htmlSrc)
	if err != nil {
		t.Fatalf("read source html after duplicate: %v", err)
	}
	if !bytes.Equal(gotHTML, originalHTML) {
		t.Errorf("source html changed: got %q, want %q", gotHTML, originalHTML)
	}
}

func TestDuplicateCatalog_RejectsNonJSON(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.txt")
	if err := os.WriteFile(src, []byte("not a catalog"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	_, err := DuplicateCatalog(src)
	if err == nil {
		t.Fatal("expected an error for a non-.json source, got nil")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected no new files to be created, dir has %d entries", len(entries))
	}
}

func TestDuplicateCatalog_RejectsMissingSource(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "does-not-exist.json")

	_, err := DuplicateCatalog(src)
	if err == nil {
		t.Fatal("expected an error for a missing source, got nil")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no files to be created, dir has %d entries", len(entries))
	}
}

// WR-03: a symlink named "<root>.html" pointing outside jsonPath's own
// directory must not be followed as the duplicate's source html. Mirrors
// TestRenameCatalog_RejectsHTMLSymlinkEscapingCatalogDir (rename_test.go)
// and internal/osutil/reveal_test.go's "symlink pointing outside
// catalogDir" case.
func TestDuplicateCatalog_RejectsHTMLSymlinkEscapingCatalogDir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photos.json")
	if err := os.WriteFile(src, []byte(`{"title":"photos"}`), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.html")
	if err := os.WriteFile(outsideFile, []byte("secret contents"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	htmlPath := filepath.Join(dir, "photos.html")
	if err := os.Symlink(outsideFile, htmlPath); err != nil {
		t.Skipf("symlinks unavailable in this environment: %v", err)
	}

	newJSONPath, err := DuplicateCatalog(src)
	if err == nil {
		t.Fatal("expected an error for an .html sibling that resolves outside the catalog directory")
	}

	// Per DuplicateCatalog's own doc comment, the .json copy is written
	// before the .html sibling is even looked at and is never rolled back
	// on a later failure -- so newJSONPath is still reported and still
	// exists; only the "-copy.html" that would have carried the symlink's
	// escape must never be created.
	if newJSONPath == "" {
		t.Fatal("expected the already-succeeded json copy path to still be reported")
	}
	if _, statErr := os.Stat(newJSONPath); statErr != nil {
		t.Errorf("expected the json copy to exist despite the html-sibling rejection: %v", statErr)
	}
	if _, statErr := os.Stat(newJSONPath[:len(newJSONPath)-len(".json")] + ".html"); !os.IsNotExist(statErr) {
		t.Errorf("expected no html copy to be created after the symlink-escape rejection, stat err: %v", statErr)
	}
}
