package watch

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- coalescer tests -- fast, deterministic, no fsnotify/real filesystem
// involved. Always-on (not gated by -short) since they are the
// deterministic half of this package's behavior.

func TestCoalescer_CollapsesBurst(t *testing.T) {
	var calls int32
	done := make(chan struct{}, 1)
	c := newCoalescer(30*time.Millisecond, func() {
		atomic.AddInt32(&calls, 1)
		select {
		case done <- struct{}{}:
		default:
		}
	})

	for i := 0; i < 50; i++ {
		c.trigger()
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("callback did not fire within bounded wait")
	}

	// Give any stray extra timer a chance to also fire before asserting.
	time.Sleep(100 * time.Millisecond)
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("want exactly 1 callback for a 50-call burst, got %d", got)
	}
}

func TestCoalescer_ResetsOnEachTrigger(t *testing.T) {
	const window = 60 * time.Millisecond
	var calls int32
	fired := make(chan time.Time, 1)
	c := newCoalescer(window, func() {
		atomic.AddInt32(&calls, 1)
		fired <- time.Now()
	})

	var last time.Time
	for i := 0; i < 4; i++ {
		time.Sleep(window / 2)
		c.trigger()
		last = time.Now()
	}

	select {
	case ft := <-fired:
		if ft.Sub(last) < window-20*time.Millisecond {
			t.Fatalf("callback fired too early: %v after the last trigger (want ~%v)", ft.Sub(last), window)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("callback never fired")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("want exactly 1 callback, got %d", got)
	}
}

func TestCoalescer_FiresNowIsImmediate(t *testing.T) {
	var calls int32
	c := newCoalescer(2*time.Second, func() {
		atomic.AddInt32(&calls, 1)
	})

	c.trigger()
	c.fireNow()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("want 1 immediate callback, got %d", got)
	}

	// Confirm the deferred timer that trigger() started was cancelled --
	// no second fire once the original window would have elapsed.
	time.Sleep(200 * time.Millisecond)
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("want still 1 callback after the original window, got %d (fireNow should have cancelled the pending deferred fire)", got)
	}
}

func TestCoalescer_StopCancelsPending(t *testing.T) {
	var calls int32
	c := newCoalescer(30*time.Millisecond, func() {
		atomic.AddInt32(&calls, 1)
	})

	c.trigger()
	c.stop()

	time.Sleep(150 * time.Millisecond)
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("want 0 callbacks after stop() cancels a pending trigger, got %d", got)
	}

	// A late trigger after stop() must not resurrect the coalescer.
	c.trigger()
	time.Sleep(150 * time.Millisecond)
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("want 0 callbacks after a trigger following stop(), got %d", got)
	}
}

func TestCoalescer_ConcurrentTriggers(t *testing.T) {
	var calls int32
	c := newCoalescer(30*time.Millisecond, func() {
		atomic.AddInt32(&calls, 1)
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.trigger()
		}()
	}
	wg.Wait()

	time.Sleep(300 * time.Millisecond)
	got := atomic.LoadInt32(&calls)
	if got < 1 || got > 5 {
		t.Fatalf("want between 1 and 5 callbacks for 100 concurrent triggers, got %d", got)
	}
}

// --- Watcher tests -- against a real temp directory (kqueue on this dev
// machine). Skippable under -short since they depend on real OS event
// delivery timing.

func newTestWatcher(t *testing.T, dir string, debounce time.Duration) (*Watcher, <-chan struct{}) {
	t.Helper()
	ch := make(chan struct{}, 100)
	w, err := New(dir, debounce, func() {
		select {
		case ch <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("New(%q): %v", dir, err)
	}
	t.Cleanup(func() { _ = w.Close() })
	return w, ch
}

func waitForSignal(t *testing.T, ch <-chan struct{}, timeout time.Duration) bool {
	t.Helper()
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

func TestWatcher_FiresOnJSONCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("real filesystem watch test")
	}
	dir := t.TempDir()
	_, ch := newTestWatcher(t, dir, 50*time.Millisecond)

	if err := os.WriteFile(filepath.Join(dir, "a.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if !waitForSignal(t, ch, 2*time.Second) {
		t.Fatal("expected a callback within bounded wait after creating a .json file")
	}
	select {
	case <-ch:
		t.Fatal("unexpected second callback for a single create")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestWatcher_FiresOnRemoveAndRename(t *testing.T) {
	if testing.Short() {
		t.Skip("real filesystem watch test")
	}
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "a.json")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, ch := newTestWatcher(t, dir, 50*time.Millisecond)

	if err := os.Remove(jsonPath); err != nil {
		t.Fatal(err)
	}
	if !waitForSignal(t, ch, 2*time.Second) {
		t.Fatal("expected a callback on remove")
	}

	b := filepath.Join(dir, "b.json")
	if err := os.WriteFile(b, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if !waitForSignal(t, ch, 2*time.Second) {
		t.Fatal("expected a callback on create (rename setup)")
	}

	c := filepath.Join(dir, "c.json")
	if err := os.Rename(b, c); err != nil {
		t.Fatal(err)
	}
	if !waitForSignal(t, ch, 2*time.Second) {
		t.Fatal("expected a callback on rename")
	}
}

func TestWatcher_IgnoresUnrelatedExtension(t *testing.T) {
	if testing.Short() {
		t.Skip("real filesystem watch test")
	}
	dir := t.TempDir()
	_, ch := newTestWatcher(t, dir, 50*time.Millisecond)

	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
		t.Fatal("unexpected callback for a non .json/.html file")
	case <-time.After(300 * time.Millisecond):
	}
}

func TestWatcher_CoalescesAtomicWritePair(t *testing.T) {
	if testing.Short() {
		t.Skip("real filesystem watch test")
	}
	dir := t.TempDir()
	_, ch := newTestWatcher(t, dir, 80*time.Millisecond)

	// The same shape WriteFileAtomic produces: a temp file (non-matching
	// extension, so its own Create/Write is filtered out), then a rename
	// onto the real .json destination -- fsnotify surfaces this as a
	// Rename+Create pair rather than a single Write (27-RESEARCH.md
	// Pitfall 5).
	tmp := filepath.Join(dir, "storcat-123456.tmp")
	if err := os.WriteFile(tmp, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "a.json")
	if err := os.Rename(tmp, dest); err != nil {
		t.Fatal(err)
	}

	if !waitForSignal(t, ch, 2*time.Second) {
		t.Fatal("expected a callback for the temp-then-rename pair")
	}
	select {
	case <-ch:
		t.Fatal("unexpected second callback for a single atomic write")
	case <-time.After(300 * time.Millisecond):
	}
}

func TestWatcher_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("real filesystem watch test")
	}
	dir := t.TempDir()
	w, ch := newTestWatcher(t, dir, 50*time.Millisecond)

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "a.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
		t.Fatal("unexpected callback after Close")
	case <-time.After(300 * time.Millisecond):
	}
}

func TestWatcher_CloseIsIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("real filesystem watch test")
	}
	dir := t.TempDir()
	w, err := New(dir, 50*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestWatcher_RejectsEmptyDir(t *testing.T) {
	w, err := New("", DefaultDebounce, func() {})
	if err == nil {
		t.Fatal("expected an error for an empty dir")
	}
	if w != nil {
		t.Fatal("expected a nil watcher on error")
	}
}

func TestWatcher_RejectsNilCallback(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, DefaultDebounce, nil)
	if err == nil {
		t.Fatal("expected an error for a nil callback")
	}
	if w != nil {
		t.Fatal("expected a nil watcher on error")
	}
}
