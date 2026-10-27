package tracking

import (
	"fmt"
	"slices"
)

// Kind clasifica el resultado de Parse.
type Kind string

// Resultados de Parse: metadata canónica válida, legacy, ausente, inválida o
// en conflicto.
const (
	KindCanonical Kind = "canonical"
	KindLegacy    Kind = "legacy"
	KindAbsent    Kind = "absent"
	KindInvalid   Kind = "invalid"
	KindConflict  Kind = "conflict"
)

// Metadata es la identidad de un artefacto. Version es 1 para la canónica y 0
// para la legacy; un campo vacío equivale a `none`. Type es spec, grill,
// project o, solo en legacy, unknown.
type Metadata struct {
	Version      int
	Type         string
	State        string
	Issue        string
	Grill        string
	SupersededBy string
	Project      string
	GeneratedAt  string
}

// State es el estado normalizado: Value es un estado de spec o grill o
// `unknown`; Provenance es canonical, legacy-explicit, legacy-inferred o absent.
type State struct {
	Value      string
	Provenance string
	Raw        string
}

// Diagnostic es un diagnóstico tipado y estable.
type Diagnostic struct {
	Code    string
	Message string
	Raw     string
}

// Result es el resultado total de Parse: nunca hace panic.
type Result struct {
	Kind        Kind
	Metadata    Metadata
	Raw         []string
	State       State
	Diagnostics []Diagnostic
}

// Error es un diagnóstico devuelto por Serialize.
type Error struct{ Diagnostic }

func (e *Error) Error() string { return "tracking: " + e.Code + ": " + e.Message }

func diag(code, format string, args ...any) *Error {
	return &Error{Diagnostic{Code: code, Message: fmt.Sprintf(format, args...)}}
}

var (
	specStates  = []string{"draft", "approved", "implemented", "superseded"}
	grillStates = []string{"paused", "finalized"}
)

func validate(m Metadata) *Error {
	if m.Version != 1 {
		return diag("unsupported-version", "unsupported metadata version: %d", m.Version)
	}
	switch m.Type {
	case "project":
		if !IsCalendarDate(m.GeneratedAt) {
			return diag("invalid-date", "invalid generated-at date: %s", m.GeneratedAt)
		}
		return nil
	case "spec", "grill":
	default:
		return diag("unknown-type", "unknown metadata type: %s", m.Type)
	}
	if m.Issue != "" && !IsIssueRef(m.Issue) {
		return diag("invalid-issue", "invalid issue reference: %s", m.Issue)
	}
	if m.Type == "spec" {
		return validateSpec(m)
	}
	if !slices.Contains(grillStates, m.State) {
		return diag("invalid-state", "invalid grill state: %s", m.State)
	}
	if !isWellFormedRef(m.Grill) {
		return diag("invalid-reference", "invalid grill reference: %s", m.Grill)
	}
	if !isWellFormedRef(m.Project) {
		return diag("invalid-reference", "invalid project reference: %s", m.Project)
	}
	return nil
}

func validateSpec(m Metadata) *Error {
	switch {
	case !slices.Contains(specStates, m.State):
		return diag("invalid-state", "invalid spec state: %s", m.State)
	case m.Grill != "" && !isWellFormedRef(m.Grill):
		return diag("invalid-reference", "invalid grill reference: %s", m.Grill)
	case m.SupersededBy != "" && !isWellFormedRef(m.SupersededBy):
		return diag("invalid-reference", "invalid superseded-by reference: %s", m.SupersededBy)
	case m.State == "superseded" && m.SupersededBy == "":
		return diag("invariant-violation", "a superseded spec requires superseded-by")
	case m.State != "superseded" && m.SupersededBy != "":
		return diag("invariant-violation", "only a superseded spec may have superseded-by")
	}
	return nil
}

func optionalRef(value string) string {
	if value == "" {
		return "none"
	}
	return EncodeRef(value)
}

// Serialize devuelve el marker canónico de una metadata v1 completa y válida.
// La metadata legacy (Version 0) no se serializa.
func Serialize(m Metadata) (string, error) {
	marker, err := serialize(m)
	if err != nil {
		return "", err
	}
	return marker, nil
}

func serialize(m Metadata) (string, *Error) {
	if err := validate(m); err != nil {
		return "", err
	}
	switch m.Type {
	case "spec":
		return "<!-- SDD-Tracking: version=1; type=spec; state=" + m.State + "; issue=" + issueOrNone(m.Issue) +
			"; grill=" + optionalRef(m.Grill) + "; superseded-by=" + optionalRef(m.SupersededBy) + " -->", nil
	case "grill":
		return "<!-- SDD-Tracking: version=1; type=grill; state=" + m.State + "; issue=" + issueOrNone(m.Issue) +
			"; grill=" + EncodeRef(m.Grill) + "; project=" + EncodeRef(m.Project) + " -->", nil
	}
	return "<!-- SDD-Tracking: version=1; type=project; generated-at=" + m.GeneratedAt + " -->", nil
}

func issueOrNone(issue string) string {
	if issue == "" {
		return "none"
	}
	return issue
}

func absentState() State { return State{Value: "unknown", Provenance: "absent"} }

// Parse lee la metadata del preámbulo de un artefacto. Un marker que declara
// version o type es candidato canónico y nunca cae a legacy; dos markers
// canónicos idénticos son válidos con `duplicate-canonical`, divergentes son
// conflicto. Sin candidatos, aplica la compatibilidad legacy.
func Parse(markdown string) Result {
	var candidates []string
	for _, line := range markerLines(markdown) {
		if declaresCanonical.MatchString(line) {
			candidates = append(candidates, line)
		}
	}
	if len(candidates) == 0 {
		return parseLegacy(markdown)
	}
	for _, c := range candidates[1:] {
		if c != candidates[0] {
			return Result{Kind: KindConflict, Raw: candidates, State: absentState(),
				Diagnostics: []Diagnostic{{Code: "conflicting-canonical", Message: "canonical markers differ"}}}
		}
	}
	raw := candidates[0]
	var dup []Diagnostic
	if len(candidates) > 1 {
		dup = []Diagnostic{{Code: "duplicate-canonical", Message: "canonical marker is duplicated", Raw: raw}}
	}
	meta, problem := parseCanonical(raw)
	if problem != nil {
		problem.Raw = raw
		return Result{Kind: KindInvalid, Raw: []string{raw}, State: absentState(),
			Diagnostics: append([]Diagnostic{problem.Diagnostic}, dup...)}
	}
	state := absentState()
	if meta.Type != "project" {
		state = State{Value: meta.State, Provenance: "canonical", Raw: meta.State}
	}
	return Result{Kind: KindCanonical, Metadata: meta, Raw: []string{raw}, State: state, Diagnostics: dup}
}
