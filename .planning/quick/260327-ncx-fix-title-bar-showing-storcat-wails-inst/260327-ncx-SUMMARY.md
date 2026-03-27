# Quick Task 260327-ncx: Fix title bar and version number

**Date:** 2026-03-27
**Status:** Complete

## Changes

1. **main.go:62** — Changed window title from `"storcat-wails"` to `"StorCat"`
2. **wails.json:16** — Updated `productVersion` from `"2.1.0"` to `"2.2.0"`
3. **frontend/index.html:6** — Changed `<title>` from `"storcat-wails"` to `"StorCat"`
4. **app_test.go:134** — Updated test comment to reference `"2.2.0"`

## Verification

- All Go tests pass (`go test ./...`)
