package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/evermeet/evermeet/internal/api"
	"github.com/evermeet/evermeet/internal/blob"
	"github.com/evermeet/evermeet/internal/config"
	"github.com/evermeet/evermeet/internal/email"
	"github.com/evermeet/evermeet/internal/node"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
	"lukechampine.com/blake3"
)

func main() {
	cfgPath := flag.String("config", "evermeet.toml", "path to config file")
	httpPort := flag.Int("port", 0, "HTTP port (overrides config)")
	p2pPort := flag.Int("p2p-port", 0, "P2P port (overrides config)")
	dataDir := flag.String("data", "", "Data directory (overrides config)")
	flag.Parse()

	logger := log.New(os.Stderr, "", log.LstdFlags)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	if *httpPort != 0 {
		cfg.Node.Port = *httpPort
	}
	if *p2pPort != 0 {
		cfg.P2P.ListenPort = *p2pPort
	}
	if *dataDir != "" {
		cfg.Node.DataDir = *dataDir
	}

	if err := os.MkdirAll(cfg.Node.DataDir, 0755); err != nil {
		logger.Fatalf("create data dir: %v", err)
	}

	// Instance key: a persistent Ed25519 key whose public key deterministically
	// derives the instance ID. Used as the server secret for user key encryption.
	instancePriv, err := loadOrGenerateInstanceKey(filepath.Join(cfg.Node.DataDir, "instance.key"))
	if err != nil {
		logger.Fatalf("instance key: %v", err)
	}
	instanceID := deriveInstanceID(instancePriv.Public().(ed25519.PublicKey))
	logger.Printf("instance id: %s", instanceID)

	// SQLite database.
	dbPath := filepath.Join(cfg.Node.DataDir, "evermeet.db")
	db, err := store.Open(dbPath)
	if err != nil {
		logger.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Optional email client (nil in dev if SMTP not configured).
	var emailClient *email.Client
	if cfg.Email.SMTPHost != "" {
		emailClient = email.New(email.Config{
			Host: cfg.Email.SMTPHost,
			Port: cfg.Email.SMTPPort,
			User: cfg.Email.SMTPUser,
			Pass: cfg.Email.SMTPPass,
			From: cfg.Email.From,
		})
	}

	baseURL := cfg.Node.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", cfg.Node.Port)
	}

	// P2P Node
	p2pNode, err := node.New(db, logger, cfg.P2P.ListenPort)
	if err != nil {
		logger.Fatalf("start p2p node: %v", err)
	}
	defer p2pNode.Close()

	blobStore, err := blob.New(filepath.Join(cfg.Node.DataDir, "blobs"))
	if err != nil {
		logger.Fatalf("blob store: %v", err)
	}

	apiServer := api.NewServer(db, blobStore, emailClient, baseURL, instancePriv.Seed(), instanceID, logger, p2pNode, cfg)

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","instance_id":%q,"p2p_peers":%d}`, instanceID, p2pNode.PeerCount())
	})

	r.Mount("/", apiServer.Router())

	// SPA fallback — must come after API routes.
	r.NotFound(spaHandler())

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Node.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("evermeet running on http://localhost:%d", cfg.Node.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Println("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
}

// deriveInstanceID returns a 16-character hex string from blake3(pubkey)[:8].
func deriveInstanceID(pub ed25519.PublicKey) string {
	h := blake3.Sum256(pub)
	return hex.EncodeToString(h[:8])
}

// loadOrGenerateInstanceKey loads a PEM-encoded Ed25519 private key from path,
// generating and saving a new one if the file does not exist.
func loadOrGenerateInstanceKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate: %w", err)
		}
		block := &pem.Block{Type: "ED25519 PRIVATE KEY", Bytes: priv.Seed()}
		if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
			return nil, fmt.Errorf("save: %w", err)
		}
		return priv, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil || len(block.Bytes) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid key file %s", path)
	}
	return ed25519.NewKeyFromSeed(block.Bytes), nil
}

// spaHandler serves static files from the embedded web/build directory.
// Unknown paths fall back to index.html so SvelteKit's client-side router handles them.
func spaHandler() http.HandlerFunc {
	sub, err := fs.Sub(webFS, "web/build")
	if err != nil {
		log.Fatalf("embed sub: %v", err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.ToSlash(filepath.Clean(r.URL.Path))
		if path == "/" {
			path = "index.html"
		} else if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		f, err := sub.Open(path)
		if err != nil {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	}
}
