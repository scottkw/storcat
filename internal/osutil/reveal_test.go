package osutil

import (
	"os"
	"path/filepath"
	"testing"
)

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRevealArgv_Darwin(t *testing.T) {
	name, args := revealArgvDarwin("/tmp/catalogs/catalog.json")
	if name != "open" {
		t.Errorf("name = %q, want %q", name, "open")
	}
	want := []string{"-R", "/tmp/catalogs/catalog.json"}
	if !equalSlices(args, want) {
		t.Errorf("args = %#v, want %#v", args, want)
	}
}

func TestRevealArgv_Windows(t *testing.T) {
	name, args := revealArgvWindows(`C:\catalogs\catalog.json`)
	if name != "explorer" {
		t.Errorf("name = %q, want %q", name, "explorer")
	}
	want := []string{`/select,C:\catalogs\catalog.json`}
	if !equalSlices(args, want) {
		t.Errorf("args = %#v, want %#v", args, want)
	}
}

func TestRevealArgv_Linux(t *testing.T) {
	name, args := revealArgvLinux("/home/user/catalogs/catalog.json")
	if name != "xdg-open" {
		t.Errorf("name = %q, want %q", name, "xdg-open")
	}
	want := []string{"/home/user/catalogs"}
	if !equalSlices(args, want) {
		t.Errorf("args = %#v, want %#v", args, want)
	}
}

func TestRevealArgvFor_UnrecognisedPlatform(t *testing.T) {
	name, args := revealArgvFor("plan9", "/tmp/catalog.json")
	if name != "" {
		t.Errorf("name = %q, want empty for an unrecognised platform", name)
	}
	if args != nil {
		t.Errorf("args = %#v, want nil for an unrecognised platform", args)
	}
}

func TestRevealArgvFor_DispatchesToEachPlatform(t *testing.T) {
	darwinName, _ := revealArgvFor("darwin", "/tmp/catalog.json")
	if darwinName != "open" {
		t.Errorf("darwin: name = %q, want %q", darwinName, "open")
	}
	windowsName, _ := revealArgvFor("windows", `C:\catalog.json`)
	if windowsName != "explorer" {
		t.Errorf("windows: name = %q, want %q", windowsName, "explorer")
	}
	linuxName, _ := revealArgvFor("linux", "/tmp/catalog.json")
	if linuxName != "xdg-open" {
		t.Errorf("linux: name = %q, want %q", linuxName, "xdg-open")
	}
}

// hostilePath carries a space, a single quote, a double quote, a semicolon,
// a double ampersand, a backtick, a dollar-parenthesis sequence, a pipe and
// a newline -- every metacharacter that could start a second command if
// this path were ever handed to a shell instead of built as its own
// argument-vector element. Each platform assertion below checks exact
// element count and exact element content, not a substring match, so an
// accidental split or an accidental join both fail the test.
const hostilePath = "/tmp/a b'c\"d;e&&f`g$(h)i|j\nk.json"

func TestRevealArgv_HostilePath_Darwin(t *testing.T) {
	_, args := revealArgvDarwin(hostilePath)
	if len(args) != 2 {
		t.Fatalf("len(args) = %d, want 2", len(args))
	}
	if args[0] != "-R" {
		t.Errorf("args[0] = %q, want %q", args[0], "-R")
	}
	if args[1] != hostilePath {
		t.Errorf("args[1] = %q, want the path byte-identical to the input", args[1])
	}
}

func TestRevealArgv_HostilePath_Windows(t *testing.T) {
	_, args := revealArgvWindows(hostilePath)
	if len(args) != 1 {
		t.Fatalf("len(args) = %d, want 1", len(args))
	}
	want := "/select," + hostilePath
	if args[0] != want {
		t.Errorf("args[0] = %q, want %q", args[0], want)
	}
}

func TestRevealArgv_HostilePath_Linux(t *testing.T) {
	_, args := revealArgvLinux(hostilePath)
	if len(args) != 1 {
		t.Fatalf("len(args) = %d, want 1", len(args))
	}
	want := filepath.Dir(hostilePath)
	if args[0] != want {
		t.Errorf("args[0] = %q, want %q", args[0], want)
	}
}

func TestRevealInFileManager_RejectsMissingPath(t *testing.T) {
	err := RevealInFileManager(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatal("expected an error for a missing path, got nil")
	}
}

func TestRevealInFileManager_RejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	err := RevealInFileManager(dir)
	if err == nil {
		t.Fatal("expected an error for a directory, got nil")
	}
}

func TestRevealInFileManager_RejectsDisallowedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-catalog.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	err := RevealInFileManager(path)
	if err == nil {
		t.Fatal("expected an error for a disallowed extension, got nil")
	}
}

// A relative path must be resolved to an absolute one before any check is
// made. This is proven without ever reaching exec.Command (which would pop
// a real Finder/Explorer/file-manager window during `go test`): a relative
// path resolves against the current working directory, so a disallowed
// extension found via that relative form -- rather than a "no such file"
// error -- is proof the path was resolved and actually stat-checked at its
// real, cwd-relative location.
func TestRevealInFileManager_ResolvesRelativePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-catalog.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	err = RevealInFileManager("not-a-catalog.txt")
	if err == nil {
		t.Fatal("expected an error for a disallowed extension via a relative path, got nil")
	}
}
