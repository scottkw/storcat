package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
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
		Title:  "storcat-wails",
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
