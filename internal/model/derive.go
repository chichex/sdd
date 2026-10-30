package model

import (
	"fmt"
	"slices"
	"time"

	"github.com/chichex/sdd/internal/events"
	"github.com/chichex/sdd/internal/github"
	"github.com/chichex/sdd/internal/scan"
	"github.com/chichex/sdd/internal/tracking"
)

const tsLayout = "2006-01-02T15:04:05Z"

func formatTS(t time.Time) string { return t.UTC().Format(tsLayout) }

var specLabel = map[string]string{"draft": "borrador", "approved": "aprobada", "implemented": "implementada", "superseded": "reemplazada"}

var grillLabel = map[string]string{"paused": "pausado", "finalized": "finalizado"}

var eventLabel = map[string]string{
	"grill.paused": "pausado", "grill.finalized": "finalizado",
	"spec.published": "borrador", "spec.approved": "aprobada",
}

// deriver arma las ramas de un grupo: la viva (principal) y las cerradas.
type deriver struct {
	repo      scan.Repo
	grp       *group
	gh        *github.Snapshot
	stale     bool
	tags      Tags
	principal []Nodo
	closed    []Rama
	prList    []int
	prTS      string
	merged    map[int]bool
	tagged    bool
}

func (d *deriver) issueRef() *string {
	if d.grp.anchor.Tipo == "issue" {
		return ptr(d.grp.anchor.Ref)
	}
	return nil
}

func (d *deriver) evs(names ...string) []events.Event {
	var out []events.Event
	for _, it := range d.grp.items {
		if it.ev != nil && slices.Contains(names, it.ev.Event) {
			out = append(out, *it.ev)
		}
	}
	return out
}

func (d *deriver) arts(kind string) []item {
	var out []item
	for _, it := range d.grp.items {
		if it.art != nil && it.art.Kind == kind {
			out = append(out, it)
		}
	}
	return out
}

// newest devuelve el artefacto más nuevo por mtime.
func newest(items []item) (item, bool) {
	if len(items) == 0 {
		return item{}, false
	}
	best := items[0]
	for _, it := range items[1:] {
		if it.art.ModTime.After(best.art.ModTime) {
			best = it
		}
	}
	return best, true
}

// firstRef devuelve el primer evento que nombra un grill o una spec sin
// transicionarla: la etapa existe aunque su historia no esté en este log.
func (d *deriver) firstRef(match func(item) bool) (string, bool) {
	for _, it := range d.grp.items {
		if it.ev != nil && match(it) {
			return it.ev.TS, true
		}
	}
	return "", false
}

func (d *deriver) grillNodes() {
	var refs []string
	for _, it := range d.grp.items {
		if it.grill != "" && !slices.Contains(refs, it.grill) {
			refs = append(refs, it.grill)
		}
	}
	for _, ref := range refs {
		var evs []events.Event
		for _, ev := range d.evs("grill.paused", "grill.finalized") {
			if ev.Grill == ref {
				evs = append(evs, ev)
			}
		}
		var arts []item
		for _, it := range d.arts("grill") {
			if it.grill == ref {
				arts = append(arts, it)
			}
		}
		n := Nodo{Etapa: Grill, SubEstado: "finalizado", Refs: Refs{Issue: d.issueRef(), Grill: ptr(ref)}}
		if len(evs) > 0 {
			last := evs[len(evs)-1]
			n.SubEstado, n.TS, n.inicio = eventLabel[last.Event], last.TS, evs[0].TS
		}
		if a, ok := newest(arts); ok {
			if label, known := grillLabel[a.art.Tracking.State.Value]; known {
				n.SubEstado = label
			}
			n.Ruta = a.art.Path
			if n.TS == "" {
				n.TS = formatTS(a.art.ModTime)
			}
		}
		if n.TS == "" {
			n.TS, _ = d.firstRef(func(it item) bool { return it.grill == ref })
		}
		d.principal = append(d.principal, n)
	}
}

func (d *deriver) supersededSpecs() []string {
	var ids []string
	for _, ev := range d.evs("spec.superseded") {
		if id := d.specID(ev.Spec); !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	for _, it := range d.arts("spec") {
		if it.art.Tracking.State.Value == "superseded" && !slices.Contains(ids, it.spec) {
			ids = append(ids, it.spec)
		}
	}
	return ids
}

func (d *deriver) specID(spec string) string {
	for _, it := range d.grp.items {
		if it.ev != nil && it.ev.Spec == spec {
			return it.spec
		}
	}
	return spec
}

func (d *deriver) specNodes() {
	superseded := d.supersededSpecs()
	live := func(it item) bool { return it.spec != "" && !slices.Contains(superseded, it.spec) }
	n := Nodo{Etapa: Spec, Refs: Refs{Issue: d.issueRef()}}
	var stateEvents []events.Event
	for _, it := range d.grp.items {
		if it.ev == nil || !live(it) {
			continue
		}
		if n.Refs.Spec == nil || tracking.IsIssueRef(it.spec) {
			n.Refs.Spec = ptr(it.spec)
		}
		if it.ev.Event == "spec.published" || it.ev.Event == "spec.approved" {
			stateEvents = append(stateEvents, *it.ev)
		}
	}
	if len(stateEvents) > 0 {
		last := stateEvents[len(stateEvents)-1]
		n.SubEstado, n.TS, n.inicio = eventLabel[last.Event], last.TS, stateEvents[0].TS
	}
	var arts []item
	for _, it := range d.arts("spec") {
		if live(it) {
			arts = append(arts, it)
		}
	}
	if a, ok := newest(arts); ok {
		if label, known := specLabel[a.art.Tracking.State.Value]; known {
			n.SubEstado = label
		} else if n.SubEstado == "" {
			n.SubEstado = "desconocido"
		}
		n.Ruta = a.art.Path
		if n.Refs.Spec == nil {
			n.Refs.Spec = ptr(a.spec)
		}
		if n.TS == "" {
			n.TS = formatTS(a.art.ModTime)
		}
	}
	if n.Refs.Spec != nil && n.SubEstado == "" {
		n.SubEstado = "aprobada" // una spec que ya corrió fue aceptada por el run
		n.TS, _ = d.firstRef(live)
	}
	if n.Refs.Spec != nil {
		d.principal = append(d.principal, n)
	}
	d.supersededRamas(superseded)
}

func (d *deriver) supersededRamas(ids []string) {
	for i, id := range ids {
		n := Nodo{Etapa: Spec, SubEstado: "reemplazada", Refs: Refs{Issue: d.issueRef(), Spec: ptr(id)}}
		r := Rama{ID: fmt.Sprintf("spec-%d", i+1), Estado: "reemplazada"}
		for _, it := range d.grp.items {
			switch {
			case it.spec != id:
			case it.ev != nil && n.inicio == "":
				n.inicio, n.TS = it.ev.TS, it.ev.TS
				if it.ev.Event == "spec.superseded" {
					r.fin = it.ev.TS
				}
			case it.ev != nil:
				n.TS = it.ev.TS
				if it.ev.Event == "spec.superseded" {
					r.fin = it.ev.TS
				}
			case it.art != nil:
				n.Ruta = it.art.Path
				if n.TS == "" {
					n.TS = formatTS(it.art.ModTime)
				}
			}
		}
		r.Nodos = []Nodo{n}
		d.closed = append(d.closed, r)
	}
}
