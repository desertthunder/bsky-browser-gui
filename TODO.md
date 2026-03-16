# v0.2.0 Roadmap

## Milestone 1 — Multi-Account Database Architecture

- [ ] Create `DatabaseManager` struct to replace global `db *sql.DB` (holds shared DB + per-account DB)
- [ ] Implement per-account database path resolution: `<config_dir>/bsky-browser/accounts/<DID>/bsky-browser.db`
- [ ] Migrate auth queries (`GetAuth`, `UpsertAuth`, `GetAuthByDID`) to use shared DB connection
- [ ] Migrate post queries (`InsertPost`, `SearchPosts`, `CountPosts`, `PostExists`) to use per-account DB connection
- [ ] Update `Open()` to initialize both shared and per-account databases with migrations
- [ ] Update `app.startup()` to auto-select the most recently updated account and open its database
- [ ] Add migration `001_add_active_account.sql` for tracking active account in shared DB
- [ ] Tests: `DatabaseManager` open/close, account switching with multiple temp databases

## Milestone 2 — Account Switcher Service & UI

- [ ] Implement `AccountService` struct with Wails binding
- [ ] `ListAccounts()` — query all rows from shared `auth` table
- [ ] `SwitchAccount(did)` — close current per-account DB, open target, emit `account:switched` event
- [ ] `RemoveAccount(did)` — delete auth row and account database directory
- [ ] `GetActiveAccount()` — return current account info
- [ ] Update `AuthService.Login()` to create per-account DB directory on new login
- [ ] Frontend: account switcher dropdown in header with handle list
- [ ] Frontend: "Add Account" and "Remove Account" actions
- [ ] Frontend: listen for `account:switched` to reset search/post state
- [ ] Tests: `ListAccounts`, `SwitchAccount`, `RemoveAccount` with temp dirs

## Milestone 3 — Profile View & Author Feed

- [ ] Add `ProfileView` model struct to `models.go`
- [ ] Implement `ProfileService` with Wails binding
- [ ] `GetProfile()` — call `bsky.ActorGetProfile` with active account DID
- [ ] `GetAuthorFeed(filter, cursor, limit)` — call `bsky.FeedGetAuthorFeed` with pagination
- [ ] Frontend: `ProfileView.svelte` component with profile card (avatar, name, bio, stats)
- [ ] Frontend: tabbed author feed below profile card (Posts / Replies / Media filters)
- [ ] Frontend: infinite scroll pagination using cursor
- [ ] Frontend: navigate to profile view from header handle/avatar click

## Milestone 4 — Search Own Posts

- [ ] Extend `SearchPosts()` with optional `authorDID` parameter
- [ ] Update FTS5 query to add `AND p.author_did = ?` when author filter is active
- [ ] Update `SearchService.Search()` to accept the author filter
- [ ] Frontend: "My Posts" toggle/tab in SearchBar that constrains to active account DID
- [ ] Tests: `SearchPosts` with author filter, mixed-author datasets

## Milestone 5 — Post Composer

- [ ] Implement `PostService` struct with Wails binding
- [ ] `CreatePost(text, replyTo)` — build `bsky.FeedPost` record, call `atproto.RepoCreateRecord`
- [ ] Facet detection for mentions, links, and hashtags
- [ ] Character limit enforcement (300 grapheme clusters)
- [ ] Frontend: compose button in header
- [ ] Frontend: compose modal with textarea, character counter, live facet preview
- [ ] Frontend: optional reply-to context from PostDetailPanel

## Milestone 6 — Drafts

- [ ] Add `drafts` table migration in per-account database
- [ ] Implement `PostService.SaveDraft()`, `ListDrafts()`, `GetDraft()`, `DeleteDraft()`
- [ ] `PublishDraft(id)` — create post and delete draft atomically
- [ ] Frontend: "Save Draft" button in compose modal
- [ ] Frontend: drafts panel with list view, click-to-edit, and delete
- [ ] Frontend: draft count badge on drafts button
- [ ] Tests: draft CRUD operations and publish flow

## Milestone 7 — System Tray

- [ ] Add `ra1phdd/systray-on-wails` dependency
- [ ] Create tray icon asset (`build/tray_icon.png`, 22×22 monochrome)
- [ ] Initialize system tray in `app.startup()` with menu items
- [ ] Menu items: active account label, Quick Post, Refresh Posts, Show/Hide Window, Quit
- [ ] Quick Post: focus window and open compose modal via Wails event
- [ ] Refresh Posts: trigger `IndexService.Refresh(0)` from tray

## Milestone 8 — Go Test Coverage

- [ ] `auth_service_test.go` — `IsAuthenticated`, `Whoami` with mocked DB state
- [ ] `index_service_test.go` — `batchWriter`, `convertPostView`, `extractFacets`, `parsePostRecord`
- [ ] `post_service_test.go` — draft CRUD, `CreatePost` record assembly (no network)
- [ ] `account_service_test.go` — multi-account switching with temp directories
- [ ] Expand `database_test.go` — `InsertPost` upsert, `CountPosts`, multi-account `UpsertAuth`, FTS5 ranking edge cases
- [x] Establish CI-friendly test target in `Taskfile.yml` (`task test`)

## Milestone 9 — Polish & Integration

- [ ] End-to-end smoke test: login → switch account → compose → post → verify in feed
- [ ] Error handling: toast notifications for tray actions, compose failures, account switch errors
- [ ] Loading states for profile view, author feed, and compose
- [ ] Keyboard shortcut: `Cmd+N` open compose modal
- [ ] Update README with Phase 2 features and screenshots
