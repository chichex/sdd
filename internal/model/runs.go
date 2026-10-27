package model

import (
	"fmt"
	"slices"
	"strconv"
)

// attempt es un intento de run: de un run.started al siguiente o a su fallo.
type attempt struct {
	n      int
	start  string
	last   string
	state  string
	failed bool
	end    string
}

// runNodes arma los intentos de run en orden de log. Un run.failed cierra su
// intento como rama fallada; un run.started posterior abre uno nuevo. El nodo
// Run de la rama viva es el último intento que no falló.
func (d *deriver) runNodes() {
	var attempts []*attempt
	var current *attempt
	open := func(ts string) *attempt {
		a := &attempt{n: len(attempts) + 1, start: ts, last: ts, state: "corriendo"}
		attempts = append(attempts, a)
		return a
	}
	for _, ev := range d.evs("run.started", "run.pr-opened", "run.committed", "run.failed") {
		if ev.Event == "run.started" || current == nil {
			current = open(ev.TS)
		}
		current.last = ev.TS
		switch ev.Event {
		case "run.pr-opened":
			current.state = "con PR"
			d.prList, d.prTS = ev.PR, ev.TS
		case "run.committed":
			current.state = "commiteado"
		case "run.failed":
			current.state, current.failed, current.end = "fallado ("+ev.Reason+")", true, ev.TS
			current = nil
		}
	}
	var live *attempt
	for _, a := range attempts {
		if !a.failed {
			live = a
		}
	}
	spec := d.specRef()
	if live != nil {
		d.principal = append(d.principal, Nodo{Etapa: Run, SubEstado: live.state, TS: live.last, inicio: live.start,
			Refs: Refs{Issue: d.issueRef(), Spec: spec}})
	}
	for _, a := range attempts {
		if a.failed {
			n := Nodo{Etapa: Run, SubEstado: a.state, TS: a.end, inicio: a.start, Refs: Refs{Issue: d.issueRef(), Spec: spec}}
			d.closed = append(d.closed, Rama{ID: "run-" + strconv.Itoa(a.n), Estado: "fallada", Nodos: []Nodo{n}, fin: a.end})
		}
	}
}

func (d *deriver) specRef() *string {
	for _, n := range d.principal {
		if n.Etapa == Spec {
			return n.Refs.Spec
		}
	}
	return nil
}

func (d *deriver) prURL(n int) *string {
	if gh, ok := d.repo.GitHubRepo(); ok {
		return ptr(fmt.Sprintf("https://github.com/%s/pull/%d", gh, n))
	}
	return nil
}

// prNodes arma un nodo por PR del último run.pr-opened, en orden de capa, y un
// nodo Mergeado por PR con land.merged.
func (d *deriver) prNodes() {
	total := len(d.prList)
	for i, n := range d.prList {
		d.principal = append(d.principal, Nodo{Etapa: PR, SubEstado: "sin dato", TS: d.prTS, inicio: d.prTS,
			Refs: Refs{Issue: d.issueRef(), PR: ptr(n)}, URL: d.prURL(n), Capa: fmt.Sprintf("%d/%d", i+1, total)})
	}
	d.merged = map[int]bool{}
	var extra []Nodo
	var inList []Nodo
	for _, ev := range d.evs("land.merged") {
		n := ev.PR[0]
		if d.merged[n] {
			continue
		}
		d.merged[n] = true
		node := Nodo{Etapa: Mergeado, SubEstado: "mergeado", TS: ev.TS, inicio: ev.TS, Refs: Refs{Issue: d.issueRef(), PR: ptr(n)}, URL: d.prURL(n)}
		if i := indexOf(d.prList, n); i >= 0 {
			node.Capa = fmt.Sprintf("%d/%d", i+1, total)
			inList = append(inList, node)
		} else {
			extra = append(extra, node)
		}
	}
	slices.SortFunc(inList, func(a, b Nodo) int { return indexOf(d.prList, *a.Refs.PR) - indexOf(d.prList, *b.Refs.PR) })
	for i := range extra {
		extra[i].Capa = fmt.Sprintf("%d/%d", len(inList)+i+1, max(total, len(inList)+len(extra)))
	}
	d.principal = append(d.principal, append(inList, extra...)...)
}

func indexOf(list []int, n int) int {
	for i, v := range list {
		if v == n {
			return i
		}
	}
	return -1
}

// mergedSummary devuelve «n/N»: PRs mergeados sobre el total conocido.
func (d *deriver) mergedSummary() string {
	total := len(d.prList)
	for n := range d.merged {
		if indexOf(d.prList, n) < 0 {
			total++
		}
	}
	return fmt.Sprintf("%d/%d", len(d.merged), total)
}

// mergedComplete: todos los PRs del último run (o los conocidos) mergeados.
func (d *deriver) mergedComplete() bool {
	if len(d.merged) == 0 {
		return false
	}
	for _, n := range d.prList {
		if !d.merged[n] {
			return false
		}
	}
	return true
}
