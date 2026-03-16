package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-version", "version":
			fmt.Println(appVersion())
			return
		}
	}

	app := NewApp()
	err := wails.Run(&options.App{
		Title:            "bsky-browser",
		Width:            1024,
		Height:           768,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []any{app, app.authService, app.indexService, app.searchService, app.logService},
	})

	if err != nil {
		LogErrorf("Application error: %v", err)
		runtime.LogErrorf(app.ctx, "Application error: %v", err)
	}
}
