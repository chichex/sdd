package api

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/workspace"
)

func TestTrabajo_DetalleYNoEncontrado(t *testing.T) {
	root := fixtureWorkspace(t)
	_, srv := start(t, workspace.Config{Root: root, WatchInterval: time.Hour})
	waitScanned(t, srv.URL)
	var det struct {
		ID    string `json:"id"`
		Gantt []struct {
			Rama   string  `json:"rama"`
			Etapa  string  `json:"etapa"`
			Inicio string  `json:"inicio"`
			Fin    *string `json:"fin"`
			Estado string  `json:"estado"`
		} `json:"gantt"`
		Links struct {
			Issue *string `json:"issue"`
			PRs   []struct {
				Numero int    `json:"numero"`
				URL    string `json:"url"`
			} `json:"prs"`
			Grills []string `json:"grills"`
			Specs  []string `json:"specs"`
		} `json:"links"`
	}
	if code := getJSON(t, srv.URL+"/api/trabajos/"+url.PathEscape("chichex/waica/issue/203"), &det); code != http.StatusOK {
		t.Fatalf("detalle = %d, want 200", code)
	}
	if det.ID != "chichex/waica/issue/203" || len(det.Gantt) != 4 || det.Gantt[3].Rama != "run-1" || det.Gantt[2].Fin != nil {
		t.Fatalf("detalle = %+v, want 4 barras con la rama run-1 y Mergeado abierto", det)
	}
	if len(det.Links.PRs) != 5 || det.Links.PRs[0].URL != "https://github.com/chichex/waica/pull/228" || len(det.Links.Specs) != 1 {
		t.Fatalf("links = %+v", det.Links)
	}
	if det.Links.Specs[0] != filepath.Join(root, "waica", ".sdd", "specs", "issue-203.md") {
		t.Fatalf("spec local = %q, want la del árbol", det.Links.Specs[0])
	}
	var grillID struct {
		ID string `json:"id"`
	}
	if code := getJSON(t, srv.URL+"/api/trabajos/chichex/waica/issue/205", &grillID); code != http.StatusOK || grillID.ID != "chichex/waica/issue/205" {
		t.Fatalf("detalle con barras sin escapar = %d %q", code, grillID.ID)
	}
	var notFound struct {
		Error string `json:"error"`
	}
	if code := getJSON(t, srv.URL+"/api/trabajos/no-existe", &notFound); code != http.StatusNotFound || notFound.Error == "" {
		t.Fatalf("id desconocido = %d %q, want 404 con error", code, notFound.Error)
	}
}
