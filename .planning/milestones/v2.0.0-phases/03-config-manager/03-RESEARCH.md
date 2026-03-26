# Phase 3: Config Manager - Research

**Researched:** 2026-03-25
**Domain:** Go config package extension + Wails v2 window lifecycle + TypeScript stub replacement
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| WIN-01 | Window size (width, height) persists across app restarts via Go config | Config struct already has `WindowWidth`/`WindowHeight`; `SetWindowSize` method exists; gap is `OnDomReady` hook not wired to restore them |
| WIN-02 | Window position (x, y) persists across app restarts (size-only on macOS if coordinate drift unresolved) | `Config` struct missing `WindowX`/`WindowY` fields; `SetWindowPosition` method missing; Wails `runtime.WindowGetPosition`/`WindowSetPosition` APIs verified |
| WIN-03 | Settings toggle enables/disables window state persistence (not a stub) | `wailsAPI.ts` has hardcoded stub for `getWindowPersistence`/`setWindowPersistence`; `Config` struct missing `WindowPersistenceEnabled` field; Go methods missing |
</phase_requirements>

---

## Summary

Phase 3 completes the window state persistence foundation so the App layer can save and restore window size, position, and the user's persistence toggle preference. The work is scoped to three areas:

1. **`internal/config/config.go`** — Add three missing fields (`WindowX`, `WindowY`, `WindowPersistenceEnabled`) and two new methods (`SetWindowPosition`, `SetWindowPersistence` / `GetWindowPersistence`).
2. **`app.go`** — Add `GetWindowPersistence`/`SetWindowPersistence` bound methods; wire `OnDomReady` in `main.go` to restore window size+position from config; add `GetWindowPosition`/`SetWindowPosition` bound methods; fix the `OnBeforeClose` hook to save final window state.
3. **`frontend/src/services/wailsAPI.ts`** — Replace the two hardcoded stubs (`getWindowPersistence`, `setWindowPersistence`) with real calls to the new Go methods.

The config package already handles JSON read/write correctly. This phase is additive: no existing fields or methods change, only new ones are added. The Wails runtime APIs needed (`runtime.WindowGetSize`, `runtime.WindowSetSize`, `runtime.WindowGetPosition`, `runtime.WindowSetPosition`) are all verified available in Wails v2.10.2.

**Primary recommendation:** Add fields + methods to config first, then wire App layer and lifecycle hooks, then replace frontend stubs. Regenerate Wails bindings after every Go public method addition.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `encoding/json` | stdlib | Config JSON marshal/unmarshal | Already in use; no third-party lib needed |
| `os` / `path/filepath` | stdlib | Config file read/write | Already in use for config path resolution |
| `github.com/wailsapp/wails/v2/pkg/runtime` | v2.10.2 | Window size/position get+set, lifecycle hooks | Only Wails runtime API for window management |
| `context` | stdlib | Passed to lifecycle hooks (`OnDomReady`, `OnBeforeClose`) | Required by all Wails runtime calls |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `wails generate module` | CLI, v2.10.2 | Regenerate TypeScript bindings after Go changes | After every new public App method or struct field change |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib `encoding/json` | `github.com/spf13/viper` | Viper is heavyweight for a 6-field config; stdlib is sufficient |
| Storing window state in Go config | localStorage (frontend only) | Electron used two separate files (preferences.json + window-state.json); Go config consolidates both into one file — simpler and acceptable for v2 parity |

**Installation:** No new dependencies. All additions use stdlib and existing Wails runtime import.

---

## Architecture Patterns

### Recommended Project Structure (no changes to layout)

```
internal/config/
└── config.go          # ADD: WindowX, WindowY int + WindowPersistenceEnabled bool fields
                       # ADD: SetWindowPosition(x, y int) error method
                       # ADD: SetWindowPersistence(enabled bool) error method
                       # ADD: GetWindowPersistence() bool method

app.go                 # ADD: GetWindowPersistence() bool bound method
                       # ADD: SetWindowPersistence(enabled bool) error bound method
                       # ADD: GetWindowPosition() (int, int, error) bound method
                       # ADD: SetWindowPosition(x, y int) error bound method
                       # ADD: domReady(ctx context.Context) lifecycle hook
                       # ADD: beforeClose(ctx context.Context) bool lifecycle hook

main.go                # ADD: OnDomReady: app.domReady
                       # ADD: OnBeforeClose: app.beforeClose

frontend/src/services/wailsAPI.ts   # REPLACE stubs with real calls
frontend/wailsjs/go/main/App.js     # REGENERATED (auto) — do not edit
frontend/wailsjs/go/main/App.d.ts   # REGENERATED (auto) — do not edit
frontend/wailsjs/go/models.ts       # REGENERATED if Config struct changes
```

### Pattern 1: Config Struct Field Addition

**What:** Add fields to the existing `Config` struct and new setter methods.
**When to use:** Whenever new persistent state is needed at the Go layer.

```go
// Source: internal/config/config.go (existing pattern, extend it)

type Config struct {
    Theme                   string `json:"theme"`
    SidebarPosition         string `json:"sidebarPosition"`
    WindowWidth             int    `json:"windowWidth"`
    WindowHeight            int    `json:"windowHeight"`
    // ADD these three fields:
    WindowX                 int    `json:"windowX"`
    WindowY                 int    `json:"windowY"`
    WindowPersistenceEnabled bool  `json:"windowPersistenceEnabled"`
}

func DefaultConfig() *Config {
    return &Config{
        Theme:                    "light",
        SidebarPosition:          "left",
        WindowWidth:              1200,
        WindowHeight:             800,
        WindowX:                  0,
        WindowY:                  0,
        WindowPersistenceEnabled: true,  // default: persistence ON (Electron behavior)
    }
}
```

### Pattern 2: Wails Lifecycle Hooks for Window Restore

**What:** Wire `OnDomReady` to restore window state; wire `OnBeforeClose` to save final state.
**When to use:** Any window state that must survive process restart.

```go
// Source: Wails v2 docs — https://wails.io/docs/guides/application-development/
// OnDomReady is the ONLY safe place to call runtime.WindowSetSize/SetPosition.
// OnStartup is NOT safe — the window initializes on a separate thread.

// In app.go:
func (a *App) domReady(ctx context.Context) {
    cfg := a.configManager.Get()
    if cfg.WindowPersistenceEnabled {
        runtime.WindowSetSize(ctx, cfg.WindowWidth, cfg.WindowHeight)
        runtime.WindowSetPosition(ctx, cfg.WindowX, cfg.WindowY)
    }
}

func (a *App) beforeClose(ctx context.Context) bool {
    cfg := a.configManager.Get()
    if cfg.WindowPersistenceEnabled {
        w, h := runtime.WindowGetSize(ctx)
        x, y := runtime.WindowGetPosition(ctx)
        _ = a.configManager.SetWindowSize(w, h)
        _ = a.configManager.SetWindowPosition(x, y)
    }
    return false  // false = allow close; true = prevent close
}

// In main.go — register hooks:
err := wails.Run(&options.App{
    // ... existing fields ...
    OnStartup:   app.startup,
    OnDomReady:  app.domReady,    // ADD
    OnBeforeClose: app.beforeClose, // ADD
    Bind: []interface{}{app},
})
```

### Pattern 3: Initial Window Size from Config (Prevent Double-Resize)

**What:** Pass saved width/height to `options.App{}` so the window opens at the correct size immediately, before `OnDomReady` fires. Prevents a visible resize flash on startup.
**When to use:** When window size persistence is active.

```go
// In main.go — read config before wails.Run:
app := NewApp()
cfg := app.GetConfig()

startWidth := cfg.WindowWidth
startHeight := cfg.WindowHeight
if !cfg.WindowPersistenceEnabled || startWidth < 400 || startHeight < 300 {
    startWidth = 1024
    startHeight = 768
}

err := wails.Run(&options.App{
    Width:  startWidth,
    Height: startHeight,
    // ...
})
```

### Pattern 4: New App-Layer Bound Methods

**What:** Expose Go config methods to the frontend via new public App methods.
**When to use:** When frontend needs to read or write a config field.

```go
// In app.go — these become Wails-bound methods (auto-generates JS/TS bindings):

func (a *App) GetWindowPersistence() bool {
    if a.configManager == nil {
        return true
    }
    return a.configManager.Get().WindowPersistenceEnabled
}

func (a *App) SetWindowPersistence(enabled bool) error {
    if a.configManager == nil {
        return nil
    }
    return a.configManager.SetWindowPersistence(enabled)
}

func (a *App) GetWindowPosition() (int, int, error) {
    if a.configManager == nil {
        return 0, 0, nil
    }
    cfg := a.configManager.Get()
    return cfg.WindowX, cfg.WindowY, nil
}

func (a *App) SetWindowPosition(x, y int) error {
    if a.configManager == nil {
        return nil
    }
    return a.configManager.SetWindowPosition(x, y)
}
```

### Pattern 5: Frontend Stub Replacement

**What:** Replace hardcoded stubs in `wailsAPI.ts` with calls to the new real Go methods.
**When to use:** After bindings are regenerated and new methods are available.

```typescript
// Source: frontend/src/services/wailsAPI.ts
// Replace these stubs:
//   getWindowPersistence: async () => { return true; }
//   setWindowPersistence: async (enabled) => { return { success: true }; }
// With:
import { GetWindowPersistence, SetWindowPersistence } from '../../wailsjs/go/main/App';

getWindowPersistence: async () => {
  try {
    return await GetWindowPersistence();
  } catch (error) {
    return true; // default to enabled on error
  }
},

setWindowPersistence: async (enabled: boolean) => {
  try {
    await SetWindowPersistence(enabled);
    return { success: true };
  } catch (error: any) {
    return { success: false, error: error.message };
  }
},
```

### Anti-Patterns to Avoid

- **Calling `runtime.WindowSetSize` in `OnStartup`:** Causes panics on Windows, silent failure on macOS. Use `OnDomReady` only. (Pitfall 8, PITFALLS.md)
- **Ignoring macOS coordinate drift:** Wails v2.10.2 — verify PR #3479 is included. The position save/restore should still be implemented as-is for Windows/Linux parity; macOS drift (if present) is bounded and acceptable for v2.0.0 per STATE.md decision.
- **Not regenerating bindings after adding Go methods:** New methods `GetWindowPersistence`, `SetWindowPersistence`, `GetWindowPosition`, `SetWindowPosition` must appear in `App.js` and `App.d.ts` before the frontend can use them. Run `wails dev` or `wails generate module` immediately after Go changes.
- **Zero-value Manager silent failure:** The existing `NewApp()` fallback creates a `&config.Manager{}` with nil config and empty path. New methods on `Manager` must guard against `m.config == nil` — or better, the fallback should use `config.NewManagerWithDefaults()` that creates an in-memory-only Manager with DefaultConfig set.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Window size/position read at close | Custom goroutine + channel to capture final state | `OnBeforeClose` hook (Wails v2) | `OnBeforeClose` is the official Wails callback for pre-close work; runs on close button, OS quit, and programmatic close |
| Window restore on startup | Call `runtime.WindowSetSize` in `OnStartup` | `OnDomReady` hook | Runtime API is unsafe in `OnStartup` — documented Wails behavior |
| Cross-platform config path | Custom `~/.storcat` or hardcoded paths | `os.UserConfigDir()` (already used) | Produces correct OS-appropriate path without platform switches |
| Config debouncing | Custom timer in Go | Debounce in frontend JS before calling `setWindowSize` | Config I/O is fast at desktop scale; debounce in JS is simpler and keeps Go layer clean |

**Key insight:** All the hard problems (platform-appropriate config paths, lifecycle hooks, window runtime API) are already solved by Go stdlib + Wails. This phase is exclusively additive plumbing.

---

## Common Pitfalls

### Pitfall 1: OnDomReady Not Wired = Silent No-Op
**What goes wrong:** `domReady` function is implemented in `app.go` but not registered in `main.go` under `OnDomReady`. Window restore code runs nowhere. No error, no warning.
**Why it happens:** Easy to write the method and forget to wire it.
**How to avoid:** After writing `domReady`, immediately add `OnDomReady: app.domReady` to `main.go`'s `wails.Run` options.
**Warning signs:** App starts at hardcoded 1024x768 even after saving a different size.

### Pitfall 2: Stale Wails Bindings After Adding Go Methods
**What goes wrong:** `GetWindowPersistence` and `SetWindowPersistence` exist in Go but `App.js` / `App.d.ts` still show old methods. TypeScript import fails or resolves to `undefined`.
**Why it happens:** Bindings are auto-generated at `wails dev` start or explicit `wails generate module` — neither runs automatically when you save a `.go` file.
**How to avoid:** After adding new public App methods, run `wails dev` (which regenerates on startup) or `wails generate module` explicitly. Verify the new function names appear in `frontend/wailsjs/go/main/App.js`.
**Warning signs:** TypeScript compile error on `import { GetWindowPersistence }` or the function not present in `App.js`.

### Pitfall 3: `windowPersistenceEnabled` Default Should Be `true`
**What goes wrong:** If `DefaultConfig()` sets `WindowPersistenceEnabled: false`, users who have no existing config file get a disabled persistence experience. The Electron version had persistence enabled by default.
**Why it happens:** Boolean zero value in Go is `false`.
**How to avoid:** Explicitly set `WindowPersistenceEnabled: true` in `DefaultConfig()`.
**Warning signs:** First-run experience: window does not save size; setting toggle appears OFF on first launch.

### Pitfall 4: Position Restore Without Bounds Clamping (Multi-Monitor)
**What goes wrong:** Saved `WindowX`/`WindowY` coordinates refer to a monitor that no longer exists (unplugged second monitor, different workstation). Window restores off-screen. User cannot see the window.
**Why it happens:** Position coordinates are absolute screen coordinates; they become invalid when monitor topology changes.
**How to avoid:** After calling `runtime.WindowSetPosition`, no Wails v2 API exists to clamp to visible area. The safest mitigation: if `WindowX == 0 && WindowY == 0`, skip position restore (use OS default placement). For non-zero positions, restore as-is and accept the multi-monitor limitation for v2.0.0.
**Warning signs:** App appears to launch but window is invisible (off-screen).

### Pitfall 5: `beforeClose` Return Value Semantics
**What goes wrong:** `OnBeforeClose` expects a `func(ctx context.Context) bool`. Returning `true` prevents the window from closing. The intent is to save state then allow close — must return `false`.
**Why it happens:** Boolean semantics of "prevent close" are not obvious without reading the docs.
**How to avoid:** Always return `false` from `beforeClose` (save state and allow close). Only return `true` if implementing a "are you sure?" dialog.
**Warning signs:** App refuses to close; clicking X does nothing.

---

## Code Examples

Verified patterns from official sources:

### Wails Lifecycle Hook Registration
```go
// Source: https://wails.io/docs/reference/options/#ondomready
err := wails.Run(&options.App{
    OnStartup:    app.startup,
    OnDomReady:   app.domReady,
    OnBeforeClose: app.beforeClose,
    // ...
})
```

### runtime.WindowGetSize / WindowSetSize
```go
// Source: https://wails.io/docs/reference/runtime/window/
// Must be called from OnDomReady or a bound method, not OnStartup.
width, height := runtime.WindowGetSize(ctx)
runtime.WindowSetSize(ctx, 1200, 800)
```

### runtime.WindowGetPosition / WindowSetPosition
```go
// Source: https://wails.io/docs/reference/runtime/window/
x, y := runtime.WindowGetPosition(ctx)
runtime.WindowSetPosition(ctx, 100, 100)
```

### Config Manager SetWindowPosition (new method to add)
```go
// Pattern follows existing SetWindowSize in internal/config/config.go
func (m *Manager) SetWindowPosition(x, y int) error {
    m.config.WindowX = x
    m.config.WindowY = y
    return m.Save()
}

func (m *Manager) SetWindowPersistence(enabled bool) error {
    m.config.WindowPersistenceEnabled = enabled
    return m.Save()
}

func (m *Manager) GetWindowPersistence() bool {
    return m.config.WindowPersistenceEnabled
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Electron: two separate files (`preferences.json` + `window-state.json`) | Go: single `~/.config/storcat/config.json` | Phase 3 (this phase) | Simpler; acceptable for v2.0.0 |
| Stubs always returning `true` / `{success:true}` | Real Go config read/write | Phase 3 (this phase) | Settings toggle becomes functional |
| Window opens at hardcoded 1024x768 | Window opens at saved size from config | Phase 3 (this phase) | WIN-01 fulfilled |
| `OnDomReady` not wired | `OnDomReady` calls `runtime.WindowSetSize` + `WindowSetPosition` | Phase 3 (this phase) | WIN-01, WIN-02 fulfilled |

---

## Open Questions

1. **macOS coordinate drift (WIN-02)**
   - What we know: Wails issue #3478 documents drift; PR #3479 claims a fix. Current project uses v2.10.2.
   - What's unclear: Whether PR #3479 is included in v2.10.2 or a later patch. STATE.md notes position restore is deferred to "size-only first."
   - Recommendation: Implement full position save/restore for WIN-02 compliance. If drift is observed during verification, add a guard `if runtime.GOOS == "darwin" { skip position restore }`. Do NOT preemptively skip — verify behavior first.

2. **`OnBeforeClose` vs frontend `beforeunload`**
   - What we know: Wails v2 supports `OnBeforeClose` in Go. The frontend could also use `window.addEventListener('beforeunload', ...)` to call `setWindowSize` before close.
   - What's unclear: Which approach is more reliable for capturing the *final* window size just before close.
   - Recommendation: Use `OnBeforeClose` in Go — it calls `runtime.WindowGetSize` which reflects the actual native window size at that moment, bypassing any frontend state lag.

---

## Environment Availability

Step 2.6: SKIPPED — Phase 3 is purely Go source code and TypeScript changes with no new external dependencies. All required tools (`go`, `wails`, `npm`) were verified available in prior phase work.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | None — no test framework detected in project |
| Config file | None |
| Quick run command | Manual: `wails dev` → settings modal → toggle → verify config file on disk |
| Full suite command | Manual: full app launch cycle (close + reopen, verify window size restored) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| WIN-01 | Config file contains `windowWidth`/`windowHeight` after size change | smoke | Manual — inspect `~/.config/storcat/config.json` after resizing | ❌ Wave 0 — no test infra |
| WIN-02 | Config file contains `windowX`/`windowY` after move; values restore on relaunch | smoke | Manual — close app, reopen, verify window position | ❌ Wave 0 |
| WIN-03 | `GetWindowPersistence` and `SetWindowPersistence` round-trip; toggle writes to disk | smoke | Manual — toggle switch OFF in settings → check config.json → toggle ON → check config.json | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** Manual inspect `~/.config/storcat/config.json` after each relevant change
- **Per wave merge:** Full launch cycle — change size/position → close → reopen → verify restoration
- **Phase gate:** All three WIN-* behaviors verified before `/gsd:verify-work`

### Wave 0 Gaps

- No automated test framework exists for this project (TEST-01 through TEST-03 deferred per REQUIREMENTS.md)
- Manual verification protocol covers all three requirements — acceptable for v2.0.0 milestone

---

## Sources

### Primary (HIGH confidence)
- Wails v2 official docs — application lifecycle: https://wails.io/docs/guides/application-development/
- Wails v2 runtime window API: https://wails.io/docs/reference/runtime/window/
- Wails v2 options reference (OnDomReady, OnBeforeClose): https://wails.io/docs/reference/options/
- Direct source analysis: `/Users/ken/dev/storcat/internal/config/config.go` — existing Config struct and Manager methods
- Direct source analysis: `/Users/ken/dev/storcat/app.go` — existing App methods and nil guards
- Direct source analysis: `/Users/ken/dev/storcat/main.go` — existing wails.Run options (OnStartup only wired)
- Direct source analysis: `/Users/ken/dev/storcat/frontend/src/services/wailsAPI.ts` — confirmed stubs at lines 156-163
- Prior project research: `.planning/research/ARCHITECTURE.md` — window state persistence section, build order §4 and §6

### Secondary (MEDIUM confidence)
- Prior project research: `.planning/research/PITFALLS.md` — Pitfall 4 (macOS coordinate drift), Pitfall 8 (OnStartup timing)
- Prior project research: `.planning/research/STACK.md` — Wails runtime API table (WindowGetSize, WindowSetSize, WindowGetPosition, WindowSetPosition)
- Wails GitHub PR #3479 — macOS coordinate system fix: https://github.com/wailsapp/wails/pull/3479
- Wails GitHub issue #2739 — multi-monitor position: https://github.com/wailsapp/wails/issues/2739

### Tertiary (LOW confidence)
- None

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all additions use stdlib + existing Wails v2.10.2 runtime; no new dependencies
- Architecture: HIGH — existing config package pattern is clear and consistent; direct source code verified
- Pitfalls: HIGH — sourced from prior verified research (PITFALLS.md) and direct code inspection

**Research date:** 2026-03-25
**Valid until:** 2026-05-25 (stable Wails v2 API; no breaking changes expected before Wails v3 stable which has no ETA)
