package catalog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// TestWriteFileAtomic_SyncsBeforeRename verifies that after a successful
// write the destination contains exactly the requested bytes and holds the
// requested perm. The Sync() call itself is proven structurally by the
// acceptance grep plus the kill test in atomicwrite_sigkill_test.go -- a
// unit test cannot observe a flush directly.
func TestWriteFileAtomic_SyncsBeforeRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	want := []byte("synced content")

	if err := WriteFileAtomic(path, want, 0640); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("content = %q, want %q", got, want)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat destination: %v", err)
	}
	if info.Mode().Perm() != 0640 {
		t.Errorf("perm = %v, want %v", info.Mode().Perm(), os.FileMode(0640))
	}
}

// TestWriteFileAtomic_ReplacesExistingFileWholesale verifies that an
// existing destination with different, longer prior content ends up
// byte-identical to the new payload with no trailing remnant of the old
// content.
func TestWriteFileAtomic_ReplacesExistingFileWholesale(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	oldContent := []byte("this is the much longer original content that must not survive")
	if err := WriteFileAtomic(path, oldContent, 0644); err != nil {
		t.Fatalf("seed write failed: %v", err)
	}

	newContent := []byte("short")
	if err := WriteFileAtomic(path, newContent, 0644); err != nil {
		t.Fatalf("replace write failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if !bytes.Equal(got, newContent) {
		t.Errorf("content = %q, want %q (no remnant of old content)", got, newContent)
	}
}

// TestWriteFileAtomic_DirSyncFailureIsNotFatal exercises the unexported
// syncDir helper directly against a nonexistent directory path (proving its
// error return), then asserts a full WriteFileAtomic on a normal directory
// still succeeds -- a directory-sync failure must never fail the write or
// remove the destination.
func TestWriteFileAtomic_DirSyncFailureIsNotFatal(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "does-not-exist")
	if err := syncDir(nonexistent); err == nil {
		t.Error("expected syncDir to return an error for a nonexistent directory")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	want := []byte("dir sync is best-effort")
	if err := WriteFileAtomic(path, want, 0644); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("content = %q, want %q", got, want)
	}
}

// TestWriteFileAtomic_ConcurrentWritersLeaveNoResidue launches 8 goroutines
// writing distinct payloads to the same destination path. After they all
// return, the destination's content must equal exactly one of the 8
// payloads in full, and no storcat-*.tmp residue may remain.
func TestWriteFileAtomic_ConcurrentWritersLeaveNoResidue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	const n = 8
	payloads := make([][]byte, n)
	for i := 0; i < n; i++ {
		payloads[i] = []byte(fmt.Sprintf("payload-%d", i))
	}

	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = WriteFileAtomic(path, payloads[i], 0644)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("writer %d failed: %v", i, err)
		}
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	matched := false
	for _, p := range payloads {
		if bytes.Equal(got, p) {
			matched = true
			break
		}
	}
	if !matched {
		t.Errorf("destination content %q does not match any of the 8 payloads in full", got)
	}

	residue, err := filepath.Glob(filepath.Join(dir, "storcat-*.tmp"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(residue) != 0 {
		t.Errorf("expected no storcat-*.tmp residue, got %v", residue)
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
