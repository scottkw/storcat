package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storcat-wails/cli"
)

func TestRunOpen_NoArg(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"open"}, "2.0.0")
		if code != 2 {
			t.Errorf("expected exit code 2 for open with no args, got %d", code)
		}
	})
	if !strings.Contains(stderr, "catalog file argument required") {
		t.Errorf("expected 'catalog file argument required' in stderr, got %q", stderr)
	}
}

func TestRunOpen_NonJsonFile(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"open", "somefile.txt"}, "2.0.0")
		if code != 2 {
			t.Errorf("expected exit code 2 for non-.json file, got %d", code)
		}
	})
	if !strings.Contains(stderr, "expected a .json catalog file") {
		t.Errorf("expected '.json catalog file' in stderr, got %q", stderr)
	}
}

func TestRunOpen_NonExistentFile(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"open", "nonexistent.json"}, "2.0.0")
		if code != 1 {
			t.Errorf("expected exit code 1 for nonexistent file, got %d", code)
		}
	})
	if stderr == "" {
		t.Errorf("expected error in stderr for nonexistent file, got empty")
	}
}

func TestRunOpen_MissingHtml(t *testing.T) {
	dir := t.TempDir()
	// Create a catalog JSON but no corresponding HTML
	writeTestCatalog(t, dir, "my-catalog")
	catalogPath := filepath.Join(dir, "my-catalog.json")

	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"open", catalogPath}, "2.0.0")
		if code != 1 {
			t.Errorf("expected exit code 1 for missing HTML, got %d", code)
		}
	})
	if !strings.Contains(stderr, "HTML file not found") {
		t.Errorf("expected 'HTML file not found' in stderr, got %q", stderr)
	}
}

func TestRunOpen_Help(t *testing.T) {
	_, _ = captureOutput(func() {
		code := cli.Run([]string{"open", "--help"}, "2.0.0")
		if code != 0 {
			t.Errorf("expected exit code 0 for --help, got %d", code)
		}
	})
}

func TestRunOpen_HtmlPathDerivation(t *testing.T) {
	// Indirectly verify HTML path derivation: JSON at /tmp/.../my-catalog.json
	// should look for /tmp/.../my-catalog.html and report it as missing.
	dir := t.TempDir()
	writeTestCatalog(t, dir, "my-catalog")
	catalogPath := filepath.Join(dir, "my-catalog.json")
	expectedHtmlPath := filepath.Join(dir, "my-catalog.html")

	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"open", catalogPath}, "2.0.0")
		if code != 1 {
			t.Errorf("expected exit code 1 for missing HTML, got %d", code)
		}
	})
	if !strings.Contains(stderr, expectedHtmlPath) {
		// Also accept just the basename as the error may use abs or rel
		basename := "my-catalog.html"
		if !strings.Contains(stderr, basename) {
			t.Errorf("expected HTML path %q in stderr, got %q", expectedHtmlPath, stderr)
		}
	}
	// Ensure no HTML file was created
	if _, err := os.Stat(expectedHtmlPath); err == nil {
		t.Error("expected HTML file to not exist, but it does")
	}
}
