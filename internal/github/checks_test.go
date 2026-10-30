package github

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func loadRollup(t *testing.T, name string) []Check {
	t.Helper()
	data, err := os.ReadFile("../../testdata/gh/rollups/" + name)
	if err != nil {
		t.Fatalf("leer %s: %v", name, err)
	}
	var doc struct {
		StatusCheckRollup []Check `json:"statusCheckRollup"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return doc.StatusCheckRollup
}

// Los valores esperados son la salida del filtro jq canónico de la Fase 3 de
// sdd-land sobre cada rollup (tabla JQ_CASES de sdd-land-gate.test.ts).
func TestChecks_RollupsDelGate(t *testing.T) {
	tests := []struct {
		fixture string
		want    Verdict
	}{
		{"brik-117.json", Verdict{Estado: "verde", Detenidos: []string{}, Supersedidos: 1}},
		{"cancelled-ultimo.json", Verdict{Estado: "detenido", Detenidos: []string{"CI - Android / android: CANCELLED"}, Supersedidos: 1}},
		{"pendiente.json", Verdict{Estado: "pendiente", Detenidos: []string{}, Supersedidos: 2}},
		{"skipped-tras-failure.json", Verdict{Estado: "detenido", Detenidos: []string{"CI / test: FAILURE"}, Supersedidos: 1}},
		{"neutral-tras-cancelled.json", Verdict{Estado: "detenido", Detenidos: []string{"CI / test: CANCELLED"}, Supersedidos: 1}},
		{"empate-success-cancelled.json", Verdict{Estado: "detenido", Detenidos: []string{"CI / test: CANCELLED"}, Supersedidos: 1}},
		{"empate-cancelled-success.json", Verdict{Estado: "detenido", Detenidos: []string{"CI / test: CANCELLED"}, Supersedidos: 1}},
	}
	for _, tc := range tests {
		t.Run(tc.fixture, func(t *testing.T) {
			got := Checks(loadRollup(t, tc.fixture))
			if got.Estado != tc.want.Estado || got.Supersedidos != tc.want.Supersedidos || !slices.Equal(got.Detenidos, tc.want.Detenidos) {
				t.Fatalf("Checks = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestChecks_SinChecksEsVerde(t *testing.T) {
	if got := Checks(nil); got.Estado != "verde" || got.Supersedidos != 0 || len(got.Detenidos) != 0 {
		t.Fatalf("Checks(nil) = %+v, want verde sin detenidos", got)
	}
}

func TestSubEstado_GateDeSddLand(t *testing.T) {
	green := []Check{{Typename: "CheckRun", Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "SUCCESS", StartedAt: "2026-10-01T10:00:00Z"}}
	red := []Check{{Typename: "CheckRun", Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "FAILURE", StartedAt: "2026-10-01T10:00:00Z"}}
	running := []Check{{Typename: "CheckRun", Name: "test", WorkflowName: "CI", Status: "IN_PROGRESS", StartedAt: "2026-10-01T10:00:00Z"}}
	open := PR{Number: 1, State: "OPEN", Mergeable: "MERGEABLE", Checks: green}
	tests := []struct {
		name string
		pr   func(PR) PR
		want string
	}{
		{"abierto, checks verdes, sin threads: validado", func(p PR) PR { return p }, "validado"},
		{"draft gana a todo", func(p PR) PR { p.IsDraft, p.Checks = true, red; return p }, "draft"},
		{"cambios pedidos", func(p PR) PR { p.ReviewDecision = "CHANGES_REQUESTED"; p.Checks = red; return p }, "cambios pedidos"},
		{"checks rojos", func(p PR) PR { p.Checks = red; return p }, "checks rojos"},
		{"checks pendientes: abierto", func(p PR) PR { p.Checks = running; return p }, "abierto"},
		{"conflicto: abierto", func(p PR) PR { p.Mergeable = "CONFLICTING"; return p }, "abierto"},
		{"thread sin resolver: abierto", func(p PR) PR { p.UnresolvedThreads = 1; return p }, "abierto"},
		{"rollup brik-117 es validado", func(p PR) PR { p.Checks = loadRollup(t, "brik-117.json"); return p }, "validado"},
		{"mergeado", func(p PR) PR { p.State = "MERGED"; return p }, "mergeado"},
		{"cerrado sin merge", func(p PR) PR { p.State = "CLOSED"; return p }, "cerrado"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := SubEstado(tc.pr(open)); got != tc.want {
				t.Fatalf("SubEstado = %q, want %q", got, tc.want)
			}
		})
	}
}
