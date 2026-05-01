// Package routing implements the DHT-based home-instance discovery protocol.
//
// Privacy model: emails are never published to the network. Only their
// blake3 hash (with a fixed domain separator) is used as a DHT key. The
// value is a signed JSON record that maps the hash to a home instance URL.
// An observer watching DHT traffic learns neither the email address nor the
// user's DID — only that some instance claims to be home for this hash.
package routing

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"lukechampine.com/blake3"
)

// domain is the fixed domain separator used in all email hashes.
// It is a public constant — its purpose is domain separation, not secrecy.
const domain = "evermeet-home-routing-v1"

// HomeRecord is the signed value stored in the DHT under an email hash.
// It is signed by the home instance's Ed25519 key (the libp2p node key),
// so a foreign instance can verify the record without trusting the DHT.
type HomeRecord struct {
	HomeInstanceURL string `json:"home_instance_url"`
	Timestamp       int64  `json:"timestamp"` // Unix seconds — newer wins on conflict
	Sig             string `json:"sig"`        // Ed25519 sig over canonical fields, base64url
}

// EmailHash returns the DHT key for a given email address.
// blake3(domain + ":" + email), hex-encoded.
func EmailHash(email string) []byte {
	h := blake3.Sum256([]byte(domain + ":" + email))
	dst := make([]byte, hex.EncodedLen(len(h)))
	hex.Encode(dst, h[:])
	return dst
}

// NewRecord builds and signs a HomeRecord for the given home instance URL
// using the instance's Ed25519 private key.
func NewRecord(homeURL string, instancePriv ed25519.PrivateKey) (*HomeRecord, error) {
	r := &HomeRecord{
		HomeInstanceURL: homeURL,
		Timestamp:       time.Now().Unix(),
	}
	payload, err := sigPayload(r)
	if err != nil {
		return nil, err
	}
	sig := ed25519.Sign(instancePriv, payload)
	r.Sig = base64.RawURLEncoding.EncodeToString(sig)
	return r, nil
}

// MarshalRecord encodes a HomeRecord to JSON bytes for DHT storage.
func MarshalRecord(r *HomeRecord) ([]byte, error) {
	return json.Marshal(r)
}

// UnmarshalRecord decodes DHT bytes into a HomeRecord.
func UnmarshalRecord(data []byte) (*HomeRecord, error) {
	var r HomeRecord
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("unmarshal home record: %w", err)
	}
	return &r, nil
}

// VerifyRecord checks the signature on a HomeRecord against the provided
// instance public key. Call this after fetching the instance's public key
// from its /.well-known/evermeet-node-key endpoint.
func VerifyRecord(r *HomeRecord, instancePub ed25519.PublicKey) error {
	payload, err := sigPayload(r)
	if err != nil {
		return err
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(r.Sig)
	if err != nil {
		return fmt.Errorf("decode sig: %w", err)
	}
	if !ed25519.Verify(instancePub, payload, sigBytes) {
		return fmt.Errorf("invalid home record signature")
	}
	return nil
}

// Publisher publishes and re-publishes home records to the DHT.
type Publisher struct {
	publish func(ctx context.Context, key, value []byte) error
	priv    ed25519.PrivateKey
	homeURL string
}

// NewPublisher creates a Publisher. publish is the DHT put function
// (node.DHTPublish), priv is the instance's Ed25519 private key,
// homeURL is this instance's base URL.
func NewPublisher(
	publish func(ctx context.Context, key, value []byte) error,
	priv ed25519.PrivateKey,
	homeURL string,
) *Publisher {
	return &Publisher{publish: publish, priv: priv, homeURL: homeURL}
}

// Publish publishes the email → homeURL mapping for a single email.
func (p *Publisher) Publish(ctx context.Context, email string) error {
	rec, err := NewRecord(p.homeURL, p.priv)
	if err != nil {
		return err
	}
	data, err := MarshalRecord(rec)
	if err != nil {
		return err
	}
	return p.publish(ctx, EmailHash(email), data)
}

// StartHeartbeat launches a goroutine that re-publishes all emails every
// interval. DHT records have a ~24h TTL; re-publish every 12h to be safe.
// emailsFn is called each tick to get the current list of emails to publish.
// Stops when ctx is cancelled.
func (p *Publisher) StartHeartbeat(ctx context.Context, interval time.Duration, emailsFn func(ctx context.Context) ([]string, error)) {
	go func() {
		// Initial publish after a short delay to let the DHT routing table
		// populate via mDNS before we try to store records.
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}
		p.publishAll(ctx, emailsFn)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.publishAll(ctx, emailsFn)
			}
		}
	}()
}

func (p *Publisher) publishAll(ctx context.Context, emailsFn func(ctx context.Context) ([]string, error)) {
	emails, err := emailsFn(ctx)
	if err != nil {
		return
	}
	for _, email := range emails {
		if ctx.Err() != nil {
			return
		}
		_ = p.Publish(ctx, email)
	}
}

// sigPayload returns the bytes that are signed/verified for a HomeRecord.
// Only the stable fields are included — Sig is excluded.
func sigPayload(r *HomeRecord) ([]byte, error) {
	return json.Marshal(struct {
		HomeInstanceURL string `json:"home_instance_url"`
		Timestamp       int64  `json:"timestamp"`
	}{r.HomeInstanceURL, r.Timestamp})
}
