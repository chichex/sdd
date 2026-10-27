package scan

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fixtureWorkspace arma el workspace de testdata/mkworkspace.sh en un
// directorio temporal. Sin git en el entorno, los tests de escaneo se saltean.
func fixtureWorkspace(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no está en PATH: los tests de escaneo necesitan repos temporales")
	}
	dir := filepath.Join(t.TempDir(), "ws")
	out, err := exec.Command("bash", "../../testdata/mkworkspace.sh", dir).CombinedOutput()
	if err != nil {
		t.Fatalf("mkworkspace: %v\n%s", err, out)
	}
	return dir
}

func byID(repos []Repo) map[string]Repo {
	m := map[string]Repo{}
	for _, r := range repos {
		m[r.ID] = r
	}
	return m
}

func TestDiscover_Workspace(t *testing.T) {
	ws := fixtureWorkspace(t)
	repos, err := Discover(context.Background(), ws)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	got := byID(repos)
	var ids []string
	for id := range got {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	want := []string{"chichex/skills", "chichex/waica", "pramaestudio/dale-que-sale", "pramaestudio/platform", "sin-git", "sin-remote"}
	if !slices.Equal(ids, want) {
		t.Fatalf("repos = %v, want %v (sin profundo ni node_modules)", ids, want)
	}
	waica := got["chichex/waica"]
	if len(waica.Rutas) != 2 || waica.Rutas[0] != filepath.Join(ws, "waica") || waica.Rutas[1] != filepath.Join(ws, "waica-wt") {
		t.Fatalf("waica rutas = %v, want el checkout principal y el worktree", waica.Rutas)
	}
	if waica.Remote != "git@github.com:chichex/waica.git" || waica.Nombre != "waica" {
		t.Fatalf("waica = %+v", waica)
	}
	if !got["sin-git"].SinGit || got["sin-remote"].SinGit || got["sin-remote"].Remote != "" {
		t.Fatalf("sin-git = %+v, sin-remote = %+v", got["sin-git"], got["sin-remote"])
	}
	if got["pramaestudio/platform"].CommonDir != filepath.Join(ws, "platform", ".git") {
		t.Fatalf("platform common-dir = %q", got["pramaestudio/platform"].CommonDir)
	}
}

func TestRepoID_FormasDeRemote(t *testing.T) {
	tests := []struct{ remote, want string }{
		{"https://github.com/o/r.git", "o/r"},
		{"https://github.com/o/r", "o/r"},
		{"git@github.com:o/r.git", "o/r"},
		{"ssh://git@github.com/o/r.git", "o/r"},
		{"ssh://git@github.com:22/o/r", "o/r"},
		{"", ""},
	}
	for _, tc := range tests {
		t.Run(tc.remote, func(t *testing.T) {
			if got := RepoID(tc.remote); got != tc.want {
				t.Fatalf("RepoID(%q) = %q, want %q", tc.remote, got, tc.want)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRead_FuentesLocales(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, ".git")
	repo := Repo{ID: "o/r", Rutas: []string{root}, CommonDir: common}
	valid := `{"v":1,"ts":"2026-10-14T14:00:00Z","harness":"claude","event":"run.started","issue":"#93","spec":"%2393","previous":"none","grill":"none","pr":[],"reason":"none"}`
	writeFile(t, filepath.Join(common, "sdd", "events.jsonl"), valid+"\nbasura\n"+valid+"\n{\"v\":2}\n"+valid)
	writeFile(t, filepath.Join(root, ".sdd", "grills", "g.md"), "# Grill — Árbol\n<!-- SDD-Tracking: version=1; type=grill; state=paused; issue=none; grill=g; project=p -->\n")
	writeFile(t, filepath.Join(common, "sdd", "grills", "h.md"), "# Grill — Common\n<!-- SDD-Tracking: version=1; type=grill; state=finalized; issue=none; grill=h; project=p -->\n")
	writeFile(t, filepath.Join(root, ".sdd", "specs", "x.md"), "# Spec — Del árbol\n<!-- SDD-Tracking: version=1; type=spec; state=approved; issue=none; grill=g; superseded-by=none -->\n")
	writeFile(t, filepath.Join(common, "sdd", "specs", "x.md"), "# Spec — Del common-dir\n<!-- SDD-Tracking: version=1; type=spec; state=draft; issue=none; grill=g; superseded-by=none -->\n")
	writeFile(t, filepath.Join(common, "sdd", "specs", "y.md"), "# Spec — Solo common\n")
	if err := os.MkdirAll(filepath.Join(root, ".sdd", "specs", "rota.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	data := Read(repo)
	if len(data.Events) != 2 || data.LineasInvalidas != 2 {
		t.Fatalf("eventos = %d, inválidas = %d; want 2 y 2 (la final sin LF no cuenta)", len(data.Events), data.LineasInvalidas)
	}
	if len(data.Grills) != 2 {
		t.Fatalf("grills = %d, want 2 (árbol + common-dir)", len(data.Grills))
	}
	specs := map[string]Artifact{}
	for _, a := range data.Specs {
		specs[a.Name] = a
	}
	if len(specs) != 2 || specs["x.md"].Title != "Spec — Del árbol" || specs["x.md"].Tracking.State.Value != "approved" {
		t.Fatalf("specs = %+v, want x.md del árbol y y.md", specs)
	}
	if specs["x.md"].Logical != ".sdd/specs/x.md" || specs["y.md"].Path != filepath.Join(common, "sdd", "specs", "y.md") {
		t.Fatalf("rutas = %q, %q", specs["x.md"].Logical, specs["y.md"].Path)
	}
	if len(data.Errores) != 1 || !strings.Contains(data.Errores[0], "rota.md") {
		t.Fatalf("errores = %v, want uno por rota.md", data.Errores)
	}
}

func TestRead_DirectoriosAusentesSonVacios(t *testing.T) {
	root := t.TempDir()
	data := Read(Repo{ID: "o/r", Rutas: []string{root}, CommonDir: filepath.Join(root, ".git")})
	if len(data.Events) != 0 || len(data.Grills) != 0 || len(data.Specs) != 0 || len(data.Errores) != 0 {
		t.Fatalf("data = %+v, want vacía y sin errores", data)
	}
}
