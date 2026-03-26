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

func TestRunCreate_MissingArg(t *testing.T) {
	exitCode := 0
	_, stderr := captureOutput(func() {
		exitCode = cli.Run([]string{"create"}, "2.0.0")
	})
	if exitCode != 2 {
		t.Errorf("expected exit code 2 for create with no args, got %d", exitCode)
	}
	if !strings.Contains(stderr, "directory argument required") {
		t.Errorf("expected 'directory argument required' in stderr, got %q", stderr)
	}
}

func TestRunCreate_NonExistentDir(t *testing.T) {
	exitCode := 0
	_, stderr := captureOutput(func() {
		exitCode = cli.Run([]string{"create", "/nonexistent/path/that/does/not/exist"}, "2.0.0")
	})
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for nonexistent dir, got %d", exitCode)
	}
	if stderr == "" {
		t.Error("expected error message on stderr for nonexistent dir, got empty string")
	}
}

func TestRunCreate_Success(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"create", dir}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for create success, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Created catalog") {
		t.Errorf("expected 'Created catalog' in stdout, got %q", stdout)
	}

	base := filepath.Base(dir)
	jsonFile := filepath.Join(dir, base+".json")
	htmlFile := filepath.Join(dir, base+".html")
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		t.Errorf("expected JSON file to exist at %s", jsonFile)
	}
	if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
		t.Errorf("expected HTML file to exist at %s", htmlFile)
	}
}

func TestRunCreate_WithFlags(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"create", dir, "--title", "Test Title", "--name", "mycat"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for create with flags, got %d", exitCode)
	}
	_ = stdout

	jsonFile := filepath.Join(dir, "mycat.json")
	htmlFile := filepath.Join(dir, "mycat.html")
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		t.Errorf("expected mycat.json to exist at %s", jsonFile)
	}
	if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
		t.Errorf("expected mycat.html to exist at %s", htmlFile)
	}
}

func TestRunCreate_WithOutput(t *testing.T) {
	dir := t.TempDir()
	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"create", dir, "--output", outDir}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for create with --output, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Copied to:") {
		t.Errorf("expected 'Copied to:' in stdout, got %q", stdout)
	}

	base := filepath.Base(dir)
	// Files should exist in source dir
	if _, err := os.Stat(filepath.Join(dir, base+".json")); os.IsNotExist(err) {
		t.Errorf("expected JSON in source dir")
	}
	// Files should also exist in output dir
	if _, err := os.Stat(filepath.Join(outDir, base+".json")); os.IsNotExist(err) {
		t.Errorf("expected JSON in output dir")
	}
	if _, err := os.Stat(filepath.Join(outDir, base+".html")); os.IsNotExist(err) {
		t.Errorf("expected HTML in output dir")
	}
}

func TestRunCreate_JSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	exitCode := 0
	stdout, _ := captureOutput(func() {
		exitCode = cli.Run([]string{"create", dir, "--json"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for create --json, got %d", exitCode)
	}

	var result models.CreateCatalogResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("expected valid JSON object in stdout, got parse error: %v\nstdout: %q", err, stdout)
		return
	}
	if result.JsonPath == "" {
		t.Error("expected non-empty JsonPath in result")
	}
	if result.FileCount == 0 {
		t.Error("expected FileCount > 0 in result")
	}
}

func TestRunCreate_Help(t *testing.T) {
	exitCode := 0
	captureOutput(func() {
		exitCode = cli.Run([]string{"create", "--help"}, "2.0.0")
	})
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for create --help, got %d", exitCode)
	}
}
