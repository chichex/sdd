package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/scan"
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

func TestWorkspace_FixtureCompleto(t *testing.T) {
	root := fixtureWorkspace(t)
	_, srv := start(t, workspace.Config{Root: root, WatchInterval: time.Hour})
	ws := waitScanned(t, srv.URL)
	if ws.Root != root || ws.ScannedAt == nil || len(ws.Repos) != 6 || len(ws.Trabajos) != 12 {
		t.Fatalf("workspace = root %q scannedAt %v repos %d trabajos %d; want 6 repos y 12 trabajos", ws.Root, ws.ScannedAt, len(ws.Repos), len(ws.Trabajos))
	}
	for i := 1; i < len(ws.Trabajos); i++ {
		if ws.Trabajos[i-1].UltimaActividad < ws.Trabajos[i].UltimaActividad {
			t.Fatalf("trabajos no ordenados por ultimaActividad descendente")
		}
	}
	var waica workspaceResp
	getJSON(t, srv.URL+"/api/workspace?repo="+url.QueryEscape("chichex/waica"), &waica)
	if len(waica.Trabajos) != 2 {
		t.Fatalf("?repo=chichex/waica = %d trabajos, want 2", len(waica.Trabajos))
	}
	for _, tr := range waica.Trabajos {
		if tr.ID == "chichex/waica/issue/203" && (tr.Titulo != "Sombras 3D" || tr.Etapa != "mergeado") {
			t.Fatalf("#203 = %+v, want título de la spec del árbol y etapa mergeado", tr)
		}
	}
	for _, r := range ws.Repos {
		if r.ID == "chichex/waica" && len(r.Rutas) != 2 {
			t.Fatalf("waica rutas = %v", r.Rutas)
		}
	}
}

func TestWorkspace_FiltroActivos(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no está en PATH")
	}
	root := t.TempDir()
	repo := filepath.Join(root, "r")
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	for _, p := range []string{filepath.Join(repo, ".sdd"), filepath.Join(repo, ".git", "sdd")} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(repo, ".sdd", "project.md"), []byte("# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	log := `{"v":1,"ts":"2026-10-01T10:00:00Z","harness":"claude","event":"run.pr-opened","issue":"none","spec":".sdd%2Fspecs%2Fa.md","previous":"none","grill":"none","pr":[7],"reason":"none"}` + "\n" +
		`{"v":1,"ts":"2026-10-02T10:00:00Z","harness":"claude","event":"land.merged","issue":"none","spec":"none","previous":"none","grill":"none","pr":[7],"reason":"none"}` + "\n" +
		`{"v":1,"ts":"2026-10-03T10:00:00Z","harness":"claude","event":"spec.published","issue":"none","spec":".sdd%2Fspecs%2Fb.md","previous":"none","grill":"none","pr":[],"reason":"none"}` + "\n"
	if err := os.WriteFile(filepath.Join(repo, ".git", "sdd", "events.jsonl"), []byte(log), 0o644); err != nil {
		t.Fatal(err)
	}
	_, srv := start(t, workspace.Config{Root: root, WatchInterval: time.Hour})
	all := waitScanned(t, srv.URL)
	var activos workspaceResp
	getJSON(t, srv.URL+"/api/workspace", &activos)
	if len(all.Trabajos) != 2 || len(activos.Trabajos) != 1 || activos.Trabajos[0].ID != "r/spec/.sdd%2Fspecs%2Fb.md" {
		t.Fatalf("todos = %d, activos = %+v; want 2 y solo la spec b (a está mergeada sin issue: terminado)", len(all.Trabajos), activos.Trabajos)
	}
	if code := getJSON(t, srv.URL+"/api/workspace?filtro=otro", nil); code != http.StatusBadRequest {
		t.Fatalf("filtro inválido = %d, want 400", code)
	}
}

func TestWorkspace_ParcialMientrasEscanea(t *testing.T) {
	root := fixtureWorkspace(t)
	release := make(chan struct{})
	var once sync.Once
	read := func(r scan.Repo) scan.Data {
		if r.ID == "chichex/waica" {
			<-release
		}
		return scan.Read(r)
	}
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	_, srv := start(t, workspace.Config{Root: root, WatchInterval: time.Hour, ReadRepo: read})
	deadline := time.Now().Add(10 * time.Second)
	for {
		var ws workspaceResp
		getJSON(t, srv.URL+"/api/workspace?filtro=todos", &ws)
		if ws.Scanning && len(ws.Repos) == 6 && len(ws.Trabajos) == 1 {
			break // skills (1 trabajo) leído; waica bloqueado; los demás repos están después
		}
		if time.Now().After(deadline) {
			t.Fatalf("no vi el parcial: scanning %v repos %d trabajos %d", ws.Scanning, len(ws.Repos), len(ws.Trabajos))
		}
		time.Sleep(20 * time.Millisecond)
	}
	once.Do(func() { close(release) })
	if ws := waitScanned(t, srv.URL); len(ws.Trabajos) != 12 {
		t.Fatalf("al terminar = %d trabajos, want 12", len(ws.Trabajos))
	}
}

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

func TestWorkspace_Vacio(t *testing.T) {
	_, srv := start(t, workspace.Config{Root: t.TempDir(), WatchInterval: time.Hour})
	ws := waitScanned(t, srv.URL)
	if ws.Repos == nil || ws.Trabajos == nil || len(ws.Repos) != 0 || len(ws.Trabajos) != 0 || ws.Scanning {
		t.Fatalf("vacío = %+v, want repos [] trabajos [] scanning false", ws)
	}
	var health struct {
		OK bool `json:"ok"`
	}
	if code := getJSON(t, srv.URL+"/api/health", &health); code != http.StatusOK || !health.OK {
		t.Fatalf("health = %d %v, want el server sirviendo", code, health.OK)
	}
}

func TestWorkspace_RefreshLocalEnVivo(t *testing.T) {
	root := fixtureWorkspace(t)
	_, srv := start(t, workspace.Config{Root: root, WatchInterval: workspace.DefaultWatchInterval})
	before := waitScanned(t, srv.URL)
	f, err := os.OpenFile(filepath.Join(root, "waica", ".git", "sdd", "events.jsonl"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	line := `{"v":1,"ts":"2026-10-27T12:00:00Z","harness":"claude","event":"grill.paused","issue":"#300","spec":"none","previous":"none","grill":"2026-10-27-nuevo","pr":[],"reason":"none"}` + "\n"
	if _, err := f.WriteString(line); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		var ws workspaceResp
		getJSON(t, srv.URL+"/api/workspace?filtro=todos", &ws)
		ids := make([]string, 0, len(ws.Trabajos))
		for _, tr := range ws.Trabajos {
			ids = append(ids, tr.ID)
		}
		if slices.Contains(ids, "chichex/waica/issue/300") {
			if *ws.ScannedAt <= *before.ScannedAt {
				t.Fatalf("scannedAt no avanzó: %s → %s", *before.ScannedAt, *ws.ScannedAt)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("el nodo nuevo no apareció en 10 s")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
