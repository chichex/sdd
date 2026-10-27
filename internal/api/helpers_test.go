package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/workspace"
)

func fixtureWorkspace(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no está en PATH: los tests de la API necesitan repos temporales")
	}
	dir := filepath.Join(t.TempDir(), "ws")
	if out, err := exec.Command("bash", "../../testdata/mkworkspace.sh", dir).CombinedOutput(); err != nil {
		t.Fatalf("mkworkspace: %v\n%s", err, out)
	}
	return dir
}

// start levanta el servicio real sobre root y un httptest.Server con la API.
func start(t *testing.T, cfg workspace.Config) (*workspace.Service, *httptest.Server) {
	t.Helper()
	svc := workspace.New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = svc.Run(ctx) // termina al cancelar el contexto
	}()
	mux := http.NewServeMux()
	New(svc, "test").Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(func() {
		srv.Close()
		cancel()
		wg.Wait()
	})
	return svc, srv
}

type workspaceResp struct {
	Root      string  `json:"root"`
	ScannedAt *string `json:"scannedAt"`
	Scanning  bool    `json:"scanning"`
	GitHub    struct {
		OK        bool    `json:"ok"`
		Reason    string  `json:"reason"`
		FetchedAt *string `json:"fetchedAt"`
	} `json:"github"`
	Repos []struct {
		ID              string   `json:"id"`
		Nombre          string   `json:"nombre"`
		Rutas           []string `json:"rutas"`
		SinGit          bool     `json:"sinGit"`
		Remote          string   `json:"remote"`
		LineasInvalidas int      `json:"lineasInvalidas"`
		Errores         []string `json:"errores"`
	} `json:"repos"`
	Trabajos []struct {
		ID              string `json:"id"`
		Repo            string `json:"repo"`
		Titulo          string `json:"titulo"`
		Etapa           string `json:"etapa"`
		SubEstado       string `json:"subEstado"`
		UltimaActividad string `json:"ultimaActividad"`
		Terminado       bool   `json:"terminado"`
		Ramas           []struct {
			ID    string `json:"id"`
			Nodos []struct {
				Etapa     string `json:"etapa"`
				SubEstado string `json:"subEstado"`
				TS        string `json:"ts"`
			} `json:"nodos"`
		} `json:"ramas"`
	} `json:"trabajos"`
}

func getJSON(t *testing.T, u string, into any) int {
	t.Helper()
	res, err := http.Get(u)
	if err != nil {
		t.Fatalf("GET %s: %v", u, err)
	}
	defer res.Body.Close()
	if into != nil {
		if err := json.NewDecoder(res.Body).Decode(into); err != nil {
			t.Fatalf("GET %s: json: %v", u, err)
		}
	}
	return res.StatusCode
}

func waitScanned(t *testing.T, base string) workspaceResp {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var ws workspaceResp
		getJSON(t, base+"/api/workspace?filtro=todos", &ws)
		if !ws.Scanning {
			return ws
		}
		if time.Now().After(deadline) {
			t.Fatalf("el escaneo inicial no terminó en 10 s")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
