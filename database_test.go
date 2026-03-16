package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := Open(dbPath); err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	t.Cleanup(func() {
		if err := Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})
}

func TestSearchPostsBrowseMode(t *testing.T) {
	openTestDB(t)

	posts := []*Post{
		{
			URI:          "at://did:plc:test/app.bsky.feed.post/1",
			CID:          "cid-1",
			AuthorDID:    "did:plc:test",
			AuthorHandle: "alice.test",
			Text:         "older saved post",
			CreatedAt:    time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC),
			Source:       "saved",
		},
		{
			URI:          "at://did:plc:test/app.bsky.feed.post/2",
			CID:          "cid-2",
			AuthorDID:    "did:plc:test",
			AuthorHandle: "alice.test",
			Text:         "newer liked post",
			CreatedAt:    time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
			Source:       "liked",
		},
	}

	for _, post := range posts {
		if err := InsertPost(post); err != nil {
			t.Fatalf("InsertPost() error = %v", err)
		}
	}

	results, err := SearchPosts("", "", 25, "created_at", "desc")
	if err != nil {
		t.Fatalf("SearchPosts(empty) error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("SearchPosts(empty) len = %d, want 2", len(results))
	}
	if results[0].URI != posts[1].URI {
		t.Fatalf("SearchPosts(empty) first URI = %q, want %q", results[0].URI, posts[1].URI)
	}
	if results[0].CreatedAt.IsZero() {
		t.Fatal("SearchPosts(empty) CreatedAt is zero, want parsed timestamp")
	}

	starResults, err := SearchPosts("*", "saved", 25, "created_at", "desc")
	if err != nil {
		t.Fatalf("SearchPosts(*) error = %v", err)
	}
	if len(starResults) != 1 {
		t.Fatalf("SearchPosts(*) len = %d, want 1", len(starResults))
	}
	if starResults[0].Source != "saved" {
		t.Fatalf("SearchPosts(*) source = %q, want %q", starResults[0].Source, "saved")
	}
}

func TestSQLiteOAuthStorePersistsSession(t *testing.T) {
	openTestDB(t)

	store := NewSQLiteOAuthStore()
	did, err := syntax.ParseDID("did:plc:xg2vq45muivyy3xwatcehspu")
	if err != nil {
		t.Fatalf("ParseDID() error = %v", err)
	}

	session := oauth.ClientSessionData{
		AccountDID:                   did,
		SessionID:                    "session-123",
		HostURL:                      "https://bsky.social",
		AuthServerURL:                "https://auth.example.com",
		AuthServerTokenEndpoint:      "https://auth.example.com/token",
		AuthServerRevocationEndpoint: "https://auth.example.com/revoke",
		Scopes:                       append([]string(nil), oauthScopes...),
		AccessToken:                  "access-1",
		RefreshToken:                 "refresh-1",
		DPoPAuthServerNonce:          "auth-nonce",
		DPoPHostNonce:                "host-nonce",
		DPoPPrivateKeyMultibase:      "private-key",
	}

	if err := store.SaveSession(context.Background(), session); err != nil {
		t.Fatalf("SaveSession() error = %v", err)
	}

	auth, err := GetAuthByDID(did.String())
	if err != nil {
		t.Fatalf("GetAuthByDID() error = %v", err)
	}
	if auth == nil {
		t.Fatal("GetAuthByDID() = nil, want auth")
	}
	if auth.RefreshJWT != session.RefreshToken {
		t.Fatalf("RefreshJWT = %q, want %q", auth.RefreshJWT, session.RefreshToken)
	}
	if auth.DPoPHostNonce != session.DPoPHostNonce {
		t.Fatalf("DPoPHostNonce = %q, want %q", auth.DPoPHostNonce, session.DPoPHostNonce)
	}

	got, err := store.GetSession(context.Background(), did, session.SessionID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if got.AccessToken != session.AccessToken {
		t.Fatalf("AccessToken = %q, want %q", got.AccessToken, session.AccessToken)
	}

	if err := store.DeleteSession(context.Background(), did, session.SessionID); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}

	deleted, err := store.GetSession(context.Background(), did, session.SessionID)
	if err == nil || deleted != nil {
		t.Fatalf("GetSession() after delete = (%v, %v), want error", deleted, err)
	}
}

func TestOpenMigratesLegacyPostsTableWithoutFacets(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")

	legacyDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	legacySchema := `
		CREATE TABLE posts (
			uri TEXT PRIMARY KEY,
			cid TEXT NOT NULL,
			author_did TEXT NOT NULL,
			author_handle TEXT NOT NULL,
			text TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			like_count INTEGER DEFAULT 0,
			repost_count INTEGER DEFAULT 0,
			reply_count INTEGER DEFAULT 0,
			source TEXT NOT NULL CHECK(source IN ('saved', 'liked')),
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := legacyDB.Exec(legacySchema); err != nil {
		t.Fatalf("creating legacy schema failed: %v", err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("legacyDB.Close() error = %v", err)
	}

	if err := Open(dbPath); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})

	hasColumn, err := columnExists("posts", "facets")
	if err != nil {
		t.Fatalf("columnExists() error = %v", err)
	}
	if !hasColumn {
		t.Fatal("posts.facets missing after migration")
	}

	post := &Post{
		URI:          "at://did:plc:test/app.bsky.feed.post/legacy",
		CID:          "cid-legacy",
		AuthorDID:    "did:plc:test",
		AuthorHandle: "legacy.test",
		Text:         "legacy post",
		CreatedAt:    time.Now().UTC(),
		Source:       "saved",
		Facets:       `[]`,
	}

	if err := InsertPost(post); err != nil {
		t.Fatalf("InsertPost() after migration error = %v", err)
	}
}
