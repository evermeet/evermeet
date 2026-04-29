package api

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/evermeet/evermeet/internal/email"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Server holds shared dependencies for all API handlers.
type Server struct {
	db           *store.DB
	email        *email.Client
	log          *log.Logger
	baseURL      string
	serverSecret []byte // used to derive per-user key encryption passwords
	webauthn     *webauthn.WebAuthn
}

// NewServer creates a Server with the given dependencies.
func NewServer(db *store.DB, emailClient *email.Client, baseURL string, serverSecret []byte, logger *log.Logger) *Server {
	u, _ := url.Parse(baseURL)
	w, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "Evermeet",
		RPID:          u.Hostname(),
		RPOrigins:     []string{baseURL},
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementRequired,
			UserVerification: protocol.VerificationPreferred,
		},
	})
	if err != nil {
		logger.Fatalf("webauthn: %v", err)
	}

	return &Server{
		db:           db,
		email:        emailClient,
		log:          logger,
		baseURL:      baseURL,
		serverSecret: serverSecret,
		webauthn:     w,
	}
}

// Router builds and returns the chi router with all routes registered.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(s.sessionMiddleware)

	// Auth
	r.Post("/api/auth/magic-link", s.handleMagicLinkRequest)
	r.Get("/api/auth/magic-link/verify", s.handleMagicLinkVerify)
	r.Post("/api/auth/logout", s.handleLogout)
	r.Get("/api/auth/me", s.handleMe)

	// Passkeys (WebAuthn)
	r.Post("/api/auth/passkey/signup/start", s.handlePasskeySignupStart)
	r.Post("/api/auth/passkey/signup/finish", s.handlePasskeySignupFinish)
	r.Post("/api/auth/passkey/register/start", s.requireAuth(s.handlePasskeyRegisterStart))
	r.Post("/api/auth/passkey/register/finish", s.requireAuth(s.handlePasskeyRegisterFinish))
	r.Post("/api/auth/passkey/login/start", s.handlePasskeyLoginStart)
	r.Post("/api/auth/passkey/login/finish", s.handlePasskeyLoginFinish)

	// Events
	r.Get("/api/events", s.handleListEvents)
	r.Post("/api/events", s.requireAuth(s.handleCreateEvent))
	r.Get("/api/events/{id}", s.handleGetEvent)
	r.Put("/api/events/{id}", s.requireAuth(s.handleUpdateEvent))
	r.Post("/api/events/{id}/rsvp", s.requireAuth(s.handleRSVP))
	r.Get("/api/events/{id}/rsvps", s.requireAuth(s.handleListRSVPs))

	// Users
	r.Get("/api/users/{did}", s.handleGetUser)
	r.Put("/api/users/me", s.requireAuth(s.handleUpdateMe))

	// Well-known: instance public key for federation auth
	r.Get("/.well-known/evermeet-node-key", s.handleNodeKey)

	return r
}

// sessionMiddleware reads the session cookie, validates it, and attaches the
// authenticated DID + decrypted private key to the request context.
func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		sess, err := s.db.GetSession(r.Context(), cookie.Value)
		if err != nil || sess == nil || time.Now().After(sess.ExpiresAt) {
			next.ServeHTTP(w, r)
			return
		}

		priv, err := s.decryptUserKey(r.Context(), sess.DID)
		if err != nil {
			s.log.Printf("session key decrypt for %s: %v", sess.DID, err)
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, withAuth(r, sess.DID, priv))
	})
}

// requireAuth wraps a handler and returns 401 if the user is not authenticated.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authDID(r) == "" {
			jsonErr(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next(w, r)
	}
}

func (s *Server) handleNodeKey(w http.ResponseWriter, r *http.Request) {
	// Placeholder — populated in Phase 6 when the libp2p node key is available.
	jsonErr(w, http.StatusNotImplemented, "node key not yet configured")
}
