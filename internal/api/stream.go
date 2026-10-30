package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// stream sirve SSE: `event: snapshot` con {scannedAt, fetchedAt} cada vez que
// cambia el modelo y un comentario `: ping` cada s.ping para mantener viva la
// conexión. Termina cuando el cliente se va.
func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	changes, unsubscribe := s.src.Subscribe()
	defer unsubscribe()
	if _, err := fmt.Fprint(w, ": conectado\n\n"); err != nil || rc.Flush() != nil {
		return
	}
	ping := time.NewTicker(s.ping)
	defer ping.Stop()
	for {
		var msg string
		select {
		case <-r.Context().Done():
			return
		case <-ping.C:
			msg = ": ping\n\n"
		case <-changes:
			snap := s.src.Snapshot()
			data, err := json.Marshal(map[string]*string{"scannedAt": optional(snap.ScannedAt), "fetchedAt": optional(snap.GitHub.FetchedAt)})
			if err != nil {
				return
			}
			msg = "event: snapshot\ndata: " + string(data) + "\n\n"
		}
		if _, err := fmt.Fprint(w, msg); err != nil || rc.Flush() != nil {
			return
		}
	}
}
