// Command killtarget is a standalone helper process used ONLY by
// atomicwrite_sigkill_test.go's TestWriteFileAtomic_SurvivesKill and
// TestWriteFileAtomic_SurvivesKill_NoPriorFile. It reproduces
// WriteFileAtomic's write sequence locally (it deliberately imports no
// package from this project and never calls the real WriteFileAtomic) with
// an injected artificial delay between the two halves of the payload write,
// widening the real-world few-millisecond temp-then-rename window to
// something a parent test can reliably schedule a SIGKILL inside.
//
// This artificial delay is a test-only variant of the write sequence and
// must never be added to production WriteFileAtomic.
//
// It lives under testdata/ so `go build ./...` and `go vet ./...` skip it
// entirely -- the Go toolchain ignores testdata directories.
//
// Usage: killtarget <dest> <payloadBytes> <delayMs>
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: killtarget <dest> <payloadBytes> <delayMs>")
		os.Exit(2)
	}

	dest := os.Args[1]
	payloadBytes, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad payload size:", err)
		os.Exit(2)
	}
	delayMs, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad delay:", err)
		os.Exit(2)
	}

	dir := filepath.Dir(dest)
	payload := make([]byte, payloadBytes)
	for i := range payload {
		payload[i] = byte('A' + i%26)
	}
	half := len(payload) / 2

	tmp, err := os.CreateTemp(dir, "storcat-*.tmp")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create temp:", err)
		os.Exit(1)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(payload[:half]); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		fmt.Fprintln(os.Stderr, "write first half:", err)
		os.Exit(1)
	}

	// Synchronization marker, not polling-on-hope: written and fsync'd
	// after the first half is written but before the artificial delay,
	// so the parent test can wait for a deterministic signal instead of
	// racing the temp file's mere appearance in the directory listing.
	markerPath := dest + ".killtarget-ready"
	marker, err := os.Create(markerPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create marker:", err)
		os.Exit(1)
	}
	if err := marker.Sync(); err != nil {
		marker.Close()
		fmt.Fprintln(os.Stderr, "sync marker:", err)
		os.Exit(1)
	}
	if err := marker.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "close marker:", err)
		os.Exit(1)
	}

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	if _, err := tmp.Write(payload[half:]); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		fmt.Fprintln(os.Stderr, "write second half:", err)
		os.Exit(1)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		fmt.Fprintln(os.Stderr, "sync:", err)
		os.Exit(1)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		fmt.Fprintln(os.Stderr, "close:", err)
		os.Exit(1)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		os.Remove(tmpPath)
		fmt.Fprintln(os.Stderr, "chmod:", err)
		os.Exit(1)
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		os.Remove(tmpPath)
		fmt.Fprintln(os.Stderr, "rename:", err)
		os.Exit(1)
	}

	// Reached the end without being killed. The parent test treats this
	// as an inconclusive iteration, not a pass -- exit 0 either way.
	os.Exit(0)
}
