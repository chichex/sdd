package model

import (
	"regexp"
	"slices"
	"strconv"

	"github.com/chichex/sdd/internal/github"
)

// bodyLink reconoce `Closes #NN` y `Refs #NN` (también `owner/repo#NN`).
var bodyLink = regexp.MustCompile(`(?i)\b(?:closes|refs)\s+([A-Za-z0-9._-]+/[A-Za-z0-9._-]+)?#([1-9][0-9]*)`)

// linkedPRs devuelve, por ancla de issue local, los PRs del snapshot que no
// figuran en ningún run.pr-opened del repo y cuyo body los vincula con ese
// issue. Solo cuentan si el trabajo del issue tiene spec (lo decide derive).
func (g *grouper) linkedPRs(snap *github.Snapshot) map[string][]int {
	out := map[string][]int{}
	if snap == nil {
		return out
	}
	for _, pr := range snap.PRs {
		if _, opened := g.prAnchor[pr.Number]; opened {
			continue
		}
		for _, m := range bodyLink.FindAllStringSubmatch(pr.Body, -1) {
			issue := g.normalizeIssue(m[1] + "#" + m[2])
			key := "issue:" + issue
			if !slices.Contains(out[key], pr.Number) {
				out[key] = append(out[key], pr.Number)
			}
		}
	}
	return out
}

// openIssue busca el issue local abierto que ancla el trabajo.
func (d *deriver) openIssue() (github.Issue, bool) {
	if d.gh == nil || d.grp.anchor.Tipo != "issue" || d.grp.anchor.Ref[0] != '#' {
		return github.Issue{}, false
	}
	n, err := strconv.Atoi(d.grp.anchor.Ref[1:])
	if err != nil {
		return github.Issue{}, false
	}
	for _, is := range d.gh.Issues {
		if is.Number == n {
			return is, true
		}
	}
	return github.Issue{}, false
}

// issueClosed: el issue local no figura en una lista completa de abiertos.
func (d *deriver) issueClosed() bool {
	_, open := d.openIssue()
	return d.gh != nil && d.gh.IssuesComplete && d.grp.anchor.Tipo == "issue" && d.grp.anchor.Ref[0] == '#' && !open
}

func (d *deriver) ideaNode() {
	is, ok := d.openIssue()
	if !ok {
		return
	}
	n := Nodo{Etapa: Idea, SubEstado: "abierto", TS: is.UpdatedAt, Refs: Refs{Issue: d.issueRef()}, URL: ptr(is.URL), Stale: d.stale}
	d.principal = append([]Nodo{n}, d.principal...)
}

// prState aplica el gate de sdd-land con dato fresco; sin dato fresco el PR
// queda `sin dato`.
func (d *deriver) prState(n int) string {
	if d.stale {
		return "sin dato"
	}
	pr, ok := d.gh.PR(n)
	if !ok {
		return "sin dato"
	}
	return github.SubEstado(pr)
}

// publishedNode suma Publicado: `taggeado` con todos los tags que contienen
// algún merge commit, o `sin dato local` si ningún merge commit está en el clon.
func (d *deriver) publishedNode() {
	var tags []string
	missing, last := false, ""
	for n := range d.merged {
		pr, ok := d.gh.PR(n)
		if !ok || pr.MergeCommit == "" {
			continue
		}
		if d.tags.Missing[pr.MergeCommit] {
			missing = true
			continue
		}
		for _, t := range d.tags.BySHA[pr.MergeCommit] {
			if !slices.Contains(tags, t) {
				tags = append(tags, t)
			}
		}
	}
	for _, node := range d.principal {
		if node.Etapa == Mergeado {
			last = max(last, node.TS)
		}
	}
	slices.Sort(tags)
	switch {
	case len(tags) > 0:
		d.principal = append(d.principal, Nodo{Etapa: Publicado, SubEstado: "taggeado", TS: last, Refs: Refs{Issue: d.issueRef()}, Tags: tags})
		d.tagged = true
	case missing:
		d.principal = append(d.principal, Nodo{Etapa: Publicado, SubEstado: "sin dato local", TS: last, Refs: Refs{Issue: d.issueRef()}})
	}
}
