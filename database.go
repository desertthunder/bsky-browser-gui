package main

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens the database connection and runs migrations
func Open(dbPath string) error {
	fmt.Printf("opening database: %s\n", dbPath)

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	fmt.Println("database connection established with WAL mode")

	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("database migrations completed successfully")
	return nil
}

func runMigrations() error {
	content, err := migrationsFS.ReadFile("migrations/000_initial_schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration: %w", err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	fmt.Println("closing database connection")
	if db != nil {
		err := db.Close()
		if err != nil {
			fmt.Printf("failed to close database: %v\n", err)
			return err
		}
		fmt.Println("database connection closed")
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
	fmt.Printf("inserting post: %s by %s\n", post.URI, post.AuthorHandle)

	exists, err := PostExists(post.URI)
	if err != nil {
		fmt.Printf("failed to check if post exists: %s, error: %v\n", post.URI, err)
		return err
	}

	if exists {
		fmt.Printf("skipping already indexed post: %s\n", post.URI)
		return nil
	}

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

	_, err = db.Exec(query,
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
		fmt.Printf("failed to insert post: %s, error: %v\n", post.URI, err)
	}

	return err
}

// UpsertAuth inserts or updates auth information
func UpsertAuth(auth *Auth) error {
	fmt.Printf("upserting auth: %s (%s)\n", auth.DID, auth.Handle)

	query := `
		INSERT INTO auth (did, handle, access_jwt, refresh_jwt, pds_url, session_id,
						  auth_server_url, auth_server_token_endpoint, auth_server_revocation_endpoint,
						  dpop_auth_nonce, dpop_host_nonce, dpop_private_key, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(did) DO UPDATE SET
			handle = excluded.handle,
			access_jwt = excluded.access_jwt,
			refresh_jwt = excluded.refresh_jwt,
			pds_url = excluded.pds_url,
			session_id = excluded.session_id,
			auth_server_url = excluded.auth_server_url,
			auth_server_token_endpoint = excluded.auth_server_token_endpoint,
			auth_server_revocation_endpoint = excluded.auth_server_revocation_endpoint,
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
		auth.DPoPAuthNonce,
		auth.DPoPHostNonce,
		auth.DPoPPrivateKey,
	)

	if err != nil {
		fmt.Printf("failed to upsert auth: %s, error: %v\n", auth.DID, err)
	}

	return err
}

// GetAuth loads the auth record from the database
func GetAuth() (*Auth, error) {
	fmt.Println("loading auth from database")

	query := `SELECT did, handle, access_jwt, refresh_jwt, pds_url, session_id,
			  auth_server_url, auth_server_token_endpoint, auth_server_revocation_endpoint,
			  dpop_auth_nonce, dpop_host_nonce, dpop_private_key, updated_at
			  FROM auth
			  ORDER BY updated_at DESC
			  LIMIT 1`

	auth, err := getAuthByQuery(query)

	if err == sql.ErrNoRows {
		fmt.Println("no auth record found in database")
		return nil, nil
	}
	if err != nil {
		fmt.Printf("failed to load auth: %v\n", err)
		return nil, err
	}

	fmt.Printf("auth loaded successfully: %s (%s)\n", auth.DID, auth.Handle)
	return auth, nil
}

// GetAuthByDID loads auth for a specific DID.
func GetAuthByDID(did string) (*Auth, error) {
	query := `SELECT did, handle, access_jwt, refresh_jwt, pds_url, session_id,
			  auth_server_url, auth_server_token_endpoint, auth_server_revocation_endpoint,
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

	var sessionID, authServerURL, authServerTokenEndpoint, authServerRevocationEndpoint, dpopAuthNonce, dpopHostNonce, dpopPrivateKey sql.NullString

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
	if dpopAuthNonce.Valid {
		auth.DPoPAuthNonce = dpopAuthNonce.String
	}
	if dpopHostNonce.Valid {
		auth.DPoPHostNonce = dpopHostNonce.String
	}
	if dpopPrivateKey.Valid {
		auth.DPoPPrivateKey = dpopPrivateKey.String
	}

	auth.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &auth, nil
}

// SearchPosts searches posts using FTS5
func SearchPosts(query string, source string) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "*" {
		query = ""
	}

	fmt.Printf("searching posts: query=%s, source=%s\n", query, source)

	if query == "" {
		return listRecentPosts(source)
	}

	sqlQuery := `
		SELECT p.uri, p.cid, p.author_did, p.author_handle, p.text, p.created_at,
			   p.like_count, p.repost_count, p.reply_count, p.source, p.indexed_at,
			   bm25(posts_fts, 5.0, 1.0) AS rank
		FROM posts_fts
		JOIN posts p ON posts_fts.rowid = p.rowid
		WHERE posts_fts MATCH ?
		  AND (? = '' OR p.source = ?)
		ORDER BY rank
		LIMIT 25
	`

	rows, err := db.Query(sqlQuery, query, source, source)
	if err != nil {
		fmt.Printf("failed to execute search query: %v\n", err)
		return nil, err
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
			return nil, err
		}

		r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		r.IndexedAt, _ = time.Parse("2006-01-02 15:04:05", indexedAt)
		results = append(results, r)
	}

	fmt.Printf("search completed: %d results\n", len(results))
	return results, rows.Err()
}

func listRecentPosts(source string) ([]SearchResult, error) {
	rows, err := db.Query(`
		SELECT uri, cid, author_did, author_handle, text, created_at,
			   like_count, repost_count, reply_count, source, indexed_at
		FROM posts
		WHERE (? = '' OR source = ?)
		ORDER BY created_at DESC
		LIMIT 25
	`, source, source)
	if err != nil {
		fmt.Printf("failed to list recent posts: %v\n", err)
		return nil, err
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
			return nil, err
		}

		r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		r.IndexedAt, _ = time.Parse("2006-01-02 15:04:05", indexedAt)
		results = append(results, r)
	}

	fmt.Printf("browse completed: %d results\n", len(results))
	return results, rows.Err()
}

// CountPosts returns the total number of posts in the database
func CountPosts() (int, error) {
	fmt.Println("counting posts in database")

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		fmt.Printf("failed to count posts: %v\n", err)
		return 0, err
	}

	fmt.Printf("post count: %d\n", count)
	return count, nil
}
