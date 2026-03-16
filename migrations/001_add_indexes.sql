-- Migration: Add performance indexes for posts table
-- These indexes improve query performance for common filter and sort operations

CREATE INDEX IF NOT EXISTS idx_posts_author_did ON posts(author_did);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
CREATE INDEX IF NOT EXISTS idx_posts_source ON posts(source);
