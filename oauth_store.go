package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

var oauthScopes = []string{"atproto", "transition:generic"}

func oauthCallbackURL(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d/callback", port)
}

func oauthConfig(port int) oauth.ClientConfig {
	return oauth.NewLocalhostConfig(oauthCallbackURL(port), append([]string(nil), oauthScopes...))
}

func oauthConfigForCallbackURL(callbackURL string) oauth.ClientConfig {
	return oauth.NewLocalhostConfig(callbackURL, append([]string(nil), oauthScopes...))
}

func newOAuthApp(store oauth.ClientAuthStore, port int) *oauth.ClientApp {
	config := oauthConfig(port)
	return oauth.NewClientApp(&config, store)
}

func newOAuthAppForAuth(store oauth.ClientAuthStore, auth *Auth) *oauth.ClientApp {
	if auth != nil && auth.OAuthCallbackURL != "" {
		config := oauthConfigForCallbackURL(auth.OAuthCallbackURL)
		return oauth.NewClientApp(&config, store)
	}

	return newOAuthApp(store, 0)
}

func authFromSessionData(sess *oauth.ClientSessionData, handle string, callbackURL string) *Auth {
	if handle == "" {
		handle = sess.AccountDID.String()
	}

	return &Auth{
		DID:                          sess.AccountDID.String(),
		Handle:                       handle,
		AccessJWT:                    sess.AccessToken,
		RefreshJWT:                   sess.RefreshToken,
		PDSURL:                       sess.HostURL,
		SessionID:                    sess.SessionID,
		AuthServerURL:                sess.AuthServerURL,
		AuthServerTokenEndpoint:      sess.AuthServerTokenEndpoint,
		AuthServerRevocationEndpoint: sess.AuthServerRevocationEndpoint,
		OAuthCallbackURL:             callbackURL,
		DPoPAuthNonce:                sess.DPoPAuthServerNonce,
		DPoPHostNonce:                sess.DPoPHostNonce,
		DPoPPrivateKey:               sess.DPoPPrivateKeyMultibase,
		UpdatedAt:                    time.Now(),
	}
}

func sessionDataFromAuth(auth *Auth) (*oauth.ClientSessionData, error) {
	did, err := syntax.ParseDID(auth.DID)
	if err != nil {
		return nil, fmt.Errorf("invalid DID in database: %w", err)
	}

	return &oauth.ClientSessionData{
		AccountDID:                   did,
		SessionID:                    auth.SessionID,
		HostURL:                      auth.PDSURL,
		AuthServerURL:                auth.AuthServerURL,
		AuthServerTokenEndpoint:      auth.AuthServerTokenEndpoint,
		AuthServerRevocationEndpoint: auth.AuthServerRevocationEndpoint,
		Scopes:                       append([]string(nil), oauthScopes...),
		AccessToken:                  auth.AccessJWT,
		RefreshToken:                 auth.RefreshJWT,
		DPoPAuthServerNonce:          auth.DPoPAuthNonce,
		DPoPHostNonce:                auth.DPoPHostNonce,
		DPoPPrivateKeyMultibase:      auth.DPoPPrivateKey,
	}, nil
}

type SQLiteOAuthStore struct {
	requests map[string]oauth.AuthRequestData
	mu       sync.Mutex
}

func NewSQLiteOAuthStore() *SQLiteOAuthStore {
	return &SQLiteOAuthStore{
		requests: make(map[string]oauth.AuthRequestData),
	}
}

func (s *SQLiteOAuthStore) GetSession(ctx context.Context, did syntax.DID, sessionID string) (*oauth.ClientSessionData, error) {
	auth, err := GetAuthByDID(did.String())
	if err != nil {
		return nil, err
	}
	if auth == nil || auth.SessionID != sessionID {
		return nil, fmt.Errorf("session not found: %s", did)
	}

	return sessionDataFromAuth(auth)
}

func (s *SQLiteOAuthStore) SaveSession(ctx context.Context, sess oauth.ClientSessionData) error {
	auth, err := GetAuthByDID(sess.AccountDID.String())
	if err != nil {
		return err
	}

	handle := ""
	callbackURL := ""
	if auth != nil {
		handle = auth.Handle
		callbackURL = auth.OAuthCallbackURL
	}

	return UpsertAuth(authFromSessionData(&sess, handle, callbackURL))
}

func (s *SQLiteOAuthStore) DeleteSession(ctx context.Context, did syntax.DID, sessionID string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM auth WHERE did = ? AND session_id = ?", did.String(), sessionID)
	return err
}

func (s *SQLiteOAuthStore) GetAuthRequestInfo(ctx context.Context, state string) (*oauth.AuthRequestData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, ok := s.requests[state]
	if !ok {
		return nil, fmt.Errorf("request info not found: %s", state)
	}
	return &info, nil
}

func (s *SQLiteOAuthStore) SaveAuthRequestInfo(ctx context.Context, info oauth.AuthRequestData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.requests[info.State]; ok {
		return fmt.Errorf("auth request already saved for state %s", info.State)
	}

	s.requests[info.State] = info
	return nil
}

func (s *SQLiteOAuthStore) DeleteAuthRequestInfo(ctx context.Context, state string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.requests, state)
	return nil
}

func listenerAddress() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to find available port: %w", err)
	}
	addr := listener.Addr().String()
	listener.Close()
	return addr, nil
}

func closeCallbackServer(server *http.Server, listener httpCloser) {
	if server != nil {
		_ = server.Close()
	}
	if listener != nil {
		_ = listener.Close()
	}
}

type httpCloser interface {
	Close() error
}
