package node

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/evermeet/evermeet/internal/store"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const (
	EventsTopicName    = "em/events"
	EventFetchProtocol = "/evermeet/event-fetch/1.0.0"
	UserFetchProtocol  = "/evermeet/user-fetch/1.0.0"
)

type Node struct {
	host    host.Host
	ps      *pubsub.PubSub
	db      *store.DB
	log     *log.Logger
	ctx     context.Context
	cancel  context.CancelFunc

	eventsTopic *pubsub.Topic
	eventsSub   *pubsub.Subscription

	peersMu sync.Mutex
	peers   map[peer.ID]peer.AddrInfo
}

func New(db *store.DB, l *log.Logger, listenPort int) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// libp2p host with default security/muxing for maximum compatibility
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
			fmt.Sprintf("/ip6/::/tcp/%d", listenPort),
		),
		// In a real app, we'd use a persistent identity here
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("libp2p host: %w", err)
	}

	// GossipSub
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("gossipsub: %w", err)
	}

	n := &Node{
		host:   h,
		ps:     ps,
		db:     db,
		log:    l,
		ctx:    ctx,
		cancel: cancel,
		peers:  make(map[peer.ID]peer.AddrInfo),
	}

	// Setup mDNS discovery
	ser := mdns.NewMdnsService(h, "evermeet", n)
	if err := ser.Start(); err != nil {
		l.Printf("mdns error: %v", err)
	}

	// Join events topic
	if err := n.joinEvents(); err != nil {
		return nil, err
	}

	// Register P2P RPC handlers
	n.host.SetStreamHandler(EventFetchProtocol, n.handleEventFetchStream)
	n.host.SetStreamHandler(UserFetchProtocol, n.handleUserFetchStream)

	l.Printf("P2P node started: %s", h.ID())
	for _, addr := range h.Addrs() {
		l.Printf("  Listening on %s/p2p/%s", addr, h.ID())
	}

	return n, nil
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

func (n *Node) BroadcastEvent(data []byte) error {
	if n.eventsTopic == nil {
		return fmt.Errorf("not joined to events topic")
	}
	return n.eventsTopic.Publish(n.ctx, data)
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
		// Small delay to avoid simultaneous open race on localhost
		time.Sleep(time.Duration(100+ (pi.ID[0] % 10)) * time.Millisecond)

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
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
	Peers     []PeerStatus `json:"peers"`
}

type PeerStatus struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
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
		peerStats = append(peerStats, PeerStatus{
			ID:        p.String(),
			Addresses: pAddrs,
		})
	}

	return NodeStatus{
		ID:        n.host.ID().String(),
		Addresses: addrs,
		Peers:     peerStats,
	}
}

func (n *Node) Close() error {
	n.cancel()
	return n.host.Close()
}
