package api

import (
	"net/http"
	"time"
)

func (s *Server) handleInstanceStatus(w http.ResponseWriter, r *http.Request) {
	if s.node == nil {
		jsonErr(w, http.StatusNotFound, "node not initialized")
		return
	}

	uptime := time.Since(s.startTime)

	safeConfig := map[string]any{
		"node": s.cfg.Node,
		"p2p":  s.cfg.P2P,
	}

	jsonOK(w, map[string]any{
		"instance_id": s.homeHost(),
		"version":     "v0.1.0-alpha",
		"started_at":  s.startTime.UTC().Format(time.RFC3339),
		"uptime_s":    int(uptime.Seconds()),
		"uptime":      formatUptime(uptime),
		"config":      safeConfig,
		"p2p":         s.node.Status(),
	})
}

func formatUptime(d time.Duration) string {
	d = d.Round(time.Second)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60
	if days > 0 {
		return time.Duration(d).String() // fallback for complex durations
	}
	if hours > 0 {
		return time.Duration(time.Duration(hours)*time.Hour + time.Duration(mins)*time.Minute + time.Duration(secs)*time.Second).String()
	}
	if mins > 0 {
		return time.Duration(time.Duration(mins)*time.Minute + time.Duration(secs)*time.Second).String()
	}
	return time.Duration(time.Duration(secs) * time.Second).String()
}
