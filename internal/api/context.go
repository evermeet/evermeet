package api

import (
	"context"
	"crypto/ed25519"
	"net/http"
)

type contextKey int

const (
	ctxDID contextKey = iota
	ctxPrivKey
)

// withAuth attaches the authenticated DID and private key to the request context.
func withAuth(r *http.Request, did string, priv ed25519.PrivateKey) *http.Request {
	ctx := context.WithValue(r.Context(), ctxDID, did)
	ctx = context.WithValue(ctx, ctxPrivKey, priv)
	return r.WithContext(ctx)
}

// authDID returns the DID of the authenticated user, or "" if unauthenticated.
func authDID(r *http.Request) string {
	v, _ := r.Context().Value(ctxDID).(string)
	return v
}

// authPrivKey returns the private key of the authenticated user.
func authPrivKey(r *http.Request) ed25519.PrivateKey {
	v, _ := r.Context().Value(ctxPrivKey).(ed25519.PrivateKey)
	return v
}
