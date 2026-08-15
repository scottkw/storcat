package osutil

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveContainedFileURL mirrors TestContainsPath's temp-dir +
// EvalSymlinks canonicalisation fixture (reveal_test.go), covering every
// case FU-23-A's containment gate for App.OpenExternal must handle.
func TestResolveContainedFileURL(t *testing.T) {
	base := t.TempDir()
	catalogDir := filepath.Join(base, "catalogs")
	if err := os.Mkdir(catalogDir, 0755); err != nil {
		t.Fatalf("mkdir catalogDir: %v", err)
	}

	// resolvedCatalogDir mirrors what ContainsPath computes internally.
	// macOS's t.TempDir() lives under a symlinked /var -> /private/var, so
	// expected values must be built from this canonical form, exactly as
	// ResolveContainedFileURL itself does by EvalSymlinks-ing the target
	// file before calling ContainsPath.
	resolvedCatalogDir, err := filepath.EvalSymlinks(catalogDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(catalogDir): %v", err)
	}

	htmlPath := filepath.Join(resolvedCatalogDir, "catalog.html")
	if err := os.WriteFile(htmlPath, []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("write html file: %v", err)
	}
	jsonPath := filepath.Join(resolvedCatalogDir, "catalog.json")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("write json file: %v", err)
	}
	txtPath := filepath.Join(resolvedCatalogDir, "notes.txt")
	if err := os.WriteFile(txtPath, []byte("hi"), 0644); err != nil {
		t.Fatalf("write txt file: %v", err)
	}
	spacePath := filepath.Join(resolvedCatalogDir, "with space.html")
	if err := os.WriteFile(spacePath, []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("write space-named file: %v", err)
	}
	subDir := filepath.Join(resolvedCatalogDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	wantHTMLURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(htmlPath)}).String()
	wantJSONURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(jsonPath)}).String()

	t.Run("bare absolute path inside catalogDir accepted", func(t *testing.T) {
		got, err := ResolveContainedFileURL(htmlPath, catalogDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != wantHTMLURL {
			t.Errorf("got %q, want %q", got, wantHTMLURL)
		}
	})

	t.Run("file:// URL for the same file accepted, same return value", func(t *testing.T) {
		fileURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(htmlPath)}).String()
		got, err := ResolveContainedFileURL(fileURL, catalogDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != wantHTMLURL {
			t.Errorf("got %q, want %q", got, wantHTMLURL)
		}
	})

	t.Run("file:// URL with percent-encoded space accepted", func(t *testing.T) {
		fileURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(spacePath)}).String()
		if !strings.Contains(fileURL, "%20") {
			t.Fatalf("test fixture bug: expected %%20 in constructed URL, got %q", fileURL)
		}
		want := (&url.URL{Scheme: "file", Path: filepath.ToSlash(spacePath)}).String()
		got, err := ResolveContainedFileURL(fileURL, catalogDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("real .json file accepted", func(t *testing.T) {
		got, err := ResolveContainedFileURL(jsonPath, catalogDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != wantJSONURL {
			t.Errorf("got %q, want %q", got, wantJSONURL)
		}
	})

	t.Run("empty catalogDir rejected", func(t *testing.T) {
		got, err := ResolveContainedFileURL(htmlPath, "")
		if err == nil {
			t.Fatal("expected an error for empty catalogDir, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string on rejection, got %q", got)
		}
	})

	t.Run("sibling directory sharing a name prefix rejected", func(t *testing.T) {
		evilDir := resolvedCatalogDir + "-evil"
		if err := os.Mkdir(evilDir, 0755); err != nil {
			t.Fatalf("mkdir evilDir: %v", err)
		}
		evilFile := filepath.Join(evilDir, "catalog.html")
		if err := os.WriteFile(evilFile, []byte("<html></html>"), 0644); err != nil {
			t.Fatalf("write evil file: %v", err)
		}
		got, err := ResolveContainedFileURL(evilFile, catalogDir)
		if err == nil {
			t.Fatal("expected an error for a name-prefix-sharing sibling directory, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string on rejection, got %q", got)
		}
	})

	t.Run("dot-dot escape rejected", func(t *testing.T) {
		escaped := filepath.Clean(filepath.Join(resolvedCatalogDir, "..", "outside.html"))
		if err := os.WriteFile(escaped, []byte("<html></html>"), 0644); err != nil {
			t.Fatalf("write outside file: %v", err)
		}
		got, err := ResolveContainedFileURL(escaped, catalogDir)
		if err == nil {
			t.Fatal("expected an error for a \"../\" escape, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string on rejection, got %q", got)
		}
	})

	t.Run("symlink pointing outside catalogDir rejected", func(t *testing.T) {
		outsideDir := filepath.Join(base, "outside-target")
		if err := os.Mkdir(outsideDir, 0755); err != nil {
			t.Fatalf("mkdir outsideDir: %v", err)
		}
		realFile := filepath.Join(outsideDir, "real.html")
		if err := os.WriteFile(realFile, []byte("<html></html>"), 0644); err != nil {
			t.Fatalf("write real file: %v", err)
		}
		link := filepath.Join(catalogDir, "link.html")
		if err := os.Symlink(realFile, link); err != nil {
			t.Skipf("symlinks unavailable in this environment: %v", err)
		}
		got, err := ResolveContainedFileURL(link, catalogDir)
		if err == nil {
			t.Fatal("expected an error for a symlink resolving outside catalogDir, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string on rejection, got %q", got)
		}
	})

	t.Run("directory inside catalogDir rejected", func(t *testing.T) {
		got, err := ResolveContainedFileURL(subDir, catalogDir)
		if err == nil {
			t.Fatal("expected an error for a directory, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string on rejection, got %q", got)
		}
	})

	t.Run(".txt file inside catalogDir rejected", func(t *testing.T) {
		got, err := ResolveContainedFileURL(txtPath, catalogDir)
		if err == nil {
			t.Fatal("expected an error for a .txt file, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string on rejection, got %q", got)
		}
	})

	t.Run("nonexistent path inside catalogDir rejected", func(t *testing.T) {
		missing := filepath.Join(resolvedCatalogDir, "does-not-exist.html")
		got, err := ResolveContainedFileURL(missing, catalogDir)
		if err == nil {
			t.Fatal("expected an error for a nonexistent path, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string on rejection, got %q", got)
		}
	})

	t.Run("relative path rejected", func(t *testing.T) {
		got, err := ResolveContainedFileURL("catalog.html", catalogDir)
		if err == nil {
			t.Fatal("expected an error for a relative path, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string on rejection, got %q", got)
		}
	})

	rejectedSchemes := []string{
		"http://example.com/x.html",
		"https://example.com/x.html",
		"javascript:alert(1)",
		"data:text/html,<b>x</b>",
	}
	for _, raw := range rejectedSchemes {
		raw := raw
		t.Run("rejected scheme: "+raw, func(t *testing.T) {
			got, err := ResolveContainedFileURL(raw, catalogDir)
			if err == nil {
				t.Fatalf("expected an error for %q, got nil", raw)
			}
			if got != "" {
				t.Errorf("expected empty string on rejection, got %q", got)
			}
			if !strings.Contains(err.Error(), raw) {
				t.Errorf("error %q does not mention rejected input %q", err.Error(), raw)
			}
		})
	}
}
