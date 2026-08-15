package volumes

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSkipMountEntry covers the shared filter's two rules against a
// hidden entry, a vendor-reserved entry, and a table of names that must
// survive unmodified -- an ordinary name, one with spaces, one with
// non-ASCII characters, and one with an emoji (CRT-02 encoding).
func TestSkipMountEntry(t *testing.T) {
	tests := []struct {
		name string
		skip bool
	}{
		{".timemachine", true},
		{"com.apple.TimeMachine.localsnapshots", true},
		{"pi-downloader", false},
		{"My Backup Drive", false},
		{"Café Backup", false},
		{"Drive 🎉", false},
	}
	for _, tt := range tests {
		if got := skipMountEntry(tt.name); got != tt.skip {
			t.Errorf("skipMountEntry(%q) = %v, want %v", tt.name, got, tt.skip)
		}
	}
}

// TestVolumeNameFromMountPath verifies the display name is the final path
// element, preserved byte-for-byte including spaces and non-ASCII
// characters -- no mangling before this name reaches the frontend or is
// passed back to the scanner (CRT-02 encoding).
func TestVolumeNameFromMountPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/Volumes/My Backup Drive", "My Backup Drive"},
		{"/Volumes/Café Backup", "Café Backup"},
		{"/Volumes/Drive 🎉", "Drive 🎉"},
		{"/media/pi/USBDRIVE", "USBDRIVE"},
	}
	for _, tt := range tests {
		if got := volumeNameFromMountPath(tt.path); got != tt.want {
			t.Errorf("volumeNameFromMountPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// TestList_ReturnsEmptySliceNotErrorWhenNoMounts exercises List's
// portable core (buildVolumeList) directly against zero candidates,
// rather than List() itself, so the assertion holds regardless of what
// this machine's real mount namespace happens to contain.
func TestList_ReturnsEmptySliceNotErrorWhenNoMounts(t *testing.T) {
	got := buildVolumeList(nil)
	if got == nil {
		t.Fatal("buildVolumeList(nil) returned a nil slice, want a non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("buildVolumeList(nil) = %+v, want an empty slice", got)
	}
}

// TestProbeReadable_UnreadableDirectory verifies a directory whose
// permissions deny read reports not-readable while a normal directory
// reports readable. Skipped on Windows, whose permission model has no
// direct equivalent to the unix mode bits this test manipulates, and when
// running as root, which ignores those bits entirely.
func TestProbeReadable_UnreadableDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-style permission bits don't apply on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses permission bits")
	}

	dir := t.TempDir()

	unreadable := filepath.Join(dir, "no-read")
	if err := os.Mkdir(unreadable, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(unreadable, 0o311); err != nil { // d--x--x--x, matches the live pi-downloader/software observation
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o755) }) // let t.TempDir's own cleanup remove it

	if probeReadable(unreadable) {
		t.Error("probeReadable(unreadable dir) = true, want false")
	}

	readable := filepath.Join(dir, "can-read")
	if err := os.Mkdir(readable, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if !probeReadable(readable) {
		t.Error("probeReadable(normal dir) = false, want true")
	}
}
