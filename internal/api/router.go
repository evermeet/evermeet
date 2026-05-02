package api

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/evermeet/evermeet/internal/blob"
	"github.com/evermeet/evermeet/internal/config"
	"github.com/evermeet/evermeet/internal/email"
	"github.com/evermeet/evermeet/internal/node"
	"github.com/evermeet/evermeet/internal/routing"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/evermeet/evermeet/internal/version"
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
	p2pMu        sync.RWMutex
	bgCtx        context.Context
	bgCtxMu      sync.RWMutex
	heartbeatMu  sync.Mutex
	dhtHBRunning bool
	setupMu      sync.Mutex
	setupToken   string
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
	setupToken := ""
	hasAdmins, err := db.HasAdminAccounts(context.Background())
	if err != nil {
		logger.Fatalf("admin setup check: %v", err)
	}
	if !hasAdmins {
		setupToken = randomHex(32)
		logger.Println("============================================================")
		logger.Printf("Evermeet %s — first-time setup required", version.Version)
		logger.Printf("Admin setup token: %s", setupToken)
		logger.Println("Open this instance in a browser and paste this token to create the first admin account.")
		logger.Println("This token is one-time and only kept in memory for this server process.")
		logger.Println("============================================================")
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
		setupToken:   setupToken,
	}
}

// SetBackgroundContext sets the context used for DHT heartbeats after first-time
// setup (and should be the same root context used for process shutdown). Call once from main.
func (s *Server) SetBackgroundContext(ctx context.Context) {
	s.bgCtxMu.Lock()
	defer s.bgCtxMu.Unlock()
	s.bgCtx = ctx
}

func (s *Server) backgroundCtx() context.Context {
	s.bgCtxMu.RLock()
	defer s.bgCtxMu.RUnlock()
	if s.bgCtx != nil {
		return s.bgCtx
	}
	return context.Background()
}

// libp2pNode returns the running libp2p host, or nil if P2P has not been started yet.
func (s *Server) libp2pNode() *node.Node {
	s.p2pMu.RLock()
	defer s.p2pMu.RUnlock()
	return s.node
}

func (s *Server) publisher() *routing.Publisher {
	s.p2pMu.RLock()
	defer s.p2pMu.RUnlock()
	return s.dhtPublisher
}

// PeerCount returns the current libp2p peer count, or 0 if P2P is not running.
func (s *Server) PeerCount() int {
	n := s.libp2pNode()
	if n == nil {
		return 0
	}
	return n.PeerCount()
}

// HandleHealth writes JSON health including live P2P peer count.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","version":%q,"instance_id":%q,"p2p_peers":%d}`, version.Version, s.homeHost(), s.PeerCount())
}

// ensureP2P starts the libp2p node and DHT publisher if they are not already running.
func (s *Server) ensureP2P() error {
	s.p2pMu.Lock()
	defer s.p2pMu.Unlock()
	if s.node != nil {
		return nil
	}
	n, err := node.New(s.db, s.log, s.cfg.P2P.ListenPort, s.cfg.Node.DataDir, s.homeHost())
	if err != nil {
		return err
	}
	s.node = n
	s.dhtPublisher = routing.NewPublisher(n.DHTPublish, s.instancePriv, s.baseURL)
	return nil
}

// CloseP2P shuts down the libp2p host if one was started. Safe to call multiple times.
func (s *Server) CloseP2P() {
	s.p2pMu.Lock()
	defer s.p2pMu.Unlock()
	if s.node != nil {
		_ = s.node.Close()
	}
	s.node = nil
	s.dhtPublisher = nil
}

// Router builds and returns the chi router with all routes registered.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	if s.cfg != nil && s.cfg.Node.Verbose {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(s.sessionMiddleware)
	r.Use(s.setupGate)

	// Auth
	r.Get("/api/setup/status", s.handleSetupStatus)
	r.Post("/api/setup/complete", s.handleSetupComplete)
	r.Get("/api/admin/overview", s.requireAdmin(s.handleAdminOverview))
	r.Get("/api/admin/objects", s.requireAdmin(s.handleAdminObjectsOverview))
	r.Get("/api/admin/objects/{type}", s.requireAdmin(s.handleAdminObjectsByType))
	r.Get("/api/admin/email", s.requireAdmin(s.handleAdminEmailConfig))
	r.Post("/api/admin/email/test", s.requireAdmin(s.handleAdminEmailTest))
	r.Get("/api/admin/admins", s.requireAdmin(s.handleAdminAdminsList))
	r.Post("/api/admin/admins", s.requireOwner(s.handleAdminAdminsCreate))
	r.Put("/api/admin/admins/{did}/role", s.requireOwner(s.handleAdminAdminsSetRole))
	r.Post("/api/auth/magic-link", s.handleMagicLinkRequest)
	r.Post("/api/auth/magic-link/status", s.handleMagicLinkStatus)
	r.Get("/api/auth/magic-link/verify", s.handleMagicLinkVerify)
	r.Post("/api/auth/logout", s.handleLogout)
	r.Get("/api/auth/me", s.handleMe)
	r.Get("/api/auth/methods", s.requireAuth(s.handleAuthMethods))
	r.Get("/api/auth/rsvp-receipts", s.requireAuth(s.handleListRSVPReceipts))
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
	r.Get("/api/events/{id}/attendees", s.handleEventAttendees)
	r.Get("/api/events/{id}/history", s.handleEventHistory)
	r.Put("/api/events/{id}", s.requireAuth(s.handleUpdateEvent))
	r.Delete("/api/events/{id}", s.requireAuth(s.handleDeleteEvent))
	r.Post("/api/events/{id}/rsvp", s.requireAuth(s.handleRSVP))
	r.Delete("/api/events/{id}/rsvp", s.requireAuth(s.handleCancelRSVP))
	r.Put("/api/events/{id}/rsvp/guest-list", s.requireAuth(s.handleUpdateRSVPGuestVisibility))
	r.Get("/api/events/{id}/rsvp/status", s.requireAuth(s.handleMyRSVPStatus))
	r.Get("/api/events/{id}/rsvps", s.requireAuth(s.handleListRSVPs))

	// Users
	r.Get("/api/users/{did}", s.handleGetUser)

	// Cross-instance auth
	r.Post("/api/auth/resolve-home", s.handleResolveHome)
	r.Post("/api/auth/delegate", s.requireAuth(s.handleCreateDelegation))
	r.Post("/api/auth/delegate-verify", s.handleVerifyDelegation)

	// Federation
	r.Post("/api/federation/rsvp-receipts", s.handleReceiveRSVPReceipt)

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
		cookie, err := s.sessionCookie(r)
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

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		ok, err := s.db.IsAdmin(r.Context(), authDID(r))
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, "admin lookup failed")
			return
		}
		if !ok {
			jsonErr(w, http.StatusForbidden, "admin access required")
			return
		}
		next(w, r)
	})
}

func (s *Server) requireOwner(next http.HandlerFunc) http.HandlerFunc {
	return s.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		role, err := s.db.GetAdminRole(r.Context(), authDID(r))
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, "admin role lookup failed")
			return
		}
		if role != "owner" {
			jsonErr(w, http.StatusForbidden, "owner access required")
			return
		}
		next(w, r)
	})
}

// StartDHTHeartbeat starts the background goroutine that re-publishes all
// local user email→homeInstance mappings to the DHT every 12 hours.
// Idempotent: safe to call from main at boot (no-op until a publisher exists) and again after lazy P2P start.
func (s *Server) StartDHTHeartbeat(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	s.heartbeatMu.Lock()
	if s.dhtHBRunning {
		s.heartbeatMu.Unlock()
		return
	}
	pub := s.publisher()
	if pub == nil {
		s.heartbeatMu.Unlock()
		return
	}
	s.dhtHBRunning = true
	s.heartbeatMu.Unlock()

	pub.StartHeartbeat(ctx, 12*time.Hour, func(ctx context.Context) ([]string, error) {
		return s.db.GetAllUserEmails(ctx)
	})
	go s.startSIWEDHTHeartbeat(ctx, 12*time.Hour)
	go s.startDIDDHTHeartbeat(ctx, 12*time.Hour)
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
	pub := s.publisher()
	if pub == nil {
		return
	}
	for _, identity := range identities {
		if ctx.Err() != nil {
			return
		}
		if err := pub.PublishEthereum(ctx, identity.ChainID, identity.Address); err != nil {
			s.log.Printf("dht publish siwe %s:%s: %v", identity.ChainID, identity.Address, err)
		}
	}
}

func (s *Server) startDIDDHTHeartbeat(ctx context.Context, interval time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}
	s.publishDIDRouting(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.publishDIDRouting(ctx)
		}
	}
}

func (s *Server) publishDIDRouting(ctx context.Context) {
	dids, err := s.db.GetAllUserDIDs(ctx)
	if err != nil {
		s.log.Printf("dht did list: %v", err)
		return
	}
	pub := s.publisher()
	if pub == nil {
		return
	}
	for _, did := range dids {
		if ctx.Err() != nil {
			return
		}
		if err := pub.PublishDID(ctx, did); err != nil {
			s.log.Printf("dht publish did %s: %v", did, err)
		}
	}
}

// homeHost returns the canonical instance address: "instanceID@hostname".
func (s *Server) homeHost() string {
	if u, err := url.Parse(s.baseURL); err == nil && u.Host != "" {
		return s.instanceID + "@" + u.Host
	}
	return s.instanceID
}

func (s *Server) handleNodeKey(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]any{
		"instance_id": s.homeHost(),
		"public_key":  s.instancePriv.Public().(ed25519.PublicKey), // raw Ed25519 public key bytes, base64 by JSON encoder
	})
}
