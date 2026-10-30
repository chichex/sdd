package events

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"

	"github.com/chichex/sdd/internal/tracking"
)

// Kind clasifica el resultado del parseo de una línea.
type Kind int

// Resultados posibles de ParseLine.
const (
	KindEvent Kind = iota
	KindUnsupported
	KindInvalid
)

// Result es el resultado total de ParseLine: un evento válido, una versión no
// soportada o una línea inválida; Code y Message describen el diagnóstico.
type Result struct {
	Kind    Kind
	Event   Event
	Code    string
	Message string
}

func invalid(code, format string, args ...any) Result {
	return Result{Kind: KindInvalid, Code: code, Message: fmt.Sprintf(format, args...)}
}

// ParseLine parsea una línea sin su LF. Es total: nunca hace panic.
func ParseLine(line []byte) Result {
	var raw map[string]json.RawMessage
	if !json.Valid(line) {
		return invalid(CodeInvalidJSON, "the line is not valid JSON")
	}
	if err := json.Unmarshal(line, &raw); err != nil || raw == nil {
		return invalid(CodeInvalidJSON, "the line is not a JSON object")
	}
	v, ok := raw["v"]
	if !ok {
		return invalid(CodeMissingKey, "missing key: v")
	}
	var version float64
	if err := json.Unmarshal(v, &version); err != nil || version != 1 {
		return Result{Kind: KindUnsupported, Code: CodeUnsupportedVersion, Message: "unsupported version: " + string(v)}
	}
	if len(line)+1 > MaxLineBytes {
		return invalid(CodeLineTooLong, "the line exceeds %d bytes with its LF", MaxLineBytes)
	}
	if res, bad := checkKeys(line); bad {
		return res
	}
	event, res, bad := decodeFields(raw)
	if bad {
		return res
	}
	canonical, err := Serialize(event)
	if err != nil {
		var diag *Error
		if errors.As(err, &diag) {
			return invalid(diag.Code, "%s", diag.Message)
		}
		return invalid(CodeInvalidJSON, "%v", err)
	}
	if canonical != string(line) {
		return invalid(CodeNonCanonical, "the line is not in its canonical byte form")
	}
	return Result{Kind: KindEvent, Event: event}
}

// topLevelKeys devuelve las claves del objeto en orden de aparición,
// incluidas las repetidas que json.Unmarshal esconde.
func topLevelKeys(line []byte) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(line))
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("events: object start: %w", err)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("events: key: %w", err)
		}
		key, _ := tok.(string)
		keys = append(keys, key)
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, fmt.Errorf("events: value: %w", err)
		}
	}
	return keys, nil
}

func checkKeys(line []byte) (Result, bool) {
	keys, err := topLevelKeys(line)
	if err != nil {
		return invalid(CodeInvalidJSON, "%v", err), true
	}
	seen := map[string]bool{}
	for _, k := range keys {
		if seen[k] {
			return invalid(CodeDuplicateKey, "repeated key: %s", k), true
		}
		seen[k] = true
	}
	for _, k := range Keys {
		if !seen[k] {
			return invalid(CodeMissingKey, "missing key: %s", k), true
		}
	}
	for _, k := range keys {
		if !slices.Contains(Keys, k) {
			return invalid(CodeExtraKey, "unexpected key: %s", k), true
		}
	}
	if !slices.Equal(keys, Keys) {
		return invalid(CodeKeyOrder, "keys are out of order"), true
	}
	return Result{}, false
}

// asString devuelve el string JSON o false si el valor es de otro tipo.
func asString(raw json.RawMessage) (string, bool) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	return s, true
}

// nullable lee un string donde `none` es el vacío.
func nullable(raw json.RawMessage) (string, bool) {
	s, ok := asString(raw)
	if !ok || s == "none" {
		return "", ok
	}
	return s, true
}

func decodeRef(raw json.RawMessage) (string, bool) {
	s, ok := asString(raw)
	if !ok {
		return "", false
	}
	if s == "none" {
		return "", true
	}
	return tracking.DecodeRef(s)
}

func decodePRs(raw json.RawMessage) ([]int, bool) {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || items == nil {
		return nil, false
	}
	prs := make([]int, 0, len(items))
	for _, item := range items {
		var f float64
		// Un string JSON ("3") no es un número aunque json.Number lo aceptaría.
		if err := json.Unmarshal(item, &f); err != nil || f != math.Trunc(f) || f > 1<<53-1 || f <= 0 {
			return nil, false
		}
		prs = append(prs, int(f))
	}
	return prs, true
}

func decodeFields(raw map[string]json.RawMessage) (Event, Result, bool) {
	e := Event{V: 1}
	var ok bool
	if e.Issue, ok = nullable(raw["issue"]); !ok {
		return e, invalid(CodeInvalidIssue, "issue must be a string"), true
	}
	for _, f := range []struct {
		key string
		dst *string
	}{{"spec", &e.Spec}, {"previous", &e.Previous}, {"grill", &e.Grill}} {
		if *f.dst, ok = decodeRef(raw[f.key]); !ok {
			return e, invalid(CodeInvalidReference, "invalid %s reference: %s", f.key, raw[f.key]), true
		}
	}
	if e.Reason, ok = nullable(raw["reason"]); !ok {
		return e, invalid(CodeInvalidReason, "reason must be a string"), true
	}
	if e.TS, ok = asString(raw["ts"]); !ok {
		return e, invalid(CodeInvalidTS, "invalid ts: %s", raw["ts"]), true
	}
	if e.Harness, ok = asString(raw["harness"]); !ok {
		return e, invalid(CodeInvalidHarness, "invalid harness: %s", raw["harness"]), true
	}
	if e.Event, ok = asString(raw["event"]); !ok {
		return e, invalid(CodeInvalidEvent, "invalid event: %s", raw["event"]), true
	}
	if e.PR, ok = decodePRs(raw["pr"]); !ok {
		return e, invalid(CodeInvalidPR, "pr must be an array of positive integers"), true
	}
	return e, Result{}, false
}

// Log es lo que ReadLog leyó: los eventos válidos en orden de archivo y
// cuántas líneas salteó por inválidas o por versión no soportada.
type Log struct {
	Events      []Event
	Invalid     int
	Unsupported int
}

// ReadLog lee un events.jsonl línea por línea: saltea las inválidas y las de
// otra versión contándolas, sigue leyendo y descarta un segmento final sin LF
// (un append en curso o truncado). Solo devuelve error de lectura.
func ReadLog(r io.Reader) (Log, error) {
	var log Log
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			return log, nil
		}
		if err != nil {
			return log, fmt.Errorf("events: read log: %w", err)
		}
		res := ParseLine(line[:len(line)-1])
		switch res.Kind {
		case KindEvent:
			log.Events = append(log.Events, res.Event)
		case KindUnsupported:
			log.Unsupported++
		default:
			log.Invalid++
		}
	}
}
