# Changelog

## 2026-03-15

### Added

- Persistence via SQLite with WAL mode and FTS5 search
- Loopback OAuth flow, token persistence, automatic refresh on startup
- Concurrent bookmark and like fetching, batch insert, facet population,
  real-time progress bar via Wails events
- FTS5 full-text search with BM25 ranking, source filtering (All/Saved/Liked),
  sortable columns, and row click to open in browser
- Real-time log panel with terminal styling, auto-scroll with scroll-lock toggle,
  and level filtering (Debug/Info/Warn/Error)
- Keyboard Shortcuts -> `Cmd+K` focus search, `Cmd+R` refresh, `Cmd+L` toggle
  log viewer
