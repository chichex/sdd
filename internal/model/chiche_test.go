package model

import (
	"testing"
	"time"

	"github.com/chichex/visor/internal/events"
	"github.com/chichex/visor/internal/scan"
	"github.com/chichex/visor/internal/tracking"
)

func chicheEv(ts, name, spec string, pr ...int) events.Event {
	return events.Event{V: 1, TS: ts, Harness: "claude", Event: name, Spec: spec, PR: pr}
}

func TestBuild_SpecMigradaEsUnSoloTrabajo(t *testing.T) {
	data := scan.Data{Repo: scan.Repo{ID: "o/r"}, Events: []events.Event{
		chicheEv("2026-10-01T10:00:00Z", "spec.published", ".sdd/specs/x.md"),
		chicheEv("2026-10-02T10:00:00Z", "run.started", ".chiche/specs/x.md"),
	}}
	tr := Build([]RepoInput{{Data: data}})
	if len(tr) != 1 {
		t.Fatalf("trabajos = %d, want 1 (la misma spec con las dos marcas)", len(tr))
	}
	if tr[0].Ancla != (Ancla{Tipo: "spec", Ref: ".chiche/specs/x.md"}) || tr[0].ID != "o/r/spec/.chiche%2Fspecs%2Fx.md" {
		t.Fatalf("ancla = %+v id %q, want .chiche/specs/x.md (un evento la nombra con chiche)", tr[0].Ancla, tr[0].ID)
	}
	if tr[0].Etapa != Run || tr[0].Titulo != "x" {
		t.Fatalf("etapa = %q, título = %q; want run y el slug x", tr[0].Etapa, tr[0].Titulo)
	}
}

func TestBuild_SpecEnDiscoDefineLaIdentidad(t *testing.T) {
	spec := artifact("spec", "x", "<!-- SDD-Tracking: version=1; type=spec; state=approved; issue=none; grill=none; superseded-by=none -->", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	data := scan.Data{Repo: scan.Repo{ID: "o/r"}, Specs: []scan.Artifact{spec}, Events: []events.Event{
		chicheEv("2026-10-02T10:00:00Z", "run.started", ".chiche/specs/x.md"),
	}}
	tr := Build([]RepoInput{{Data: data}})
	if len(tr) != 1 || tr[0].Ancla.Ref != ".sdd/specs/x.md" || tr[0].Etapa != Run {
		t.Fatalf("trabajos = %+v, want uno anclado en .sdd/specs/x.md (la ruta en disco) y en run", tr)
	}
}

func TestBuild_SpecMovidaConLaOtraMarca(t *testing.T) {
	moved := chicheEv("2026-10-02T10:00:00Z", "spec.moved", ".sdd/specs/b.md")
	moved.Previous = ".sdd/specs/a.md"
	data := scan.Data{Repo: scan.Repo{ID: "o/r"}, Events: []events.Event{
		chicheEv("2026-10-01T10:00:00Z", "spec.published", ".sdd/specs/a.md"),
		moved,
		chicheEv("2026-10-03T10:00:00Z", "run.started", ".chiche/specs/a.md"),
	}}
	tr := Build([]RepoInput{{Data: data}})
	if len(tr) != 1 || tr[0].Ancla.Ref != ".sdd/specs/b.md" || tr[0].Etapa != Run {
		t.Fatalf("trabajos = %+v, want uno en .sdd/specs/b.md (el spec.moved vale para las dos marcas) y en run", tr)
	}
}

func TestBuild_SpecsDistintasNoSeJuntan(t *testing.T) {
	data := scan.Data{Repo: scan.Repo{ID: "o/r"}, Events: []events.Event{
		chicheEv("2026-10-01T10:00:00Z", "spec.published", ".sdd/specs/x.md"),
		chicheEv("2026-10-02T10:00:00Z", "spec.published", ".chiche/specs/y.md"),
	}}
	if tr := Build([]RepoInput{{Data: data}}); len(tr) != 2 {
		t.Fatalf("trabajos = %d, want 2 (nombres distintos)", len(tr))
	}
}

func TestBuild_ArtefactoChicheV2DefineLaIdentidad(t *testing.T) {
	text := "# Spec — x\n<!-- Chiche-Tracking: version=2; type=spec; state=approved; issue=none; grill=none; superseded-by=none -->\n"
	spec := scan.Artifact{Kind: "spec", Name: "x.md", Path: "/r/.chiche/specs/x.md", Logical: ".chiche/specs/x.md",
		Title: "Spec — Migrada", ModTime: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Tracking: tracking.Parse(text)}
	data := scan.Data{Repo: scan.Repo{ID: "o/r"}, Specs: []scan.Artifact{spec}, Events: []events.Event{
		chicheEv("2026-10-01T10:00:00Z", "spec.published", ".sdd/specs/x.md"),
		chicheEv("2026-10-02T10:00:00Z", "run.started", ".sdd/specs/x.md"),
	}}
	tr := Build([]RepoInput{{Data: data}})
	if len(tr) != 1 || tr[0].Ancla.Ref != ".chiche/specs/x.md" || tr[0].ID != "o/r/spec/.chiche%2Fspecs%2Fx.md" {
		t.Fatalf("trabajos = %+v, want uno anclado en .chiche/specs/x.md (el artefacto v2 en disco)", tr)
	}
	if tr[0].Etapa != Run || tr[0].Titulo != "Migrada" || len(tr[0].Links.Specs) != 1 || tr[0].Links.Specs[0] != "/r/.chiche/specs/x.md" {
		t.Fatalf("etapa %q título %q links %v; want run, Migrada y el archivo de .chiche/", tr[0].Etapa, tr[0].Titulo, tr[0].Links.Specs)
	}
}

func TestBuild_SpecMovidaSoloDeMarca(t *testing.T) {
	moved := chicheEv("2026-10-02T10:00:00Z", "spec.moved", ".chiche/specs/a.md")
	moved.Previous = ".sdd/specs/a.md"
	data := scan.Data{Repo: scan.Repo{ID: "o/r"}, Events: []events.Event{
		chicheEv("2026-10-01T10:00:00Z", "spec.published", ".sdd/specs/a.md"),
		moved,
		chicheEv("2026-10-03T10:00:00Z", "run.started", ".chiche/specs/a.md"),
	}}
	tr := Build([]RepoInput{{Data: data}})
	if len(tr) != 1 || tr[0].Ancla.Ref != ".chiche/specs/a.md" || tr[0].Etapa != Run {
		t.Fatalf("trabajos = %+v, want uno en .chiche/specs/a.md y en run", tr)
	}
}
