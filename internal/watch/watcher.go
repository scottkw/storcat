// Package watch watches a single, flat directory for filesystem changes and
// coalesces event bursts into a single debounced callback. Catalogs live as
// flat .json/.html files directly inside one configured directory -- never
// nested -- so the watch is deliberately non-recursive.
//
// This package imports nothing from the Wails runtime and knows nothing of
// a frontend, so it stays usable from the CLI with no Wails runtime
// attached (COMPAT-04). The caller supplies a plain func(); app.go is the
// sole place in this repository that translates that callback into a
// runtime.EventsEmit call.
//
// The supplied callback may be invoked from a timer goroutine (via
// time.AfterFunc), not from the watcher's own event-loop goroutine, so the
// caller's closure must be safe to call concurrently with its own state.
package watch

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// DefaultDebounce is the locked ~300ms trailing debounce window: a named
// constant rather than a magic number, so a future tuning has one place to
// change.
const DefaultDebounce = 300 * time.Millisecond

// coalescer collapses repeated trigger() calls arriving within d of one
// another into a single fn() invocation, fired d after the last trigger
// (trailing debounce).
type coalescer struct {
	mu      sync.Mutex
	d       time.Duration
	fn      func()
	timer   *time.Timer
	stopped bool
}

func newCoalescer(d time.Duration, fn func()) *coalescer {
	return &coalescer{d: d, fn: fn}
}

// trigger (re)starts the trailing debounce window. A no-op once stopped.
func (c *coalescer) trigger() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopped {
		return
	}
	if c.timer != nil {
		c.timer.Stop()
	}
	c.timer = time.AfterFunc(c.d, c.fn)
}

// fireNow invokes fn immediately, cancelling any pending deferred fire. The
// mutex is released before fn runs so a slow callback cannot block the
// event loop or deadlock against a concurrent trigger().
func (c *coalescer) fireNow() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	fn := c.fn
	c.mu.Unlock()
	fn()
}

// stop cancels any pending fire and marks the coalescer stopped, so a late
// trigger() cannot resurrect it after Close().
func (c *coalescer) stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	c.stopped = true
}

// Watcher wraps an fsnotify.Watcher on a single directory, coalescing
// Create/Write/Remove/Rename events on .json/.html files into one debounced
// callback and draining the Errors channel in the same select loop so an
// unread error can never stall the watcher.
type Watcher struct {
	fsw  *fsnotify.Watcher
	c    *coalescer
	done chan struct{}

	closeOnce sync.Once
	closeErr  error
}

// New starts watching dir -- non-recursively, the single directory only --
// and calls onChange, debounced by debounce (falling back to
// DefaultDebounce when debounce is non-positive), whenever a
// Create/Write/Remove/Rename event fires on a .json or .html file inside
// it. onChange must not import the Wails runtime; the caller supplies a
// closure that does the eventual emit.
func New(dir string, debounce time.Duration, onChange func()) (*Watcher, error) {
	if dir == "" {
		return nil, errors.New("watch: dir must not be empty")
	}
	if onChange == nil {
		return nil, errors.New("watch: onChange must not be nil")
	}
	if debounce <= 0 {
		debounce = DefaultDebounce
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("watch: new watcher: %w", err)
	}
	// The single directory only, never a recursive walk of its contents --
	// catalogs are flat .json/.html files directly inside it.
	if err := fsw.Add(dir); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("watch: add %s: %w", dir, err)
	}

	w := &Watcher{
		fsw:  fsw,
		c:    newCoalescer(debounce, onChange),
		done: make(chan struct{}),
	}
	go w.loop()
	return w, nil
}

func (w *Watcher) loop() {
	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			ext := filepath.Ext(event.Name)
			if ext != ".json" && ext != ".html" {
				continue
			}
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) ||
				event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				w.c.trigger()
			}
		case _, ok := <-w.fsw.Errors:
			// Drained in the same select as Events -- an unread Errors
			// channel can stall the watcher outright. Any error, including
			// fsnotify.ErrEventOverflow, gets the same recovery as a real
			// change: "we may have missed events" and "something changed"
			// both resolve to the same idempotent re-list, so there is no
			// delta to reconcile and nothing to log-and-drop.
			if !ok {
				return
			}
			w.c.fireNow()
		case <-w.done:
			return
		}
	}
}

// Close genuinely releases the underlying OS watch handle (inotify fd,
// kqueue fd, or Windows handle) -- WATCH-03 requires release, not merely
// ignoring events. Idempotent and safe to call more than once.
func (w *Watcher) Close() error {
	w.closeOnce.Do(func() {
		w.c.stop()
		close(w.done)
		w.closeErr = w.fsw.Close()
	})
	return w.closeErr
}
