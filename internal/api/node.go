package api

import (
	"net/http"
)

func (s *Server) handleNodeStatus(w http.ResponseWriter, r *http.Request) {
	if s.node == nil {
		jsonErr(w, http.StatusNotFound, "node not initialized")
		return
	}
	jsonOK(w, s.node.Status())
}
