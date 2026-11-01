package scan

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func gitInit(t *testing.T, dir string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatal("git no está en PATH: este test necesita repos temporales")
	}
	if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
}

func TestDiscover_ChicheYSDD(t *testing.T) {
	ws := t.TempDir()
	for _, tc := range []struct{ repo, brand string }{{"nuevo", ".chiche"}, {"viejo", ".sdd"}, {"ambos", ".chiche"}, {"ambos", ".sdd"}, {"nada", ".otro"}} {
		writeFile(t, filepath.Join(ws, tc.repo, tc.brand, "project.md"), "# Contrato\n")
	}
	for _, r := range []string{"nuevo", "viejo", "ambos", "nada"} {
		gitInit(t, filepath.Join(ws, r))
	}
	repos, err := Discover(context.Background(), ws)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	var ids []string
	for _, r := range repos {
		ids = append(ids, r.ID)
	}
	if want := []string{"ambos", "nuevo", "viejo"}; !slices.Equal(ids, want) {
		t.Fatalf("repos = %v, want %v (uno por carpeta, con .chiche/ o .sdd/)", ids, want)
	}
}

func TestRead_CuatroLugaresConPrecedencia(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, ".git")
	repo := Repo{ID: "o/r", Rutas: []string{root}, CommonDir: common}
	places := map[string]string{
		"arbol-chiche":  filepath.Join(root, ".chiche", "specs"),
		"arbol-sdd":     filepath.Join(root, ".sdd", "specs"),
		"common-chiche": filepath.Join(common, "chiche", "specs"),
		"common-sdd":    filepath.Join(common, "sdd", "specs"),
	}
	order := []string{"arbol-chiche", "arbol-sdd", "common-chiche", "common-sdd"}
	// x.md está en los cuatro; y.md en los tres últimos; z.md en los dos del
	// common-dir; w.md solo en <common>/sdd.
	for i, place := range order {
		for _, name := range []string{"x", "y", "z", "w"}[:i+1] {
			writeFile(t, filepath.Join(places[place], name+".md"), "# Spec — "+place+"\n")
		}
	}
	specs := map[string]Artifact{}
	for _, a := range Read(repo).Specs {
		specs[a.Name] = a
	}
	want := map[string]struct{ title, logical, path string }{
		"x.md": {"Spec — arbol-chiche", ".chiche/specs/x.md", filepath.Join(places["arbol-chiche"], "x.md")},
		"y.md": {"Spec — arbol-sdd", ".sdd/specs/y.md", filepath.Join(places["arbol-sdd"], "y.md")},
		"z.md": {"Spec — common-chiche", ".chiche/specs/z.md", filepath.Join(places["common-chiche"], "z.md")},
		"w.md": {"Spec — common-sdd", ".sdd/specs/w.md", filepath.Join(places["common-sdd"], "w.md")},
	}
	if len(specs) != len(want) {
		t.Fatalf("specs = %d, want %d (una por nombre)", len(specs), len(want))
	}
	for name, w := range want {
		got := specs[name]
		if got.Title != w.title || got.Logical != w.logical || got.Path != w.path {
			t.Fatalf("%s = %q %q %q, want %q %q %q", name, got.Title, got.Logical, got.Path, w.title, w.logical, w.path)
		}
	}
}

func TestRead_GrillsDeChiche(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, ".git")
	writeFile(t, filepath.Join(common, "chiche", "grills", "g.md"), "# Grill — Nuevo\n<!-- Chiche-Tracking: version=2; type=grill; state=paused; issue=none; grill=g; project=p -->\n")
	data := Read(Repo{ID: "o/r", Rutas: []string{root}, CommonDir: common})
	if len(data.Grills) != 1 || data.Grills[0].Logical != ".chiche/grills/g.md" || data.Grills[0].Tracking.Metadata.Version != 2 {
		t.Fatalf("grills = %+v, want g.md de <common>/chiche con tracking v2", data.Grills)
	}
}

func TestRead_DosLogsOrdenadosPorTS(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, ".git")
	ev := func(ts, name string) string {
		return `{"v":1,"ts":"` + ts + `","harness":"claude","event":"` + name + `","issue":"#9","spec":"%239","previous":"none","grill":"none","pr":[],"reason":"none"}` + "\n"
	}
	writeFile(t, filepath.Join(common, "sdd", "events.jsonl"),
		ev("2026-10-01T10:00:00Z", "spec.published")+ev("2026-10-03T10:00:00Z", "run.started")+"basura\n")
	writeFile(t, filepath.Join(common, "chiche", "events.jsonl"),
		ev("2026-10-02T10:00:00Z", "spec.approved")+ev("2026-10-03T10:00:00Z", "run.committed")+"{\"v\":2}\n")
	data := Read(Repo{ID: "o/r", Rutas: []string{root}, CommonDir: common})
	var got []string
	for _, e := range data.Events {
		got = append(got, e.Event)
	}
	want := []string{"spec.published", "spec.approved", "run.started", "run.committed"}
	if !slices.Equal(got, want) || data.LineasInvalidas != 2 || len(data.Errores) != 0 {
		t.Fatalf("eventos = %v, inválidas = %d, errores = %v; want %v, 2 y ninguno (empate: sdd antes que chiche)", got, data.LineasInvalidas, data.Errores, want)
	}
}

func TestRead_UnSoloLogConservaSuOrden(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, ".git")
	ev := func(ts, name string) string {
		return `{"v":1,"ts":"` + ts + `","harness":"claude","event":"` + name + `","issue":"#9","spec":"%239","previous":"none","grill":"none","pr":[],"reason":"none"}` + "\n"
	}
	writeFile(t, filepath.Join(common, "sdd", "events.jsonl"), ev("2026-10-03T10:00:00Z", "run.started")+ev("2026-10-01T10:00:00Z", "spec.published"))
	var got []string
	for _, e := range Read(Repo{ID: "o/r", Rutas: []string{root}, CommonDir: common}).Events {
		got = append(got, e.Event)
	}
	if want := []string{"run.started", "spec.published"}; !slices.Equal(got, want) {
		t.Fatalf("eventos = %v, want %v (orden del log, sin reordenar por ts)", got, want)
	}
}

func TestRead_SoloLogDeChiche(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, ".git")
	writeFile(t, filepath.Join(common, "chiche", "events.jsonl"), `{"v":1,"ts":"2026-10-02T10:00:00Z","harness":"claude","event":"spec.published","issue":"#9","spec":"%239","previous":"none","grill":"none","pr":[],"reason":"none"}`+"\n")
	data := Read(Repo{ID: "o/r", Rutas: []string{root}, CommonDir: common})
	if len(data.Events) != 1 || len(data.Errores) != 0 {
		t.Fatalf("eventos = %d, errores = %v; want 1 y ninguno", len(data.Events), data.Errores)
	}
}

func TestStamp_CubreLasRutasDeChiche(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, ".git")
	repo := Repo{ID: "o/r", Rutas: []string{root}, CommonDir: common}
	if err := os.MkdirAll(common, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(common, "chiche", "events.jsonl"),
		filepath.Join(root, ".chiche", "specs", "a.md"),
		filepath.Join(common, "chiche", "grills", "g.md"),
	} {
		before := Stamp(repo)
		writeFile(t, path, "x\n")
		if Stamp(repo) == before {
			t.Fatalf("Stamp no cambió al escribir %s", path)
		}
	}
}
