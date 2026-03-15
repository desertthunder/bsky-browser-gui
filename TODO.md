# Wails Desktop App — Roadmap

## Milestone 1 — Scaffold & Tooling

- [x] Install Tailwind v4 (`tailwindcss`, `@tailwindcss/vite`) and wire `vite.config.ts`
- [x] Install Fontsource packages (`@fontsource/jetbrains-mono`, `@fontsource-variable/geist`, `@fontsource-variable/lora`)
- [x] Create `frontend/src/index.css` with `@import "tailwindcss"` and `@theme` tokens (palette, fonts)
- [x] Put font CSS imports in `App.svelte`
- [x] Set up `Taskfile.yml` for build tasks
- [x] Verify `wails dev` hot-reloads a "Hello World" page with correct fonts and theme

## Milestone 2 — Database Layer

- [x] Implement `database.go` — `Open()`, `Close()`, embedded migrations via `//go:embed`
- [x] Copy existing migrations from CLI (`000`–`003`) and add `004_add_facets_column.sql` (`ALTER TABLE posts ADD COLUMN facets TEXT`)
- [x] Enable WAL mode (`PRAGMA journal_mode=WAL`) on connection open
- [x] Implement `models.go` — `Post`, `Auth`, `SearchResult` structs (add `Facets` field to `Post`)
- [x] Implement `PostExists`, `InsertPost`, `UpsertAuth`, `GetAuth`, `SearchPosts`, `CountPosts`
- [x] Verify the desktop app and CLI can read/write the same database concurrently

## Milestone 3 — Authentication

- [x] Implement `AuthService` struct with Wails service binding
- [x] `Login(handle)` — loopback OAuth via `oauth.NewLocalhostConfig`, persist tokens to shared DB
- [x] `Whoami(force)` — load auth from DB, optionally resolve handle from DID
- [x] `IsAuthenticated()` — check for valid auth row
- [x] Automatic token refresh on `OnStartup` lifecycle hook
- [x] Frontend: login view with handle input, "Login" button, and status display

## Milestone 4 — Indexing & Progress

- [x] Implement `IndexService` struct with Wails service binding
- [x] `Refresh(limit)` — concurrent bookmark + like fetch, batch insert (port `RefreshAndIndex` logic)
- [x] Populate `facets` column from `FeedPost.Facets` during `convertPostView`
- [x] `IsIndexing()` — thread-safe boolean guard to prevent concurrent refreshes
- [x] Emit Wails events: `index:started`, `index:progress`, `index:done`
- [x] Frontend: "Refresh" button in header, optional limit input
- [x] Frontend: bottom-pinned progress bar component driven by `index:*` events

## Milestone 5 — Search & Data Table

- [ ] Implement `SearchService` struct with Wails service binding
- [ ] `Search(query, source)` — FTS5 query with BM25 ranking and source filter
- [ ] `CountPosts()` — total indexed post count
- [ ] Frontend: search bar with query input and source filter (All / Saved / Liked segmented control)
- [ ] Frontend: tabbed data table component (Saved / Liked / All tabs)
- [ ] Columns: Author Handle, Text (truncated), Created At, ♥ Likes, 🔁 Reposts, 💬 Replies, Source
- [ ] Client-side column sorting (click header to toggle asc/desc)
- [ ] Row click → open post URL in default browser via `runtime.BrowserOpenURL`

## Milestone 6 — Facets & Log Viewer

- [ ] Frontend: facet parser — convert UTF-8 byte offsets to JS string indices
- [ ] Frontend: facet renderer — links (`<a>`), mentions (`@handle`), hashtags (`#tag`)
- [ ] Integrate rendered facets into post text in data table rows
- [ ] Implement `LogService` — custom `io.Writer` that emits `log:line` events
- [ ] Wire `LogService` writer into `log.NewWithOptions` alongside file writer
- [ ] Frontend: log viewer panel with terminal-style dark background, monospace text
- [ ] Auto-scroll to bottom with scroll-lock toggle
- [ ] Level filter buttons: Debug, Info, Warn, Error

## Milestone 7 — Polish

- [ ] Empty state: show "No posts indexed" with prompt to refresh
- [ ] Error handling: toast/notification for network failures, auth expiry
- [ ] Keyboard shortcuts: `Cmd+K` focus search, `Cmd+R` refresh, `Cmd+L` toggle log viewer
- [ ] Window title and app icon (`build/appicon.png`)
- [ ] Production build verification (`wails3 build` → macOS `.app` bundle)
- [ ] README with build instructions, screenshots, and usage
