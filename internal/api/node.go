package api

import (
	"net/http"
)

func (s *Server) handleInstanceStatus(w http.ResponseWriter, r *http.Request) {
	if s.node == nil {
		jsonErr(w, http.StatusNotFound, "node not initialized")
		return
	}
	
	// Create a safe copy of config omitting Email
	safeConfig := map[string]any{
		"node":     s.cfg.Node,
		"identity": s.cfg.Identity,
		"p2p":      s.cfg.P2P,
	}

	jsonOK(w, map[string]any{
		"home_host": s.baseURL,
		"version":   "v0.1.0-alpha",
		"config":    safeConfig,
		"p2p":       s.node.Status(),
	})
}
