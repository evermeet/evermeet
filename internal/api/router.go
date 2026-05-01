package api

import (
	"context"
	"crypto/ed25519"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/blob"
	"github.com/evermeet/evermeet/internal/config"
	"github.com/evermeet/evermeet/internal/email"
	"github.com/evermeet/evermeet/internal/node"
	"github.com/evermeet/evermeet/internal/routing"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Server holds shared dependencies for all API handlers.
type Server struct {
	db           *store.DB
	blobs        *blob.Store
	email        *email.Client
	log          *log.Logger
	baseURL      string
	serverSecret []byte // used to derive per-user key encryption passwords
	instancePriv ed25519.PrivateKey
	instanceID   string
	webauthn     *webauthn.WebAuthn
	node         *node.Node
	cfg          *config.Config
	startTime    time.Time
	dhtPublisher *routing.Publisher
}

// NewServer creates a Server with the given dependencies.
func NewServer(db *store.DB, blobStore *blob.Store, emailClient *email.Client, baseURL string, serverSecret []byte, instanceID string, logger *log.Logger, p2pNode *node.Node, cfg *config.Config) *Server {
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

	// Reconstruct the instance Ed25519 private key from its seed (serverSecret).
	instancePriv := ed25519.NewKeyFromSeed(serverSecret)

	var pub *routing.Publisher
	if p2pNode != nil {
		pub = routing.NewPublisher(p2pNode.DHTPublish, instancePriv, baseURL)
	}

	return &Server{
		db:           db,
		blobs:        blobStore,
		email:        emailClient,
		log:          logger,
		baseURL:      baseURL,
		serverSecret: serverSecret,
		instancePriv: instancePriv,
		instanceID:   instanceID,
		webauthn:     w,
		node:         p2pNode,
		cfg:          cfg,
		startTime:    time.Now(),
		dhtPublisher: pub,
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
	r.Post("/api/auth/magic-link/status", s.handleMagicLinkStatus)
	r.Get("/api/auth/magic-link/verify", s.handleMagicLinkVerify)
	r.Post("/api/auth/logout", s.handleLogout)
	r.Get("/api/auth/me", s.handleMe)
	r.Put("/api/auth/profile", s.requireAuth(s.handleUpdateProfile))
	r.Get("/api/users/{did}", s.handleGetUser)

	// Passkeys (WebAuthn)
	r.Post("/api/auth/passkey/signup/start", s.handlePasskeySignupStart)
	r.Post("/api/auth/passkey/signup/finish", s.handlePasskeySignupFinish)
	r.Post("/api/auth/passkey/register/start", s.requireAuth(s.handlePasskeyRegisterStart))
	r.Post("/api/auth/passkey/register/finish", s.requireAuth(s.handlePasskeyRegisterFinish))
	r.Post("/api/auth/passkey/login/start", s.handlePasskeyLoginStart)
	r.Post("/api/auth/passkey/login/finish", s.handlePasskeyLoginFinish)
	r.Post("/api/auth/siwe/start", s.handleSIWEStart)
	r.Post("/api/auth/siwe/finish", s.handleSIWEFinish)

	// Calendars
	r.Get("/api/calendars", s.requireAuth(s.handleListCalendars))
	r.Get("/api/calendars/discover", s.handleDiscoverCalendars)
	r.Post("/api/calendars", s.requireAuth(s.handleCreateCalendar))
	r.Get("/api/calendars/{id}", s.handleGetCalendar)
	r.Put("/api/calendars/{id}", s.requireAuth(s.handleUpdateCalendar))
	r.Post("/api/calendars/{id}/subscribe", s.requireAuth(s.handleSubscribeCalendar))
	r.Delete("/api/calendars/{id}/subscribe", s.requireAuth(s.handleUnsubscribeCalendar))

	// Events
	r.Get("/api/events", s.handleListEvents)
	r.Post("/api/events", s.requireAuth(s.handleCreateEvent))
	r.Post("/api/events/import/preview", s.requireAuth(s.handleImportEventPreview))
	r.Get("/api/events/{id}", s.handleGetEvent)
	r.Get("/api/events/{id}/history", s.handleEventHistory)
	r.Put("/api/events/{id}", s.requireAuth(s.handleUpdateEvent))
	r.Delete("/api/events/{id}", s.requireAuth(s.handleDeleteEvent))
	r.Post("/api/events/{id}/rsvp", s.requireAuth(s.handleRSVP))
	r.Get("/api/events/{id}/rsvps", s.requireAuth(s.handleListRSVPs))

	// Users
	r.Get("/api/users/{did}", s.handleGetUser)

	// Cross-instance auth
	r.Post("/api/auth/resolve-home", s.handleResolveHome)
	r.Post("/api/auth/delegate", s.requireAuth(s.handleCreateDelegation))
	r.Post("/api/auth/delegate-verify", s.handleVerifyDelegation)

	// Blobs
	r.Post("/api/blobs", s.requireAuth(s.handleUploadBlob))
	r.Get("/api/blobs/{hash}", s.handleGetBlob)

	// Well-known: instance public key for federation auth
	r.Get("/.well-known/evermeet-node-key", s.handleNodeKey)

	// Instance status (Public)
	r.Get("/api/instance/status", s.handleInstanceStatus)
	r.Get("/api/node/status", s.handleInstanceStatus) // kept for backwards compat

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
			user, userErr := s.db.GetUser(r.Context(), sess.DID)
			if userErr != nil || user == nil || user.Endpoint == "" || strings.TrimRight(user.Endpoint, "/") == strings.TrimRight(s.baseURL, "/") {
				s.log.Printf("session key decrypt for %s: %v", sess.DID, err)
				next.ServeHTTP(w, r)
				return
			}
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

// StartDHTHeartbeat starts the background goroutine that re-publishes all
// local user email→homeInstance mappings to the DHT every 12 hours.
// Call this once after the server is fully initialised.
func (s *Server) StartDHTHeartbeat(ctx context.Context) {
	if s.dhtPublisher == nil {
		return
	}
	s.dhtPublisher.StartHeartbeat(ctx, 12*time.Hour, func(ctx context.Context) ([]string, error) {
		return s.db.GetAllUserEmails(ctx)
	})
	go s.startSIWEDHTHeartbeat(ctx, 12*time.Hour)
}

func (s *Server) startSIWEDHTHeartbeat(ctx context.Context, interval time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}
	s.publishSIWERouting(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.publishSIWERouting(ctx)
		}
	}
}

func (s *Server) publishSIWERouting(ctx context.Context) {
	identities, err := s.db.GetLocalSIWEIdentities(ctx)
	if err != nil {
		s.log.Printf("dht siwe identities: %v", err)
		return
	}
	for _, identity := range identities {
		if ctx.Err() != nil {
			return
		}
		if err := s.dhtPublisher.PublishEthereum(ctx, identity.ChainID, identity.Address); err != nil {
			s.log.Printf("dht publish siwe %s:%s: %v", identity.ChainID, identity.Address, err)
		}
	}
}

// homeHost returns the canonical instance address: "instanceID@hostname".
func (s *Server) homeHost() string {
	if u, err := url.Parse(s.baseURL); err == nil && u.Hostname() != "" {
		return s.instanceID + "@" + u.Hostname()
	}
	return s.instanceID
}

func (s *Server) handleNodeKey(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]any{
		"instance_id": s.instanceID,
		"public_key":  s.instancePriv.Public().(ed25519.PublicKey), // raw Ed25519 public key bytes, base64 by JSON encoder
	})
}
