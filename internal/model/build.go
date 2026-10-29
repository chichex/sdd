package model

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/chichex/sdd/internal/scan"
	"github.com/chichex/sdd/internal/tracking"
)

var etapaRank = map[Etapa]int{Idea: 0, Grill: 1, Spec: 2, Run: 3, PR: 4, Mergeado: 5, Publicado: 6}

// Build arma los trabajos de todos los repos, ordenados por última actividad
// descendente. Los repos sin git no se procesan.
func Build(inputs []RepoInput) []Trabajo {
	var out []Trabajo
	for _, in := range inputs {
		if in.Data.Repo.SinGit {
			continue
		}
		g := newGrouper(in.Data.Repo.ID)
		groups := g.run(in.Data)
		linked := g.linkedPRs(in.GitHub)
		for _, grp := range append(groups, g.ideas(in.GitHub)...) {
			out = append(out, derive(in, grp, linked["issue:"+grp.anchor.Ref]))
		}
	}
	slices.SortStableFunc(out, func(a, b Trabajo) int {
		if c := strings.Compare(b.UltimaActividad, a.UltimaActividad); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	return out
}

func derive(in RepoInput, grp *group, linked []int) Trabajo {
	repo := in.Data.Repo
	d := &deriver{repo: repo, grp: grp, gh: in.GitHub, stale: in.Stale, tags: in.Tags}
	d.grillNodes()
	d.specNodes()
	d.runNodes()
	if d.specRef() == nil || grp.anchor.Tipo != "issue" {
		linked = nil // un PR vinculado por body solo entra a un trabajo con spec
	}
	d.prNodes(linked)
	d.publishedNode()
	d.ideaNode()
	if d.principal == nil {
		d.principal = []Nodo{}
	}
	tr := Trabajo{
		ID:     trabajoID(repo.ID, grp.anchor),
		Repo:   repo.ID,
		Titulo: d.title(),
		Ancla:  grp.anchor,
		Ramas:  append([]Rama{{ID: "principal", Estado: "viva", Nodos: d.principal}}, d.closed...),
	}
	tr.SinHistoria = len(grp.items) > 0 && !slices.ContainsFunc(grp.items, func(it item) bool { return it.ev != nil })
	tr.Etapa, tr.SubEstado = d.current()
	for _, r := range tr.Ramas {
		for _, n := range r.Nodos {
			tr.UltimaActividad = max(tr.UltimaActividad, n.TS)
		}
	}
	tr.Terminado = d.tagged || (d.mergedComplete() && (grp.anchor.Tipo != "issue" || d.issueClosed()))
	tr.Gantt = d.gantt(tr.Ramas)
	tr.Links = d.links()
	return tr
}

// trabajoID es `<repo>/issue/<NN>`, `/grill/<ref>` o `/spec/<ruta>`, con la
// referencia codificada como `<ref>` de SDD-Tracking (salvo el número local).
func trabajoID(repoID string, a Ancla) string {
	ref := tracking.EncodeRef(a.Ref)
	if a.Tipo == "issue" && strings.HasPrefix(a.Ref, "#") {
		ref = a.Ref[1:]
	}
	return repoID + "/" + a.Tipo + "/" + ref
}

// current es la etapa más avanzada de la rama viva y su sub-estado.
func (d *deriver) current() (Etapa, string) {
	etapa, sub := Idea, ""
	for _, n := range d.principal {
		if etapaRank[n.Etapa] < etapaRank[etapa] || (n.Etapa == Publicado && n.SubEstado != "taggeado") {
			continue // sin dato local no dice que se haya publicado
		}
		laterPR := n.Etapa == PR && etapa == PR // el sub-estado de PR es el de la capa 1
		if !laterPR {
			sub = n.SubEstado
		}
		etapa = n.Etapa
	}
	if etapa == Mergeado {
		sub = d.mergedSummary()
	}
	return etapa, sub
}

var titlePrefix = regexp.MustCompile(`(?i)^(?:spec|grill)\s*[:—–-]\s*`)

func cleanTitle(h1 string) string { return strings.TrimSpace(titlePrefix.ReplaceAllString(h1, "")) }

var (
	specSlug  = regexp.MustCompile(`^\.sdd/specs/([^/]+)\.md$`)
	grillSlug = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-(.+)$`)
)

// anchorSlug es el nombre legible de un ancla sin título: el archivo de una
// spec sin su ruta ni extensión y el grill sin su prefijo de fecha. Cualquier
// otra referencia (un issue, una ruta que no es `.sdd/specs/<slug>.md`) queda
// tal cual.
func anchorSlug(a Ancla) string {
	var re *regexp.Regexp
	switch a.Tipo {
	case "spec":
		re = specSlug
	case "grill":
		re = grillSlug
	default:
		return a.Ref
	}
	if m := re.FindStringSubmatch(a.Ref); m != nil {
		return m[1]
	}
	return a.Ref
}

// title: título del issue, si no H1 de la spec más nueva, si no H1 del grill,
// si no el slug del ancla.
func (d *deriver) title() string {
	if is, ok := d.openIssue(); ok && is.Title != "" {
		return is.Title
	}
	for _, kind := range []string{"spec", "grill"} {
		if a, ok := newest(d.arts(kind)); ok {
			if t := cleanTitle(a.art.Title); t != "" {
				return t
			}
		}
	}
	return anchorSlug(d.grp.anchor)
}

func issueURL(repo scan.Repo, issue string) *string {
	gh, ok := repo.GitHubRepo()
	if !ok || !tracking.IsIssueRef(issue) {
		return nil
	}
	owner, num, _ := strings.Cut(issue, "#")
	if owner == "" {
		owner = gh
	}
	return ptr(fmt.Sprintf("https://github.com/%s/issues/%s", owner, num))
}

func (d *deriver) links() Links {
	l := Links{PRs: []PRLink{}, Grills: []string{}, Specs: []string{}}
	if d.grp.anchor.Tipo == "issue" {
		l.Issue = issueURL(d.repo, d.grp.anchor.Ref)
	}
	var prs []int
	for _, n := range d.principal {
		if n.Refs.PR != nil && !slices.Contains(prs, *n.Refs.PR) {
			prs = append(prs, *n.Refs.PR)
			l.PRs = append(l.PRs, PRLink{Numero: *n.Refs.PR, URL: n.URL})
		}
	}
	for _, it := range d.grp.items {
		if it.art == nil {
			continue
		}
		if it.art.Kind == "grill" {
			l.Grills = append(l.Grills, it.art.Path)
		} else {
			l.Specs = append(l.Specs, it.art.Path)
		}
	}
	return l
}

// gantt arma una barra por etapa y por rama con los nodos que salen de un
// evento: inicio = primer evento de la etapa en la rama, fin = inicio de la
// etapa siguiente, el cierre de una rama cerrada o nil si sigue abierta.
func (d *deriver) gantt(ramas []Rama) []Barra {
	var out []Barra
	for _, r := range ramas {
		var stages []Etapa
		starts := map[Etapa]string{}
		estado := map[Etapa]string{}
		for _, n := range r.Nodos {
			if n.inicio == "" {
				continue
			}
			if _, seen := starts[n.Etapa]; !seen {
				stages = append(stages, n.Etapa)
				starts[n.Etapa], estado[n.Etapa] = n.inicio, n.SubEstado
			}
			starts[n.Etapa] = min(starts[n.Etapa], n.inicio)
		}
		if r.ID == "principal" && starts[Mergeado] != "" {
			estado[Mergeado] = d.mergedSummary()
		}
		for i, e := range stages {
			b := Barra{Rama: r.ID, Etapa: e, Inicio: starts[e], Estado: estado[e]}
			switch {
			case i+1 < len(stages):
				b.Fin = ptr(starts[stages[i+1]])
			case r.fin != "":
				b.Fin = ptr(r.fin)
			}
			out = append(out, b)
		}
	}
	return out
}
