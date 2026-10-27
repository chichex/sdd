package events

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"
)

func readLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leer %s: %v", path, err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

// vectorBlock extrae las líneas del bloque delimitado copiado verbatim del doc
// normativo: un cambio en ese bloque rompe el test aunque el .jsonl no cambie.
func vectorBlock(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile("../../testdata/vectors/sdd-events-v1.vectors.md")
	if err != nil {
		t.Fatalf("leer bloque de vectores: %v", err)
	}
	text := string(data)
	start := strings.Index(text, "```jsonl\n")
	end := strings.LastIndex(text, "\n```")
	if !strings.HasPrefix(text, "<!-- conformance-vectors:start -->") || start < 0 || end <= start {
		t.Fatalf("el bloque de vectores no tiene la forma del doc")
	}
	return strings.Split(text[start+len("```jsonl\n"):end], "\n")
}

func TestParseLine_Vectores(t *testing.T) {
	lines := readLines(t, "../../testdata/vectors/events-v1.jsonl")
	block := vectorBlock(t)
	if len(lines) != 12 {
		t.Fatalf("vectores = %d, want 12", len(lines))
	}
	if strings.Join(block, "\n") != strings.Join(lines, "\n") {
		t.Fatalf("events-v1.jsonl difiere del bloque de vectores del doc")
	}
	seen := map[string]bool{}
	for _, line := range lines {
		got := ParseLine([]byte(line))
		if got.Kind != KindEvent {
			t.Fatalf("ParseLine(%s) = kind %d code %q, want evento válido", line, got.Kind, got.Code)
		}
		seen[got.Event.Event] = true
		out, err := Serialize(got.Event)
		if err != nil || out != line {
			t.Fatalf("Serialize = %q, %v; want %q", out, err, line)
		}
	}
	if len(seen) != 11 {
		t.Fatalf("eventos del catálogo cubiertos = %d, want 11", len(seen))
	}
}

func TestParseLine_DecodificaReferencias(t *testing.T) {
	line := `{"v":1,"ts":"2026-10-14T12:30:05Z","harness":"pi","event":"grill.finalized","issue":"owner/repo#9","spec":"none","previous":"none","grill":"session%209%2F%C3%A1","pr":[],"reason":"none"}`
	got := ParseLine([]byte(line))
	want := Event{V: 1, TS: "2026-10-14T12:30:05Z", Harness: "pi", Event: "grill.finalized", Issue: "owner/repo#9", Grill: "session 9/á", PR: []int{}}
	if got.Kind != KindEvent || got.Event.Grill != want.Grill || got.Event.Issue != want.Issue || got.Event.Spec != "" || got.Event.Reason != "" {
		t.Fatalf("ParseLine = %+v, want %+v", got, want)
	}
}

// backslash arma el escape JSON de la letra a sin escribirlo literal en el fuente.
const backslash = string(rune(92))

const base = `{"v":1,"ts":"2026-10-14T14:00:00Z","harness":"claude","event":"run.started","issue":"#93","spec":"%2393","previous":"none","grill":"none","pr":[],"reason":"none"}`

func TestParseLine_Invalidas(t *testing.T) {
	long := strings.Replace(base, `"grill":"none"`, `"grill":"`+strings.Repeat("a", 900)+`"`, 1)
	tests := []struct {
		name string
		line string
		kind Kind
		code string
	}{
		{name: "no es JSON", line: `{"v":1,`, kind: KindInvalid, code: "invalid-json"},
		{name: "no es un objeto", line: `[1]`, kind: KindInvalid, code: "invalid-json"},
		{name: "espacio extra", line: strings.Replace(base, `"v":1,`, `"v": 1,`, 1), kind: KindInvalid, code: "non-canonical"},
		{name: "escape unicode en vez del byte A", line: strings.Replace(base, `"claude"`, `"cl`+backslash+`u0061ude"`, 1), kind: KindInvalid, code: "non-canonical"},
		{name: "v como 1.0", line: strings.Replace(base, `"v":1,`, `"v":1.0,`, 1), kind: KindInvalid, code: "non-canonical"},
		{name: "falta una clave", line: strings.Replace(base, `"previous":"none",`, ``, 1), kind: KindInvalid, code: "missing-key"},
		{name: "clave extra", line: strings.Replace(base, `"reason":"none"}`, `"reason":"none","x":1}`, 1), kind: KindInvalid, code: "extra-key"},
		{name: "clave repetida", line: strings.Replace(base, `"reason":"none"}`, `"reason":"none","reason":"none"}`, 1), kind: KindInvalid, code: "duplicate-key"},
		{name: "claves desordenadas", line: `{"ts":"2026-10-14T14:00:00Z","v":1,"harness":"claude","event":"run.started","issue":"#93","spec":"%2393","previous":"none","grill":"none","pr":[],"reason":"none"}`, kind: KindInvalid, code: "key-order"},
		{name: "más de 1024 bytes", line: long, kind: KindInvalid, code: "line-too-long"},
		{name: "versión 2", line: strings.Replace(base, `"v":1,`, `"v":2,`, 1), kind: KindUnsupported, code: "unsupported-version"},
		{name: "sin v", line: strings.Replace(base, `"v":1,`, ``, 1), kind: KindInvalid, code: "missing-key"},
		{name: "ts fuera de calendario", line: strings.Replace(base, `2026-10-14T`, `2026-02-30T`, 1), kind: KindInvalid, code: "invalid-ts"},
		{name: "ts con milisegundos", line: strings.Replace(base, `14:00:00Z`, `14:00:00.5Z`, 1), kind: KindInvalid, code: "invalid-ts"},
		{name: "harness desconocido", line: strings.Replace(base, `"claude"`, `"cursor"`, 1), kind: KindInvalid, code: "invalid-harness"},
		{name: "evento desconocido", line: strings.Replace(base, `"run.started"`, `"run.paused"`, 1), kind: KindInvalid, code: "invalid-event"},
		{name: "issue desconocido", line: strings.Replace(base, `"issue":"#93"`, `"issue":"93"`, 1), kind: KindInvalid, code: "invalid-issue"},
		{name: "reason desconocida", line: strings.Replace(base, `"reason":"none"`, `"reason":"timeout"`, 1), kind: KindInvalid, code: "invalid-reason"},
		{name: "referencia con %2f en minúscula", line: strings.Replace(base, `"spec":"%2393"`, `"spec":".sdd%2fspecs%2fx.md"`, 1), kind: KindInvalid, code: "invalid-reference"},
		{name: "referencia con barra cruda", line: strings.Replace(base, `"spec":"%2393"`, `"spec":".sdd/specs/x.md"`, 1), kind: KindInvalid, code: "invalid-reference"},
		{name: "none codificado", line: strings.Replace(base, `"spec":"%2393"`, `"spec":"%6E%6F%6E%65"`, 1), kind: KindInvalid, code: "invalid-reference"},
		{name: "pr con cero", line: strings.Replace(strings.Replace(base, `"run.started"`, `"run.pr-opened"`, 1), `"pr":[]`, `"pr":[0]`, 1), kind: KindInvalid, code: "invalid-pr"},
		{name: "pr negativo", line: strings.Replace(strings.Replace(base, `"run.started"`, `"run.pr-opened"`, 1), `"pr":[]`, `"pr":[-3]`, 1), kind: KindInvalid, code: "invalid-pr"},
		{name: "pr como string", line: strings.Replace(strings.Replace(base, `"run.started"`, `"run.pr-opened"`, 1), `"pr":[]`, `"pr":["3"]`, 1), kind: KindInvalid, code: "invalid-pr"},
		{name: "land.merged con dos PRs", line: strings.Replace(strings.Replace(base, `"run.started"`, `"land.merged"`, 1), `"pr":[]`, `"pr":[101,102]`, 1), kind: KindInvalid, code: "invariant-violation"},
		{name: "run.pr-opened sin PRs", line: strings.Replace(base, `"run.started"`, `"run.pr-opened"`, 1), kind: KindInvalid, code: "invariant-violation"},
		{name: "run.failed sin reason", line: strings.Replace(base, `"run.started"`, `"run.failed"`, 1), kind: KindInvalid, code: "invariant-violation"},
		{name: "previous fuera de spec.moved", line: strings.Replace(base, `"previous":"none"`, `"previous":"%2394"`, 1), kind: KindInvalid, code: "invariant-violation"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseLine([]byte(tc.line))
			if got.Kind != tc.kind || got.Code != tc.code {
				t.Fatalf("ParseLine = kind %d code %q, want kind %d code %q", got.Kind, got.Code, tc.kind, tc.code)
			}
		})
	}
}

func TestReadLog_SaltaInvalidasYSegmentoFinal(t *testing.T) {
	vectors := readLines(t, "../../testdata/vectors/events-v1.jsonl")
	var buf bytes.Buffer
	buf.WriteString(vectors[0] + "\n")
	buf.WriteString("basura\n")
	buf.WriteString(strings.Replace(vectors[1], `"v":1,`, `"v":2,`, 1) + "\n")
	buf.WriteString(strings.Replace(vectors[2], `"harness":"codex"`, `"harness":"cursor"`, 1) + "\n")
	buf.WriteString(vectors[3] + "\n")
	buf.WriteString(vectors[4]) // segmento final sin LF: append en curso
	log, err := ReadLog(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("ReadLog: %v", err)
	}
	if len(log.Events) != 2 || log.Invalid != 2 || log.Unsupported != 1 {
		t.Fatalf("ReadLog = %d eventos, %d inválidas, %d no soportadas; want 2, 2, 1", len(log.Events), log.Invalid, log.Unsupported)
	}
	if log.Events[0].Event != "grill.paused" || log.Events[1].Event != "spec.moved" {
		t.Fatalf("eventos = %q, %q; want grill.paused, spec.moved", log.Events[0].Event, log.Events[1].Event)
	}
}

func TestSerialize_RechazaEventoInvalido(t *testing.T) {
	_, err := Serialize(Event{V: 1, TS: "2026-10-14T14:00:00Z", Harness: "claude", Event: "land.merged", PR: []int{1, 2}})
	if err == nil {
		t.Fatalf("Serialize aceptó un land.merged con dos PRs")
	}
}
