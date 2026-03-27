package main

import (
	"embed"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"storcat-wails/cli"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Filter macOS Gatekeeper -psn_* args before any inspection
	args := filterMacOSArgs(os.Args[1:])

	// CLI dispatch: known subcommand -> CLI mode, exit before GUI
	if len(args) > 0 {
		switch args[0] {
		case "version", "create", "search", "list", "show", "open", "help", "--help", "-h":
			os.Exit(cli.Run(args, Version))
		}
	}

	// GUI mode: no args, or unrecognized args fall through
	runGUI()
}

// filterMacOSArgs removes macOS Gatekeeper -psn_* arguments injected on first Finder launch.
// Applied on all platforms (no-op on Windows/Linux).
func filterMacOSArgs(args []string) []string {
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-psn_") {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

// runGUI starts the Wails GUI application.
func runGUI() {
	// Create an instance of the app structure
	app := NewApp()

	// Read config for initial window size (prevents resize flash)
	cfg := app.GetConfig()
	startWidth := cfg.WindowWidth
	startHeight := cfg.WindowHeight
	if !cfg.WindowPersistenceEnabled || startWidth < 400 || startHeight < 300 {
		startWidth = 1024
		startHeight = 768
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "StorCat",
		Width:  startWidth,
		Height: startHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
