package main

import (
	_ "embed"
	"encoding/json"
)

//go:embed wails.json
var wailsJSON []byte

// Version returns the productVersion from wails.json, or "dev" if parsing fails.
var Version = func() string {
	var cfg struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(wailsJSON, &cfg); err != nil || cfg.Info.ProductVersion == "" {
		return "dev"
	}
	return cfg.Info.ProductVersion
}()
