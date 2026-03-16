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
	ctx           context.Context
	authService   *AuthService
	indexService  *IndexService
	searchService *SearchService
	logService    *LogService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		authService:   NewAuthService(),
		indexService:  NewIndexService(),
		searchService: NewSearchService(),
		logService:    NewLogService(),
	}
}

// startup is called when the app starts. The context is saved so we can call
// the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.authService.setContext(ctx)
	a.indexService.setContext(ctx)
	a.logService.setContext(ctx)

	if err := a.logService.Initialize(); err != nil {
		runtime.LogErrorf(a.ctx, "failed to initialize log service: %v", err)
	} else {
		InitLogger(a.logService)
		LogInfo("Application started")
	}

	dbPath := getDBPath()
	if err := Open(dbPath); err != nil {
		LogErrorf("failed to open database: %v", err)
		runtime.LogErrorf(a.ctx, "failed to open database: %v", err)
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Fatal Error",
			Message: fmt.Sprintf("Failed to open database: %v\n\nThe application will now exit.", err),
		})
		os.Exit(1)
	}

	if a.authService.IsAuthenticated() {
		if err := a.authService.RefreshSession(); err != nil {
			LogWarnf("token refresh failed on startup: %v", err)
			runtime.LogWarningf(a.ctx, "token refresh failed on startup: %v", err)
		}
	}
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	if err := Close(); err != nil {
		LogErrorf("failed to close database: %v", err)
		runtime.LogErrorf(ctx, "failed to close database: %v", err)
	}
	if err := a.logService.Close(); err != nil {
		runtime.LogErrorf(ctx, "failed to close log service: %v", err)
	}
}

// GetVersion returns the current application build version.
func (a *App) GetVersion() string {
	return appVersion()
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
