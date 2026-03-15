package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	rt "runtime"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

// AuthService provides authentication functionality via Wails bindings
type AuthService struct {
	app      *oauth.ClientApp
	server   *http.Server
	listener net.Listener
	codeChan chan string
	errChan  chan error
	port     int
}

// NewAuthService creates a new AuthService instance
func NewAuthService() *AuthService {
	return &AuthService{
		codeChan: make(chan string, 1),
		errChan:  make(chan error, 1),
	}
}

// Login initiates OAuth login flow for the given handle
func (s *AuthService) Login(handle string) error {
	ctx := context.Background()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", s.port)
	scopes := []string{"atproto", "transition:generic"}

	config := oauth.NewLocalhostConfig(redirectURI, scopes)
	store := oauth.NewMemStore()
	s.app = oauth.NewClientApp(&config, store)

	redirectURL, err := s.app.StartAuthFlow(ctx, handle)
	if err != nil {
		return fmt.Errorf("failed to start auth flow: %w", err)
	}

	s.startCallbackServer()

	if err := openBrowser(redirectURL); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	select {
	case code := <-s.codeChan:
		return s.exchangeCode(ctx, code)
	case err := <-s.errChan:
		return fmt.Errorf("authorization error: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *AuthService) startCallbackServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		code := query.Get("code")
		if code == "" {
			errMsg := query.Get("error")
			if errMsg == "" {
				errMsg = "missing authorization code"
			}
			errDesc := query.Get("error_description")
			s.errChan <- fmt.Errorf("authorization failed: %s - %s", errMsg, errDesc)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Authorization failed: %s\n", errMsg)
			return
		}

		state := query.Get("state")
		iss := query.Get("iss")
		s.codeChan <- fmt.Sprintf("%s|%s|%s", code, state, iss)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Authorization successful! You can close this window.")
	})

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		if err := s.server.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			s.errChan <- err
		}
	}()
}

func (s *AuthService) exchangeCode(ctx context.Context, data string) error {
	parts := strings.SplitN(data, "|", 3)
	if len(parts) < 2 {
		return fmt.Errorf("invalid callback data")
	}

	params := make(map[string][]string)
	params["code"] = []string{parts[0]}
	params["state"] = []string{parts[1]}
	if len(parts) > 2 && parts[2] != "" {
		params["iss"] = []string{parts[2]}
	}

	sessData, err := s.app.ProcessCallback(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to process callback: %w", err)
	}

	auth := &Auth{
		DID:                          sessData.AccountDID.String(),
		Handle:                       sessData.AccountDID.String(),
		AccessJWT:                    sessData.AccessToken,
		RefreshJWT:                   sessData.RefreshToken,
		PDSURL:                       sessData.HostURL,
		SessionID:                    sessData.SessionID,
		AuthServerURL:                sessData.AuthServerURL,
		AuthServerTokenEndpoint:      sessData.AuthServerTokenEndpoint,
		AuthServerRevocationEndpoint: sessData.AuthServerRevocationEndpoint,
		DPoPAuthNonce:                sessData.DPoPAuthServerNonce,
		DPoPHostNonce:                sessData.DPoPHostNonce,
		DPoPPrivateKey:               sessData.DPoPPrivateKeyMultibase,
		UpdatedAt:                    time.Now(),
	}

	if err := UpsertAuth(auth); err != nil {
		return fmt.Errorf("failed to persist auth: %w", err)
	}

	return nil
}

// Whoami returns the current authenticated user, optionally resolving handle from DID
//
// TODO: store [context.Context] in [AuthService] to be able to use wails' runtime.LogWarningf
func (s *AuthService) Whoami(force bool) (*Auth, error) {
	auth, err := GetAuth()
	if err != nil {
		return nil, fmt.Errorf("failed to load auth: %w", err)
	}
	if auth == nil {
		return nil, fmt.Errorf("not logged in")
	}

	if force || strings.HasPrefix(auth.Handle, "did:") {
		did, err := syntax.ParseDID(auth.DID)
		if err != nil {
			return nil, fmt.Errorf("invalid DID in database: %w", err)
		}

		dir := &identity.BaseDirectory{}
		ident, err := dir.LookupDID(context.Background(), did)
		if err != nil {
			return auth, nil
		}

		auth.Handle = ident.Handle.String()

	}

	return auth, nil
}

// IsAuthenticated checks if there is a valid auth record
func (s *AuthService) IsAuthenticated() bool {
	auth, err := GetAuth()
	if err != nil {
		return false
	}
	return auth != nil
}

// RefreshSession attempts to refresh the access token if needed
func (s *AuthService) RefreshSession() error {
	auth, err := GetAuth()
	if err != nil {
		return fmt.Errorf("failed to load auth: %w", err)
	}
	if auth == nil {
		return fmt.Errorf("no session found")
	}

	if auth.SessionID == "" {
		return nil // Cannot refresh without session ID
	}

	redirectURI := "http://127.0.0.1/callback"
	scopes := []string{"atproto", "transition:generic"}
	config := oauth.NewLocalhostConfig(redirectURI, scopes)
	store := oauth.NewMemStore()
	app := oauth.NewClientApp(&config, store)

	did, err := syntax.ParseDID(auth.DID)
	if err != nil {
		return fmt.Errorf("invalid DID in database: %w", err)
	}

	session, err := app.ResumeSession(context.Background(), did, auth.SessionID)
	if err != nil {
		return fmt.Errorf("failed to resume session: %w", err)
	}

	newAccessToken, err := session.RefreshTokens(context.Background())
	if err != nil {
		return fmt.Errorf("failed to refresh tokens: %w", err)
	}

	if newAccessToken != "" {
		auth.AccessJWT = newAccessToken
		auth.UpdatedAt = time.Now()
		if err := UpsertAuth(auth); err != nil {
			return fmt.Errorf("failed to update refreshed tokens: %w", err)
		}
	}

	return nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch rt.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}
