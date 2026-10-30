package api

import (
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/github/fake"
	"github.com/chichex/sdd/internal/workspace"
)

const mensajeMostrar = "mostrar debe ser una lista de idea, en-curso, terminado o sin-traza"

// gitRepo crea en root/name un repo con contrato; remote vacío no agrega origin.
func gitRepo(t *testing.T, root, name, remote string) string {
	t.Helper()
	repo := filepath.Join(root, name)
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if remote != "" {
		if out, err := exec.Command("git", "-C", repo, "remote", "add", "origin", remote).CombinedOutput(); err != nil {
			t.Fatalf("git remote add: %v %s", err, out)
		}
	}
	for _, p := range []string{filepath.Join(repo, ".sdd", "specs"), filepath.Join(repo, ".git", "sdd")} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write(t, filepath.Join(repo, ".sdd", "project.md"), "# x\n")
	return repo
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// mostrarWorkspace arma un repo con un trabajo de cada categoría (el issue
// abierto sale del snapshot grabado de GitHub) y un repo vacío.
func mostrarWorkspace(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no está en PATH")
	}
	root := t.TempDir()
	repo := gitRepo(t, root, "categorias", "https://github.com/chichex/categorias.git")
	gitRepo(t, root, "vacio", "")
	// c es una spec sin ningún evento: sin-traza.
	write(t, filepath.Join(repo, ".sdd", "specs", "c.md"), "# Spec — c\n<!-- SDD-Tracking: version=1; type=spec; state=approved; issue=none; grill=none; superseded-by=none -->\n")
	log := `{"v":1,"ts":"2026-10-01T10:00:00Z","harness":"claude","event":"run.pr-opened","issue":"none","spec":".sdd%2Fspecs%2Fa.md","previous":"none","grill":"none","pr":[7],"reason":"none"}` + "\n" +
		`{"v":1,"ts":"2026-10-02T10:00:00Z","harness":"claude","event":"land.merged","issue":"none","spec":"none","previous":"none","grill":"none","pr":[7],"reason":"none"}` + "\n" +
		`{"v":1,"ts":"2026-10-03T10:00:00Z","harness":"claude","event":"spec.published","issue":"none","spec":".sdd%2Fspecs%2Fb.md","previous":"none","grill":"none","pr":[],"reason":"none"}` + "\n"
	write(t, filepath.Join(repo, ".git", "sdd", "events.jsonl"), log)
	return root
}

func TestWorkspace_Mostrar(t *testing.T) {
	root := mostrarWorkspace(t)
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh"}
	_, srv := start(t, ghConfig(root, gh, ""))
	waitGitHub(t, srv.URL, func(ws ghResp) bool { return ws.GitHub.FetchedAt != nil })
	const (
		a     = "chichex/categorias/spec/.sdd%2Fspecs%2Fa.md" // mergeada sin issue: terminado
		b     = "chichex/categorias/spec/.sdd%2Fspecs%2Fb.md" // publicada: en-curso
		c     = "chichex/categorias/spec/.sdd%2Fspecs%2Fc.md" // sin eventos: sin-traza
		issue = "chichex/categorias/issue/9"                  // abierto sin trabajo SDD: idea
	)
	todas := map[string]int{"idea": 1, "en-curso": 1, "terminado": 1, "sin-traza": 1}
	tests := []struct {
		name       string
		query      string
		want       []string
		categorias map[string]int
	}{
		{"sin parámetro: en curso e ideas", "", []string{b, issue}, todas},
		{"el default escrito es igual al default", "?mostrar=en-curso,idea", []string{b, issue}, todas},
		{"solo en-curso", "?mostrar=en-curso", []string{b}, todas},
		{"en-curso y terminado", "?mostrar=en-curso,terminado", []string{a, b}, todas},
		{"solo sin-traza", "?mostrar=sin-traza", []string{c}, todas},
		{"las cuatro categorías devuelven todo", queryTodas, []string{a, b, c, issue}, todas},
		{"un id repetido no duplica trabajos", "?mostrar=terminado,terminado", []string{a}, todas},
		{"?filtro=activos ya no existe: se ignora", "?filtro=activos", []string{b, issue}, todas},
		{"?filtro=todos tampoco devuelve todo", "?filtro=todos", []string{b, issue}, todas},
		{"?repo= aplica antes: solo cuenta ese repo", "?repo=chichex/categorias&mostrar=terminado", []string{a}, todas},
		{"un repo sin trabajos: cuatro ceros", "?repo=vacio" + "&mostrar=idea,en-curso,terminado,sin-traza", []string{}, map[string]int{"idea": 0, "en-curso": 0, "terminado": 0, "sin-traza": 0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ws workspaceResp
			if code := getJSON(t, srv.URL+"/api/workspace"+tc.query, &ws); code != http.StatusOK {
				t.Fatalf("GET %q = %d, want 200", tc.query, code)
			}
			got := make([]string, 0, len(ws.Trabajos))
			for _, tr := range ws.Trabajos {
				got = append(got, tr.ID)
			}
			slices.Sort(got)
			if want := slices.Sorted(slices.Values(tc.want)); !slices.Equal(got, want) {
				t.Fatalf("GET %q = %v, want %v", tc.query, got, want)
			}
			if !maps.Equal(ws.Categorias, tc.categorias) {
				t.Fatalf("GET %q categorias = %v, want %v", tc.query, ws.Categorias, tc.categorias)
			}
		})
	}
}

func TestWorkspace_MostrarDevuelveLaCategoriaDeCadaTrabajo(t *testing.T) {
	root := mostrarWorkspace(t)
	_, srv := start(t, ghConfig(root, &fake.Runner{T: t, Dir: "../../testdata/gh"}, ""))
	ws := waitGitHub(t, srv.URL, func(ws ghResp) bool { return ws.GitHub.FetchedAt != nil })
	var all workspaceResp
	getJSON(t, srv.URL+"/api/workspace"+queryTodas, &all)
	got := map[string]string{}
	for _, tr := range all.Trabajos {
		got[tr.ID] = tr.Categoria
	}
	want := map[string]string{
		"chichex/categorias/spec/.sdd%2Fspecs%2Fa.md": "terminado",
		"chichex/categorias/spec/.sdd%2Fspecs%2Fb.md": "en-curso",
		"chichex/categorias/spec/.sdd%2Fspecs%2Fc.md": "sin-traza",
		"chichex/categorias/issue/9":                  "idea",
	}
	if !maps.Equal(got, want) {
		t.Fatalf("categoria por trabajo = %v, want %v (github %+v)", got, want, ws.GitHub)
	}
}

func TestWorkspace_MostrarInvalido(t *testing.T) {
	root := mostrarWorkspace(t)
	_, srv := start(t, workspace.Config{Root: root, WatchInterval: time.Hour})
	waitScanned(t, srv.URL)
	for _, query := range []string{"?mostrar=", "?mostrar=otro", "?mostrar=en-curso,otro", "?mostrar=,", "?mostrar=en-curso,%20idea", "?mostrar=activos"} {
		t.Run(query, func(t *testing.T) {
			var bad struct {
				Error string `json:"error"`
			}
			if code := getJSON(t, srv.URL+"/api/workspace"+query, &bad); code != http.StatusBadRequest || bad.Error != mensajeMostrar {
				t.Fatalf("GET %q = %d %q, want 400 con %q", query, code, bad.Error, mensajeMostrar)
			}
		})
	}
}
