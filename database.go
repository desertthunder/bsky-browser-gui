// Database access and migrations
//
// NOTE: migrations should be "registered" in the [`runMigrations`] function and
// should be idempotent.
package main

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens the database connection and runs migrations
func Open(dbPath string) error {
	LogInfof("opening database: %s", dbPath)

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return wrapDBError("failed to create database directory", err)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return wrapDBError("failed to open database", err)
	}

	if err := db.Ping(); err != nil {
		return wrapDBError("failed to ping database", err)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return wrapDBError("failed to enable WAL mode", err)
	}

	LogInfo("database connection established with WAL mode")

	if err := runMigrations(); err != nil {
		return wrapDBError("failed to run migrations", err)
	}

	LogInfo("database migrations completed successfully")
	return nil
}

func runMigrations() error {
	migrations := []string{
		"migrations/000_initial_schema.sql",
		"migrations/001_add_indexes.sql",
		"migrations/002_add_oauth_callback_url.sql",
	}

	for _, migration := range migrations {
		if migration == "migrations/002_add_oauth_callback_url.sql" {
			hasColumn, err := columnExists("auth", "oauth_callback_url")
			if err != nil {
				return wrapDBError("failed to inspect auth.oauth_callback_url before migration", err)
			}
			if hasColumn {
				LogInfof("migration skipped because schema is already up to date: %s", migration)
				continue
			}
		}

		content, err := migrationsFS.ReadFile(migration)
		if err != nil {
			return wrapDBError("failed to read migration "+migration, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return wrapDBError("failed to execute migration "+migration, err)
		}
		LogInfof("migration applied successfully: %s", migration)
	}

	if err := ensureSchemaCompatibility(); err != nil {
		return wrapDBError("failed to upgrade schema", err)
	}

	return nil
}

// validSchemaIdentifiers is the allow-list of valid table and column names
var validSchemaIdentifiers = map[string]map[string]bool{
	"posts": {
		"facets": true,
	},
	"auth": {
		"session_id":                      true,
		"auth_server_url":                 true,
		"auth_server_token_endpoint":      true,
		"auth_server_revocation_endpoint": true,
		"oauth_callback_url":              true,
		"dpop_auth_nonce":                 true,
		"dpop_host_nonce":                 true,
		"dpop_private_key":                true,
	},
}

// validIndexIdentifiers is the allow-list of valid index names
var validIndexIdentifiers = map[string]bool{
	"idx_posts_author_did": true,
	"idx_posts_created_at": true,
	"idx_posts_source":     true,
}

func ensureSchemaCompatibility() error {
	columnsByTable := map[string][]struct {
		name       string
		definition string
	}{
		"posts": {
			{name: "facets", definition: "TEXT"},
		},
		"auth": {
			{name: "session_id", definition: "TEXT"},
			{name: "auth_server_url", definition: "TEXT"},
			{name: "auth_server_token_endpoint", definition: "TEXT"},
			{name: "auth_server_revocation_endpoint", definition: "TEXT"},
			{name: "oauth_callback_url", definition: "TEXT"},
			{name: "dpop_auth_nonce", definition: "TEXT"},
			{name: "dpop_host_nonce", definition: "TEXT"},
			{name: "dpop_private_key", definition: "TEXT"},
		},
	}

	for table, columns := range columnsByTable {
		// Validate table name against allow-list
		if _, ok := validSchemaIdentifiers[table]; !ok {
			return fmt.Errorf("invalid table name in schema migration: %s", table)
		}
		exists, err := tableExists(table)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}

		for _, column := range columns {
			// Validate column name against allow-list
			if !validSchemaIdentifiers[table][column.name] {
				return fmt.Errorf("invalid column name in schema migration: %s.%s", table, column.name)
			}

			hasColumn, err := columnExists(table, column.name)
			if err != nil {
				return err
			}
			if hasColumn {
				continue
			}

			query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column.name, column.definition)
			if _, err := db.Exec(query); err != nil {
				return wrapDBError("failed to add "+table+"."+column.name, err)
			}
		}
	}

	// Create performance indexes if they don't exist
	indexes := []struct {
		name    string
		table   string
		columns string
	}{
		{"idx_posts_author_did", "posts", "author_did"},
		{"idx_posts_created_at", "posts", "created_at"},
		{"idx_posts_source", "posts", "source"},
	}

	for _, idx := range indexes {
		// Validate index name against allow-list
		if !validIndexIdentifiers[idx.name] {
			return fmt.Errorf("invalid index name in schema migration: %s", idx.name)
		}
		exists, err := indexExists(idx.name)
		if err != nil {
			return err
		}
		if !exists {
			query := fmt.Sprintf("CREATE INDEX %s ON %s(%s)", idx.name, idx.table, idx.columns)
			if _, err := db.Exec(query); err != nil {
				return wrapDBError("failed to create index "+idx.name, err)
			}
		}
	}

	return nil
}

func indexExists(name string) (bool, error) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?", name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func tableExists(table string) (bool, error) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func columnExists(table, column string) (bool, error) {
	rows, err := db.Query("SELECT name FROM pragma_table_info(?)", table)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return false, err
		}

		if name == column {
			return true, nil
		}
	}

	return false, rows.Err()
}

// Close closes the database connection
func Close() error {
	LogInfo("closing database connection")
	if db != nil {
		err := db.Close()
		if err != nil {
			LogErrorf("failed to close database: %v", err)
			return err
		}
		LogInfo("database connection closed")
	}
	return nil
}

// PostExists checks if a post with the given URI already exists in the database
func PostExists(uri string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE uri = ?)", uri).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// InsertPost inserts a post into the database
func InsertPost(post *Post) error {
	LogInfof("inserting post: %s by %s", post.URI, post.AuthorHandle)

	query := `
		INSERT INTO posts (uri, cid, author_did, author_handle, text, created_at, like_count, repost_count, reply_count, source, facets)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uri) DO UPDATE SET
			cid = excluded.cid,
			author_did = excluded.author_did,
			author_handle = excluded.author_handle,
			text = excluded.text,
			created_at = excluded.created_at,
			like_count = excluded.like_count,
			repost_count = excluded.repost_count,
			reply_count = excluded.reply_count,
			source = excluded.source,
			facets = excluded.facets,
			indexed_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query,
		post.URI,
		post.CID,
		post.AuthorDID,
		post.AuthorHandle,
		post.Text,
		post.CreatedAt,
		post.LikeCount,
		post.RepostCount,
		post.ReplyCount,
		post.Source,
		post.Facets,
	)

	if err != nil {
		LogErrorf("failed to insert post: %s, error: %v", post.URI, err)
	}

	return err
}

// insertPostTx inserts a post using an existing transaction
func insertPostTx(tx *sql.Tx, post *Post) error {
	query := `
		INSERT INTO posts (uri, cid, author_did, author_handle, text, created_at, like_count, repost_count, reply_count, source, facets)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uri) DO UPDATE SET
			cid = excluded.cid,
			author_did = excluded.author_did,
			author_handle = excluded.author_handle,
			text = excluded.text,
			created_at = excluded.created_at,
			like_count = excluded.like_count,
			repost_count = excluded.repost_count,
			reply_count = excluded.reply_count,
			source = excluded.source,
			facets = excluded.facets,
			indexed_at = CURRENT_TIMESTAMP
	`

	_, err := tx.Exec(query,
		post.URI,
		post.CID,
		post.AuthorDID,
		post.AuthorHandle,
		post.Text,
		post.CreatedAt,
		post.LikeCount,
		post.RepostCount,
		post.ReplyCount,
		post.Source,
		post.Facets,
	)

	return err
}

// UpsertAuth inserts or updates auth information
func UpsertAuth(auth *Auth) error {
	LogInfof("upserting auth: %s (%s)", auth.DID, auth.Handle)

	query := `
		INSERT INTO auth (did, handle, access_jwt, refresh_jwt, pds_url, session_id,
						  auth_server_url, auth_server_token_endpoint, auth_server_revocation_endpoint, oauth_callback_url,
						  dpop_auth_nonce, dpop_host_nonce, dpop_private_key, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(did) DO UPDATE SET
			handle = excluded.handle,
			access_jwt = excluded.access_jwt,
			refresh_jwt = excluded.refresh_jwt,
			pds_url = excluded.pds_url,
			session_id = excluded.session_id,
			auth_server_url = excluded.auth_server_url,
			auth_server_token_endpoint = excluded.auth_server_token_endpoint,
			auth_server_revocation_endpoint = excluded.auth_server_revocation_endpoint,
			oauth_callback_url = excluded.oauth_callback_url,
			dpop_auth_nonce = excluded.dpop_auth_nonce,
			dpop_host_nonce = excluded.dpop_host_nonce,
			dpop_private_key = excluded.dpop_private_key,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query,
		auth.DID,
		auth.Handle,
		auth.AccessJWT,
		auth.RefreshJWT,
		auth.PDSURL,
		auth.SessionID,
		auth.AuthServerURL,
		auth.AuthServerTokenEndpoint,
		auth.AuthServerRevocationEndpoint,
		auth.OAuthCallbackURL,
		auth.DPoPAuthNonce,
		auth.DPoPHostNonce,
		auth.DPoPPrivateKey,
	)

	if err != nil {
		LogErrorf("failed to upsert auth: %s, error: %v", auth.DID, err)
	}

	return err
}

// ClearAuth removes all persisted auth rows for this desktop client.
func ClearAuth() error {
	_, err := db.Exec("DELETE FROM auth")
	return err
}

// GetAuth loads the auth record from the database
func GetAuth() (*Auth, error) {
	LogInfo("loading auth from database")

	query := `SELECT did, handle, access_jwt, refresh_jwt, pds_url, session_id,
			  auth_server_url, auth_server_token_endpoint, auth_server_revocation_endpoint, oauth_callback_url,
			  dpop_auth_nonce, dpop_host_nonce, dpop_private_key, updated_at
			  FROM auth
			  ORDER BY updated_at DESC
			  LIMIT 1`

	auth, err := getAuthByQuery(query)

	if err == sql.ErrNoRows {
		LogInfo("no auth record found in database")
		return nil, nil
	}
	if err != nil {
		LogErrorf("failed to load auth: %v", err)
		return nil, err
	}

	LogInfof("auth loaded successfully: %s (%s)", auth.DID, auth.Handle)
	return auth, nil
}

// GetAuthByDID loads auth for a specific DID.
func GetAuthByDID(did string) (*Auth, error) {
	query := `SELECT did, handle, access_jwt, refresh_jwt, pds_url, session_id,
			  auth_server_url, auth_server_token_endpoint, auth_server_revocation_endpoint, oauth_callback_url,
			  dpop_auth_nonce, dpop_host_nonce, dpop_private_key, updated_at
			  FROM auth
			  WHERE did = ?
			  LIMIT 1`

	auth, err := getAuthByQuery(query, did)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return auth, nil
}

func getAuthByQuery(query string, args ...any) (*Auth, error) {
	var auth Auth
	var updatedAt string

	var sessionID, authServerURL, authServerTokenEndpoint, authServerRevocationEndpoint, oauthCallbackURL, dpopAuthNonce, dpopHostNonce, dpopPrivateKey sql.NullString

	err := db.QueryRow(query, args...).Scan(
		&auth.DID,
		&auth.Handle,
		&auth.AccessJWT,
		&auth.RefreshJWT,
		&auth.PDSURL,
		&sessionID,
		&authServerURL,
		&authServerTokenEndpoint,
		&authServerRevocationEndpoint,
		&oauthCallbackURL,
		&dpopAuthNonce,
		&dpopHostNonce,
		&dpopPrivateKey,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if sessionID.Valid {
		auth.SessionID = sessionID.String
	}
	if authServerURL.Valid {
		auth.AuthServerURL = authServerURL.String
	}
	if authServerTokenEndpoint.Valid {
		auth.AuthServerTokenEndpoint = authServerTokenEndpoint.String
	}
	if authServerRevocationEndpoint.Valid {
		auth.AuthServerRevocationEndpoint = authServerRevocationEndpoint.String
	}
	if oauthCallbackURL.Valid {
		auth.OAuthCallbackURL = oauthCallbackURL.String
	}
	if dpopAuthNonce.Valid {
		auth.DPoPAuthNonce = dpopAuthNonce.String
	}
	if dpopHostNonce.Valid {
		auth.DPoPHostNonce = dpopHostNonce.String
	}
	if dpopPrivateKey.Valid {
		auth.DPoPPrivateKey = dpopPrivateKey.String
	}

	auth.UpdatedAt = parseStoredTime(updatedAt)
	return &auth, nil
}

// validSortColumns defines which columns can be used for sorting
var validSortColumns = map[string]bool{
	"created_at":    true,
	"indexed_at":    true,
	"like_count":    true,
	"repost_count":  true,
	"reply_count":   true,
	"author_handle": true,
	"text":          true,
}

// SearchPosts searches posts using FTS5 with server-side pagination.
func SearchPosts(query string, source string, page int, pageSize int, sortColumn string, sortDirection string) (SearchPage, error) {
	const defaultPageSize = 25
	const maxPageSize = 100

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset := (page - 1) * pageSize

	// Validate and default sort parameters
	if sortColumn == "" || !validSortColumns[sortColumn] {
		sortColumn = "created_at"
	}
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	query = strings.TrimSpace(query)
	if query == "*" {
		query = ""
	}

	LogInfof(
		"searching posts: query=%s, source=%s, page=%d, pageSize=%d, sort=%s %s",
		query,
		source,
		page,
		pageSize,
		sortColumn,
		sortDirection,
	)

	if query == "" {
		return listRecentPosts(source, page, pageSize, offset, sortColumn, sortDirection)
	}

	total, err := countFilteredPosts(query, source)
	if err != nil {
		return SearchPage{}, err
	}

	sqlQuery := `
		SELECT p.uri, p.cid, p.author_did, p.author_handle, p.text, p.created_at,
			   p.like_count, p.repost_count, p.reply_count, p.source, p.indexed_at,
			   bm25(posts_fts, 5.0, 1.0) AS rank
		FROM posts_fts
		JOIN posts p ON posts_fts.rowid = p.rowid
		WHERE posts_fts MATCH ?
		  AND (? = '' OR p.source = ?)
		ORDER BY p.` + sortColumn + ` ` + sortDirection + `
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(sqlQuery, query, source, source, pageSize, offset)
	if err != nil {
		LogErrorf("failed to execute search query: %v", err)
		return SearchPage{}, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var createdAt, indexedAt string

		err := rows.Scan(
			&r.URI,
			&r.CID,
			&r.AuthorDID,
			&r.AuthorHandle,
			&r.Text,
			&createdAt,
			&r.LikeCount,
			&r.RepostCount,
			&r.ReplyCount,
			&r.Source,
			&indexedAt,
			&r.Rank,
		)
		if err != nil {
			return SearchPage{}, err
		}

		r.CreatedAt = parseStoredTime(createdAt)
		r.IndexedAt = parseStoredTime(indexedAt)
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return SearchPage{}, err
	}

	LogInfof("search completed: %d results (total=%d)", len(results), total)
	return SearchPage{
		Results:  results,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func listRecentPosts(source string, page int, pageSize int, offset int, sortColumn string, sortDirection string) (SearchPage, error) {
	total, err := countRecentPosts(source)
	if err != nil {
		return SearchPage{}, err
	}

	// Build query with proper ORDER BY - note: sortColumn and sortDirection are validated by caller
	query := fmt.Sprintf(`
		SELECT uri, cid, author_did, author_handle, text, created_at,
			   like_count, repost_count, reply_count, source, indexed_at
		FROM posts
		WHERE (? = '' OR source = ?)
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, sortColumn, sortDirection)

	rows, err := db.Query(query, source, source, pageSize, offset)
	if err != nil {
		LogErrorf("failed to list recent posts: %v", err)
		return SearchPage{}, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var createdAt, indexedAt string

		err := rows.Scan(
			&r.URI,
			&r.CID,
			&r.AuthorDID,
			&r.AuthorHandle,
			&r.Text,
			&createdAt,
			&r.LikeCount,
			&r.RepostCount,
			&r.ReplyCount,
			&r.Source,
			&indexedAt,
		)
		if err != nil {
			return SearchPage{}, err
		}

		r.CreatedAt = parseStoredTime(createdAt)
		r.IndexedAt = parseStoredTime(indexedAt)
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return SearchPage{}, err
	}

	LogInfof("browse completed: %d results (total=%d)", len(results), total)
	return SearchPage{
		Results:  results,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func countRecentPosts(source string) (int, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM posts
		WHERE (? = '' OR source = ?)
	`

	var total int
	if err := db.QueryRow(countQuery, source, source).Scan(&total); err != nil {
		LogErrorf("failed to count recent posts: %v", err)
		return 0, err
	}

	return total, nil
}

func countFilteredPosts(query string, source string) (int, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM posts_fts
		JOIN posts p ON posts_fts.rowid = p.rowid
		WHERE posts_fts MATCH ?
		  AND (? = '' OR p.source = ?)
	`

	var total int
	if err := db.QueryRow(countQuery, query, source, source).Scan(&total); err != nil {
		LogErrorf("failed to count filtered posts: %v", err)
		return 0, err
	}

	return total, nil
}

// parseStoredTimeCache caches the last successful time layout for faster parsing
var (
	parseStoredTimeCache    string
	parseStoredTimeCacheMux sync.RWMutex
	parseStoredTimeLayouts  = []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05",
	}
)

func parseStoredTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	// Try cached layout first
	parseStoredTimeCacheMux.RLock()
	cachedLayout := parseStoredTimeCache
	parseStoredTimeCacheMux.RUnlock()

	if cachedLayout != "" {
		if parsed, err := time.Parse(cachedLayout, value); err == nil {
			return parsed
		}
	}

	// Fall back to trying all layouts
	for _, layout := range parseStoredTimeLayouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			// Cache the successful layout
			parseStoredTimeCacheMux.Lock()
			parseStoredTimeCache = layout
			parseStoredTimeCacheMux.Unlock()
			return parsed
		}
	}

	return time.Time{}
}

// CountPosts returns the total number of posts in the database
func CountPosts() (int, error) {
	LogInfo("counting posts in database")

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		LogErrorf("failed to count posts: %v", err)
		return 0, err
	}

	LogInfof("post count: %d", count)
	return count, nil
}

func wrapDBError(message string, err error) error {
	return &dbError{message: message, err: err}
}

type dbError struct {
	message string
	err     error
}

func (e *dbError) Error() string {
	return e.message + ": " + e.err.Error()
}

func (e *dbError) Unwrap() error {
	return e.err
}
