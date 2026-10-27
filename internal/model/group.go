package model

import (
	"slices"
	"strconv"
	"strings"

	"github.com/chichex/sdd/internal/events"
	"github.com/chichex/sdd/internal/github"
	"github.com/chichex/sdd/internal/scan"
	"github.com/chichex/sdd/internal/tracking"
)

// item es un evento o un artefacto con su identidad ya resuelta.
type item struct {
	idx   int
	ev    *events.Event
	art   *scan.Artifact
	issue string
	grill string
	spec  string
}

// group junta lo que comparte ancla; los eventos quedan en orden de log.
type group struct {
	anchor Ancla
	items  []item
}

// grouper agrupa por ancla con prioridad issue > grill > spec: lo que nombra un
// issue va a ese issue; lo que no, al issue que comparta su grill o su spec y,
// si no hay, al componente grill/spec que lo une con el resto.
type grouper struct {
	repoID     string
	moved      map[string]string
	groups     map[string]*group
	order      []string
	grillIssue map[string][]Ancla
	specIssue  map[string][]Ancla
	prAnchor   map[int]Ancla
	parent     map[string]string
}

func newGrouper(repoID string) *grouper {
	return &grouper{repoID: repoID, moved: map[string]string{}, groups: map[string]*group{},
		grillIssue: map[string][]Ancla{}, specIssue: map[string][]Ancla{}, prAnchor: map[int]Ancla{}, parent: map[string]string{}}
}

// normalizeIssue lleva `owner/repo#NN` del propio repo a `#NN`.
func (g *grouper) normalizeIssue(issue string) string {
	if i := strings.LastIndex(issue, "#"); i > 0 && strings.EqualFold(issue[:i], g.repoID) {
		return issue[i:]
	}
	return issue
}

// resolveSpec sigue los `spec.moved` hasta la identidad vigente.
func (g *grouper) resolveSpec(spec string) string {
	for hops := 0; hops < 8; hops++ {
		next, ok := g.moved[spec]
		if !ok {
			break
		}
		spec = next
	}
	return spec
}

func (g *grouper) identify(it *item) {
	if it.ev != nil {
		it.issue, it.grill, it.spec = g.normalizeIssue(it.ev.Issue), it.ev.Grill, g.resolveSpec(it.ev.Spec)
	} else {
		meta := it.art.Tracking.Metadata
		it.issue, it.grill = g.normalizeIssue(meta.Issue), meta.Grill
		if it.art.Kind == "grill" && it.grill == "" {
			it.grill = strings.TrimSuffix(it.art.Name, ".md")
		}
		if it.art.Kind == "spec" {
			it.spec = g.resolveSpec(it.art.Logical)
		}
	}
	if it.issue == "" && tracking.IsIssueRef(it.spec) {
		it.issue = g.normalizeIssue(it.spec)
	}
}

func usable(a scan.Artifact) bool {
	k := a.Tracking.Kind
	return (k == tracking.KindCanonical || k == tracking.KindLegacy) && a.Tracking.Metadata.Type == a.Kind
}

func (g *grouper) add(a Ancla, it item) {
	key := a.Tipo + ":" + a.Ref
	grp, ok := g.groups[key]
	if !ok {
		grp = &group{anchor: a}
		g.groups[key] = grp
		g.order = append(g.order, key)
	}
	grp.items = append(grp.items, it)
}

// ideas devuelve un grupo vacío por cada issue abierto que todavía no ancla
// ningún trabajo: entra como Idea.
func (g *grouper) ideas(snap *github.Snapshot) []*group {
	if snap == nil {
		return nil
	}
	var out []*group
	for _, is := range snap.Issues {
		a := Ancla{Tipo: "issue", Ref: "#" + strconv.Itoa(is.Number)}
		if _, ok := g.groups["issue:"+a.Ref]; !ok {
			out = append(out, &group{anchor: a})
		}
	}
	return out
}

func (g *grouper) find(id string) string {
	for g.parent[id] != id {
		id = g.parent[id]
	}
	return id
}

func (g *grouper) union(ids ...string) {
	for _, id := range ids {
		if _, ok := g.parent[id]; !ok {
			g.parent[id] = id
		}
	}
	for _, id := range ids[1:] {
		g.parent[g.find(id)] = g.find(ids[0])
	}
}

func appendAnchor(list []Ancla, a Ancla) []Ancla {
	if slices.Contains(list, a) {
		return list
	}
	return append(list, a)
}

// run agrupa los eventos (en orden de log) y los artefactos usables del repo.
func (g *grouper) run(d scan.Data) []*group {
	for _, ev := range d.Events {
		if ev.Event == "spec.moved" {
			g.moved[ev.Previous] = ev.Spec
		}
	}
	var items []item
	for i := range d.Events {
		items = append(items, item{idx: i, ev: &d.Events[i]})
	}
	for _, arts := range [][]scan.Artifact{d.Grills, d.Specs} {
		for i := range arts {
			if usable(arts[i]) {
				items = append(items, item{idx: len(items), art: &arts[i]})
			}
		}
	}
	var deferred []item
	for i := range items {
		it := &items[i]
		g.identify(it)
		if it.issue == "" {
			deferred = append(deferred, *it)
			continue
		}
		a := Ancla{Tipo: "issue", Ref: it.issue}
		g.add(a, *it)
		g.remember(a, *it)
	}
	g.attachDeferred(deferred)
	result := make([]*group, 0, len(g.order))
	for _, key := range g.order {
		grp := g.groups[key]
		slices.SortStableFunc(grp.items, func(a, b item) int { return a.idx - b.idx })
		result = append(result, grp)
	}
	return result
}

// remember registra qué issue ancla cada grill y cada spec, y los PRs.
func (g *grouper) remember(a Ancla, it item) {
	if it.grill != "" {
		g.grillIssue[it.grill] = appendAnchor(g.grillIssue[it.grill], a)
	}
	if it.spec != "" && !tracking.IsIssueRef(it.spec) {
		g.specIssue[it.spec] = appendAnchor(g.specIssue[it.spec], a)
	}
	g.rememberPRs(a, it)
}

// rememberPRs registra el ancla de los PRs que abrió un run.
func (g *grouper) rememberPRs(a Ancla, it item) {
	if it.ev != nil && it.ev.Event == "run.pr-opened" {
		for _, n := range it.ev.PR {
			if _, ok := g.prAnchor[n]; !ok {
				g.prAnchor[n] = a
			}
		}
	}
}

func (g *grouper) attachDeferred(deferred []item) {
	var loose, merges []item
	for _, it := range deferred {
		if it.ev != nil && it.ev.Event == "land.merged" {
			merges = append(merges, it)
			continue
		}
		var cands []Ancla
		for _, a := range append(append([]Ancla{}, g.grillIssue[it.grill]...), g.specIssue[it.spec]...) {
			cands = appendAnchor(cands, a)
		}
		for _, a := range cands {
			g.add(a, it)
		}
		if len(cands) == 0 {
			loose = append(loose, it)
		} else if len(cands) == 1 {
			g.rememberPRs(cands[0], it)
		}
	}
	g.attachLoose(loose)
	for _, it := range merges {
		if a, ok := g.prAnchor[it.ev.PR[0]]; ok {
			g.add(a, it)
		}
	}
}

func looseIDs(it item) []string {
	var ids []string
	if it.grill != "" {
		ids = append(ids, "grill:"+it.grill)
	}
	if it.spec != "" {
		ids = append(ids, "spec:"+it.spec)
	}
	return ids
}

// attachLoose une grills y specs sin issue por componentes: el ancla de cada
// componente es su primer grill o, si no tiene, su primera spec.
func (g *grouper) attachLoose(loose []item) {
	for _, it := range loose {
		if ids := looseIDs(it); len(ids) > 0 {
			g.union(ids...)
		}
	}
	anchors := map[string]Ancla{}
	for _, preferGrill := range []bool{true, false} {
		for _, it := range loose {
			for _, id := range looseIDs(it) {
				rep := g.find(id)
				tipo, ref, _ := strings.Cut(id, ":")
				if _, done := anchors[rep]; !done && (tipo == "grill") == preferGrill {
					anchors[rep] = Ancla{Tipo: tipo, Ref: ref}
				}
			}
		}
	}
	for _, it := range loose {
		ids := looseIDs(it)
		if len(ids) == 0 {
			continue
		}
		a := anchors[g.find(ids[0])]
		g.add(a, it)
		g.rememberPRs(a, it)
	}
}
