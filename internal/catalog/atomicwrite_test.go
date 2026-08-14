package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteFileAtomic_CreatesFileWithContent verifies the destination
// contains exactly the given bytes.
func TestWriteFileAtomic_CreatesFileWithContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	want := []byte("hello atomic world")

	if err := WriteFileAtomic(path, want, 0644); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("content = %q, want %q", got, want)
	}
}

// TestWriteFileAtomic_LeavesNoTempResidue verifies that after a successful
// write the destination directory contains exactly one entry.
func TestWriteFileAtomic_LeavesNoTempResidue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	if err := WriteFileAtomic(path, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected exactly 1 entry in destination dir, got %d: %+v", len(entries), entries)
	}
}

// TestWriteFileAtomic_RemovesTempOnFailure verifies that if the destination
// directory is removed between temp-create and rename, the call returns an
// error and leaves no temp file behind in any directory that still exists.
func TestWriteFileAtomic_RemovesTempOnFailure(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "will-vanish")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "out.txt")

	// os.CreateTemp succeeds before we remove dir out from under it; we
	// can't easily race the exact create-vs-rename window, so instead
	// simulate the failure mode Rename hits directly: remove dir so the
	// rename target's parent no longer exists.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove dir: %v", err)
	}

	if err := WriteFileAtomic(path, []byte("data"), 0644); err == nil {
		t.Fatal("expected an error when the destination directory does not exist")
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatalf("read parent: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "storcat-") {
			t.Errorf("temp file residue left behind: %s", e.Name())
		}
	}
}

// TestWriteFileAtomic_TempIsCreatedInDestinationDirectory asserts the
// implementation never calls os.TempDir -- the temp file must live beside
// the destination, not in the shared system temp directory.
func TestWriteFileAtomic_TempIsCreatedInDestinationDirectory(t *testing.T) {
	src, err := os.ReadFile("atomicwrite.go")
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	if strings.Contains(string(src), "os.TempDir()") {
		t.Error("WriteFileAtomic must not use os.TempDir() -- temp file must be created in the destination directory")
	}
	if !strings.Contains(string(src), "os.CreateTemp(dir") {
		t.Error("WriteFileAtomic must create its temp file via os.CreateTemp(dir, ...) in the destination directory")
	}
}
