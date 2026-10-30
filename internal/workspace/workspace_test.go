package workspace

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/github/fake"
	"github.com/chichex/sdd/internal/model"
)

const (
	fixtureGH = "../../testdata/gh"
	// roundsPerCase repite el poll sobre servicios nuevos: la carrera es una
	// ventana de microsegundos y una sola pasada puede no caer en ella.
	roundsPerCase = 10
)

func fixtureWorkspace(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatalf("git no está en PATH: el test necesita repos temporales: %v", err)
	}
	dir := filepath.Join(t.TempDir(), "ws")
	if out, err := exec.Command("bash", "../../testdata/mkworkspace.sh", dir).CombinedOutput(); err != nil {
		t.Fatalf("mkworkspace: %v\n%s", err, out)
	}
	return dir
}

// inconsistencia dice por qué snap mezcla estados de GitHub y trabajos de
// momentos distintos, o "" si la vista es coherente. caido es el repo que el
// poll reporta caído.
func inconsistencia(snap Snapshot, caido string) string {
	if snap.GitHub.FetchedAt != "" && snap.RepoGitHub[caido].OK {
		return fmt.Sprintf("github.fetchedAt = %q pero %s figura ok: el poll se ve a medias", snap.GitHub.FetchedAt, caido)
	}
	for _, tr := range snap.Trabajos {
		if snap.RepoGitHub[tr.Repo].OK {
			continue
		}
		for _, n := range prNodes(tr) {
			if !n.Stale {
				return fmt.Sprintf("%s figura caído pero su PR %d no es stale", tr.Repo, *n.Refs.PR)
			}
		}
	}
	return ""
}

func prNodes(tr model.Trabajo) []model.Nodo {
	var out []model.Nodo
	for _, r := range tr.Ramas {
		for _, n := range r.Nodos {
			if n.Etapa == model.PR && n.Refs.PR != nil {
				out = append(out, n)
			}
		}
	}
	return out
}

// aplicado dice si snap ya muestra el poll completo: hay fetchedAt y el repo
// caído tiene PRs, todos stale.
func aplicado(snap Snapshot, caido string) bool {
	if snap.GitHub.FetchedAt == "" {
		return false
	}
	found := false
	for _, tr := range snap.Trabajos {
		if tr.Repo != caido {
			continue
		}
		for _, n := range prNodes(tr) {
			if !n.Stale {
				return false
			}
			found = true
		}
	}
	return found
}

// leerHastaAplicar toma Snapshot() en loop, comprobando la coherencia en cada
// lectura, hasta que el poll se ve completo. Suelta gate cuando termina el
// escaneo inicial, para que el poll arranque con el lector ya leyendo.
func leerHastaAplicar(ctx context.Context, svc *Service, gate chan struct{}, caido string) (reads int, problema string) {
	released := false
	for ctx.Err() == nil {
		snap := svc.Snapshot()
		reads++
		if p := inconsistencia(snap, caido); p != "" {
			return reads, p
		}
		if !snap.Scanning && !released {
			close(gate)
			released = true
		}
		if aplicado(snap, caido) {
			return reads, ""
		}
	}
	return reads, "el poll no se aplicó a tiempo"
}

func TestSnapshot_GitHubYTrabajosCoincidenAlAplicarUnPoll(t *testing.T) {
	root := fixtureWorkspace(t)
	tests := []struct {
		name  string
		caido string // nombre del repo en testdata/gh
		id    string
	}{
		{"cae waica y platform responde", "waica", "chichex/waica"},
		{"cae platform y waica responde", "platform", "pramaestudio/platform"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for round := 1; round <= roundsPerCase; round++ {
				gate := make(chan struct{})
				gh := &fake.Runner{T: t, Dir: fixtureGH, Fail: map[string]string{tc.caido: "dial tcp: i/o timeout"}, Gate: gate}
				svc := New(Config{Root: root, WatchInterval: time.Hour, GitHub: gh, GitHubInterval: time.Hour})
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = svc.Run(ctx) // termina al cancelar ctx
				}()
				reads, problema := leerHastaAplicar(ctx, svc, gate, tc.id)
				cancel()
				wg.Wait()
				if problema != "" {
					t.Fatalf("ronda %d, tras %d lecturas: %s", round, reads, problema)
				}
			}
		})
	}
}
