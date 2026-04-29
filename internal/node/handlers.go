package node

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/evermeet/evermeet/internal/events"
	"github.com/evermeet/evermeet/internal/identity"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/libp2p/go-libp2p/core/network"
)

type GossipEvent struct {
	Founding *events.FoundingDoc  `json:"founding"`
	State    *events.MutableState `json:"state"`
}

type EventFetchRequest struct {
	ID string `json:"id"`
}

type EventFetchResponse struct {
	Founding *events.FoundingDoc  `json:"founding,omitempty"`
	State    *events.MutableState `json:"state,omitempty"`
	Error    string               `json:"error,omitempty"`
}

type UserFetchRequest struct {
	DID string `json:"did"`
}

type UserFetchResponse struct {
	User  *store.User `json:"user,omitempty"`
	Error string      `json:"error,omitempty"`
}

type ProfileGossip struct {
	User *store.User `json:"user"`
	Sig  string      `json:"sig"`
}

func (n *Node) handleEventFetchStream(s network.Stream) {
	defer s.Close()

	var req EventFetchRequest
	if err := json.NewDecoder(s).Decode(&req); err != nil {
		return
	}

	res := EventFetchResponse{}
	ctx := context.Background()

	founding, err := n.db.GetEventFounding(ctx, req.ID)
	if err == nil && founding != nil {
		var fd events.FoundingDoc
		if err := json.Unmarshal([]byte(founding.Payload), &fd); err == nil {
			res.Founding = &fd
		}
	}

	state, err := n.db.GetCurrentEventState(ctx, req.ID)
	if err == nil && state != nil {
		var ms events.MutableState
		if err := json.Unmarshal([]byte(state.Payload), &ms); err == nil {
			res.State = &ms
		}
	}

	if res.Founding == nil || res.State == nil {
		res.Error = "not found"
	}

	json.NewEncoder(s).Encode(res)
}

func (n *Node) handleUserFetchStream(s network.Stream) {
	defer s.Close()

	var req UserFetchRequest
	if err := json.NewDecoder(s).Decode(&req); err != nil {
		return
	}

	res := UserFetchResponse{}
	u, err := n.db.GetUser(context.Background(), req.DID)
	if err == nil && u != nil {
		res.User = u
	} else {
		res.Error = "not found"
	}

	json.NewEncoder(s).Encode(res)
}

func (n *Node) FetchEvent(id string) error {
	peers := n.host.Network().Peers()
	if len(peers) == 0 {
		return fmt.Errorf("no peers connected")
	}

	for _, p := range peers {
		ctx, cancel := context.WithTimeout(n.ctx, 5*time.Second)
		s, err := n.host.NewStream(ctx, p, EventFetchProtocol)
		cancel()
		if err != nil {
			continue
		}

		err = json.NewEncoder(s).Encode(EventFetchRequest{ID: id})
		if err != nil {
			s.Reset()
			continue
		}

		var res EventFetchResponse
		err = json.NewDecoder(s).Decode(&res)
		s.Close()
		if err != nil {
			continue
		}

		if res.Error != "" || res.Founding == nil || res.State == nil {
			continue
		}

		// Save the received event
		n.log.Printf("fetched event %s from peer %s", id, p)
		n.handleIncomingEventRaw(res.Founding, res.State)
		return nil
	}

	return fmt.Errorf("event not found on network")
}

func (n *Node) FetchUser(did string) error {
	peers := n.host.Network().Peers()
	for _, p := range peers {
		ctx, cancel := context.WithTimeout(n.ctx, 5*time.Second)
		s, err := n.host.NewStream(ctx, p, UserFetchProtocol)
		cancel()
		if err != nil {
			continue
		}

		err = json.NewEncoder(s).Encode(UserFetchRequest{DID: did})
		if err != nil {
			s.Reset()
			continue
		}

		var res UserFetchResponse
		err = json.NewDecoder(s).Decode(&res)
		s.Close()
		if err != nil {
			continue
		}

		if res.Error != "" || res.User == nil {
			continue
		}

		// Save the received user
		n.log.Printf("fetched user %s from peer %s", did, p)
		return n.db.UpsertUser(context.Background(), res.User)
	}
	return fmt.Errorf("user not found on network")
}

func (n *Node) handleIncomingEventRaw(founding *events.FoundingDoc, state *events.MutableState) {
	data, _ := json.Marshal(GossipEvent{Founding: founding, State: state})
	n.handleIncomingEvent(data)
}

func (n *Node) readEventsLoop() {
	for {
		msg, err := n.eventsSub.Next(n.ctx)
		if err != nil {
			if n.ctx.Err() == nil {
				n.log.Printf("pubsub error: %v", err)
			}
			return
		}

		// Don't process our own messages
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		go n.handleIncomingEvent(msg.Data)
	}
}

func (n *Node) handleIncomingEvent(data []byte) {
	var ge GossipEvent
	if err := json.Unmarshal(data, &ge); err != nil {
		n.log.Printf("failed to unmarshal gossip event: %v", err)
		return
	}

	if ge.Founding == nil || ge.State == nil {
		return
	}

	// Verify the event record
	err := events.Verify(ge.State, func(did string) (ed25519.PublicKey, error) {
		u, err := n.db.GetUser(context.Background(), did)
		if (err != nil || u == nil) && n != nil {
			// Try fetching the user from network
			if err := n.FetchUser(did); err == nil {
				u, err = n.db.GetUser(context.Background(), did)
			}
		}
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, fmt.Errorf("unknown user: %s", did)
		}
		pkBytes, err := hex.DecodeString(u.CurrentPK)
		if err != nil {
			return nil, fmt.Errorf("invalid public key hex: %w", err)
		}
		return ed25519.PublicKey(pkBytes), nil
	})

	if err != nil {
		n.log.Printf("invalid event received via P2P (%s): %v", ge.State.ID, err)
		return
	}

	n.log.Printf("received event via P2P: %s (%s)", ge.State.Title, ge.State.ID)

	// Save founding doc
	foundingPayload, _ := json.Marshal(ge.Founding)
	err = n.db.InsertEventFounding(context.Background(), &store.EventFounding{
		ID:      ge.State.ID,
		Payload: string(foundingPayload),
	})
	if err != nil {
		n.log.Printf("failed to save founding doc: %v", err)
	}

	// Save mutable state
	statePayload, _ := json.Marshal(ge.State)
	stateHash, _ := identity.ContentHash(ge.State)
	updatedAt, _ := time.Parse(time.RFC3339, ge.State.UpdatedAt)

	err = n.db.AppendEventState(context.Background(), &store.EventState{
		Hash:      stateHash,
		ID:        ge.State.ID,
		Prev:      ge.State.Prev,
		Payload:   string(statePayload),
		IsCurrent: true,
		CreatedAt: updatedAt,
	})
	if err != nil {
		n.log.Printf("failed to save event state: %v", err)
	}
}

func (n *Node) readUsersLoop() {
	for {
		msg, err := n.usersSub.Next(n.ctx)
		if err != nil {
			if n.ctx.Err() == nil {
				n.log.Printf("users pubsub error: %v", err)
			}
			return
		}

		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		go n.handleIncomingUser(msg.Data)
	}
}

func (n *Node) handleIncomingUser(data []byte) {
	var pg ProfileGossip
	if err := json.Unmarshal(data, &pg); err != nil {
		return
	}

	u := pg.User
	if u == nil {
		return
	}

	// Verify signature
	pkBytes, err := hex.DecodeString(u.CurrentPK)
	if err != nil {
		return
	}
	valid, err := identity.Verify(ed25519.PublicKey(pkBytes), u, pg.Sig)
	if err != nil || !valid {
		n.log.Printf("invalid profile signature from %s", u.DID)
		return
	}

	// Check if newer
	existing, err := n.db.GetUser(context.Background(), u.DID)
	if err == nil && existing != nil && !u.UpdatedAt.After(existing.UpdatedAt) {
		return
	}

	n.log.Printf("received profile update via P2P: %s (%s)", u.DisplayName, u.DID)
	n.db.UpsertUser(context.Background(), u)
}

func (n *Node) BroadcastUserProfile(u *store.User, priv ed25519.PrivateKey) error {
	sig, err := identity.Sign(priv, u)
	if err != nil {
		return err
	}

	data, err := json.Marshal(ProfileGossip{User: u, Sig: sig})
	if err != nil {
		return err
	}

	return n.BroadcastUser(data)
}
