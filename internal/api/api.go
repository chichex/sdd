// Package api sirve la API JSON del visor sobre net/http: el workspace con sus
// trabajos y el detalle de un trabajo con su Gantt y sus links.
package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chichex/sdd/internal/model"
	"github.com/chichex/sdd/internal/workspace"
)

// Source es de donde la API lee el estado.
type Source interface {
	Snapshot() workspace.Snapshot
}

// Server sirve /api/*.
type Server struct {
	src     Source
	version string
}

// New crea el server de la API.
func New(src Source, version string) *Server { return &Server{src: src, version: version} }

// Register monta las rutas de la API en mux.
func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/workspace", s.workspace)
	mux.HandleFunc("GET /api/trabajos/{id...}", s.trabajo)
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
