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
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/evermeet/evermeet/internal/api"
	"github.com/evermeet/evermeet/internal/blob"
	"github.com/evermeet/evermeet/internal/config"
	"github.com/evermeet/evermeet/internal/email"
	"github.com/evermeet/evermeet/internal/node"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/evermeet/evermeet/internal/version"
	"github.com/go-chi/chi/v5"
	"lukechampine.com/blake3"
)

func main() {
	cfgPath := flag.String("config", "evermeet.toml", "path to config file")
	httpPort := flag.Int("port", 0, "HTTP port (overrides config)")
	p2pPort := flag.Int("p2p-port", 0, "P2P port (overrides config)")
	dataDir := flag.String("data", "", "Data directory (overrides config)")
	verbose := flag.Bool("verbose", false, "Enable verbose backend logging, including HTTP request logs")
	bootstrapMode := flag.Bool("bootstrap", false, "Run as a DHT bootstrap node only (no HTTP server or admin UI)")
	announceAddrsFlag := flag.String("announce-addrs", "", "comma-separated extra multiaddrs to advertise when behind Docker/NAT (no /p2p/ suffix); e.g. /ip4/203.0.113.50/tcp/4002")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)

	if *bootstrapMode {
		runBootstrap(logger, *p2pPort, *dataDir, splitCommaNonEmpty(*announceAddrsFlag))
		return
	}

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
	if *verbose {
		cfg.Node.Verbose = true
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

	baseURL := cfg.Node.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", cfg.Node.Port)
	}
	mailDefaultHost := ""
	homeHostStr := instanceID
	if u, err := url.Parse(baseURL); err == nil {
		mailDefaultHost = u.Hostname()
		if u.Host != "" {
			homeHostStr = instanceID + "@" + u.Host
		}
	}

	logger.Printf("evermeet %s starting", version.Version)
	logger.Printf("  instance: %s", homeHostStr)
	logger.Printf("  public base URL: %s  (HTTP listen :%d)", baseURL, cfg.Node.Port)

	// SQLite database.
	dbPath := filepath.Join(cfg.Node.DataDir, "evermeet.db")
	db, err := store.Open(dbPath)
	if err != nil {
		logger.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Optional email client: SMTP from config, or local sendmail (From or evermeet@<base host>).
	emailClient, emailDesc := email.NewClient(cfg.Email, mailDefaultHost)
	if emailDesc != "" {
		logger.Printf("email: %s", emailDesc)
	}

	ctxBoot := context.Background()
	hasAdmins, err := db.HasAdminAccounts(ctxBoot)
	if err != nil {
		logger.Fatalf("admin check: %v", err)
	}

	var p2pNode *node.Node
	if hasAdmins {
		p2pNode, err = node.New(db, logger, cfg.P2P.ListenPort, cfg.Node.DataDir, homeHostStr, cfg.P2P.BootstrapPeers, cfg.P2P.AnnounceAddrs)
		if err != nil {
			logger.Fatalf("start p2p node: %v", err)
		}
	} else {
		logger.Printf("P2P: not started (%s); first-time setup — will start automatically when setup completes.", version.Version)
	}

	blobStore, err := blob.New(filepath.Join(cfg.Node.DataDir, "blobs"))
	if err != nil {
		logger.Fatalf("blob store: %v", err)
	}

	apiServer := api.NewServer(db, blobStore, emailClient, baseURL, instancePriv.Seed(), instanceID, logger, p2pNode, cfg, *cfgPath)

	r := chi.NewRouter()

	r.Get("/health", apiServer.HandleHealth)

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

	apiServer.SetBackgroundContext(ctx)
	apiServer.StartDHTHeartbeat(ctx)

	go func() {
		logger.Printf("listening on http://localhost:%d", cfg.Node.Port)
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
	apiServer.CloseP2P()
}

func splitCommaNonEmpty(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func runBootstrap(logger *log.Logger, p2pPort int, dataDir string, announceAddrs []string) {
	if p2pPort == 0 {
		p2pPort = 4001
	}
	if dataDir == "" {
		dataDir = "./data"
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Fatalf("create data dir: %v", err)
	}

	n, err := node.NewBootstrap(logger, p2pPort, dataDir, nil, announceAddrs)
	if err != nil {
		logger.Fatalf("bootstrap node: %v", err)
	}
	defer n.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	logger.Println("bootstrap node shutting down...")
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
