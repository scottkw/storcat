package cli_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storcat-wails/cli"
)

// captureOutput runs fn while capturing os.Stdout and os.Stderr.
// Returns (stdout, stderr).
func captureOutput(fn func()) (string, string) {
	// Save originals
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Pipe for stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// Pipe for stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	fn()

	// Close writers so readers EOF
	wOut.Close()
	wErr.Close()

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String()
}

func TestRunVersion_PrintsVersion(t *testing.T) {
	stdout, _ := captureOutput(func() {
		code := cli.Run([]string{"version"}, "2.0.0")
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})
	if stdout != "storcat 2.0.0\n" {
		t.Errorf("expected 'storcat 2.0.0\\n', got %q", stdout)
	}
}

func TestRunVersion_Help(t *testing.T) {
	_, _ = captureOutput(func() {
		code := cli.Run([]string{"version", "--help"}, "2.0.0")
		if code != 0 {
			t.Errorf("expected exit code 0 for --help, got %d", code)
		}
	})
}

func TestRunVersion_Help_H(t *testing.T) {
	_, _ = captureOutput(func() {
		code := cli.Run([]string{"version", "-h"}, "2.0.0")
		if code != 0 {
			t.Errorf("expected exit code 0 for -h, got %d", code)
		}
	})
}

func TestRunHelp_Global(t *testing.T) {
	stdout, _ := captureOutput(func() {
		code := cli.Run([]string{"--help"}, "2.0.0")
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})
	if !strings.Contains(stdout, "Commands:") {
		t.Errorf("expected stdout to contain 'Commands:', got %q", stdout)
	}
}

func TestRunHelp_HFlag(t *testing.T) {
	_, _ = captureOutput(func() {
		code := cli.Run([]string{"-h"}, "2.0.0")
		if code != 0 {
			t.Errorf("expected exit code 0 for -h, got %d", code)
		}
	})
}

func TestRunHelp_HelpCmd(t *testing.T) {
	_, _ = captureOutput(func() {
		code := cli.Run([]string{"help"}, "2.0.0")
		if code != 0 {
			t.Errorf("expected exit code 0 for help, got %d", code)
		}
	})
}

func TestRun_UnknownCommand(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"badcmd"}, "2.0.0")
		if code != 2 {
			t.Errorf("expected exit code 2 for unknown command, got %d", code)
		}
	})
	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("expected stderr to contain 'unknown command', got %q", stderr)
	}
}

func TestRun_StdoutStderrSeparation(t *testing.T) {
	// version: stdout has content, stderr is empty
	stdout, stderr := captureOutput(func() {
		cli.Run([]string{"version"}, "2.0.0")
	})
	if stdout == "" {
		t.Error("version: expected stdout to have content")
	}
	if stderr != "" {
		t.Errorf("version: expected stderr to be empty, got %q", stderr)
	}

	// unknown command: stderr has content, stdout is empty
	stdout2, stderr2 := captureOutput(func() {
		cli.Run([]string{"badcmd"}, "2.0.0")
	})
	if stdout2 != "" {
		t.Errorf("badcmd: expected stdout to be empty, got %q", stdout2)
	}
	if stderr2 == "" {
		t.Error("badcmd: expected stderr to have content")
	}
}

func TestRun_NoArgs(t *testing.T) {
	stdout, _ := captureOutput(func() {
		code := cli.Run([]string{}, "2.0.0")
		if code != 0 {
			t.Errorf("expected exit code 0 for no args, got %d", code)
		}
	})
	if !strings.Contains(stdout, "Commands:") {
		t.Errorf("expected stdout to contain 'Commands:' for no args, got %q", stdout)
	}
}

func TestRun_StubCreate(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"create"}, "2.0.0")
		if code != 1 {
			t.Errorf("expected exit code 1 for create stub, got %d", code)
		}
	})
	if !strings.Contains(stderr, "not yet implemented") {
		t.Errorf("expected 'not yet implemented' in stderr, got %q", stderr)
	}
}

func TestRun_StubSearch(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"search"}, "2.0.0")
		if code != 1 {
			t.Errorf("expected exit code 1 for search stub, got %d", code)
		}
	})
	if !strings.Contains(stderr, "not yet implemented") {
		t.Errorf("expected 'not yet implemented' in stderr, got %q", stderr)
	}
}

func TestRun_StubList(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"list"}, "2.0.0")
		if code != 1 {
			t.Errorf("expected exit code 1 for list stub, got %d", code)
		}
	})
	if !strings.Contains(stderr, "not yet implemented") {
		t.Errorf("expected 'not yet implemented' in stderr, got %q", stderr)
	}
}

func TestRun_StubShow(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"show"}, "2.0.0")
		if code != 1 {
			t.Errorf("expected exit code 1 for show stub, got %d", code)
		}
	})
	if !strings.Contains(stderr, "not yet implemented") {
		t.Errorf("expected 'not yet implemented' in stderr, got %q", stderr)
	}
}

func TestRun_StubOpen(t *testing.T) {
	_, stderr := captureOutput(func() {
		code := cli.Run([]string{"open"}, "2.0.0")
		if code != 1 {
			t.Errorf("expected exit code 1 for open stub, got %d", code)
		}
	})
	if !strings.Contains(stderr, "not yet implemented") {
		t.Errorf("expected 'not yet implemented' in stderr, got %q", stderr)
	}
}

func TestRun_StubHelp(t *testing.T) {
	// --help on stub commands should return 0
	stubs := []string{"create", "search", "list", "show", "open"}
	for _, cmd := range stubs {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			_, _ = captureOutput(func() {
				code := cli.Run([]string{cmd, "--help"}, "2.0.0")
				if code != 0 {
					t.Errorf("%s --help: expected exit code 0, got %d", cmd, code)
				}
			})
		})
	}
}

func TestRun_NoCobra(t *testing.T) {
	// Verify no cobra import in any cli/*.go file
	cliDir := filepath.Join(".") // package cli dir
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		t.Fatalf("failed to read cli dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(cliDir, e.Name()))
		if err != nil {
			t.Fatalf("failed to read %s: %v", e.Name(), err)
		}
		if strings.Contains(string(content), "github.com/spf13/cobra") {
			t.Errorf("file %s contains cobra import — must use stdlib flag.FlagSet", e.Name())
		}
		if strings.Contains(string(content), "wailsapp") {
			t.Errorf("file %s contains wails import — cli/ must have zero wails dependencies", e.Name())
		}
	}
}
