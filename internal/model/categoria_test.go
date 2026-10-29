package model

import (
	"testing"
	"time"

	"github.com/chichex/sdd/internal/events"
	"github.com/chichex/sdd/internal/github"
	"github.com/chichex/sdd/internal/scan"
)

func TestBuild_Categoria(t *testing.T) {
	mtime := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	ev := func(ts, name, spec string, pr ...int) events.Event {
		return events.Event{V: 1, TS: ts, Harness: "claude", Event: name, Spec: spec, PR: pr}
	}
	sinEventos := artifact("spec", "sin-eventos", "<!-- SDD-Tracking: version=1; type=spec; state=approved; issue=none; grill=none; superseded-by=none -->", mtime)
	local := RepoInput{Data: scan.Data{Repo: scan.Repo{ID: "o/r"}, Specs: []scan.Artifact{sinEventos}, Events: []events.Event{
		ev("2026-10-01T10:00:00Z", "run.pr-opened", ".sdd/specs/mergeada.md", 7),
		ev("2026-10-02T10:00:00Z", "land.merged", "", 7),
		ev("2026-10-03T10:00:00Z", "spec.published", ".sdd/specs/publicada.md"),
	}}}
	// Un issue abierto que ningún artefacto ni evento ancla: grupo sin ítems.
	conIssue := RepoInput{
		Data:   scan.Data{Repo: scan.Repo{ID: "o/i"}},
		GitHub: &github.Snapshot{Repo: "o/i", IssuesComplete: true, Issues: []github.Issue{{Number: 9, Title: "Idea suelta", URL: "https://github.com/o/i/issues/9", UpdatedAt: "2026-10-04T10:00:00Z"}}},
	}
	trabajos := Build([]RepoInput{local, conIssue})
	tests := []struct {
		name string
		id   string
		want Categoria
	}{
		{"un issue abierto sin trabajo SDD es idea", "o/i/issue/9", CategoriaIdea},
		{"una spec sin ningún evento es sin-traza", "o/r/spec/.sdd%2Fspecs%2Fsin-eventos.md", CategoriaSinTraza},
		{"una spec mergeada sin issue es terminado", "o/r/spec/.sdd%2Fspecs%2Fmergeada.md", CategoriaTerminado},
		{"una spec publicada y sin terminar es en-curso", "o/r/spec/.sdd%2Fspecs%2Fpublicada.md", CategoriaEnCurso},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := find(t, trabajos, tc.id).Categoria; got != tc.want {
				t.Fatalf("categoria de %s = %q, want %q", tc.id, got, tc.want)
			}
		})
	}
	if len(trabajos) != len(tests) {
		t.Fatalf("trabajos = %d, want %d (uno por categoría)", len(trabajos), len(tests))
	}
}

func TestBuild_CategoriaIdeaNoCambiaSinHistoria(t *testing.T) {
	idea := find(t, Build(withGitHub(t, false)), "chichex/waica/issue/210")
	if idea.Categoria != CategoriaIdea || idea.SinHistoria || idea.Terminado {
		t.Fatalf("idea = categoria %q sinHistoria %v terminado %v, want idea con sinHistoria y terminado en false", idea.Categoria, idea.SinHistoria, idea.Terminado)
	}
}

// Cada trabajo cae en exactamente una categoría y esta coincide con sus banderas.
func TestBuild_CategoriaUnicaEnLogsReales(t *testing.T) {
	validas := map[Categoria]bool{CategoriaIdea: true, CategoriaEnCurso: true, CategoriaTerminado: true, CategoriaSinTraza: true}
	for name, inputs := range map[string][]RepoInput{"solo logs": realLogs(t), "con GitHub": withGitHub(t, false)} {
		for _, tr := range Build(inputs) {
			t.Run(name+"/"+tr.ID, func(t *testing.T) {
				if !validas[tr.Categoria] {
					t.Fatalf("categoria = %q, want una de idea, en-curso, terminado o sin-traza", tr.Categoria)
				}
				if (tr.Categoria == CategoriaSinTraza) != tr.SinHistoria {
					t.Errorf("categoria %q con sinHistoria %v: sin-traza es exactamente sinHistoria", tr.Categoria, tr.SinHistoria)
				}
				if (tr.Categoria == CategoriaTerminado) != (tr.Terminado && !tr.SinHistoria) {
					t.Errorf("categoria %q con terminado %v sinHistoria %v: terminado es lo terminado con traza", tr.Categoria, tr.Terminado, tr.SinHistoria)
				}
				if tr.Categoria == CategoriaIdea && (tr.Etapa != Idea || tr.SinHistoria || tr.Terminado) {
					t.Errorf("idea en etapa %q sinHistoria %v terminado %v, want etapa idea sin banderas", tr.Etapa, tr.SinHistoria, tr.Terminado)
				}
			})
		}
	}
}
