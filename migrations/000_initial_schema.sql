-- Combined initial schema for bsky-browser-desktop
-- Includes all migrations: auth, posts, FTS5, OAuth fields, and facets

-- Auth table with all OAuth fields
CREATE TABLE IF NOT EXISTS auth (
    did                           TEXT PRIMARY KEY,
    handle                        TEXT NOT NULL,
    access_jwt                    TEXT NOT NULL,
    refresh_jwt                   TEXT NOT NULL,
    pds_url                       TEXT NOT NULL,
    session_id                    TEXT,
    auth_server_url               TEXT,
    auth_server_token_endpoint    TEXT,
    auth_server_revocation_endpoint TEXT,
    dpop_auth_nonce               TEXT,
    dpop_host_nonce               TEXT,
    dpop_private_key              TEXT,
    updated_at                    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Posts table with facets support
CREATE TABLE IF NOT EXISTS posts (
    uri           TEXT PRIMARY KEY,
    cid           TEXT NOT NULL,
    author_did    TEXT NOT NULL,
    author_handle TEXT NOT NULL,
    text          TEXT NOT NULL DEFAULT '',
    created_at    DATETIME NOT NULL,
    like_count    INTEGER DEFAULT 0,
    repost_count  INTEGER DEFAULT 0,
    reply_count   INTEGER DEFAULT 0,
    source        TEXT NOT NULL CHECK(source IN ('saved', 'liked')),
    facets        TEXT,
    indexed_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- FTS5 virtual table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(
    text,
    author_handle,
    content='posts',
    content_rowid='rowid',
    tokenize='unicode61'
);

-- Triggers to keep FTS5 index in sync with posts table
CREATE TRIGGER IF NOT EXISTS posts_ai AFTER INSERT ON posts BEGIN
    INSERT INTO posts_fts(rowid, text, author_handle)
    VALUES (new.rowid, new.text, new.author_handle);
END;

CREATE TRIGGER IF NOT EXISTS posts_ad AFTER DELETE ON posts BEGIN
    INSERT INTO posts_fts(posts_fts, rowid, text, author_handle)
    VALUES ('delete', old.rowid, old.text, old.author_handle);
END;

CREATE TRIGGER IF NOT EXISTS posts_au AFTER UPDATE ON posts BEGIN
    INSERT INTO posts_fts(posts_fts, rowid, text, author_handle)
    VALUES ('delete', old.rowid, old.text, old.author_handle);
    INSERT INTO posts_fts(rowid, text, author_handle)
    VALUES (new.rowid, new.text, new.author_handle);
END;
