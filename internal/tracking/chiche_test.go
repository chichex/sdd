package tracking

import (
	"os"
	"strings"
	"testing"
)

func chicheVectors(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile("../../testdata/vectors/chiche-tracking-v2.vectors.md")
	if err != nil {
		t.Fatalf("leer vectores v2: %v", err)
	}
	text := string(data)
	start := strings.Index(text, "```html\n")
	end := strings.LastIndex(text, "\n```")
	if start < 0 || end <= start {
		t.Fatalf("bloque de vectores v2 mal formado")
	}
	return strings.Split(text[start+len("```html\n"):end], "\n")
}

func TestParse_ChicheVectoresRoundTrip(t *testing.T) {
	v2, v1 := chicheVectors(t), trackingVectors(t)
	if len(v2) != 3 || len(v1) != 3 {
		t.Fatalf("vectores v2 = %d, v1 = %d; want 3 y 3", len(v2), len(v1))
	}
	for i, marker := range v2 {
		got := Parse("# Artefacto\n" + marker + "\n\n## Cuerpo\n")
		if got.Kind != KindCanonical || got.Metadata.Version != 2 || len(got.Diagnostics) != 0 {
			t.Fatalf("Parse(%s) = %q v%d %+v, want canonical v2 sin diagnósticos", marker, got.Kind, got.Metadata.Version, got.Diagnostics)
		}
		out, err := Serialize(got.Metadata)
		if err != nil || out != marker {
			t.Fatalf("Serialize = %q, %v; want %q", out, err, marker)
		}
		old := Parse(v1[i]).Metadata
		old.Version = 2
		if old != got.Metadata {
			t.Fatalf("v2 decodificada = %+v, want la misma identidad que v1 %+v", got.Metadata, old)
		}
	}
	if got := Parse(v1[0]); got.Kind != KindCanonical || got.Metadata.Version != 1 {
		t.Fatalf("SDD-Tracking v1 = %q v%d, want canonical v1", got.Kind, got.Metadata.Version)
	}
}

func TestParse_ChicheInvalidos(t *testing.T) {
	spec := "type=spec; state=approved; issue=none; grill=none; superseded-by=none -->"
	tests := []struct{ name, marker, code string }{
		{"Chiche-Tracking con version=1", "<!-- Chiche-Tracking: version=1; " + spec, "marker-version-mismatch"},
		{"SDD-Tracking con version=2", "<!-- SDD-Tracking: version=2; " + spec, "marker-version-mismatch"},
		{"Chiche-Tracking con version=3", "<!-- Chiche-Tracking: version=3; " + spec, "unsupported-version"},
		{"SDD-Tracking con version=3", "<!-- SDD-Tracking: version=3; " + spec, "unsupported-version"},
		{"Chiche-Tracking sin version ni type no es legacy", "<!-- Chiche-Tracking: estado=aprobada -->", "missing-key"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Parse("# Spec — x\n" + tc.marker + "\n")
			if got.Kind != KindInvalid || len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != tc.code {
				t.Fatalf("Parse = %q %+v, want invalid con %s", got.Kind, got.Diagnostics, tc.code)
			}
			if got.State.Provenance != "absent" || got.Metadata != (Metadata{}) {
				t.Fatalf("estado = %+v, metadata = %+v; want ausentes", got.State, got.Metadata)
			}
		})
	}
}

func TestParse_ChicheYSDDJuntosSonConflicto(t *testing.T) {
	fields := "type=spec; state=approved; issue=none; grill=none; superseded-by=none -->"
	doc := "# Spec — x\n<!-- SDD-Tracking: version=1; " + fields + "\n<!-- Chiche-Tracking: version=2; " + fields + "\n"
	got := Parse(doc)
	if got.Kind != KindConflict || len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != "conflicting-canonical" || len(got.Raw) != 2 {
		t.Fatalf("Parse = %q %+v raw %v, want conflict con conflicting-canonical y los dos markers", got.Kind, got.Diagnostics, got.Raw)
	}
}

func TestParse_ChicheDuplicadoIdentico(t *testing.T) {
	marker := "<!-- Chiche-Tracking: version=2; type=project; generated-at=2026-10-25 -->"
	got := Parse("# Contrato\n" + marker + "\n" + marker + "\n")
	if got.Kind != KindCanonical || got.Metadata.Version != 2 || len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != "duplicate-canonical" {
		t.Fatalf("Parse = %q %+v, want canonical v2 con duplicate-canonical", got.Kind, got.Diagnostics)
	}
}
