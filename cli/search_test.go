package cli_test

import (
	"encoding/json"
	"strings"
	"testing"

	"storcat-wails/cli"
	"storcat-wails/pkg/models"
)

func TestRunSearch_MissingArgs(t *testing.T) {
	exitCode := 0
	_, stderr := captureOutput(func() {
		exitCode = cli.Run([]string{"search"}, "2.0.0")
	})
	if exitCode != 2 {
		t.Errorf("expected exit code 2 for search with no args, got %d", exitCode)
	}
	if !strings.Contains(stderr, "search term required") {
		t.Errorf("expected 'search term required' in stderr, got %q", stderr)
	}
}

func TestRunSearch_NoResults(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "mycat")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"search", "zzznomatchzzzxxx", dir}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for no results, got %d", exitCode)
	}
	if !strings.Contains(stdout, "No results found.") {
		t.Errorf("expected 'No results found.' in stdout, got %q", stdout)
	}
}

func TestRunSearch_WithResults(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "mycat")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"search", "test", dir}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for search with results, got %d", exitCode)
	}
	if !strings.Contains(stdout, "test.txt") {
		t.Errorf("expected 'test.txt' in stdout, got %q", stdout)
	}
}

func TestRunSearch_JSON(t *testing.T) {
	dir := t.TempDir()
	writeTestCatalog(t, dir, "mycat")

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"search", "test", dir, "--json"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for search --json, got %d", exitCode)
	}

	var results []*models.SearchResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Errorf("expected valid JSON array in stdout, got parse error: %v\nstdout: %q", err, stdout)
		return
	}
	if len(results) == 0 {
		t.Error("expected at least one search result, got empty array")
	}
}

func TestRunSearch_NonExistentDir(t *testing.T) {
	exitCode := 0
	_, stderr := captureOutput(func() {
		exitCode = cli.Run([]string{"search", "term", "/nonexistent/path/that/does/not/exist"}, "2.0.0")
	})
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for nonexistent dir, got %d", exitCode)
	}
	if stderr == "" {
		t.Error("expected error message on stderr for nonexistent dir, got empty string")
	}
}

func TestRunSearch_Help(t *testing.T) {
	exitCode := 0
	captureOutput(func() {
		exitCode = cli.Run([]string{"search", "--help"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for search --help, got %d", exitCode)
	}
}
