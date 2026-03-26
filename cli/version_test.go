package cli_test

import (
	"strings"
	"testing"

	"storcat-wails/cli"
)

func TestRunVersion(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		version  string
		wantCode int
		wantOut  string
	}{
		{
			name:     "prints version to stdout",
			args:     []string{"version"},
			version:  "2.0.0",
			wantCode: 0,
			wantOut:  "storcat 2.0.0\n",
		},
		{
			name:     "different version string",
			args:     []string{"version"},
			version:  "1.2.3",
			wantCode: 0,
			wantOut:  "storcat 1.2.3\n",
		},
		{
			name:     "--help returns 0",
			args:     []string{"version", "--help"},
			version:  "2.0.0",
			wantCode: 0,
		},
		{
			name:     "-h returns 0",
			args:     []string{"version", "-h"},
			version:  "2.0.0",
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := captureOutput(func() {
				code := cli.Run(tt.args, tt.version)
				if code != tt.wantCode {
					t.Errorf("exit code: want %d, got %d", tt.wantCode, code)
				}
			})
			if tt.wantOut != "" && stdout != tt.wantOut {
				t.Errorf("stdout: want %q, got %q", tt.wantOut, stdout)
			}
		})
	}
}

func TestRunVersion_HelpGoesToStderr(t *testing.T) {
	_, stderr := captureOutput(func() {
		cli.Run([]string{"version", "--help"}, "2.0.0")
	})
	// Help text should be non-empty on stderr
	if !strings.Contains(stderr, "storcat version") {
		t.Errorf("expected help text on stderr, got %q", stderr)
	}
}
