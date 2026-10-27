// Package api sirve la API JSON del visor sobre net/http: el workspace con sus
// trabajos y el detalle de un trabajo con su Gantt y sus links.
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/chichex/sdd/internal/model"
	"github.com/chichex/sdd/internal/workspace"
)

// Source es de donde la API lee el estado: la foto actual, los avisos de
// cambio y el refresh manual.
type Source interface {
	Snapshot() workspace.Snapshot
	Subscribe() (<-chan struct{}, func())
	Refresh(ctx context.Context) string
}

// DefaultPing es cada cuánto el stream SSE manda un comentario de vida.
const DefaultPing = 30 * time.Second

// Server sirve /api/*.
type Server struct {
	src     Source
	version string
	ping    time.Duration
}

// New crea el server de la API.
func New(src Source, version string) *Server {
	return &Server{src: src, version: version, ping: DefaultPing}
}

// Register monta las rutas de la API en mux.
func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/workspace", s.workspace)
	mux.HandleFunc("GET /api/trabajos/{id...}", s.trabajo)
	mux.HandleFunc("POST /api/refresh", s.refresh)
	mux.HandleFunc("GET /api/stream", s.stream)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("api: json:", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "version": s.version, "root": s.src.Snapshot().Root})
}

func (s *Server) workspace(w http.ResponseWriter, r *http.Request) {
	filtro := r.URL.Query().Get("filtro")
	if filtro == "" {
		filtro = "activos"
	}
	if filtro != "activos" && filtro != "todos" {
		writeError(w, http.StatusBadRequest, "filtro debe ser activos o todos")
		return
	}
	snap := s.src.Snapshot()
	repo := r.URL.Query().Get("repo")
	trabajos := make([]model.Trabajo, 0, len(snap.Trabajos))
	for _, tr := range snap.Trabajos {
		if (repo == "" || tr.Repo == repo) && (filtro == "todos" || !tr.Terminado) {
			trabajos = append(trabajos, tr)
		}
	}
	writeJSON(w, http.StatusOK, workspaceView(snap, trabajos))
}

func (s *Server) trabajo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, tr := range s.src.Snapshot().Trabajos {
		if tr.ID == id {
			writeJSON(w, http.StatusOK, detalleView(tr))
			return
		}
	}
	writeError(w, http.StatusNotFound, "trabajo no encontrado: "+id)
}

// refresh consulta GitHub ya (fusionándose con una consulta en vuelo) y
// responde 202 con el fetchedAt resultante cuando termina.
func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusAccepted, map[string]*string{"fetchedAt": optional(s.src.Refresh(r.Context()))})
}
