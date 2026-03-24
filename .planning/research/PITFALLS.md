# Domain Pitfalls: Electron-to-Wails Migration

**Domain:** Desktop app migration (Electron/Node.js to Go/Wails)
**Project:** StorCat v2.0.0
**Researched:** 2026-03-24

---

## Critical Pitfalls

Mistakes that cause silent data corruption, runtime panics, or feature regressions that are hard to detect in testing.

---

### Pitfall 1: IPC Response Envelope Mismatch

**What goes wrong:** Electron's `ipcMain`/`ipcRenderer` pattern conventionally wraps responses in `{success: boolean, data: any}` envelopes. Wails' IPC layer does not — bound Go methods return raw values directly (resolved as the Promise value) or throw errors (rejected as the Promise reason). When a Wails wrapper shim recreates `window.electronAPI`, it must manually construct the `{success, data}` envelope in the wrapper, not rely on the Go method to produce it.

**Why it happens:** Developers assume the IPC transport preserves the Electron contract. In Wails, a Go method returning `(string, error)` resolves the Promise with the raw `string`. The frontend's `.then(result => result.data)` silently gets `undefined`, because there is no `.data` property on a bare string.

**Consequences:** Silent `undefined` failures throughout the frontend. Operations appear to succeed (Promise resolved, no rejection) but data is never rendered. Regression only visible at runtime, not in TypeScript compilation.

**Prevention:**
- Every wrapper in the Wails API shim that replaces an Electron handler must explicitly construct `{success: true, data: result}` in the JavaScript wrapper function, not in Go.
- Alternatively, define Go structs for every response type and have Go return them — then the TypeScript type system will enforce the shape.
- Review every call site in the frontend that destructures `{success, data, ...}` from an API call; trace it back to the Go method and verify the shape is being constructed somewhere.

**Detection:** TypeScript types for wailsjs bindings will show the raw Go return type. If the frontend expects `{success: boolean}` but the binding shows `string`, the wrapper layer is missing the envelope.

**Phase:** Address in the wrapper/shim fix phase before any functional testing.

---

### Pitfall 2: Nil Slice Serializes as JSON `null`, Not `[]`

**What goes wrong:** In Go, a declared-but-uninitialized slice (`var contents []Item`) marshals to JSON `null`. An initialized-but-empty slice (`contents := make([]Item, 0)` or `contents := []Item{}`) marshals to `[]`. Electron's Node.js backend always produces `[]` for empty arrays, never `null`. Frontend code that does `items.map(...)` throws `TypeError: Cannot read properties of null` when Go returns `null`.

**Why it happens:** Go's zero value for a slice is `nil`, which is JSON `null`. The distinction between "no items" and "not set" is meaningful in Go but invisible to the developer writing the struct definition. Adding `omitempty` to the struct tag makes it worse: it omits the field entirely, so the frontend gets no field at all rather than an empty array.

**Consequences:** Directory entries with no children render broken. The specific case in StorCat: `contents` field of leaf nodes or empty directories becomes `null`, breaking `Array.isArray()` checks and `.map()` calls silently (when the array is null, `.map` throws; when missing with omitempty, accessing it returns `undefined`).

**Prevention:**
- Initialize all slice fields explicitly: `Contents: []FileNode{}` not `var Contents []FileNode`.
- Do NOT use `omitempty` on slice fields unless `null` and missing are both acceptable to the frontend.
- For the JSON output format: after marshaling, audit the output for `null` array values with a test fixture.

**Detection:** Run the catalog creation against an empty directory and inspect raw JSON output. A `null` where `[]` is expected confirms the bug. TypeScript `Array.isArray(node.contents)` returning `false` at runtime is the symptom.

**Phase:** Fix during JSON output format correctness work.

---

### Pitfall 3: filepath.Walk Skips Symlinks Silently

**What goes wrong:** `filepath.Walk` and `filepath.WalkDir` use `os.Lstat` internally and explicitly do NOT follow symbolic links. A directory tree containing symlinks to subdirectories will produce an incomplete catalog — symlinked directories appear as files with a symlink mode bit, not as traversed directories. Node.js's `fs.stat` follows symlinks by default. This is a silent discrepancy: no error is thrown, the catalog just misses content.

**Why it happens:** Go's walk functions are designed this way intentionally, to avoid infinite loops from circular symlinks. The developer using `filepath.Walk` assumes "it walks everything" without reading the symlink caveat.

**Consequences:** A catalog created in Wails/Go will have fewer entries than one created in Electron for the same directory tree if any symlinks exist. Users with symlink-heavy project structures (monorepos, version-controlled dotfiles, NixOS/Nix-based systems) will see silent data loss.

**Prevention:**
- Replace any `filepath.Walk` or `filepath.WalkDir` with a custom recursive walker that uses `os.Stat` (follows symlinks) to get file info after detecting a symlink via `os.Lstat`.
- Guard against circular symlinks by tracking visited real paths (via `os.Readlink` + `filepath.EvalSymlinks`).
- The pattern: `info, err := os.Lstat(path)` to detect type; if `ModeSymlink` is set, call `os.Stat(path)` to resolve the target.

**Detection:** Create a test directory with a symlink to a subdirectory. Compare Electron catalog output with Go catalog output. Missing entries = symlink traversal broken.

**Phase:** Fix during symlink handling work; add a test fixture with symlinks.

---

### Pitfall 4: Window State Persistence Coordinate System Drift on macOS

**What goes wrong:** Wails v2 has a documented bug where `WindowGetPosition()` and `WindowSetPosition()` use inconsistent coordinate systems on macOS — `GetPosition` uses `visibleFrame` (which subtracts the menubar height) while `SetPosition` uses `frame`. Saving position on close and restoring it on open causes the window to drift upward by ~24px on each launch.

**Why it happens:** macOS has two coordinate systems: screen frame (origin at bottom-left of physical screen) and visible frame (origin adjusted for menubar and dock). Wails internally mixed them. This was fixed in a patch release, but the fix requires being on a recent enough Wails version.

**Consequences:** If window position persistence is implemented naively against an unfixed Wails version, the window migrates toward the top of the screen across launches. On multi-monitor setups, the behavior is compounded: `WindowSetPosition` and `WindowGetPosition` are relative to the monitor the window is currently on, not absolute screen coordinates. Moving to a different monitor invalidates saved position.

**Prevention:**
- Verify the Wails version in use has the macOS coordinate fix (PR #3479 — "Fix Position() and SetPosition() using inconsistent coordinate systems on macOS").
- When saving/restoring position, also save the monitor the window was on and validate that the monitor still exists before restoring position.
- Clamp restored position to visible screen bounds as a safety net.
- Use `OnDomReady` (not `OnStartup`) to restore window state — the runtime API is not reliably available in `OnStartup`.

**Detection:** Launch app, move window, close, reopen. Check if window position drifted. Test on multi-monitor setup.

**Phase:** Address during window state persistence implementation.

---

## Moderate Pitfalls

---

### Pitfall 5: Wails Drag Regions Require Different CSS Than Electron

**What goes wrong:** Electron uses `-webkit-app-region: drag` (or `WebkitAppRegion: 'drag'` in React inline styles) for frameless window drag areas. Wails uses a custom CSS variable `--wails-draggable: drag`. These are not the same. Code migrated from Electron with the WebKit property applied but without the Wails CSS variable will produce a frameless window that cannot be dragged.

**Why it happens:** The property is different and there is no fallback — browsers simply ignore unknown CSS variables. No error is thrown in the console.

**Consequences:** The macOS custom titlebar is rendered but cannot be used to drag the window. Users cannot move the window at all (on macOS frameless builds). This is a total regression of a basic windowing feature.

**Prevention:**
- Replace every instance of `WebkitAppRegion: 'drag'` with `'--wails-draggable': 'drag'` (note: CSS variable syntax in React inline styles requires the string key form).
- Also replace `WebkitAppRegion: 'no-drag'` with `'--wails-draggable': 'no-drag'` for interactive elements inside the drag region.
- Test by attempting to drag the window from the custom titlebar area on macOS.

**Detection:** Frameless window on macOS cannot be moved by clicking and dragging the header area. Browser DevTools will not show any errors — the property silently does nothing.

**Phase:** Fix during header/drag region implementation.

---

### Pitfall 6: Wails-Generated TypeScript Bindings Are Not Automatically Kept in Sync

**What goes wrong:** The `wailsjs/` directory contains auto-generated TypeScript bindings from the Go struct definitions. If Go methods or structs are changed without regenerating bindings (via `wails generate module` or `wails build`), the TypeScript types are stale. The frontend may compile cleanly against old types while the runtime behavior has changed.

**Why it happens:** The generation step is separate from compilation. It is easy to edit Go code and rebuild without running the generation step, especially in a migration context where many methods are being added or changed.

**Consequences:** TypeScript compiles with old types. Return value types are wrong. Method signatures mismatch at runtime despite passing type-checking. Bugs only appear when the specific function is called.

**Prevention:**
- Add `wails generate module` as a pre-build step or makefile target run before any frontend development.
- After every Go struct or method signature change, immediately regenerate bindings before writing frontend code against the new shape.
- Check `wailsjs/go/` output files into version control so diffs make stale bindings visible in PR review.

**Detection:** Compare `wailsjs/go/models.ts` types against current Go struct definitions. Any discrepancy = stale bindings.

**Phase:** Address at project setup; enforce in the build workflow.

---

### Pitfall 7: Go time.Time Serializes as RFC3339, Not Unix Timestamp or Milliseconds

**What goes wrong:** Go's `time.Time` marshals to RFC3339 format by default (e.g., `"2024-01-15T10:30:00Z"`). Electron/Node.js backends typically return Unix timestamps (milliseconds since epoch) or locale-formatted strings. Frontend code written against the Electron API may use `new Date(timestamp)` where `timestamp` is a number. `new Date("2024-01-15T10:30:00Z")` also works, but `new Date(result.modified)` breaks if the frontend previously expected a number and now gets a string.

**Why it happens:** The types are compatible in one direction (string parses fine) but the format is unexpected if the original backend returned numbers. Also, Wails' TypeScript binding generator produces its own `Time` type rather than TypeScript's `Date`, requiring manual handling.

**Consequences:** Date display works inconsistently. Relative time calculations (`Date.now() - file.modified`) break silently if `file.modified` is a string. The `created` field in StorCat specifically has a semantic mismatch: it reflects ctime (inode change time on Unix) not creation time, which is different behavior from macOS's birthtime that Electron's `fs.stat` may have exposed.

**Prevention:**
- Audit every date/time field in the API response surface. Define whether the frontend expects a number (milliseconds) or ISO string.
- If the frontend expects numbers: use a custom struct with `MarshalJSON` returning Unix milliseconds, or compute the value in the Go method and return `int64`.
- For the `created`/`modified` field semantic issue: document explicitly that `created` means ctime (inode change), not file creation time, and consider renaming to avoid confusion.

**Detection:** Render a file listing in the Browse tab and inspect the date values in browser DevTools. A string where a number is expected, or an unexpected date value, confirms the issue.

**Phase:** Fix during browse metadata fields work.

---

### Pitfall 8: Runtime API Unavailable in OnStartup

**What goes wrong:** Calling Wails runtime functions (e.g., `runtime.WindowSetSize`, `runtime.EventsEmit`, dialog functions) from the `OnStartup` callback causes panics on Windows and unreliable behavior on all platforms because the window is still initializing in a separate goroutine.

**Why it happens:** `OnStartup` is called after the frontend loads `index.html` but before the DOM is fully ready and the window is fully initialized. The runtime's window API requires the window to be in a usable state.

**Consequences:** App panics on Windows on startup. On macOS, calls may silently fail or produce incorrect results. Window state restoration (size, position) applied in `OnStartup` may not take effect.

**Prevention:**
- Move all runtime API calls (especially window state restoration) to `OnDomReady`.
- `OnStartup` is safe only for: reading config files, initializing Go-side state, non-window operations.
- `OnDomReady` is safe for: window resize/reposition, emitting startup events, showing dialogs.

**Detection:** App panics immediately on startup (Windows), or window size/position is not restored on first launch.

**Phase:** Address during window state persistence implementation.

---

## Minor Pitfalls

---

### Pitfall 9: File Dialog Behavior Differs by Platform

**What goes wrong:** Wails file dialogs have platform-specific option availability. Mac-only options (`CanCreateDirectories`, `ResolvesAliases`, `AllowDirectories`) are silently ignored on Windows and Linux. File extension filters work differently: on macOS, multiple `FileFilter` entries are merged into one pattern set; on Windows and Linux they appear as separate choosable filter entries.

**Prevention:**
- Use only cross-platform dialog options for the core flow.
- Test file dialogs on all three target platforms before shipping.
- Do not rely on `CanCreateDirectories` being available — handle the case where the user picks a non-existent path.

**Phase:** Verify during cross-platform testing.

---

### Pitfall 10: JSON Root Object Format (Array vs Bare Object)

**What goes wrong:** Go's `json.Marshal` on a slice produces `[{...}]` (a JSON array). The Electron version produces a bare object `{...}`. If the Go backend marshals the catalog root as a slice, existing v1 catalog files produced by Electron will be read correctly (different schema version), but new files produced by Go will be structurally different from what external tools or v1-format readers expect.

**Why it happens:** The Go developer modeled the catalog as a `[]FileNode` and marshaled it directly. The Electron version returned a single root node object, not an array.

**Consequences:** v2-produced catalogs cannot be loaded by v1 readers. External tools parsing the JSON format break. Backward compatibility is violated for future users who might load v2 catalogs in v1 or third-party tools.

**Prevention:**
- Wrap the root node in a struct and marshal that: `json.Marshal(RootObject{Root: tree})` or serialize the single root `FileNode` directly as an object.
- Write a round-trip test: produce a catalog, read it back, verify the root is a JSON object not an array.
- Add a test fixture comparing Go output format against a known-good Electron output file.

**Phase:** Fix during JSON output format work (already identified as active requirement).

---

### Pitfall 11: webview Rendering Differences Across Platforms

**What goes wrong:** Wails uses the OS-native webview (WKWebView on macOS, WebView2 on Windows, WebKitGTK on Linux). Each has different CSS support, JavaScript engine versions, and rendering behavior. Electron bundles Chromium and is consistent across platforms.

**Why it happens:** This is inherent to Wails' architecture — it is a feature (smaller binary) but also a constraint.

**Consequences:** CSS that works in Electron's Chromium may behave differently on older WebView2 (Windows) or WebKitGTK (Linux). Ant Design components generally work, but platform-specific CSS features (especially CSS grid subgrid, newer selectors) may render differently.

**Prevention:**
- Test the full UI on Windows and Linux after development on macOS, not just at release time.
- Avoid cutting-edge CSS features not supported by WebView2 on Windows 10 (the oldest common target).
- Use `wails doctor` to confirm the webview versions on each platform before starting testing.

**Phase:** Verify during cross-platform testing pass.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| API wrapper / shim fixes | IPC envelope mismatch (Pitfall 1) | Explicitly wrap all Go return values in `{success, data}` in JS wrapper |
| JSON output format | Nil-slice null vs `[]` (Pitfall 2), root array vs object (Pitfall 10) | Initialize all slices; write round-trip format test |
| Symlink handling | Silent traversal skip (Pitfall 3) | Custom walker with os.Stat follow; cycle detection |
| Window state persistence | Coordinate drift on macOS (Pitfall 4), OnStartup timing (Pitfall 8) | Use OnDomReady; validate Wails version has the coordinate fix |
| Header drag region | Wrong CSS property (Pitfall 5) | Replace WebkitAppRegion with --wails-draggable |
| Browse metadata fields | time.Time format (Pitfall 7), created/ctime semantics | Audit date types; document ctime vs birthtime difference |
| Any Go struct changes | Stale wailsjs bindings (Pitfall 6) | Regenerate bindings immediately after struct changes |
| Cross-platform testing | Platform webview differences (Pitfall 11), dialog options (Pitfall 9) | Test on all 3 platforms before declaring complete |

---

## Sources

- Wails frameless window docs: https://wails.io/docs/guides/frameless/
- Wails window runtime API: https://wails.io/docs/reference/runtime/window/
- Wails IPC internals: https://wails.io/docs/howdoesitwork/
- Wails application development guide: https://wails.io/docs/guides/application-development/
- Wails dialog reference: https://wails.io/docs/reference/runtime/dialog/
- Wails macOS window position fix (PR #3479): https://github.com/wailsapp/wails/pull/3479
- Wails window position multi-monitor issue #2739: https://github.com/wailsapp/wails/issues/2739
- Wails TypeScript type mismatch issue #2258: https://github.com/wailsapp/wails/issues/2258
- Wails OnStartup runtime panic issue #1660: https://github.com/wailsapp/wails/issues/1660
- Go filepath.Walk symlink behavior (issue #4759): https://github.com/golang/go/issues/4759
- Go nil slice JSON null vs empty array: https://apoorvam.github.io/blog/2017/golang-json-marshal-slice-as-empty-array-not-null/
- Go JSON surprises and gotchas: https://www.alexedwards.net/blog/json-surprises-and-gotchas
- Go os.Stat vs os.Lstat: https://dev.to/moseeh_52/understanding-osstat-vs-oslstat-in-go-file-and-symlink-handling-3p5d
- Go time.Time JSON serialization: https://www.willem.dev/articles/change-time-format-json/
- Wails as Electron alternative (migration overview): https://dev.to/kartik_patel/wails-as-electron-alternative-4dmn
- Getting started with Wails replacing Electron: https://www.codingexplorations.com/blog/getting-started-with-wails-replacing-electron-app
