package node

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/evermeet/evermeet/internal/routing"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/ipfs/boxo/ipns"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/crypto"
	cryptopb "github.com/libp2p/go-libp2p/core/crypto/pb"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	lukeblake3 "lukechampine.com/blake3"
)

const (
	EventsTopicName         = "em/events"
	UsersTopicName          = "em/users"
	EventFetchProtocol      = "/evermeet/event-fetch/1.0.0"
	UserFetchProtocol       = "/evermeet/user-fetch/1.0.0"
	InstanceInfoProtocol    = "/evermeet/instance-info/1.0.0"
	instanceInfoMaxBodySize = 8192

	dhtProtocolPrefix = "/evermeet"
	dhtNamespace      = "evermeet"
)

type Node struct {
	host   host.Host
	ps     *pubsub.PubSub
	dht    *dht.IpfsDHT
	rd     *drouting.RoutingDiscovery
	db     *store.DB
	log    *log.Logger
	ctx    context.Context
	cancel context.CancelFunc

	eventsTopic *pubsub.Topic
	eventsSub   *pubsub.Subscription
	usersTopic  *pubsub.Topic
	usersSub    *pubsub.Subscription

	isBootstrap bool

	peersMu sync.Mutex
	peers   map[peer.ID]peer.AddrInfo

	evermeetHomeID   string
	peerEvermeet     map[peer.ID]string
	peerBootstrap    map[peer.ID]bool
	peerEvermeetLock sync.RWMutex
}

func New(db *store.DB, l *log.Logger, listenPort int, dataDir, evermeetHomeID string, bootstrapPeers []string) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Load or generate a persistent Ed25519 identity for this node.
	// A stable identity means a stable DHT node ID across restarts.
	privKey, err := loadOrGenerateP2PKey(filepath.Join(dataDir, "p2p.key"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("p2p identity: %w", err)
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
			fmt.Sprintf("/ip6/::/tcp/%d", listenPort),
		),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("libp2p host: %w", err)
	}

	// Parse configured bootstrap peers into AddrInfo.
	var parsedBootstrap []peer.AddrInfo
	for _, s := range bootstrapPeers {
		ai, err := peer.AddrInfoFromString(s)
		if err != nil {
			l.Printf("p2p: skipping invalid bootstrap peer %q: %v", s, err)
			continue
		}
		parsedBootstrap = append(parsedBootstrap, *ai)
	}

	// Kademlia DHT in server mode so this node stores and serves records.
	kad, err := dht.New(ctx, h,
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix(dhtProtocolPrefix),
		dht.BootstrapPeers(parsedBootstrap...),
		dht.Validator(record.NamespacedValidator{
			"pk":         record.PublicKeyValidator{},
			"ipns":       ipns.Validator{KeyBook: h.Peerstore()},
			dhtNamespace: evermeetValidator{},
		}),
	)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("kademlia dht: %w", err)
	}

	if err := kad.Bootstrap(ctx); err != nil {
		l.Printf("dht bootstrap: %v", err)
	}

	// Proactively connect to bootstrap peers so the routing table populates
	// immediately rather than waiting for the periodic bootstrap walk.
	for _, ai := range parsedBootstrap {
		if err := h.Connect(ctx, ai); err != nil {
			l.Printf("p2p: bootstrap connect %s: %v", ai.ID, err)
		}
	}

	// GossipSub
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		kad.Close()
		h.Close()
		cancel()
		return nil, fmt.Errorf("gossipsub: %w", err)
	}

	n := &Node{
		host:             h,
		ps:               ps,
		dht:              kad,
		rd:               drouting.NewRoutingDiscovery(kad),
		db:               db,
		log:              l,
		ctx:              ctx,
		cancel:           cancel,
		peers:            make(map[peer.ID]peer.AddrInfo),
		evermeetHomeID:   evermeetHomeID,
		peerEvermeet:     make(map[peer.ID]string),
		peerBootstrap:    make(map[peer.ID]bool),
		peerEvermeetLock: sync.RWMutex{},
	}

	// mDNS discovery — peers found here also populate the DHT routing table
	// via the Connect call in HandlePeerFound.
	ser := mdns.NewMdnsService(h, "evermeet", n)
	if err := ser.Start(); err != nil {
		l.Printf("mdns error: %v", err)
	}

	if err := n.joinEvents(); err != nil {
		return nil, err
	}
	if err := n.joinUsers(); err != nil {
		return nil, err
	}

	n.host.SetStreamHandler(EventFetchProtocol, n.handleEventFetchStream)
	n.host.SetStreamHandler(UserFetchProtocol, n.handleUserFetchStream)
	n.host.SetStreamHandler(InstanceInfoProtocol, n.handleInstanceInfoStream)

	n.host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(_ network.Network, c network.Conn) {
			remote := c.RemotePeer()
			if remote == n.host.ID() {
				return
			}
			go func() {
				ctx, cancel := context.WithTimeout(n.ctx, 5*time.Second)
				defer cancel()
				n.exchangeInstanceInfo(ctx, remote)
			}()
		},
		DisconnectedF: func(_ network.Network, c network.Conn) {
			n.peerEvermeetLock.Lock()
			delete(n.peerEvermeet, c.RemotePeer())
			delete(n.peerBootstrap, c.RemotePeer())
			n.peerEvermeetLock.Unlock()
		},
	})

	l.Printf("P2P started peer_id=%s", h.ID())
	for _, addr := range h.Addrs() {
		l.Printf("  listen %s", addr.String())
	}

	return n, nil
}

// NewBootstrap starts a minimal libp2p node that only participates in the
// Evermeet DHT as a bootstrap/relay peer. No database, no stream handlers,
// no pubsub — just enough to help other nodes discover each other.
func NewBootstrap(l *log.Logger, listenPort int, dataDir string, bootstrapPeers []string) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	privKey, err := loadOrGenerateP2PKey(filepath.Join(dataDir, "p2p.key"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("p2p identity: %w", err)
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
			fmt.Sprintf("/ip6/::/tcp/%d", listenPort),
		),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("libp2p host: %w", err)
	}

	var parsedBootstrap []peer.AddrInfo
	for _, s := range bootstrapPeers {
		ai, err := peer.AddrInfoFromString(s)
		if err != nil {
			l.Printf("bootstrap: skipping invalid peer %q: %v", s, err)
			continue
		}
		parsedBootstrap = append(parsedBootstrap, *ai)
	}

	kad, err := dht.New(ctx, h,
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix(dhtProtocolPrefix),
		dht.BootstrapPeers(parsedBootstrap...),
		dht.Validator(record.NamespacedValidator{
			"pk":         record.PublicKeyValidator{},
			"ipns":       ipns.Validator{KeyBook: h.Peerstore()},
			dhtNamespace: evermeetValidator{},
		}),
	)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("kademlia dht: %w", err)
	}

	if err := kad.Bootstrap(ctx); err != nil {
		l.Printf("dht bootstrap: %v", err)
	}

	for _, ai := range parsedBootstrap {
		if err := h.Connect(ctx, ai); err != nil {
			l.Printf("bootstrap connect %s: %v", ai.ID, err)
		}
	}

	n := &Node{
		host:             h,
		dht:              kad,
		rd:               drouting.NewRoutingDiscovery(kad),
		log:              l,
		ctx:              ctx,
		cancel:           cancel,
		isBootstrap:      true,
		peers:            make(map[peer.ID]peer.AddrInfo),
		peerEvermeet:     make(map[peer.ID]string),
		peerBootstrap:    make(map[peer.ID]bool),
		peerEvermeetLock: sync.RWMutex{},
	}

	n.host.SetStreamHandler(InstanceInfoProtocol, n.handleInstanceInfoStream)
	n.host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(_ network.Network, c network.Conn) {
			remote := c.RemotePeer()
			if remote == n.host.ID() {
				return
			}
			go func() {
				ctx, cancel := context.WithTimeout(n.ctx, 5*time.Second)
				defer cancel()
				n.exchangeInstanceInfo(ctx, remote)
			}()
		},
		DisconnectedF: func(_ network.Network, c network.Conn) {
			n.peerEvermeetLock.Lock()
			delete(n.peerEvermeet, c.RemotePeer())
			delete(n.peerBootstrap, c.RemotePeer())
			n.peerEvermeetLock.Unlock()
		},
	})

	ser := mdns.NewMdnsService(h, "evermeet", n)
	if err := ser.Start(); err != nil {
		l.Printf("mdns error: %v", err)
	}

	l.Printf("bootstrap node started peer_id=%s", h.ID())
	for _, addr := range h.Addrs() {
		l.Printf("  listen %s/p2p/%s", addr, h.ID())
	}

	return n, nil
}

// DHTPublish stores value under key in the Kademlia DHT.
// key should be an email hash; value is a signed JSON record.
func (n *Node) DHTPublish(ctx context.Context, key []byte, value []byte) error {
	hexKey := hex.EncodeToString(key)
	return n.dht.PutValue(ctx, "/evermeet/"+hexKey, value)
}

// DHTPeerLookup retrieves the value stored under key from the DHT.
// Returns the raw signed record bytes, or an error if not found.
func (n *Node) DHTPeerLookup(ctx context.Context, key []byte) ([]byte, error) {
	hexKey := hex.EncodeToString(key)
	return n.dht.GetValue(ctx, "/evermeet/"+hexKey)
}

type evermeetValidator struct{}

func (evermeetValidator) Validate(key string, value []byte) error {
	if _, _, err := record.SplitKey(key); err != nil {
		return err
	}
	rec, err := routing.UnmarshalRecord(value)
	if err != nil {
		return err
	}
	if rec.HomeInstanceURL == "" || rec.Timestamp <= 0 || rec.Sig == "" {
		return fmt.Errorf("invalid evermeet routing record")
	}
	return nil
}

func (evermeetValidator) Select(key string, values [][]byte) (int, error) {
	var selected int
	var selectedTimestamp int64 = -1
	for i, value := range values {
		rec, err := routing.UnmarshalRecord(value)
		if err != nil {
			continue
		}
		if rec.Timestamp > selectedTimestamp {
			selected = i
			selectedTimestamp = rec.Timestamp
		}
	}
	if selectedTimestamp < 0 {
		return 0, fmt.Errorf("no valid evermeet routing records")
	}
	return selected, nil
}

func (n *Node) joinEvents() error {
	topic, err := n.ps.Join(EventsTopicName)
	if err != nil {
		return fmt.Errorf("join topic: %w", err)
	}
	sub, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	n.eventsTopic = topic
	n.eventsSub = sub
	go n.readEventsLoop()
	return nil
}

func (n *Node) joinUsers() error {
	topic, err := n.ps.Join(UsersTopicName)
	if err != nil {
		return fmt.Errorf("join topic: %w", err)
	}
	sub, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	n.usersTopic = topic
	n.usersSub = sub
	go n.readUsersLoop()
	return nil
}

func (n *Node) BroadcastUser(data []byte) error {
	if n.usersTopic == nil {
		return fmt.Errorf("not joined to users topic")
	}
	return n.usersTopic.Publish(n.ctx, data)
}

func (n *Node) BroadcastEvent(data []byte) error {
	if n.eventsTopic == nil {
		return fmt.Errorf("not joined to events topic")
	}
	return n.eventsTopic.Publish(n.ctx, data)
}

type instanceInfoMsg struct {
	EvermeetInstanceID string `json:"evermeet_instance_id,omitempty"`
	Bootstrap          bool   `json:"bootstrap,omitempty"`
}

func (n *Node) handleInstanceInfoStream(s network.Stream) {
	defer s.Close()
	remote := s.Conn().RemotePeer()
	var req instanceInfoMsg
	if err := json.NewDecoder(io.LimitReader(s, instanceInfoMaxBodySize)).Decode(&req); err != nil {
		return
	}
	n.peerEvermeetLock.Lock()
	if req.EvermeetInstanceID != "" {
		n.peerEvermeet[remote] = req.EvermeetInstanceID
	}
	if req.Bootstrap {
		n.peerBootstrap[remote] = true
	}
	n.peerEvermeetLock.Unlock()
	_ = json.NewEncoder(s).Encode(instanceInfoMsg{
		EvermeetInstanceID: n.evermeetHomeID,
		Bootstrap:          n.isBootstrap,
	})
}

func (n *Node) exchangeInstanceInfo(ctx context.Context, remote peer.ID) {
	s, err := n.host.NewStream(ctx, remote, InstanceInfoProtocol)
	if err != nil {
		return
	}
	defer s.Close()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}
	_ = s.SetDeadline(deadline)
	if err := json.NewEncoder(s).Encode(instanceInfoMsg{
		EvermeetInstanceID: n.evermeetHomeID,
		Bootstrap:          n.isBootstrap,
	}); err != nil {
		return
	}
	var resp instanceInfoMsg
	if err := json.NewDecoder(io.LimitReader(s, instanceInfoMaxBodySize)).Decode(&resp); err != nil {
		return
	}
	n.peerEvermeetLock.Lock()
	if resp.EvermeetInstanceID != "" {
		n.peerEvermeet[remote] = resp.EvermeetInstanceID
	}
	if resp.Bootstrap {
		n.peerBootstrap[remote] = true
	}
	n.peerEvermeetLock.Unlock()
}

func (n *Node) HandlePeerFound(pi peer.AddrInfo) {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	if pi.ID == n.host.ID() {
		return
	}
	if _, ok := n.peers[pi.ID]; ok {
		return
	}

	n.log.Printf("discovered peer: %s", pi.ID)
	n.peers[pi.ID] = pi

	go func() {
		time.Sleep(time.Duration(100+(pi.ID[0]%10)) * time.Millisecond)
		ctx, cancel := context.WithTimeout(n.ctx, 5*time.Second)
		defer cancel()
		if err := n.host.Connect(ctx, pi); err != nil {
			n.log.Printf("failed to connect to %s: %v", pi.ID, err)
		} else {
			n.log.Printf("connected to %s", pi.ID)
		}
	}()
}

func (n *Node) PeerCount() int {
	return len(n.host.Network().Peers())
}

type NodeStatus struct {
	EvermeetInstanceID string       `json:"evermeet_instance_id"`
	Libp2pPeerID       string       `json:"libp2p_peer_id"`
	Addresses          []string     `json:"addresses"`
	Peers              []PeerStatus `json:"peers"`
}

type PeerStatus struct {
	EvermeetInstanceID string   `json:"evermeet_instance_id,omitempty"`
	Libp2pPeerID       string   `json:"libp2p_peer_id"`
	Libp2pFingerprint  string   `json:"libp2p_fingerprint,omitempty"`
	Addresses          []string `json:"addresses"`
	Bootstrap          bool     `json:"bootstrap,omitempty"`
}

func fingerprintFromLibp2pPubKey(pub crypto.PubKey) string {
	if pub == nil || pub.Type() != cryptopb.KeyType_Ed25519 {
		return ""
	}
	raw, err := pub.Raw()
	if err != nil || len(raw) != ed25519.PublicKeySize {
		return ""
	}
	h := lukeblake3.Sum256(raw)
	return hex.EncodeToString(h[:8])
}

func (n *Node) Status() NodeStatus {
	var addrs []string
	for _, a := range n.host.Addrs() {
		addrs = append(addrs, a.String())
	}
	peers := n.host.Network().Peers()
	peerStats := make([]PeerStatus, 0, len(peers))
	for _, p := range peers {
		var pAddrs []string
		for _, a := range n.host.Peerstore().Addrs(p) {
			pAddrs = append(pAddrs, a.String())
		}
		fp := ""
		if pk := n.host.Peerstore().PubKey(p); pk != nil {
			fp = fingerprintFromLibp2pPubKey(pk)
		}
		n.peerEvermeetLock.RLock()
		home := n.peerEvermeet[p]
		isBS := n.peerBootstrap[p]
		n.peerEvermeetLock.RUnlock()
		peerStats = append(peerStats, PeerStatus{
			EvermeetInstanceID: home,
			Libp2pPeerID:       p.String(),
			Libp2pFingerprint:  fp,
			Addresses:          pAddrs,
			Bootstrap:          isBS,
		})
	}
	return NodeStatus{
		EvermeetInstanceID: n.evermeetHomeID,
		Libp2pPeerID:       n.host.ID().String(),
		Addresses:          addrs,
		Peers:              peerStats,
	}
}

// InstancePubKey returns the raw Ed25519 public key bytes of this node's
// persistent identity. Used by /.well-known/evermeet-node-key.
func (n *Node) InstancePubKey() ([]byte, error) {
	pub := n.host.Peerstore().PubKey(n.host.ID())
	if pub == nil {
		return nil, fmt.Errorf("no public key in peerstore")
	}
	return crypto.MarshalPublicKey(pub)
}

func (n *Node) SignInstancePayload(payload []byte) (string, error) {
	priv := n.host.Peerstore().PrivKey(n.host.ID())
	if priv == nil {
		return "", fmt.Errorf("no private key in peerstore")
	}
	sig, err := priv.Sign(payload)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sig), nil
}

func (n *Node) Close() error {
	n.cancel()
	return n.host.Close()
}

// loadOrGenerateP2PKey loads a libp2p Ed25519 private key from path,
// generating and persisting a new one if the file does not exist.
func loadOrGenerateP2PKey(path string) (crypto.PrivKey, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate key: %w", err)
		}
		b, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			return nil, fmt.Errorf("marshal key: %w", err)
		}
		if err := os.WriteFile(path, b, 0600); err != nil {
			return nil, fmt.Errorf("save key: %w", err)
		}
		return priv, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}
	return crypto.UnmarshalPrivateKey(data)
}
