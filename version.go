package main

// Version is injected at build time via:
//   wails build -ldflags "-X main.Version=2.0.0"
// Falls back to "dev" in development mode.
var Version = "dev"
