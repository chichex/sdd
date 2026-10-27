package api

import (
	"net/http"
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
