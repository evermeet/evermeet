package api

import (
	"encoding/json"
	"net/http"

	"github.com/evermeet/evermeet/internal/eventimport"
)

func (s *Server) handleImportEventPreview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.URL == "" {
		jsonErr(w, http.StatusBadRequest, "url required")
		return
	}

	preview, err := eventimport.DefaultManager().Preview(r.Context(), req.URL)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, preview)
}
