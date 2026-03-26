package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterMacOSArgs(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "RemovesPsn",
			input: []string{"-psn_0_12345", "version"},
			want:  []string{"version"},
		},
		{
			name:  "PreservesNormalArgs",
			input: []string{"version", "--help"},
			want:  []string{"version", "--help"},
		},
		{
			name:  "EmptyArgs",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "OnlyPsn",
			input: []string{"-psn_0_99999"},
			want:  []string{},
		},
		{
			name:  "MultiplePsn",
			input: []string{"-psn_0_1", "-psn_0_2", "version"},
			want:  []string{"version"},
		},
		{
			name:  "PsnLikeButNot",
			input: []string{"--psn_flag"},
			want:  []string{"--psn_flag"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterMacOSArgs(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("filterMacOSArgs(%v) = %v (len %d), want %v (len %d)",
					tc.input, got, len(got), tc.want, len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("filterMacOSArgs(%v)[%d] = %q, want %q",
						tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestDispatch_UnrecognizedArgsFallThrough verifies CLPC-03: wails dev flags and
// other unrecognized args are NOT in the main.go dispatch switch case list, so they
// fall through to runGUI() instead of triggering os.Exit. This tests the observable
// contract that the switch has no default case matching unknown args.
func TestDispatch_UnrecognizedArgsFallThrough(t *testing.T) {
	// Read main.go to inspect the dispatch switch contents.
	// We verify that wails dev hot-reload flags are absent from the case list.
	mainGo, err := os.ReadFile(filepath.Join(".", "main.go"))
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	content := string(mainGo)

	// Verify the known subcommands ARE present (sanity check the switch exists).
	knownCmds := []string{"version", "create", "search", "list", "show", "open", "help", "--help", "-h"}
	for _, cmd := range knownCmds {
		if !strings.Contains(content, `"`+cmd+`"`) {
			t.Errorf("main.go dispatch switch is missing expected case %q", cmd)
		}
	}

	// Verify there is NO default case that calls os.Exit — unrecognized args must
	// fall through to runGUI(), not trigger an exit.
	// A default case with os.Exit would look like `default:` followed by `os.Exit`.
	// We check that the switch block does not contain a default branch with exit logic.
	switchBlock := extractDispatchSwitch(content)
	if strings.Contains(switchBlock, "default:") {
		// A default case exists — verify it does NOT contain os.Exit
		defaultIdx := strings.Index(switchBlock, "default:")
		afterDefault := switchBlock[defaultIdx:]
		if strings.Contains(afterDefault, "os.Exit") {
			t.Error("main.go dispatch switch has a default: case that calls os.Exit — unrecognized args must fall through to runGUI(), not exit")
		}
	}

	// Verify wails dev flags are not listed as known cases.
	// These flags are injected by `wails dev` for hot-reload; they must not be matched.
	wailsDevFlags := []string{
		`"-appargs"`,
		`"--wails-dev-mode"`,
		`"-wails-dev-mode"`,
		`"--loglevel"`,
		`"-reload"`,
	}
	for _, flag := range wailsDevFlags {
		if strings.Contains(switchBlock, flag) {
			t.Errorf("main.go dispatch switch unexpectedly contains wails dev flag %s — it should fall through to runGUI()", flag)
		}
	}

	// Verify the fall-through path: runGUI() is called after the switch (not inside it).
	afterSwitch := content[strings.Index(content, "// GUI mode"):]
	if !strings.Contains(afterSwitch, "runGUI()") {
		t.Error("main.go does not call runGUI() after the dispatch switch — fall-through path is missing")
	}
}

// extractDispatchSwitch extracts the text of the dispatch switch block from main.go content.
// Returns from "switch args[0]" to the closing brace of that block.
func extractDispatchSwitch(content string) string {
	start := strings.Index(content, "switch args[0]")
	if start == -1 {
		return ""
	}
	// Find the matching closing brace by counting braces
	depth := 0
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[start : i+1]
			}
		}
	}
	return content[start:]
}

// TestInstallCLIScript_Content verifies CLPC-04: scripts/install-cli.sh has the
// correct content to create /usr/local/bin/storcat symlink. We validate the script's
// structure and key content without actually running it (which would require sudo
// and /Applications/StorCat.app to be installed).
func TestInstallCLIScript_Content(t *testing.T) {
	scriptPath := filepath.Join(".", "scripts", "install-cli.sh")

	// Verify the file exists and is executable.
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("scripts/install-cli.sh not found: %v", err)
	}
	mode := info.Mode()
	if mode&0111 == 0 {
		t.Error("scripts/install-cli.sh is not executable — run: chmod +x scripts/install-cli.sh")
	}

	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read scripts/install-cli.sh: %v", err)
	}
	text := string(content)

	tests := []struct {
		name    string
		pattern string
		desc    string
	}{
		{
			name:    "shebang",
			pattern: "#!/bin/sh",
			desc:    "script must start with POSIX shebang #!/bin/sh",
		},
		{
			name:    "set -e",
			pattern: "set -e",
			desc:    "script must use set -e for error propagation",
		},
		{
			name:    "app binary path",
			pattern: `APP_BINARY="/Applications/StorCat.app/Contents/MacOS/StorCat"`,
			desc:    "script must reference the correct .app bundle binary path",
		},
		{
			name:    "link path",
			pattern: `LINK_PATH="/usr/local/bin/storcat"`,
			desc:    "script must set symlink destination to /usr/local/bin/storcat",
		},
		{
			name:    "ln -sf command",
			pattern: `ln -sf "$APP_BINARY" "$LINK_PATH"`,
			desc:    "script must use ln -sf to create/overwrite the symlink",
		},
		{
			name:    "missing app check",
			pattern: `[ ! -f "$APP_BINARY" ]`,
			desc:    "script must check if .app binary exists before symlinking",
		},
		{
			name:    "error exit on missing app",
			pattern: "exit 1",
			desc:    "script must exit 1 when .app binary is not found",
		},
		{
			name:    "error to stderr",
			pattern: ">&2",
			desc:    "script must write error messages to stderr",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(text, tc.pattern) {
				t.Errorf("scripts/install-cli.sh missing %q — %s", tc.pattern, tc.desc)
			}
		})
	}
}
