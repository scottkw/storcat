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

// writeTestCatalog writes a minimal valid JSON catalog to dir/name.json.
func writeTestCatalog(t *testing.T, dir, name string) {
	t.Helper()
	catalog := models.CatalogItem{
		Type: "directory",
		Name: "./",
		Size: 1024,
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./test.txt", Size: 100},
		},
	}
	data, _ := json.MarshalIndent(catalog, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, name+".json"), data, 0644); err != nil {
		t.Fatalf("writeTestCatalog: %v", err)
	}
}

func TestRunList_MissingDir_DefaultsCwd(t *testing.T) {
	// storcat list with no dir arg defaults to cwd; cwd may have no .json catalogs
	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"list"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for list with no dir, got %d", exitCode)
	}
	// Outcome depends on cwd contents; we only assert exit code 0
	_ = stdout
}

func TestRunList_NonExistentDir(t *testing.T) {
	exitCode := 0
	_, stderr := captureOutput(func() {
		exitCode = cli.Run([]string{"list", "/nonexistent/path/that/does/not/exist"}, "2.0.0")
	})
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for nonexistent dir, got %d", exitCode)
	}
	if stderr == "" {
		t.Error("expected error message on stderr for nonexistent dir, got empty string")
	}
}

func TestRunList_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"list", dir}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for empty dir, got %d", exitCode)
	}
	if !strings.Contains(stdout, "No catalogs found.") {
		t.Errorf("expected 'No catalogs found.' in stdout, got %q", stdout)
	}
}

func TestRunList_WithCatalogs(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "my-catalog")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"list", dir}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for dir with catalog, got %d", exitCode)
	}
	if !strings.Contains(stdout, "my-catalog") {
		t.Errorf("expected catalog name 'my-catalog' in stdout, got %q", stdout)
	}
}

func TestRunList_JSON(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "test-catalog")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"list", dir, "--json"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for --json, got %d", exitCode)
	}
	if !strings.Contains(stdout, "[") {
		t.Errorf("expected JSON array in stdout, got %q", stdout)
	}

	// Parse the JSON output
	var results []models.CatalogMetadata
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Errorf("expected valid JSON array in stdout, got parse error: %v\nstdout: %q", err, stdout)
		return
	}
	if len(results) == 0 {
		t.Error("expected at least one catalog in JSON output, got empty array")
	}
}

func TestRunList_JSON_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"list", dir, "--json"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for empty dir --json, got %d", exitCode)
	}

	// Should output empty JSON array
	var results []models.CatalogMetadata
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Errorf("expected valid JSON array for empty dir, got parse error: %v\nstdout: %q", err, stdout)
		return
	}
	if len(results) != 0 {
		t.Errorf("expected empty JSON array for empty dir, got %d items", len(results))
	}
}

func TestRunList_Help(t *testing.T) {
	exitCode := 0
	captureOutput(func() {
		exitCode = cli.Run([]string{"list", "--help"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for list --help, got %d", exitCode)
	}
}
