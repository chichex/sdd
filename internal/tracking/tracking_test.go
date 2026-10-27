package tracking

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type expectedFixture struct {
	Kind         Kind   `json:"kind"`
	Type         string `json:"type"`
	State        string `json:"state"`
	Provenance   string `json:"provenance"`
	Raw          string `json:"raw"`
	Diagnostic   string `json:"diagnostic"`
	Issue        string `json:"issue"`
	Grill        string `json:"grill"`
	Project      string `json:"project"`
	GeneratedAt  string `json:"generatedAt"`
	SupersededBy string `json:"supersededBy"`
}

func loadExpected(t *testing.T) map[string]expectedFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/tracking/expected.json")
	if err != nil {
		t.Fatalf("leer expected.json: %v", err)
	}
	var doc struct {
		Fixtures map[string]expectedFixture `json:"fixtures"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("expected.json: %v", err)
	}
	return doc.Fixtures
}

func TestParse_Fixtures(t *testing.T) {
	expected := loadExpected(t)
	files, err := filepath.Glob("../../testdata/tracking/*.md")
	if err != nil || len(files) != 21 || len(expected) != 21 {
		t.Fatalf("fixtures = %d, esperados = %d, err %v; want 21 y 21", len(files), len(expected), err)
	}
	for _, file := range files {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			want, ok := expected[name]
			if !ok {
				t.Fatalf("%s no figura en expected.json", name)
			}
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("leer: %v", err)
			}
			got := Parse(string(data))
			if got.Kind != want.Kind || got.State.Value != want.State || got.State.Provenance != want.Provenance {
				t.Fatalf("Parse = kind %q state %q/%q, want %q state %q/%q", got.Kind, got.State.Value, got.State.Provenance, want.Kind, want.State, want.Provenance)
			}
			if want.Raw != "" && got.State.Raw != want.Raw {
				t.Fatalf("state raw = %q, want %q", got.State.Raw, want.Raw)
			}
			firstDiag := ""
			if len(got.Diagnostics) > 0 {
				firstDiag = got.Diagnostics[0].Code
			}
			if firstDiag != want.Diagnostic {
				t.Fatalf("primer diagnóstico = %q, want %q", firstDiag, want.Diagnostic)
			}
			m := got.Metadata
			if want.Type != "" && m.Type != want.Type {
				t.Fatalf("type = %q, want %q", m.Type, want.Type)
			}
			if m.Issue != want.Issue || m.Grill != want.Grill || m.Project != want.Project || m.GeneratedAt != want.GeneratedAt || m.SupersededBy != want.SupersededBy {
				t.Fatalf("identidad = %+v, want issue %q grill %q project %q generatedAt %q supersededBy %q", m, want.Issue, want.Grill, want.Project, want.GeneratedAt, want.SupersededBy)
			}
		})
	}
}

func trackingVectors(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile("../../testdata/vectors/sdd-tracking-v1.vectors.md")
	if err != nil {
		t.Fatalf("leer vectores: %v", err)
	}
	text := string(data)
	start := strings.Index(text, "```html\n")
	end := strings.LastIndex(text, "\n```")
	if start < 0 || end <= start {
		t.Fatalf("bloque de vectores mal formado")
	}
	return strings.Split(text[start+len("```html\n"):end], "\n")
}

func TestParse_VectoresRoundTrip(t *testing.T) {
	vectors := trackingVectors(t)
	if len(vectors) != 3 {
		t.Fatalf("vectores = %d, want 3", len(vectors))
	}
	for _, marker := range vectors {
		got := Parse("# Artefacto\n" + marker + "\n\n## Cuerpo\n")
		if got.Kind != KindCanonical {
			t.Fatalf("Parse(%s) = %q %+v, want canonical", marker, got.Kind, got.Diagnostics)
		}
		out, err := Serialize(got.Metadata)
		if err != nil || out != marker {
			t.Fatalf("Serialize = %q, %v; want %q", out, err, marker)
		}
	}
	first := Parse(vectors[0])
	if first.Metadata.Grill != "session 9/á" || first.Metadata.Issue != "owner/repo#9" {
		t.Fatalf("vector de spec decodificado = %+v", first.Metadata)
	}
}

func TestSerialize_SoloCanonica(t *testing.T) {
	tests := []struct {
		name string
		meta Metadata
	}{
		{name: "metadata legacy sin versión", meta: Metadata{Type: "spec", State: "approved"}},
		{name: "spec reemplazada sin superseded-by", meta: Metadata{Version: 1, Type: "spec", State: "superseded"}},
		{name: "spec aprobada con superseded-by", meta: Metadata{Version: 1, Type: "spec", State: "approved", SupersededBy: "x"}},
		{name: "grill sin project", meta: Metadata{Version: 1, Type: "grill", State: "paused", Grill: "g"}},
		{name: "estado unknown", meta: Metadata{Version: 1, Type: "spec", State: "unknown"}},
		{name: "fecha imposible", meta: Metadata{Version: 1, Type: "project", GeneratedAt: "2026-02-30"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if out, err := Serialize(tc.meta); err == nil {
				t.Fatalf("Serialize = %q, want error", out)
			}
		})
	}
}

func TestParse_Preambulo(t *testing.T) {
	marker := "<!-- SDD-Tracking: version=1; type=spec; state=draft; issue=none; grill=none; superseded-by=none -->"
	tests := []struct {
		name  string
		doc   string
		kind  Kind
		state string
	}{
		{name: "BOM al inicio se conserva fuera del H1", doc: "\xEF\xBB\xBF# Spec — x\n" + marker + "\n## A\n", kind: KindCanonical, state: "draft"},
		{name: "marker después del primer H2 es cuerpo", doc: "# Notas\n\n## A\n" + marker + "\n", kind: KindAbsent, state: "unknown"},
		{name: "un ## dentro de un fence no corta el preámbulo", doc: "# Spec — x\n```md\n## no es H2\n```\n" + marker + "\n## A\n", kind: KindCanonical, state: "draft"},
		{name: "fence de tildes", doc: "# Spec — x\n~~~\n## tampoco\n~~~\n" + marker + "\n", kind: KindCanonical, state: "draft"},
		{name: "documento sin señales es ausente", doc: "# Notas\n\nTexto libre.\n", kind: KindAbsent, state: "unknown"},
		{name: "mención en prosa no es marker", doc: "# Notas\nSDD-Tracking: version=1; type=spec\n", kind: KindAbsent, state: "unknown"},
		{name: "H1 Spec con dos puntos es spec legacy", doc: "# Spec: Algo\n<!-- Estado: reemplazada por .sdd/specs/Nueva.md -->\n", kind: KindLegacy, state: "superseded"},
		{name: "estados legacy divergentes son conflicto", doc: "# Spec — x\n<!-- Estado: aprobada -->\n<!-- Status: draft -->\n", kind: KindConflict, state: "unknown"},
		{name: "H1 Spec y estado de grill son conflicto de tipo", doc: "# Spec — x\n<!-- Estado: paused -->\n", kind: KindConflict, state: "unknown"},
		{name: "Resultado de ejecucion con sufijo infiere implementada", doc: "# Spec — x\n\n## Resultado de ejecución (2026-10-27 · HEAD abc)\n", kind: KindLegacy, state: "implemented"},
		{name: "estado explícito gana a la inferencia", doc: "# Spec — x\n<!-- Estado: borrador -->\n## Resultado de ejecucion\n", kind: KindLegacy, state: "draft"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Parse(tc.doc)
			if got.Kind != tc.kind || got.State.Value != tc.state {
				t.Fatalf("Parse = %q/%q %+v, want %q/%q", got.Kind, got.State.Value, got.Diagnostics, tc.kind, tc.state)
			}
		})
	}
	superseded := Parse("# Spec: Algo\n<!-- Estado: reemplazada por .sdd/specs/Nueva.md -->\n")
	if superseded.Metadata.SupersededBy != ".sdd/specs/Nueva.md" {
		t.Fatalf("supersededBy = %q, want la ref cruda con mayúsculas", superseded.Metadata.SupersededBy)
	}
}

func TestParse_AliasesLegacy(t *testing.T) {
	tests := []struct{ raw, want string }{
		{"DRAFT", "draft"}, {"Borrador", "draft"}, {"APROBADA", "approved"},
		{"pendiente   de EJECUCIÓN", "approved"}, {"Implementada", "implemented"},
		{"reemplazada", "superseded"}, {"Reemplazada por #12", "superseded"},
		{"paused", "paused"}, {"finalized", "finalized"}, {"en revisión", "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			got := Parse("# Spec — x\n<!-- Estado: " + tc.raw + " -->\n")
			if tc.want == "paused" || tc.want == "finalized" {
				got = Parse("# Grill — x\n<!-- Estado: " + tc.raw + " -->\n")
			}
			if got.State.Value != tc.want || got.State.Provenance != "legacy-explicit" || got.State.Raw != tc.raw {
				t.Fatalf("estado = %+v, want %q legacy-explicit raw %q", got.State, tc.want, tc.raw)
			}
		})
	}
}
