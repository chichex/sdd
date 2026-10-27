package model

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/github"
	"github.com/chichex/sdd/internal/github/fake"
)

func snapshot(t *testing.T, repo string) *github.Snapshot {
	t.Helper()
	gh := &fake.Runner{T: t, Dir: "../../testdata/gh"}
	s, err := github.Fetch(context.Background(), gh, repo, time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Fetch(%s): %v", repo, err)
	}
	return s
}

func withGitHub(t *testing.T, stale bool) []RepoInput {
	t.Helper()
	inputs := realLogs(t)
	for i := range inputs {
		inputs[i].GitHub = snapshot(t, inputs[i].Data.Repo.ID)
		inputs[i].Stale = stale
	}
	return inputs
}

func TestBuild_IdeaDesdeIssuesAbiertos(t *testing.T) {
	trabajos := Build(withGitHub(t, false))
	if len(trabajos) != 14 {
		t.Fatalf("trabajos = %d, want 14 (12 + las ideas waica#210 y platform#120)", len(trabajos))
	}
	idea := find(t, trabajos, "chichex/waica/issue/210")
	if idea.Etapa != Idea || idea.Titulo != "Idea sintética sin artefactos" || idea.UltimaActividad != "2026-10-26T09:00:00Z" {
		t.Fatalf("idea = %+v", idea)
	}
	if n := idea.Ramas[0].Nodos; len(n) != 1 || n[0].URL == nil || *n[0].URL != "https://github.com/chichex/waica/issues/210" {
		t.Fatalf("nodo idea = %+v", n)
	}
	p57 := find(t, trabajos, "pramaestudio/platform/issue/57")
	if p57.Titulo != "Issue 57 sintético" || len(nodos(p57, "principal", Idea)) != 1 {
		t.Fatalf("#57 = título %q ideas %d, want el título del issue sin duplicar el trabajo", p57.Titulo, len(nodos(p57, "principal", Idea)))
	}
	if got := anclas(trabajos, "pramaestudio/dale-que-sale"); len(got) != 3 {
		t.Fatalf("un issue cerrado sin artefactos no debe aparecer")
	}
}

func TestBuild_SubEstadosDePR(t *testing.T) {
	trabajos := Build(withGitHub(t, false))
	want := map[int]string{98: "validado", 106: "draft", 108: "cambios pedidos", 109: "validado", 111: "checks rojos", 112: "abierto", 113: "abierto", 114: "abierto"}
	for _, tr := range trabajos {
		for _, n := range nodos(tr, "principal", PR) {
			if w, ok := want[*n.Refs.PR]; ok && (n.SubEstado != w || n.Stale) {
				t.Fatalf("PR %d = %q stale %v, want %q fresco", *n.Refs.PR, n.SubEstado, n.Stale, w)
			}
		}
	}
	p56 := find(t, trabajos, "pramaestudio/platform/issue/56")
	if p56.Etapa != PR || p56.SubEstado != "checks rojos" {
		t.Fatalf("#56 = %q %q, want pr con el sub-estado de la capa 1 (111 checks rojos)", p56.Etapa, p56.SubEstado)
	}
}

func TestBuild_MergeadoDesdeGitHub(t *testing.T) {
	trabajos := Build(withGitHub(t, false))
	p100 := find(t, trabajos, "pramaestudio/platform/issue/100")
	merged := nodos(p100, "principal", Mergeado)
	if len(merged) != 1 || *merged[0].Refs.PR != 105 || merged[0].TS != "2026-10-26T18:10:00Z" || p100.SubEstado != "1/2" {
		t.Fatalf("#100 mergeados = %+v sub %q, want el 105 con ts = mergedAt y 1/2", merged, p100.SubEstado)
	}
	p54 := nodos(find(t, trabajos, "pramaestudio/platform/issue/54"), "principal", Mergeado)
	if len(p54) != 3 || p54[0].TS != "2026-10-25T18:53:46Z" || p54[0].Refs.MergedAt == nil || *p54[0].Refs.MergedAt != "2026-10-25T18:53:10Z" {
		t.Fatalf("#54 mergeados = %+v, want ts del evento y mergedAt en refs", p54)
	}
}

func TestBuild_VinculacionPorBody(t *testing.T) {
	trabajos := Build(withGitHub(t, false))
	var prs []int
	for _, n := range nodos(find(t, trabajos, "pramaestudio/platform/issue/57"), "principal", PR) {
		prs = append(prs, *n.Refs.PR)
	}
	if !slices.Equal(prs, []int{108, 109, 130}) {
		t.Fatalf("#57 PRs = %v, want 108, 109 y el 130 por `Closes #57`", prs)
	}
	for _, id := range []string{"pramaestudio/platform/issue/120", "chichex/waica/issue/210"} {
		if got := nodos(find(t, trabajos, id), "*", PR); len(got) != 0 {
			t.Fatalf("%s tiene PRs %+v; un body que apunta a un trabajo sin spec queda fuera", id, got)
		}
	}
}

func TestBuild_Taggeado(t *testing.T) {
	inputs := withGitHub(t, false)
	sha := func(n int) string { p, _ := inputs[2].GitHub.PR(n); return p.MergeCommit }
	inputs[2].Tags = Tags{BySHA: map[string][]string{sha(95): {"api/v1.0.57", "web/v2.1.0"}, sha(96): {"api/v1.0.57"}, sha(97): {}}}
	skillsPR, _ := inputs[3].GitHub.PR(101)
	inputs[3].Tags = Tags{Missing: map[string]bool{skillsPR.MergeCommit: true}}
	trabajos := Build(inputs)
	p54 := find(t, trabajos, "pramaestudio/platform/issue/54")
	pub := nodos(p54, "principal", Publicado)
	if len(pub) != 1 || pub[0].SubEstado != "taggeado" || !slices.Equal(pub[0].Tags, []string{"api/v1.0.57", "web/v2.1.0"}) || p54.Etapa != Publicado || !p54.Terminado {
		t.Fatalf("#54 publicado = %+v etapa %q terminado %v", pub, p54.Etapa, p54.Terminado)
	}
	s100 := find(t, trabajos, "chichex/skills/issue/100")
	if pub := nodos(s100, "principal", Publicado); len(pub) != 1 || pub[0].SubEstado != "sin dato local" || s100.Etapa != Mergeado {
		t.Fatalf("skills#100 publicado = %+v etapa %q, want sin dato local y etapa mergeado", pub, s100.Etapa)
	}
	for _, tr := range trabajos {
		for _, n := range nodos(tr, "*", Publicado) {
			if n.SubEstado == "en prod" {
				t.Fatalf("en prod no se produce en v1")
			}
		}
	}
}

func TestBuild_DegradadoMarcaStale(t *testing.T) {
	trabajos := Build(withGitHub(t, true))
	for _, tr := range trabajos {
		for _, n := range nodos(tr, "*", PR) {
			if n.SubEstado != "sin dato" || !n.Stale {
				t.Fatalf("%s PR %d = %q stale %v, want sin dato y stale", tr.ID, *n.Refs.PR, n.SubEstado, n.Stale)
			}
		}
		for _, n := range nodos(tr, "*", Idea) {
			if !n.Stale {
				t.Fatalf("%s idea sin stale", tr.ID)
			}
		}
	}
	if s := find(t, trabajos, "chichex/skills/issue/100"); !s.Terminado {
		t.Fatalf("skills#100 mergeado con issue cerrado debería estar terminado")
	}
}
