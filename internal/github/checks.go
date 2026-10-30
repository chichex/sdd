package github

import (
	"cmp"
	"slices"
	"strings"
)

// Verdict es la salida del filtro de checks de la Fase 3 de sdd-land.
type Verdict struct {
	Estado       string   `json:"estado"`
	Detenidos    []string `json:"detenidos"`
	Supersedidos int      `json:"supersedidos"`
}

type evaluated struct {
	id, t, v, r string
}

func checkTime(c Check) string {
	for _, t := range []string{c.StartedAt, c.CompletedAt} {
		if t != "" && !strings.HasPrefix(t, "0001-") {
			return t
		}
	}
	return "9999" // todavía no arrancó: es el más reciente
}

func checkID(c Check) string {
	if c.Typename == "StatusContext" {
		return c.Context
	}
	return c.WorkflowName + " / " + c.Name
}

func checkVerdict(c Check) string {
	if c.Typename == "StatusContext" {
		switch c.State {
		case "SUCCESS":
			return "verde"
		case "PENDING", "EXPECTED":
			return "pendiente"
		}
		return "detenido"
	}
	if c.Status != "COMPLETED" {
		return "pendiente"
	}
	switch c.Conclusion {
	case "SUCCESS", "NEUTRAL", "SKIPPED":
		return "verde"
	}
	return "detenido"
}

var weight = map[string]int{"verde": 0, "pendiente": 1, "detenido": 2}

// Checks porta el filtro jq canónico de sdd-land: deduplica por identidad
// (`workflowName / name` o `context`), cuenta el más reciente por
// startedAt/completedAt, un SKIPPED o NEUTRAL nunca supersede a un run real y
// ante un empate gana el veredicto conservador.
func Checks(rollup []Check) Verdict {
	groups := map[string][]evaluated{}
	for _, c := range rollup {
		r := c.Conclusion
		if r == "" {
			r = c.State
		}
		e := evaluated{id: checkID(c), t: checkTime(c), v: checkVerdict(c), r: r}
		groups[e.id] = append(groups[e.id], e)
	}
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	out := Verdict{Estado: "verde", Detenidos: []string{}}
	for _, id := range ids {
		group := groups[id]
		out.Supersedidos += len(group) - 1
		slices.SortStableFunc(group, func(a, b evaluated) int {
			return cmp.Or(strings.Compare(a.t, b.t), weight[a.v]-weight[b.v])
		})
		current := group[len(group)-1]
		for _, e := range group {
			if e.r != "SKIPPED" && e.r != "NEUTRAL" {
				current = e // el último run real del orden gana a un omitido
			}
		}
		switch {
		case current.v == "detenido":
			out.Detenidos = append(out.Detenidos, current.id+": "+current.r)
			out.Estado = "detenido"
		case current.v == "pendiente" && out.Estado == "verde":
			out.Estado = "pendiente"
		}
	}
	return out
}

// SubEstado deriva el sub-estado de un PR con el gate de sdd-land: draft,
// cambios pedidos, checks rojos, validado (abierto, no draft, sin conflicto,
// sin cambios pedidos, checks verdes y cero threads sin resolver) o abierto.
// Un PR mergeado o cerrado lo dice.
func SubEstado(pr PR) string {
	switch {
	case pr.State == "MERGED":
		return "mergeado"
	case pr.State == "CLOSED":
		return "cerrado"
	case pr.IsDraft:
		return "draft"
	case pr.ReviewDecision == "CHANGES_REQUESTED":
		return "cambios pedidos"
	}
	checks := Checks(pr.Checks).Estado
	switch {
	case checks == "detenido":
		return "checks rojos"
	case pr.Mergeable != "CONFLICTING" && checks == "verde" && pr.UnresolvedThreads == 0:
		return "validado"
	}
	return "abierto"
}
