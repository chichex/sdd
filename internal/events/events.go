// Package events porta a Go el log local events-v1 (skills/docs/sdd-events-v1.md):
// una línea JSON canónica por evento en <git-common-dir>/sdd/events.jsonl.
// El parseo es total: nunca hace panic y devuelve un evento válido, una
// versión no soportada o una línea inválida con su código de diagnóstico.
package events

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/chichex/sdd/internal/tracking"
)

// Event es un evento events-v1 con las referencias decodificadas. Un campo que
// en la línea vale `none` queda vacío; `PR` vacío es `[]`.
type Event struct {
	V        int
	TS       string
	Harness  string
	Event    string
	Issue    string
	Spec     string
	Previous string
	Grill    string
	PR       []int
	Reason   string
}

// Códigos de diagnóstico estables del doc normativo.
const (
	CodeInvalidJSON        = "invalid-json"
	CodeDuplicateKey       = "duplicate-key"
	CodeMissingKey         = "missing-key"
	CodeExtraKey           = "extra-key"
	CodeKeyOrder           = "key-order"
	CodeLineTooLong        = "line-too-long"
	CodeInvalidTS          = "invalid-ts"
	CodeInvalidHarness     = "invalid-harness"
	CodeInvalidEvent       = "invalid-event"
	CodeInvalidIssue       = "invalid-issue"
	CodeInvalidReference   = "invalid-reference"
	CodeInvalidPR          = "invalid-pr"
	CodeInvalidReason      = "invalid-reason"
	CodeInvariantViolation = "invariant-violation"
	CodeNonCanonical       = "non-canonical"
	CodeUnsupportedVersion = "unsupported-version"
)

// MaxLineBytes es el tope de una línea, contando su LF final.
const MaxLineBytes = 1024

// Keys son las diez claves de una línea, en su orden canónico.
var Keys = []string{"v", "ts", "harness", "event", "issue", "spec", "previous", "grill", "pr", "reason"}

// Harnesses y Names son los catálogos cerrados de events-v1.
var (
	Harnesses = []string{"claude", "codex", "opencode", "pi", "antigravity"}
	Names     = []string{
		"grill.paused", "grill.finalized", "spec.published", "spec.moved", "spec.approved",
		"spec.superseded", "run.started", "run.pr-opened", "run.committed", "run.failed", "land.merged",
	}
)

// Error es un diagnóstico tipado de events-v1.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return "events: " + e.Code + ": " + e.Message }

func fail(code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

var timestamp = regexp.MustCompile(`^([0-9]{4}-[0-9]{2}-[0-9]{2})T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]Z$`)

func isTimestamp(ts string) bool {
	m := timestamp.FindStringSubmatch(ts)
	return m != nil && tracking.IsCalendarDate(m[1])
}

func invariant(e Event) string {
	if e.Event == "run.failed" {
		if e.Reason == "" {
			return "run.failed requires reason=aborted or reason=interrupted"
		}
	} else if e.Reason != "" {
		return e.Event + " requires reason=none"
	}
	if e.Event == "spec.moved" {
		switch {
		case e.Spec == "":
			return "spec.moved requires spec"
		case e.Previous == "":
			return "spec.moved requires previous"
		case len(e.PR) != 0:
			return "spec.moved requires pr=[]"
		}
		return ""
	}
	if e.Previous != "" {
		return e.Event + " requires previous=none"
	}
	switch e.Event {
	case "grill.paused", "grill.finalized":
		if e.Grill == "" {
			return e.Event + " requires grill"
		}
		return emptyPR(e)
	case "land.merged":
		if len(e.PR) != 1 {
			return "land.merged requires exactly one pr"
		}
		return ""
	}
	if e.Spec == "" {
		return e.Event + " requires spec"
	}
	if e.Event == "run.pr-opened" {
		if len(e.PR) == 0 {
			return "run.pr-opened requires at least one pr"
		}
		return ""
	}
	return emptyPR(e)
}

func emptyPR(e Event) string {
	if len(e.PR) != 0 {
		return e.Event + " requires pr=[]"
	}
	return ""
}

// validate devuelve el primer problema del evento, en el orden de la referencia TS.
func validate(e Event) *Error {
	if !isTimestamp(e.TS) {
		return fail(CodeInvalidTS, "invalid ts: %s", e.TS)
	}
	if !slices.Contains(Harnesses, e.Harness) {
		return fail(CodeInvalidHarness, "invalid harness: %s", e.Harness)
	}
	if !slices.Contains(Names, e.Event) {
		return fail(CodeInvalidEvent, "invalid event: %s", e.Event)
	}
	if e.Issue != "" && !tracking.IsIssueRef(e.Issue) {
		return fail(CodeInvalidIssue, "invalid issue: %s", e.Issue)
	}
	for _, ref := range []string{e.Spec, e.Previous, e.Grill} {
		if ref == "" {
			continue
		}
		if decoded, ok := tracking.DecodeRef(tracking.EncodeRef(ref)); !ok || decoded != ref {
			return fail(CodeInvalidReference, "invalid reference: %s", ref)
		}
	}
	for _, n := range e.PR {
		if n <= 0 {
			return fail(CodeInvalidPR, "pr must be an array of positive integers")
		}
	}
	if e.Reason != "" && e.Reason != "aborted" && e.Reason != "interrupted" {
		return fail(CodeInvalidReason, "invalid reason: %s", e.Reason)
	}
	if problem := invariant(e); problem != "" {
		return fail(CodeInvariantViolation, "%s", problem)
	}
	return nil
}

func wire(value string) string {
	if value == "" {
		return "none"
	}
	return value
}

func wireRef(value string) string {
	if value == "" {
		return "none"
	}
	return tracking.EncodeRef(value)
}

// Serialize devuelve la forma canónica de un evento válido, sin LF. Ante un
// evento inválido devuelve un *Error con el código del diagnóstico.
func Serialize(e Event) (string, error) {
	if e.V != 1 {
		return "", fail(CodeUnsupportedVersion, "unsupported version: %d", e.V)
	}
	if err := validate(e); err != nil {
		return "", err
	}
	prs := make([]string, len(e.PR))
	for i, n := range e.PR {
		prs[i] = strconv.Itoa(n)
	}
	line := `{"v":1,"ts":"` + e.TS + `","harness":"` + e.Harness + `","event":"` + e.Event +
		`","issue":"` + wire(e.Issue) + `","spec":"` + wireRef(e.Spec) + `","previous":"` + wireRef(e.Previous) +
		`","grill":"` + wireRef(e.Grill) + `","pr":[` + strings.Join(prs, ",") + `],"reason":"` + wire(e.Reason) + `"}`
	if len(line)+1 > MaxLineBytes {
		return "", fail(CodeLineTooLong, "the line exceeds %d bytes with its LF", MaxLineBytes)
	}
	return line, nil
}
