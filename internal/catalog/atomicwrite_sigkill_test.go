package catalog

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// buildKillTargetHelper builds the standalone testdata/killtarget helper
// binary once and returns its path. A build failure is t.Fatal.
func buildKillTargetHelper(t *testing.T) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "killtarget-bin")
	cmd := exec.Command("go", "build", "-o", binPath, "./testdata/killtarget")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build killtarget helper: %v\n%s", err, out)
	}
	return binPath
}

// killDelaysMs is the varied set of injected mid-write delays (milliseconds)
// cycled across iterations so the SIGKILL lands at different points in the
// write window rather than proving only one timing.
var killDelaysMs = []int{20, 50, 120}

// waitForMarker polls for markerPath's existence with a short bounded poll.
// A missed marker is a t.Fatal, never a silent skip -- a test that quietly
// stops proving anything is worse than a red one.
func waitForMarker(t *testing.T, markerPath string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(markerPath); err == nil {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("killtarget-ready marker never appeared at %s within deadline", markerPath)
}

// waitForProcessDeath waits for cmd (already Kill()ed) to finish and asserts
// it really died from a signal rather than exiting normally -- ProcessState
// .Exited() is portable across platforms and returns false when the process
// was terminated by a signal (Unix SIGKILL) rather than exiting on its own.
func waitForProcessDeath(t *testing.T, cmd *exec.Cmd, iteration int) {
	t.Helper()
	err := cmd.Wait()
	if err == nil {
		t.Fatalf("iteration %d: helper process exited cleanly (exit 0) instead of being killed", iteration)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("iteration %d: unexpected Wait error (not an ExitError): %v", iteration, err)
	}
	if exitErr.ProcessState.Exited() {
		t.Fatalf("iteration %d: helper process exited normally (code %d) instead of being killed by a signal", iteration, exitErr.ExitCode())
	}
}

// cleanupIteration removes any storcat-*.tmp residue and the marker file
// left behind by a killed iteration, returning the count of temp files
// found. Temp residue at the temp path is EXPECTED after a SIGKILL -- the
// killed process never reaches its own os.Remove -- so this is reported,
// not asserted to be zero. What must be zero is residue at the destination
// path itself, asserted separately by the caller.
func cleanupIteration(t *testing.T, dir, markerPath string) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "storcat-*.tmp"))
	if err != nil {
		t.Fatalf("glob temp residue: %v", err)
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			t.Logf("cleanup: failed to remove residue %s: %v", m, err)
		}
	}
	os.Remove(markerPath) // best-effort; may already be absent
	return len(matches)
}

// TestWriteFileAtomic_SurvivesKill proves ACT-09's actual claim with a real,
// separately-launched OS process: a SIGKILL delivered mid-write leaves a
// pre-existing destination file byte-identical to its content before the
// write started, across at least 20 iterations at varied delays.
func TestWriteFileAtomic_SurvivesKill(t *testing.T) {
	if testing.Short() {
		t.Skip("subprocess SIGKILL harness is slow; skipped under -short")
	}

	helperBin := buildKillTargetHelper(t)
	dir := t.TempDir()
	dest := filepath.Join(dir, "catalog.json")
	markerPath := dest + ".killtarget-ready"

	seeded := []byte(`{"name":"pre-existing-catalog","contents":[]}`)
	if err := WriteFileAtomic(dest, seeded, 0644); err != nil {
		t.Fatalf("seed destination: %v", err)
	}
	seededSum := sha256.Sum256(seeded)

	const iterations = 21 // >= 20, exact multiple of len(killDelaysMs)
	totalResidue := 0

	for i := 0; i < iterations; i++ {
		delayMs := killDelaysMs[i%len(killDelaysMs)]

		cmd := exec.Command(helperBin, dest, "512", fmt.Sprintf("%d", delayMs))
		if err := cmd.Start(); err != nil {
			t.Fatalf("iteration %d: start helper: %v", i, err)
		}

		waitForMarker(t, markerPath)

		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("iteration %d: Kill: %v", i, err)
		}
		waitForProcessDeath(t, cmd, i)

		got, err := os.ReadFile(dest)
		if err != nil {
			t.Fatalf("iteration %d: read destination after kill: %v", i, err)
		}
		gotSum := sha256.Sum256(got)
		if len(got) != len(seeded) || !bytes.Equal(gotSum[:], seededSum[:]) {
			t.Fatalf("iteration %d (delay %dms): destination corrupted -- len=%d want=%d, sha256=%x want=%x",
				i, delayMs, len(got), len(seeded), gotSum, seededSum)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(got, &parsed); err != nil {
			t.Fatalf("iteration %d (delay %dms): destination does not parse as JSON after kill (truncated?): %v", i, delayMs, err)
		}

		totalResidue += cleanupIteration(t, dir, markerPath)
	}

	t.Logf("SurvivesKill: %d iterations, %d total storcat-*.tmp residue files left behind by killed processes (expected -- the killed process never reaches its own cleanup)", iterations, totalResidue)
}

// TestWriteFileAtomic_SurvivesKill_NoPriorFile proves the second half of
// ACT-09's claim: with no pre-existing destination, a mid-write kill leaves
// the destination ABSENT, never present-but-truncated.
func TestWriteFileAtomic_SurvivesKill_NoPriorFile(t *testing.T) {
	if testing.Short() {
		t.Skip("subprocess SIGKILL harness is slow; skipped under -short")
	}

	helperBin := buildKillTargetHelper(t)
	dir := t.TempDir()
	dest := filepath.Join(dir, "new-catalog.json")
	markerPath := dest + ".killtarget-ready"

	const iterations = 21
	totalResidue := 0

	for i := 0; i < iterations; i++ {
		delayMs := killDelaysMs[i%len(killDelaysMs)]

		cmd := exec.Command(helperBin, dest, "512", fmt.Sprintf("%d", delayMs))
		if err := cmd.Start(); err != nil {
			t.Fatalf("iteration %d: start helper: %v", i, err)
		}

		waitForMarker(t, markerPath)

		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("iteration %d: Kill: %v", i, err)
		}
		waitForProcessDeath(t, cmd, i)

		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			t.Fatalf("iteration %d (delay %dms): destination must be absent after a kill with no prior file, got err=%v", i, delayMs, err)
		}

		totalResidue += cleanupIteration(t, dir, markerPath)
	}

	t.Logf("SurvivesKill_NoPriorFile: %d iterations, %d total storcat-*.tmp residue files left behind by killed processes (expected)", iterations, totalResidue)
}
