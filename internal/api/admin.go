package api

import (
	"net/http"
	"time"
)

func (s *Server) handleAdminOverview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	adminCount, err := s.db.CountAdminAccounts(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count admins failed")
		return
	}
	userCount, err := s.db.CountUsers(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count users failed")
		return
	}
	eventCount, err := s.db.CountCurrentEvents(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count events failed")
		return
	}
	calendarCount, err := s.db.CountCalendars(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count calendars failed")
		return
	}
	blobCount, err := s.db.CountBlobs(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count blobs failed")
		return
	}

	uptime := time.Since(s.startTime)
	var p2pStatus any = map[string]any{}
	if s.node != nil {
		p2pStatus = s.node.Status()
	}

	jsonOK(w, map[string]any{
		"instance_id": s.homeHost(),
		"base_url":    s.baseURL,
		"version":     "v0.1.0-alpha",
		"started_at":  s.startTime.UTC().Format(time.RFC3339),
		"uptime":      formatUptime(uptime),
		"counts": map[string]int{
			"admins":    adminCount,
			"users":     userCount,
			"events":    eventCount,
			"calendars": calendarCount,
			"blobs":     blobCount,
		},
		"p2p": p2pStatus,
		"config": map[string]any{
			"node": s.cfg.Node,
			"p2p":  s.cfg.P2P,
		},
	})
}
