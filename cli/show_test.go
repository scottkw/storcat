package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storcat-wails/cli"
	"storcat-wails/pkg/models"
)

// writeDeepTestCatalog creates a catalog with nested dirs for depth testing.
func writeDeepTestCatalog(t *testing.T, dir, name string) string {
	t.Helper()
	catalog := models.CatalogItem{
		Type: "directory", Name: "./", Size: 2048,
		Contents: []*models.CatalogItem{
			{Type: "directory", Name: "./subdir", Size: 1024,
				Contents: []*models.CatalogItem{
					{Type: "file", Name: "./subdir/deep.txt", Size: 100},
				},
			},
			{Type: "file", Name: "./root.txt", Size: 512},
		},
	}
	data, _ := json.MarshalIndent(catalog, "", "  ")
	path := filepath.Join(dir, name+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writeDeepTestCatalog: %v", err)
	}
	return path
}

func TestRunShow_TreeOutput(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "my-catalog")
	catalogPath := filepath.Join(dir, "my-catalog.json")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"show", catalogPath}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	// Check for tree connectors and file names
	if !strings.Contains(stdout, "──") {
		t.Errorf("expected tree connectors in stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "test.txt") {
		t.Errorf("expected 'test.txt' in stdout, got %q", stdout)
	}
}

func TestRunShow_RootNodePrinted(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "my-catalog")
	catalogPath := filepath.Join(dir, "my-catalog.json")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"show", catalogPath}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	// First line should contain the root directory name (filepath.Base of "./")
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) == 0 {
		t.Fatal("expected output lines, got none")
	}
	// Root name is "." (filepath.Base of "./")
	if !strings.Contains(lines[0], ".") {
		t.Errorf("expected root directory name in first line, got %q", lines[0])
	}
}

func TestRunShow_JSON(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "my-catalog")
	catalogPath := filepath.Join(dir, "my-catalog.json")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"show", catalogPath, "--json"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	// Should be valid JSON
	var result models.CatalogItem
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("expected valid JSON in stdout, got parse error: %v\nstdout: %q", err, stdout)
	}
	if result.Type != "directory" {
		t.Errorf("expected root type 'directory', got %q", result.Type)
	}
}

func TestRunShow_Depth0(t *testing.T) {
	dir := t.TempDir()
	writeDeepTestCatalog(t, dir, "deep-catalog")
	catalogPath := filepath.Join(dir, "deep-catalog.json")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"show", catalogPath, "--depth", "0"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	// Root only, no children in output
	if strings.Contains(stdout, "subdir") {
		t.Errorf("depth 0: expected no children in stdout, but found 'subdir': %q", stdout)
	}
	if strings.Contains(stdout, "root.txt") {
		t.Errorf("depth 0: expected no children in stdout, but found 'root.txt': %q", stdout)
	}
}

func TestRunShow_Depth1(t *testing.T) {
	dir := t.TempDir()
	writeDeepTestCatalog(t, dir, "deep-catalog")
	catalogPath := filepath.Join(dir, "deep-catalog.json")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"show", catalogPath, "--depth", "1"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	// Immediate children visible, grandchildren not
	if !strings.Contains(stdout, "subdir") {
		t.Errorf("depth 1: expected 'subdir' in stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "root.txt") {
		t.Errorf("depth 1: expected 'root.txt' in stdout, got %q", stdout)
	}
	// Grandchild should not be visible
	if strings.Contains(stdout, "deep.txt") {
		t.Errorf("depth 1: expected 'deep.txt' to be hidden, but found it: %q", stdout)
	}
}

func TestRunShow_NoArg(t *testing.T) {
	exitCode := 0
	_, stderr := captureOutput(func() {
		exitCode = cli.Run([]string{"show"}, "2.0.0")
	})
	if exitCode != 2 {
		t.Errorf("expected exit code 2 for no args, got %d", exitCode)
	}
	if !strings.Contains(stderr, "catalog file argument required") {
		t.Errorf("expected 'catalog file argument required' in stderr, got %q", stderr)
	}
}

func TestRunShow_NonJsonFile(t *testing.T) {
	exitCode := 0
	_, stderr := captureOutput(func() {
		exitCode = cli.Run([]string{"show", "somefile.txt"}, "2.0.0")
	})
	if exitCode != 2 {
		t.Errorf("expected exit code 2 for non-.json file, got %d", exitCode)
	}
	if !strings.Contains(stderr, "expected a .json catalog file") {
		t.Errorf("expected 'expected a .json catalog file' in stderr, got %q", stderr)
	}
}

func TestRunShow_NonExistentFile(t *testing.T) {
	exitCode := 0
	_, stderr := captureOutput(func() {
		exitCode = cli.Run([]string{"show", "/nonexistent/path/catalog.json"}, "2.0.0")
	})
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for nonexistent file, got %d", exitCode)
	}
	if stderr == "" {
		t.Error("expected error message on stderr for nonexistent file")
	}
}

func TestRunShow_Help(t *testing.T) {
	exitCode := 0
	captureOutput(func() {
		exitCode = cli.Run([]string{"show", "--help"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for --help, got %d", exitCode)
	}
}

func TestRunShow_NoColor(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "my-catalog")
	catalogPath := filepath.Join(dir, "my-catalog.json")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"show", catalogPath, "--no-color"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	// No ANSI escape sequences
	if strings.Contains(stdout, "\x1b") {
		t.Errorf("expected no ANSI escape sequences with --no-color, got %q", stdout)
	}
}

func TestRunShow_DepthIgnoredForJSON(t *testing.T) {
	dir := t.TempDir()
	writeDeepTestCatalog(t, dir, "deep-catalog")
	catalogPath := filepath.Join(dir, "deep-catalog.json")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"show", catalogPath, "--json", "--depth", "1"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	// Full JSON output — depth flag ignored for --json
	var result models.CatalogItem
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("expected valid JSON in stdout, got parse error: %v\nstdout: %q", err, stdout)
	}
	// Should have nested content (grandchildren visible)
	if len(result.Contents) == 0 {
		t.Error("expected root contents in full JSON output")
	}
	// Check that the subdir has its contents (depth not applied to JSON)
	var subdirFound bool
	for _, child := range result.Contents {
		if child.Type == "directory" {
			subdirFound = true
			if len(child.Contents) == 0 {
				t.Errorf("expected subdir to have contents in full JSON (depth ignored), got empty contents")
			}
		}
	}
	if !subdirFound {
		t.Errorf("expected a directory child in JSON output, got: %+v", result.Contents)
	}
}
