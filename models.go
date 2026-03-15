package main

import (
	"time"
)

// Post represents a Bluesky post in the database
type Post struct {
	URI          string    `json:"uri"`
	CID          string    `json:"cid"`
	AuthorDID    string    `json:"author_did"`
	AuthorHandle string    `json:"author_handle"`
	Text         string    `json:"text"`
	CreatedAt    time.Time `json:"created_at"`
	LikeCount    int       `json:"like_count"`
	RepostCount  int       `json:"repost_count"`
	ReplyCount   int       `json:"reply_count"`
	Source       string    `json:"source"` // 'saved' or 'liked'
	Facets       string    `json:"facets"` // JSON-encoded facets
	IndexedAt    time.Time `json:"indexed_at"`
}

// Auth represents OAuth authentication information
type Auth struct {
	DID                          string    `json:"did"`
	Handle                       string    `json:"handle"`
	AccessJWT                    string    `json:"access_jwt"`
	RefreshJWT                   string    `json:"refresh_jwt"`
	PDSURL                       string    `json:"pds_url"`
	SessionID                    string    `json:"session_id"`
	AuthServerURL                string    `json:"auth_server_url"`
	AuthServerTokenEndpoint      string    `json:"auth_server_token_endpoint"`
	AuthServerRevocationEndpoint string    `json:"auth_server_revocation_endpoint"`
	DPoPAuthNonce                string    `json:"dpop_auth_nonce"`
	DPoPHostNonce                string    `json:"dpop_host_nonce"`
	DPoPPrivateKey               string    `json:"dpop_private_key"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

// SearchResult represents a search result with BM25 ranking
type SearchResult struct {
	Post
	Rank float64 `json:"rank"`
}
