package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	authService *AuthService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		authService: NewAuthService(),
	}
}

// startup is called when the app starts.
//
// The context is saved so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dbPath := getDBPath()
	if err := Open(dbPath); err != nil {
		fmt.Printf("failed to open database: %v\n", err)
		return
	}

	if a.authService.IsAuthenticated() {
		if err := a.authService.RefreshSession(); err != nil {
			fmt.Printf("token refresh failed on startup: %v\n", err)
		}
	}
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	runtime.LogInfo(ctx, "Shutting down")
	if err := Close(); err != nil {
		runtime.LogErrorf(ctx, "failed to close database: %v", err)
	}
}

// getDBPath returns the path to the shared database
func getDBPath() string {
	if dir := os.Getenv("BSKY_BROWSER_DATA"); dir != "" {
		return filepath.Join(dir, "bsky-browser.db")
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}

	return filepath.Join(configDir, "bsky-browser", "bsky-browser.db")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
