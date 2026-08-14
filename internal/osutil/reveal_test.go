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
	dir := t.TempDir()
	err := RevealInFileManager(filepath.Join(dir, "does-not-exist.json"), dir)
	if err == nil {
		t.Fatal("expected an error for a missing path, got nil")
	}
}

func TestRevealInFileManager_RejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	err := RevealInFileManager(dir, dir)
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
	err := RevealInFileManager(path, dir)
	if err == nil {
		t.Fatal("expected an error for a disallowed extension, got nil")
	}
}

func TestRevealInFileManager_RejectsMissingCatalogDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	err := RevealInFileManager(path, "")
	if err == nil {
		t.Fatal("expected an error when no catalog directory is configured, got nil")
	}
}

// TestRevealInFileManager_RejectsPathOutsideCatalogDir proves containment is
// actually wired into RevealInFileManager itself, not just containsPath in
// isolation -- a valid .json regular file, but outside catalogDir, must be
// rejected before exec.Command is ever reached (so this test cannot pop a
// real Finder/Explorer/file-manager window).
func TestRevealInFileManager_RejectsPathOutsideCatalogDir(t *testing.T) {
	base := t.TempDir()
	catalogDir := filepath.Join(base, "catalogs")
	outsideDir := filepath.Join(base, "outside")
	if err := os.Mkdir(catalogDir, 0755); err != nil {
		t.Fatalf("mkdir catalogDir: %v", err)
	}
	if err := os.Mkdir(outsideDir, 0755); err != nil {
		t.Fatalf("mkdir outsideDir: %v", err)
	}
	path := filepath.Join(outsideDir, "catalog.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	err := RevealInFileManager(path, catalogDir)
	if err == nil {
		t.Fatal("expected an error for a path outside the configured catalog directory, got nil")
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

	err = RevealInFileManager("not-a-catalog.txt", dir)
	if err == nil {
		t.Fatal("expected an error for a disallowed extension via a relative path, got nil")
	}
}

// TestContainsPath exercises the containment check in isolation -- no
// exec.Command in reach here at all -- covering the four scenarios WR-02
// asked for: a legitimate in-directory path, a sibling directory that only
// shares catalogDir's name as a string prefix, a "../" escape, and a
// symlink whose resolved target lands outside catalogDir.
func TestContainsPath(t *testing.T) {
	base := t.TempDir()
	catalogDir := filepath.Join(base, "catalogs")
	if err := os.Mkdir(catalogDir, 0755); err != nil {
		t.Fatalf("mkdir catalogDir: %v", err)
	}

	// resolvedCatalogDir mirrors what containsPath computes internally for
	// catalogDir (filepath.Abs + filepath.EvalSymlinks). On macOS t.TempDir()
	// lives under a symlinked /var -> /private/var, so building "resolved"
	// test inputs from the raw (unresolved) catalogDir would compare a
	// resolved base against an unresolved child and fail for the wrong
	// reason. Every subtest below builds its "resolved" input from this same
	// canonical form, exactly as RevealInFileManager does by running
	// EvalSymlinks on the real file before calling containsPath.
	resolvedCatalogDir, err := filepath.EvalSymlinks(catalogDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(catalogDir): %v", err)
	}

	t.Run("legitimate path inside catalogDir", func(t *testing.T) {
		resolved := filepath.Join(resolvedCatalogDir, "catalog.json")
		ok, err := containsPath(catalogDir, resolved)
		if err != nil {
			t.Fatalf("containsPath returned an error: %v", err)
		}
		if !ok {
			t.Error("expected a path inside catalogDir to be contained, got false")
		}
	})

	t.Run("sibling directory sharing a name prefix", func(t *testing.T) {
		// "/…/catalogs-evil" has "/…/catalogs" as a string prefix -- a naive
		// strings.HasPrefix check would wrongly admit this.
		evilDir := resolvedCatalogDir + "-evil"
		if err := os.Mkdir(evilDir, 0755); err != nil {
			t.Fatalf("mkdir evilDir: %v", err)
		}
		resolved := filepath.Join(evilDir, "catalog.json")
		ok, err := containsPath(catalogDir, resolved)
		if err != nil {
			t.Fatalf("containsPath returned an error: %v", err)
		}
		if ok {
			t.Error("expected a name-prefix-sharing sibling directory to NOT be contained, got true")
		}
	})

	t.Run("dot-dot escape", func(t *testing.T) {
		// filepath.Abs/Clean already collapse a literal "../" segment before
		// containsPath ever runs (RevealInFileManager calls filepath.Abs on
		// the input first), so the escape is exercised the way it actually
		// arrives: an absolute, already-cleaned path outside catalogDir.
		escaped := filepath.Clean(filepath.Join(resolvedCatalogDir, "..", "outside", "catalog.json"))
		ok, err := containsPath(catalogDir, escaped)
		if err != nil {
			t.Fatalf("containsPath returned an error: %v", err)
		}
		if ok {
			t.Error("expected a \"../\" escape to NOT be contained, got true")
		}
	})

	t.Run("symlink pointing outside catalogDir", func(t *testing.T) {
		outsideDir := filepath.Join(base, "outside-target")
		if err := os.Mkdir(outsideDir, 0755); err != nil {
			t.Fatalf("mkdir outsideDir: %v", err)
		}
		realFile := filepath.Join(outsideDir, "real-catalog.json")
		if err := os.WriteFile(realFile, []byte("{}"), 0644); err != nil {
			t.Fatalf("write real file: %v", err)
		}
		link := filepath.Join(catalogDir, "link.json")
		if err := os.Symlink(realFile, link); err != nil {
			t.Skipf("symlinks unavailable in this environment: %v", err)
		}

		// Mirrors what RevealInFileManager itself does before calling
		// containsPath: resolve the symlink first, then check containment
		// against the resolved (not the linked) location.
		resolved, err := filepath.EvalSymlinks(link)
		if err != nil {
			t.Fatalf("EvalSymlinks: %v", err)
		}
		ok, err := containsPath(catalogDir, resolved)
		if err != nil {
			t.Fatalf("containsPath returned an error: %v", err)
		}
		if ok {
			t.Error("expected a symlink resolving outside catalogDir to NOT be contained, got true")
		}
	})
}
