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
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AuthService provides authentication functionality via Wails bindings
type AuthService struct {
	ctx      context.Context
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

func (s *AuthService) setContext(ctx context.Context) {
	s.ctx = ctx
}

// Login initiates OAuth login flow for the given handle
func (s *AuthService) Login(handle string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	s.codeChan = make(chan string, 1)
	s.errChan = make(chan error, 1)

	// Find an available port first
	listenerAddr, err := listenerAddress()
	if err != nil {
		return fmt.Errorf("failed to find available port for OAuth callback: %w", err)
	}

	listener, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	store := NewSQLiteOAuthStore()
	s.app = newOAuthApp(store, s.port)

	redirectURL, err := s.app.StartAuthFlow(ctx, handle)
	if err != nil {
		closeCallbackServer(nil, s.listener)
		s.listener = nil
		s.port = 0
		return fmt.Errorf("failed to start auth flow: %w", err)
	}

	s.startCallbackServer()
	defer s.stopCallbackServer()

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

func (s *AuthService) stopCallbackServer() {
	closeCallbackServer(s.server, s.listener)
	s.server = nil
	s.listener = nil
	s.port = 0
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

	current, err := GetAuthByDID(sessData.AccountDID.String())
	if err != nil {
		return fmt.Errorf("failed to load persisted auth: %w", err)
	}

	handle := ""
	if current != nil {
		handle = current.Handle
	}

	auth := authFromSessionData(sessData, handle)

	if err := UpsertAuth(auth); err != nil {
		return fmt.Errorf("failed to persist auth: %w", err)
	}

	return nil
}

// Whoami returns the current authenticated user, optionally resolving handle from DID.
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		ident, err := dir.LookupDID(ctx, did)
		cancel()
		if err != nil {
			LogWarnf("failed to resolve handle for %s: %v", auth.DID, err)
			return auth, nil
		}

		auth.Handle = ident.Handle.String()
		if err := UpsertAuth(auth); err != nil {
			return nil, fmt.Errorf("failed to persist resolved handle: %w", err)
		}

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
		return nil
	}

	store := NewSQLiteOAuthStore()
	app := newOAuthApp(store, 0)

	did, err := syntax.ParseDID(auth.DID)
	if err != nil {
		return fmt.Errorf("invalid DID in database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	session, err := app.ResumeSession(ctx, did, auth.SessionID)
	if err != nil {
		return fmt.Errorf("failed to resume session: %w", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()
	if _, err := session.RefreshTokens(ctx2); err != nil {
		return fmt.Errorf("failed to refresh tokens: %w", err)
	}

	if err := UpsertAuth(authFromSessionData(session.Data, auth.Handle)); err != nil {
		return fmt.Errorf("failed to persist refreshed session: %w", err)
	}

	return nil
}

// Logout revokes the current session when possible and clears local auth state.
func (s *AuthService) Logout() error {
	auth, err := GetAuth()
	if err != nil {
		return fmt.Errorf("failed to load auth: %w", err)
	}
	if auth == nil {
		return nil
	}

	if auth.SessionID != "" {
		store := NewSQLiteOAuthStore()
		app := newOAuthApp(store, 0)

		did, err := syntax.ParseDID(auth.DID)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			session, resumeErr := app.ResumeSession(ctx, did, auth.SessionID)
			cancel()
			if resumeErr == nil {
				ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
				if revokeErr := session.RevokeSession(ctx2); revokeErr != nil {
					LogWarnf("failed to revoke remote session for %s: %v", auth.DID, revokeErr)
				}
				cancel2()
			} else {
				LogWarnf("failed to resume session for logout (%s): %v", auth.DID, resumeErr)
			}
		} else {
			LogWarnf("failed to parse DID for logout (%s): %v", auth.DID, err)
		}
	}

	if err := ClearAuth(); err != nil {
		return fmt.Errorf("failed to clear auth: %w", err)
	}

	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "auth:logout", map[string]any{
			"did":    auth.DID,
			"handle": auth.Handle,
		})
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
