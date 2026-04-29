package main

import (
	"context"
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
	"github.com/evermeet/evermeet/internal/config"
	"github.com/evermeet/evermeet/internal/email"
	"github.com/evermeet/evermeet/internal/identity"
	"github.com/evermeet/evermeet/internal/node"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
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
		cfg.Identity.KeyDir = filepath.Join(*dataDir, "keys")
	}

	if err := os.MkdirAll(cfg.Node.DataDir, 0755); err != nil {
		logger.Fatalf("create data dir: %v", err)
	}

	// Node identity keypair (used for protocol signing).
	kp, err := identity.LoadOrGenerate(cfg.Identity.KeyDir)
	if err != nil {
		logger.Fatalf("load keypair: %v", err)
	}
	nodeDID := identity.DeriveDID(kp.SigningPub)
	logger.Printf("node identity: %s", nodeDID)

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

	// Server secret: used to derive per-user key-encryption passwords.
	// In production, load this from a file or env var. For now, derive from node key seed.
	serverSecret := kp.SigningPriv.Seed()

	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Node.Port)
	
	// P2P Node
	p2pNode, err := node.New(db, logger, cfg.P2P.ListenPort)
	if err != nil {
		logger.Fatalf("start p2p node: %v", err)
	}
	defer p2pNode.Close()

	apiServer := api.NewServer(db, emailClient, baseURL, serverSecret, logger, p2pNode)

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","did":%q,"p2p_peers":%d}`, nodeDID, p2pNode.PeerCount())
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
