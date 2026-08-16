package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storcat-wails/pkg/models"
)

// v2BareObjectFixture is a minimal, realistic v2.0.0 bare-object catalog:
// a root directory with one nested file, matching the shape writeJSONFile
// actually produces (no title key -- title is added by RenameCatalog).
const v2BareObjectFixture = `{"type":"directory","name":"Photos","size":100,"contents":[{"type":"file","name":"Photos/a.jpg","size":100,"contents":[]}]}`

// v1ArrayEnvelopeFixture mimics the legacy bash-script/`tree -J` output: the
// tree root as element 0, followed by a trailing report object at element 1
// -- the exact shape TestRenameCatalog_PreservesArrayEnvelope must survive.
const v1ArrayEnvelopeFixture = `[{"type":"directory","name":"Photos","contents":[{"type":"file","name":"Photos/a.jpg","size":100,"contents":[]}],"size":100},{"type":"report","directories":1,"files":1}]`

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
	return path
}

// realisticHTMLFixture mirrors writeHTMLFile's actual output shape closely
// enough to exercise both tag sites and the byte-identical-elsewhere
// requirement: tree structure, counts and the VERSION footer.
func realisticHTMLFixture(title string) string {
	return `<!DOCTYPE html>
<html>
<head>
 <title>` + title + `</title>
</head>
<body>
	<h1>` + title + `</h1><p>
	tree structure here
	<br><br>
	</p>
	<p>
 100B used in 1 directories, 1 files
	<br><br>
	</p>
	<hr>
	<p class="VERSION">
		 tree v1.7.0 (c) 1996 - 2014 by Steve Baker and Thomas Moore
	</p>
</body>
</html>`
}

func TestRenameCatalog_WritesJSONTitle(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)

	if err := RenameCatalog(jsonPath, "Photos 2024"); err != nil {
		t.Fatalf("RenameCatalog: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	var item models.CatalogItem
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if item.Title != "Photos 2024" {
		t.Errorf("Title = %q, want %q", item.Title, "Photos 2024")
	}
	if item.Type != "directory" || item.Name != "Photos" || item.Size != 100 {
		t.Errorf("unexpected root fields after rename: %+v", item)
	}
	if len(item.Contents) != 1 || item.Contents[0].Name != "Photos/a.jpg" {
		t.Errorf("unexpected contents after rename: %+v", item.Contents)
	}
}

func TestRenameCatalog_PreservesArrayEnvelope(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v1ArrayEnvelopeFixture)

	var before []json.RawMessage
	if err := json.Unmarshal([]byte(v1ArrayEnvelopeFixture), &before); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if err := RenameCatalog(jsonPath, "Photos 2024"); err != nil {
		t.Fatalf("RenameCatalog: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	var after []json.RawMessage
	if err := json.Unmarshal(data, &after); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(after) != 2 {
		t.Fatalf("array length = %d, want 2", len(after))
	}
	if string(after[1]) != string(before[1]) {
		t.Errorf("trailing report element changed:\nbefore: %s\nafter:  %s", before[1], after[1])
	}
}

func TestRenameCatalog_PreservesNestedContentsBytes(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)

	var before map[string]json.RawMessage
	if err := json.Unmarshal([]byte(v2BareObjectFixture), &before); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if err := RenameCatalog(jsonPath, "Photos 2024"); err != nil {
		t.Fatalf("RenameCatalog: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	var after map[string]json.RawMessage
	if err := json.Unmarshal(data, &after); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if string(after["contents"]) != string(before["contents"]) {
		t.Errorf("contents bytes changed:\nbefore: %s\nafter:  %s", before["contents"], after["contents"])
	}
}

func TestRenameCatalog_IsIdempotentOnKey(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)

	if err := RenameCatalog(jsonPath, "First Title"); err != nil {
		t.Fatalf("first RenameCatalog: %v", err)
	}
	if err := RenameCatalog(jsonPath, "Second Title"); err != nil {
		t.Fatalf("second RenameCatalog: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	root := string(data)
	// Only the root object's own key list matters -- count occurrences of
	// the literal key token, not any nested value that might contain the
	// substring "title" by coincidence (none do here, but this is the
	// documented intent).
	count := strings.Count(root, `"title":`)
	if count != 1 {
		t.Errorf(`occurrences of "title": = %d, want 1 (root: %s)`, count, root)
	}

	var item models.CatalogItem
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if item.Title != "Second Title" {
		t.Errorf("Title = %q, want %q", item.Title, "Second Title")
	}
}

func TestRenameCatalog_RewritesBothHTMLOccurrences(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)
	htmlBefore := realisticHTMLFixture("Photos")
	htmlPath := writeFixture(t, dir, "photos.html", htmlBefore)

	if err := RenameCatalog(jsonPath, "Photos 2024"); err != nil {
		t.Fatalf("RenameCatalog: %v", err)
	}

	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read html result: %v", err)
	}
	got := string(data)

	if !strings.Contains(got, "<title>Photos 2024</title>") {
		t.Errorf("<title> not updated: %s", got)
	}
	if !strings.Contains(got, "<h1>Photos 2024</h1>") {
		t.Errorf("<h1> not updated: %s", got)
	}
	if !strings.Contains(got, "tree structure here") {
		t.Errorf("tree structure body changed unexpectedly: %s", got)
	}
	if !strings.Contains(got, "100B used in 1 directories, 1 files") {
		t.Errorf("counts changed unexpectedly: %s", got)
	}
	if !strings.Contains(got, `class="VERSION"`) {
		t.Errorf("VERSION footer changed unexpectedly: %s", got)
	}
}

func TestRenameCatalog_EscapesHTMLSpecials(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)
	writeFixture(t, dir, "photos.html", realisticHTMLFixture("Photos"))

	title := `Tom & Jerry <2024>`
	if err := RenameCatalog(jsonPath, title); err != nil {
		t.Fatalf("RenameCatalog: %v", err)
	}

	htmlData, err := os.ReadFile(filepath.Join(dir, "photos.html"))
	if err != nil {
		t.Fatalf("read html result: %v", err)
	}
	htmlGot := string(htmlData)
	want := "Tom &amp; Jerry &lt;2024&gt;"
	if strings.Count(htmlGot, want) != 2 {
		t.Errorf("expected escaped title twice (title+h1), got: %s", htmlGot)
	}
	if strings.Contains(htmlGot, title) {
		t.Errorf("raw unescaped title leaked into HTML: %s", htmlGot)
	}

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json result: %v", err)
	}
	var item models.CatalogItem
	if err := json.Unmarshal(jsonData, &item); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if item.Title != title {
		t.Errorf("JSON title = %q, want raw unescaped %q", item.Title, title)
	}
}

func TestRenameCatalog_NoHTMLIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)

	if err := RenameCatalog(jsonPath, "Photos 2024"); err != nil {
		t.Fatalf("RenameCatalog: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "photos.html")); !os.IsNotExist(err) {
		t.Errorf("expected no .html to be created, stat err = %v", err)
	}
}

func TestRenameCatalog_HTMLWithoutH1(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)
	titleOnly := "<html><head><title>Photos</title></head><body>no heading here</body></html>"
	htmlPath := writeFixture(t, dir, "photos.html", titleOnly)

	if err := RenameCatalog(jsonPath, "Photos 2024"); err != nil {
		t.Fatalf("RenameCatalog: %v", err)
	}

	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read html result: %v", err)
	}
	got := string(data)
	want := "<html><head><title>Photos 2024</title></head><body>no heading here</body></html>"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenameCatalog_RejectsEmptyTitle(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)
	htmlPath := writeFixture(t, dir, "photos.html", realisticHTMLFixture("Photos"))

	jsonBefore, _ := os.ReadFile(jsonPath)
	htmlBefore, _ := os.ReadFile(htmlPath)

	if err := RenameCatalog(jsonPath, "   "); err == nil {
		t.Fatal("expected an error for a title that trims to empty")
	}

	jsonAfter, _ := os.ReadFile(jsonPath)
	htmlAfter, _ := os.ReadFile(htmlPath)
	if string(jsonBefore) != string(jsonAfter) {
		t.Errorf("json was modified despite empty-title rejection")
	}
	if string(htmlBefore) != string(htmlAfter) {
		t.Errorf("html was modified despite empty-title rejection")
	}
}

func TestRenameCatalog_RejectsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", "{not valid json")

	before, _ := os.ReadFile(jsonPath)

	if err := RenameCatalog(jsonPath, "Photos 2024"); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}

	after, _ := os.ReadFile(jsonPath)
	if string(before) != string(after) {
		t.Errorf("invalid json file was modified despite rejection")
	}
}

// WR-03: a symlink named "<root>.html" pointing outside jsonPath's own
// directory must not be followed -- neither read from nor written to.
// Mirrors internal/osutil/reveal_test.go's "symlink pointing outside
// catalogDir" case.
func TestRenameCatalog_RejectsHTMLSymlinkEscapingCatalogDir(t *testing.T) {
	dir := t.TempDir()
	jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.html")
	if err := os.WriteFile(outsideFile, []byte("<html><title>secret</title></html>"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	htmlPath := filepath.Join(dir, "photos.html")
	if err := os.Symlink(outsideFile, htmlPath); err != nil {
		t.Skipf("symlinks unavailable in this environment: %v", err)
	}

	outsideBefore, _ := os.ReadFile(outsideFile)

	if err := RenameCatalog(jsonPath, "Photos 2024"); err == nil {
		t.Fatal("expected an error for an .html sibling that resolves outside the catalog directory")
	}

	outsideAfter, _ := os.ReadFile(outsideFile)
	if string(outsideBefore) != string(outsideAfter) {
		t.Errorf("file outside the catalog directory was modified via the .html symlink escape")
	}
}

// WR-01 (27-REVIEW.md iteration 2): a rejected .html sibling step must leave
// the JSON file's title completely untouched -- not silently renamed anyway.
// Table-driven over every way resolveContainedSibling/the html read can fail,
// since both share the same "validate before mutating" ordering fix.
func TestRenameCatalog_RejectedHTMLStepLeavesJSONTitleUnchanged(t *testing.T) {
	tests := []struct {
		name     string
		setupDir func(t *testing.T, dir string)
	}{
		{
			name: "html symlink escapes catalog directory",
			setupDir: func(t *testing.T, dir string) {
				outsideDir := t.TempDir()
				outsideFile := filepath.Join(outsideDir, "secret.html")
				if err := os.WriteFile(outsideFile, []byte("<html><title>secret</title></html>"), 0644); err != nil {
					t.Fatalf("write outside file: %v", err)
				}
				if err := os.Symlink(outsideFile, filepath.Join(dir, "photos.html")); err != nil {
					t.Skipf("symlinks unavailable in this environment: %v", err)
				}
			},
		},
		{
			name: "html file unreadable",
			setupDir: func(t *testing.T, dir string) {
				htmlPath := filepath.Join(dir, "photos.html")
				if err := os.WriteFile(htmlPath, []byte("<html></html>"), 0644); err != nil {
					t.Fatalf("write html fixture: %v", err)
				}
				if err := os.Chmod(htmlPath, 0000); err != nil {
					t.Skipf("chmod unavailable in this environment: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(htmlPath, 0644) })
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			jsonPath := writeFixture(t, dir, "photos.json", v2BareObjectFixture)
			tt.setupDir(t, dir)

			jsonBefore, err := os.ReadFile(jsonPath)
			if err != nil {
				t.Fatalf("read json before: %v", err)
			}

			if err := RenameCatalog(jsonPath, "Photos 2024"); err == nil {
				t.Fatal("expected an error when the .html sibling step is rejected")
			}

			jsonAfter, err := os.ReadFile(jsonPath)
			if err != nil {
				t.Fatalf("read json after: %v", err)
			}
			if string(jsonBefore) != string(jsonAfter) {
				t.Errorf("json title was mutated despite the rename being rejected:\nbefore: %s\nafter:  %s", jsonBefore, jsonAfter)
			}
			if strings.Contains(string(jsonAfter), "Photos 2024") {
				t.Errorf("json contains the new title despite the rename being rejected: %s", jsonAfter)
			}
		})
	}
}
