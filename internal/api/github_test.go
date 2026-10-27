package api

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/github"
	"github.com/chichex/sdd/internal/github/fake"
	"github.com/chichex/sdd/internal/workspace"
)

type ghResp struct {
	GitHub struct {
		OK        bool    `json:"ok"`
		Reason    *string `json:"reason"`
		FetchedAt *string `json:"fetchedAt"`
	} `json:"github"`
	Repos []struct {
		ID     string `json:"id"`
		GitHub struct {
			OK     bool    `json:"ok"`
			Reason *string `json:"reason"`
		} `json:"github"`
	} `json:"repos"`
	Trabajos []struct {
		ID    string `json:"id"`
		Ramas []struct {
			Nodos []struct {
				Etapa     string `json:"etapa"`
				SubEstado string `json:"subEstado"`
				Stale     bool   `json:"stale"`
				Refs      struct {
					PR *int `json:"pr"`
				} `json:"refs"`
			} `json:"nodos"`
		} `json:"ramas"`
	} `json:"trabajos"`
}

func waitGitHub(t *testing.T, base string, done func(ghResp) bool) ghResp {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var ws ghResp
		getJSON(t, base+"/api/workspace?filtro=todos", &ws)
		if done(ws) {
			return ws
		}
		if time.Now().After(deadline) {
			t.Fatalf("estado de GitHub no alcanzado: %+v", ws.GitHub)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func prNode(ws ghResp, trabajo string, pr int) (string, bool, bool) {
	for _, tr := range ws.Trabajos {
		if tr.ID != trabajo {
			continue
		}
		for _, r := range tr.Ramas {
			for _, n := range r.Nodos {
				if n.Etapa == "pr" && n.Refs.PR != nil && *n.Refs.PR == pr {
					return n.SubEstado, n.Stale, true
				}
			}
		}
	}
	return "", false, false
}

func ghConfig(root string, gh github.Runner, cache string) workspace.Config {
	return workspace.Config{Root: root, WatchInterval: time.Hour, GitHub: gh, GitHubInterval: time.Hour, CacheDir: cache}
}

func TestRefresh_ConsultaYResponde202(t *testing.T) {
	root := fixtureWorkspace(t)
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh"}
	_, srv := start(t, ghConfig(root, gh, t.TempDir()))
	waitGitHub(t, srv.URL, func(ws ghResp) bool { return ws.GitHub.FetchedAt != nil })
	if gh.CountFor("waica") != 1 {
		t.Fatalf("consultas a waica tras el arranque = %d, want 1", gh.CountFor("waica"))
	}
	res, err := http.Post(srv.URL+"/api/refresh", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		FetchedAt *string `json:"fetchedAt"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil || res.StatusCode != http.StatusAccepted || body.FetchedAt == nil {
		t.Fatalf("POST /api/refresh = %d %+v %v, want 202 con fetchedAt", res.StatusCode, body, err)
	}
	if gh.CountFor("waica") != 2 || gh.CountFor("platform") != 2 || gh.CountFor("sin-remote") != 0 {
		t.Fatalf("consultas tras el refresh: waica %d platform %d; want 2 y 2", gh.CountFor("waica"), gh.CountFor("platform"))
	}
	ws := waitGitHub(t, srv.URL, func(ws ghResp) bool { return true })
	if sub, stale, ok := prNode(ws, "pramaestudio/platform/issue/56", 111); !ok || sub != "checks rojos" || stale {
		t.Fatalf("PR 111 = %q stale %v, want checks rojos fresco", sub, stale)
	}
}

func TestCache_ArrancaSinGhYSirveLoCacheado(t *testing.T) {
	root := fixtureWorkspace(t)
	cache := t.TempDir()
	cached, err := github.Fetch(context.Background(), &fake.Runner{T: t, Dir: "../../testdata/gh"}, "chichex/waica", time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := github.SaveCache(cache, cached); err != nil {
		t.Fatal(err)
	}
	gate := make(chan struct{})
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh", NotFound: true, Gate: gate}
	_, srv := start(t, ghConfig(root, gh, cache))
	ws := waitGitHub(t, srv.URL, func(ws ghResp) bool { return ws.GitHub.FetchedAt != nil })
	if *ws.GitHub.FetchedAt != "2026-10-20T10:00:00Z" || !hasTrabajo(ws, "chichex/waica/issue/210") {
		t.Fatalf("mientras corre la primera consulta = %+v, want la caché con su fetchedAt y la idea #210", ws.GitHub)
	}
	close(gate)
	ws = waitGitHub(t, srv.URL, func(ws ghResp) bool { return !ws.GitHub.OK })
	if *ws.GitHub.Reason != "sin-gh" || *ws.GitHub.FetchedAt != "2026-10-20T10:00:00Z" || !hasTrabajo(ws, "chichex/waica/issue/210") {
		t.Fatalf("sin gh = %+v, want reason sin-gh, fetchedAt de la caché e idea cacheada", ws.GitHub)
	}
	if _, err := os.Stat(github.CachePath(cache, "chichex/waica")); err != nil {
		t.Fatalf("la caché desapareció: %v", err)
	}
}

func hasTrabajo(ws ghResp, id string) bool {
	for _, tr := range ws.Trabajos {
		if tr.ID == id {
			return true
		}
	}
	return false
}

func TestDegradado_UnRepoFallaYLosDemasNo(t *testing.T) {
	root := fixtureWorkspace(t)
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh", Fail: map[string]string{"waica": "dial tcp: lookup api.github.com: i/o timeout"}}
	_, srv := start(t, ghConfig(root, gh, ""))
	ws := waitGitHub(t, srv.URL, func(ws ghResp) bool { return ws.GitHub.FetchedAt != nil })
	if !ws.GitHub.OK {
		t.Fatalf("github global = %+v, want ok: un repo caído no tumba a los demás", ws.GitHub)
	}
	reasons := map[string]string{}
	for _, r := range ws.Repos {
		if r.GitHub.Reason != nil {
			reasons[r.ID] = *r.GitHub.Reason
		}
	}
	if reasons["chichex/waica"] != "red" || reasons["sin-remote"] != "sin-remote" || reasons["pramaestudio/platform"] != "" {
		t.Fatalf("razones por repo = %v", reasons)
	}
	if sub, stale, _ := prNode(ws, "chichex/waica/issue/205", 239); sub != "sin dato" || !stale {
		t.Fatalf("PR 239 de waica = %q stale %v, want sin dato y stale", sub, stale)
	}
	if sub, stale, _ := prNode(ws, "pramaestudio/platform/issue/57", 108); sub != "cambios pedidos" || stale {
		t.Fatalf("PR 108 de platform = %q stale %v, want cambios pedidos fresco", sub, stale)
	}
}

func TestDegradado_SinGhSirveLoLocal(t *testing.T) {
	root := fixtureWorkspace(t)
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh", NotFound: true}
	_, srv := start(t, ghConfig(root, gh, ""))
	ws := waitGitHub(t, srv.URL, func(ws ghResp) bool { return !ws.GitHub.OK && ws.GitHub.Reason != nil })
	if *ws.GitHub.Reason != "sin-gh" || ws.GitHub.FetchedAt != nil || len(ws.Trabajos) != 12 {
		t.Fatalf("sin gh = %+v con %d trabajos; want sin-gh, fetchedAt null y los 12 locales", ws.GitHub, len(ws.Trabajos))
	}
	if sub, stale, _ := prNode(ws, "pramaestudio/platform/issue/54", 98); sub != "sin dato" || !stale {
		t.Fatalf("PR 98 = %q stale %v, want sin dato y stale", sub, stale)
	}
	var health struct {
		OK bool `json:"ok"`
	}
	if code := getJSON(t, srv.URL+"/api/health", &health); code != http.StatusOK || !health.OK {
		t.Fatalf("el proceso dejó de servir por GitHub: %d", code)
	}
}

func TestStream_EmiteSnapshotAlCambiar(t *testing.T) {
	root := fixtureWorkspace(t)
	_, srv := start(t, workspace.Config{Root: root, WatchInterval: 200 * time.Millisecond})
	waitScanned(t, srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/stream", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q", ct)
	}
	f, err := os.OpenFile(filepath.Join(root, "chichex", "skills", ".git", "sdd", "events.jsonl"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"v":1,"ts":"2026-10-27T13:00:00Z","harness":"pi","event":"grill.paused","issue":"#400","spec":"none","previous":"none","grill":"g-400","pr":[],"reason":"none"}` + "\n"); err != nil {
		t.Fatal(err)
	}
	_ = f.Close() // solo se escribió una línea
	sc := bufio.NewScanner(res.Body)
	event := ""
	for sc.Scan() {
		line := sc.Text()
		if line == "event: snapshot" {
			event = line
			continue
		}
		if data, ok := strings.CutPrefix(line, "data: "); ok && event != "" {
			var payload map[string]any
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				t.Fatalf("data no es JSON: %q", data)
			}
			if _, ok := payload["scannedAt"]; !ok {
				t.Fatalf("data sin scannedAt: %q", data)
			}
			if _, ok := payload["fetchedAt"]; !ok {
				t.Fatalf("data sin fetchedAt: %q", data)
			}
			return
		}
	}
	t.Fatalf("no llegó ningún event: snapshot tras el cambio: %v", sc.Err())
}
