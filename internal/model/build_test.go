package model

import (
	"os"
	"slices"
	"testing"
	"time"

	"github.com/chichex/sdd/internal/events"
	"github.com/chichex/sdd/internal/scan"
	"github.com/chichex/sdd/internal/tracking"
)

func logData(t *testing.T, file, id string) scan.Data {
	t.Helper()
	f, err := os.Open("../../testdata/events/" + file)
	if err != nil {
		t.Fatalf("abrir %s: %v", file, err)
	}
	defer f.Close()
	log, err := events.ReadLog(f)
	if err != nil || log.Invalid != 0 {
		t.Fatalf("ReadLog(%s): %v, inválidas %d", file, err, log.Invalid)
	}
	return scan.Data{Repo: scan.Repo{ID: id, Remote: "https://github.com/" + id + ".git"}, Events: log.Events}
}

func realLogs(t *testing.T) []RepoInput {
	t.Helper()
	return []RepoInput{
		{Data: logData(t, "dale-que-sale.jsonl", "pramaestudio/dale-que-sale")},
		{Data: logData(t, "waica.jsonl", "chichex/waica")},
		{Data: logData(t, "platform.jsonl", "pramaestudio/platform")},
		{Data: logData(t, "skills.jsonl", "chichex/skills")},
	}
}

func anclas(trabajos []Trabajo, repo string) []string {
	var out []string
	for _, tr := range trabajos {
		if tr.Repo == repo {
			out = append(out, tr.Ancla.Ref)
		}
	}
	slices.Sort(out)
	return out
}

func find(t *testing.T, trabajos []Trabajo, id string) Trabajo {
	t.Helper()
	for _, tr := range trabajos {
		if tr.ID == id {
			return tr
		}
	}
	t.Fatalf("no hay trabajo %q", id)
	return Trabajo{}
}

func TestBuild_TrabajosDeLogsReales(t *testing.T) {
	trabajos := Build(realLogs(t))
	tests := []struct {
		repo string
		want []string
	}{
		{"pramaestudio/dale-que-sale", []string{"#287", "#288", "#289"}},
		{"chichex/waica", []string{"#203", "#205"}},
		{"pramaestudio/platform", []string{"#100", "#101", "#102", "#54", "#56", "#57"}},
		{"chichex/skills", []string{"#100"}},
	}
	for _, tc := range tests {
		t.Run(tc.repo, func(t *testing.T) {
			if got := anclas(trabajos, tc.repo); !slices.Equal(got, tc.want) {
				t.Fatalf("anclas = %v, want %v", got, tc.want)
			}
		})
	}
	if len(trabajos) != 12 {
		t.Fatalf("trabajos = %d, want 12 (el PR 249 sin ancla queda fuera)", len(trabajos))
	}
	for _, id := range []string{"chichex/waica/issue/203", "chichex/waica/issue/205"} {
		if prs := nodos(find(t, trabajos, id), "principal", PR); len(prs) != 5 {
			t.Fatalf("%s: nodos PR = %d, want 5", id, len(prs))
		}
	}
	for i := 1; i < len(trabajos); i++ {
		if trabajos[i-1].UltimaActividad < trabajos[i].UltimaActividad {
			t.Fatalf("trabajos no ordenados por ultimaActividad descendente")
		}
	}
}

func nodos(tr Trabajo, rama string, etapa Etapa) []Nodo {
	var out []Nodo
	for _, r := range tr.Ramas {
		if r.ID != rama && rama != "*" {
			continue
		}
		for _, n := range r.Nodos {
			if n.Etapa == etapa {
				out = append(out, n)
			}
		}
	}
	return out
}

func TestBuild_EtapasYRamas(t *testing.T) {
	trabajos := Build(realLogs(t))
	w203 := find(t, trabajos, "chichex/waica/issue/203")
	runs := nodos(w203, "*", Run)
	if len(w203.Ramas) != 2 || len(runs) != 2 || runs[1].SubEstado != "fallado (interrupted)" || runs[0].SubEstado != "con PR" {
		t.Fatalf("#203 ramas = %+v, want principal con PR y una rama fallado (interrupted)", w203.Ramas)
	}
	if w203.Ramas[1].Estado != "fallada" || w203.Ramas[0].Estado != "viva" {
		t.Fatalf("#203 estados de rama = %q, %q", w203.Ramas[0].Estado, w203.Ramas[1].Estado)
	}
	if merged := nodos(w203, "principal", Mergeado); len(merged) != 5 || w203.Etapa != Mergeado || w203.SubEstado != "5/5" {
		t.Fatalf("#203 mergeados = %d, etapa %q %q; want 5, mergeado 5/5", len(merged), w203.Etapa, w203.SubEstado)
	}
	p54 := find(t, trabajos, "pramaestudio/platform/issue/54")
	if p54.Etapa != Mergeado || p54.SubEstado != "3/4" {
		t.Fatalf("#54 = %q %q, want mergeado 3/4", p54.Etapa, p54.SubEstado)
	}
	prs := nodos(p54, "principal", PR)
	if len(prs) != 4 || *prs[3].Refs.PR != 98 || prs[3].SubEstado != "sin dato" || prs[3].Capa != "4/4" {
		t.Fatalf("#54 PRs = %+v, want 4 con el 98 como 4/4 sin dato", prs)
	}
	if grill := nodos(p54, "principal", Grill); len(grill) != 1 || grill[0].SubEstado != "finalizado" {
		t.Fatalf("#54 grill = %+v, want finalizado", grill)
	}
	if spec := nodos(p54, "principal", Spec); len(spec) != 1 || spec[0].SubEstado != "aprobada" {
		t.Fatalf("#54 spec = %+v, want aprobada", spec)
	}
	p57 := find(t, trabajos, "pramaestudio/platform/issue/57")
	if p57.Etapa != PR || p57.SubEstado != "sin dato" {
		t.Fatalf("#57 = %q %q, want pr sin dato", p57.Etapa, p57.SubEstado)
	}
	if p57.UltimaActividad != "2026-10-26T17:30:47Z" {
		t.Fatalf("#57 ultimaActividad = %q", p57.UltimaActividad)
	}
	d289 := find(t, trabajos, "pramaestudio/dale-que-sale/issue/289")
	if d289.Etapa != Mergeado || d289.SubEstado != "1/1" || d289.Terminado {
		t.Fatalf("#289 = %q %q terminado %v; want mergeado 1/1 no terminado (issue sin dato de GitHub)", d289.Etapa, d289.SubEstado, d289.Terminado)
	}
}

func TestBuild_RunNuevoTrasFallo(t *testing.T) {
	ev := func(ts, name, reason string, pr ...int) events.Event {
		return events.Event{V: 1, TS: ts, Harness: "claude", Event: name, Spec: ".sdd/specs/x.md", Reason: reason, PR: pr}
	}
	data := scan.Data{Repo: scan.Repo{ID: "o/r"}, Events: []events.Event{
		ev("2026-10-01T10:00:00Z", "spec.published", ""),
		ev("2026-10-01T11:00:00Z", "run.started", ""),
		ev("2026-10-01T11:30:00Z", "run.failed", "aborted"),
	}}
	tr := Build([]RepoInput{{Data: data}})
	if len(tr) != 1 || tr[0].Etapa != Spec || len(tr[0].Ramas) != 2 || tr[0].Ramas[1].Nodos[0].SubEstado != "fallado (aborted)" {
		t.Fatalf("trabajo = %+v, want etapa spec y una rama fallado (aborted)", tr)
	}
	if tr[0].Ancla != (Ancla{Tipo: "spec", Ref: ".sdd/specs/x.md"}) || tr[0].ID != "o/r/spec/.sdd%2Fspecs%2Fx.md" {
		t.Fatalf("ancla = %+v id %q", tr[0].Ancla, tr[0].ID)
	}
	data.Events = append(data.Events, ev("2026-10-02T09:00:00Z", "run.started", ""))
	tr = Build([]RepoInput{{Data: data}})
	if tr[0].Etapa != Run || tr[0].SubEstado != "corriendo" || len(tr[0].Ramas) != 2 {
		t.Fatalf("tras un run.started nuevo = %q %q ramas %d; want run corriendo con la fallada aparte", tr[0].Etapa, tr[0].SubEstado, len(tr[0].Ramas))
	}
}

func artifact(kind, name, marker string, mtime time.Time) scan.Artifact {
	text := "# " + kind + " — " + name + "\n" + marker + "\n"
	return scan.Artifact{Kind: kind, Name: name + ".md", Path: "/r/.sdd/" + kind + "s/" + name + ".md",
		Logical: ".sdd/" + kind + "s/" + name + ".md", Title: "Spec — " + name, ModTime: mtime, Tracking: tracking.Parse(text)}
}

func TestBuild_SinHistoria(t *testing.T) {
	mtime := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	data := scan.Data{Repo: scan.Repo{ID: "o/r"}, Specs: []scan.Artifact{
		artifact("spec", "vieja", "<!-- SDD-Tracking: version=1; type=spec; state=implemented; issue=none; grill=none; superseded-by=none -->", mtime),
	}}
	tr := Build([]RepoInput{{Data: data}})
	if len(tr) != 1 {
		t.Fatalf("trabajos = %d, want 1", len(tr))
	}
	got := tr[0]
	if !got.SinHistoria || got.Etapa != Spec || got.SubEstado != "implementada" || len(got.Ramas) != 1 || len(got.Ramas[0].Nodos) != 1 {
		t.Fatalf("trabajo = %+v, want un nodo spec implementada sin historia", got)
	}
	if got.Ramas[0].Nodos[0].TS != "2026-09-01T12:00:00Z" || len(got.Gantt) != 0 {
		t.Fatalf("ts = %q, gantt = %v; want el mtime y sin barras", got.Ramas[0].Nodos[0].TS, got.Gantt)
	}
}

func TestBuild_Gantt(t *testing.T) {
	w203 := find(t, Build(realLogs(t)), "chichex/waica/issue/203")
	want := []Barra{
		{Rama: "principal", Etapa: Run, Inicio: "2026-10-17T10:11:07Z", Fin: ptr("2026-10-17T10:31:38Z"), Estado: "con PR"},
		{Rama: "principal", Etapa: PR, Inicio: "2026-10-17T10:31:38Z", Fin: ptr("2026-10-23T14:45:00Z"), Estado: "sin dato"},
		{Rama: "principal", Etapa: Mergeado, Inicio: "2026-10-23T14:45:00Z", Fin: nil, Estado: "5/5"},
		{Rama: "run-1", Etapa: Run, Inicio: "2026-10-17T04:52:40Z", Fin: ptr("2026-10-17T06:20:45Z"), Estado: "fallado (interrupted)"},
	}
	if len(w203.Gantt) != len(want) {
		t.Fatalf("gantt = %+v, want %d barras", w203.Gantt, len(want))
	}
	for i, b := range want {
		got := w203.Gantt[i]
		if got.Rama != b.Rama || got.Etapa != b.Etapa || got.Inicio != b.Inicio || got.Estado != b.Estado || !sameEnd(got.Fin, b.Fin) {
			t.Fatalf("barra %d = %+v (fin %v), want %+v (fin %v)", i, got, deref(got.Fin), b, deref(b.Fin))
		}
	}
	if w203.Links.Issue == nil || *w203.Links.Issue != "https://github.com/chichex/waica/issues/203" || len(w203.Links.PRs) != 5 {
		t.Fatalf("links = %+v", w203.Links)
	}
}

func sameEnd(a, b *string) bool { return (a == nil && b == nil) || (a != nil && b != nil && *a == *b) }

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func TestBuild_TituloSlug(t *testing.T) {
	mtime := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	ev := func(name, issue, spec, grill string) events.Event {
		return events.Event{V: 1, TS: "2026-10-01T10:00:00Z", Harness: "claude", Event: name, Issue: issue, Spec: spec, Grill: grill}
	}
	conH1 := artifact("spec", "vieja", "<!-- SDD-Tracking: version=1; type=spec; state=approved; issue=none; grill=none; superseded-by=none -->", mtime)
	conH1.Title = "Spec — Título real"
	tests := []struct {
		name       string
		data       scan.Data
		wantRef    string
		wantID     string
		wantTitulo string
	}{
		{"spec solo en eventos: el slug sin ruta ni extensión",
			scan.Data{Events: []events.Event{ev("spec.published", "", ".sdd/specs/foo-bar.md", "")}},
			".sdd/specs/foo-bar.md", "o/r/spec/.sdd%2Fspecs%2Ffoo-bar.md", "foo-bar"},
		{"grill sin H1: el slug sin el prefijo de fecha",
			scan.Data{Events: []events.Event{ev("grill.finalized", "", "", "2026-10-01-norte")}},
			"2026-10-01-norte", "o/r/grill/2026-10-01-norte", "norte"},
		{"issue sin título: la referencia tal cual",
			scan.Data{Events: []events.Event{ev("spec.published", "#9", ".sdd/specs/x.md", "")}},
			"#9", "o/r/issue/9", "#9"},
		{"ruta de spec anidada: la referencia tal cual",
			scan.Data{Events: []events.Event{ev("spec.published", "", ".sdd/specs/sub/anidada.md", "")}},
			".sdd/specs/sub/anidada.md", "o/r/spec/.sdd%2Fspecs%2Fsub%2Fanidada.md", ".sdd/specs/sub/anidada.md"},
		{"grill sin prefijo de fecha: la referencia tal cual",
			scan.Data{Events: []events.Event{ev("grill.finalized", "", "", "norte")}},
			"norte", "o/r/grill/norte", "norte"},
		{"con H1 en la spec el título es el H1, no el slug",
			scan.Data{Specs: []scan.Artifact{conH1}},
			".sdd/specs/vieja.md", "o/r/spec/.sdd%2Fspecs%2Fvieja.md", "Título real"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.data.Repo = scan.Repo{ID: "o/r"}
			got := Build([]RepoInput{{Data: tc.data}})
			if len(got) != 1 {
				t.Fatalf("trabajos = %d, want 1", len(got))
			}
			if got[0].Titulo != tc.wantTitulo {
				t.Errorf("Titulo = %q, want %q", got[0].Titulo, tc.wantTitulo)
			}
			if got[0].Ancla.Ref != tc.wantRef || got[0].ID != tc.wantID {
				t.Errorf("Ancla.Ref = %q, ID = %q; want %q, %q (la identidad no cambia)", got[0].Ancla.Ref, got[0].ID, tc.wantRef, tc.wantID)
			}
		})
	}
}
